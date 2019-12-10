package v1alpha1

import (
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	K8sMetadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"os"
)

// CreatePodPredictions add pod predictions information to database
func (s *ServiceV1alpha1) CreatePodPredictions(ctx context.Context, in *ApiPredictions.CreatePodPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreatePodPredictionsRequestExtended{CreatePodPredictionsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoPrediction.NewPodPredictionsDAO(*s.Config)
	err := predictionDAO.CreatePredictions(requestExtended.ProducePredictions())
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

// ListPodPredictions list pods' predictions
func (s *ServiceV1alpha1) ListPodPredictions(ctx context.Context, in *ApiPredictions.ListPodPredictionsRequest) (*ApiPredictions.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	_, err := os.Stat("prediction_cpu.csv")
	if !os.IsNotExist(err) {
		return s.ListPodPredictionsDemo(ctx, in)
	}

	requestExt := FormatRequest.ListPodPredictionsRequestExtended{Request: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiPredictions.ListPodPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	predictionDAO := DaoPrediction.NewPodPredictionsDAO(*s.Config)
	podsPredictionMap, err := predictionDAO.ListPredictions(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListPodPredictions failed: %+v", err)
		return &ApiPredictions.ListPodPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	// TODO: must handle filldays in prediction DAO !!!
	/*
		if in.GetFillDays() > 0 {
			predictionDAO.FillPodPredictions(podsPredictions, in.GetFillDays())
		}
	*/

	datahubPodPredictions := make([]*ApiPredictions.PodPrediction, 0)
	for _, podPrediction := range podsPredictionMap.MetricMap {
		podPredictionExtended := FormatResponse.PodPredictionExtended{PodPrediction: podPrediction}
		datahubPodPrediction := podPredictionExtended.ProducePredictions()
		datahubPodPredictions = append(datahubPodPredictions, datahubPodPrediction)
	}

	return &ApiPredictions.ListPodPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPredictions: datahubPodPredictions,
	}, nil
}

// ListPodPredictions list pods' predictions for demo
func (s *ServiceV1alpha1) ListPodPredictionsDemo(ctx context.Context, in *ApiPredictions.ListPodPredictionsRequest) (*ApiPredictions.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictionsDemo grpc function: " + AlamedaUtils.InterfaceToString(in))

	demoPodPredictionList := make([]*ApiPredictions.PodPrediction, 0)
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	if endTime == 0 {
		return &ApiPredictions.ListPodPredictionsResponse{
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
		return &ApiPredictions.ListPodPredictionsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INVALID_ARGUMENT),
			},
			PodPredictions: demoPodPredictionList,
		}, errors.Errorf("Invalid EndTime")
	}

	tempObjectMeta := K8sMetadata.ObjectMeta{
		Namespace: in.ObjectMeta[0].Namespace,
		Name:      in.ObjectMeta[0].Name,
	}

	demoContainerPredictionList := make([]*ApiPredictions.ContainerPrediction, 0)
	demoContainerPrediction := ApiPredictions.ContainerPrediction{
		Name:             in.ObjectMeta[0].Name,
		PredictedRawData: make([]*ApiPredictions.MetricData, 0),
	}
	demoContainerPredictionList = append(demoContainerPredictionList, &demoContainerPrediction)

	demoPredictionDataCPU := ApiPredictions.MetricData{
		MetricType: ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		Data:       make([]*ApiPredictions.Sample, 0),
	}

	demoPredictionDataMem := ApiPredictions.MetricData{
		MetricType: ApiCommon.MetricType_MEMORY_USAGE_BYTES,
		Data:       make([]*ApiPredictions.Sample, 0),
	}

	demoDataMapCPU, _ := DatahubUtils.ReadCSV("prediction_cpu.csv")
	demoDataMapMem, _ := DatahubUtils.ReadCSV("prediction_memory.csv")

	demoKey := in.ObjectMeta[0].Namespace + "_" + in.ObjectMeta[0].Name
	startTime := endTime - int64(step*len(demoDataMapCPU[demoKey]))

	for index, value := range demoDataMapCPU[demoKey] {
		second := startTime + int64(index*step)
		demoPredictionDataCPU.Data = append(demoPredictionDataCPU.Data, &ApiPredictions.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	for index, value := range demoDataMapMem[demoKey] {
		second := startTime + int64(index*step)
		demoPredictionDataMem.Data = append(demoPredictionDataMem.Data, &ApiPredictions.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	demoContainerPrediction.PredictedRawData = append(demoContainerPrediction.PredictedRawData, &demoPredictionDataCPU)
	demoContainerPrediction.PredictedRawData = append(demoContainerPrediction.PredictedRawData, &demoPredictionDataMem)

	demoPodMetric := ApiPredictions.PodPrediction{
		ObjectMeta:           FormatResponse.NewObjectMeta(&tempObjectMeta),
		ContainerPredictions: demoContainerPredictionList,
	}
	demoPodPredictionList = append(demoPodPredictionList, &demoPodMetric)

	return &ApiPredictions.ListPodPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPredictions: demoPodPredictionList,
	}, nil
}
