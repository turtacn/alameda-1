package clusterstatus

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	NamespaceTime        influxdb.Tag   = "time"
	NamespaceName        influxdb.Tag   = "name"
	NamespaceClusterName influxdb.Tag   = "cluster_name"
	NamespaceUid         influxdb.Tag   = "uid"
	NamespaceValue       influxdb.Field = "value"
)

var (
	NamespaceTags = []influxdb.Tag{
		NamespaceName,
		NamespaceClusterName,
		NamespaceUid,
	}

	NamespaceFields = []influxdb.Field{
		NamespaceValue,
	}

	NamespaceColumns = []string{
		string(NamespaceName),
		string(NamespaceClusterName),
		string(NamespaceUid),
		string(NamespaceValue),
	}
)

type NamespaceEntity struct {
	Time        time.Time
	Name        *string
	Namespace   *string
	NodeName    *string
	ClusterName *string
}
