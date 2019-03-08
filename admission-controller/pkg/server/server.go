package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/containers-ai/alameda/admission-controller/pkg/recommendator/resource"
	admission_controller_utils "github.com/containers-ai/alameda/admission-controller/pkg/utils"
	alamedascaler_reconciler "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedascaler"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources"
	metadata_utils "github.com/containers-ai/alameda/pkg/utils/kubernetes/metadata"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	admission_v1beta1 "k8s.io/api/admission/v1beta1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	patchType = admission_v1beta1.PatchTypeJSONPatch
	scope     = log.RegisterScope("admission-controller", "admission-controller", 0)
)

type admitFunc func(*admission_v1beta1.AdmissionReview) *admission_v1beta1.AdmissionResponse

type admissionController struct {
	config *Config

	lock                                       *sync.Mutex
	controllerPodResourceRecommendationMap     map[namespaceKindName]*controllerPodResourceRecommendation
	controllerPodResourceRecommendationLockMap map[namespaceKindName]*sync.Mutex
	resourceRecommendator                      resource.ResourceRecommendator
	resourceRecommendatorSyncTimeout           time.Duration
	resourceRecommendatorSyncRetryTime         int
	resourceRecommendatorSyncWaitInterval      time.Duration

	k8sConfig            *rest.Config
	ownerReferenceTracer *metadata_utils.OwnerReferenceTracer
}

func NewAdmissionControllerWithConfig(cfg Config, resourceRecommendator resource.ResourceRecommendator) (AdmissionController, error) {

	defaultOwnerReferenceTracer, err := metadata_utils.NewDefaultOwnerReferenceTracer()
	if err != nil {
		return nil, errors.Wrap(err, "new AdmissionController failed")
	}

	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "new AdmissionController failed")
	}

	ac := &admissionController{
		config: &cfg,

		lock:                                   &sync.Mutex{},
		controllerPodResourceRecommendationMap: make(map[namespaceKindName]*controllerPodResourceRecommendation),
		controllerPodResourceRecommendationLockMap: make(map[namespaceKindName]*sync.Mutex),
		resourceRecommendator:                      resourceRecommendator,
		resourceRecommendatorSyncTimeout:           10 * time.Second,
		resourceRecommendatorSyncRetryTime:         3,
		resourceRecommendatorSyncWaitInterval:      5 * time.Second,

		k8sConfig:            k8sConfig,
		ownerReferenceTracer: defaultOwnerReferenceTracer,
	}

	return ac, nil
}

func (ac *admissionController) MutatePod(w http.ResponseWriter, r *http.Request) {
	ac.serve(w, r, ac.mutatePod)
}

func (ac *admissionController) serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {

	var admissionResponse *admission_v1beta1.AdmissionResponse

	if ac.config.Enable {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			scope.Errorf("contentType=%s, expect application/json", contentType)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			scope.Errorf("read http request failed: %s", err.Error())
		}

		admissionReview := &admission_v1beta1.AdmissionReview{}
		if err := json.Unmarshal(body, admissionReview); err != nil {
			scope.Errorf("unmarshal AdmissionReview failed: %s", err.Error())
		} else {
			admissionResponse = admit(admissionReview)
			if admissionResponse == nil {
				scope.Warnf("received nill AdmissionResponse, skip mutating pod, AdmissionReview: %+v", admissionReview)
			} else {
				admissionResponse.UID = admissionReview.Request.UID
			}
		}
	} else {
		scope.Warn("admission-controller is not enabled")
	}

	newAdmissionReview := admission_v1beta1.AdmissionReview{
		Response: admissionResponse,
	}
	admissionReviewBytes, err := json.Marshal(newAdmissionReview)
	if err != nil {
		scope.Errorf("marshal AdmissionReview  failed: %s", err.Error())
	}

	_, err = w.Write(admissionReviewBytes)
	if err != nil {
		scope.Errorf("write AdmissionReview failed: %s", err.Error())
	}

}

