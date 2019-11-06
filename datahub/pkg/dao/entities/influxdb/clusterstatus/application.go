package clusterstatus

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ApplicationTime        influxdb.Tag   = "time"
	ApplicationName        influxdb.Tag   = "name"
	ApplicationNamespace   influxdb.Tag   = "namespace"
	ApplicationNodeName    influxdb.Tag   = "node_name"
	ApplicationClusterName influxdb.Tag   = "cluster_name"
	ApplicationUid         influxdb.Tag   = "uid"
	ApplicationValue       influxdb.Field = "value"
)

var (
	ApplicationTags = []influxdb.Tag{
		ApplicationName,
		ApplicationNamespace,
		ApplicationNodeName,
		ApplicationClusterName,
		ApplicationUid,
	}

	ApplicationFields = []influxdb.Field{
		ApplicationValue,
	}

	ApplicationColumns = []string{
		string(ApplicationName),
		string(ApplicationNamespace),
		string(ApplicationNodeName),
		string(ApplicationClusterName),
		string(ApplicationUid),
		string(ApplicationValue),
	}
)

type ApplicationEntity struct {
	Time        time.Time
	Name        *string
	Namespace   *string
	NodeName    *string
	ClusterName *string
}
