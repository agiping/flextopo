package collector

import (
	"reflect"
	"testing"

	"flextopo/pkg/utils"
)

// var logger = &utils.SimpleLogger{}

func TestParseLSCPUOutput(t *testing.T) {
	testInput := `# The following is the parsable format, which can be fed to other
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

	// Call the function being tested
	result := utils.ParseLSCPUOutput(testInput)

	// Verify the result
	expectedLength := 64
	if len(result) != expectedLength {
		t.Errorf("Expected %d CPUInfo items, but got %d", expectedLength, len(result))
	}

	// Check some specific values
	expectedValues := []utils.CPUInfo{
		{CPUID: 0, CoreID: 0, SocketID: 0, NumaNodeID: 0},
		{CPUID: 7, CoreID: 7, SocketID: 0, NumaNodeID: 0},
		{CPUID: 8, CoreID: 8, SocketID: 0, NumaNodeID: 1},
		{CPUID: 31, CoreID: 31, SocketID: 0, NumaNodeID: 3},
		{CPUID: 32, CoreID: 32, SocketID: 1, NumaNodeID: 4},
		{CPUID: 63, CoreID: 63, SocketID: 1, NumaNodeID: 7},
	}

	for _, expected := range expectedValues {
		if !containsCPUInfo(result, expected) {
			t.Errorf("Expected CPUInfo not found: %+v", expected)
		}
	}
}

// Helper function: check if a slice contains a specific CPUInfo
func containsCPUInfo(slice []utils.CPUInfo, item utils.CPUInfo) bool {
	for _, v := range slice {
		if reflect.DeepEqual(v, item) {
			return true
		}
	}
	return false
}
