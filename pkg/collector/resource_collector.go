package collector

import (
	"context"
	"encoding/json"
	"flextopo/pkg/graph"
	"flextopo/pkg/utils"
	"fmt"
	"os/exec"

	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	// 获取当前节点上的 Pod 列表
	pods, err := rc.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.nodeName", rc.nodeName).String(),
	})
	if err != nil {
		return err
	}

	// 遍历 Pod，更新资源分配状态
	for _, pod := range pods.Items {
		rc.processPod(&pod, graph)
	}

	return nil
}

// processPod 处理单个 Pod，更新资源分配状态
func (rc *ResourceCollector) processPod(pod *corev1.Pod, graph *graph.FlexTopoGraph) {
	// 获取 Pod 的所有容器的进程 ID
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerID := containerStatus.ContainerID
		if containerID == "" {
			continue
		}

		// 提取容器运行时和容器 ID
		parts := strings.Split(containerID, "://")
		if len(parts) != 2 {
			continue
		}
		runtime := parts[0]
		id := parts[1]

		// 获取容器的进程 ID（PID）
		pid, err := rc.getContainerPID(runtime, id)
		if err != nil {
			rc.logger.Warn("Failed to get PID for container " + id + ": " + err.Error())
			continue
		}

		// 获取容器实际使用的 CPU 核心
		cpuCores, err := rc.getContainerCPUCores(pid)
		if err != nil {
			rc.logger.Warn("Failed to get CPU cores for container " + id + ": " + err.Error())
			continue
		}

		// 更新拓扑图中对应 CPU Core 节点的状态
		graph.UpdateCPUUsage(pod.Name, cpuCores)

		// 获取容器实际使用的 GPU
		gpuUUIDs, err := rc.getContainerGPUs(pid)
		if err != nil {
			rc.logger.Warn("Failed to get GPUs for container " + id + ": " + err.Error())
			continue
		}

		// 更新拓扑图中对应 GPU 节点的状态
		graph.UpdateGPUUsage(pod.Name, gpuUUIDs)
	}
}

// getContainerPID 获取容器的主进程 PID
func (rc *ResourceCollector) getContainerPID(runtime, id string) (string, error) {
	// 针对不同的容器运行时，使用不同的方法获取 PID
	// 以下以 containerd 为例
	if runtime == "containerd" {
		// 使用 crictl 工具
		out, err := exec.Command("crictl", "inspect", "--output=json", id).Output()
		if err != nil {
			return "", err
		}
		// 解析 JSON，获取 PID
		var data map[string]interface{}
		json.Unmarshal(out, &data)
		info, ok := data["info"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("failed to parse container info")
		}
		pid := fmt.Sprintf("%.0f", info["pid"].(float64))
		return pid, nil
	}
	// TODO: 支持其他容器运行时
	return "", fmt.Errorf("unsupported runtime: %s", runtime)
}

// getContainerCPUCores 获取容器实际使用的 CPU 核心列表
func (rc *ResourceCollector) getContainerCPUCores(pid string) ([]int, error) {
	// 读取 /proc/<pid>/status 中的 Cpus_allowed_list
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
	// 解析 CPU 列表
	cpuCores, err := utils.ParseCPUList(cpuListLine)
	if err != nil {
		return nil, err
	}
	return cpuCores, nil
}

// getContainerGPUs 获取容器实际使用的 GPU UUID 列表
func (rc *ResourceCollector) getContainerGPUs(pid string) ([]string, error) {
	// 使用 nvidia-smi 工具，获取每个 GPU 上运行的进程信息
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
