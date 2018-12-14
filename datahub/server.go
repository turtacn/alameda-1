package datahub

import (
	go_context "context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics/prometheus"
	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	"github.com/containers-ai/alameda/operator/pkg/controller/alamedaresource"
	"github.com/containers-ai/alameda/pkg/utils/log"
	operator_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/operator"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	apicorev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Server struct {
	err    chan error
	server *grpc.Server

	Config    Config
	K8SClient client.Client
	MetricsDB metrics.MetricsDB
}

var (
	scope = log.RegisterScope("gRPC", "gRPC server log", 0)
)

func NewServer(cfg Config) (*Server, error) {

	var (
		err error

		server    *Server
		k8sCli    client.Client
		metricsDB metrics.MetricsDB
	)

	if err = cfg.Validate(); err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	k8sClientConfig, err := config.GetConfig()
	if err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	if k8sCli, err = client.New(k8sClientConfig, client.Options{}); err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	if metricsDB, err = prometheus.New(*cfg.Prometheus); err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	server = &Server{
		err: make(chan error),

		Config:    cfg,
		K8SClient: k8sCli,
		MetricsDB: metricsDB,
	}

	return server, nil
}

func (s *Server) Run() error {

	// Open metrics database
	if err := s.MetricsDB.Connect(); err != nil {
		return err
	}

	// build server listener
	scope.Info(("starting gRPC server"))
	ln, err := net.Listen("tcp", s.Config.BindAddress)
	if err != nil {
		scope.Error("gRPC server failed listen: " + err.Error())
		return fmt.Errorf("GRPC server failed to bind address: %s", s.Config.BindAddress)
	}
	scope.Info("gRPC server listening on " + s.Config.BindAddress)

	server, err := s.newGRPCServer()
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	s.server = server

	s.registGRPCServer(server)
	reflection.Register(server)

	if err := server.Serve(ln); err != nil {
		s.err <- fmt.Errorf("GRPC server failed to serve: %s", err.Error())
	}

	return nil
}

func (s *Server) Stop() error {

	if err := s.MetricsDB.Close(); err != nil {
		return err
	}

	s.server.Stop()

	return nil
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) newGRPCServer() (*grpc.Server, error) {

	var (
		server *grpc.Server
	)

	server = grpc.NewServer()

	return server, nil
}

func (s *Server) registGRPCServer(server *grpc.Server) {

	operator_v1alpha1.RegisterOperatorServiceServer(server, s)
}

func (s *Server) ListMetrics(ctx context.Context, in *operator_v1alpha1.ListMetricsRequest) (*operator_v1alpha1.ListMetricsResponse, error) {

	var resp *operator_v1alpha1.ListMetricsResponse

	// Validate request
	err := ValidateListMetricsRequest(in)
	if err != nil {
		resp = &operator_v1alpha1.ListMetricsResponse{}
		resp.Status = &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: err.Error(),
		}
		return resp, nil
	}

	// build query instance to query metrics db
	q := buildMetircQuery(in)

	// query to metrics db
	quertResp, err := s.MetricsDB.Query(q)
	if err != nil {
		resp = &operator_v1alpha1.ListMetricsResponse{}
		resp.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return resp, nil
	}

	// convert response of query metrics db to containers-ai.operator.v1alpha1.ListMetricssResposne
	resp = convertMetricsQueryResponseToProtoResponse(&quertResp)
	resp.Status = &status.Status{
		Code: int32(code.Code_OK),
	}
	return resp, nil
}

func (s *Server) ListMetricsSum(ctx context.Context, in *operator_v1alpha1.ListMetricsSumRequest) (*operator_v1alpha1.ListMetricsSumResponse, error) {

	return &operator_v1alpha1.ListMetricsSumResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}

