package plannings

type nodeTag = string
type nodeField = string

const (
	NodePlanningId   nodeTag = "planning_id"
	NodePlanningType nodeTag = "planning_type"
	NodeTime         nodeTag = "time"
	NodeName         nodeTag = "name"
	NodeGranularity  nodeTag = "granularity"

	NodeResourceRequestCPU           nodeField = "resource_request_cpu"
	NodeResourceRequestMemory        nodeField = "resource_request_memory"
	NodeResourceLimitCPU             nodeField = "resource_limit_cpu"
	NodeResourceLimitMemory          nodeField = "resource_limit_memory"
	NodeInitialResourceRequestCPU    nodeField = "initial_resource_request_cpu"
	NodeInitialResourceRequestMemory nodeField = "initial_resource_request_memory"
	NodeInitialResourceLimitCPU      nodeField = "initial_resource_limit_cpu"
	NodeInitialResourceLimitMemory   nodeField = "initial_resource_limit_memory"
	NodeStartTime                    nodeField = "start_time"
	NodeEndTime                      nodeField = "end_time"
	NodeTotalCost                    nodeField = "total_cost"
	NodeApplyPlanningNow             nodeField = "apply_planning_now"
)

var (
	NodeTags = []nodeTag{
		NodePlanningId,
		NodePlanningType,
		NodeTime,
		NodeName,
		NodeGranularity,
	}

	NodeFields = []nodeField{
		NodeResourceRequestCPU,
		NodeResourceRequestMemory,
		NodeResourceLimitCPU,
		NodeResourceLimitMemory,
		NodeInitialResourceRequestCPU,
		NodeInitialResourceRequestMemory,
		NodeInitialResourceLimitCPU,
		NodeInitialResourceLimitMemory,
		NodeStartTime,
		NodeEndTime,
		NodeTotalCost,
		NodeApplyPlanningNow,
	}
)
