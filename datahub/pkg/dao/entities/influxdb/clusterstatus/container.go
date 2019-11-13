package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	//ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"time"
)

const (
	ContainerTime        influxdb.Tag = "time"
	ContainerName        influxdb.Tag = "name"
	ContainerPodName     influxdb.Tag = "pod_name"
	ContainerNamespace   influxdb.Tag = "namespace"
	ContainerNodeName    influxdb.Tag = "node_name"
	ContainerClusterName influxdb.Tag = "cluster_name"
	ContainerUid         influxdb.Tag = "uid"

	ContainerResourceRequestCPU                  influxdb.Field = "resource_request_cpu"
	ContainerResourceRequestMemory               influxdb.Field = "resource_request_memory"
	ContainerResourceLimitCPU                    influxdb.Field = "resource_limit_cpu"
	ContainerResourceLimitMemory                 influxdb.Field = "resource_limit_memory"
	ContainerStatusWaitingReason                 influxdb.Field = "status_waiting_reason"
	ContainerStatusWaitingMessage                influxdb.Field = "status_waiting_message"
	ContainerStatusRunningStartedAt              influxdb.Field = "status_running_start_at"
	ContainerStatusTerminatedExitCode            influxdb.Field = "status_terminated_exit_code"
	ContainerStatusTerminatedReason              influxdb.Field = "status_terminated_reason"
	ContainerStatusTerminatedMessage             influxdb.Field = "status_terminated_message"
	ContainerStatusTerminatedStartedAt           influxdb.Field = "status_terminated_started_at"
	ContainerStatusTerminatedFinishedAt          influxdb.Field = "status_terminated_finished_at"
	ContainerLastTerminationWaitingReason        influxdb.Field = "last_termination_status_waiting_reason"
	ContainerLastTerminationWaitingMessage       influxdb.Field = "last_termination_status_waiting_message"
	ContainerLastTerminationRunningStartedAt     influxdb.Field = "last_termination_status_running_start_at"
	ContainerLastTerminationTerminatedExitCode   influxdb.Field = "last_termination_status_terminated_exit_code"
	ContainerLastTerminationTerminatedReason     influxdb.Field = "last_termination_status_terminated_reason"
	ContainerLastTerminationTerminatedMessage    influxdb.Field = "last_termination_status_terminated_message"
	ContainerLastTerminationTerminatedStartedAt  influxdb.Field = "last_termination_status_terminated_started_at"
	ContainerLastTerminationTerminatedFinishedAt influxdb.Field = "last_termination_status_terminated_finished_at"
	ContainerRestartCount                        influxdb.Field = "restart_count"
)

var (
	// ContainerTags is the list of container measurement tags
	ContainerTags = []influxdb.Tag{
		ContainerTime,
		ContainerName,
		ContainerPodName,
		ContainerNamespace,
		ContainerNodeName,
		ContainerClusterName,
		ContainerUid,
	}

	// ContainerFields is the list of container measurement fields
	ContainerFields = []influxdb.Field{
		ContainerResourceRequestCPU,
		ContainerResourceRequestMemory,
		ContainerResourceLimitCPU,
		ContainerResourceLimitMemory,
		ContainerStatusWaitingReason,
		ContainerStatusWaitingMessage,
		ContainerStatusRunningStartedAt,
		ContainerStatusTerminatedExitCode,
		ContainerStatusTerminatedReason,
		ContainerStatusTerminatedMessage,
		ContainerStatusTerminatedStartedAt,
		ContainerStatusTerminatedFinishedAt,
		ContainerLastTerminationWaitingReason,
		ContainerLastTerminationWaitingMessage,
		ContainerLastTerminationRunningStartedAt,
		ContainerLastTerminationTerminatedExitCode,
		ContainerLastTerminationTerminatedReason,
		ContainerLastTerminationTerminatedMessage,
		ContainerLastTerminationTerminatedStartedAt,
		ContainerLastTerminationTerminatedFinishedAt,
		ContainerRestartCount,
	}
)

