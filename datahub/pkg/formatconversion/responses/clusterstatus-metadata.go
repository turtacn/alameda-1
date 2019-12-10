package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewObjectMeta(objectMeta *metadata.ObjectMeta) *resources.ObjectMeta {
	meta := resources.ObjectMeta{
		Name:        objectMeta.Name,
		Namespace:   objectMeta.Namespace,
		NodeName:    objectMeta.NodeName,
		ClusterName: objectMeta.ClusterName,
		Uid:         objectMeta.Uid,
	}
	return &meta
}
