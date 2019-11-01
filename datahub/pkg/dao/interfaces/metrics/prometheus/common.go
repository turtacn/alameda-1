package prometheus

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = log.RegisterScope("dao_metrics_prometheu", "metrics dao implementing with Prometheus", 0)
)

func filterObjectMetaByClusterUID(clusterUID string, objectMetas []metadata.ObjectMeta) []metadata.ObjectMeta {
	newObjectMetas := make([]metadata.ObjectMeta, 0, len(objectMetas))
	for _, objectMeta := range objectMetas {
		if objectMeta.ClusterName == clusterUID {
			newObjectMetas = append(newObjectMetas, objectMeta)
		}
	}
	return newObjectMetas
}
