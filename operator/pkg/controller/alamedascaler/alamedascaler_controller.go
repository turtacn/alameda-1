/*
Copyright 2019 The Alameda Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alamedascaler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	scope = logUtil.RegisterScope("alamedascaler", "alamedascaler log", 0)
)

// AlamedaScaler is alameda scaler
type AlamedaScaler string

//
const (
	AlamedaDeployment AlamedaScaler = "Deployment"
	UpdateRetry                     = 3
)

// AlamedaK8sController is key of AlamedaScaler annotation
const AlamedaK8sController = "annotation-k8s-controller"

// JSONIndent is ident of formatted json string
const JSONIndent = "  "

// Container struct
type Container struct {
	Name string
}

// Pod struct
type Pod struct {
	UID        string
	Namespace  string
	Name       string
	Containers []Container
}

// Deployment struct
type Deployment struct {
	UID       string
	Namespace string
	Name      string
	PodMap    map[string]Pod
}

// K8SControllerAnnotation struct
type K8SControllerAnnotation struct {
	DeploymentMap map[string]Deployment
}

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AlamedaScaler Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this autoscaling.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAlamedaScaler{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("alamedascaler-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &autoscalingv1alpha1.AlamedaScaler{}}, &handler.EnqueueRequestForObject{}); err != nil {
		scope.Error(err.Error())
	}

	if err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}); err != nil {
		scope.Error(err.Error())
	}

	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAlamedaScaler{}

// ReconcileAlamedaScaler reconciles a AlamedaScaler object
type ReconcileAlamedaScaler struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AlamedaScaler object and makes changes based on the state read
// and what is in the AlamedaScaler .Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.containers.ai,resources=alamedascalers,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileAlamedaScaler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the AlamedaScaler instance
	deleteEvt := true
	ns := request.Namespace
	name := request.Name
	alamedascaler := &autoscalingv1alpha1.AlamedaScaler{}
	scope.Infof(fmt.Sprintf("Try to get AlamedaScaler (%s/%s)", ns, name))
	err := r.Get(context.TODO(), request.NamespacedName, alamedascaler)
	if err != nil && errors.IsNotFound(err) {
		//delete Alameda Resource Predict CR due to no Alameda CR with the same name exist
		alamedaResourcePrediction := &autoscalingv1alpha1.AlamedaResourcePrediction{}
		err = r.Get(context.TODO(), request.NamespacedName, alamedaResourcePrediction)
		if err == nil {
			r.Delete(context.TODO(), alamedaResourcePrediction)
		}
	} else if err == nil {
		scope.Infof(fmt.Sprintf("AlamedaScaler found. (%s/%s)", ns, name))
		r.updateAlamedaScalerAnnotation(alamedascaler.Spec.Selector.MatchLabels, alamedascaler.GetAnnotations(), ns, name)
		//check in alameda predict CR exist state
		alamedaResourcePrediction := &autoscalingv1alpha1.AlamedaResourcePrediction{}

		err := r.Get(context.TODO(), request.NamespacedName, alamedaResourcePrediction)
		if err != nil && errors.IsNotFound(err) {
			scope.Infof(fmt.Sprintf("Create AlamedaResourcePrediction for AlamedaScaler. (%s/%s)", alamedascaler.Namespace, alamedascaler.Name))
			err = CreateAlamedaPrediction(r, r.scheme, alamedascaler)
			if err != nil {
				scope.Error(err.Error())
			}
		}
		deleteEvt = false
	} else {
		scope.Error(err.Error())
	}

	deploymentFound := &appsv1.Deployment{}
	err = r.Get(context.TODO(), request.NamespacedName, deploymentFound)
	if err != nil {
		if errors.IsNotFound(err) {
			scope.Warnf(fmt.Sprintf("Deployment not found (%s/%s).", ns, name))

		} else {
			scope.Error(err.Error())
		}
	} else {
		r.updateAlamedaK8SControllerByDeployment(ns, deploymentFound)
		deleteEvt = false
	}

	if deleteEvt {
		scope.Info(fmt.Sprintf("Delete event."))
		alamedaResourceList := &autoscalingv1alpha1.AlamedaScalerList{}
		err = r.List(context.TODO(),
			client.InNamespace(ns),
			alamedaResourceList)
		for _, ala := range alamedaResourceList.Items {
			r.updateAlamedaAnnotationByDeleteEvt(&ala, request)
		}
	}
	return reconcile.Result{}, nil
}

// CreateAlamedaPrediction Creates AlamedaScalerAlamedaResourcePrediction CR
func CreateAlamedaPrediction(r client.Client, scheme *runtime.Scheme, alamedascaler *autoscalingv1alpha1.AlamedaScaler) error {
	newAlamedaResourcePrediction := &autoscalingv1alpha1.AlamedaResourcePrediction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      alamedascaler.Name,
			Namespace: alamedascaler.Namespace,
		},
		Spec: autoscalingv1alpha1.AlamedaResourcePredictionSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: alamedascaler.Spec.Selector.MatchLabels,
			},
		},
		Status: autoscalingv1alpha1.AlamedaResourcePredictionStatus{
			Prediction: autoscalingv1alpha1.AlamedaPrediction{
				Deployments: map[autoscalingv1alpha1.DeploymentUID]autoscalingv1alpha1.PredictDeployment{},
			},
		},
	}
	/*
		if err := controllerutil.SetControllerReference(alamedascaler, newAlamedaResourcePrediction, scheme); err != nil {
			return err
		}
	*/

	err := r.Create(context.TODO(), newAlamedaResourcePrediction)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileAlamedaScaler) updateAlamedaScalerAnnotation(matchLabels, alamedaAnnotations map[string]string, namespace, name string) {
	noAlaPod := false
	newPodMaps := map[string]Pod{}
	newAlamedaAnnotations := map[string]string{}
	if alamedaAnnotations == nil {
		newAlamedaAnnotations = map[string]string{}
		noAlaPod = true
	} else {
		for k, v := range alamedaAnnotations {
			newAlamedaAnnotations[k] = v
		}
	}
	if newAlamedaAnnotations[AlamedaK8sController] == "" {
		newAlamedaAnnotations[AlamedaK8sController] = alamedaK8sControllerDefautlAnno()
		noAlaPod = true
	}
	//find matched deployment controller
	matchedDeploymentList := &appsv1.DeploymentList{}
	err := r.List(context.TODO(),
		client.InNamespace(namespace).
			MatchingLabels(matchLabels),
		matchedDeploymentList)
	if err == nil {
		akcMap := convertk8scontrollerJSONString(newAlamedaAnnotations[AlamedaK8sController])

		for _, deploy := range matchedDeploymentList.Items {
			akcMap.DeploymentMap[string(deploy.GetUID())] = *r.getControllerMapForAnno("deployment", &deploy).(*Deployment)
			// New AlamedaScaler created, records Alameda pods to register AI server
			if noAlaPod {
				for k, v := range akcMap.DeploymentMap[string(deploy.GetUID())].PodMap {
					newPodMaps[k] = v
				}
			}
		}
		updatemd, _ := json.MarshalIndent(akcMap, "", JSONIndent)
		newAlamedaAnnotations[AlamedaK8sController] = string(updatemd)
	}
	if len(newAlamedaAnnotations) > 0 && !reflect.DeepEqual(newAlamedaAnnotations, alamedaAnnotations) {
		for retry := 0; retry < UpdateRetry; retry++ {
			time.Sleep(1 * time.Second)
			alamedascaler := &autoscalingv1alpha1.AlamedaScaler{}
			err := r.Get(context.TODO(), types.NamespacedName{
				Namespace: namespace,
				Name:      name,
			}, alamedascaler)
			alamedascalerAnno := getAnnotations(alamedascaler)
			alamedascalerAnno[AlamedaK8sController] = newAlamedaAnnotations[AlamedaK8sController]
			alamedascaler.SetAnnotations(alamedascalerAnno)
			if err != nil {
				scope.Error(err.Error())
			} else {
				err = r.Update(context.TODO(), alamedascaler)
				if err != nil {
					scope.Error(err.Error())
				} else {
					registerPodPrediction(alamedascaler, namespace, name, newPodMaps, nil)
					break
				}
			}
		}
	}
}

