package plannings

type clusterTag = string
type clusterField = string

const (
	ClusterPlanningId   clusterTag = "planning_id"
	ClusterPlanningType clusterTag = "planning_type"
	ClusterTime         clusterTag = "time"
	ClusterName         clusterTag = "name"
	ClusterGranularity  clusterTag = "granularity"

	ClusterResourceRequestCPU           clusterField = "resource_request_cpu"
	ClusterResourceRequestMemory        clusterField = "resource_request_memory"
	ClusterResourceLimitCPU             clusterField = "resource_limit_cpu"
	ClusterResourceLimitMemory          clusterField = "resource_limit_memory"
	ClusterInitialResourceRequestCPU    clusterField = "initial_resource_request_cpu"
	ClusterInitialResourceRequestMemory clusterField = "initial_resource_request_memory"
	ClusterInitialResourceLimitCPU      clusterField = "initial_resource_limit_cpu"
	ClusterInitialResourceLimitMemory   clusterField = "initial_resource_limit_memory"
	ClusterStartTime                    clusterField = "start_time"
	ClusterEndTime                      clusterField = "end_time"
	ClusterTotalCost                    clusterField = "total_cost"
	ClusterApplyPlanningNow             clusterField = "apply_planning_now"
)

var (
	ClusterTags = []clusterTag{
		ClusterPlanningId,
		ClusterPlanningType,
		ClusterTime,
		ClusterName,
		ClusterGranularity,
	}

	ClusterFields = []clusterField{
		ClusterResourceRequestCPU,
		ClusterResourceRequestMemory,
		ClusterResourceLimitCPU,
		ClusterResourceLimitMemory,
		ClusterInitialResourceRequestCPU,
		ClusterInitialResourceRequestMemory,
		ClusterInitialResourceLimitCPU,
		ClusterInitialResourceLimitMemory,
		ClusterStartTime,
		ClusterEndTime,
		ClusterTotalCost,
		ClusterApplyPlanningNow,
	}
)