func (s *Server) CreatePredictResult(ctx context.Context, in *operator_v1alpha1.CreatePredictResultRequest) (*operator_v1alpha1.CreatePredictResultResponse, error) {
	// 1. Get namespace list information from predicted pods
	nsRange := map[string]bool{}
	for _, predictPod := range in.GetPredictPods() {
		if _, ok := nsRange[predictPod.GetNamespace()]; !ok {
			nsRange[predictPod.GetNamespace()] = true
		}
	}
	// 2. Get AlamedaResource list from namespace list
	alaListRange := []autoscalingv1alpha1.AlamedaResource{}
	for namespace, _ := range nsRange {
		alamedaresourceList := &autoscalingv1alpha1.AlamedaResourceList{}
		err := s.K8SClient.List(go_context.TODO(), client.InNamespace(namespace), alamedaresourceList)
		if err == nil {
			alaListRange = append(alaListRange, alamedaresourceList.Items...)
		}
	}
	if len(alaListRange) == 0 {
		return &operator_v1alpha1.CreatePredictResultResponse{
			Status: &status.Status{
				Code:    int32(code.Code_NOT_FOUND),
				Message: "AlamedaResource not found.",
			},
		}, nil
	}
	for _, ala := range alaListRange {
		alaAnno := ala.GetAnnotations()
		predictPodsInAla := []*operator_v1alpha1.PredictPod{}
		if alaAnno == nil {
			scope.Warnf(fmt.Sprintf("No annotation found in AlamedaResouce %s in namespace %s in AlamedaResource list, try searching next item", ala.GetName(), ala.GetNamespace()))
			continue
		}
		if _, ok := alaAnno[alamedaresource.AlamedaK8sController]; !ok {
			scope.Warnf(fmt.Sprintf("No k8s controller annotation key found in AlamedaResouce %s in namespace %s in AlamedaResource list, try searching next item", ala.GetName(), ala.GetNamespace()))
			continue
		}
		scope.Infof(fmt.Sprintf("K8s controller annotation found %s in AlamedaResouce %s in namespace %s in AlamedaResource list", alaAnno[alamedaresource.AlamedaK8sController], ala.GetName(), ala.GetNamespace()))
		for _, predictPod := range in.GetPredictPods() {
			alaK8sCtrStr := alaAnno[alamedaresource.AlamedaK8sController]
			if isAlamedaPod(alaK8sCtrStr, predictPod.GetUid()) {
				predictPodsInAla = append(predictPodsInAla, predictPod)
			} else {
				scope.Infof(fmt.Sprintf("Pod %s do not belong to AlamedaResource (%s/%s)", predictPod.GetUid(), ala.GetNamespace(), ala.GetName()))
			}
		}
		if len(predictPodsInAla) > 0 {
			s.updateAlamedaResourcePredict(ala, predictPodsInAla)
		}
	}
	inBin, _ := json.Marshal(*in)
	return &operator_v1alpha1.CreatePredictResultResponse{
		Status: &status.Status{
			Code:    int32(code.Code_OK),
			Message: string(inBin),
		},
	}, nil
}

func (s *Server) updateAlamedaResourcePredict(ala autoscalingv1alpha1.AlamedaResource, predictPods []*operator_v1alpha1.PredictPod) {
	alamedaresourcePredict := &autoscalingv1alpha1.AlamedaResourcePrediction{}
	s.K8SClient.Get(go_context.TODO(), types.NamespacedName{
		Name:      ala.GetName(),
		Namespace: ala.GetNamespace(),
	}, alamedaresourcePredict)

	for _, predictPod := range predictPods {
		for _, deployment := range alamedaresourcePredict.Status.Prediction.Deployments {
			if _, ok := deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())]; ok {
				for _, predictContainer := range predictPod.GetPredictContainers() {
					for containerName, _ := range deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers {
						if predictContainer.GetName() == string(containerName) {
							scope.Infof(fmt.Sprintf("Update Prediction from AI service. (%s/%s)", ala.GetNamespace(), ala.GetName()))
							recommendation, initialResource := s.updateAlamedaResourcePredictContainer(deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers[containerName], predictContainer)
							alaPredictContainer := deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers[containerName]
							alaPredictContainer.Recommendations = recommendation
							alaPredictContainer.InitialResource = initialResource
							deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers[containerName] = alaPredictContainer
						}
					}
				}
			}
		}
	}

	err := s.K8SClient.Update(context.TODO(), alamedaresourcePredict)
	if err != nil {
		scope.Error(err.Error())
	}
}

func (s *Server) updateAlamedaResourcePredictContainer(alaPredictContainer autoscalingv1alpha1.PredictContainer, predictContainer *operator_v1alpha1.PredictContainer) ([]autoscalingv1alpha1.Recommendation, apicorev1.ResourceRequirements) {
	for resource, predictData := range predictContainer.RowPredictData {
		tsData := autoscalingv1alpha1.TimeSeriesData{
			PredictData: []autoscalingv1alpha1.PredictData{},
		}
		for _, data := range predictData.GetPredictData() {
			if data.Time == nil {
				dataBin, _ := json.Marshal(data)
				scope.Infof(fmt.Sprintf("Predict data from AI server contains no time field. %s", dataBin))
			} else {
				date := time.Unix(data.Time.Seconds, 0)
				tsData.PredictData = append(tsData.PredictData, autoscalingv1alpha1.PredictData{
					Time:  data.Time.Seconds,
					Value: data.Value,
					Date:  date.String(),
				})
			}
		}
		alaPredictContainer.RawPredict[autoscalingv1alpha1.ResourceType(resource)] = tsData
	}
	recommendations := []autoscalingv1alpha1.Recommendation{}
	for _, recommendation := range predictContainer.Recommendations {
		resource := apicorev1.ResourceRequirements{
			Limits:   map[apicorev1.ResourceName]apiresource.Quantity{},
			Requests: map[apicorev1.ResourceName]apiresource.Quantity{},
		}
		for limitKey, limit := range recommendation.Resource.Limit {
			resource.Limits[apicorev1.ResourceName(limitKey)] = apiresource.MustParse(limit)
		}
		for requestKey, request := range recommendation.Resource.Request {
			resource.Requests[apicorev1.ResourceName(requestKey)] = apiresource.MustParse(request)
		}
		date := time.Unix(recommendation.Time.Seconds, 0)
		recommendations = append(recommendations, autoscalingv1alpha1.Recommendation{
			Time:      recommendation.Time.Seconds,
			Date:      date.String(),
			Resources: resource,
		})
	}

	alaPredictContainer.Recommendations = recommendations

	initialResource := apicorev1.ResourceRequirements{
		Limits:   map[apicorev1.ResourceName]apiresource.Quantity{},
		Requests: map[apicorev1.ResourceName]apiresource.Quantity{},
	}
	if predictContainer.InitialResource != nil {
		for limitKey, limit := range predictContainer.InitialResource.Limit {
			initialResource.Limits[apicorev1.ResourceName(limitKey)] = apiresource.MustParse(limit)
		}
		for requestKey, request := range predictContainer.InitialResource.Request {
			initialResource.Requests[apicorev1.ResourceName(requestKey)] = apiresource.MustParse(request)
		}
	}

	alaPredictContainer.InitialResource = initialResource
	return recommendations, initialResource
}

