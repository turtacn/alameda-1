package server

import (
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"

	"github.com/golang/protobuf/ptypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	componentName = "admission-controller"
)

func newPodPatchEvent(namespace, clusterID string, ownRef metav1.OwnerReference) datahub_v1alpha1.Event {

	now := ptypes.TimestampNow()
	id := uuid.NewUUID()
	source := datahub_v1alpha1.EventSource{
		Host:      "",
		Component: componentName,
	}
	eventType := datahub_v1alpha1.EventType_EVENT_TYPE_VPA_RECOMMENDATION_EXECUTE
	version := datahub_v1alpha1.EventVersion_EVENT_VERSION_V1
	level := datahub_v1alpha1.EventLevel_EVENT_LEVEL_INFO
	subject := datahub_v1alpha1.K8SObjectReference{
		Kind:       ownRef.Kind,
		ApiVersion: ownRef.APIVersion,
		Namespace:  namespace,
		Name:       ownRef.Name,
	}
	message := "Patch resource recommendation to new created pod"
	data := ""

	event := datahub_v1alpha1.Event{
		Time:      now,
		Id:        string(id),
		ClusterId: clusterID,
		Source:    &source,
		Type:      eventType,
		Version:   version,
		Level:     level,
		Subject:   &subject,
		Message:   message,
		Data:      data,
	}

	return event
}
