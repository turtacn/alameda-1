package recommendations

type appTag = string
type appField = string

const (
	AppTime      appTag = "time"
	AppNamespace appTag = "namespace"
	AppName      appTag = "name"
	AppType      appTag = "type"

	AppKind              appField = "kind"
	AppCurrentReplicas   appField = "current_replicas"
	AppDesiredReplicas   appField = "desired_replicas"
	AppCreateTime        appField = "create_time"
	AppCurrentCPURequest appField = "current_cpu_requests"
	AppCurrentMEMRequest appField = "current_mem_requests"
	AppCurrentCPULimit   appField = "current_cpu_limits"
	AppCurrentMEMLimit   appField = "current_mem_limits"
	AppDesiredCPULimit   appField = "desired_cpu_limits"
	AppDesiredMEMLimit   appField = "desired_mem_limits"
	AppTotalCost         appField = "total_cost"
)

var (
	AppTags = []appTag{
		AppTime,
		AppNamespace,
		AppName,
		AppType,
	}

	AppFields = []appField{
		AppKind,
		AppCurrentReplicas,
		AppDesiredReplicas,
		AppCreateTime,
		AppCurrentCPURequest,
		AppCurrentMEMRequest,
		AppCurrentCPULimit,
		AppCurrentMEMLimit,
		AppDesiredCPULimit,
		AppDesiredMEMLimit,
		AppTotalCost,
	}
)
