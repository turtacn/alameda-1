package plannings

type clusterTag = string
type clusterField = string

const (
	ClusterTime clusterTag = "time"
	ClusterName clusterTag = "name"

	ClusterValue clusterField = "value"
)

var (
	ClusterTags = []clusterTag{
		ClusterTime,
		ClusterName,
	}

	ClusterFields = []clusterField{
		ClusterValue,
	}
)