// ContainerEntity Entity in database
type ContainerEntity struct {
	Time        time.Time
	Name        string
	PodName     string
	Namespace   string
	NodeName    string
	ClusterName string
	Uid         string

	ResourceRequestCPU                  string // TODO: check if type string or float64
	ResourceRequestMemory               string // TODO: check if type string or float64
	ResourceLimitCPU                    string // TODO: check if type string or float64
	ResourceLimitMemory                 string // TODO: check if type string or float64
	StatusWaitingReason                 string
	StatusWaitingMessage                string
	StatusRunningStartedAt              int64
	StatusTerminatedExitCode            int32
	StatusTerminatedReason              string
	StatusTerminatedMessage             string
	StatusTerminatedStartedAt           int64
	StatusTerminatedFinishedAt          int64
	LastTerminationWaitingReason        string
	LastTerminationWaitingMessage       string
	LastTerminationRunningStartedAt     int64
	LastTerminationTerminatedExitCode   int32
	LastTerminationTerminatedReason     string
	LastTerminationTerminatedMessage    string
	LastTerminationTerminatedStartedAt  int64
	LastTerminationTerminatedFinishedAt int64
	RestartCount                        int32
}

// NewContainerEntityFromMap Build entity from map
func NewContainerEntity(data map[string]string) *ContainerEntity {
	entity := ContainerEntity{}

	tempTimestamp, _ := utils.ParseTime(data[string(ContainerTime)])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(ContainerName)]; exist {
		entity.Name = value
	}
	if value, exist := data[string(ContainerPodName)]; exist {
		entity.PodName = value
	}
	if value, exist := data[string(ContainerNamespace)]; exist {
		entity.Namespace = value
	}
	if value, exist := data[string(ContainerNodeName)]; exist {
		entity.NodeName = value
	}
	if value, exist := data[string(ContainerClusterName)]; exist {
		entity.ClusterName = value
	}
	if value, exist := data[string(ContainerUid)]; exist {
		entity.Uid = value
	}

	// InfluxDB fields
	if value, exist := data[string(ContainerResourceRequestCPU)]; exist {
		if data[string(ContainerResourceRequestCPU)] != "" {
			entity.ResourceRequestCPU = value
		}
	}
	if value, exist := data[string(ContainerResourceRequestMemory)]; exist {
		if data[string(ContainerResourceRequestMemory)] != "" {
			entity.ResourceRequestMemory = value
		}
	}
	if value, exist := data[string(ContainerResourceLimitCPU)]; exist {
		if data[string(ContainerResourceLimitCPU)] != "" {
			entity.ResourceLimitCPU = value
		}
	}
	if value, exist := data[string(ContainerResourceLimitMemory)]; exist {
		if data[string(ContainerResourceLimitMemory)] != "" {
			entity.ResourceLimitMemory = value
		}
	}
	if value, exist := data[string(ContainerStatusWaitingReason)]; exist {
		entity.StatusWaitingReason = value
	}
	if value, exist := data[string(ContainerStatusWaitingMessage)]; exist {
		entity.StatusWaitingMessage = value
	}
	if value, exist := data[string(ContainerStatusRunningStartedAt)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.StatusRunningStartedAt = valueInt64
	}
	if value, exist := data[string(ContainerStatusTerminatedExitCode)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.StatusTerminatedExitCode = int32(valueInt64)
	}
	if value, exist := data[string(ContainerStatusTerminatedReason)]; exist {
		entity.StatusTerminatedReason = value
	}
	if value, exist := data[string(ContainerStatusTerminatedMessage)]; exist {
		entity.StatusTerminatedMessage = value
	}
	if value, exist := data[string(ContainerStatusTerminatedStartedAt)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.StatusTerminatedStartedAt = valueInt64
	}
	if value, exist := data[string(ContainerStatusTerminatedFinishedAt)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.StatusTerminatedFinishedAt = valueInt64
	}
	if value, exist := data[string(ContainerLastTerminationWaitingReason)]; exist {
		entity.LastTerminationWaitingReason = value
	}
	if value, exist := data[string(ContainerLastTerminationWaitingMessage)]; exist {
		entity.LastTerminationWaitingMessage = value
	}
	if value, exist := data[string(ContainerLastTerminationRunningStartedAt)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.LastTerminationRunningStartedAt = valueInt64
	}
	if value, exist := data[string(ContainerLastTerminationTerminatedExitCode)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.LastTerminationTerminatedExitCode = int32(valueInt64)
	}
	if value, exist := data[string(ContainerLastTerminationTerminatedReason)]; exist {
		entity.LastTerminationTerminatedReason = value
	}
	if value, exist := data[string(ContainerLastTerminationTerminatedMessage)]; exist {
		entity.LastTerminationTerminatedMessage = value
	}
	if value, exist := data[string(ContainerLastTerminationTerminatedStartedAt)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.LastTerminationTerminatedStartedAt = valueInt64
	}
	if value, exist := data[string(ContainerLastTerminationTerminatedFinishedAt)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.LastTerminationTerminatedFinishedAt = valueInt64
	}
	if value, exist := data[string(ContainerRestartCount)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.RestartCount = int32(valueInt64)
	}

	return &entity
}