func isAlamedaPod(alaK8sCtrAnnoStr, podUid string) bool {
	akcMap := alamedaresource.GetDefaultAlamedaK8SControllerAnno()
	err := json.Unmarshal([]byte(alaK8sCtrAnnoStr), akcMap)
	if err != nil {
		return false
	}
	for _, deployment := range akcMap.DeploymentMap {
		if _, ok := deployment.PodMap[podUid]; ok {
			return true
		} else {
			scope.Infof(fmt.Sprintf("Pod %s does not belong to K8s controller %s", podUid, alaK8sCtrAnnoStr))
		}
	}
	return false
}

func buildMetircQuery(req *operator_v1alpha1.ListMetricsRequest) metrics.Query {

	var q = metrics.Query{}

	switch req.GetMetricType() {
	case operator_v1alpha1.MetricType_CONTAINER_CPU_USAGE_TOTAL:
		q.Metric = metrics.MetricTypeContainerCPUUsageTotal
	case operator_v1alpha1.MetricType_CONTAINER_CPU_USAGE_TOTAL_RATE:
		q.Metric = metrics.MetricTypeContainerCPUUsageTotalRate
	case operator_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE:
		q.Metric = metrics.MetricTypeContainerMemoryUsage
	}

	for _, labelSelector := range req.GetConditions() {

		k := labelSelector.GetKey()
		v := labelSelector.GetValue()
		var op metrics.StringOperator
		switch labelSelector.GetOp() {
		case operator_v1alpha1.StrOp_Equal:
			op = metrics.StringOperatorEqueal
		case operator_v1alpha1.StrOp_NotEqual:
			op = metrics.StringOperatorNotEqueal
		}

		q.LabelSelectors = append(q.LabelSelectors, metrics.LabelSelector{Key: k, Op: op, Value: v})
	}

	// assign difference type of time to query instance by type of gRPC request time
	switch req.TimeSelector.(type) {
	case nil:
		q.TimeSelector = nil
	case *operator_v1alpha1.ListMetricsRequest_Time:
		q.TimeSelector = &metrics.Timestamp{T: time.Unix(req.GetTime().GetSeconds(), int64(req.GetTime().GetNanos()))}
	case *operator_v1alpha1.ListMetricsRequest_Duration:
		d, _ := ptypes.Duration(req.GetDuration())
		q.TimeSelector = &metrics.Since{
			Duration: d,
		}
	case *operator_v1alpha1.ListMetricsRequest_TimeRange:
		startTime := req.GetTimeRange().GetStartTime()
		endTime := req.GetTimeRange().GetEndTime()
		step, _ := ptypes.Duration(req.GetTimeRange().GetStep())
		q.TimeSelector = &metrics.TimeRange{
			StartTime: time.Unix(startTime.GetSeconds(), int64(startTime.GetNanos())),
			EndTime:   time.Unix(endTime.GetSeconds(), int64(endTime.GetNanos())),
			Step:      step,
		}
	}

	return q
}

func convertMetricsQueryResponseToProtoResponse(resp *metrics.QueryResponse) *operator_v1alpha1.ListMetricsResponse {

	// initiallize proto response
	ListMetricssResponse := &operator_v1alpha1.ListMetricsResponse{}
	ListMetricssResponse.Metrics = []*operator_v1alpha1.MetricResult{}

	for _, result := range resp.Results {
		series := &operator_v1alpha1.MetricResult{}

		series.Labels = result.Labels
		for _, sample := range result.Samples {
			s := &operator_v1alpha1.Sample{}

			timestampProto, err := ptypes.TimestampProto(sample.Time)
			if err != nil {
				scope.Error("convert time.Time to google.protobuf.Timestamp failed")
			}
			s.Time = timestampProto
			s.Value = sample.Value
			series.Samples = append(series.Samples, s)
		}
		ListMetricssResponse.Metrics = append(ListMetricssResponse.Metrics, series)
	}

	return ListMetricssResponse
}
