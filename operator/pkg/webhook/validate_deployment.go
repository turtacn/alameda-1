package webhook

import (
	"context"
	"net/http"

	"github.com/containers-ai/alameda/pkg/utils"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	admissiontypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var scope = logUtil.RegisterScope("operator_webhook", "Operator K8S webhook.", 0)

type deploymentLabeler struct {
	client  client.Client
	decoder admissiontypes.Decoder
}

var _ admission.Handler = &deploymentLabeler{}

func (labeler *deploymentLabeler) Handle(ctx context.Context, req admissiontypes.Request) admissiontypes.Response {
	deployment := &extensionsv1.Deployment{}

	err := labeler.decoder.Decode(req, deployment)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	scope.Debugf("Deployment received to validate as following %s", utils.InterfaceToString(deployment))
	res, err := labeler.validateDeploymentsFn(ctx, deployment)

	if err != nil {
		return admission.ValidationResponse(res, err.Error())
	}
	return admission.ValidationResponse(res, "")
}

var _ inject.Decoder = &deploymentLabeler{}

// InjectDecoder injects the decoder into the deploymentLabeler
func (labeler *deploymentLabeler) InjectDecoder(d admissiontypes.Decoder) error {
	labeler.decoder = d
	return nil
}

var _ inject.Client = &deploymentLabeler{}

// InjectClient injects the client into the deploymentLabeler
func (labeler *deploymentLabeler) InjectClient(c client.Client) error {
	labeler.client = c
	return nil
}

// validateDeploymentsFn validate the given deployment
func (labeler *deploymentLabeler) validateDeploymentsFn(ctx context.Context, deployment *extensionsv1.Deployment) (bool, error) {
	return isTopControllerValid(&labeler.client, &validatingObject{
		namespace: deployment.GetNamespace(),
		name:      deployment.GetName(),
		kind:      deployment.GetObjectKind().GroupVersionKind().Kind,
		labels:    deployment.GetLabels(),
	})
}

func GetDeploymentHandler() *deploymentLabeler {
	return &deploymentLabeler{}
}
