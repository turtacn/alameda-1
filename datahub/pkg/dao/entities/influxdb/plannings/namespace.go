package plannings

type namespaceTag = string
type namespaceField = string

const (
	NamespacePlanningId   namespaceTag = "planning_id"
	NamespacePlanningType namespaceTag = "planning_type"
	NamespaceTime         namespaceTag = "time"
	NamespaceName         namespaceTag = "name"
	NamespaceGranularity  namespaceTag = "granularity"

	NamespaceResourceRequestCPU           namespaceField = "resource_request_cpu"
	NamespaceResourceRequestMemory        namespaceField = "resource_request_memory"
	NamespaceResourceLimitCPU             namespaceField = "resource_limit_cpu"
	NamespaceResourceLimitMemory          namespaceField = "resource_limit_memory"
	NamespaceInitialResourceRequestCPU    namespaceField = "initial_resource_request_cpu"
	NamespaceInitialResourceRequestMemory namespaceField = "initial_resource_request_memory"
	NamespaceInitialResourceLimitCPU      namespaceField = "initial_resource_limit_cpu"
	NamespaceInitialResourceLimitMemory   namespaceField = "initial_resource_limit_memory"
	NamespaceStartTime                    namespaceField = "start_time"
	NamespaceEndTime                      namespaceField = "end_time"
	NamespaceTotalCost                    namespaceField = "total_cost"
	NamespaceApplyPlanningNow             namespaceField = "apply_planning_now"
)

var (
	NamespaceTags = []namespaceTag{
		NamespacePlanningId,
		NamespacePlanningType,
		NamespaceTime,
		NamespaceName,
		NamespaceGranularity,
	}

	NamespaceFields = []namespaceField{
		NamespaceResourceRequestCPU,
		NamespaceResourceRequestMemory,
		NamespaceResourceLimitCPU,
		NamespaceResourceLimitMemory,
		NamespaceInitialResourceRequestCPU,
		NamespaceInitialResourceRequestMemory,
		NamespaceInitialResourceLimitCPU,
		NamespaceInitialResourceLimitMemory,
		NamespaceStartTime,
		NamespaceEndTime,
		NamespaceTotalCost,
		NamespaceApplyPlanningNow,
	}
)
