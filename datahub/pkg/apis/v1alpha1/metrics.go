package v1alpha1

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	DaoMetricPromth "github.com/containers-ai/alameda/datahub/pkg/dao/metric/prometheus"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"os"
)

// ListNodeMetrics list nodes' metrics
func (s *ServiceV1alpha1) ListNodeMetrics(ctx context.Context, in *DatahubV1alpha1.ListNodeMetricsRequest) (*DatahubV1alpha1.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err error

		metricDAO DaoMetric.MetricsDAO

		requestExt     datahubListNodeMetricsRequestExtended
		nodeNames      []string
		queryCondition DBCommon.QueryCondition

		nodesMetricMap     DaoMetric.NodesMetricMap
		datahubNodeMetrics []*DatahubV1alpha1.NodeMetric
	)

	requestExt = datahubListNodeMetricsRequestExtended{*in}
	if err = requestExt.validate(); err != nil {
		return &DatahubV1alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = DaoMetricPromth.NewWithConfig(*s.Config.Prometheus)

	nodeNames = in.GetNodeNames()
	queryCondition = datahubQueryConditionExtend{queryCondition: in.GetQueryCondition()}.daoQueryCondition()
	listNodeMetricsRequest := DaoMetric.ListNodeMetricsRequest{
		NodeNames:      nodeNames,
		QueryCondition: queryCondition,
	}

	nodesMetricMap, err = metricDAO.ListNodesMetric(listNodeMetricsRequest)
	if err != nil {
		scope.Errorf("ListNodeMetrics failed: %+v", err)
		return &DatahubV1alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	for _, nodeMetric := range nodesMetricMap {
		nodeMetricExtended := daoNodeMetricExtended{nodeMetric}
		datahubNodeMetric := nodeMetricExtended.datahubNodeMetric()
		datahubNodeMetrics = append(datahubNodeMetrics, datahubNodeMetric)
	}

	return &DatahubV1alpha1.ListNodeMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodeMetrics: datahubNodeMetrics,
	}, nil
}

// ListPodMetrics list pods' metrics
func (s *ServiceV1alpha1) ListPodMetrics(ctx context.Context, in *DatahubV1alpha1.ListPodMetricsRequest) (*DatahubV1alpha1.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	//--------------------------------------------------------
	_, err := os.Stat("metric_cpu.csv")
	if !os.IsNotExist(err) {
		return s.ListPodMetricsDemo(ctx, in)
	}

	//--------------------------------------------------------
	var (
		metricDAO DaoMetric.MetricsDAO

		requestExt     datahubListPodMetricsRequestExtended
		namespace      = ""
		podName        = ""
		queryCondition DBCommon.QueryCondition

		podsMetricMap     DaoMetric.PodsMetricMap
		datahubPodMetrics []*DatahubV1alpha1.PodMetric
	)

	requestExt = datahubListPodMetricsRequestExtended{*in}
	if err = requestExt.validate(); err != nil {
		return &DatahubV1alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = DaoMetricPromth.NewWithConfig(*s.Config.Prometheus)

	if in.GetNamespacedName() != nil {
		namespace = in.GetNamespacedName().GetNamespace()
		podName = in.GetNamespacedName().GetName()
	}
	queryCondition = datahubQueryConditionExtend{queryCondition: in.GetQueryCondition()}.daoQueryCondition()
	listPodMetricsRequest := DaoMetric.ListPodMetricsRequest{
		Namespace:      namespace,
		PodName:        podName,
		QueryCondition: queryCondition,
	}

	podsMetricMap, err = metricDAO.ListPodMetrics(listPodMetricsRequest)
	if err != nil {
		scope.Errorf("ListPodMetrics failed: %+v", err)
		return &DatahubV1alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	for _, podMetric := range podsMetricMap {
		podMetricExtended := daoPodMetricExtended{podMetric}
		datahubPodMetric := podMetricExtended.datahubPodMetric()
		datahubPodMetrics = append(datahubPodMetrics, datahubPodMetric)
	}

	return &DatahubV1alpha1.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: datahubPodMetrics,
	}, nil
}

// ListPodMetrics list pods' metrics for demo
func (s *ServiceV1alpha1) ListPodMetricsDemo(ctx context.Context, in *DatahubV1alpha1.ListPodMetricsRequest) (*DatahubV1alpha1.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetricsDemo grpc function: " + AlamedaUtils.InterfaceToString(in))

	demoPodMetricList := make([]*DatahubV1alpha1.PodMetric, 0)
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	if endTime == 0 {
		return &DatahubV1alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INVALID_ARGUMENT),
			},
			PodMetrics: demoPodMetricList,
		}, errors.Errorf("Invalid EndTime")
	}

	if endTime%3600 != 0 {
		endTime = endTime - (endTime % 3600) + 3600
	}

	//step := int(in.GetQueryCondition().GetTimeRange().GetStep().GetSeconds())
	step := 3600
	if step == 0 {
		step = 3600
	}

	tempNamespacedName := DatahubV1alpha1.NamespacedName{
		Namespace: in.NamespacedName.Namespace,
		Name:      in.NamespacedName.Name,
	}

	demoContainerMetricList := make([]*DatahubV1alpha1.ContainerMetric, 0)
	demoContainerMetric := DatahubV1alpha1.ContainerMetric{
		Name:       in.NamespacedName.Name,
		MetricData: make([]*DatahubV1alpha1.MetricData, 0),
	}
	demoContainerMetricList = append(demoContainerMetricList, &demoContainerMetric)

	demoMetricDataCPU := DatahubV1alpha1.MetricData{
		MetricType: DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		Data:       make([]*DatahubV1alpha1.Sample, 0),
	}

	demoMetricDataMem := DatahubV1alpha1.MetricData{
		MetricType: DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES,
		Data:       make([]*DatahubV1alpha1.Sample, 0),
	}

	demoDataMapCPU, _ := DatahubUtils.ReadCSV("metric_cpu.csv")
	demoDataMapMem, _ := DatahubUtils.ReadCSV("metric_memory.csv")

	demoKey := in.NamespacedName.Namespace + "_" + in.NamespacedName.Name

	startTime := endTime - int64(step*len(demoDataMapCPU[demoKey]))
	for index, value := range demoDataMapCPU[demoKey] {
		second := startTime + int64(index*step)
		demoMetricDataCPU.Data = append(demoMetricDataCPU.Data, &DatahubV1alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	for index, value := range demoDataMapMem[demoKey] {
		second := startTime + int64(index*step)
		demoMetricDataMem.Data = append(demoMetricDataMem.Data, &DatahubV1alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	demoContainerMetric.MetricData = append(demoContainerMetric.MetricData, &demoMetricDataCPU)
	demoContainerMetric.MetricData = append(demoContainerMetric.MetricData, &demoMetricDataMem)

	demoPodMetric := DatahubV1alpha1.PodMetric{
		NamespacedName:   &tempNamespacedName,
		ContainerMetrics: demoContainerMetricList,
	}
	demoPodMetricList = append(demoPodMetricList, &demoPodMetric)

	return &DatahubV1alpha1.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: demoPodMetricList,
	}, nil
}
