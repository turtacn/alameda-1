package eviction

import (
	"context"
	"fmt"
	"math"
	"time"

	datahubutils "github.com/containers-ai/alameda/datahub/pkg/utils"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	"github.com/containers-ai/alameda/pkg/utils"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/rpc/code"
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
	checkCycle  int64
	datahubClnt datahub_v1alpha1.DatahubServiceClient
	k8sClienit  client.Client
	evictCfg    Config
}

// NewEvictioner return Evictioner instance
func NewEvictioner(checkCycle int64,
	datahubClnt datahub_v1alpha1.DatahubServiceClient,
	k8sClienit client.Client,
	evictCfg Config) *Evictioner {
	return &Evictioner{
		checkCycle:  checkCycle,
		datahubClnt: datahubClnt,
		k8sClienit:  k8sClienit,
		evictCfg:    evictCfg,
	}
}

// Start checking pods need to apply recommendation
func (evictioner *Evictioner) Start() {
	go evictioner.evictProcess()
}

func (evictioner *Evictioner) evictProcess() {
	for {
		if !evictioner.evictCfg.Enable {
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
		err = evictioner.k8sClienit.Delete(context.TODO(), recPodIns)
		if err != nil {
			scope.Errorf("Evict pod (%s,%s) failed: %s", recPodIns.GetNamespace(), recPodIns.GetName(), err.Error())
		}
	}
}

func (evictioner *Evictioner) listAppliablePodRecommendation() ([]*datahub_v1alpha1.PodRecommendation, error) {
	cpuTriggerThreashold := evictioner.evictCfg.TriggerThreashold.CPU
	memoryTriggerThreashold := evictioner.evictCfg.TriggerThreashold.Memory
	appliablePodRecList := []*datahub_v1alpha1.PodRecommendation{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	nowTimestamp := time.Now().Unix()
	in := &datahub_v1alpha1.ListPodRecommendationsRequest{
		QueryCondition: &datahub_v1alpha1.QueryCondition{
			TimeRange: &datahub_v1alpha1.TimeRange{
				EndTime: &timestamp.Timestamp{
					Seconds: nowTimestamp,
				},
			},
			Order: datahub_v1alpha1.QueryCondition_DESC,
		},
	}
	resp, err := evictioner.datahubClnt.ListPodRecommendations(ctx, in)
	if err != nil {
		return appliablePodRecList, err
	} else if resp.Status == nil {
		return appliablePodRecList, fmt.Errorf("Receive nil status from datahub")
	} else if resp.Status.Code != int32(code.Code_OK) {
		return appliablePodRecList, fmt.Errorf("Status code not 0: receive status code: %d,message: %s", resp.GetStatus().GetCode(), resp.GetStatus().GetMessage())
	}
	scope.Debugf("Possible applicable pod recommendation lists: %s", utils.InterfaceToString(resp.GetPodRecommendations()))

	for _, rec := range resp.GetPodRecommendations() {
		evictPod := false
		if rec.GetStartTime().GetSeconds() >= nowTimestamp || nowTimestamp <= rec.GetEndTime().GetSeconds() {
			continue
		}
		getResource := utilsresource.NewGetResource(evictioner.k8sClienit)
		if rec.GetNamespacedName() == nil {
			scope.Warn("receive pod recommendation with nil NamespacedName, skip this recommendation")
			continue
		}
		pod, err := getResource.GetPod(rec.GetNamespacedName().GetNamespace(), rec.GetNamespacedName().GetName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				scope.Debugf(err.Error())
			} else {
				scope.Errorf(err.Error())
			}
			continue
		}
		for _, container := range pod.Spec.Containers {
			for _, recContainer := range rec.GetContainerRecommendations() {
				if container.Name == recContainer.GetName() {
					if &container.Resources == nil || container.Resources.Limits == nil || container.Resources.Requests == nil {
						scope.Infof("Pod %s/%s selected to evict due to some resource of container %s not defined.",
							pod.GetNamespace(), pod.GetName(), recContainer.GetName())
						evictPod = true
						break
					}

					for _, resourceType := range []corev1.ResourceName{
						corev1.ResourceMemory,
						corev1.ResourceCPU,
					} {
						// resource limit check
						if _, ok := container.Resources.Limits[resourceType]; !ok {
							scope.Infof("Pod %s/%s selected to evict due to resource limit %s of container %s not defined.",
								pod.GetNamespace(), pod.GetName(), resourceType, recContainer.GetName())
							evictPod = true
							break
						}

						for _, limitRec := range recContainer.GetLimitRecommendations() {
							if resourceType == corev1.ResourceMemory && limitRec.GetMetricType() == datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES && len(limitRec.GetData()) > 0 {
								if limitRecVal, err := datahubutils.StringToInt64(limitRec.GetData()[0].GetNumValue()); err == nil {
									limitQuan := container.Resources.Limits[resourceType]
									delta := (math.Abs(float64(100*(limitRecVal-(&limitQuan).Value()))) / float64((&limitQuan).Value()))

									if delta >= memoryTriggerThreashold {
										scope.Infof("Resource limit of %s pod %s/%s container %s checking eviction threshold (%v perentage). Current setting: %v, Recommended setting: %v",
											resourceType, pod.GetNamespace(), pod.GetName(), recContainer.GetName(), memoryTriggerThreashold, limitQuan.Value(), limitRecVal)
										scope.Infof("Decide to evict pod %s/%s due to delta is %v >= %v (threashold)", pod.GetNamespace(), pod.GetName(), delta, memoryTriggerThreashold)
										evictPod = true
										break
									}
								}
							}
							if resourceType == corev1.ResourceCPU && limitRec.GetMetricType() == datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE && len(limitRec.GetData()) > 0 {
								if limitRecVal, err := datahubutils.StringToInt64(limitRec.GetData()[0].GetNumValue()); err == nil {
									limitQuan := container.Resources.Limits[resourceType]
									delta := (math.Abs(float64(100*(limitRecVal-(&limitQuan).Value()))) / float64((&limitQuan).Value()))
									if delta >= cpuTriggerThreashold {
										scope.Infof("Resource limit of %s pod %s/%s container %s checking eviction threshold (%v perentage). Current setting: %v, Recommended setting: %v",
											resourceType, pod.GetNamespace(), pod.GetName(), recContainer.GetName(), cpuTriggerThreashold, limitQuan.Value(), limitRecVal)
										scope.Infof("Decide to evict pod %s/%s due to delta is %v >= %v (threashold)", pod.GetNamespace(), pod.GetName(), delta, cpuTriggerThreashold)
										evictPod = true
										break
									}
								}
							}
						}

						if evictPod {
							break
						}

						// resource request check
						if _, ok := container.Resources.Requests[resourceType]; !ok {
							scope.Infof("Pod %s/%s selected to evict due to resource request %s of container %s not defined.",
								pod.GetNamespace(), pod.GetName(), resourceType, recContainer.GetName())
							evictPod = true
							break
						}
						for _, reqRec := range recContainer.GetRequestRecommendations() {
							if resourceType == corev1.ResourceMemory && reqRec.GetMetricType() == datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES && len(reqRec.GetData()) > 0 {
								if requestRecVal, err := datahubutils.StringToInt64(reqRec.GetData()[0].GetNumValue()); err == nil {
									requestQuan := container.Resources.Requests[resourceType]
									delta := (math.Abs(float64(100*(requestRecVal-(&requestQuan).Value()))) / float64((&requestQuan).Value()))
									scope.Debugf("Resource request of %s pod %s/%s container %s checking eviction threshold (%v perentage). Current setting: %v, Recommended setting: %v",
										resourceType, pod.GetNamespace(), pod.GetName(), recContainer.GetName(), memoryTriggerThreashold, requestQuan.Value(), requestRecVal)
									if delta >= memoryTriggerThreashold {
										scope.Infof("Decide to evict pod %s/%s due to delta is %v >= %v (threashold)", pod.GetNamespace(), pod.GetName(), delta, memoryTriggerThreashold)
										evictPod = true
										break
									}
								}
							}
							if resourceType == corev1.ResourceCPU && reqRec.GetMetricType() == datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE && len(reqRec.GetData()) > 0 {
								if requestRecVal, err := datahubutils.StringToInt64(reqRec.GetData()[0].GetNumValue()); err == nil {
									requestQuan := container.Resources.Requests[resourceType]
									delta := (math.Abs(float64(100*(requestRecVal-(&requestQuan).Value()))) / float64((&requestQuan).Value()))
									scope.Debugf("Resource request of %s pod %s/%s container %s checking eviction threshold (%v perentage). Current setting: %v, Recommended setting: %v",
										resourceType, pod.GetNamespace(), pod.GetName(), recContainer.GetName(), cpuTriggerThreashold, requestQuan.Value(), requestRecVal)
									if delta >= cpuTriggerThreashold {
										scope.Infof("Decide to evict pod %s/%s due to delta is %v >= %v (threashold)", pod.GetNamespace(), pod.GetName(), delta, cpuTriggerThreashold)
										evictPod = true
										break
									}
								}
							}
						}

						if evictPod {
							break
						}
					}

					if evictPod {
						break
					}
				}
				if evictPod {
					break
				}
			}
		}
		if evictPod {
			appliablePodRecList = append(appliablePodRecList, rec)
		}
	}
	return appliablePodRecList, nil
}
