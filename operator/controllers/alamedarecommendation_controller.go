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

package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/containers-ai/alameda/datahub/pkg/utils"
	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_recommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var alamedarecommendationFirstSynced = false

// AlamedaRecommendationReconciler reconciles a AlamedaRecommendation object
type AlamedaRecommendationReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	DatahubClient datahub_v1alpha1.DatahubServiceClient

	ClusterUID string
}

// Reconcile reads that state of the cluster for a AlamedaRecommendation object and makes changes based on the state read
// and what is in the AlamedaRecommendation.Spec
func (r *AlamedaRecommendationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	requeueAfter := 5 * time.Second
	if !alamedarecommendationFirstSynced {
		time.Sleep(5 * time.Second)
	}
	alamedarecommendationFirstSynced = true

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	alamedaRecommendation := autoscalingv1alpha1.AlamedaRecommendation{}
	err := r.Client.Get(ctx, req.NamespacedName, &alamedaRecommendation)
	if err != nil && !k8serrors.IsNotFound(err) {
		scope.Warnf("Get AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", req.Namespace, req.Name, requeueAfter.Seconds(), err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if k8serrors.IsNotFound(err) {
		return ctrl.Result{Requeue: false}, nil
	}

	// Delete AlamedaRecommendation if it is not watching by any AlamedaScaler
	alamedaScaler, err := r.getWatchingAlamedaScaler(ctx, alamedaRecommendation)
	if err != nil {
		scope.Warnf("Get watching AlamedaScaler failed: retry after %f seconds, errorMsg: %s", requeueAfter.Seconds(), err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if alamedaScaler.Namespace == "" || alamedaScaler.Name == "" {
		if err = r.Delete(ctx, &alamedaRecommendation); err != nil {
			scope.Warnf("Delete AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", alamedaRecommendation.Namespace, alamedaRecommendation.Name, requeueAfter.Seconds(), err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
		return ctrl.Result{}, nil
	} else if !alamedaScaler.HasAlamedaPod(alamedaRecommendation.Namespace, alamedaRecommendation.Name) {
		if err = r.Delete(ctx, &alamedaRecommendation); err != nil {
			scope.Warnf("Delete AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", req.Namespace, req.Name, requeueAfter.Seconds(), err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
		return ctrl.Result{}, nil
	}

	// Update this AlamedaRecommendation with the latest recommendation value from Datahub
	resp, err := r.DatahubClient.ListPodRecommendations(ctx, &datahub_recommendations.ListPodRecommendationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				Namespace:   req.Namespace,
				Name:        req.Name,
				ClusterName: r.ClusterUID,
			},
		},
		Kind: datahub_resources.Kind_KIND_UNDEFINED,
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
		return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if resp == nil || resp.Status == nil {
		scope.Warnf("List PodRecommendations from datahub failed, receive nil response.Status, retry after %f seconds.", requeueAfter.Seconds())
		return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	} else if resp.Status.Code != 0 {
		scope.Warnf("List PodRecommendations from datahub failed, receive receive code: %d, message: %s, retry after %f seconds.", resp.Status.Code, resp.Status.Message, requeueAfter.Seconds())
		return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	}
	if len(resp.PodRecommendations) >= 1 && resp.PodRecommendations[0] != nil {
		alamedaRecommendation, err := r.updateAlamedaRecommendationWithLatestDatahubPodRecommendation(alamedaRecommendation, *resp.PodRecommendations[0])
		if err != nil {
			scope.Warnf("Update AlamedaRecommendation(%s/%s) with databub PodRecommendation failed, retry after %f seconds, errorMsg: %s", req.Namespace, req.Name, requeueAfter.Seconds(), err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
		if err = r.Client.Update(ctx, &alamedaRecommendation); err != nil {
			scope.Warnf("Update AlamedaRecommendation(%s/%s) failed, retry after %f seconds, errorMsg: %s", req.Namespace, req.Name, requeueAfter.Seconds(), err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *AlamedaRecommendationReconciler) getWatchingAlamedaScaler(ctx context.Context, alamedaRecommendation autoscalingv1alpha1.AlamedaRecommendation) (autoscalingv1alpha1.AlamedaScaler, error) {

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

func (r *AlamedaRecommendationReconciler) updateAlamedaRecommendationWithLatestDatahubPodRecommendation(alamedaRecommendation autoscalingv1alpha1.AlamedaRecommendation, podRecommendation datahub_recommendations.PodRecommendation) (autoscalingv1alpha1.AlamedaRecommendation, error) {

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

func (r *AlamedaRecommendationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingv1alpha1.AlamedaRecommendation{}).
		Complete(r)
}
