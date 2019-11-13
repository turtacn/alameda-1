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
	ControllerOwnerName                  influxdb.Tag   = "owner_name"
	ControllerOwnerNamespace             influxdb.Tag   = "owner_namespace"
	ControllerOwnerKind                  influxdb.Field = "owner_kind"
	ControllerKind                       influxdb.Field = "kind"
	ControllerReplicas                   influxdb.Field = "replicas"
	ControllerSpecReplicas               influxdb.Field = "spec_replicas"
	ControllerAlamedaSpecName            influxdb.Field = "spec_name"
	ControllerAlamedaSpecNamespace       influxdb.Field = "spec_namespace"
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
		ControllerOwnerName,
		ControllerOwnerNamespace,
	}

	// ControllerFields is list of fields of alameda_controller_recommendation measurement
	ControllerFields = []influxdb.Field{
		ControllerOwnerKind,
		ControllerKind,
		ControllerReplicas,
		ControllerSpecReplicas,
		ControllerAlamedaSpecName,
		ControllerAlamedaSpecNamespace,
		ControllerAlamedaSpecPolicy,
		ControllerAlamedaSpecEnableExecution,
	}

	ControllerColumns = []string{
		string(ControllerTime),
		string(ControllerName),
		string(ControllerNamespace),
		string(ControllerClusterName),
		string(ControllerUid),
		string(ControllerOwnerName),
		string(ControllerOwnerNamespace),
		string(ControllerOwnerKind),
		string(ControllerKind),
		string(ControllerReplicas),
		string(ControllerSpecReplicas),
		string(ControllerAlamedaSpecName),
		string(ControllerAlamedaSpecNamespace),
		string(ControllerAlamedaSpecPolicy),
		string(ControllerAlamedaSpecEnableExecution),
	}
)
