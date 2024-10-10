package reporter

import (
	"context"
	"flextopo/pkg/crd"
	"flextopo/pkg/graph"
	"flextopo/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"
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
		Group:    "flextopo.baichuan-inc.com",
		Version:  "v1alpha1",
		Resource: "flextopos",
	}

	// Convert topology graph to CRD Spec
	spec := graph.ToSpec()

	// Build CRD object
	flextopo := &crd.FlexTopo{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FlexTopo",
			APIVersion: "flextopo.baichuan-inc.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: r.nodeName, // Use node name as the CRD object name
		},
		Spec: *spec,
	}

	// Convert CRD object to unstructured.Unstructured
	unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(flextopo)
	if err != nil {
		return err
	}
	unstructuredObj := &unstructured.Unstructured{Object: unstructuredData}

	// Try to get existing resource
	existing, err := r.dynamicClient.Resource(gvr).Get(context.TODO(), r.nodeName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			r.logger.Error("Failed to get FlexTopo CRD: " + err.Error())
			return err
		}
		// Resource doesn't exist, create a new one
		_, err = r.dynamicClient.Resource(gvr).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			r.logger.Error("Failed to create FlexTopo CRD: " + err.Error())
			return err
		}
		r.logger.Info("Successfully created FlexTopo CRD for node: " + r.nodeName)
	} else {
		// Resource exists, update it
		existing.Object["spec"] = unstructuredObj.Object["spec"]
		_, err = r.dynamicClient.Resource(gvr).Update(context.TODO(), existing, metav1.UpdateOptions{})
		if err != nil {
			r.logger.Error("Failed to update FlexTopo CRD: " + err.Error())
			return err
		}
		r.logger.Info("Successfully updated FlexTopo CRD for node: " + r.nodeName)
	}

	return nil
}
