package v1alpha1

import (
	"fmt"
)

const (

	// MonitoringAlamedaScalerIDKey is key of label to identify which AlamedaScaler is the resource belongs to.
	MonitoringAlamedaScalerIDKey = "containers.ai/alamedascaler.name"
	// MonitoringAlamedaScalerIDValueFormat is format of the value, format arguments:
	// 1. Name of AlamedaScaler
	// 2. Namespace of AlamedaScaler
	// example: alamedascaler-test.webapp
	MonitoringAlamedaScalerIDValueFormat = "%s.%s"

	// MonitoringAlamedaScalerAlamedaControllerIDKey is key of label to identify which workload controller of which AlamedaScaler is the resource belongs to.
	MonitoringAlamedaScalerAlamedaControllerIDKey = "containers.ai/alamedascaler.status.alamedacontroller.name"
	// MonitoringAlamedaScalerAlamedaControllerIDValueFormat is format of the value, format arguments:
	// 1. Name of AlamedaScaler
	// 2. Type of workload controller
	// 3. Name of workload controller
	// example: alamedascaler-test.deploymentconfig.nginx
	MonitoringAlamedaScalerAlamedaControllerIDValueFormat = "%s.%s.%s"
)

func GenerateMonitoringAlamedaScalerIdentityLabels(alamedaScalerNamespace, alamedaScalerName string) map[string]string {

	labels := make(map[string]string)

	idKey := MonitoringAlamedaScalerIDKey
	idValue := fmt.Sprintf(MonitoringAlamedaScalerIDValueFormat, alamedaScalerName, alamedaScalerNamespace)

	labels[idKey] = idValue

	return labels
}

func GenerateMonitoringAlamedaScalerAlamedaControllerIdentityLabels(alamedaScalerName string, controllerName string, controllerType AlamedaControllerType) map[string]string {

	labels := make(map[string]string)

	idKey := MonitoringAlamedaScalerAlamedaControllerIDKey
	idValue := fmt.Sprintf(MonitoringAlamedaScalerAlamedaControllerIDValueFormat, alamedaScalerName, AlamedaControllerTypeName[controllerType], controllerName)

	labels[idKey] = idValue

	return labels
}
