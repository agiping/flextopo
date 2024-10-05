package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FlexTopo 定义了 CRD 的结构
type FlexTopo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlexTopoSpec   `json:"spec,omitempty"`
	Status FlexTopoStatus `json:"status,omitempty"`
}

// FlexTopoSpec 定义了拓扑图的内容
type FlexTopoSpec struct {
	Nodes []FlexTopoNode `json:"nodes"`
	Edges []FlexTopoEdge `json:"edges"`
}

// FlexTopoNode 表示拓扑图中的一个节点
type FlexTopoNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
}

// FlexTopoEdge 表示拓扑图中的一条边
type FlexTopoEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

// FlexTopoStatus 可用于存储状态信息
type FlexTopoStatus struct {
	// 可根据需要添加状态字段
}
