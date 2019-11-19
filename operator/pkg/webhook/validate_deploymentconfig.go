package webhook

import (
	"context"
	"net/http"

	"github.com/containers-ai/alameda/pkg/utils"
	osappsapi "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func HandleDeploymentConfig(decoder *admission.Decoder, client client.Client,
	ctx context.Context, req admission.Request) admission.Response {
	deploymentConfig := &osappsapi.DeploymentConfig{}

	err := decoder.Decode(req, deploymentConfig)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	scope.Debugf("DeploymentConfig received to validate as following %s",
		utils.InterfaceToString(deploymentConfig))
	res, err := validateDeploymentConfigsFn(client, ctx, deploymentConfig)
	if err != nil {
		return admission.ValidationResponse(res, err.Error())
	}
	return admission.ValidationResponse(res, "")
}

// validateDeploymentsFn validate the given deploymentConfig
func validateDeploymentConfigsFn(client client.Client, ctx context.Context,
	deploymentConfig *osappsapi.DeploymentConfig) (bool, error) {
	return isTopControllerValid(&client, &validatingObject{
		namespace: deploymentConfig.GetNamespace(),
		name:      deploymentConfig.GetName(),
		kind:      deploymentConfig.GetObjectKind().GroupVersionKind().Kind,
		labels:    deploymentConfig.GetLabels(),
	})
}