func (ac *admissionController) mutatePod(ar *admission_v1beta1.AdmissionReview) *admission_v1beta1.AdmissionResponse {

	scope.Debug("mutate pod")

	namespace := ar.Request.Namespace

	podResource := meta_v1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		err := errors.Errorf("expect resource to be %s, get %s", podResource.String(), ar.Request.Resource.String())
		scope.Error(err.Error())
		return nil
	}

	raw := ar.Request.Object.Raw
	pod := core_v1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		scope.Errorf("deserialize admission request raw to Pod struct failed: %s", err.Error())
		return nil
	}

	reviewResponse := admission_v1beta1.AdmissionResponse{
		Allowed: true,
	}

	controllerKind, controllerName, err := ac.ownerReferenceTracer.GetRootControllerKindAndNameOfOwnerReferences(namespace, pod.OwnerReferences)
	if err != nil {
		scope.Warnf("get root controller information of Pod failed, skip mutate pod, errMsg: %s", err.Error())
		return &reviewResponse
	}

	controllerID := newNamespaceKindName(namespace, controllerKind, controllerName)
	recommendation, err := ac.getPodResourceRecommendationByControllerID(controllerID)
	if err != nil {
		scope.Error(err.Error())
		return nil
	} else if recommendation == nil {
		scope.Error("fetch empty recommendation, skip mutate pod")
		return &reviewResponse
	}

	patches, err := admission_controller_utils.GetPatchesFromPodResourceRecommendation(&pod, recommendation)
	if err != nil {
		scope.Errorf("get patches to mutate pod resource failed: %s", err.Error())
		return nil
	}

	reviewResponse.Patch = []byte(patches)
	reviewResponse.PatchType = &patchType

	return &reviewResponse
}

func (ac *admissionController) getPodResourceRecommendationByControllerID(controllerID namespaceKindName) (*resource.PodResourceRecommendation, error) {

	var recommendation *resource.PodResourceRecommendation

	controllerRecommendation := ac.getControllerPodResourceRecommendation(controllerID)
	controllerRecommendationLock := ac.getControllerPodResourceRecommendationLock(controllerID)

	retryTime := ac.resourceRecommendatorSyncRetryTime
	controllerRecommendationLock.Lock()
	defer controllerRecommendationLock.Unlock()
	for recommendation == nil && retryTime > 0 {
		if newRecommendations, err := ac.fetchNewRecommendations(controllerID); err != nil {
			scope.Error(err.Error())
		} else {
			controllerRecommendation.setRecommendations(newRecommendations)
			break
		}
		retryTime--
	}
	validRecommedations, err := ac.listValidRecommendations(controllerID, controllerRecommendation.getRecommendations())
	if err != nil {
		scope.Error(err.Error())
	}
	controllerRecommendation.setRecommendations(validRecommedations)
	recommendation = controllerRecommendation.dispatchOneValidRecommendation(time.Now())

	return recommendation, nil
}

func (ac *admissionController) getControllerPodResourceRecommendation(controllerID namespaceKindName) *controllerPodResourceRecommendation {

	ac.lock.Lock()
	controllerRecommendation, exist := ac.controllerPodResourceRecommendationMap[controllerID]
	if !exist {
		scope.Debugf("controllerID: %s, controller recommendation not exist, create new recommendation.", controllerID)
		ac.controllerPodResourceRecommendationMap[controllerID] = NewControllerPodResourceRecommendation()
		controllerRecommendation = ac.controllerPodResourceRecommendationMap[controllerID]
	}
	ac.lock.Unlock()

	return controllerRecommendation
}

func (ac *admissionController) getControllerPodResourceRecommendationLock(controllerID namespaceKindName) *sync.Mutex {

	ac.lock.Lock()
	controllerRecommendationLock, exist := ac.controllerPodResourceRecommendationLockMap[controllerID]
	if !exist {
		scope.Debugf("controllerID: %s, controller recommendation not exist, create new recommendation.", controllerID)
		ac.controllerPodResourceRecommendationLockMap[controllerID] = &sync.Mutex{}
		controllerRecommendationLock = ac.controllerPodResourceRecommendationLockMap[controllerID]
	}
	ac.lock.Unlock()

	return controllerRecommendationLock
}

