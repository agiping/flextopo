package collector

import (
	"fmt"
	"os/exec"
	"strings"

	"flextopo/pkg/graph"
	"flextopo/pkg/utils"
)

// HardwareCollector 负责收集硬件拓扑信息
type HardwareCollector struct {
	logger utils.Logger
}

// NewHardwareCollector 创建一个新的 HardwareCollector 实例
func NewHardwareCollector(logger utils.Logger) *HardwareCollector {
	return &HardwareCollector{
		logger: logger,
	}
}

// CollectHardwareInfo 收集硬件拓扑信息
func (hc *HardwareCollector) CollectHardwareInfo() (*graph.FlexTopoGraph, error) {
	// 从配置获取 CoreGroupSize
	coreGroupSize := utils.GetConfig().CoreGroupSize
	graph := graph.NewFlexTopoGraph(coreGroupSize)

	// 收集 CPU、NUMA 信息
	err := hc.collectCPUNUMAInfo(graph)
	if err != nil {
		return nil, err
	}

	// 收集 GPU 信息
	err = hc.collectGPUInfo(graph)
	if err != nil {
		return nil, err
	}

	// 收集 PCIe 拓扑信息
	// err = hc.collectPCIeInfo(graph)
	// if err != nil {
	// 	return nil, err
	// }

	return graph, nil
}

// collectCPUNUMAInfo 收集 CPU 和 NUMA 节点信息
func (hc *HardwareCollector) collectCPUNUMAInfo(graph *graph.FlexTopoGraph) error {
	hc.logger.Info("Collecting CPU and NUMA information")

	// 使用 lscpu 命令获取 CPU 和 NUMA 信息
	out, err := exec.Command("lscpu", "-p=CPU,Core,Socket,Node").Output()
	if err != nil {
		return fmt.Errorf("failed to execute lscpu: %v", err)
	}

	// 解析输出，构建节点和关系
	cpuInfo := utils.ParseLSCPUOutput(string(out))
	graph.BuildCPUNodes(cpuInfo)

	return nil
}

// collectGPUInfo 收集 GPU 信息
func (hc *HardwareCollector) collectGPUInfo(graph *graph.FlexTopoGraph) error {
	hc.logger.Info("Collecting GPU information")

	// 使用 nvidia-smi 命令获取 GPU 信息
	out, err := exec.Command("nvidia-smi", "--query-gpu=index,uuid,name,memory.total", "--format=csv,noheader,nounits").Output()
	if err != nil {
		hc.logger.Warn("nvidia-smi command failed, assuming no GPUs present")
		return nil // 无 GPU，直接返回
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

// // collectPCIeInfo 收集 PCIe 拓扑信息
// func (hc *HardwareCollector) collectPCIeInfo(graph *graph.FlexTopoGraph) error {
// 	hc.logger.Info("Collecting PCIe topology information")

// 	// PCIe 拓扑信息的收集可能比较复杂，视情况而定
// 	// 这里可以使用 lspci 命令，或者读取 /sys/bus/pci 目录下的信息

// 	// 简化处理，暂不实现
// 	return nil
// }
