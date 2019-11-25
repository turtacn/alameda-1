package eviction

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	"github.com/containers-ai/alameda/pkg/utils"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_client "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_events "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
	datahub_recommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	openshift_apps_v1 "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	apps_v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scope = logUtil.RegisterScope("evictioner", "alamedascaler evictioner", 0)
)

// Evictioner deletes pods which need to apply recommendation
type Evictioner struct {
	checkCycle              int64
	datahubClnt             datahub_client.DatahubServiceClient
	k8sClienit              client.Client
	evictCfg                Config
	purgeContainerCPUMemory bool

	clusterID string
}

// NewEvictioner return Evictioner instance
func NewEvictioner(checkCycle int64,
	datahubClnt datahub_client.DatahubServiceClient,
	k8sClienit client.Client,
	evictCfg Config,
	purgeContainerCPUMemory bool,
	clusterID string) *Evictioner {
	return &Evictioner{
		checkCycle:              checkCycle,
		datahubClnt:             datahubClnt,
		k8sClienit:              k8sClienit,
		evictCfg:                evictCfg,
		purgeContainerCPUMemory: purgeContainerCPUMemory,
		clusterID:               clusterID,
	}
}

// Start checking pods need to apply recommendation
func (evictioner *Evictioner) Start() {
	go evictioner.evictProcess()
}

func (evictioner *Evictioner) evictProcess() {
	for {
		if !evictioner.evictCfg.Enable {
			scope.Warn("Evictioner is not enabled")
			return
		}
		appliablePodRecList, err := evictioner.listAppliablePodRecommendation()
		if err != nil {
			scope.Errorf("List appliable PodRecommendation failed: %s", err.Error())
		}
		scope.Debugf("Applicable pod recommendation lists: %s", utils.InterfaceToString(appliablePodRecList))
		evictioner.evictPods(appliablePodRecList)
		time.Sleep(time.Duration(evictioner.checkCycle) * time.Second)
	}
}

func (evictioner *Evictioner) evictPods(recommendations []*datahub_recommendations.PodRecommendation) {

	events := make([]*datahub_events.Event, 0, len(recommendations))

	for _, recommendation := range recommendations {

		if recommendation.ObjectMeta == nil {
			continue
		}

		pod := &corev1.Pod{}
		err := evictioner.k8sClienit.Get(context.TODO(), types.NamespacedName{
			Namespace: recommendation.ObjectMeta.GetNamespace(),
			Name:      recommendation.ObjectMeta.GetName(),
		}, pod)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				scope.Errorf("Get Pod(%s/%s) failed: %s", recommendation.ObjectMeta.GetNamespace(), recommendation.ObjectMeta.GetName(), err.Error())
			}
			continue
		}
		if evictioner.purgeContainerCPUMemory {
			topController := recommendation.TopController
			if topController == nil || topController.ObjectMeta == nil {
				scope.Errorf("Purge pod (%s,%s) resources failed: get empty topController from PodRecommendation", pod.GetNamespace(), pod.GetName())
				continue

			}
			topControllerNamespace := topController.ObjectMeta.Namespace
			topControllerName := topController.ObjectMeta.Name
			topControllerKind := topController.Kind
			topControllerInstance, err := evictioner.getTopController(topControllerNamespace, topControllerName, topControllerKind)
			if err != nil {
				scope.Errorf("Purge pod (%s,%s) resources failed: get topController failed: %s", pod.GetNamespace(), pod.GetName(), err.Error())
				continue
			}
			if needToPurge, err := evictioner.needToPurgeTopControllerContainerResources(topControllerInstance, topControllerKind); err != nil {
				scope.Errorf("Purge pod (%s,%s) resources failed: %s", pod.GetNamespace(), pod.GetName(), err.Error())
			} else if needToPurge {
				if err = evictioner.purgeTopControllerContainerResources(topControllerInstance, topControllerKind); err != nil {
					scope.Errorf("Purge pod (%s,%s) resources failed: %s", pod.GetNamespace(), pod.GetName(), err.Error())
					continue
				}
			} else {
				err = evictioner.k8sClienit.Delete(context.TODO(), pod)
				if err != nil {
					scope.Errorf("Evict pod (%s,%s) failed: %s", pod.GetNamespace(), pod.GetName(), err.Error())
				} else {
					e := newPodEvictEvent(evictioner.clusterID, &pod.ObjectMeta, pod.TypeMeta)
					events = append(events, &e)
				}
			}
		} else {
			err = evictioner.k8sClienit.Delete(context.TODO(), pod)
			if err != nil {
				scope.Errorf("Evict pod (%s,%s) failed: %s", pod.GetNamespace(), pod.GetName(), err.Error())
			} else {
				e := newPodEvictEvent(evictioner.clusterID, &pod.ObjectMeta, pod.TypeMeta)
				events = append(events, &e)
			}
		}
	}

	if err := evictioner.sendEvents(events); err != nil {
		scope.Warnf("Send events to datahub failed: %s\n", err.Error())
	}
}

