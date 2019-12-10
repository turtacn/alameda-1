package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type Container struct {
	Name                     string
	PodName                  string
	Namespace                string
	NodeName                 string
	ClusterName              string
	Uid                      string
	TopControllerName        string
	TopControllerKind        string
	AlamedaScalerName        string
	AlamedaScalerScalingTool string
	Resources                *ResourceRequirements
	Status                   *ContainerStatus
}

type ContainerObjectMeta struct {
	Name                     string
	PodName                  string
	Namespace                string
	NodeName                 string
	ClusterName              string
	TopControllerName        string
	TopControllerKind        string
	AlamedaScalerName        string
	AlamedaScalerScalingTool string
}

type ListContainersRequest struct {
	common.QueryCondition
	ContainerObjectMeta []*ContainerObjectMeta
}

type DeleteContainersRequest struct {
	ContainerObjectMeta []*ContainerObjectMeta
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
	request.ContainerObjectMeta = make([]*ContainerObjectMeta, 0)
	return request
}

func NewDeleteContainersRequest() DeleteContainersRequest {
	request := DeleteContainersRequest{}
	request.ContainerObjectMeta = make([]*ContainerObjectMeta, 0)
	return request
}

func NewResourceRequirements(limitCpu, limitMemory, reqCpu, reqMemory string) *ResourceRequirements {
	requirements := ResourceRequirements{}
	requirements.Limits = make(map[int32]string)
	requirements.Requests = make(map[int32]string)
	if limitCpu != "" {
		requirements.Limits[int32(ApiCommon.ResourceName_CPU)] = limitCpu
	}
	if limitMemory != "" {
		requirements.Limits[int32(ApiCommon.ResourceName_MEMORY)] = limitMemory
	}
	if reqCpu != "" {
		requirements.Requests[int32(ApiCommon.ResourceName_CPU)] = reqCpu
	}
	if reqMemory != "" {
		requirements.Requests[int32(ApiCommon.ResourceName_MEMORY)] = reqMemory
	}
	return &requirements
}

func (p *Container) BuildEntity() *clusterstatus.ContainerEntity {
	entity := clusterstatus.ContainerEntity{}

	// Build misc information
	entity.Time = influxdb.ZeroTime
	entity.Name = p.Name
	entity.PodName = p.PodName
	entity.Namespace = p.Namespace
	entity.NodeName = p.NodeName
	entity.ClusterName = p.ClusterName
	entity.Uid = p.Uid
	entity.TopControllerName = p.TopControllerName
	entity.TopControllerKind = p.TopControllerKind
	entity.AlamedaScalerName = p.AlamedaScalerName
	entity.AlamedaScalerScalingTool = p.AlamedaScalerScalingTool
	entity.RestartCount = p.Status.RestartCount

	// Build resources
	if p.Resources != nil {
		if value, exist := p.Resources.Limits[int32(ApiCommon.ResourceName_CPU)]; exist {
			entity.ResourceLimitCPU = value
		}
		if value, exist := p.Resources.Limits[int32(ApiCommon.ResourceName_MEMORY)]; exist {
			entity.ResourceLimitMemory = value
		}
		if value, exist := p.Resources.Requests[int32(ApiCommon.ResourceName_CPU)]; exist {
			entity.ResourceRequestCPU = value
		}
		if value, exist := p.Resources.Requests[int32(ApiCommon.ResourceName_MEMORY)]; exist {
			entity.ResourceRequestMemory = value
		}
	}

	// Build status
	if p.Status != nil {
		if p.Status.State != nil {
			if p.Status.State.Waiting != nil {
				entity.StatusWaitingReason = p.Status.State.Waiting.Reason
				entity.StatusWaitingMessage = p.Status.State.Waiting.Message
			}
			if p.Status.State.Running != nil {
				entity.StatusRunningStartedAt = p.Status.State.Running.StartedAt.GetSeconds()
			}
			if p.Status.State.Terminated != nil {
				entity.StatusTerminatedExitCode = p.Status.State.Terminated.ExitCode
				entity.StatusTerminatedReason = p.Status.State.Terminated.Reason
				entity.StatusTerminatedMessage = p.Status.State.Terminated.Message
				entity.StatusTerminatedStartedAt = p.Status.State.Terminated.StartedAt.GetSeconds()
				entity.StatusTerminatedFinishedAt = p.Status.State.Terminated.FinishedAt.GetSeconds()
			}

		}

		if p.Status.LastTerminationState != nil {
			if p.Status.LastTerminationState.Waiting != nil {
				entity.LastTerminationWaitingReason = p.Status.LastTerminationState.Waiting.Reason
				entity.LastTerminationWaitingMessage = p.Status.LastTerminationState.Waiting.Message
			}
			if p.Status.LastTerminationState.Running != nil {
				entity.LastTerminationRunningStartedAt = p.Status.LastTerminationState.Running.StartedAt.GetSeconds()
			}
			if p.Status.LastTerminationState.Terminated != nil {
				entity.LastTerminationTerminatedExitCode = p.Status.LastTerminationState.Terminated.ExitCode
				entity.LastTerminationTerminatedReason = p.Status.LastTerminationState.Terminated.Reason
				entity.LastTerminationTerminatedMessage = p.Status.LastTerminationState.Terminated.Message
				entity.LastTerminationTerminatedStartedAt = p.Status.LastTerminationState.Terminated.StartedAt.GetSeconds()
				entity.LastTerminationTerminatedFinishedAt = p.Status.LastTerminationState.Terminated.FinishedAt.GetSeconds()
			}
		}
	}

	return &entity
}

