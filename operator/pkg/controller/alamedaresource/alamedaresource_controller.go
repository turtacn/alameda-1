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
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	appsv1 "k8s.io/api/apps/v1"

	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AlamedaResource is alameda resource
type AlamedaResource string

const (
	AlamedaDeployment AlamedaResource = "Deployment"
)
const AlamedaK8sController = "annotation-k8s-controller"

type Pod struct {
	UID  string
	Name string
}

type Deployment struct {
	UID    string
	Name   string
	PodMap map[string]Pod
}

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
		logUtil.GetLogger().Error(err, fmt.Sprintf("Watch Deployment controller failed."))
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
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.containers.ai,resources=alamedaresources,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileAlamedaResource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	alamedaAnnotations := map[string]string{}
	newAlamedaAnnotations := map[string]string{}
	// Fetch the AlamedaResource instance
	deleteEvt := true
	ns := request.Namespace
	name := request.Name
	alamedaresource := &autoscalingv1alpha1.AlamedaResource{}
	logUtil.GetLogger().Info(fmt.Sprintf("Try to get AlamedaResource (%s/%s)", ns, name))
	err := r.Get(context.TODO(), request.NamespacedName, alamedaresource)
	if err != nil {
		if errors.IsNotFound(err) {
			//logUtil.GetLogger().Info(fmt.Sprintf("AlamedaResource not found. (%s/%s)", ns, name))
			//return reconcile.Result{}, nil
		}
		//logUtil.GetLogger().Info(fmt.Sprintf("Get AlamedaResource failed. (%s/%s)", ns, name))
		//return reconcile.Result{}, err

	} else {
		logUtil.GetLogger().Info(fmt.Sprintf("AlamedaResource found. (%s/%s)", ns, name))

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
			akcMap := convertk8scontrollerJsonString(newAlamedaAnnotations[AlamedaK8sController])
			for _, deploy := range matchedDeploymentList.Items {
				akcMap.DeploymentMap[string(deploy.GetUID())] = *r.getControllerMapForAnno("deployment", &deploy).(*Deployment)
			}
			updatemd, _ := json.Marshal(akcMap)
			newAlamedaAnnotations[AlamedaK8sController] = string(updatemd)
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
		alamedaResourceList := &autoscalingv1alpha1.AlamedaResourceList{}
		err = r.List(context.TODO(),
			client.InNamespace(ns),
			alamedaResourceList)
		for _, ala := range alamedaResourceList.Items {
			r.updateAlamedaAnnotationByDeployment(&ala, deploymentFound)
		}
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
	} else if len(newAlamedaAnnotations) > 0 && !reflect.DeepEqual(newAlamedaAnnotations, alamedaAnnotations) {
		alamedaresource.SetAnnotations(newAlamedaAnnotations)
		err = r.Update(context.TODO(), alamedaresource)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAlamedaResource) updateAlamedaAnnotationByDeleteEvt(ala *autoscalingv1alpha1.AlamedaResource, request reconcile.Request) {
	needUpdated := false
	name := request.Name
	anno := ala.GetAnnotations()
	if anno != nil && anno[AlamedaK8sController] != "" {
		k8sc := convertk8scontrollerJsonString(anno[AlamedaK8sController])
		//handle deployment controller
		for k, v := range k8sc.DeploymentMap {
			if v.Name == name {
				delete(k8sc.DeploymentMap, k)
				needUpdated = true
			}
		}
		if needUpdated {
			updated, _ := json.Marshal(k8sc)
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
	k8sc := convertk8scontrollerJsonString(anno[AlamedaK8sController])
	if isLabelsMatched(dL, alaML) {
		if _, found := k8sc.DeploymentMap[string(dpUID)]; !found {
			k8sc.DeploymentMap[string(dpUID)] = *r.getControllerMapForAnno("deployment", deploy).(*Deployment)
			logUtil.GetLogger().Info(fmt.Sprintf("Alameda Deployment found. (%s/%s).", deploy.GetNamespace(), deploy.GetName()))
			needUpdated = true
		}
	} else {
		if _, found := k8sc.DeploymentMap[string(dpUID)]; found {
			delete(k8sc.DeploymentMap, string(deploy.GetUID()))
			needUpdated = true
		}
	}
	if needUpdated {
		updated, _ := json.Marshal(k8sc)
		anno[AlamedaK8sController] = string(updated)
		ala.SetAnnotations(anno)
		_ = r.Update(context.TODO(), ala)
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
	md, _ := json.Marshal(*getDefaultAlamedaK8SControllerAnno())
	return string(md)
}

func getDefaultAlamedaK8SControllerAnno() *K8SControllerAnnotation {
	return &K8SControllerAnnotation{
		DeploymentMap: map[string]Deployment{},
	}
}

//annotation-k8s-controller annotation struct definition
func convertk8scontrollerJsonString(jsonStr string) *K8SControllerAnnotation {
	akcMap := getDefaultAlamedaK8SControllerAnno()
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
			podMap[string(pod.GetUID())] = Pod{
				Name: pod.GetName(),
				UID:  string(pod.GetUID()),
			}
		}
		return &Deployment{
			Name:   name,
			UID:    string(deploy.(*appsv1.Deployment).GetUID()),
			PodMap: podMap}
	}
	return nil
}