func (r *ReconcileAlamedaScaler) updateAlamedaK8SControllerByDeployment(ns string, deploymentFound *appsv1.Deployment) {
	alamedaResourceList := &autoscalingv1alpha1.AlamedaScalerList{}
	err := r.List(context.TODO(),
		client.InNamespace(ns),
		alamedaResourceList)
	if err != nil {
		scope.Error(err.Error())
		return
	}

	for _, ala := range alamedaResourceList.Items {
		r.updateAlamedaAnnotationByDeployment(&ala, deploymentFound)
	}
}

func (r *ReconcileAlamedaScaler) updateAlamedaAnnotationByDeleteEvt(ala *autoscalingv1alpha1.AlamedaScaler, request reconcile.Request) {
	needUpdated := false
	name := request.Name
	anno := getAnnotations(ala)
	if anno != nil && anno[AlamedaK8sController] != "" {
		k8sc := convertk8scontrollerJSONString(anno[AlamedaK8sController])
		//handle deployment controller
		for k, v := range k8sc.DeploymentMap {
			if v.Name == name {
				delete(k8sc.DeploymentMap, k)
				needUpdated = true
			}
		}
		if needUpdated {
			updated, _ := json.MarshalIndent(k8sc, "", JSONIndent)
			anno[AlamedaK8sController] = string(updated)
			ala.SetAnnotations(anno)
			_ = r.Update(context.TODO(), ala)
		}
	}
}