func (evictioner *Evictioner) getTopController(namespace string, name string, kind datahub_resources.Kind) (interface{}, error) {

	getResource := utilsresource.NewGetResource(evictioner.k8sClienit)

	switch kind {
	case datahub_resources.Kind_DEPLOYMENT:
		return getResource.GetDeployment(namespace, name)
	case datahub_resources.Kind_DEPLOYMENTCONFIG:
		return getResource.GetDeploymentConfig(namespace, name)
	case datahub_resources.Kind_STATEFULSET:
		return getResource.GetStatefulSet(namespace, name)
	default:
		return nil, errors.Errorf("not supported controller type %s", datahub_resources.Kind_name[int32(kind)])
	}
}

func (evictioner *Evictioner) needToPurgeTopControllerContainerResources(controller interface{}, kind datahub_resources.Kind) (bool, error) {

	switch kind {
	case datahub_resources.Kind_DEPLOYMENT:
		deployment := controller.(*apps_v1.Deployment)
		for _, container := range deployment.Spec.Template.Spec.Containers {
			resourceLimits := container.Resources.Limits
			if resourceLimits != nil {
				_, cpuSpecExist := resourceLimits[corev1.ResourceCPU]
				_, memorySpecExist := resourceLimits[corev1.ResourceMemory]
				if cpuSpecExist || memorySpecExist {
					return true, nil
				}
			}
			resourceRequests := container.Resources.Requests
			if resourceRequests != nil {
				_, cpuSpecExist := resourceRequests[corev1.ResourceCPU]
				_, memorySpecExist := resourceRequests[corev1.ResourceMemory]
				if cpuSpecExist || memorySpecExist {
					return true, nil
				}
			}
		}
		return false, nil
	case datahub_resources.Kind_DEPLOYMENTCONFIG:
		deploymentConfig := controller.(*openshift_apps_v1.DeploymentConfig)
		for _, container := range deploymentConfig.Spec.Template.Spec.Containers {
			resourceLimits := container.Resources.Limits
			if resourceLimits != nil {
				_, cpuSpecExist := resourceLimits[corev1.ResourceCPU]
				_, memorySpecExist := resourceLimits[corev1.ResourceMemory]
				if cpuSpecExist || memorySpecExist {
					return true, nil
				}
			}
			resourceRequests := container.Resources.Requests
			if resourceRequests != nil {
				_, cpuSpecExist := resourceRequests[corev1.ResourceCPU]
				_, memorySpecExist := resourceRequests[corev1.ResourceMemory]
				if cpuSpecExist || memorySpecExist {
					return true, nil
				}
			}
		}
		return false, nil
	case datahub_resources.Kind_STATEFULSET:
		statefulSet := controller.(*apps_v1.StatefulSet)
		for _, container := range statefulSet.Spec.Template.Spec.Containers {
			resourceLimits := container.Resources.Limits
			if resourceLimits != nil {
				_, cpuSpecExist := resourceLimits[corev1.ResourceCPU]
				_, memorySpecExist := resourceLimits[corev1.ResourceMemory]
				if cpuSpecExist || memorySpecExist {
					return true, nil
				}
			}
			resourceRequests := container.Resources.Requests
			if resourceRequests != nil {
				_, cpuSpecExist := resourceRequests[corev1.ResourceCPU]
				_, memorySpecExist := resourceRequests[corev1.ResourceMemory]
				if cpuSpecExist || memorySpecExist {
					return true, nil
				}
			}
		}
		return false, nil
	default:
		return false, errors.Errorf("not supported controller type %s", datahub_resources.Kind_name[int32(kind)])
	}
}