func (p *Container) Initialize(entity *clusterstatus.ContainerEntity) {
	p.Name = entity.Name
	p.PodName = entity.PodName
	p.Namespace = entity.Namespace
	p.NodeName = entity.NodeName
	p.ClusterName = entity.ClusterName
	p.Uid = entity.Uid

	// Build Resources
	p.Resources = NewResourceRequirements(entity.ResourceLimitCPU, entity.ResourceLimitMemory, entity.ResourceRequestCPU, entity.ResourceRequestMemory)

	// Build Status
	p.Status = &ContainerStatus{}
	p.Status.State = p.newState(entity)
	p.Status.LastTerminationState = p.newLastTerminationState(entity)
	p.Status.RestartCount = entity.RestartCount
}

func (p *Container) newState(entity *clusterstatus.ContainerEntity) *ContainerState {
	waiting := p.waiting(entity.StatusWaitingReason, entity.StatusWaitingMessage)
	running := p.running(entity.StatusRunningStartedAt)
	terminated := p.terminated(
		entity.StatusTerminatedReason,
		entity.StatusTerminatedMessage,
		entity.StatusTerminatedStartedAt,
		entity.StatusTerminatedFinishedAt,
		entity.StatusTerminatedExitCode)
	if waiting != nil || running != nil || terminated != nil {
		state := &ContainerState{}
		state.Waiting = waiting
		state.Running = running
		state.Terminated = terminated
		return state
	}
	return nil
}

func (p *Container) newLastTerminationState(entity *clusterstatus.ContainerEntity) *ContainerState {
	waiting := p.waiting(entity.LastTerminationWaitingReason, entity.LastTerminationWaitingMessage)
	running := p.running(entity.LastTerminationRunningStartedAt)
	terminated := p.terminated(
		entity.LastTerminationTerminatedReason,
		entity.LastTerminationTerminatedMessage,
		entity.LastTerminationTerminatedStartedAt,
		entity.LastTerminationTerminatedFinishedAt,
		entity.LastTerminationTerminatedExitCode)
	if waiting != nil || running != nil || terminated != nil {
		state := &ContainerState{}
		state.Waiting = waiting
		state.Running = running
		state.Terminated = terminated
		return state
	}
	return nil
}

func (p *Container) waiting(reason, message string) *ContainerStateWaiting {
	if reason != "" || message != "" {
		state := &ContainerStateWaiting{}
		state.Reason = reason
		state.Message = message
		return state
	}
	return nil
}

func (p *Container) running(startedAt int64) *ContainerStateRunning {
	if startedAt != 0 {
		state := &ContainerStateRunning{}
		state.StartedAt = &timestamp.Timestamp{Seconds: startedAt}
		return state
	}
	return nil
}

func (p *Container) terminated(reason, message string, startedAt, finishedAt int64, exitCode int32) *ContainerStateTerminated {
	if reason != "" || message != "" || startedAt != 0 || finishedAt != 0 {
		state := &ContainerStateTerminated{}
		state.Reason = reason
		state.Message = message
		state.StartedAt = &timestamp.Timestamp{Seconds: startedAt}
		state.FinishedAt = &timestamp.Timestamp{Seconds: finishedAt}
		state.ExitCode = exitCode
		return state
	}
	return nil
}
