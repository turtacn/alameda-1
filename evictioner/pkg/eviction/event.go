package eviction

import (
	"fmt"

	datahub_events "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"

	"github.com/golang/protobuf/ptypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	componentName = "alameda-evictioner"

	defaultPodKind       = "Pod"
	defaultPodAPIVersion = "v1"
)

func newPodEvictEvent(clusterID string, subjectObject metav1.Object, subjectType metav1.TypeMeta) datahub_events.Event {

	if subjectType.Kind == "" {
		subjectType.Kind = defaultPodKind
	}
	if subjectType.APIVersion == "" {
		subjectType.APIVersion = defaultPodAPIVersion
	}

	now := ptypes.TimestampNow()
	id := uuid.NewUUID()
	source := datahub_events.EventSource{
		Host:      "",
		Component: componentName,
	}
	eventType := datahub_events.EventType_EVENT_TYPE_VPA_RECOMMENDATION_EXECUTE
	version := datahub_events.EventVersion_EVENT_VERSION_V1
	level := datahub_events.EventLevel_EVENT_LEVEL_INFO
	subject := datahub_events.K8SObjectReference{
		Kind:       subjectType.Kind,
		ApiVersion: subjectType.APIVersion,
		Namespace:  subjectObject.GetNamespace(),
		Name:       subjectObject.GetName(),
	}
	message := fmt.Sprintf("Pod %s/%s is evicted", subjectObject.GetNamespace(), subjectObject.GetName())
	data := ""

	event := datahub_events.Event{
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
