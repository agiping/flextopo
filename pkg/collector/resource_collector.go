package collector

import (
	"context"

	"flextopo/pkg/graph"
	"flextopo/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ResourceCollector 负责收集动态资源分配信息
type ResourceCollector struct {
	clientset *kubernetes.Clientset
	nodeName  string
	logger    utils.Logger
}

// NewResourceCollector 创建一个新的 ResourceCollector 实例
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

// CollectResourceInfo 收集动态资源分配信息
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
	// 获取 Pod 使用的资源，包括 CPU、内存、GPU 等
	for _, container := range pod.Spec.Containers {
		resources := container.Resources

		// 更新 CPU 核心的分配状态
		cpuQuantity := resources.Requests.Cpu()
		if cpuQuantity != nil && cpuQuantity.MilliValue() > 0 {
			// 更新对应的 CPU 核心节点状态
			graph.UpdateCPUAllocation(pod, cpuQuantity)
		}

		// 更新 GPU 的分配状态
		gpuQuantity := resources.Limits["nvidia.com/gpu"]
		if gpuQuantity.Value() > 0 {
			// 更新对应的 GPU 节点状态
			graph.UpdateGPUAllocation(pod, &gpuQuantity)
		}

		// 其他资源的处理
	}
}
