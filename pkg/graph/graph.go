package graph

import (
	"flextopo/pkg/crd"
	"flextopo/pkg/utils"
	"fmt"
	"math"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// FlexTopoGraph 表示整个拓扑图
type FlexTopoGraph struct {
	Nodes map[string]*Node
	Edges []*Edge
	// 新增字段
	CoreGroupSize int // CPU 核心分组大小，可配置
}

// NewFlexTopoGraph 创建一个新的 FlexTopoGraph 实例
func NewFlexTopoGraph(coreGroupSize int) *FlexTopoGraph {
	return &FlexTopoGraph{
		Nodes:         make(map[string]*Node),
		Edges:         []*Edge{},
		CoreGroupSize: coreGroupSize,
	}
}

// BuildCPUNodes 根据 CPU 信息构建节点和边
func (g *FlexTopoGraph) BuildCPUNodes(cpuInfos []utils.CPUInfo) {
	// 按照 Socket、NUMA、CoreID 对 CPUInfo 进行排序
	sort.Slice(cpuInfos, func(i, j int) bool {
		if cpuInfos[i].SocketID != cpuInfos[j].SocketID {
			return cpuInfos[i].SocketID < cpuInfos[j].SocketID
		}
		if cpuInfos[i].NumaNodeID != cpuInfos[j].NumaNodeID {
			return cpuInfos[i].NumaNodeID < cpuInfos[j].NumaNodeID
		}
		return cpuInfos[i].CoreID < cpuInfos[j].CoreID
	})

	// 构建节点
	for _, cpuInfo := range cpuInfos {
		socketID := fmt.Sprintf("socket-%d", cpuInfo.SocketID)
		numaNodeID := fmt.Sprintf("numa-%d", cpuInfo.NumaNodeID)
		coreID := fmt.Sprintf("core-%d", cpuInfo.CoreID)

		// 创建或获取 Socket 节点
		socketNode := g.getNode(socketID, "Socket")

		// 创建或获取 NUMA Node 节点
		numaNode := g.getNode(numaNodeID, "NUMANode")
		g.addEdge(socketNode, numaNode, "contains")

		// 创建或获取 Core Group 节点
		coreGroupID := fmt.Sprintf("coregroup-%d-%d", cpuInfo.NumaNodeID, cpuInfo.CoreID/g.CoreGroupSize)
		coreGroupNode := g.getNode(coreGroupID, "CoreGroup")
		coreGroupNode.Attributes["nodeID"] = cpuInfo.NumaNodeID
		coreGroupNode.Attributes["groupIndex"] = cpuInfo.CoreID / g.CoreGroupSize
		g.addEdge(numaNode, coreGroupNode, "contains")

		// 创建 CPU Core 节点
		coreNode := g.getNode(coreID, "CPUCore")
		coreNode.Attributes["status"] = "free"
		g.addEdge(coreGroupNode, coreNode, "contains")
	}
}

// NewGPUNode 创建一个新的 GPU 节点
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

// AddNode 将节点添加到图中
func (g *FlexTopoGraph) AddNode(node *Node) {
	g.Nodes[node.ID] = node
}

// UpdateCPUAllocation 更新 CPU 核心的分配状态
func (g *FlexTopoGraph) UpdateCPUAllocation(pod *corev1.Pod, cpuQuantity *resource.Quantity) {
	// 计算需要分配的核心数
	cpuCoresNeeded := int(math.Ceil(float64(cpuQuantity.MilliValue()) / 1000))

	// 查找可用的 Core Group
	availableCoreGroups := []*Node{}
	for _, node := range g.Nodes {
		if node.Type == "CoreGroup" {
			// 检查 Core Group 是否可用
			isAvailable := true
			for _, child := range node.Children {
				status, ok := child.Attributes["status"].(string)
				if !ok || status != "free" {
					isAvailable = false
					break
				}
			}
			if isAvailable {
				availableCoreGroups = append(availableCoreGroups, node)
			}
		}
	}

	// 分配核心
	coresAllocated := 0
	for _, coreGroup := range availableCoreGroups {
		for _, coreNode := range coreGroup.Children {
			if coresAllocated >= cpuCoresNeeded {
				break
			}
			coreNode.Attributes["status"] = "allocated"
			coreNode.Attributes["allocatedTo"] = pod.Name
			coresAllocated++
		}
		if coresAllocated >= cpuCoresNeeded {
			break
		}
	}

	if coresAllocated < cpuCoresNeeded {
		// 资源不足，记录警告或错误
		// 可以在这里实现资源不足的处理逻辑
	}
}

// UpdateGPUAllocation 更新 GPU 的分配状态
func (g *FlexTopoGraph) UpdateGPUAllocation(pod *corev1.Pod, gpuQuantity *resource.Quantity) {
	// 计算需要分配的 GPU 数量
	gpusNeeded := int(gpuQuantity.Value())

	// 查找可用的 GPU 节点
	availableGPUs := []*Node{}
	for _, node := range g.Nodes {
		if node.Type == "GPU" {
			status, ok := node.Attributes["status"].(string)
			if ok && status == "free" {
				availableGPUs = append(availableGPUs, node)
			}
		}
	}

	// 分配 GPU
	gpusAllocated := 0
	for _, gpuNode := range availableGPUs {
		if gpusAllocated >= gpusNeeded {
			break
		}
		gpuNode.Attributes["status"] = "allocated"
		gpuNode.Attributes["allocatedTo"] = pod.Name
		gpusAllocated++
	}

	if gpusAllocated < gpusNeeded {
		// 资源不足，记录警告或错误
		// 可以在这里实现资源不足的处理逻辑
	}
}

// getNode 获取或创建节点
func (g *FlexTopoGraph) getNode(id, nodeType string) *Node {
	if node, exists := g.Nodes[id]; exists {
		return node
	}
	node := &Node{
		ID:         id,
		Type:       nodeType,
		Attributes: make(map[string]interface{}),
	}
	g.Nodes[id] = node
	return node
}

// addEdge 添加边到图中，并维护节点的 Children
func (g *FlexTopoGraph) addEdge(source, target *Node, edgeType string) {
	edge := &Edge{
		Source: source,
		Target: target,
		Type:   edgeType,
	}
	g.Edges = append(g.Edges, edge)

	// 维护 Children 字段
	source.Children = append(source.Children, target)
}

// ToSpec 将 FlexTopoGraph 转换为 FlexTopoSpec
func (g *FlexTopoGraph) ToSpec() *crd.FlexTopoSpec {
	spec := &crd.FlexTopoSpec{
		Nodes: []crd.FlexTopoNode{},
		Edges: []crd.FlexTopoEdge{},
	}

	// 转换节点
	for _, node := range g.Nodes {
		specNode := crd.FlexTopoNode{
			ID:         node.ID,
			Type:       node.Type,
			Attributes: node.Attributes,
		}
		spec.Nodes = append(spec.Nodes, specNode)
	}

	// 转换边
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
