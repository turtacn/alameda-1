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
	logUtil.GetLogger().Info(fmt.Sprintf("Start reconciling."))
	alamedaAnnotations := map[string]string{}
	newAlamedaAnnotations := map[string]string{}
	// Fetch the AlamedaResource instance
	deleteEvt := true
	ns := request.Namespace
	name := request.Name
	alamedaresource := &autoscalingv1alpha1.AlamedaResource{}
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
			akcMap.(map[string]map[string]map[string]string)["deployment"] = map[string]map[string]string{}
			for _, deploy := range matchedDeploymentList.Items {
				akcMap.(map[string]map[string]map[string]string)["deployment"][string(deploy.GetUID())] = getControllerMapForAnno("deployment", &deploy)
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
		//logUtil.GetLogger().Error(err, fmt.Sprintf("Get Deployment failed. (%s/%s)", ns, name))
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
		k8sc := convertk8scontrollerJsonString(anno[AlamedaK8sController]).(map[string]map[string]map[string]string)
		//handle deployment controller
		for k, v := range k8sc["deployment"] {
			if v["name"] == name {
				delete(k8sc["deployment"], k)
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
	k8sc := convertk8scontrollerJsonString(anno[AlamedaK8sController]).(map[string]map[string]map[string]string)
	if isLabelsMatched(dL, alaML) {
		if k8sc["deployment"][string(dpUID)] == nil {
			k8sc["deployment"][string(dpUID)] = getControllerMapForAnno("deployment", deploy)
			logUtil.GetLogger().Info(fmt.Sprintf("Alameda Deployment found. (%s/%s).", deploy.GetNamespace(), deploy.GetName()))
			needUpdated = true
		}
	} else {
		if k8sc["deployment"][string(dpUID)] == nil {
			delete(k8sc["deployment"], string(deploy.GetUID()))
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
	emp := map[string]interface{}{}
	emp["deployment"] = map[string]interface{}{}
	md, _ := json.Marshal(emp)
	return string(md)
}

//annotation-k8s-controller annotation struct definition
func convertk8scontrollerJsonString(jsonStr string) interface{} {
	akcMap := map[string]map[string]map[string]string{
		"deployment": {},
	}
	err := json.Unmarshal([]byte(jsonStr), &akcMap)
	if err != nil {
		logUtil.GetLogger().Error(err, fmt.Sprintf("Json string decode failed"))
	}
	return akcMap
}

func getControllerMapForAnno(kind string, deploy interface{}) map[string]string {
	if kind == "deployment" {
		return map[string]string{
			"name": deploy.(*appsv1.Deployment).GetName(),
			"uid":  string(deploy.(*appsv1.Deployment).GetUID())}
	}
	return nil
}
