package collector

import (
	"flextopo/pkg/graph"
	"flextopo/pkg/utils"
)

// Collector interface defines methods for collecting topology information
type Collector interface {
	// Collect gathers hardware topology and resource allocation information, returns the constructed FlexTopo graph
	Collect() (*graph.FlexTopoGraph, error)
}

// DefaultCollector implements the Collector interface
type DefaultCollector struct {
	hardwareCollector *HardwareCollector
	resourceCollector *ResourceCollector
	logger            utils.Logger
}

// NewCollector creates a new instance of DefaultCollector
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

// Collect gathers hardware and resource allocation information
func (dc *DefaultCollector) Collect() (*graph.FlexTopoGraph, error) {
	// Collect hardware topology information
	graph, err := dc.hardwareCollector.CollectHardwareInfo()
	if err != nil {
		return nil, err
	}

	// Collect resource allocation information
	err = dc.resourceCollector.CollectResourceInfo(graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
