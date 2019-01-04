package clusterstatus

type nodeField string
type nodeTag string

const (
	NodeTime nodeTag = "time"

	NodeName      nodeField = "name"
	NodeGroup     nodeField = "group"
	NodeInCluster nodeField = "in_cluster"
)

var (
	NodeTags   = []nodeTag{NodeTime}
	NodeFields = []nodeField{NodeName, NodeGroup, NodeInCluster}
)