func (evictioner *Evictioner) purgeTopControllerContainerResources(controller interface{}, kind datahub_resources.Kind) error {

	switch kind {
	case datahub_resources.Kind_DEPLOYMENT:
		deployment := controller.(*apps_v1.Deployment)
		deploymentCopy := deployment.DeepCopy()
		for _, container := range deploymentCopy.Spec.Template.Spec.Containers {
			resourceLimits := container.Resources.Limits
			if resourceLimits != nil {
				delete(resourceLimits, corev1.ResourceCPU)
				delete(resourceLimits, corev1.ResourceMemory)
			}
			resourceRequests := container.Resources.Requests
			if resourceRequests != nil {
				delete(resourceRequests, corev1.ResourceCPU)
				delete(resourceRequests, corev1.ResourceMemory)
			}
		}
		ctx := context.TODO()
		err := evictioner.k8sClienit.Update(ctx, deploymentCopy)
		if err != nil {
			return errors.Wrapf(err, "purge topController failed: %s", err.Error())
		}
		return nil
	case datahub_resources.Kind_DEPLOYMENTCONFIG:
		deploymentConfig := controller.(*openshift_apps_v1.DeploymentConfig)
		deploymentConfigCopy := deploymentConfig.DeepCopy()
		for _, container := range deploymentConfigCopy.Spec.Template.Spec.Containers {
			resourceLimits := container.Resources.Limits
			if resourceLimits != nil {
				delete(resourceLimits, corev1.ResourceCPU)
				delete(resourceLimits, corev1.ResourceMemory)
			}
			resourceRequests := container.Resources.Requests
			if resourceRequests != nil {
				delete(resourceRequests, corev1.ResourceCPU)
				delete(resourceRequests, corev1.ResourceMemory)
			}
		}
		ctx := context.TODO()
		err := evictioner.k8sClienit.Update(ctx, deploymentConfigCopy)
		if err != nil {
			return errors.Wrapf(err, "purge topController failed: %s", err.Error())
		}
		return nil
	case datahub_resources.Kind_STATEFULSET:
		statefulSet := controller.(*apps_v1.StatefulSet)
		statefulSetCopy := statefulSet.DeepCopy()
		for _, container := range statefulSetCopy.Spec.Template.Spec.Containers {
			resourceLimits := container.Resources.Limits
			if resourceLimits != nil {
				delete(resourceLimits, corev1.ResourceCPU)
				delete(resourceLimits, corev1.ResourceMemory)
			}
			resourceRequests := container.Resources.Requests
			if resourceRequests != nil {
				delete(resourceRequests, corev1.ResourceCPU)
				delete(resourceRequests, corev1.ResourceMemory)
			}
		}
		ctx := context.TODO()
		err := evictioner.k8sClienit.Update(ctx, statefulSetCopy)
		if err != nil {
			return errors.Wrapf(err, "purge topController failed: %s", err.Error())
		}
		return nil
	default:
		return errors.Errorf("not supported controller type %s", datahub_resources.Kind_name[int32(kind)])
	}
}

