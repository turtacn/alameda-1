package plannings

type namespaceTag = string
type namespaceField = string

const (
	NamespacePlanningType namespaceTag = "planning_type"
	NamespaceTime         namespaceTag = "time"
	NamespaceName         namespaceTag = "name"
	NamespaceType         namespaceTag = "type"

	NamespaceKind              namespaceField = "kind"
	NamespaceCurrentReplicas   namespaceField = "current_replicas"
	NamespaceDesiredReplicas   namespaceField = "desired_replicas"
	NamespaceCreateTime        namespaceField = "create_time"
	NamespaceCurrentCPURequest namespaceField = "current_cpu_requests"
	NamespaceCurrentMEMRequest namespaceField = "current_mem_requests"
	NamespaceCurrentCPULimit   namespaceField = "current_cpu_limits"
	NamespaceCurrentMEMLimit   namespaceField = "current_mem_limits"
	NamespaceDesiredCPULimit   namespaceField = "desired_cpu_limits"
	NamespaceDesiredMEMLimit   namespaceField = "desired_mem_limits"
	NamespaceTotalCost         namespaceField = "total_cost"
)

var (
	NamespaceTags = []namespaceTag{
		NamespacePlanningType,
		NamespaceTime,
		NamespaceName,
		NamespaceType,
	}

	NamespaceFields = []namespaceField{
		NamespaceKind,
		NamespaceCurrentReplicas,
		NamespaceDesiredReplicas,
		NamespaceCreateTime,
		NamespaceCurrentCPURequest,
		NamespaceCurrentMEMRequest,
		NamespaceCurrentCPULimit,
		NamespaceCurrentMEMLimit,
		NamespaceDesiredCPULimit,
		NamespaceDesiredMEMLimit,
		NamespaceTotalCost,
	}
)
