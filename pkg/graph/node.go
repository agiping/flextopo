package graph

// Node 表示拓扑图中的一个节点
type Node struct {
	ID         string
	Type       string
	Attributes map[string]interface{}
	// 新增字段
	Children []*Node // 子节点列表，用于表示层次结构
}
