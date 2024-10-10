package collector

import (
	"fmt"
	"os/exec"
	"strings"

	"flextopo/pkg/graph"
	"flextopo/pkg/utils"
)

// HardwareCollector is responsible for collecting hardware topology information
type HardwareCollector struct {
	logger utils.Logger
}

// NewHardwareCollector creates a new instance of HardwareCollector
func NewHardwareCollector(logger utils.Logger) *HardwareCollector {
	return &HardwareCollector{
		logger: logger,
	}
}

// CollectHardwareInfo collects hardware topology information
func (hc *HardwareCollector) CollectHardwareInfo() (*graph.FlexTopoGraph, error) {
	// Get CoreGroupSize from configuration
	coreGroupSize := utils.GetConfig().CoreGroupSize
	graph := graph.NewFlexTopoGraph(coreGroupSize)

	// Collect CPU and NUMA information
	err := hc.collectCPUNUMAInfo(graph)
	if err != nil {
		return nil, err
	}

	// Collect GPU information
	err = hc.collectGPUInfo(graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}

// collectCPUNUMAInfo collects CPU and NUMA node information
func (hc *HardwareCollector) collectCPUNUMAInfo(graph *graph.FlexTopoGraph) error {
	hc.logger.Info("Collecting CPU and NUMA information")

	// Use lscpu command to get CPU and NUMA information
	out, err := exec.Command("lscpu", "-p=CPU,Core,Socket,Node").Output()
	if err != nil {
		return fmt.Errorf("failed to execute lscpu: %v", err)
	}

	// Parse output, build nodes and relationships
	cpuInfo := utils.ParseLSCPUOutput(string(out))
	graph.BuildCPUNodes(cpuInfo)

	return nil
}

// collectGPUInfo collects GPU information
func (hc *HardwareCollector) collectGPUInfo(graph *graph.FlexTopoGraph) error {
	hc.logger.Info("Collecting GPU information")

	// Use nvidia-smi command to get GPU information
	out, err := exec.Command("/host-bin/nvidia-smi", "--query-gpu=index,uuid,name,memory.total", "--format=csv,noheader,nounits").Output()
	if err != nil {
		hc.logger.Warn("nvidia-smi command failed, assuming no GPUs present")
		return nil // No GPUs, return directly
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ", ")
		if len(fields) < 4 {
			continue
		}
		index := utils.Atoi(fields[0])
		uuid := fields[1]
		name := fields[2]
		memoryTotal := utils.Atoi(fields[3])

		gpuNode := graph.NewGPUNode(index, uuid, name, memoryTotal)
		graph.AddNode(gpuNode)
	}

	return nil
}
