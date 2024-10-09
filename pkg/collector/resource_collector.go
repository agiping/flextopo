package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"flextopo/pkg/graph"
	"flextopo/pkg/utils"
)

type ResourceCollector struct {
	clientset *kubernetes.Clientset
	nodeName  string
	logger    utils.Logger
}

func NewResourceCollector(nodeName string, logger utils.Logger) (*ResourceCollector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &ResourceCollector{
		clientset: clientset,
		nodeName:  nodeName,
		logger:    logger,
	}, nil
}

func (rc *ResourceCollector) CollectResourceInfo(graph *graph.FlexTopoGraph) error {
	rc.logger.Info("Collecting dynamic resource allocation information")

	// Get the list of Pods on the current node
	pods, err := rc.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.nodeName", rc.nodeName).String(),
	})
	if err != nil {
		return err
	}

	// Iterate through Pods and update resource allocation status
	for _, pod := range pods.Items {
		rc.processPod(&pod, graph)
	}

	return nil
}

// processPod processes a single Pod and updates the resource allocation status on the FlexTopo graph
func (rc *ResourceCollector) processPod(pod *corev1.Pod, graph *graph.FlexTopoGraph) {
	// Get the process IDs of all containers in the Pod
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerID := containerStatus.ContainerID
		if containerID == "" {
			continue
		}

		// Extract container runtime and container ID
		parts := strings.Split(containerID, "://")
		if len(parts) != 2 {
			continue
		}
		runtime := parts[0]
		id := parts[1]

		// Get the process ID (PID) of the container
		pid, err := rc.getContainerPID(runtime, id)
		if err != nil {
			rc.logger.Warn("Failed to get PID for container " + id + ": " + err.Error())
			continue
		}

		// Get the CPU cores actually used by the container
		cpuCores, err := rc.getContainerCPUCores(pid)
		if err != nil {
			rc.logger.Warn("Failed to get CPU cores for container " + id + ": " + err.Error())
			continue
		}

		// Update the status of corresponding CPU Core nodes in the topology graph
		graph.UpdateCPUUsage(pod.Name, cpuCores)

		// Get the GPUs actually used by the container
		gpuUUIDs, err := rc.getContainerGPUs(pid)
		if err != nil {
			rc.logger.Warn("Failed to get GPUs for container " + id + ": " + err.Error())
			continue
		}

		// Update the status of corresponding GPU nodes in the topology graph
		graph.UpdateGPUUsage(pod.Name, gpuUUIDs)
	}
}

// getContainerPID gets the main process PID of the container
func (rc *ResourceCollector) getContainerPID(runtime, id string) (string, error) {
	// Use different methods to get PID for different container runtimes
	// The following is for containerd
	if runtime == "containerd" {
		// Use crictl tool
		out, err := exec.Command("crictl", "inspect", "--output=json", id).Output()
		if err != nil {
			return "", err
		}
		// Parse JSON to get PID
		var data map[string]interface{}
		json.Unmarshal(out, &data)
		info, ok := data["info"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("failed to parse container info")
		}
		pidFloat, ok := info["pid"].(float64)
		if !ok {
			return "", fmt.Errorf("PID is not a number")
		}
		pid := strconv.Itoa(int(pidFloat))
		return pid, nil
	}
	// TODO: Support other container runtimes
	return "", fmt.Errorf("unsupported runtime: %s", runtime)
}

// getContainerCPUCores gets the list of CPU cores actually used by the container
func (rc *ResourceCollector) getContainerCPUCores(pid string) ([]int, error) {
	// Read Cpus_allowed_list from /proc/<pid>/status
	path := filepath.Join("/proc", pid, "status")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	var cpuListLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "Cpus_allowed_list:") {
			cpuListLine = strings.TrimSpace(strings.TrimPrefix(line, "Cpus_allowed_list:"))
			break
		}
	}
	if cpuListLine == "" {
		return nil, fmt.Errorf("failed to find Cpus_allowed_list in %s", path)
	}
	// Parse CPU list
	cpuCores, err := utils.ParseCPUList(cpuListLine)
	if err != nil {
		return nil, err
	}
	return cpuCores, nil
}

// getContainerGPUs gets the list of GPU UUIDs actually used by the container
func (rc *ResourceCollector) getContainerGPUs(pid string) ([]string, error) {
	// Use nvidia-smi tool to get process information running on each GPU
	out, err := exec.Command("nvidia-smi", "--query-compute-apps=pid,gpu_uuid", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var gpuUUIDs []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ", ")
		if len(fields) != 2 {
			continue
		}
		gpuPID := fields[0]
		gpuUUID := fields[1]
		if gpuPID == pid {
			gpuUUIDs = append(gpuUUIDs, gpuUUID)
		}
	}
	return gpuUUIDs, nil
}
