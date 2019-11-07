package enumconv

import (
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	CoreV1 "k8s.io/api/core/v1"
)

var PodPhaseEnumDatahubToK8S map[ApiResources.PodPhase]CoreV1.PodPhase = map[ApiResources.PodPhase]CoreV1.PodPhase{
	ApiResources.PodPhase_PENDING:   CoreV1.PodPending,
	ApiResources.PodPhase_RUNNING:   CoreV1.PodRunning,
	ApiResources.PodPhase_SUCCEEDED: CoreV1.PodSucceeded,
	ApiResources.PodPhase_FAILED:    CoreV1.PodFailed,
	ApiResources.PodPhase_UNKNOWN:   CoreV1.PodUnknown,
}

var PodPhaseEnumK8SToDatahub map[CoreV1.PodPhase]ApiResources.PodPhase = map[CoreV1.PodPhase]ApiResources.PodPhase{
	CoreV1.PodPending:   ApiResources.PodPhase_PENDING,
	CoreV1.PodRunning:   ApiResources.PodPhase_RUNNING,
	CoreV1.PodSucceeded: ApiResources.PodPhase_SUCCEEDED,
	CoreV1.PodFailed:    ApiResources.PodPhase_FAILED,
	CoreV1.PodUnknown:   ApiResources.PodPhase_UNKNOWN,
}
