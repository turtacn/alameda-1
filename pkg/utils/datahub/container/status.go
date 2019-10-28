package container

import (
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/ptypes/timestamp"
	corev1 "k8s.io/api/core/v1"
)

func NewStatus(containerStatus *corev1.ContainerStatus) *ApiResources.ContainerStatus {
	state := &ApiResources.ContainerState{}
	if containerStatus.State.Running != nil {
		state.Running = &ApiResources.ContainerStateRunning{
			StartedAt: &timestamp.Timestamp{
				Seconds: containerStatus.State.Running.StartedAt.Unix(),
			},
		}
	} else if containerStatus.State.Terminated != nil {
		state.Terminated = &ApiResources.ContainerStateTerminated{
			ExitCode: containerStatus.State.Terminated.ExitCode,
			Reason:   containerStatus.State.Terminated.Reason,
			Message:  containerStatus.State.Terminated.Message,
			StartedAt: &timestamp.Timestamp{
				Seconds: containerStatus.State.Terminated.StartedAt.Unix(),
			},
			FinishedAt: &timestamp.Timestamp{
				Seconds: containerStatus.State.Terminated.FinishedAt.Unix(),
			},
		}
	} else if containerStatus.State.Waiting != nil {
		state.Waiting = &ApiResources.ContainerStateWaiting{
			Reason:  containerStatus.State.Waiting.Reason,
			Message: containerStatus.State.Waiting.Message,
		}
	}
	lastTerminationState := &ApiResources.ContainerState{}
	if containerStatus.LastTerminationState.Running != nil {
		lastTerminationState.Running = &ApiResources.ContainerStateRunning{
			StartedAt: &timestamp.Timestamp{
				Seconds: containerStatus.LastTerminationState.Running.StartedAt.Unix(),
			},
		}
	} else if containerStatus.LastTerminationState.Terminated != nil {
		lastTerminationState.Terminated = &ApiResources.ContainerStateTerminated{
			ExitCode: containerStatus.LastTerminationState.Terminated.ExitCode,
			Reason:   containerStatus.LastTerminationState.Terminated.Reason,
			Message:  containerStatus.LastTerminationState.Terminated.Message,
			StartedAt: &timestamp.Timestamp{
				Seconds: containerStatus.LastTerminationState.Terminated.StartedAt.Unix(),
			},
			FinishedAt: &timestamp.Timestamp{
				Seconds: containerStatus.LastTerminationState.Terminated.FinishedAt.Unix(),
			},
		}
	} else if containerStatus.LastTerminationState.Waiting != nil {
		lastTerminationState.Waiting = &ApiResources.ContainerStateWaiting{
			Reason:  containerStatus.LastTerminationState.Waiting.Reason,
			Message: containerStatus.LastTerminationState.Waiting.Message,
		}
	}
	return &ApiResources.ContainerStatus{
		RestartCount:         containerStatus.RestartCount,
		State:                state,
		LastTerminationState: lastTerminationState,
	}
}
