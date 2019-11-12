package types

import (
	//"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	//"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	//"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"github.com/golang/protobuf/ptypes/timestamp"
	"strconv"
)

type Container struct {
	Name        string
	PodName     string
	Namespace   string
	NodeName    string
	ClusterName string
	Uid         string
	Resources   *ResourceRequirements
	Status      *ContainerStatus
}

type ContainerObjectMeta struct {
	ObjectMeta metadata.ObjectMeta
	PodName    string
}

type ListContainersRequest struct {
	common.QueryCondition
	ContainerObjectMeta []ContainerObjectMeta
}

type ResourceRequirements struct {
	// limits describes the maximum amount of compute resources allowed
	// use enum ResourceName as key of the map which defined in common
	Limits map[int32]string
	// requests describes the minimum amount of compute resources required
	// use enum ResourceName as key of the map which defined in common
	Requests map[int32]string
}

type ContainerStatus struct {
	State                *ContainerState
	LastTerminationState *ContainerState
	RestartCount         int32
}

type ContainerState struct {
	Waiting    *ContainerStateWaiting
	Running    *ContainerStateRunning
	Terminated *ContainerStateTerminated
}

type ContainerStateWaiting struct {
	Reason  string
	Message string
}

type ContainerStateRunning struct {
	StartedAt *timestamp.Timestamp
}

type ContainerStateTerminated struct {
	ExitCode   int32
	Reason     string
	Message    string
	StartedAt  *timestamp.Timestamp
	FinishedAt *timestamp.Timestamp
}

func NewContainer() *Container {
	container := Container{}
	container.Resources = &ResourceRequirements{}
	container.Resources.Limits = make(map[int32]string)
	container.Resources.Requests = make(map[int32]string)
	container.Status = &ContainerStatus{}

	return &container
}

func NewListContainersRequest() ListContainersRequest {
	request := ListContainersRequest{}
	request.ContainerObjectMeta = make([]ContainerObjectMeta, 0)
	return request
}

