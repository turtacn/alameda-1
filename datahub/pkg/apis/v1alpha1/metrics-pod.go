package v1alpha1

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	K8sMetadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"os"
)

func (s *ServiceV1alpha1) CreatePodMetrics(ctx context.Context, in *ApiMetrics.CreatePodMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreatePodMetricsRequestExtended{CreatePodMetricsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	metricDAO := DaoMetric.NewPodMetricsWriterDAO(*s.Config)
	err := metricDAO.CreateMetrics(ctx, requestExtended.ProduceMetrics())
	if err != nil {
		scope.Errorf("failed to create pod metrics: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListPodMetrics(ctx context.Context, in *ApiMetrics.ListPodMetricsRequest) (*ApiMetrics.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	_, err := os.Stat("metric_cpu.csv")
	if !os.IsNotExist(err) {
		return s.ListPodMetricsDemo(ctx, in)
	}

	requestExt := FormatRequest.ListPodMetricsRequestExtended{Request: in}
	if err = requestExt.Validate(); err != nil {
		return &ApiMetrics.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}
	requestExt.SetDefaultWithMetricsDBType(s.Config.Apis.Metrics.Source)

	metricDAO := DaoMetric.NewPodMetricsReaderDAO(*s.Config)
	podMetricMap, err := metricDAO.ListMetrics(ctx, requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListPodMetrics failed: %+v", err)
		return &ApiMetrics.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	datahubPodMetrics := make([]*ApiMetrics.PodMetric, 0)
	for _, podMetric := range podMetricMap.MetricMap {
		podMetricExtended := FormatResponse.PodMetricExtended{PodMetric: podMetric}
		datahubPodMetric := podMetricExtended.ProduceMetrics()
		datahubPodMetrics = append(datahubPodMetrics, datahubPodMetric)
	}

	return &ApiMetrics.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: datahubPodMetrics,
	}, nil
}

// ListPodMetrics list pods' metrics for demo
func (s *ServiceV1alpha1) ListPodMetricsDemo(ctx context.Context, in *ApiMetrics.ListPodMetricsRequest) (*ApiMetrics.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetricsDemo grpc function: " + AlamedaUtils.InterfaceToString(in))

	demoPodMetricList := make([]*ApiMetrics.PodMetric, 0)
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	if endTime == 0 {
		return &ApiMetrics.ListPodMetricsResponse{
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

	tempObjectMeta := K8sMetadata.ObjectMeta{
		Namespace: in.ObjectMeta[0].Namespace,
		Name:      in.ObjectMeta[0].Name,
	}

	demoContainerMetricList := make([]*ApiMetrics.ContainerMetric, 0)
	demoContainerMetric := ApiMetrics.ContainerMetric{
		Name:       in.ObjectMeta[0].Name,
		MetricData: make([]*ApiCommon.MetricData, 0),
	}
	demoContainerMetricList = append(demoContainerMetricList, &demoContainerMetric)

	demoMetricDataCPU := ApiCommon.MetricData{
		MetricType: ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		Data:       make([]*ApiCommon.Sample, 0),
	}

	demoMetricDataMem := ApiCommon.MetricData{
		MetricType: ApiCommon.MetricType_MEMORY_USAGE_BYTES,
		Data:       make([]*ApiCommon.Sample, 0),
	}

	demoDataMapCPU, _ := DatahubUtils.ReadCSV("metric_cpu.csv")
	demoDataMapMem, _ := DatahubUtils.ReadCSV("metric_memory.csv")

	demoKey := in.ObjectMeta[0].Namespace + "_" + in.ObjectMeta[0].Name

	startTime := endTime - int64(step*len(demoDataMapCPU[demoKey]))
	for index, value := range demoDataMapCPU[demoKey] {
		second := startTime + int64(index*step)
		demoMetricDataCPU.Data = append(demoMetricDataCPU.Data, &ApiCommon.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	for index, value := range demoDataMapMem[demoKey] {
		second := startTime + int64(index*step)
		demoMetricDataMem.Data = append(demoMetricDataMem.Data, &ApiCommon.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	demoContainerMetric.MetricData = append(demoContainerMetric.MetricData, &demoMetricDataCPU)
	demoContainerMetric.MetricData = append(demoContainerMetric.MetricData, &demoMetricDataMem)

	demoPodMetric := ApiMetrics.PodMetric{
		ObjectMeta:       FormatResponse.NewObjectMeta(&tempObjectMeta),
		ContainerMetrics: demoContainerMetricList,
	}
	demoPodMetricList = append(demoPodMetricList, &demoPodMetric)

	return &ApiMetrics.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: demoPodMetricList,
	}, nil
}