func (r *ReconcileAlamedaScaler) updateAlamedaAnnotationByDeployment(ala *autoscalingv1alpha1.AlamedaScaler, deploy *appsv1.Deployment) {
	needUpdated := false
	alaML := ala.Spec.Selector.MatchLabels
	dL := deploy.GetLabels()
	dpUID := deploy.GetUID()
	anno := getAnnotations(ala)

	k8sc := convertk8scontrollerJSONString(anno[AlamedaK8sController])
	deletePodMaps := map[string]Pod{}
	newPodMaps := map[string]Pod{}
	if isLabelsMatched(dL, alaML) {
		curDeployment := *r.getControllerMapForAnno("deployment", deploy).(*Deployment)
		if _, found := k8sc.DeploymentMap[string(dpUID)]; found {
			legacyDeployment := k8sc.DeploymentMap[string(dpUID)]

			for k, v := range legacyDeployment.PodMap {
				if _, found := curDeployment.PodMap[k]; !found {
					deletePodMaps[k] = v
				}
			}
			for k, v := range curDeployment.PodMap {
				if _, found := legacyDeployment.PodMap[k]; !found {
					newPodMaps[k] = v
				}
			}
		} else {
			for k, v := range curDeployment.PodMap {
				newPodMaps[k] = v
			}
		}
		k8sc.DeploymentMap[string(dpUID)] = curDeployment
		deletePodMapsBin, _ := json.MarshalIndent(deletePodMaps, "", JSONIndent)
		newPodMapsBin, _ := json.MarshalIndent(newPodMaps, "", JSONIndent)

		scope.Infof(fmt.Sprintf("Alameda Deployment Pods to add %s. (%s/%s).", string(newPodMapsBin), deploy.GetNamespace(), deploy.GetName()))
		scope.Infof(fmt.Sprintf("Alameda Deployment Pods to delete %s. (%s/%s).", string(deletePodMapsBin), deploy.GetNamespace(), deploy.GetName()))
		needUpdated = true
	} else {
		if _, found := k8sc.DeploymentMap[string(dpUID)]; found {
			for k, v := range k8sc.DeploymentMap[string(deploy.GetUID())].PodMap {
				deletePodMaps[k] = v
			}
			delete(k8sc.DeploymentMap, string(deploy.GetUID()))
			needUpdated = true
		}
	}
	if needUpdated {
		for retry := 0; retry < UpdateRetry; retry++ {
			time.Sleep(1 * time.Second)
			updated, _ := json.MarshalIndent(k8sc, "", JSONIndent)
			alaIns := &autoscalingv1alpha1.AlamedaScaler{}
			err := r.Get(context.TODO(), types.NamespacedName{
				Namespace: ala.GetNamespace(),
				Name:      ala.GetName(),
			}, alaIns)
			if err != nil {
				scope.Error(err.Error())
				continue
			}

			alaInsAnno := getAnnotations(alaIns)
			alaInsAnno[AlamedaK8sController] = string(updated)
			alaIns.SetAnnotations(alaInsAnno)
			err = r.Update(context.TODO(), ala)
			if err != nil {
				scope.Error(err.Error())
				continue
			}

			registerPodPrediction(alaIns, deploy.GetNamespace(), deploy.GetName(), newPodMaps, deletePodMaps)
			break
		}
	}
}

