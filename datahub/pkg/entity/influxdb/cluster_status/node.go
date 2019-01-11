package clusterstatus

type nodeField string
type nodeTag string

const (
	// NodeTime is the time node information is inserted to databse
	NodeTime nodeTag = "time"
	// NodeName is the name of node
	NodeName nodeTag = "name"

	// NodeGroup is node group name
	NodeGroup nodeField = "group"
	// NodeInCluster is the state node is in cluster or not
	NodeInCluster nodeField = "in_cluster"
)

var (
	// NodeTags list tags of node measurement
	NodeTags = []nodeTag{NodeTime, NodeName}
	// NodeFields list fields of node measurement
	NodeFields = []nodeField{NodeGroup, NodeInCluster}
)
