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

func (p *Container) Initialize(entity *clusterstatus.ContainerEntity) {
	p.Name = entity.Name
	p.PodName = entity.PodName
	p.Namespace = entity.Namespace
	p.NodeName = entity.NodeName
	p.ClusterName = entity.ClusterName
	p.Uid = entity.Uid

	// Build Resources
	p.Resources = &ResourceRequirements{}
	p.Resources.Limits = make(map[int32]string)
	p.Resources.Requests = make(map[int32]string)
	if entity.ResourceLimitCPU != "" {
		p.Resources.Limits[int32(ApiCommon.ResourceName_CPU)] = entity.ResourceLimitCPU
	}
	if entity.ResourceLimitMemory != "" {
		p.Resources.Limits[int32(ApiCommon.ResourceName_MEMORY)] = entity.ResourceLimitMemory
	}
	if entity.ResourceRequestCPU != "" {
		p.Resources.Requests[int32(ApiCommon.ResourceName_CPU)] = entity.ResourceRequestCPU
	}
	if entity.ResourceRequestMemory != "" {
		p.Resources.Requests[int32(ApiCommon.ResourceName_MEMORY)] = entity.ResourceRequestMemory
	}

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
