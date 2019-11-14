package requests

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewObjectMeta(objectMeta *resources.ObjectMeta) metadata.ObjectMeta {
	meta := metadata.ObjectMeta{
		Name:        objectMeta.GetName(),
		Namespace:   objectMeta.GetNamespace(),
		NodeName:    objectMeta.GetNodeName(),
		ClusterName: objectMeta.GetClusterName(),
		Uid:         objectMeta.GetUid(),
	}
	return meta
}