func registerPodPrediction(alamedascaler *autoscalingv1alpha1.AlamedaScaler, namespace, name string, newPodMaps, deletePodMaps map[string]Pod) {
	scope.Info("Start registering pod prediction.")
	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())

	if err != nil {
		scope.Error(err.Error())
		return
	}

	defer conn.Close()
	aiServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	if len(newPodMaps) > 0 {
		req := datahub_v1alpha1.CreatePodsRequest{
			Pods: []*datahub_v1alpha1.Pod{},
		}
		for _, v := range newPodMaps {
			policy := datahub_v1alpha1.RecommendationPolicy_STABLE
			containers := []*datahub_v1alpha1.Container{}
			if strings.ToLower(string(alamedascaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicyCOMPACT)) {
				policy = datahub_v1alpha1.RecommendationPolicy_COMPACT
			} else if strings.ToLower(string(alamedascaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicySTABLE)) {
				policy = datahub_v1alpha1.RecommendationPolicy_STABLE
			}
			for _, c := range v.Containers {
				containers = append(containers, &datahub_v1alpha1.Container{
					Name: c.Name,
				})
			}
			req.Pods = append(req.Pods, &datahub_v1alpha1.Pod{
				IsAlameda: true,
				AlamedaScaler: &datahub_v1alpha1.NamespacedName{
					Namespace: alamedascaler.GetNamespace(),
					Name:      alamedascaler.GetName(),
				},
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: v.Namespace,
					Name:      v.Name,
				},
				Policy:     datahub_v1alpha1.RecommendationPolicy(policy),
				Containers: containers,

				// TODO
				NodeName:     "",
				ResourceLink: "",
				StartTime:    &timestamp.Timestamp{},
			})
		}
		reqBin, _ := json.MarshalIndent(req, "", JSONIndent)
		scope.Infof(fmt.Sprintf("Add/Update alameda pod %s to datahub. (%s/%s).", string(reqBin), namespace, name))
		reqRes, err := aiServiceClnt.CreatePods(context.Background(), &req)
		if err != nil {
			scope.Error(err.Error())
		} else {
			reqResBin, _ := json.Marshal(*reqRes)
			scope.Infof(fmt.Sprintf("Add/Update alameda pod %s to datahub successfully (%s).", string(reqBin), reqResBin))
		}
	}
	if len(deletePodMaps) > 0 {
		req := datahub_v1alpha1.DeletePodsRequest{
			Pods: []*datahub_v1alpha1.Pod{},
		}
		for _, v := range deletePodMaps {
			req.Pods = append(req.Pods, &datahub_v1alpha1.Pod{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: v.Namespace,
					Name:      v.Name,
				},
			})
		}
		reqBin, _ := json.MarshalIndent(req, "", JSONIndent)
		scope.Infof(fmt.Sprintf("Remove alameda pod %s from datahub. (%s/%s).", string(reqBin), namespace, name))
		reqRes, err := aiServiceClnt.DeletePods(context.Background(), &req)
		if err != nil {
			scope.Error(err.Error())
		} else {
			reqResBin, _ := json.Marshal(*reqRes)
			scope.Infof(fmt.Sprintf("Remove alameda pod %s from datahub successfully (%s).", string(reqBin), reqResBin))
		}
	}
}

func isLabelsMatched(labels, matchlabels map[string]string) bool {
	if len(matchlabels) > len(labels) {
		return false
	}
	for k, v := range matchlabels {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func alamedaK8sControllerDefautlAnno() string {
	md, _ := json.MarshalIndent(*GetDefaultAlamedaK8SControllerAnno(), "", JSONIndent)
	return string(md)
}

// GetDefaultAlamedaK8SControllerAnno get default AlamedaScaler annotation of K8S controller
func GetDefaultAlamedaK8SControllerAnno() *K8SControllerAnnotation {
	return &K8SControllerAnnotation{
		DeploymentMap: map[string]Deployment{},
	}
}

//annotation-k8s-controller annotation struct definition
func convertk8scontrollerJSONString(jsonStr string) *K8SControllerAnnotation {
	akcMap := GetDefaultAlamedaK8SControllerAnno()
	err := json.Unmarshal([]byte(jsonStr), akcMap)
	if err != nil {
		scope.Error(err.Error())
	}
	return akcMap
}

func (r *ReconcileAlamedaScaler) getControllerMapForAnno(kind string, deploy interface{}) interface{} {
	if kind == "deployment" {
		namespace := deploy.(*appsv1.Deployment).GetNamespace()
		name := deploy.(*appsv1.Deployment).GetName()
		listPods := utilsresource.NewListPods(r)
		podList := listPods.ListPods(namespace, name, "deployment")
		podMap := map[string]Pod{}
		for _, pod := range podList {
			containers := []Container{}
			for _, container := range pod.Spec.Containers {
				containers = append(containers, Container{Name: container.Name})
			}
			podMap[string(pod.GetUID())] = Pod{
				Name:       pod.GetName(),
				Namespace:  namespace,
				UID:        string(pod.GetUID()),
				Containers: containers,
			}
		}
		return &Deployment{
			Name:      name,
			Namespace: namespace,
			UID:       string(deploy.(*appsv1.Deployment).GetUID()),
			PodMap:    podMap,
		}
	}
	return nil
}

func getAnnotations(ala *autoscalingv1alpha1.AlamedaScaler) map[string]string {
	anno := ala.GetAnnotations()
	if anno == nil {
		anno = map[string]string{}
	}
	if _, ok := anno[AlamedaK8sController]; !ok {
		anno[AlamedaK8sController] = alamedaK8sControllerDefautlAnno()
	}
	return anno
}
