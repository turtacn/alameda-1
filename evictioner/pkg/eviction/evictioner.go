package eviction

import (
	"context"
	"fmt"
	"sort"
	"time"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	"github.com/containers-ai/alameda/pkg/consts"
	"github.com/containers-ai/alameda/pkg/utils"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	openshift_apps_v1 "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	apps_v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
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
	datahubClnt             datahub_v1alpha1.DatahubServiceClient
	k8sClienit              client.Client
	evictCfg                Config
	purgeContainerCPUMemory bool
}

// NewEvictioner return Evictioner instance
func NewEvictioner(checkCycle int64,
	datahubClnt datahub_v1alpha1.DatahubServiceClient,
	k8sClienit client.Client,
	evictCfg Config,
	purgeContainerCPUMemory bool) *Evictioner {
	return &Evictioner{
		checkCycle:              checkCycle,
		datahubClnt:             datahubClnt,
		k8sClienit:              k8sClienit,
		evictCfg:                evictCfg,
		purgeContainerCPUMemory: purgeContainerCPUMemory,
	}
}

// Start checking pods need to apply recommendation
func (evictioner *Evictioner) Start() {
	go evictioner.evictProcess()
}

func (evictioner *Evictioner) evictProcess() {
	for {
		if !evictioner.evictCfg.Enable {
			scope.Warn("evictioner is not enabled")
			return
		}
		appliablePodRecList, err := evictioner.listAppliablePodRecommendation()
		if err != nil {
			scope.Error(err.Error())
		}
		scope.Debugf("Applicable pod recommendation lists: %s", utils.InterfaceToString(appliablePodRecList))
		evictioner.evictPods(appliablePodRecList)
		time.Sleep(time.Duration(evictioner.checkCycle) * time.Second)
	}
}

func (evictioner *Evictioner) evictPods(recPodList []*datahub_v1alpha1.PodRecommendation) {
	for _, recPod := range recPodList {
		recPodIns := &corev1.Pod{}
		err := evictioner.k8sClienit.Get(context.TODO(), types.NamespacedName{
			Namespace: recPod.GetNamespacedName().GetNamespace(),
			Name:      recPod.GetNamespacedName().GetName(),
		}, recPodIns)
		if err != nil {
			if !k8s_errors.IsNotFound(err) {
				scope.Error(err.Error())
			}
			continue
		}
		if evictioner.purgeContainerCPUMemory {
			topController := recPod.TopController
			if topController == nil || topController.NamespacedName == nil {
				scope.Errorf("Purge pod (%s,%s) resources failed: get empty topController from PodRecommendation", recPodIns.GetNamespace(), recPodIns.GetName())
				continue

			}
			topControllerNamespace := topController.NamespacedName.Namespace
			topControllerName := topController.NamespacedName.Name
			topControllerKind := topController.Kind
			topControllerInstance, err := evictioner.getTopController(topControllerNamespace, topControllerName, topControllerKind)
			if err != nil {
				scope.Errorf("Purge pod (%s,%s) resources failed: get topController failed: %s", recPodIns.GetNamespace(), recPodIns.GetName(), err.Error())
				continue
			}
			if needToPurge, err := evictioner.needToPurgeTopControllerContainerResources(topControllerInstance, topControllerKind); err != nil {
				scope.Errorf("Purge pod (%s,%s) resources failed: %s", recPodIns.GetNamespace(), recPodIns.GetName(), err.Error())
			} else if needToPurge {
				if err = evictioner.purgeTopControllerContainerResources(topControllerInstance, topControllerKind); err != nil {
					scope.Errorf("Purge pod (%s,%s) resources failed: %s", recPodIns.GetNamespace(), recPodIns.GetName(), err.Error())
					continue
				}
			} else {
				err = evictioner.k8sClienit.Delete(context.TODO(), recPodIns)
				if err != nil {
					scope.Errorf("Evict pod (%s,%s) failed: %s", recPodIns.GetNamespace(), recPodIns.GetName(), err.Error())
				}
			}
		} else {
			err = evictioner.k8sClienit.Delete(context.TODO(), recPodIns)
			if err != nil {
				scope.Errorf("Evict pod (%s,%s) failed: %s", recPodIns.GetNamespace(), recPodIns.GetName(), err.Error())
			}
		}
	}
}

func (evictioner *Evictioner) getTopController(namespace string, name string, kind datahub_v1alpha1.Kind) (interface{}, error) {

	getResource := utilsresource.NewGetResource(evictioner.k8sClienit)

	switch kind {
	case datahub_v1alpha1.Kind_DEPLOYMENT:
		return getResource.GetDeployment(namespace, name)
	case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
		return getResource.GetDeploymentConfig(namespace, name)
	default:
		return nil, errors.Errorf("not supported controller type %s", datahub_v1alpha1.Kind_name[int32(kind)])
	}
}

