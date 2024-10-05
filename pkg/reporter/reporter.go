package reporter

import (
	"context"
	"flextopo/pkg/crd"
	"flextopo/pkg/graph"
	"flextopo/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Reporter struct {
	dynamicClient dynamic.Interface
	nodeName      string
	logger        utils.Logger
}

func NewReporter(nodeName string, logger utils.Logger) (*Reporter, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Reporter{
		dynamicClient: dynamicClient,
		nodeName:      nodeName,
		logger:        logger,
	}, nil
}

func (r *Reporter) Report(graph *graph.FlexTopoGraph) error {
	gvr := schema.GroupVersionResource{
		Group:    "flextopo.example.com",
		Version:  "v1alpha1",
		Resource: "flextopos",
	}

	// 将拓扑图转换为 CRD 的 Spec
	spec := graph.ToSpec()

	// 构建 CRD 对象
	flextopo := &crd.FlexTopo{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FlexTopo",
			APIVersion: "flextopo.example.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: r.nodeName, // 使用节点名称作为 CRD 对象的名称
		},
		Spec: *spec,
	}

	// 将 CRD 对象转换为 unstructured.Unstructured
	unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(flextopo)
	if err != nil {
		return err
	}
	unstructuredObj := &unstructured.Unstructured{Object: unstructuredData}

	// 尝试更新 CRD 对象
	_, err = r.dynamicClient.Resource(gvr).Update(context.TODO(), unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		// 如果更新失败，尝试创建
		_, err = r.dynamicClient.Resource(gvr).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			r.logger.Error("Failed to create or update FlexTopo CRD: " + err.Error())
			return err
		}
	}
	r.logger.Info("Successfully reported FlexTopo for node " + r.nodeName)
	return nil
}