func (e *ContainerEntity) BuildInfluxPoint(measurement string) (*InfluxClient.Point, error) {
	// Pack influx tags
	tags := map[string]string{
		string(ContainerName):        e.Name,
		string(ContainerPodName):     e.PodName,
		string(ContainerNamespace):   e.Namespace,
		string(ContainerNodeName):    e.NodeName,
		string(ContainerClusterName): e.ClusterName,
		string(ContainerUid):         e.Uid,
	}

	// Pack influx fields
	fields := map[string]interface{}{
		string(ContainerResourceRequestCPU):                  e.ResourceRequestCPU,
		string(ContainerResourceRequestMemory):               e.ResourceRequestMemory,
		string(ContainerResourceLimitCPU):                    e.ResourceLimitCPU,
		string(ContainerResourceLimitMemory):                 e.ResourceLimitMemory,
		string(ContainerStatusWaitingReason):                 e.StatusWaitingReason,
		string(ContainerStatusWaitingMessage):                e.StatusWaitingMessage,
		string(ContainerStatusRunningStartedAt):              e.StatusRunningStartedAt,
		string(ContainerStatusTerminatedExitCode):            e.StatusTerminatedExitCode,
		string(ContainerStatusTerminatedReason):              e.StatusTerminatedReason,
		string(ContainerStatusTerminatedMessage):             e.StatusTerminatedMessage,
		string(ContainerStatusTerminatedStartedAt):           e.StatusTerminatedStartedAt,
		string(ContainerStatusTerminatedFinishedAt):          e.StatusTerminatedFinishedAt,
		string(ContainerLastTerminationWaitingReason):        e.LastTerminationWaitingReason,
		string(ContainerLastTerminationWaitingMessage):       e.LastTerminationWaitingMessage,
		string(ContainerLastTerminationRunningStartedAt):     e.LastTerminationRunningStartedAt,
		string(ContainerLastTerminationTerminatedExitCode):   e.LastTerminationTerminatedExitCode,
		string(ContainerLastTerminationTerminatedReason):     e.LastTerminationTerminatedReason,
		string(ContainerLastTerminationTerminatedMessage):    e.LastTerminationTerminatedMessage,
		string(ContainerLastTerminationTerminatedStartedAt):  e.LastTerminationTerminatedStartedAt,
		string(ContainerLastTerminationTerminatedFinishedAt): e.LastTerminationTerminatedFinishedAt,
		string(ContainerRestartCount):                        e.RestartCount,
	}

	return InfluxClient.NewPoint(measurement, tags, fields, e.Time)
}
