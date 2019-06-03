package metrics

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"

	metric_dao "github.com/containers-ai/alameda/datapipe/pkg/dao/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"

	prometheusMetricDAO "github.com/containers-ai/alameda/datapipe/pkg/dao/metrics/prometheus"
	datahubMetricsAPI "github.com/containers-ai/api/datahub/metrics"
	dataPipeMetricsAPI "github.com/containers-ai/api/datapipe/metrics"

	"github.com/containers-ai/alameda/datapipe/pkg/requests"
	commonAPI "github.com/containers-ai/api/common"

	containerEntity "github.com/containers-ai/alameda/datapipe/pkg/entities/influxdb/metrics/container"
	nodeEntity "github.com/containers-ai/alameda/datapipe/pkg/entities/influxdb/metrics/node"

	apiServer "github.com/containers-ai/alameda/datapipe/pkg/repositories/apiserver"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceMetric struct {
	Config *DatapipeConfig.Config
}

func NewServiceMetric(cfg *DatapipeConfig.Config) *ServiceMetric {
	service := ServiceMetric{}
	service.Config = cfg
	return &service
}

func (c *ServiceMetric) CreatePodMetrics(ctx context.Context, in *dataPipeMetricsAPI.CreatePodMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	rowDataList := make([]*commonAPI.WriteRawdata, 0)

	rowData := &commonAPI.WriteRawdata{
		Database:    containerEntity.MetricDatabaseName,
		Table:       containerEntity.MetricMeasurementName,
		Columns:     containerEntity.MetricColumns,
		Rows:        make([]*commonAPI.Row, 0),
		ColumnTypes: containerEntity.MetricColumnsTypes,
		DataTypes:   containerEntity.MetricDataTypes,
	}

	//pod_namespace, pod_name, name, metric_type, value
	for _, pod := range in.GetPodMetrics() {
		podNamespace := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()

		for _, container := range pod.GetContainerMetrics() {
			containerName := container.GetName()
			for key, value := range container.GetMetricData() {
				metricType := datahubMetricsAPI.MetricType(key).String()
				for _, sample := range value.GetData() {
					value := sample.GetNumValue()
					row := &commonAPI.Row{
						Time: sample.GetStartTime(),
						Values: []string{
							podNamespace,
							podName,
							containerName,
							metricType,
							value,
						},
					}

					rowData.Rows = append(rowData.Rows, row)
				}
			}
		}
	}

	rowDataList = append(rowDataList, rowData)

	retStatus, err := apiServer.WriteRawdata(c.Config.APIServer.Address, rowDataList)
	return retStatus, err
}

func (c *ServiceMetric) CreateNodeMetrics(ctx context.Context, in *dataPipeMetricsAPI.CreateNodeMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	rowDataList := make([]*commonAPI.WriteRawdata, 0)

	rowData := &commonAPI.WriteRawdata{
		Database:    nodeEntity.MetricDatabaseName,
		Table:       nodeEntity.MetricMeasurementName,
		Columns:     nodeEntity.MetricColumns,
		Rows:        make([]*commonAPI.Row, 0),
		ColumnTypes: nodeEntity.MetricColumnsTypes,
		DataTypes:   nodeEntity.MetricDataTypes,
	}

	//pod_namespace, pod_name, name, metric_type, value
	for _, node := range in.GetNodeMetrics() {
		name := node.GetName()

		for key, value := range node.GetMetricData() {
			metricType := datahubMetricsAPI.MetricType(key).String()
			for _, sample := range value.GetData() {
				value := sample.GetNumValue()
				row := &commonAPI.Row{
					Time: sample.GetStartTime(),
					Values: []string{
						name,
						metricType,
						value,
					},
				}

				rowData.Rows = append(rowData.Rows, row)
			}
		}
	}

	rowDataList = append(rowDataList, rowData)

	retStatus, err := apiServer.WriteRawdata(c.Config.APIServer.Address, rowDataList)
	return retStatus, err
}

func (c *ServiceMetric) ListPodMetrics(ctx context.Context, in *dataPipeMetricsAPI.ListPodMetricsRequest) (*dataPipeMetricsAPI.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	//--------------------------------------------------------
	var (
		metricDAO metric_dao.MetricsDAO

		requestExt requests.DatahubListPodMetricsRequestExtended
		namespace  = ""
		podName    = ""

		podsMetricMap     metric_dao.PodsMetricMap
		datahubPodMetrics []*datahubMetricsAPI.PodMetric
	)

	requestExt = requests.DatahubListPodMetricsRequestExtended{*in}
	if err := requestExt.Validate(); err != nil {
		return &dataPipeMetricsAPI.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = prometheusMetricDAO.NewWithConfig(*c.Config.Prometheus)

	if in.GetNamespacedName() != nil {
		namespace = in.GetNamespacedName().GetNamespace()
		podName = in.GetNamespacedName().GetName()
	}

	queryCondition := requests.DatahubQueryConditionExtend{QueryCondition: in.GetQueryCondition()}.DaoQueryCondition()
	listPodMetricsRequest := metric_dao.ListPodMetricsRequest{
		Namespace:      namespace,
		PodName:        podName,
		QueryCondition: queryCondition,
	}

	podsMetricMap, err := metricDAO.ListPodMetrics(listPodMetricsRequest)
	if err != nil {
		scope.Errorf("ListPodMetrics failed: %+v", err)
		return &dataPipeMetricsAPI.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	for _, podMetric := range podsMetricMap {
		podMetricExtended := requests.DaoPodMetricExtended{podMetric}
		datahubPodMetric := podMetricExtended.DatahubPodMetric()
		datahubPodMetrics = append(datahubPodMetrics, datahubPodMetric)
	}

	return &dataPipeMetricsAPI.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: datahubPodMetrics,
	}, nil
}

func (c *ServiceMetric) ListNodeMetrics(ctx context.Context, in *dataPipeMetricsAPI.ListNodeMetricsRequest) (*dataPipeMetricsAPI.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err error

		metricDAO metric_dao.MetricsDAO

		requestExt requests.DatahubListNodeMetricsRequestExtended
		nodeNames  []string

		nodesMetricMap     metric_dao.NodesMetricMap
		datahubNodeMetrics []*datahubMetricsAPI.NodeMetric
	)

	requestExt = requests.DatahubListNodeMetricsRequestExtended{*in}
	if err = requestExt.Validate(); err != nil {
		return &dataPipeMetricsAPI.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = prometheusMetricDAO.NewWithConfig(*c.Config.Prometheus)

	nodeNames = in.GetNodeNames()
	queryCondition := requests.DatahubQueryConditionExtend{QueryCondition: in.GetQueryCondition()}.DaoQueryCondition()
	listNodeMetricsRequest := metric_dao.ListNodeMetricsRequest{
		NodeNames:      nodeNames,
		QueryCondition: queryCondition,
	}

	nodesMetricMap, err = metricDAO.ListNodesMetric(listNodeMetricsRequest)
	if err != nil {
		scope.Errorf("ListNodeMetrics failed: %+v", err)
		return &dataPipeMetricsAPI.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	for _, nodeMetric := range nodesMetricMap {
		nodeMetricExtended := requests.DaoNodeMetricExtended{nodeMetric}
		datahubNodeMetric := nodeMetricExtended.DatahubNodeMetric()
		datahubNodeMetrics = append(datahubNodeMetrics, datahubNodeMetric)
	}

	return &dataPipeMetricsAPI.ListNodeMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodeMetrics: datahubNodeMetrics,
	}, nil

}
