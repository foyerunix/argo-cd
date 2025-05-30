package argo

import (
	"os"
	"testing"

	"github.com/argoproj/argo-cd/v3/util/kube"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/argoproj/argo-cd/v3/common"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
)

func TestSetAppInstanceLabel(t *testing.T) {
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)

	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)

	resourceTracking := NewResourceTracking()

	err = resourceTracking.SetAppInstance(&obj, common.LabelKeyAppInstance, "my-app", "", v1alpha1.TrackingMethodLabel, "")
	require.NoError(t, err)
	app := resourceTracking.GetAppName(&obj, common.LabelKeyAppInstance, v1alpha1.TrackingMethodLabel, "")
	assert.Equal(t, "my-app", app)
}

func TestSetAppInstanceAnnotation(t *testing.T) {
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)

	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)

	resourceTracking := NewResourceTracking()

	err = resourceTracking.SetAppInstance(&obj, common.AnnotationKeyAppInstance, "my-app", "", v1alpha1.TrackingMethodAnnotation, "")
	require.NoError(t, err)

	app := resourceTracking.GetAppName(&obj, common.AnnotationKeyAppInstance, v1alpha1.TrackingMethodAnnotation, "")
	assert.Equal(t, "my-app", app)
}

func TestSetAppInstanceAnnotationAndLabel(t *testing.T) {
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)
	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)

	resourceTracking := NewResourceTracking()

	err = resourceTracking.SetAppInstance(&obj, common.LabelKeyAppInstance, "my-app", "", v1alpha1.TrackingMethodAnnotationAndLabel, "")
	require.NoError(t, err)

	app := resourceTracking.GetAppName(&obj, common.LabelKeyAppInstance, v1alpha1.TrackingMethodAnnotationAndLabel, "")
	assert.Equal(t, "my-app", app)
}

func TestSetAppInstanceAnnotationAndLabelLongName(t *testing.T) {
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)
	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)

	resourceTracking := NewResourceTracking()

	err = resourceTracking.SetAppInstance(&obj, common.LabelKeyAppInstance, "my-app-with-an-extremely-long-name-that-is-over-sixty-three-characters", "", v1alpha1.TrackingMethodAnnotationAndLabel, "")
	require.NoError(t, err)

	// the annotation should still work, so the name from GetAppName should not be truncated
	app := resourceTracking.GetAppName(&obj, common.LabelKeyAppInstance, v1alpha1.TrackingMethodAnnotationAndLabel, "")
	assert.Equal(t, "my-app-with-an-extremely-long-name-that-is-over-sixty-three-characters", app)

	// the label should be truncated to 63 characters
	assert.Equal(t, "my-app-with-an-extremely-long-name-that-is-over-sixty-three-cha", obj.GetLabels()[common.LabelKeyAppInstance])
}

func TestSetAppInstanceAnnotationAndLabelLongNameBadEnding(t *testing.T) {
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)
	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)

	resourceTracking := NewResourceTracking()

	err = resourceTracking.SetAppInstance(&obj, common.LabelKeyAppInstance, "the-very-suspicious-name-with-precisely-sixty-three-characters-with-hyphen", "", v1alpha1.TrackingMethodAnnotationAndLabel, "")
	require.NoError(t, err)

	// the annotation should still work, so the name from GetAppName should not be truncated
	app := resourceTracking.GetAppName(&obj, common.LabelKeyAppInstance, v1alpha1.TrackingMethodAnnotationAndLabel, "")
	assert.Equal(t, "the-very-suspicious-name-with-precisely-sixty-three-characters-with-hyphen", app)

	// the label should be truncated to 63 characters, AND the hyphen should be removed
	assert.Equal(t, "the-very-suspicious-name-with-precisely-sixty-three-characters", obj.GetLabels()[common.LabelKeyAppInstance])
}

func TestSetAppInstanceAnnotationAndLabelOutOfBounds(t *testing.T) {
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)
	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)

	resourceTracking := NewResourceTracking()

	err = resourceTracking.SetAppInstance(&obj, common.LabelKeyAppInstance, "----------------------------------------------------------------", "", v1alpha1.TrackingMethodAnnotationAndLabel, "")
	// this should error because it can't truncate to a valid value
	assert.EqualError(t, err, "failed to set app instance label: unable to truncate label to not end with a special character")
}

func TestSetAppInstanceAnnotationNotFound(t *testing.T) {
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)

	var obj unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)

	resourceTracking := NewResourceTracking()

	app := resourceTracking.GetAppName(&obj, common.LabelKeyAppInstance, v1alpha1.TrackingMethodAnnotation, "")
	assert.Empty(t, app)
}

func TestParseAppInstanceValue(t *testing.T) {
	resourceTracking := NewResourceTracking()
	appInstanceValue, err := resourceTracking.ParseAppInstanceValue("app:<group>/<kind>:<namespace>/<name>")
	require.NoError(t, err)
	assert.Equal(t, "app", appInstanceValue.ApplicationName)
	assert.Equal(t, "<group>", appInstanceValue.Group)
	assert.Equal(t, "<kind>", appInstanceValue.Kind)
	assert.Equal(t, "<namespace>", appInstanceValue.Namespace)
	assert.Equal(t, "<name>", appInstanceValue.Name)
}

