package webhook

import (
	"context"
	"net/http"

	"github.com/containers-ai/alameda/pkg/utils"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var scope = logUtil.RegisterScope("operator_webhook", "Operator K8S webhook.", 0)

func HandleDeployment(decoder *admission.Decoder, client client.Client, ctx context.Context, req admission.Request) admission.Response {
	deployment, err := getDeploymentTypeMetaAndObjectMeta(decoder, req)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	scope.Debugf("Deployment received to validate as following %s", utils.InterfaceToString(deployment))
	res, err := validateDeploymentsFn(client, ctx, deployment)

	if err != nil {
		return admission.ValidationResponse(res, err.Error())
	}
	return admission.ValidationResponse(res, "")
}

func getDeploymentTypeMetaAndObjectMeta(decoder *admission.Decoder, req admission.Request) (struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}, error) {
	deployment := struct {
		metav1.TypeMeta
		metav1.ObjectMeta
	}{}

	switch req.AdmissionRequest.Kind.Group {
	case extensionsv1.GroupName:
		switch req.AdmissionRequest.Kind.Version {
		case extensionsv1.SchemeGroupVersion.Version:

			extensionsDep := &extensionsv1.Deployment{}
			err := decoder.Decode(req, extensionsDep)
			if err != nil {
				return deployment, errors.Errorf("decode deployment failed: %s", err.Error())
			}
			deployment.TypeMeta = extensionsDep.TypeMeta
			deployment.ObjectMeta = extensionsDep.ObjectMeta
		default:
			return deployment, errors.Errorf("get deployment TypeMeta and ObjectMeta failed, received GroupVersionKind not supported, GroupVersionKind: %s", req.AdmissionRequest.Kind.String())
		}
	case appsv1.GroupName:
		switch req.AdmissionRequest.Kind.Version {
		case appsv1.SchemeGroupVersion.Version:
			appsDep := &appsv1.Deployment{}
			err := decoder.Decode(req, appsDep)
			if err != nil {
				return deployment, errors.Errorf("decode deployment failed: %s", err.Error())
			}
			deployment.TypeMeta = appsDep.TypeMeta
			deployment.ObjectMeta = appsDep.ObjectMeta
		case appsv1beta1.SchemeGroupVersion.Version:
			appsDep := &appsv1beta1.Deployment{}
			err := decoder.Decode(req, appsDep)
			if err != nil {
				return deployment, errors.Errorf("decode deployment failed: %s", err.Error())
			}
			deployment.TypeMeta = appsDep.TypeMeta
			deployment.ObjectMeta = appsDep.ObjectMeta
		case appsv1beta2.SchemeGroupVersion.Version:
			appsDep := &appsv1beta2.Deployment{}
			err := decoder.Decode(req, appsDep)
			if err != nil {
				return deployment, errors.Errorf("decode deployment failed: %s", err.Error())
			}
			deployment.TypeMeta = appsDep.TypeMeta
			deployment.ObjectMeta = appsDep.ObjectMeta
		default:
			return deployment, errors.Errorf("get deployment TypeMeta and ObjectMeta failed, received GroupVersionKind not supported, GroupVersionKind: %s", req.AdmissionRequest.Kind.String())
		}
	}

	return deployment, nil
}

// validateDeploymentsFn validate the given deployment
func validateDeploymentsFn(client client.Client, ctx context.Context, deployment struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}) (bool, error) {
	return isTopControllerValid(&client, &validatingObject{
		namespace: deployment.GetNamespace(),
		name:      deployment.GetName(),
		kind:      deployment.GetObjectKind().GroupVersionKind().Kind,
		labels:    deployment.GetLabels(),
	})
}
