package recommendations

type nodeTag = string
type nodeField = string

const (
	NodeTime        nodeTag = "time"
	NodeClusterName nodeTag = "cluster_name"
	NodeName        nodeTag = "name"
	NodeType        nodeTag = "type"

	NodeKind              nodeField = "kind"
	NodeCurrentReplicas   nodeField = "current_replicas"
	NodeDesiredReplicas   nodeField = "desired_replicas"
	NodeCreateTime        nodeField = "create_time"
	NodeCurrentCPURequest nodeField = "current_cpu_requests"
	NodeCurrentMEMRequest nodeField = "current_mem_requests"
	NodeCurrentCPULimit   nodeField = "current_cpu_limits"
	NodeCurrentMEMLimit   nodeField = "current_mem_limits"
	NodeDesiredCPULimit   nodeField = "desired_cpu_limits"
	NodeDesiredMEMLimit   nodeField = "desired_mem_limits"
	NodeTotalCost         nodeField = "total_cost"
)

var (
	NodeTags = []nodeTag{
		NodeTime,
		NodeClusterName,
		NodeName,
		NodeType,
	}

	NodeFields = []nodeField{
		NodeKind,
		NodeCurrentReplicas,
		NodeDesiredReplicas,
		NodeCreateTime,
		NodeCurrentCPURequest,
		NodeCurrentMEMRequest,
		NodeCurrentCPULimit,
		NodeCurrentMEMLimit,
		NodeDesiredCPULimit,
		NodeDesiredMEMLimit,
		NodeTotalCost,
	}
)