func (evictioner *Evictioner) listAppliablePodRecommendation() ([]*datahub_recommendations.PodRecommendation, error) {

	appliablePodRecList := []*datahub_recommendations.PodRecommendation{}
	nowTime := time.Now()
	nowTimestamp := time.Now().Unix()

	recommendations, err := evictioner.listPodRecommendations(nowTimestamp)
	if err != nil {
		return nil, errors.Wrap(err, "list pod recommendations failed")
	}
	controllerRecommendationInfoMap := NewControllerRecommendationInfoMap(evictioner.k8sClienit, recommendations)
	for _, controllerRecommendationInfo := range controllerRecommendationInfoMap {
		podRecommendationInfos := controllerRecommendationInfo.podRecommendationInfos
		sort.Slice(podRecommendationInfos, func(i, j int) bool {
			return podRecommendationInfos[i].pod.ObjectMeta.CreationTimestamp.UnixNano() < podRecommendationInfos[j].pod.ObjectMeta.CreationTimestamp.UnixNano()
		})
	}
	for _, controllerRecommendationInfo := range controllerRecommendationInfoMap {

		// Create eviction restriction
		maxUnavailable := controllerRecommendationInfo.getMaxUnavailable()
		triggerThreshold, err := controllerRecommendationInfo.buildTriggerThreshold()
		if err != nil {
			scope.Errorf("Build triggerThreshold of controller (%s/%s, kind: %s) faild, skip evicting controller's pod: %s",
				controllerRecommendationInfo.namespace, controllerRecommendationInfo.name, controllerRecommendationInfo.kind, err.Error())
			continue
		}
		podRecommendations := make([]*datahub_recommendations.PodRecommendation, len(controllerRecommendationInfo.podRecommendationInfos))
		for i := range controllerRecommendationInfo.podRecommendationInfos {
			podRecommendations[i] = controllerRecommendationInfo.podRecommendationInfos[i].recommendation
		}
		evictionRestriction := NewEvictionRestriction(evictioner.k8sClienit, maxUnavailable, triggerThreshold, podRecommendations)

		for _, podRecommendationInfo := range controllerRecommendationInfo.podRecommendationInfos {
			pod := podRecommendationInfo.pod
			podRecommendation := podRecommendationInfo.recommendation
			if !controllerRecommendationInfo.isScalingToolTypeVPA() {
				scope.Infof("Pod (%s/%s) cannot be evicted due to AlamedaScaler's scaling tool is type of %s",
					pod.GetNamespace(), pod.GetName(), controllerRecommendationInfo.alamedaScaler.Spec.ScalingTool.Type)
				continue
			}
			if ok, err := podRecommendationInfo.isApplicableAtTime(nowTime); err != nil {
				scope.Infof("Pod (%s/%s) cannot be evicted due to PodRecommendation validate error, %s",
					pod.GetNamespace(), pod.GetName(), err.Error())
				continue
			} else if !ok {
				scope.Infof("Pod (%s/%s) cannot be evicted due to current time (%d) is not applicable on PodRecommendation's startTime (%d) and endTime(%d) interval",
					pod.GetNamespace(), pod.GetName(), nowTime.Unix(), podRecommendation.GetStartTime().GetSeconds(), podRecommendation.GetEndTime().GetSeconds())
				continue
			}
			if isEvictabel, err := evictionRestriction.IsEvictabel(pod); err != nil {
				scope.Infof("Pod (%s/%s) cannot be evicted due to eviction restriction checking error: %s", pod.GetNamespace(), pod.GetName(), err.Error())
				continue
			} else if !isEvictabel {
				scope.Infof("Pod (%s/%s) cannot be evicted.", pod.GetNamespace(), pod.GetName())
				continue
			} else {
				scope.Infof("Pod (%s/%s) can be evicted.", pod.GetNamespace(), pod.GetName())
				appliablePodRecList = append(appliablePodRecList, podRecommendation)
			}
		}
	}

	return appliablePodRecList, nil
}

func (evictioner *Evictioner) listPodRecommendations(nowTimestamp int64) ([]*datahub_recommendations.PodRecommendation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	in := &datahub_recommendations.ListPodRecommendationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: evictioner.clusterID,
			},
		},
		Kind: datahub_resources.Kind_KIND_UNDEFINED,
		QueryCondition: &datahub_common.QueryCondition{
			TimeRange: &datahub_common.TimeRange{
				ApplyTime: &timestamp.Timestamp{
					Seconds: nowTimestamp,
				},
			},
			Order: datahub_common.QueryCondition_DESC,
			Limit: 1,
		},
	}
	scope.Debugf("Request of ListAvailablePodRecommendations is %s.", utils.InterfaceToString(in))
	resp, err := evictioner.datahubClnt.ListAvailablePodRecommendations(ctx, in)
	if err != nil {
		return nil, err
	} else if resp == nil || resp.Status == nil {
		return nil, fmt.Errorf("receive nil status from datahub")
	} else if resp.Status.Code != int32(code.Code_OK) {
		return nil, fmt.Errorf("status code not 0: receive status code: %d, message: %s", resp.GetStatus().GetCode(), resp.GetStatus().GetMessage())
	}

	return resp.GetPodRecommendations(), nil
}

