package validate

import (
	autoscalingapi "github.com/containers-ai/alameda/operator/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AlamedaScalerValidate interface {
	IsScalerValid(client *client.Client,
		topCtl *autoscalingapi.ValidatingObject) (bool, error)
}

type DeploymentValidate interface {
	IsTopControllerValid(client *client.Client,
		topCtl *autoscalingapi.ValidatingObject) (bool, error)
}

type DeploymentConfigValidate interface {
	IsTopControllerValid(client *client.Client,
		topCtl *autoscalingapi.ValidatingObject) (bool, error)
}
