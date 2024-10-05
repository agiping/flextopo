package collector

import (
	"flextopo/pkg/graph"
	"flextopo/pkg/utils"
)

// Collector 接口定义了收集拓扑信息的方法
type Collector interface {
	// Collect 收集硬件拓扑和资源分配信息，返回构建的 FlexTopo 图
	Collect() (*graph.FlexTopoGraph, error)
}

// DefaultCollector 实现了 Collector 接口
type DefaultCollector struct {
	hardwareCollector *HardwareCollector
	resourceCollector *ResourceCollector
	logger            utils.Logger
}

// NewCollector 创建一个新的 DefaultCollector 实例
func NewCollector(nodeName string, logger utils.Logger) (Collector, error) {
	hardwareCollector := NewHardwareCollector(logger)
	resourceCollector, err := NewResourceCollector(nodeName, logger)
	if err != nil {
		return nil, err
	}
	return &DefaultCollector{
		hardwareCollector: hardwareCollector,
		resourceCollector: resourceCollector,
		logger:            logger,
	}, nil
}

// Collect 收集硬件和资源分配信息
func (dc *DefaultCollector) Collect() (*graph.FlexTopoGraph, error) {
	// 收集硬件拓扑信息
	graph, err := dc.hardwareCollector.CollectHardwareInfo()
	if err != nil {
		return nil, err
	}

	// 收集资源分配信息
	err = dc.resourceCollector.CollectResourceInfo(graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
