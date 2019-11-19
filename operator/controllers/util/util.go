package util

import (
	"github.com/pkg/errors"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	alamedaScalerNameAnnotationKey = "alamedascalers.autoscaling.containers.ai/name"
)

// SetLastMonitorAlamedaScaler sets the last AlamedaScaler's name into the object's annotation
func SetLastMonitorAlamedaScaler(obj metav1.Object, alamedaScalerName string) {

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[alamedaScalerNameAnnotationKey] = alamedaScalerName

	obj.SetAnnotations(annotations)
}

// GetLastMonitorAlamedaScaler gets the last AlamedaScaler's name from the object's annotation
func GetLastMonitorAlamedaScaler(obj metav1.Object) string {

	annotations := obj.GetAnnotations()
	if annotations == nil {
		return ""
	}

	return annotations[alamedaScalerNameAnnotationKey]
}

// TriggerAlamedaScaler will update the provided AlamedaScaler's CustomResourceVersion to trigger the reconcile process
func TriggerAlamedaScaler(client *utilsresource.UpdateResource, alamedaScaler *autoscalingv1alpha1.AlamedaScaler) error {

	alamedaScaler.SetCustomResourceVersion(alamedaScaler.GenCustomResourceVersion())
	err := client.UpdateAlamedaScaler(alamedaScaler)
	if err != nil {
		return errors.Errorf("Update AlamedaScaler falied: error:%s", err.Error())
	}

	return nil
}