func (evictioner *Evictioner) sendEvents(events []*datahub_events.Event) error {

	if len(events) == 0 {
		return nil
	}

	request := datahub_events.CreateEventsRequest{
		Events: events,
	}
	status, err := evictioner.datahubClnt.CreateEvents(context.TODO(), &request)
	if err != nil {
		return errors.Errorf("send events to Datahub failed: %s", err.Error())
	} else if status == nil {
		return errors.Errorf("send events to Datahub failed: receive nil status")
	} else if status.Code != int32(code.Code_OK) {
		return errors.Errorf("send events to Datahub failed: statusCode: %d, message: %s", status.Code, status.Message)
	}

	return nil
}

type podRecommendationInfo struct {
	pod            *corev1.Pod
	recommendation *datahub_recommendations.PodRecommendation
}

func (p *podRecommendationInfo) isApplicableAtTime(t time.Time) (bool, error) {

	startTime := p.recommendation.GetStartTime()
	endTime := p.recommendation.GetEndTime()

	if startTime == nil || endTime == nil {
		return false, errors.Errorf("starTime and endTime cannot be nil")
	}

	if startTime.GetSeconds() >= t.Unix() || t.Unix() >= endTime.GetSeconds() {
		return false, nil
	}

	return true, nil
}

func (p *podRecommendationInfo) podRunsLongerThan(d time.Duration) bool {
	now := time.Now()
	podCreationTime := p.pod.CreationTimestamp.Time
	return podCreationTime.Add(d).Before(now)
}

type controllerRecommendationInfo struct {
	namespace              string
	name                   string
	kind                   string
	alamedaScaler          *autoscalingv1alpha1.AlamedaScaler
	podRecommendationInfos []*podRecommendationInfo
}

func (c controllerRecommendationInfo) getMaxUnavailable() string {

	var maxUnavailable string

	scalingTool := c.alamedaScaler.Spec.ScalingTool
	if scalingTool.ExecutionStrategy == nil {
		maxUnavailable = autoscalingv1alpha1.DefaultMaxUnavailablePercentage
		return maxUnavailable
	}

	maxUnavailable = scalingTool.ExecutionStrategy.MaxUnavailable
	return maxUnavailable
}

func (c controllerRecommendationInfo) isScalingToolTypeVPA() bool {
	return c.alamedaScaler.IsScalingToolTypeVPA()
}

func (c controllerRecommendationInfo) buildTriggerThreshold() (triggerThreshold, error) {

	var triggerThreshold triggerThreshold

	cpu := c.alamedaScaler.Spec.ScalingTool.ExecutionStrategy.TriggerThreshold.CPU
	cpu = strings.TrimSuffix(cpu, "%")
	cpuValue, err := strconv.ParseFloat(cpu, 64)
	if err != nil {
		return triggerThreshold, errors.Errorf("parse cpu trigger threshold failed: %s", err.Error())
	}
	triggerThreshold.CPU = cpuValue

	memory := c.alamedaScaler.Spec.ScalingTool.ExecutionStrategy.TriggerThreshold.Memory
	memory = strings.TrimSuffix(memory, "%")
	memoryValue, err := strconv.ParseFloat(memory, 64)
	if err != nil {
		return triggerThreshold, errors.Errorf("parse memory trigger threshold failed: %s", err.Error())
	}
	triggerThreshold.Memory = memoryValue

	return triggerThreshold, nil
}

