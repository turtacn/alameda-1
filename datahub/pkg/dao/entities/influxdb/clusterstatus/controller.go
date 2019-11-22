package clusterstatus

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	ControllerTime                       influxdb.Tag   = "time"
	ControllerName                       influxdb.Tag   = "name"
	ControllerNamespace                  influxdb.Tag   = "namespace"
	ControllerClusterName                influxdb.Tag   = "cluster_name"
	ControllerUid                        influxdb.Tag   = "uid"
	ControllerKind                       influxdb.Tag   = "kind"
	ControllerAlamedaSpecScalerName      influxdb.Tag   = "spec_name"
	ControllerAlamedaSpecScalerNamespace influxdb.Tag   = "spec_namespace"
	ControllerReplicas                   influxdb.Field = "replicas"
	ControllerSpecReplicas               influxdb.Field = "spec_replicas"
	ControllerAlamedaSpecScalingTool     influxdb.Field = "spec_scaling_tool"
	ControllerAlamedaSpecPolicy          influxdb.Field = "policy"
	ControllerAlamedaSpecEnableExecution influxdb.Field = "enable_execution"
)

var (
	// ControllerTags is list of tags of alameda_controller_recommendation measurement
	ControllerTags = []influxdb.Tag{
		ControllerTime,
		ControllerName,
		ControllerNamespace,
		ControllerClusterName,
		ControllerUid,
		ControllerKind,
		ControllerAlamedaSpecScalerName,
		ControllerAlamedaSpecScalerNamespace,
	}

	// ControllerFields is list of fields of alameda_controller_recommendation measurement
	ControllerFields = []influxdb.Field{
		ControllerReplicas,
		ControllerSpecReplicas,
		ControllerAlamedaSpecScalingTool,
		ControllerAlamedaSpecPolicy,
		ControllerAlamedaSpecEnableExecution,
	}

	ControllerColumns = []string{
		string(ControllerTime),
		string(ControllerName),
		string(ControllerNamespace),
		string(ControllerClusterName),
		string(ControllerUid),
		string(ControllerKind),
		string(ControllerAlamedaSpecScalerName),
		string(ControllerAlamedaSpecScalerNamespace),
		string(ControllerReplicas),
		string(ControllerSpecReplicas),
		string(ControllerAlamedaSpecScalingTool),
		string(ControllerAlamedaSpecPolicy),
		string(ControllerAlamedaSpecEnableExecution),
	}
)
