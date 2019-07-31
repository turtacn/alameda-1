package enumconv

import (
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

const (
	Pod              string = "Pod"
	Deployment       string = "Deployment"
	DeploymentConfig string = "DeploymentConfig"
	AlamedaScaler    string = "AlamedaScaler"
	StatefulSet      string = "StatefulSet"
)

var KindEnum map[string]datahub_v1alpha1.Kind = map[string]datahub_v1alpha1.Kind{
	Pod:              datahub_v1alpha1.Kind_POD,
	Deployment:       datahub_v1alpha1.Kind_DEPLOYMENT,
	DeploymentConfig: datahub_v1alpha1.Kind_DEPLOYMENTCONFIG,
	AlamedaScaler:    datahub_v1alpha1.Kind_ALAMEDASCALER,
	StatefulSet:      datahub_v1alpha1.Kind_STATEFULSET,
}

var KindDisp map[datahub_v1alpha1.Kind]string = map[datahub_v1alpha1.Kind]string{
	datahub_v1alpha1.Kind_POD:              Pod,
	datahub_v1alpha1.Kind_DEPLOYMENT:       Deployment,
	datahub_v1alpha1.Kind_DEPLOYMENTCONFIG: DeploymentConfig,
	datahub_v1alpha1.Kind_ALAMEDASCALER:    AlamedaScaler,
	datahub_v1alpha1.Kind_STATEFULSET:      StatefulSet,
}
