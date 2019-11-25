package datahub

import (
	"context"
	"math"
	"strconv"

	"github.com/containers-ai/alameda/admission-controller/pkg/recommendator/resource"
	"github.com/containers-ai/alameda/pkg/framework/datahub"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_client "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_recommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	core_v1 "k8s.io/api/core/v1"
	k8s_resource "k8s.io/apimachinery/pkg/api/resource"
)

var (
	scope               = log.RegisterScope("resource-recommendator", "Datahub resource recommendator", 0)
	k8sKind_DatahubKind = map[string]datahub_resources.Kind{
		"Undefined":        datahub_resources.Kind_KIND_UNDEFINED,
		"Deployment":       datahub_resources.Kind_DEPLOYMENT,
		"DeploymentConfig": datahub_resources.Kind_DEPLOYMENTCONFIG,
		"StatefulSet":      datahub_resources.Kind_STATEFULSET,
	}
	datahubKind_K8SKind = map[datahub_resources.Kind]string{
		datahub_resources.Kind_KIND_UNDEFINED:   "Undefined",
		datahub_resources.Kind_DEPLOYMENT:       "Deployment",
		datahub_resources.Kind_DEPLOYMENTCONFIG: "DeploymentConfig",
		datahub_resources.Kind_STATEFULSET:      "StatefulSet",
	}
	datahubMetricType_K8SResourceName = map[datahub_common.MetricType]core_v1.ResourceName{
		datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE: core_v1.ResourceCPU,
		datahub_common.MetricType_MEMORY_USAGE_BYTES:           core_v1.ResourceMemory,
	}
)

var _ resource.ResourceRecommendator = &datahubResourceRecommendator{}

type datahubResourceRecommendator struct {
	datahubServiceClient datahub_client.DatahubServiceClient
	clusterName          string
}

func NewDatahubResourceRecommendator(client datahub_client.DatahubServiceClient, clusterName string) (resource.ResourceRecommendator, error) {

	return &datahubResourceRecommendator{
		datahubServiceClient: client,
		clusterName:          clusterName,
	}, nil
}

func (dr *datahubResourceRecommendator) ListControllerPodResourceRecommendations(req resource.ListControllerPodResourceRecommendationsRequest) ([]*resource.PodResourceRecommendation, error) {

	recommendations := make([]*resource.PodResourceRecommendation, 0)

	datahubRequest, err := dr.buildListAvailablePodRecommendationsRequest(req)
	if err != nil {
		return recommendations, errors.Wrap(err, "list controller pod resource recommendations failed")
	}
	scope.Debugf("query ListAvailablePodRecommendations to datahub, send request: %+v", datahubRequest)
	resp, err := dr.datahubServiceClient.ListAvailablePodRecommendations(context.Background(), datahubRequest)
	scope.Debugf("query ListAvailablePodRecommendations to datahub, received response: %+v", resp)
	if err != nil {
		return recommendations, errors.Wrap(err, "list controller pod resource recommendations failed")
	} else if _, err := datahub.IsResponseStatusOK(resp.Status); err != nil {
		return recommendations, errors.Wrap(err, "list controller pod resource recommendations failed")
	}

	for _, datahubPodRecommendation := range resp.GetPodRecommendations() {
		podRecommendation := buildPodResourceRecommendationFromDatahubPodRecommendation(datahubPodRecommendation)
		recommendations = append(recommendations, podRecommendation)
	}

	return recommendations, nil
}

func (dr *datahubResourceRecommendator) buildListAvailablePodRecommendationsRequest(request resource.ListControllerPodResourceRecommendationsRequest) (*datahub_recommendations.ListPodRecommendationsRequest, error) {

	var datahubRequest *datahub_recommendations.ListPodRecommendationsRequest

	datahubKind, exist := k8sKind_DatahubKind[request.Kind]
	if !exist {
		return datahubRequest, errors.Errorf("build Datahub ListPodRecommendationsRequest failed: no mapping Datahub kind for k8s kind: %s", request.Kind)
	}

	var queryTime *timestamp.Timestamp
	var err error
	if request.Time != nil {
		queryTime, err = ptypes.TimestampProto(*request.Time)
		if err != nil {
			return datahubRequest, errors.Errorf("build Datahub ListPodRecommendationsRequest failed: convert time.Time to google.Timestamp failed: %s", err.Error())
		}
	}

	datahubRequest = &datahub_recommendations.ListPodRecommendationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: dr.clusterName,
				Namespace:   request.Namespace,
				Name:        request.Name,
			},
		},
		Kind: datahubKind,
		QueryCondition: &datahub_common.QueryCondition{
			TimeRange: &datahub_common.TimeRange{
				ApplyTime: queryTime,
			},
			Order: datahub_common.QueryCondition_DESC,
			Limit: 1,
		},
	}
	return datahubRequest, nil
}

