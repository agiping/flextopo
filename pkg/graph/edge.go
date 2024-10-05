package graph

// Edge 表示拓扑图中的一条边
type Edge struct {
	Source *Node
	Target *Node
	Type   string
}
