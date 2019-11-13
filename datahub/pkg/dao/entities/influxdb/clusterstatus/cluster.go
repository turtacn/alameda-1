package clusterstatus

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ClusterTime  influxdb.Tag   = "time"
	ClusterName  influxdb.Tag   = "name"
	ClusterUid   influxdb.Tag   = "uid"
	ClusterValue influxdb.Field = "value"
)

var (
	ClusterTags = []influxdb.Tag{
		ClusterName,
		ClusterUid,
	}

	ClusterFields = []influxdb.Field{
		ClusterValue,
	}

	ClusterColumns = []string{
		string(ClusterName),
		string(ClusterUid),
		string(ClusterValue),
	}
)

type ClusterEntity struct {
	Time time.Time
	Name *string
	Uid  *string
}
