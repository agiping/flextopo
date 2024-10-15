package reporter

import (
	"flextopo/pkg/graph"
	"testing"

	flextopov1alpha1 "github.com/agiping/flextopo-api/pkg/apis/flextopo/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestReporter_Report_Preparation(t *testing.T) {

	testGraph := graph.NewFlexTopoGraph(1)

	spec := testGraph.ToSpec()

	// Verify the generated FlexTopo object
	expectedFlexTopo := &flextopov1alpha1.FlexTopo{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FlexTopo",
			APIVersion: "flextopo.baichuan-inc.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: *spec,
	}

	// Convert FlexTopo object to unstructured.Unstructured
	unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(expectedFlexTopo)
	assert.NoError(t, err)

	// Verify the converted unstructured object
	assert.Equal(t, "FlexTopo", unstructuredData["kind"])
	assert.Equal(t, "flextopo.baichuan-inc.com/v1alpha1", unstructuredData["apiVersion"])
	assert.Equal(t, "test-node", unstructuredData["metadata"].(map[string]interface{})["name"])
}
