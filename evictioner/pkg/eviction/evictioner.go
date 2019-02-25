package eviction

import (
	"context"
	"fmt"
	"time"

	"github.com/containers-ai/alameda/pkg/utils"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/rpc/code"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	evictionerscope = logUtil.RegisterScope("evictioner", "alamedascaler evictioner", 0)
)

// Evictioner deletes pods which need to apply recommendation
type Evictioner struct {
	checkCycle  int64
	datahubClnt datahub_v1alpha1.DatahubServiceClient
	k8sClienit  client.Client
}

// NewEvictioner return Evictioner instance
func NewEvictioner(checkCycle int64,
	datahubClnt datahub_v1alpha1.DatahubServiceClient,
	k8sClienit client.Client) *Evictioner {
	return &Evictioner{
		checkCycle:  checkCycle,
		datahubClnt: datahubClnt,
		k8sClienit:  k8sClienit,
	}
}

// Start checking pods need to apply recommendation
func (evictioner *Evictioner) Start() {
	go evictioner.evictProcess()
}

func (evictioner *Evictioner) evictProcess() {
	for {
		appliablePodRecList, err := evictioner.listAppliablePodRecommendation()
		if err != nil {
			evictionerscope.Error(err.Error())
		}
		evictionerscope.Debugf("Applicable pod recommendation lists: %s", utils.InterfaceToString(appliablePodRecList))
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
				evictionerscope.Error(err.Error())
			}
			continue
		}
		err = evictioner.k8sClienit.Delete(context.TODO(), recPodIns)
		if err != nil {
			evictionerscope.Errorf("Evict pod (%s,%s) failed: %s", recPodIns.GetNamespace(), recPodIns.GetName(), err.Error())
		}
	}
}

func (evictioner *Evictioner) listAppliablePodRecommendation() ([]*datahub_v1alpha1.PodRecommendation, error) {
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
	evictionerscope.Debugf("Possible applicable pod recommendation lists: %s", utils.InterfaceToString(resp.GetPodRecommendations()))

	for _, rec := range resp.GetPodRecommendations() {
		if rec.GetStartTime().GetSeconds() < nowTimestamp && nowTimestamp > rec.GetEndTime().GetSeconds() {
			appliablePodRecList = append(appliablePodRecList, rec)
		}
	}
	return appliablePodRecList, nil
}
