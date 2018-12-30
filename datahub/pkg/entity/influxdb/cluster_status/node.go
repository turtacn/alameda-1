package clusterstatus

type NodeField string
type NodeTag string

const (
	NodeTime NodeTag = "time"

	NodeName      NodeField = "name"
	NodeGroup     NodeField = "group"
	NodeInCluster NodeField = "in_cluster"
)

var (
	NodeTags   = []NodeTag{NodeTime}
	NodeFields = []NodeField{NodeName, NodeGroup, NodeInCluster}
)