func (evictioner *Evictioner) needToPurgeTopControllerContainerResources(controller interface{}, kind datahub_v1alpha1.Kind) (bool, error) {

	switch kind {
	case datahub_v1alpha1.Kind_DEPLOYMENT:
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
	case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
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
	default:
		return false, errors.Errorf("not supported controller type %s", datahub_v1alpha1.Kind_name[int32(kind)])
	}
}

func (evictioner *Evictioner) purgeTopControllerContainerResources(controller interface{}, kind datahub_v1alpha1.Kind) error {

	switch kind {
	case datahub_v1alpha1.Kind_DEPLOYMENT:
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
	case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
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
	default:
		return errors.Errorf("not supported controller type %s", datahub_v1alpha1.Kind_name[int32(kind)])
	}

	return nil
}

func (evictioner *Evictioner) listAppliablePodRecommendation() ([]*datahub_v1alpha1.PodRecommendation, error) {

	appliablePodRecList := []*datahub_v1alpha1.PodRecommendation{}
	nowTimestamp := time.Now().Unix()

	resp, err := evictioner.listPodRecommsPossibleToApply(nowTimestamp)
	if err != nil {
		return appliablePodRecList, err
	} else if resp.Status == nil {
		return appliablePodRecList, fmt.Errorf("Receive nil status from datahub")
	} else if resp.Status.Code != int32(code.Code_OK) {
		return appliablePodRecList, fmt.Errorf("Status code not 0: receive status code: %d,message: %s", resp.GetStatus().GetCode(), resp.GetStatus().GetMessage())
	}

	podRecommsPossibleToApply := resp.GetPodRecommendations()
	scope.Debugf("Possible applicable pod recommendation lists: %s", utils.InterfaceToString(podRecommsPossibleToApply))

	topControllerIDToPodRecommendationInfosMap := NewTopControllerIDToPodRecommendationInfosMap(evictioner.k8sClienit, podRecommsPossibleToApply)
	for _, podRecommendationInfos := range topControllerIDToPodRecommendationInfosMap {
		sort.Slice(podRecommendationInfos, func(i, j int) bool {
			return podRecommendationInfos[i].pod.ObjectMeta.CreationTimestamp.UnixNano() < podRecommendationInfos[j].pod.ObjectMeta.CreationTimestamp.UnixNano()
		})
	}
	for _, podRecommendationInfos := range topControllerIDToPodRecommendationInfosMap {

		evictionRestriction := NewEvictionRestriction(evictioner.k8sClienit, evictioner.getPreservationPercentage(), evictioner.evictCfg.TriggerThreshold, podRecommsPossibleToApply)
		enableScalerMap := map[string]bool{}
		for _, podRecommendationInfo := range podRecommendationInfos {

			rec := podRecommendationInfo.recommendation

			startTime := rec.GetStartTime().GetSeconds()
			endTime := rec.GetEndTime().GetSeconds()
			if startTime >= nowTimestamp || nowTimestamp >= endTime {
				continue
			}

			if rec.GetNamespacedName() == nil {
				scope.Warn("receive pod recommendation with nil NamespacedName, skip this recommendation")
				continue
			}

			recNS := rec.GetNamespacedName().GetNamespace()
			recName := rec.GetNamespacedName().GetName()
			pod := podRecommendationInfo.pod

			alamRecomm, err := evictioner.getAlamRecommInfo(recNS, recName)
			if err != nil {
				scope.Errorf("Get AlamedaRecommendation (%s/%s) failed due to %s.", recNS, recName, err.Error())
				continue
			}

			if !evictioner.isPodEnableExecution(alamRecomm, enableScalerMap) {
				scope.Debugf("Pod (%s/%s) cannot be evicted because its execution is not enabled.", pod.GetNamespace(), pod.GetName())
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
				appliablePodRecList = append(appliablePodRecList, rec)
			}
		}
	}

	return appliablePodRecList, nil
}

func (evictioner *Evictioner) getPreservationPercentage() float64 {
	return evictioner.evictCfg.PreservationPercentage / 100
}

func (evictioner *Evictioner) listPodRecommsPossibleToApply(nowTimestamp int64) (*datahub_v1alpha1.ListPodRecommendationsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	in := &datahub_v1alpha1.ListPodRecommendationsRequest{
		QueryCondition: &datahub_v1alpha1.QueryCondition{
			TimeRange: &datahub_v1alpha1.TimeRange{
				ApplyTime: &timestamp.Timestamp{
					Seconds: nowTimestamp,
				},
			},
			Order: datahub_v1alpha1.QueryCondition_DESC,
			Limit: 1,
		},
	}
	scope.Debugf("Request of ListAvailablePodRecommendations is %s.", utils.InterfaceToString(in))

	return evictioner.datahubClnt.ListAvailablePodRecommendations(ctx, in)
}

