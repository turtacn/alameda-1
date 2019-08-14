package planning

type containerTag = string
type containerField = string

const (
	ContainerPlanningType containerTag = "planning_type"
	ContainerTime         containerTag = "time"
	ContainerNamespace    containerTag = "namespace"
	ContainerName         containerTag = "name"
	ContainerPodName      containerTag = "pod_name"
	ContainerGranularity  containerTag = "granularity"

	ContainerPolicy                       containerField = "policy"
	ContainerPolicyTime                   containerField = "policy_time"
	ContainerResourceRequestCPU           containerField = "resource_request_cpu"
	ContainerResourceRequestMemory        containerField = "resource_request_memory"
	ContainerResourceLimitCPU             containerField = "resource_limit_cpu"
	ContainerResourceLimitMemory          containerField = "resource_limit_memory"
	ContainerInitialResourceRequestCPU    containerField = "initial_resource_request_cpu"
	ContainerInitialResourceRequestMemory containerField = "initial_resource_request_memory"
	ContainerInitialResourceLimitCPU      containerField = "initial_resource_limit_cpu"
	ContainerInitialResourceLimitMemory   containerField = "initial_resource_limit_memory"
	ContainerStartTime                    containerField = "start_time"
	ContainerEndTime                      containerField = "end_time"
	ContainerTopControllerName            containerField = "top_controller_name"
	ContainerTopControllerKind            containerField = "top_controller_kind"
	ContainerPodTotalCost                 containerField = "pod_total_cost"
)

const (
	ContainerMetricKindLimit       = "limit"
	ContainerMetricKindRequest     = "request"
	ContainerMetricKindInitLimit   = "initLimit"
	ContainerMetricKindInitRequest = "initRequest"
)

var (
	ContainerMetricKinds = []string{
		ContainerMetricKindLimit,
		ContainerMetricKindRequest,
		ContainerMetricKindInitLimit,
		ContainerMetricKindInitRequest,
	}
)

var (
	ContainerTags = []containerTag{
		ContainerPlanningType,
		ContainerTime,
		ContainerNamespace,
		ContainerName,
		ContainerPodName,
		ContainerGranularity,
	}

	ContainerFields = []containerField{
		ContainerPolicy,
		ContainerResourceRequestCPU,
		ContainerResourceRequestMemory,
		ContainerResourceLimitCPU,
		ContainerResourceLimitMemory,
		ContainerInitialResourceRequestCPU,
		ContainerInitialResourceRequestMemory,
		ContainerInitialResourceLimitCPU,
		ContainerInitialResourceLimitMemory,
		ContainerStartTime, ContainerEndTime,
		ContainerTopControllerName,
		ContainerTopControllerKind,
	}
)
