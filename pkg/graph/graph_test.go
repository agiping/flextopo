package graph

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"flextopo/pkg/utils"
)

func TestBuildCPUNodes(t *testing.T) {
	// Prepare test data, the following is the hardware structure data of a 4090 Server
	lscpuOutput := `# CPU,Core,Socket,Node
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

	// Parse CPU information
	cpuInfos := utils.ParseLSCPUOutput(lscpuOutput)

	// Create FlexTopoGraph instance
	coreGroupSize := 8
	graph := NewFlexTopoGraph(coreGroupSize)

	// Call the function being tested
	graph.BuildCPUNodes(cpuInfos)

	// Verify results
	assert.Equal(t, 2, len(graph.getNodesByType("Socket")), "There should be 2 Socket nodes")
	assert.Equal(t, 8, len(graph.getNodesByType("NUMANode")), "There should be 8 NUMA nodes")
	assert.Equal(t, 8, len(graph.getNodesByType("CoreGroup")), "There should be 8 CoreGroup nodes")
	assert.Equal(t, 64, len(graph.getNodesByType("CPUCore")), "There should be 64 CPU Core nodes")

	// Verify connections from Socket to NUMA
	socket0 := graph.getNode("socket-0", "Socket")
	socket1 := graph.getNode("socket-1", "Socket")
	assert.Equal(t, 4, len(graph.getEdges(socket0, "contains")), "Socket 0 should contain 4 NUMA nodes")
	assert.Equal(t, 4, len(graph.getEdges(socket1, "contains")), "Socket 1 should contain 4 NUMA nodes")

	// Verify connections from NUMA to CoreGroup
	for i := 0; i < 8; i++ {
		numaNode := graph.getNode(fmt.Sprintf("numa-%d", i), "NUMANode")
		assert.Equal(t, 1, len(graph.getEdges(numaNode, "contains")), "Each NUMA node should contain 1 CoreGroup")
	}

	// Verify connections from CoreGroup to CPUCore
	for i := 0; i < 8; i++ {
		coreGroup := graph.getNode(fmt.Sprintf("coregroup-%d-%d", i, i), "CoreGroup")
		assert.Equal(t, 8, len(graph.getEdges(coreGroup, "contains")), "Each CoreGroup should contain 8 CPU Cores")
	}

	// Verify CPUCore attributes
	for i := 0; i < 64; i++ {
		cpuCore := graph.getNode(fmt.Sprintf("core-%d", i), "CPUCore")
		assert.Equal(t, "free", cpuCore.Attributes["status"], "The status of each CPU Core should be free")
	}
}
