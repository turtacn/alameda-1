package requests

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewContainerStatus(containerStatus *resources.ContainerStatus) *types.ContainerStatus {
	if containerStatus != nil {
		status := &types.ContainerStatus{}
		if state := containerStatus.GetState(); state != nil {
			status.State = &types.ContainerState{}
			if waiting := state.GetWaiting(); waiting != nil {
				status.State.Waiting = &types.ContainerStateWaiting{}
				status.State.Waiting.Reason = waiting.GetReason()
				status.State.Waiting.Message = waiting.GetMessage()
			}
			if running := containerStatus.GetState().GetRunning(); running != nil {
				status.State.Running = &types.ContainerStateRunning{}
				status.State.Running.StartedAt = running.GetStartedAt()
			}
			if terminated := containerStatus.GetState().GetTerminated(); terminated != nil {
				status.State.Terminated = &types.ContainerStateTerminated{}
				status.State.Terminated.ExitCode = terminated.GetExitCode()
				status.State.Terminated.Reason = terminated.GetReason()
				status.State.Terminated.Message = terminated.GetMessage()
				status.State.Terminated.StartedAt = terminated.GetStartedAt()
				status.State.Terminated.FinishedAt = terminated.GetFinishedAt()

			}
		}

		if state := containerStatus.GetLastTerminationState(); state != nil {
			status.LastTerminationState = &types.ContainerState{}
			if waiting := state.GetWaiting(); waiting != nil {
				status.LastTerminationState.Waiting = &types.ContainerStateWaiting{}
				status.LastTerminationState.Waiting.Reason = waiting.GetReason()
				status.LastTerminationState.Waiting.Message = waiting.GetMessage()
			}
			if running := containerStatus.GetState().GetRunning(); running != nil {
				status.LastTerminationState.Running = &types.ContainerStateRunning{}
				status.LastTerminationState.Running.StartedAt = running.GetStartedAt()
			}
			if terminated := containerStatus.GetState().GetTerminated(); terminated != nil {
				status.LastTerminationState.Terminated = &types.ContainerStateTerminated{}
				status.LastTerminationState.Terminated.ExitCode = terminated.GetExitCode()
				status.LastTerminationState.Terminated.Reason = terminated.GetReason()
				status.LastTerminationState.Terminated.Message = terminated.GetMessage()
				status.LastTerminationState.Terminated.StartedAt = terminated.GetStartedAt()
				status.LastTerminationState.Terminated.FinishedAt = terminated.GetFinishedAt()

			}
		}

		status.RestartCount = containerStatus.GetRestartCount()

		return status
	}
	return nil
}

func NewPodStatus(podStatus *resources.PodStatus) *types.PodStatus {
	if podStatus != nil {
		status := &types.PodStatus{}
		status.Phase = podStatus.GetPhase().String()
		status.Message = podStatus.GetMessage()
		status.Reason = podStatus.GetReason()
		return status
	}
	return nil
}