func buildPodResourceRecommendationFromDatahubPodRecommendation(datahubPodRecommendation *datahub_recommendations.PodRecommendation) *resource.PodResourceRecommendation {

	namespace := ""
	name := ""
	if objectMeta := datahubPodRecommendation.ObjectMeta; objectMeta != nil {
		namespace = objectMeta.Namespace
		name = objectMeta.Name
	}

	startTime, _ := ptypes.Timestamp(datahubPodRecommendation.GetStartTime())
	endTime, _ := ptypes.Timestamp(datahubPodRecommendation.GetEndTime())

	topControllerKind := ""
	topControllerName := ""
	if datahubPodRecommendation.TopController != nil {
		topControllerKind = datahubKind_K8SKind[datahubPodRecommendation.TopController.Kind]
		if datahubPodRecommendation.TopController.ObjectMeta != nil {
			topControllerName = datahubPodRecommendation.TopController.ObjectMeta.Name
		}
	}

	podRecommendation := &resource.PodResourceRecommendation{
		Namespace:                        namespace,
		Name:                             name,
		TopControllerKind:                topControllerKind,
		TopControllerName:                topControllerName,
		ContainerResourceRecommendations: make([]*resource.ContainerResourceRecommendation, 0),
		ValidStartTime:                   startTime,
		ValidEndTime:                     endTime,
	}
	for _, datahubContainerRecommendation := range datahubPodRecommendation.GetContainerRecommendations() {
		containerResourceRecommendation := buildContainerResourceRecommendationFromDatahubContainerRecommendation(datahubContainerRecommendation)
		podRecommendation.ContainerResourceRecommendations = append(podRecommendation.ContainerResourceRecommendations, containerResourceRecommendation)
	}

	return podRecommendation
}

func buildContainerResourceRecommendationFromDatahubContainerRecommendation(datahubContainerRecommendation *datahub_recommendations.ContainerRecommendation) *resource.ContainerResourceRecommendation {

	containerResourceRecommendation := &resource.ContainerResourceRecommendation{
		Name: datahubContainerRecommendation.Name,
	}

	resourceLimitMap := datahubMetricDataSliceToMetricTypeValueMap(datahubContainerRecommendation.GetLimitRecommendations())
	containerResourceRecommendation.Limits = buildK8SReosurceListFromMetricTypeValueMap(resourceLimitMap)

	resourceRequestMap := datahubMetricDataSliceToMetricTypeValueMap(datahubContainerRecommendation.GetRequestRecommendations())
	containerResourceRecommendation.Requests = buildK8SReosurceListFromMetricTypeValueMap(resourceRequestMap)

	return containerResourceRecommendation
}

func datahubMetricDataSliceToMetricTypeValueMap(metricDataSlice []*datahub_common.MetricData) map[datahub_common.MetricType]string {

	resourceMap := make(map[datahub_common.MetricType]string)

	for _, metricData := range metricDataSlice {
		sample := choseOneSample(metricData.GetData())
		if sample != nil {
			resourceMap[metricData.MetricType] = sample.NumValue
		}
	}

	return resourceMap
}

func choseOneSample(samples []*datahub_common.Sample) *datahub_common.Sample {

	if len(samples) > 0 {
		return samples[0]
	} else {
		return nil
	}
}

func buildK8SReosurceListFromMetricTypeValueMap(metricTypeValueMap map[datahub_common.MetricType]string) core_v1.ResourceList {

	resourceList := make(core_v1.ResourceList)

	for metricType, value := range metricTypeValueMap {

		resourceUnit := ""
		if metricType == datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE {
			cpuMilliCores, err := strconv.ParseFloat(value, 64)
			if err != nil {

			}
			cpuMilliCores = math.Ceil(cpuMilliCores)
			value = strconv.FormatFloat(cpuMilliCores, 'f', 0, 64)
			resourceUnit = "m"
		}
		value = value + resourceUnit

		quantity, err := k8s_resource.ParseQuantity(value)
		if err != nil {
			scope.Warnf("parse value to k8s resource.Quantity failed, skip this recommendation: metricType:%s, value: %s, errMsg: %s", datahub_common.MetricType_name[int32(metricType)], value, err.Error())
			continue
		}

		if k8sResourceName, exist := datahubMetricType_K8SResourceName[metricType]; !exist {
			scope.Warnf("no mapping k8s core_v1.ResourceName found for Datahub MetricType, skip this recommendation: metricType: %d", metricType)
			continue
		} else {
			resourceList[k8sResourceName] = quantity
		}
	}

	return resourceList
}
