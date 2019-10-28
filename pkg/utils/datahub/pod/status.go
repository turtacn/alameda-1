package pod

import (
	AlamedaUtils "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	AlamedaConsts "github.com/containers-ai/alameda/pkg/consts"
	AlamedaEnum "github.com/containers-ai/alameda/pkg/utils/datahub/enumconv"
	AlamedaLog "github.com/containers-ai/alameda/pkg/utils/log"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	CoreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scope = AlamedaLog.RegisterScope("datahubpodutils", "datahub pod utils", 0)
)

// NewStatus return pod status struct of datahub
func NewStatus(pod *CoreV1.Pod) *ApiResources.PodStatus {
	return &ApiResources.PodStatus{
		Message: pod.Status.Message,
		Reason:  pod.Status.Reason,
		Phase:   AlamedaEnum.PodPhaseEnumK8SToDatahub[pod.Status.Phase],
	}
}

// GetReplicasFromPod return number of replicas of pod
func GetReplicasFromPod(pod *CoreV1.Pod, client client.Client) int32 {
	getResource := AlamedaUtils.NewGetResource(client)

	for _, or := range pod.OwnerReferences {
		if or.Kind == AlamedaConsts.K8S_KIND_REPLICASET {
			rs, err := getResource.GetReplicaSet(pod.GetNamespace(), or.Name)
			if err == nil {
				return rs.Status.Replicas
			} else {
				scope.Errorf("Get replicaset for number of replicas failed due to %s", err.Error())
			}
		} else if or.Kind == AlamedaConsts.K8S_KIND_REPLICATIONCONTROLLER {
			rc, err := getResource.GetReplicationController(pod.GetNamespace(), or.Name)
			if err == nil {
				return rc.Status.Replicas
			} else {
				scope.Errorf("Get replicationcontroller for number of replicas failed due to %s", err.Error())
			}
		} else if or.Kind == AlamedaConsts.K8S_KIND_STATEFULSET {
			sts, err := getResource.GetStatefulSet(pod.GetNamespace(), or.Name)
			if err == nil {
				return sts.Status.Replicas
			} else {
				scope.Errorf("Get StatefulSet for number of replicas failed due to %s", err.Error())
			}
		}
	}
	return int32(-1)
}
