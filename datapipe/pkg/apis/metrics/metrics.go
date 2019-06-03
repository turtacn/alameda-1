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

	commonAPI "github.com/containers-ai/api/common"
	fedRawdataAPI "github.com/containers-ai/federatorai-api/apiserver/rawdata"

	"github.com/containers-ai/alameda/datapipe/pkg/requests"
	fedRawAPI "github.com/containers-ai/federatorai-api/apiserver/rawdata"

	"fmt"
	"google.golang.org/grpc"
	"time"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type loginCreds struct {
	Username string
	Password string
}

func (c *loginCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.Username,
		"password": c.Password,
	}, nil
}
func (c *loginCreds) RequireTransportSecurity() bool {
	return false
}

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
		Database: "alameda_metric",
		Table:    "container",
		Columns:  []string{"pod_namespace", "pod_name", "name", "metric_type", "value"},
		Rows:     make([]*commonAPI.Row, 0),
		ColumnTypes: []commonAPI.ColumnType{
			commonAPI.ColumnType_COLUMNTYPE_TAG,
			commonAPI.ColumnType_COLUMNTYPE_TAG,
			commonAPI.ColumnType_COLUMNTYPE_TAG,
			commonAPI.ColumnType_COLUMNTYPE_TAG,
			commonAPI.ColumnType_COLUMNTYPE_FIELD},
		DataTypes: []commonAPI.DataType{
			commonAPI.DataType_DATATYPE_STRING,
			commonAPI.DataType_DATATYPE_STRING,
			commonAPI.DataType_DATATYPE_STRING,
			commonAPI.DataType_DATATYPE_STRING,
			commonAPI.DataType_DATATYPE_FLOAT32,
		},
	}

	//pod_namespace, pod_name, name, metric_type, value

	for _, pod := range in.GetPodMetrics() {
		podNamespace := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()

		for _, container := range pod.GetContainerMetrics() {
			containerName := container.GetName()
			for key, value := range container.MetricData {
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

	request := &fedRawdataAPI.WriteRawdataRequest{
		Rawdata: rowDataList,
	}

	conn, err := grpc.Dial(c.Config.APIServer.Address, grpc.WithInsecure(), grpc.WithPerRPCCredentials(&loginCreds{Username: "shofan", Password: "password"}))
	if err != nil {
		fmt.Print(err)
	}
	defer conn.Close()
	ctx, _ = context.WithTimeout(context.Background(), time.Second)

	client := fedRawAPI.NewRawdataServiceClient(conn)
	_, err = client.WriteRawdata(ctx, request)
	if err != nil {
		fmt.Print(err)
	}

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceMetric) CreateNodeMetrics(ctx context.Context, in *dataPipeMetricsAPI.CreateNodeMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
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
