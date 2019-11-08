package recommendations

type clusterTag = string
type clusterField = string

const (
	ClusterTime clusterTag = "time"
	ClusterName clusterTag = "name"
	ClusterType clusterTag = "type"

	ClusterKind              clusterField = "kind"
	ClusterCurrentReplicas   clusterField = "current_replicas"
	ClusterDesiredReplicas   clusterField = "desired_replicas"
	ClusterCreateTime        clusterField = "create_time"
	ClusterCurrentCPURequest clusterField = "current_cpu_requests"
	ClusterCurrentMEMRequest clusterField = "current_mem_requests"
	ClusterCurrentCPULimit   clusterField = "current_cpu_limits"
	ClusterCurrentMEMLimit   clusterField = "current_mem_limits"
	ClusterDesiredCPULimit   clusterField = "desired_cpu_limits"
	ClusterDesiredMEMLimit   clusterField = "desired_mem_limits"
	ClusterTotalCost         clusterField = "total_cost"
)

var (
	ClusterTags = []clusterTag{
		ClusterTime,
		ClusterName,
		ClusterType,
	}

	ClusterFields = []clusterField{
		ClusterKind,
		ClusterCurrentReplicas,
		ClusterDesiredReplicas,
		ClusterCreateTime,
		ClusterCurrentCPURequest,
		ClusterCurrentMEMRequest,
		ClusterCurrentCPULimit,
		ClusterCurrentMEMLimit,
		ClusterDesiredCPULimit,
		ClusterDesiredMEMLimit,
		ClusterTotalCost,
	}
)
