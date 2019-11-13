package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewContainerStatus(containerStatus *types.ContainerStatus) *resources.ContainerStatus {
	if containerStatus != nil {
		cntStatus := resources.ContainerStatus{}

		if containerStatus.State != nil {
			cntStatus.State = &resources.ContainerState{}
			if containerStatus.State.Waiting != nil {
				cntStatus.State.Waiting = &resources.ContainerStateWaiting{}
				cntStatus.State.Waiting.Reason = containerStatus.State.Waiting.Reason
				cntStatus.State.Waiting.Message = containerStatus.State.Waiting.Message
			}
			if containerStatus.State.Running != nil {
				cntStatus.State.Running = &resources.ContainerStateRunning{}
				cntStatus.State.Running.StartedAt = containerStatus.State.Running.StartedAt
			}
			if containerStatus.State.Terminated != nil {
				cntStatus.State.Terminated = &resources.ContainerStateTerminated{}
				cntStatus.State.Terminated.ExitCode = containerStatus.State.Terminated.ExitCode
				cntStatus.State.Terminated.Reason = containerStatus.State.Terminated.Reason
				cntStatus.State.Terminated.Message = containerStatus.State.Terminated.Message
				cntStatus.State.Terminated.StartedAt = containerStatus.State.Terminated.StartedAt
				cntStatus.State.Terminated.FinishedAt = containerStatus.State.Terminated.FinishedAt
			}
		}

		if containerStatus.LastTerminationState != nil {
			cntStatus.LastTerminationState = &resources.ContainerState{}
			if containerStatus.LastTerminationState.Waiting != nil {
				cntStatus.LastTerminationState.Waiting = &resources.ContainerStateWaiting{}
				cntStatus.LastTerminationState.Waiting.Reason = containerStatus.LastTerminationState.Waiting.Reason
				cntStatus.LastTerminationState.Waiting.Message = containerStatus.LastTerminationState.Waiting.Message
			}
			if containerStatus.LastTerminationState.Running != nil {
				cntStatus.LastTerminationState.Running = &resources.ContainerStateRunning{}
				cntStatus.LastTerminationState.Running.StartedAt = containerStatus.LastTerminationState.Running.StartedAt
			}
			if containerStatus.LastTerminationState.Terminated != nil {
				cntStatus.LastTerminationState.Terminated = &resources.ContainerStateTerminated{}
				cntStatus.LastTerminationState.Terminated.ExitCode = containerStatus.LastTerminationState.Terminated.ExitCode
				cntStatus.LastTerminationState.Terminated.Reason = containerStatus.LastTerminationState.Terminated.Reason
				cntStatus.LastTerminationState.Terminated.Message = containerStatus.LastTerminationState.Terminated.Message
				cntStatus.LastTerminationState.Terminated.StartedAt = containerStatus.LastTerminationState.Terminated.StartedAt
				cntStatus.LastTerminationState.Terminated.FinishedAt = containerStatus.LastTerminationState.Terminated.FinishedAt
			}
		}

		cntStatus.RestartCount = containerStatus.RestartCount

		return &cntStatus
	}
	return nil
}

func NewPodStatus(podStatus *types.PodStatus) *resources.PodStatus {
	if podStatus != nil {
		status := resources.PodStatus{}
		status.Phase = resources.PodPhase(resources.PodPhase_value[podStatus.Phase])
		status.Message = podStatus.Message
		status.Reason = podStatus.Reason
		return &status
	}
	return nil
}