func (evictioner *Evictioner) getPodInfo(namespace, name string) (*corev1.Pod, error) {
	getResource := utilsresource.NewGetResource(evictioner.k8sClienit)
	pod, err := getResource.GetPod(namespace, name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			scope.Debugf(err.Error())
		} else {
			scope.Errorf(err.Error())
		}
	}
	return pod, err
}

func (evictioner *Evictioner) getAlamRecommInfo(namespace, name string) (*autoscalingv1alpha1.AlamedaRecommendation, error) {
	getResource := utilsresource.NewGetResource(evictioner.k8sClienit)
	alamRecomm, err := getResource.GetAlamedaRecommendation(namespace, name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			scope.Debugf(err.Error())
		} else {
			scope.Errorf(err.Error())
		}
	}
	return alamRecomm, err
}

func (evictioner *Evictioner) isPodEnableExecution(alamRecomm *autoscalingv1alpha1.AlamedaRecommendation, enableScalerMap map[string]bool) bool {

	for _, or := range alamRecomm.OwnerReferences {
		if or.Kind != consts.K8S_KIND_ALAMEDASCALER {
			continue
		}

		if enabled, ok := enableScalerMap[fmt.Sprintf("%s/%s", alamRecomm.GetNamespace(), or.Name)]; enabled && ok {
			return true
		} else if !enabled && ok {
			return false
		}

		scaler, err := evictioner.getAlamedaScalerInfo(alamRecomm.GetNamespace(), or.Name)
		if err == nil {
			enableScalerMap[fmt.Sprintf("%s/%s", alamRecomm.GetNamespace(), or.Name)] = scaler.Spec.EnableExecution
			return scaler.Spec.EnableExecution
		}
		return false
	}
	return false
}

func (evictioner *Evictioner) getAlamedaScalerInfo(namespace, name string) (*autoscalingv1alpha1.AlamedaScaler, error) {
	getResource := utilsresource.NewGetResource(evictioner.k8sClienit)
	scaler, err := getResource.GetAlamedaScaler(namespace, name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			scope.Debugf(err.Error())
		} else {
			scope.Errorf(err.Error())
		}
	}
	return scaler, err
}

type podRecommendationInfo struct {
	pod            *corev1.Pod
	recommendation *datahub_v1alpha1.PodRecommendation
}

func NewTopControllerIDToPodRecommendationInfosMap(client client.Client, podRecommendations []*datahub_v1alpha1.PodRecommendation) map[string][]*podRecommendationInfo {

	getResource := utilsresource.NewGetResource(client)
	topControllerIDToPodRecommendationsMap := make(map[string][]*podRecommendationInfo)
	for _, podRecommendation := range podRecommendations {

		copyPodRecommendation := proto.Clone(podRecommendation)
		podRecommendation = copyPodRecommendation.(*datahub_v1alpha1.PodRecommendation)

		recommendationNamespacedName := podRecommendation.NamespacedName
		if recommendationNamespacedName == nil {
			scope.Errorf("skip PodRecommendation due to PodRecommendation has empty NamespacedName")
			continue
		}

		topController := podRecommendation.TopController
		if topController == nil {
			scope.Errorf("skip PodRecommendation (%s/%s) due to PodRecommendation has empty topController", recommendationNamespacedName.Namespace, recommendationNamespacedName.Name)
			continue
		} else if topController.NamespacedName == nil {
			scope.Errorf("skip PodRecommendation (%s/%s) due to topController has empty NamespacedName", recommendationNamespacedName.Namespace, recommendationNamespacedName.Name)
			continue
		}

		podNamespace := recommendationNamespacedName.Namespace
		podName := recommendationNamespacedName.Name
		pod, err := getResource.GetPod(podNamespace, podName)
		if err != nil {
			scope.Errorf("skip PodRecommendation due to get Pod (%s/%s) failed: %s", podNamespace, podName, err.Error())
			continue
		}

		topControllerID := fmt.Sprintf("%s.%s.%s", topController.Kind, topController.NamespacedName.Namespace, topController.NamespacedName.Name)
		if _, exist := topControllerIDToPodRecommendationsMap[topControllerID]; !exist {
			topControllerIDToPodRecommendationsMap[topControllerID] = make([]*podRecommendationInfo, 0)
		}
		podRecommendationInfo := &podRecommendationInfo{
			pod:            pod,
			recommendation: podRecommendation,
		}
		topControllerIDToPodRecommendationsMap[topControllerID] = append(topControllerIDToPodRecommendationsMap[topControllerID], podRecommendationInfo)
	}

	return topControllerIDToPodRecommendationsMap
}