func TestParseAppInstanceValueColon(t *testing.T) {
	resourceTracking := NewResourceTracking()
	appInstanceValue, err := resourceTracking.ParseAppInstanceValue("app:<group>/<kind>:<namespace>/<name>:<colon>")
	require.NoError(t, err)
	assert.Equal(t, "app", appInstanceValue.ApplicationName)
	assert.Equal(t, "<group>", appInstanceValue.Group)
	assert.Equal(t, "<kind>", appInstanceValue.Kind)
	assert.Equal(t, "<namespace>", appInstanceValue.Namespace)
	assert.Equal(t, "<name>:<colon>", appInstanceValue.Name)
}

func TestParseAppInstanceValueWrongFormat1(t *testing.T) {
	resourceTracking := NewResourceTracking()
	_, err := resourceTracking.ParseAppInstanceValue("app")
	require.ErrorIs(t, err, ErrWrongResourceTrackingFormat)
}

func TestParseAppInstanceValueWrongFormat2(t *testing.T) {
	resourceTracking := NewResourceTracking()
	_, err := resourceTracking.ParseAppInstanceValue("app;group/kind/ns")
	require.ErrorIs(t, err, ErrWrongResourceTrackingFormat)
}

func TestParseAppInstanceValueCorrectFormat(t *testing.T) {
	resourceTracking := NewResourceTracking()
	_, err := resourceTracking.ParseAppInstanceValue("app:group/kind:test/ns")
	require.NoError(t, err)
}

func sampleResource(t *testing.T) *unstructured.Unstructured {
	t.Helper()
	yamlBytes, err := os.ReadFile("testdata/svc.yaml")
	require.NoError(t, err)
	var obj *unstructured.Unstructured
	err = yaml.Unmarshal(yamlBytes, &obj)
	require.NoError(t, err)
	return obj
}

func TestResourceIdNormalizer_Normalize(t *testing.T) {
	rt := NewResourceTracking()

	// live object is a resource that has old style tracking label
	liveObj := sampleResource(t)
	err := rt.SetAppInstance(liveObj, common.LabelKeyAppInstance, "my-app", "", v1alpha1.TrackingMethodLabel, "")
	require.NoError(t, err)

	// config object is a resource that has new style tracking annotation
	configObj := sampleResource(t)
	err = rt.SetAppInstance(configObj, common.AnnotationKeyAppInstance, "my-app2", "", v1alpha1.TrackingMethodAnnotation, "")
	require.NoError(t, err)

	_ = rt.Normalize(configObj, liveObj, common.LabelKeyAppInstance, string(v1alpha1.TrackingMethodAnnotation))

	// the normalization should affect add the new style annotation and drop old tracking label from live object
	annotation, err := kube.GetAppInstanceAnnotation(configObj, common.AnnotationKeyAppInstance)
	require.NoError(t, err)
	assert.Equal(t, liveObj.GetAnnotations()[common.AnnotationKeyAppInstance], annotation)
	_, hasOldLabel := liveObj.GetLabels()[common.LabelKeyAppInstance]
	assert.False(t, hasOldLabel)
}

func TestResourceIdNormalizer_NormalizeCRD(t *testing.T) {
	rt := NewResourceTracking()

	// live object is a CRD resource
	liveObj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]any{
				"name": "crontabs.stable.example.com",
				"labels": map[string]any{
					common.LabelKeyAppInstance: "my-app",
				},
			},
			"spec": map[string]any{
				"group": "stable.example.com",
				"scope": "Namespaced",
			},
		},
	}

	// config object is a CRD resource
	configObj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]any{
				"name": "crontabs.stable.example.com",
				"labels": map[string]any{
					common.LabelKeyAppInstance: "my-app",
				},
			},
			"spec": map[string]any{
				"group": "stable.example.com",
				"scope": "Namespaced",
			},
		},
	}

	require.NoError(t, rt.Normalize(configObj, liveObj, common.LabelKeyAppInstance, string(v1alpha1.TrackingMethodAnnotation)))
	// the normalization should not apply any changes to the live object
	require.NotContains(t, liveObj.GetAnnotations(), common.AnnotationKeyAppInstance)
}

func TestResourceIdNormalizer_Normalize_ConfigHasOldLabel(t *testing.T) {
	rt := NewResourceTracking()

	// live object is a resource that has old style tracking label
	liveObj := sampleResource(t)
	err := rt.SetAppInstance(liveObj, common.LabelKeyAppInstance, "my-app", "", v1alpha1.TrackingMethodLabel, "")
	require.NoError(t, err)

	// config object is a resource that has new style tracking annotation
	configObj := sampleResource(t)
	err = rt.SetAppInstance(configObj, common.AnnotationKeyAppInstance, "my-app2", "", v1alpha1.TrackingMethodAnnotation, "")
	require.NoError(t, err)
	err = rt.SetAppInstance(configObj, common.LabelKeyAppInstance, "my-app", "", v1alpha1.TrackingMethodLabel, "")
	require.NoError(t, err)

	_ = rt.Normalize(configObj, liveObj, common.LabelKeyAppInstance, string(v1alpha1.TrackingMethodAnnotation))

	// the normalization should affect add the new style annotation and drop old tracking label from live object
	annotation, err := kube.GetAppInstanceAnnotation(configObj, common.AnnotationKeyAppInstance)
	require.NoError(t, err)
	assert.Equal(t, liveObj.GetAnnotations()[common.AnnotationKeyAppInstance], annotation)
	_, hasOldLabel := liveObj.GetLabels()[common.LabelKeyAppInstance]
	assert.True(t, hasOldLabel)
}

func TestIsOldTrackingMethod(t *testing.T) {
	assert.True(t, IsOldTrackingMethod(string(v1alpha1.TrackingMethodLabel)))
}