func (p *Container) Initialize(values map[string]string) {
	if value, ok := values[string(clusterstatus.ContainerName)]; ok {
		p.Name = value
	}
	if value, ok := values[string(clusterstatus.ContainerPodName)]; ok {
		p.PodName = value
	}
	if value, ok := values[string(clusterstatus.ContainerNamespace)]; ok {
		p.Namespace = value
	}
	if value, ok := values[string(clusterstatus.ContainerNodeName)]; ok {
		p.NodeName = value
	}
	if value, ok := values[string(clusterstatus.ContainerClusterName)]; ok {
		p.ClusterName = value
	}
	if value, ok := values[string(clusterstatus.ContainerUid)]; ok {
		p.Uid = value
	}

	if value, ok := values[string(clusterstatus.ContainerResourceLimitCPU)]; ok {
		if value != "" {
			p.Resources.Limits[int32(ApiCommon.ResourceName_CPU)] = value
		}
	}
	if value, ok := values[string(clusterstatus.ContainerResourceLimitMemory)]; ok {
		if value != "" {
			p.Resources.Limits[int32(ApiCommon.ResourceName_CPU)] = value
		}
	}
	if value, ok := values[string(clusterstatus.ContainerResourceRequestCPU)]; ok {
		if value != "" {
			p.Resources.Requests[int32(ApiCommon.ResourceName_CPU)] = value
		}
	}
	if value, ok := values[string(clusterstatus.ContainerResourceRequestMemory)]; ok {
		if value != "" {
			p.Resources.Requests[int32(ApiCommon.ResourceName_MEMORY)] = value
		}
	}

	// TODO: remove empty state !!
	p.Status = &ContainerStatus{}

	p.Status.State = &ContainerState{}
	p.Status.State.Waiting = &ContainerStateWaiting{}
	p.Status.State.Running = &ContainerStateRunning{}
	p.Status.State.Terminated = &ContainerStateTerminated{}

	p.Status.LastTerminationState = &ContainerState{}
	p.Status.LastTerminationState.Waiting = &ContainerStateWaiting{}
	p.Status.LastTerminationState.Running = &ContainerStateRunning{}
	p.Status.LastTerminationState.Terminated = &ContainerStateTerminated{}

	if value, ok := values[string(clusterstatus.ContainerStatusWaitingReason)]; ok {
		p.Status.State.Waiting.Reason = value
	}
	if value, ok := values[string(clusterstatus.ContainerStatusWaitingMessage)]; ok {
		p.Status.State.Waiting.Message = value
	}

	if value, ok := values[string(clusterstatus.ContainerStatusRunningStartedAt)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.State.Running.StartedAt = &timestamp.Timestamp{Seconds: valueInt64}
	}

	if value, ok := values[string(clusterstatus.ContainerStatusTerminatedExitCode)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.State.Terminated.ExitCode = int32(valueInt64)
	}
	if value, ok := values[string(clusterstatus.ContainerStatusTerminatedReason)]; ok {
		p.Status.State.Terminated.Reason = value
	}
	if value, ok := values[string(clusterstatus.ContainerStatusTerminatedMessage)]; ok {
		p.Status.State.Terminated.Message = value
	}
	if value, ok := values[string(clusterstatus.ContainerStatusTerminatedStartedAt)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.State.Terminated.StartedAt = &timestamp.Timestamp{Seconds: valueInt64}
	}
	if value, ok := values[string(clusterstatus.ContainerStatusTerminatedFinishedAt)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.State.Terminated.FinishedAt = &timestamp.Timestamp{Seconds: valueInt64}
	}

	if value, ok := values[string(clusterstatus.ContainerLastTerminationWaitingReason)]; ok {
		p.Status.LastTerminationState.Waiting.Reason = value
	}
	if value, ok := values[string(clusterstatus.ContainerLastTerminationWaitingMessage)]; ok {
		p.Status.LastTerminationState.Waiting.Message = value
	}

	if value, ok := values[string(clusterstatus.ContainerLastTerminationRunningStartedAt)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.LastTerminationState.Running.StartedAt = &timestamp.Timestamp{Seconds: valueInt64}
	}

	if value, ok := values[string(clusterstatus.ContainerLastTerminationTerminatedExitCode)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.LastTerminationState.Terminated.ExitCode = int32(valueInt64)
	}
	if value, ok := values[string(clusterstatus.ContainerLastTerminationTerminatedReason)]; ok {
		p.Status.LastTerminationState.Terminated.Reason = value
	}
	if value, ok := values[string(clusterstatus.ContainerLastTerminationTerminatedMessage)]; ok {
		p.Status.LastTerminationState.Terminated.Message = value
	}
	if value, ok := values[string(clusterstatus.ContainerLastTerminationTerminatedStartedAt)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.LastTerminationState.Terminated.StartedAt = &timestamp.Timestamp{Seconds: valueInt64}
	}
	if value, ok := values[string(clusterstatus.ContainerLastTerminationTerminatedFinishedAt)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.LastTerminationState.Terminated.FinishedAt = &timestamp.Timestamp{Seconds: valueInt64}
	}

	if value, ok := values[string(clusterstatus.ContainerRestartCount)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Status.RestartCount = int32(valueInt64)
	}

}

func (p *Container) initStatusWaiting(values map[string]string) {
	if value, ok := values[string(clusterstatus.ContainerStatusWaitingReason)]; ok {
		if value != "" {
			if p.Status == nil {
				p.Status = &ContainerStatus{}
			}
			if p.Status.State == nil {
				p.Status.State = &ContainerState{}
			}
			if p.Status.State.Waiting == nil {
				p.Status.State.Waiting = &ContainerStateWaiting{}
			}
		}
	}
	if value, ok := values[string(clusterstatus.ContainerStatusWaitingMessage)]; ok {
		if value != "" {
			if p.Status == nil {
				p.Status = &ContainerStatus{}
			}
			if p.Status.State == nil {
				p.Status.State = &ContainerState{}
			}
			if p.Status.State.Waiting == nil {
				p.Status.State.Waiting = &ContainerStateWaiting{}
			}
		}
	}
}

func (p *Container) initStatusRunning(values map[string]string) {
	if value, ok := values[string(clusterstatus.ContainerStatusRunningStartedAt)]; ok {
		if value != "" {
			if p.Status == nil {
				p.Status = &ContainerStatus{}
			}
			if p.Status.State == nil {
				p.Status.State = &ContainerState{}
			}
			if p.Status.State.Running == nil {
				p.Status.State.Running = &ContainerStateRunning{}
			}
		}
	}
}
