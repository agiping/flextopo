package graph

// Edge represents an edge in the FlexTopo graph
type Edge struct {
	Source *Node
	Target *Node
	Type   string
}