// NewControllerRecommendationInfoMap returns
func NewControllerRecommendationInfoMap(client client.Client, podRecommendations []*datahub_recommendations.PodRecommendation) map[string]*controllerRecommendationInfo {

	getResource := utilsresource.NewGetResource(client)
	alamedaScalerMap := make(map[string]*autoscalingv1alpha1.AlamedaScaler)
	controllerRecommendationInfoMap := make(map[string]*controllerRecommendationInfo)
	for _, podRecommendation := range podRecommendations {

		// Filter out invalid PodRecommendation
		copyPodRecommendation := proto.Clone(podRecommendation)
		podRecommendation = copyPodRecommendation.(*datahub_recommendations.PodRecommendation)
		recommendationObjectMeta := podRecommendation.ObjectMeta
		if recommendationObjectMeta == nil {
			scope.Errorf("Skip PodRecommendation due to PodRecommendation has nil ObjectMeta")
			continue
		}

		// Get AlamedaScaler owns this PodRecommendation and validate the AlamedaScaler is enabled execution.
		alamedaRecommendation, err := getResource.GetAlamedaRecommendation(recommendationObjectMeta.Namespace, recommendationObjectMeta.Name)
		if err != nil {
			scope.Errorf("Skip PodRecommendation (%s/%s) due to get AlamedaRecommendation falied: %s", recommendationObjectMeta.Namespace, recommendationObjectMeta.Name, err.Error())
			continue
		}
		alamedaScalerNamespace := ""
		alamedaScalerName := ""
		for _, or := range alamedaRecommendation.OwnerReferences {
			if or.Kind == "AlamedaScaler" {
				alamedaScalerNamespace = alamedaRecommendation.Namespace
				alamedaScalerName = or.Name
				break
			}
		}
		alamedaScaler, exist := alamedaScalerMap[fmt.Sprintf("%s/%s", alamedaScalerNamespace, alamedaScalerName)]
		if !exist {
			alamedaScaler, err = getResource.GetAlamedaScaler(alamedaScalerNamespace, alamedaScalerName)
			if err != nil {
				scope.Errorf("Skip PodRecommendation (%s/%s) due to get AlamedaScaler falied: %s", recommendationObjectMeta.Namespace, recommendationObjectMeta.Name, err.Error())
				continue
			}
			alamedaScalerMap[fmt.Sprintf("%s/%s", alamedaScalerNamespace, alamedaScalerName)] = alamedaScaler
		}
		if !alamedaScaler.IsEnableExecution() {
			scope.Errorf("Skip PodRecommendation (%s/%s) because it's execution is not enabled.", recommendationObjectMeta.Namespace, recommendationObjectMeta.Name)
			continue
		}

		// Get Pod instance of this PodRecommendation
		podNamespace := recommendationObjectMeta.Namespace
		podName := recommendationObjectMeta.Name
		pod, err := getResource.GetPod(podNamespace, podName)
		if err != nil {
			scope.Errorf("Skip PodRecommendation due to get Pod (%s/%s) failed: %s", podNamespace, podName, err.Error())
			continue
		}

		// Get topmost controller namespace, name and kind controlling this pod
		controller := podRecommendation.TopController
		if controller == nil {
			scope.Errorf("Skip PodRecommendation (%s/%s) due to PodRecommendation has nil topmost controller", podNamespace, podName)
			continue
		} else if controller.ObjectMeta == nil {
			scope.Errorf("Skip PodRecommendation (%s/%s) due to topmost controller has nil ObjectMeta", podNamespace, podName)
			continue
		}

		// Append podRecommendationInfos into controllerRecommendationInfo
		controllerID := fmt.Sprintf("%s.%s.%s", controller.Kind, controller.ObjectMeta.Namespace, controller.ObjectMeta.Name)
		_, exist = controllerRecommendationInfoMap[controllerID]
		if !exist {
			controllerRecommendationInfoMap[controllerID] = &controllerRecommendationInfo{
				namespace:              controller.ObjectMeta.Namespace,
				name:                   controller.ObjectMeta.Name,
				kind:                   datahub_resources.Kind_name[int32(controller.Kind)],
				alamedaScaler:          alamedaScaler,
				podRecommendationInfos: make([]*podRecommendationInfo, 0),
			}
		}
		podRecommendationInfo := &podRecommendationInfo{
			pod:            pod,
			recommendation: podRecommendation,
		}
		controllerRecommendationInfoMap[controllerID].podRecommendationInfos = append(
			controllerRecommendationInfoMap[controllerID].podRecommendationInfos,
			podRecommendationInfo,
		)
	}

	return controllerRecommendationInfoMap
}