func (ac *admissionController) listValidRecommendations(controllerID namespaceKindName, recommendations []*resource.PodResourceRecommendation) ([]*resource.PodResourceRecommendation, error) {

	validRecommendations := make([]*resource.PodResourceRecommendation, 0)

	initRecommendationNumberMap := buildRecommendationNumberMap(recommendations)
	scope.Debugf("initRecommendationNumberMap %+v", initRecommendationNumberMap)

	pods, err := ac.listPodControlledByControllerID(controllerID)
	if err != nil {
		return validRecommendations, errors.Wrap(err, "list valid recommendations failed")
	}
	currentRunningPods := make([]*core_v1.Pod, 0)
	for _, pod := range pods {
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		currentRunningPods = append(currentRunningPods, pod)
	}
	decreaseRecommendationNuberMapByPods(initRecommendationNumberMap, pods)

	validRecommendations = getValidRecommedationFromRecommendationNumberMap(initRecommendationNumberMap, recommendations)
	scope.Debugf("validRecommendations %+v", validRecommendations)

	return validRecommendations, nil
}

func (ac *admissionController) listPodControlledByControllerID(controllerID namespaceKindName) ([]*core_v1.Pod, error) {
	pods := make([]*core_v1.Pod, 0)

	sigsK8SClient, err := ac.newSigsK8SIOClient()
	if err != nil {
		return pods, errors.Wrapf(err, "list pods controlled by controllerID: %s failed", controllerID.String())
	}

	podsInCluster := make([]core_v1.Pod, 0)
	listResource := resources.NewListResources(sigsK8SClient)
	switch controllerID.getKind() {
	case "Deployment":
		podsInCluster, err = listResource.ListPodsByDeployment(controllerID.getNamespace(), controllerID.getName())
		if err != nil {
			return pods, errors.Wrapf(err, "list pods controlled by controllerID: %s failed", controllerID.String())
		}
	case "DeploymentConfig":
		podsInCluster, err = listResource.ListPodsByDeploymentConfig(controllerID.getNamespace(), controllerID.getName())
		if err != nil {
			return pods, errors.Wrapf(err, "list pods controlled by controllerID: %s failed", controllerID.String())
		}
	default:
		return pods, errors.Errorf("no matching resource lister for controller kind: %s", controllerID.getKind())
	}

	for _, pod := range podsInCluster {
		copyPod := pod
		pods = append(pods, &copyPod)
	}

	return pods, nil
}

func (ac *admissionController) newSigsK8SIOClient() (client.Client, error) {

	return client.New(ac.k8sConfig, client.Options{Scheme: scheme})
}

func (ac *admissionController) fetchNewRecommendations(controllerID namespaceKindName) ([]*resource.PodResourceRecommendation, error) {

	scope.Debug("fetching new recommendations from recommendator")

	var err error
	recommendations := make([]*resource.PodResourceRecommendation, 0)
	done := make(chan bool)

	go func(chan bool) {
		queryTime := time.Now()
		recommendations, err = ac.resourceRecommendator.ListControllerPodResourceRecommendations(resource.ListControllerPodResourceRecommendationsRequest{
			Namespace: controllerID.getNamespace(),
			Name:      controllerID.getName(),
			Kind:      controllerID.getKind(),
			Time:      &queryTime,
		})
		done <- true
	}(done)

	select {
	case _ = <-done:
	case _ = <-time.After(ac.resourceRecommendatorSyncTimeout):
		err = errors.Errorf("fetch recommendations failed: timeout after %f seconds", ac.resourceRecommendatorSyncTimeout.Seconds())
	}

	return recommendations, err
}

