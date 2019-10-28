package enumconv

import (
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

const (
	Pod              string = "Pod"
	Deployment       string = "Deployment"
	DeploymentConfig string = "DeploymentConfig"
	AlamedaScaler    string = "AlamedaScaler"
	StatefulSet      string = "StatefulSet"
)

var KindEnum map[string]ApiResources.Kind = map[string]ApiResources.Kind{
	Pod:              ApiResources.Kind_POD,
	Deployment:       ApiResources.Kind_DEPLOYMENT,
	DeploymentConfig: ApiResources.Kind_DEPLOYMENTCONFIG,
	AlamedaScaler:    ApiResources.Kind_ALAMEDASCALER,
	StatefulSet:      ApiResources.Kind_STATEFULSET,
}

var KindDisp map[ApiResources.Kind]string = map[ApiResources.Kind]string{
	ApiResources.Kind_POD:              Pod,
	ApiResources.Kind_DEPLOYMENT:       Deployment,
	ApiResources.Kind_DEPLOYMENTCONFIG: DeploymentConfig,
	ApiResources.Kind_ALAMEDASCALER:    AlamedaScaler,
	ApiResources.Kind_STATEFULSET:      StatefulSet,
}
