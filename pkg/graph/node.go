package graph

import "encoding/json"

// Node represents a node in the topology graph
type Node struct {
	ID         string
	Type       string
	Attributes map[string]json.RawMessage
	// New field
	Children []*Node // List of child nodes, used to represent hierarchical structure
}
