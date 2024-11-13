package collector

import (
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
	// err = hc.collectGPUInfo(graph)
	// if err != nil {
	// 	return nil, err
	// }

	return graph, nil
}

// collectCPUNUMAInfo collects CPU and NUMA node information
func (hc *HardwareCollector) collectCPUNUMAInfo(graph *graph.FlexTopoGraph) error {
	hc.logger.Info("Collecting CPU and NUMA information")

	// Use lscpu command to get CPU and NUMA information
	// out, err := exec.Command("lscpu", "-p=CPU,Core,Socket,Node").Output()
	// if err != nil {
	// 	return fmt.Errorf("failed to execute lscpu: %v", err)
	// }

	out := `# The following is the parsable format, which can be fed to other
# programs. Each different item in every column has an unique ID
# starting from zero.
# CPU,Core,Socket,Node
0,0,0,0
1,1,0,0
2,2,0,0
3,3,0,0
4,4,0,0
5,5,0,0
6,6,0,0
7,7,0,0
8,8,0,1
9,9,0,1
10,10,0,1
11,11,0,1
12,12,0,1
13,13,0,1
14,14,0,1
15,15,0,1
16,16,0,2
17,17,0,2
18,18,0,2
19,19,0,2
20,20,0,2
21,21,0,2
22,22,0,2
23,23,0,2
24,24,0,3
25,25,0,3
26,26,0,3
27,27,0,3
28,28,0,3
29,29,0,3
30,30,0,3
31,31,0,3
32,32,1,4
33,33,1,4
34,34,1,4
35,35,1,4
36,36,1,4
37,37,1,4
38,38,1,4
39,39,1,4
40,40,1,5
41,41,1,5
42,42,1,5
43,43,1,5
44,44,1,5
45,45,1,5
46,46,1,5
47,47,1,5
48,48,1,6
49,49,1,6
50,50,1,6
51,51,1,6
52,52,1,6
53,53,1,6
54,54,1,6
55,55,1,6
56,56,1,7
57,57,1,7
58,58,1,7
59,59,1,7
60,60,1,7
61,61,1,7
62,62,1,7
63,63,1,7`

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
