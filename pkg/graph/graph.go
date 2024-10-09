package graph

import (
	"flextopo/pkg/crd"
	"flextopo/pkg/utils"
	"fmt"
	"sort"
)

// FlexTopoGraph represents the entire topology graph
type FlexTopoGraph struct {
	Nodes map[string]*Node // key: nodeID, value: node
	// key: edgeID, value: edge
	// edgeID = sourceID + "-" + targetID + "-" + edgeType
	Edges map[string]*Edge
	// CPU core group size, configurable
	CoreGroupSize int
}

// NewFlexTopoGraph creates a new instance of FlexTopoGraph
func NewFlexTopoGraph(coreGroupSize int) *FlexTopoGraph {
	return &FlexTopoGraph{
		Nodes:         make(map[string]*Node),
		Edges:         make(map[string]*Edge), // Initialize as an empty map
		CoreGroupSize: coreGroupSize,
	}
}

// BuildCPUNodes builds nodes and edges based on CPU information
func (g *FlexTopoGraph) BuildCPUNodes(cpuInfos []utils.CPUInfo) {
	// Sort CPUInfo by Socket, NUMA, CoreID
	sort.Slice(cpuInfos, func(i, j int) bool {
		if cpuInfos[i].SocketID != cpuInfos[j].SocketID {
			return cpuInfos[i].SocketID < cpuInfos[j].SocketID
		}
		if cpuInfos[i].NumaNodeID != cpuInfos[j].NumaNodeID {
			return cpuInfos[i].NumaNodeID < cpuInfos[j].NumaNodeID
		}
		return cpuInfos[i].CoreID < cpuInfos[j].CoreID
	})

	// Build nodes
	for _, cpuInfo := range cpuInfos {
		socketID := fmt.Sprintf("socket-%d", cpuInfo.SocketID)
		numaNodeID := fmt.Sprintf("numa-%d", cpuInfo.NumaNodeID)
		coreID := fmt.Sprintf("core-%d", cpuInfo.CoreID)

		// Create or get Socket node
		socketNode := g.getNode(socketID, "Socket")

		// Create or get NUMA Node node
		numaNode := g.getNode(numaNodeID, "NUMANode")
		g.addEdge(socketNode, numaNode, "contains")

		// Create or get Core Group node
		coreGroupID := fmt.Sprintf("coregroup-%d-%d", cpuInfo.NumaNodeID, cpuInfo.CoreID/g.CoreGroupSize)
		coreGroupNode := g.getNode(coreGroupID, "CoreGroup")
		coreGroupNode.Attributes["nodeID"] = cpuInfo.NumaNodeID
		coreGroupNode.Attributes["groupIndex"] = cpuInfo.CoreID / g.CoreGroupSize
		g.addEdge(numaNode, coreGroupNode, "contains")

		// Create CPU Core node
		coreNode := g.getNode(coreID, "CPUCore")
		coreNode.Attributes["status"] = "free"
		g.addEdge(coreGroupNode, coreNode, "contains")
	}
}

// NewGPUNode creates a new GPU node
func (g *FlexTopoGraph) NewGPUNode(index int, uuid, name string, memoryTotal int) *Node {
	gpuID := fmt.Sprintf("gpu-%d", index)
	gpuNode := &Node{
		ID:   gpuID,
		Type: "GPU",
		Attributes: map[string]interface{}{
			"uuid":        uuid,
			"name":        name,
			"memoryTotal": memoryTotal,
			"status":      "free",
		},
	}
	return gpuNode
}

// AddNode adds a node to the graph
func (g *FlexTopoGraph) AddNode(node *Node) {
	g.Nodes[node.ID] = node
}

// UpdateCPUUsage updates the usage status of CPU Core nodes
func (g *FlexTopoGraph) UpdateCPUUsage(podName string, cpuCores []int) {
	for _, coreID := range cpuCores {
		nodeID := fmt.Sprintf("core-%d", coreID)
		if node, exists := g.Nodes[nodeID]; exists {
			node.Attributes["status"] = "used"
			node.Attributes["usedBy"] = podName
		}
	}
}

// UpdateGPUUsage updates the usage status of GPU nodes
func (g *FlexTopoGraph) UpdateGPUUsage(podName string, gpuUUIDs []string) {
	for _, uuid := range gpuUUIDs {
		// Find the corresponding GPU node
		for _, node := range g.Nodes {
			if node.Type == "GPU" {
				if node.Attributes["uuid"] == uuid {
					node.Attributes["status"] = "used"
					node.Attributes["usedBy"] = podName
					break
				}
			}
		}
	}
}

// addEdge adds an edge to the graph and maintains the Children field of nodes
func (g *FlexTopoGraph) addEdge(source, target *Node, edgeType string) {
	edgeKey := fmt.Sprintf("%s-%s-%s", source.ID, target.ID, edgeType)
	fmt.Println("edgeKey:", edgeKey)
	if _, exists := g.Edges[edgeKey]; !exists {
		edge := &Edge{
			Source: source,
			Target: target,
			Type:   edgeType,
		}
		g.Edges[edgeKey] = edge

		// Maintain the Children field
		source.Children = append(source.Children, target)
	}
}

// ToSpec converts FlexTopoGraph to FlexTopoSpec
func (g *FlexTopoGraph) ToSpec() *crd.FlexTopoSpec {
	spec := &crd.FlexTopoSpec{
		Nodes: []crd.FlexTopoNode{},
		Edges: []crd.FlexTopoEdge{},
	}

	// Convert nodes
	for _, node := range g.Nodes {
		specNode := crd.FlexTopoNode{
			ID:         node.ID,
			Type:       node.Type,
			Attributes: node.Attributes,
		}
		spec.Nodes = append(spec.Nodes, specNode)
	}

	// Convert edges
	for _, edge := range g.Edges {
		specEdge := crd.FlexTopoEdge{
			Source: edge.Source.ID,
			Target: edge.Target.ID,
			Type:   edge.Type,
		}
		spec.Edges = append(spec.Edges, specEdge)
	}

	return spec
}

// getNode gets or creates a node
func (g *FlexTopoGraph) getNode(id, nodeType string) *Node {
	if node, exists := g.Nodes[id]; exists {
		return node
	}
	node := &Node{
		ID:         id,
		Type:       nodeType,
		Attributes: make(map[string]interface{}),
		Children:   []*Node{},
	}
	g.Nodes[id] = node
	return node
}

// Helper function to get all nodes of a specific type
func (g *FlexTopoGraph) getNodesByType(nodeType string) []*Node {
	var nodes []*Node
	for _, node := range g.Nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// Helper function to get edges of a specific type originating from a source node
func (g *FlexTopoGraph) getEdges(source *Node, edgeType string) []*Edge {
	var edges []*Edge
	for _, edge := range g.Edges {
		if edge.Source == source && edge.Type == edgeType {
			edges = append(edges, edge)
		}
	}
	return edges
}
