/*
Copyright 2018 The Alameda Authors.

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

package alamedaresource

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	grpcutils "github.com/containers-ai/alameda/operator/pkg/utils/grpc"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	aiservice_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/ai_service"
	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"

	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AlamedaResource is alameda resource
type AlamedaResource string

//
const (
	AlamedaDeployment AlamedaResource = "Deployment"
)

// AlamedaK8sController is key of AlamedaResource annotation
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
	Name       string
	Containers []Container
}

// Deployment struct
type Deployment struct {
	UID    string
	Name   string
	PodMap map[string]Pod
}

// K8SControllerAnnotation struct
type K8SControllerAnnotation struct {
	DeploymentMap map[string]Deployment
}

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AlamedaResource Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this autoscaling.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAlamedaResource{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("alamedaresource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &autoscalingv1alpha1.AlamedaResource{}}, &handler.EnqueueRequestForObject{}); err != nil {
		logUtil.GetLogger().Error(err, fmt.Sprintf("Watch AlamedaResource failed"))
	}

	if err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}); err != nil {
		logUtil.GetLogger().Error(err, fmt.Sprintf("Watch Deployment controller for AlemedaResource failed."))
	}

	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAlamedaResource{}

// ReconcileAlamedaResource reconciles a AlamedaResource object
type ReconcileAlamedaResource struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AlamedaResource object and makes changes based on the state read
// and what is in the AlamedaResource.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.containers.ai,resources=alamedaresources,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileAlamedaResource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the AlamedaResource instance
	deleteEvt := true
	ns := request.Namespace
	name := request.Name

	alamedaresource := &autoscalingv1alpha1.AlamedaResource{}
	logUtil.GetLogger().Info(fmt.Sprintf("Try to get AlamedaResource (%s/%s)", ns, name))
	err := r.Get(context.TODO(), request.NamespacedName, alamedaresource)
	if err != nil {
		if errors.IsNotFound(err) {
			//delete Alameda Resource Predict CR due to no Alameda CR with the same name exist
			alamedaResourcePrediction := &autoscalingv1alpha1.AlamedaResourcePrediction{}
			err = r.Get(context.TODO(), request.NamespacedName, alamedaResourcePrediction)
			if err == nil {
				r.Delete(context.TODO(), alamedaResourcePrediction)
			}
		}
		//logUtil.GetLogger().Info(fmt.Sprintf("Get AlamedaResource failed. (%s/%s)", ns, name))
		//return reconcile.Result{}, err

	} else {
		logUtil.GetLogger().Info(fmt.Sprintf("AlamedaResource found. (%s/%s)", ns, name))
		r.updateAlamedaResourceAnnotation(alamedaresource, ns)
		//check in alameda predict CR exist state
		alamedaResourcePrediction := &autoscalingv1alpha1.AlamedaResourcePrediction{}

		err := r.Get(context.TODO(), request.NamespacedName, alamedaResourcePrediction)
		if err != nil {
			if errors.IsNotFound(err) {
				newAlamedaResourcePrediction := &autoscalingv1alpha1.AlamedaResourcePrediction{
					ObjectMeta: metav1.ObjectMeta{
						Name:      request.Name,
						Namespace: request.Namespace,
					},
					Spec: autoscalingv1alpha1.AlamedaResourcePredictionSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: alamedaresource.Spec.Selector.MatchLabels,
						},
					},
					Status: autoscalingv1alpha1.AlamedaResourcePredictionStatus{
						Prediction: autoscalingv1alpha1.AlamedaPrediction{
							Deployments: map[autoscalingv1alpha1.DeploymentUID]autoscalingv1alpha1.PredictDeployment{},
						},
					},
				}
				if err := controllerutil.SetControllerReference(alamedaresource, newAlamedaResourcePrediction, r.scheme); err != nil {
					return reconcile.Result{}, err
				}

				r.Create(context.TODO(), newAlamedaResourcePrediction)
			}
		}
		deleteEvt = false
	}

	deploymentFound := &appsv1.Deployment{}
	err = r.Get(context.TODO(), request.NamespacedName, deploymentFound)
	if err != nil {
		if errors.IsNotFound(err) {
			//logUtil.GetLogger().Info(fmt.Sprintf("Deployment not found. (%s/%s)", ns, name))
			//return reconcile.Result{}, nil
		}
		logUtil.GetLogger().Info(fmt.Sprintf("Get Deployment for AlamedaResource controller failed. (%s/%s)", ns, name))
		//		return reconcile.Result{}, err
	} else {
		r.updateAlamedaK8SControllerByDeployment(ns, deploymentFound)
		deleteEvt = false
	}

	if deleteEvt {
		logUtil.GetLogger().Info(fmt.Sprintf("Delete event."))
		alamedaResourceList := &autoscalingv1alpha1.AlamedaResourceList{}
		err = r.List(context.TODO(),
			client.InNamespace(ns),
			alamedaResourceList)
		for _, ala := range alamedaResourceList.Items {
			r.updateAlamedaAnnotationByDeleteEvt(&ala, request)
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAlamedaResource) updateAlamedaResourceAnnotation(alamedaresource *autoscalingv1alpha1.AlamedaResource, ns string) {
	alamedaAnnotations := map[string]string{}
	newAlamedaAnnotations := map[string]string{}
	alamedaAnnotations = alamedaresource.GetAnnotations()
	if alamedaAnnotations == nil {
		newAlamedaAnnotations = map[string]string{}
	} else {
		for k, v := range alamedaAnnotations {
			newAlamedaAnnotations[k] = v
		}
	}
	if newAlamedaAnnotations[AlamedaK8sController] == "" {
		newAlamedaAnnotations[AlamedaK8sController] = alamedaK8sControllerDefautlAnno()
	}
	//find matched deployment controller
	matchedDeploymentList := &appsv1.DeploymentList{}
	err := r.List(context.TODO(),
		client.InNamespace(ns).
			MatchingLabels(alamedaresource.Spec.Selector.MatchLabels),
		matchedDeploymentList)
	if err == nil {
		akcMap := convertk8scontrollerJSONString(newAlamedaAnnotations[AlamedaK8sController])
		for _, deploy := range matchedDeploymentList.Items {
			akcMap.DeploymentMap[string(deploy.GetUID())] = *r.getControllerMapForAnno("deployment", &deploy).(*Deployment)
		}
		updatemd, _ := json.MarshalIndent(akcMap, "", JSONIndent)
		newAlamedaAnnotations[AlamedaK8sController] = string(updatemd)
	}
	if len(newAlamedaAnnotations) > 0 && !reflect.DeepEqual(newAlamedaAnnotations, alamedaAnnotations) {
		alamedaresource.SetAnnotations(newAlamedaAnnotations)
		r.Update(context.TODO(), alamedaresource)
	}
}

func (r *ReconcileAlamedaResource) updateAlamedaK8SControllerByDeployment(ns string, deploymentFound *appsv1.Deployment) {
	alamedaResourceList := &autoscalingv1alpha1.AlamedaResourceList{}
	r.List(context.TODO(),
		client.InNamespace(ns),
		alamedaResourceList)
	for _, ala := range alamedaResourceList.Items {
		r.updateAlamedaAnnotationByDeployment(&ala, deploymentFound)
	}
}

func (r *ReconcileAlamedaResource) updateAlamedaAnnotationByDeleteEvt(ala *autoscalingv1alpha1.AlamedaResource, request reconcile.Request) {
	needUpdated := false
	name := request.Name
	anno := ala.GetAnnotations()
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

func (r *ReconcileAlamedaResource) updateAlamedaAnnotationByDeployment(ala *autoscalingv1alpha1.AlamedaResource, deploy *appsv1.Deployment) {
	needUpdated := false
	alaML := ala.Spec.Selector.MatchLabels
	dL := deploy.GetLabels()
	dpUID := deploy.GetUID()
	anno := ala.GetAnnotations()
	if anno == nil {
		anno[AlamedaK8sController] = alamedaK8sControllerDefautlAnno()
	}
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

		logUtil.GetLogger().Info(fmt.Sprintf("Alameda Deployment Pods to add %s. (%s/%s).", string(newPodMapsBin), deploy.GetNamespace(), deploy.GetName()))
		logUtil.GetLogger().Info(fmt.Sprintf("Alameda Deployment Pods to delete %s. (%s/%s).", string(deletePodMapsBin), deploy.GetNamespace(), deploy.GetName()))
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
		updated, _ := json.MarshalIndent(k8sc, "", JSONIndent)
		anno[AlamedaK8sController] = string(updated)
		ala.SetAnnotations(anno)
		err := r.Update(context.TODO(), ala)
		if err != nil {
			logUtil.GetLogger().Error(err, fmt.Sprintf("Update Annotation failed"))
			return
		}

		conn, err := grpc.Dial(grpcutils.GetAIServiceAddress(), grpc.WithInsecure())
		if err != nil {
			logUtil.GetLogger().Error(err, fmt.Sprintf("Connect to AI server failed"))
			return
		}

		defer conn.Close()
		aiServiceClnt := aiservice_v1alpha1.NewAlamendaAIServiceClient(conn)
		if len(newPodMaps) > 0 {
			req := aiservice_v1alpha1.PredictionObjectListCreationRequest{
				Objects: []*aiservice_v1alpha1.Object{},
			}
			for _, v := range newPodMaps {
				req.Objects = append(req.Objects, &aiservice_v1alpha1.Object{
					Type:      aiservice_v1alpha1.Object_POD,
					Uid:       v.UID,
					Namespace: deploy.GetNamespace(),
					Name:      v.Name,
				})
			}
			reqBin, _ := json.MarshalIndent(req, "", JSONIndent)
			logUtil.GetLogger().Info(fmt.Sprintf("Create prediction object %s to AI server. (%s/%s).", string(reqBin), deploy.GetNamespace(), deploy.GetName()))
			aiServiceClnt.CreatePredictionObjects(context.Background(), &req)
		}
		if len(deletePodMaps) > 0 {
			req := aiservice_v1alpha1.PredictionObjectListDeletionRequest{
				Objects: []*aiservice_v1alpha1.Object{},
			}
			for _, v := range deletePodMaps {
				req.Objects = append(req.Objects, &aiservice_v1alpha1.Object{
					Type:      aiservice_v1alpha1.Object_POD,
					Uid:       v.UID,
					Namespace: deploy.GetNamespace(),
					Name:      v.Name,
				})
			}
			reqBin, _ := json.MarshalIndent(req, "", JSONIndent)
			logUtil.GetLogger().Info(fmt.Sprintf("Delete prediction object %s to AI server. (%s/%s).", string(reqBin), deploy.GetNamespace(), deploy.GetName()))
			aiServiceClnt.DeletePredictionObjects(context.Background(), &req)
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

// GetDefaultAlamedaK8SControllerAnno get default AlamedaResource annotation of K8S controller
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
		logUtil.GetLogger().Error(err, fmt.Sprintf("Json string decode failed"))
	}
	return akcMap
}

func (r *ReconcileAlamedaResource) getControllerMapForAnno(kind string, deploy interface{}) interface{} {
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
				UID:        string(pod.GetUID()),
				Containers: containers,
			}
		}
		return &Deployment{
			Name:   name,
			UID:    string(deploy.(*appsv1.Deployment).GetUID()),
			PodMap: podMap,
		}
	}
	return nil
}
