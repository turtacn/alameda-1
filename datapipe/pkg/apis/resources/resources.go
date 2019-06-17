package resources

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Resources "github.com/containers-ai/api/datapipe/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"

	commonAPI "github.com/containers-ai/api/common"

	entity "github.com/containers-ai/alameda/datapipe/pkg/entities/influxdb/cluster_status/container"
	apiServer "github.com/containers-ai/alameda/datapipe/pkg/repositories/apiserver"

	datahubMetricsAPI "github.com/containers-ai/api/datahub/metrics"
	"github.com/golang/protobuf/ptypes/timestamp"
	"strconv"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceResource struct {
	Config *DatapipeConfig.Config
}

func NewServiceResource(cfg *DatapipeConfig.Config) *ServiceResource {
	service := ServiceResource{}
	service.Config = cfg
	return &service
}

func (c *ServiceResource) CreateContainers(ctx context.Context, in *Resources.CreateContainersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateContainers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) CreatePods(ctx context.Context, in *Resources.CreatePodsRequest) (*status.Status, error) {
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

	metricTypeCpu := int32(datahubMetricsAPI.MetricType_CPU_USAGE_SECONDS_PERCENTAGE)
	metricTypeMem := int32(datahubMetricsAPI.MetricType_MEMORY_USAGE_BYTES)

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
		alamedaPodSpec := pod.GetAlamedaPodSpec()

		for _, container := range pod.GetContainers() {
			containerName := container.GetName()
			containerResources := container.GetResources()
			containerStatus := container.GetStatus()

			state := containerStatus.GetState()
			lastTerminationState := containerStatus.GetLastTerminationState()

			values := make([]string, len(entity.ContainerColumns))
			values[indexMap[entity.ContainerNamespace]] = podNamespace
			values[indexMap[entity.ContainerPodName]] = podName
			values[indexMap[entity.ContainerAlamedaScalerNamespace]] = alamedaPodSpec.GetScaler().GetNamespace()
			values[indexMap[entity.ContainerAlamedaScalerName]] = alamedaPodSpec.GetScaler().GetName()
			values[indexMap[entity.ContainerNodeName]] = nodeName
			values[indexMap[entity.ContainerName]] = containerName
			values[indexMap[entity.ContainerAppName]] = alamedaPodSpec.GetAppName()
			values[indexMap[entity.ContainerAppPartOf]] = alamedaPodSpec.GetAppPartOf()

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

			values[indexMap[entity.ContainerResourceRequestCPU]] = containerResources.GetRequests()[metricTypeCpu]
			values[indexMap[entity.ContainerResourceRequestMemory]] = containerResources.GetRequests()[metricTypeMem]
			values[indexMap[entity.ContainerResourceLimitCPU]] = containerResources.GetLimits()[metricTypeCpu]
			values[indexMap[entity.ContainerResourceLimitMemory]] = containerResources.GetLimits()[metricTypeMem]

			values[indexMap[entity.ContainerPolicy]] = alamedaPodSpec.GetPolicy().String()
			values[indexMap[entity.ContainerPodCreateTime]] = strconv.FormatInt(startTime.GetSeconds(), 10)
			values[indexMap[entity.ContainerResourceLink]] = resourceLink
			values[indexMap[entity.ContainerTopControllerName]] = topController.GetNamespacedName().GetName()
			values[indexMap[entity.ContainerTopControllerKind]] = topController.GetKind().String()
			values[indexMap[entity.ContainerTpoControllerReplicas]] = strconv.FormatInt(int64(topController.GetReplicas()), 10)
			values[indexMap[entity.ContainerEnableHPA]] = strconv.FormatBool(alamedaPodSpec.GetEnable_HPA())
			values[indexMap[entity.ContainerEnableVPA]] = strconv.FormatBool(alamedaPodSpec.GetEnable_VPA())

			row := &commonAPI.Row{
				Time:   &timestamp.Timestamp{Seconds: 0},
				Values: values,
			}

			rowData.Rows = append(rowData.Rows, row)
		}
	}

	rowDataList = append(rowDataList, rowData)

	retStatus, err := apiServer.WriteRawdata(c.Config.APIServer.Address, commonAPI.DatabaseType_INFLUXDB, rowDataList)
	return retStatus, err
}

func (c *ServiceResource) CreateControllers(ctx context.Context, in *Resources.CreateControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) CreateNodes(ctx context.Context, in *Resources.CreateNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) ListContainers(ctx context.Context, in *Resources.ListContainersRequest) (*Resources.ListContainersResponse, error) {
	scope.Debug("Request received from ListContainers grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListContainersResponse)
	return out, nil
}

func (c *ServiceResource) ListPods(ctx context.Context, in *Resources.ListPodsRequest) (*Resources.ListPodsResponse, error) {
	scope.Debug("Request received from ListPods grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListPodsResponse)
	return out, nil
}

func (c *ServiceResource) ListPodsByNodeName(ctx context.Context, in *Resources.ListPodsByNodeNamesRequest) (*Resources.ListPodsResponse, error) {
	scope.Debug("Request received from ListPodsByNodeName grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListPodsResponse)
	return out, nil
}

func (c *ServiceResource) ListControllers(ctx context.Context, in *Resources.ListControllersRequest) (*Resources.ListControllersResponse, error) {
	scope.Debug("Request received from ListControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListControllersResponse)
	return out, nil
}

func (c *ServiceResource) ListNodes(ctx context.Context, in *Resources.ListNodesRequest) (*Resources.ListNodesResponse, error) {
	scope.Debug("Request received from ListNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListNodesResponse)
	return out, nil
}

func (c *ServiceResource) DeleteContainers(ctx context.Context, in *Resources.DeleteContainersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteContainers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) DeletePods(ctx context.Context, in *Resources.DeletePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeletePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) DeleteControllers(ctx context.Context, in *Resources.DeleteControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) DeleteNodes(ctx context.Context, in *Resources.DeleteNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
