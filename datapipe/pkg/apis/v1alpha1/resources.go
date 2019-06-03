package v1alpha1

import (
	entity "github.com/containers-ai/alameda/datapipe/pkg/entities/influxdb/cluster_status"
	apiServer "github.com/containers-ai/alameda/datapipe/pkg/repositories/apiserver"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	commonAPI "github.com/containers-ai/api/common"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/rpc/status"
	"strconv"
)

func (s *ServiceV1alpha1) CreatePodsImpl(in *datahub_v1alpha1.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	rowDataList := make([]*commonAPI.WriteRawdata, 0)

	rowData := &commonAPI.WriteRawdata{
		Database:    entity.ContainerDatabaseName,
		Table:       entity.ContainerMeasurementName,
		Columns:     entity.ContainerColumns,
		Rows:        make([]*commonAPI.Row, 0),
		ColumnTypes: entity.ContainerColumnsTypes,
		DataTypes:   entity.ContainerDataTypes,
	}

	indexMap := entity.ContainerColIndexMap

	for _, pod := range in.GetPods() {
		podNamespace := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()
		//meta := pod.GetMeta()
		nodeName := pod.GetNodeName()
		resourceLink := pod.GetResourceLink()
		status := pod.GetStatus()
		topController := pod.GetTopController()
		startTime := pod.GetStartTime()

		for _, container := range pod.GetContainers() {
			containerName := container.GetName()
			containerLimitResources := container.GetLimitResource()
			containerRequestResources := container.GetRequestResource()

			containerLimitResourcesCPU := ""
			containerLimitResourcesMEM := ""

			containerRequestResourcesCPU := ""
			containerRequestResourcesMEM := ""

			for _, value := range containerLimitResources {
				if value.GetMetricType() == datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE {
					if len(value.GetData()) > 0 {
						containerLimitResourcesCPU = value.GetData()[0].GetNumValue()
					}
				}
				if value.GetMetricType() == datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES {
					if len(value.GetData()) > 0 {
						containerLimitResourcesMEM = value.GetData()[0].GetNumValue()
					}
				}
			}

			for _, value := range containerRequestResources {
				if value.GetMetricType() == datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE {
					if len(value.GetData()) > 0 {
						containerRequestResourcesCPU = value.GetData()[0].GetNumValue()
					}
				}
				if value.GetMetricType() == datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES {
					if len(value.GetData()) > 0 {
						containerRequestResourcesMEM = value.GetData()[0].GetNumValue()
					}
				}
			}

			containerStatus := container.GetStatus()

			state := containerStatus.GetState()
			lastTerminationState := containerStatus.GetLastTerminationState()

			values := make([]string, len(entity.ContainerColumns))
			values[indexMap[entity.ContainerNamespace]] = podNamespace
			values[indexMap[entity.ContainerPodName]] = podName
			values[indexMap[entity.ContainerAlamedaScalerNamespace]] = pod.GetAlamedaScaler().GetNamespace()
			values[indexMap[entity.ContainerAlamedaScalerName]] = pod.GetAlamedaScaler().GetName()
			values[indexMap[entity.ContainerNodeName]] = nodeName
			values[indexMap[entity.ContainerName]] = containerName
			values[indexMap[entity.ContainerAppName]] = pod.GetAppName()
			values[indexMap[entity.ContainerAppPartOf]] = pod.GetAppPartOf()

			values[indexMap[entity.ContainerPodPhase]] = status.GetPhase().String()
			values[indexMap[entity.ContainerPodMessage]] = status.GetMessage()
			values[indexMap[entity.ContainerPodReason]] = status.GetReason()

			values[indexMap[entity.ContainerStatusWaitingReason]] = state.GetWaiting().GetReason()
			values[indexMap[entity.ContainerStatusWaitingMessage]] = state.GetWaiting().GetMessage()
			values[indexMap[entity.ContainerStatusRunningStartedAt]] = strconv.FormatInt(state.GetRunning().GetStartedAt().GetSeconds(), 10)
			values[indexMap[entity.ContainerStatusTerminatedExitCode]] = strconv.FormatInt(int64(state.GetTerminated().GetExitCode()), 10)
			values[indexMap[entity.ContainerStatusTerminatedReason]] = state.GetTerminated().GetReason()
			values[indexMap[entity.ContainerStatusTerminatedMessage]] = state.GetTerminated().GetMessage()
			values[indexMap[entity.ContainerStatusTerminatedStartedAt]] = strconv.FormatInt(state.GetTerminated().GetStartedAt().GetSeconds(), 10)
			values[indexMap[entity.ContainerStatusTerminatedFinishedAt]] = strconv.FormatInt(state.GetTerminated().GetFinishedAt().GetSeconds(), 10)

			values[indexMap[entity.ContainerLastTerminationStatusWaitingReason]] = lastTerminationState.GetWaiting().GetReason()
			values[indexMap[entity.ContainerLastTerminationStatusWaitingMessage]] = lastTerminationState.GetWaiting().GetMessage()
			values[indexMap[entity.ContainerLastTerminationStatusRunningStartedAt]] = strconv.FormatInt(lastTerminationState.GetRunning().GetStartedAt().GetSeconds(), 10)
			values[indexMap[entity.ContainerLastTerminationStatusTerminatedExitCode]] = strconv.FormatInt(int64(lastTerminationState.GetTerminated().GetExitCode()), 10)
			values[indexMap[entity.ContainerLastTerminationStatusTerminatedReason]] = lastTerminationState.GetTerminated().GetReason()
			values[indexMap[entity.ContainerLastTerminationStatusTerminatedMessage]] = lastTerminationState.GetTerminated().GetMessage()
			values[indexMap[entity.ContainerLastTerminationStatusTerminatedStartedAt]] = strconv.FormatInt(lastTerminationState.GetTerminated().GetStartedAt().GetSeconds(), 10)
			values[indexMap[entity.ContainerLastTerminationStatusTerminatedFinishedAt]] = strconv.FormatInt(lastTerminationState.GetTerminated().GetFinishedAt().GetSeconds(), 10)

			values[indexMap[entity.ContainerRestartCount]] = strconv.FormatInt(int64(containerStatus.GetRestartCount()), 10)

			values[indexMap[entity.ContainerResourceRequestCPU]] = containerRequestResourcesCPU
			values[indexMap[entity.ContainerResourceRequestMemory]] = containerRequestResourcesMEM
			values[indexMap[entity.ContainerResourceLimitCPU]] = containerLimitResourcesCPU
			values[indexMap[entity.ContainerResourceLimitMemory]] = containerLimitResourcesMEM

			values[indexMap[entity.ContainerPolicy]] = pod.GetPolicy().String()
			values[indexMap[entity.ContainerPodCreateTime]] = strconv.FormatInt(startTime.GetSeconds(), 10)
			values[indexMap[entity.ContainerResourceLink]] = resourceLink
			values[indexMap[entity.ContainerTopControllerName]] = topController.GetNamespacedName().GetName()
			values[indexMap[entity.ContainerTopControllerKind]] = topController.GetKind().String()
			values[indexMap[entity.ContainerTpoControllerReplicas]] = strconv.FormatInt(int64(topController.GetReplicas()), 10)
			values[indexMap[entity.ContainerEnableHPA]] = strconv.FormatBool(pod.GetEnable_HPA())
			values[indexMap[entity.ContainerEnableVPA]] = strconv.FormatBool(pod.GetEnable_VPA())

			row := &commonAPI.Row{
				Time:   &timestamp.Timestamp{Seconds: 0},
				Values: values,
			}

			rowData.Rows = append(rowData.Rows, row)
		}
	}

	rowDataList = append(rowDataList, rowData)

	retStatus, err := apiServer.WriteRawdata(s.Config.APIServer.Address, rowDataList)
	return retStatus, err
}