func buildRecommendationNumberMap(recommendations []*resource.PodResourceRecommendation) map[string]int {
	currentTime := time.Now()
	recommendationNumberMap := make(map[string]int)
	for _, recommendation := range recommendations {
		if !(recommendation.ValidStartTime.Unix() < currentTime.Unix() && currentTime.Unix() < recommendation.ValidEndTime.Unix()) {
			continue
		}
		recommendationID := buildPodResourceIDFromPodRecommendation(recommendation)
		recommendationNumberMap[recommendationID]++
	}
	return recommendationNumberMap
}

func decreaseRecommendationNuberMapByPods(recommendationNumberMap map[string]int, pods []*core_v1.Pod) {
	for _, pod := range pods {
		scope.Debugf("try to decrease recommendation from pod: %+v", pod)
		if pod.ObjectMeta.DeletionTimestamp != nil {
			scope.Debugf("skip decreate recommendation cause pod %s/%s has deletion timestamp", pod.Namespace, pod.Name)
			continue
		}
		if !alamedascaler_reconciler.PodIsMonitoredByAlameda(pod) {
			scope.Debugf("skip decreate recommendation cause pod's %s/%s phase: %s is not monitored by Alameda", pod.Namespace, pod.Name, pod.Status.Phase)
			continue
		}
		recommendationID := buildPodResourceIDFromPod(pod)
		if _, exist := recommendationNumberMap[recommendationID]; exist {
			scope.Debugf("decreate recommendation for pod %s/%s", pod.Namespace, pod.Name)
			recommendationNumberMap[recommendationID]--
		} else {
			scope.Debugf("no matched key found in recommendationMap: key: %s", recommendationID)
		}
	}
}

func getValidRecommedationFromRecommendationNumberMap(recommendationNumberMap map[string]int, recommendations []*resource.PodResourceRecommendation) []*resource.PodResourceRecommendation {

	validRecommendations := make([]*resource.PodResourceRecommendation, 0)
	for _, recommendation := range recommendations {
		copyRecommendation := recommendation
		recommendationID := buildPodResourceIDFromPodRecommendation(recommendation)
		if remainRecommendationsNum := recommendationNumberMap[recommendationID]; remainRecommendationsNum > 0 {
			recommendationNumberMap[recommendationID]--
			validRecommendations = append(validRecommendations, copyRecommendation)
		}
	}
	return validRecommendations
}

func buildPodResourceIDFromPod(pod *core_v1.Pod) string {

	containers := pod.Spec.Containers

	sort.SliceStable(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	id := ""
	for _, container := range containers {
		requestCPU := container.Resources.Requests[core_v1.ResourceCPU]
		requestMem := container.Resources.Requests[core_v1.ResourceMemory]
		limitsCPU := container.Resources.Limits[core_v1.ResourceCPU]
		limitsMem := container.Resources.Limits[core_v1.ResourceMemory]
		id += fmt.Sprintf("container-name-%s/requset-cpu-%s-mem-%s/limit-cpu-%s-mem-%s/", container.Name,
			requestCPU.String(), requestMem.String(),
			limitsCPU.String(), limitsMem.String(),
		)
	}

	return id
}

func buildPodResourceIDFromPodRecommendation(recommendation *resource.PodResourceRecommendation) string {

	containerRecommendations := recommendation.ContainerResourceRecommendations
	sort.SliceStable(containerRecommendations, func(i, j int) bool {
		return containerRecommendations[i].Name < containerRecommendations[j].Name
	})

	id := ""
	for _, containerRecommendation := range containerRecommendations {
		requestCPU := containerRecommendation.Requests[core_v1.ResourceCPU]
		requestMem := containerRecommendation.Requests[core_v1.ResourceMemory]
		limitsCPU := containerRecommendation.Limits[core_v1.ResourceCPU]
		limitsMem := containerRecommendation.Limits[core_v1.ResourceMemory]
		id += fmt.Sprintf("container-name-%s/requset-cpu-%s-mem-%s/limit-cpu-%s-mem-%s/", containerRecommendation.Name,
			requestCPU.String(), requestMem.String(),
			limitsCPU.String(), limitsMem.String(),
		)
	}
	return id
}
