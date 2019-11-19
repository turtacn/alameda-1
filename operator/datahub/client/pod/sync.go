package pod

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	k8sutils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SyncWithDatahub(k8sClient client.Client, conn *grpc.ClientConn) error {
	err := deleteRedudantPodFromDatahub(k8sClient, conn)
	if err != nil {
		return errors.Wrap(err, "delete redudant pods from Datahub failed")
	}
	return nil
}

func deleteRedudantPodFromDatahub(k8sClient client.Client, conn *grpc.ClientConn) error {

	clusterUID, err := k8sutils.GetClusterUID(k8sClient)
	if err != nil {
		return errors.Wrap(err, "get cluster uid failed")
	}
	datahubPodRepo := NewPodRepository(conn, clusterUID)
	pods, err := datahubPodRepo.ListAlamedaPods()
	if err != nil {
		return errors.Wrap(err, "list pods from Datahub failed")
	}

	podsNeedToBeDeleted := []*datahub_resources.ObjectMeta{}
	for _, pod := range pods {
		if pod == nil || pod.ObjectMeta == nil || pod.ObjectMeta.Namespace == "" || pod.ObjectMeta.Name == "" {
			continue
		}

		p := corev1.Pod{}
		namespace := pod.ObjectMeta.Namespace
		name := pod.ObjectMeta.Name
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, &p)
		if err != nil && k8serrors.IsNotFound(err) {
			podsNeedToBeDeleted = append(podsNeedToBeDeleted, pod.ObjectMeta)
			continue
		} else if err != nil {
			return errors.Wrapf(err, "get Pod(%s/%s) failed", namespace, name)
		}

		if exist, err := isMonitoringAlamedaScalerOfPodExist(k8sClient, *pod); err != nil {
			return errors.Wrapf(err, "check if monitoring AlamedaScaler of Pod(%s/%s) is exist failed", namespace, name)
		} else if !exist {
			podsNeedToBeDeleted = append(podsNeedToBeDeleted, pod.ObjectMeta)
		}
	}

	if len(podsNeedToBeDeleted) <= 0 {
		return nil
	}
	if err := datahubPodRepo.DeletePods(context.TODO(), podsNeedToBeDeleted); err != nil {
		return errors.Wrap(err, "delete pods from datahub failed")
	}
	return nil
}

func isMonitoringAlamedaScalerOfPodExist(k8sClient client.Client, pod datahub_resources.Pod) (bool, error) {

	if pod.AlamedaPodSpec == nil || pod.AlamedaPodSpec.AlamedaScaler == nil ||
		pod.AlamedaPodSpec.AlamedaScaler.Namespace == "" || pod.AlamedaPodSpec.AlamedaScaler.Name == "" {
		return false, nil
	}

	alamedaScaler := autoscalingv1alpha1.AlamedaScaler{}
	namespace := pod.AlamedaPodSpec.AlamedaScaler.Namespace
	name := pod.AlamedaPodSpec.AlamedaScaler.Name
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, &alamedaScaler)
	if err != nil && k8serrors.IsNotFound(err) {
		return false, errors.Wrapf(err, "get AlamedaScaler(%s/%s) failed", namespace, name)
	} else if k8serrors.IsNotFound(err) {
		return false, nil
	}

	return true, nil
}
