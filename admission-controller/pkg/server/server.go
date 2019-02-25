package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/containers-ai/alameda/admission-controller/pkg/recommendator/resource"
	admission_controller_utils "github.com/containers-ai/alameda/admission-controller/pkg/utils"
	metadata_utils "github.com/containers-ai/alameda/pkg/utils/kubernetes/metadata"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	admission_v1beta1 "k8s.io/api/admission/v1beta1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	patchType = admission_v1beta1.PatchTypeJSONPatch
	scope     = log.RegisterScope("admission-controller", "admission-controller", 0)
)

type admitFunc func(*admission_v1beta1.AdmissionReview) *admission_v1beta1.AdmissionResponse

type admissionController struct {
	config *Config

	lock                                   *sync.Mutex
	controllerPodResourceRecommendationMap map[namespaceKindName]*controllerPodResourceRecommendation
	resourceRecommendator                  resource.ResourceRecommendator
	resourceRecommendatorSyncTimeout       time.Duration
	resourceRecommendatorSyncWaitInterval  time.Duration
	ownerReferenceTracer                   *metadata_utils.OwnerReferenceTracer
}

func NewAdmissionControllerWithConfig(cfg Config, resourceRecommendator resource.ResourceRecommendator) (AdmissionController, error) {

	defaultOwnerReferenceTracer, err := metadata_utils.NewDefaultOwnerReferenceTracer()
	if err != nil {
		return nil, errors.Wrap(err, "new AdmissionController failed")
	}

	ac := &admissionController{
		config: &cfg,

		lock:                                   &sync.Mutex{},
		controllerPodResourceRecommendationMap: make(map[namespaceKindName]*controllerPodResourceRecommendation),
		resourceRecommendator:                  resourceRecommendator,
		resourceRecommendatorSyncTimeout:       10 * time.Second,
		resourceRecommendatorSyncWaitInterval:  5 * time.Second,
		ownerReferenceTracer:                   defaultOwnerReferenceTracer,
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

	recommendation = controllerRecommendation.dispatchOneValidRecommendation(time.Now())
	if recommendation == nil {

		state := controllerRecommendation.waitOrSync()
		if state != recommendationWaitsSynchronizing {
			if err := ac.doRecommendationSync(controllerID); err != nil {
				return recommendation, errors.Wrap(err, "fetch controller pod resource recommendation failed")
			}
		} else {
			ac.waitRecommendationSync(controllerID)
		}
		recommendation = controllerRecommendation.dispatchOneValidRecommendation(time.Now())
	}

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

func (ac *admissionController) doRecommendationSync(controllerID namespaceKindName) error {

	controllerRecommendation := ac.getControllerPodResourceRecommendation(controllerID)

	// unblock others waiting goroutine
	defer controllerRecommendation.finishSync()

	if newRecommendations, err := ac.fetchNewRecommendations(controllerID); err != nil {
		return errors.Wrapf(err, "do recommendation sync failed: controllerID: %s", controllerID)
	} else {
		controllerRecommendation.appendRecommendations(newRecommendations)
	}
	return nil
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

func (ac *admissionController) waitRecommendationSync(controllerID namespaceKindName) {
	scope.Debug("waiting recommendations synchronizning")
	controllerRecommendation := ac.getControllerPodResourceRecommendation(controllerID)
	<-controllerRecommendation.syncChan
	scope.Debug("finish waiting recommendations synchronizning")
}
