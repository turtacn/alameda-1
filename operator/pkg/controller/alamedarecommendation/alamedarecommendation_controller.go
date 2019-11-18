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

package alamedarecommendation

import (
	"context"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/datahub/pkg/utils"
	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	k8sutils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_recommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	scope        = logUtil.RegisterScope("alamedarecommendation", "alameda recommendation", 0)
	requeueAfter = 5 * time.Second
)

var cachedFirstSynced = false

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AlamedaRecommendation Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {

	conn, _ := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())

	k8sClient, err := client.New(mgr.GetConfig(), client.Options{})
	if err != nil {
		panic(errors.Wrap(err, "new kubernetes client failed"))
	}
	clusterUID, err := k8sutils.GetClusterUID(k8sClient)
	if err != nil {
		panic(errors.Wrap(err, "get cluster uid failed"))
	}

	return &ReconcileAlamedaRecommendation{
		Client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		datahubClient: datahub_v1alpha1.NewDatahubServiceClient(conn),

		clusterUID: clusterUID,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("alamedarecommendation-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to AlamedaRecommendation
	err = c.Watch(&source.Kind{Type: &autoscalingv1alpha1.AlamedaRecommendation{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAlamedaRecommendation{}

// ReconcileAlamedaRecommendation reconciles a AlamedaRecommendation object
type ReconcileAlamedaRecommendation struct {
	client.Client
	scheme        *runtime.Scheme
	datahubClient datahub_v1alpha1.DatahubServiceClient

	clusterUID string
}

// Reconcile reads that state of the cluster for a AlamedaRecommendation object and makes changes based on the state read
// and what is in the AlamedaRecommendation.Spec
func (r *ReconcileAlamedaRecommendation) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if !cachedFirstSynced {
		time.Sleep(5 * time.Second)
	}
	cachedFirstSynced = true

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	alamedaRecommendation := autoscalingv1alpha1.AlamedaRecommendation{}
	err := r.Client.Get(ctx, request.NamespacedName, &alamedaRecommendation)
	if err != nil && !k8serrors.IsNotFound(err) {
		scope.Warnf("Get AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", request.Namespace, request.Name, requeueAfter.Seconds(), err.Error())
		return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if k8serrors.IsNotFound(err) {
		return reconcile.Result{Requeue: false}, nil
	}

	// Delete AlamedaRecommendation if it is not watching by any AlamedaScaler
	alamedaScaler, err := r.getWatchingAlamedaScaler(ctx, alamedaRecommendation)
	if err != nil {
		scope.Warnf("Get watching AlamedaScaler failed: retry after %f seconds, errorMsg: %s", requeueAfter.Seconds(), err.Error())
		return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if alamedaScaler.Namespace == "" || alamedaScaler.Name == "" {
		if err = r.Delete(ctx, &alamedaRecommendation); err != nil {
			scope.Warnf("Delete AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", alamedaRecommendation.Namespace, alamedaRecommendation.Name, requeueAfter.Seconds(), err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
		return reconcile.Result{}, nil
	} else if !alamedaScaler.HasAlamedaPod(alamedaRecommendation.Namespace, alamedaRecommendation.Name) {
		if err = r.Delete(ctx, &alamedaRecommendation); err != nil {
			scope.Warnf("Delete AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", request.Namespace, request.Name, requeueAfter.Seconds(), err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
		return reconcile.Result{}, nil
	}

	// Update this AlamedaRecommendation with the latest recommendation value from Datahub
	resp, err := r.datahubClient.ListPodRecommendations(ctx, &datahub_recommendations.ListPodRecommendationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				Namespace:   request.Namespace,
				Name:        request.Name,
				ClusterName: r.clusterUID,
			},
		},
		Kind: datahub_resources.Kind_POD,
		QueryCondition: &datahub_common.QueryCondition{
			TimeRange: &datahub_common.TimeRange{
				EndTime: ptypes.TimestampNow(),
			},
			Order: datahub_common.QueryCondition_DESC,
			Limit: 1,
		},
	})
	if err != nil {
		scope.Warnf("List PodRecommendations from datahub failed, retry after %f seconds, errorMsg: %s", requeueAfter.Seconds(), err.Error())
		return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if resp == nil || resp.Status == nil {
		scope.Warnf("List PodRecommendations from datahub failed, receive nil response.Status, retry after %f seconds.", requeueAfter.Seconds())
		return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if resp.Status.Code != 0 {
		scope.Warnf("List PodRecommendations from datahub failed, receive receive code: %d, message: %s, retry after %f seconds.", resp.Status.Code, resp.Status.Message, requeueAfter.Seconds())
		return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	}
	if len(resp.PodRecommendations) >= 1 && resp.PodRecommendations[0] != nil {
		alamedaRecommendation, err := r.updateAlamedaRecommendationWithLatestDatahubPodRecommendation(alamedaRecommendation, *resp.PodRecommendations[0])
		if err != nil {
			scope.Warnf("Update AlamedaRecommendation(%s/%s) with databub PodRecommendation failed, retry after %f seconds, errorMsg: %s", request.Namespace, request.Name, requeueAfter.Seconds(), err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
		if err = r.Client.Update(ctx, &alamedaRecommendation); err != nil {
			scope.Warnf("Update AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", request.Namespace, request.Name, requeueAfter.Seconds(), err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
	}

	return reconcile.Result{Requeue: false}, nil
}

func (r *ReconcileAlamedaRecommendation) getWatchingAlamedaScaler(ctx context.Context, alamedaRecommendation autoscalingv1alpha1.AlamedaRecommendation) (autoscalingv1alpha1.AlamedaScaler, error) {

	for _, or := range alamedaRecommendation.OwnerReferences {
		if or.Controller != nil && *or.Controller && strings.ToLower(or.Kind) == "alamedascaler" {
			alamedaScaler := autoscalingv1alpha1.AlamedaScaler{}
			namespace := alamedaRecommendation.Namespace
			name := or.Name
			err := r.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &alamedaScaler)
			if err != nil && !k8serrors.IsNotFound(err) {
				return autoscalingv1alpha1.AlamedaScaler{}, errors.Wrapf(err, "get AlamedaScaler(%s/%s) failed", namespace, name)
			}
			return alamedaScaler, nil
		}
	}
	return autoscalingv1alpha1.AlamedaScaler{}, nil
}

func (r *ReconcileAlamedaRecommendation) updateAlamedaRecommendationWithLatestDatahubPodRecommendation(alamedaRecommendation autoscalingv1alpha1.AlamedaRecommendation, podRecommendation datahub_recommendations.PodRecommendation) (autoscalingv1alpha1.AlamedaRecommendation, error) {

	for i, container := range alamedaRecommendation.Spec.Containers {
		for _, containerRecommendation := range podRecommendation.ContainerRecommendations {
			if containerRecommendation == nil || container.Name != containerRecommendation.Name {
				continue
			}
			if container.Resources.Limits == nil {
				container.Resources.Limits = corev1.ResourceList{}
			}
			if container.Resources.Requests == nil {
				container.Resources.Requests = corev1.ResourceList{}
			}
			for _, metricData := range containerRecommendation.LimitRecommendations {
				switch metricData.MetricType {
				case datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
					latestTime := int64(0)
					for _, data := range metricData.Data {
						dataTime := utils.TimeStampToNanoSecond(data.Time)
						if numVal, err := utils.StringToInt64(data.NumValue); err == nil && dataTime > latestTime {
							container.Resources.Limits[corev1.ResourceCPU] = *resource.NewMilliQuantity(numVal, resource.DecimalSI)
							latestTime = dataTime
						} else if err != nil {
							return alamedaRecommendation, errors.Wrap(err, "convert data value failed")
						}
					}
				case datahub_common.MetricType_MEMORY_USAGE_BYTES:
					latestTime := int64(0)
					for _, data := range metricData.Data {
						dataTime := utils.TimeStampToNanoSecond(data.Time)
						if numVal, err := utils.StringToInt64(data.NumValue); err == nil && dataTime > latestTime {
							container.Resources.Limits[corev1.ResourceMemory] = *resource.NewQuantity(numVal, resource.BinarySI)
							latestTime = dataTime
						} else if err != nil {
							return alamedaRecommendation, errors.Wrap(err, "convert data value failed")
						}
					}
				}
			}
			for _, metricData := range containerRecommendation.RequestRecommendations {
				switch metricData.MetricType {
				case datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
					latestTime := int64(0)
					for _, data := range metricData.Data {
						dataTime := utils.TimeStampToNanoSecond(data.Time)
						if numVal, err := utils.StringToInt64(data.NumValue); err == nil && dataTime > latestTime {
							container.Resources.Requests[corev1.ResourceCPU] = *resource.NewMilliQuantity(numVal, resource.DecimalSI)
							latestTime = dataTime
						} else if err != nil {
							return alamedaRecommendation, errors.Wrap(err, "convert data value failed")
						}
					}
				case datahub_common.MetricType_MEMORY_USAGE_BYTES:
					latestTime := int64(0)
					for _, data := range metricData.Data {
						dataTime := utils.TimeStampToNanoSecond(data.Time)
						if numVal, err := utils.StringToInt64(data.NumValue); err == nil && dataTime > latestTime {
							container.Resources.Requests[corev1.ResourceMemory] = *resource.NewQuantity(numVal, resource.BinarySI)
							latestTime = dataTime
						} else if err != nil {
							return alamedaRecommendation, errors.Wrap(err, "convert data value failed")
						}
					}
				}
			}
		}
		alamedaRecommendation.Spec.Containers[i] = container
	}
	return alamedaRecommendation, nil
}
