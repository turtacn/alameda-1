package v1alpha1

import (
	DaoPredictionImpl "github.com/containers-ai/alameda/datahub/pkg/dao/prediction/impl"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"os"
)

// CreateNodePredictions add node predictions information to database
func (s *ServiceV1alpha1) CreateNodePredictions(ctx context.Context, in *DatahubV1alpha1.CreateNodePredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)
	err := predictionDAO.CreateNodePredictions(in)
	if err != nil {
		scope.Errorf("create node predictions failed: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreatePodPredictions add pod predictions information to database
func (s *ServiceV1alpha1) CreatePodPredictions(ctx context.Context, in *DatahubV1alpha1.CreatePodPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)
	err := predictionDAO.CreateContainerPredictions(in)
	if err != nil {
		scope.Errorf("create pod predictions failed: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// ListNodePredictions list nodes' predictions
func (s *ServiceV1alpha1) ListNodePredictions(ctx context.Context, in *DatahubV1alpha1.ListNodePredictionsRequest) (*DatahubV1alpha1.ListNodePredictionsResponse, error) {
	scope.Debug("Request received from ListNodePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)

	datahubListNodePredictionsRequestExtended := datahubListNodePredictionsRequestExtended{in}
	listNodePredictionRequest := datahubListNodePredictionsRequestExtended.daoListNodePredictionsRequest()
	nodePredictions, err := predictionDAO.ListNodePredictions(listNodePredictionRequest)
	if err != nil {
		scope.Errorf("ListNodePredictions failed: %+v", err)
		return &DatahubV1alpha1.ListNodePredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return &DatahubV1alpha1.ListNodePredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodePredictions: nodePredictions,
	}, nil
}

// ListPodPredictions list pods' predictions
func (s *ServiceV1alpha1) ListPodPredictions(ctx context.Context, in *DatahubV1alpha1.ListPodPredictionsRequest) (*DatahubV1alpha1.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	//--------------------------------------------------------
	_, err := os.Stat("prediction_cpu.csv")
	if !os.IsNotExist(err) {
		return s.ListPodPredictionsDemo(ctx, in)
	}

	//--------------------------------------------------------
	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)

	datahubListPodPredictionsRequestExtended := datahubListPodPredictionsRequestExtended{in}
	listPodPredictionsRequest := datahubListPodPredictionsRequestExtended.daoListPodPredictionsRequest()

	podsPredictions, err := predictionDAO.ListPodPredictions(listPodPredictionsRequest)

	if err != nil {
		scope.Errorf("ListPodPrediction failed: %+v", err)
		return &DatahubV1alpha1.ListPodPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	if in.GetFillDays() > 0 {
		predictionDAO.FillPodPredictions(podsPredictions, in.GetFillDays())
	}

	return &DatahubV1alpha1.ListPodPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPredictions: podsPredictions,
	}, nil
}

// ListPodPredictions list pods' predictions for demo
func (s *ServiceV1alpha1) ListPodPredictionsDemo(ctx context.Context, in *DatahubV1alpha1.ListPodPredictionsRequest) (*DatahubV1alpha1.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictionsDemo grpc function: " + AlamedaUtils.InterfaceToString(in))

	demoPodPredictionList := make([]*DatahubV1alpha1.PodPrediction, 0)
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	if endTime == 0 {
		return &DatahubV1alpha1.ListPodPredictionsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INVALID_ARGUMENT),
			},
			PodPredictions: demoPodPredictionList,
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

	if endTime == 0 {
		return &DatahubV1alpha1.ListPodPredictionsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INVALID_ARGUMENT),
			},
			PodPredictions: demoPodPredictionList,
		}, errors.Errorf("Invalid EndTime")
	}

	tempNamespacedName := DatahubV1alpha1.NamespacedName{
		Namespace: in.NamespacedName.Namespace,
		Name:      in.NamespacedName.Name,
	}

	demoContainerPredictionList := make([]*DatahubV1alpha1.ContainerPrediction, 0)
	demoContainerPrediction := DatahubV1alpha1.ContainerPrediction{
		Name:             in.NamespacedName.Name,
		PredictedRawData: make([]*DatahubV1alpha1.MetricData, 0),
	}
	demoContainerPredictionList = append(demoContainerPredictionList, &demoContainerPrediction)

	demoPredictionDataCPU := DatahubV1alpha1.MetricData{
		MetricType: DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		Data:       make([]*DatahubV1alpha1.Sample, 0),
	}

	demoPredictionDataMem := DatahubV1alpha1.MetricData{
		MetricType: DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES,
		Data:       make([]*DatahubV1alpha1.Sample, 0),
	}

	demoDataMapCPU, _ := DatahubUtils.ReadCSV("prediction_cpu.csv")
	demoDataMapMem, _ := DatahubUtils.ReadCSV("prediction_memory.csv")

	demoKey := in.NamespacedName.Namespace + "_" + in.NamespacedName.Name
	startTime := endTime - int64(step*len(demoDataMapCPU[demoKey]))

	for index, value := range demoDataMapCPU[demoKey] {
		second := startTime + int64(index*step)
		demoPredictionDataCPU.Data = append(demoPredictionDataCPU.Data, &DatahubV1alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	for index, value := range demoDataMapMem[demoKey] {
		second := startTime + int64(index*step)
		demoPredictionDataMem.Data = append(demoPredictionDataMem.Data, &DatahubV1alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	demoContainerPrediction.PredictedRawData = append(demoContainerPrediction.PredictedRawData, &demoPredictionDataCPU)
	demoContainerPrediction.PredictedRawData = append(demoContainerPrediction.PredictedRawData, &demoPredictionDataMem)

	demoPodMetric := DatahubV1alpha1.PodPrediction{
		NamespacedName:       &tempNamespacedName,
		ContainerPredictions: demoContainerPredictionList,
	}
	demoPodPredictionList = append(demoPodPredictionList, &demoPodMetric)

	return &DatahubV1alpha1.ListPodPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPredictions: demoPodPredictionList,
	}, nil
}
