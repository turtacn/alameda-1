package plannings

type appTag = string
type appField = string

const (
	AppPlanningId   appTag = "planning_id"
	AppPlanningType appTag = "planning_type"
	AppTime         appTag = "time"
	AppNamespace    appTag = "namespace"
	AppName         appTag = "name"
	AppGranularity  appTag = "granularity"

	AppResourceRequestCPU           appField = "resource_request_cpu"
	AppResourceRequestMemory        appField = "resource_request_memory"
	AppResourceLimitCPU             appField = "resource_limit_cpu"
	AppResourceLimitMemory          appField = "resource_limit_memory"
	AppInitialResourceRequestCPU    appField = "initial_resource_request_cpu"
	AppInitialResourceRequestMemory appField = "initial_resource_request_memory"
	AppInitialResourceLimitCPU      appField = "initial_resource_limit_cpu"
	AppInitialResourceLimitMemory   appField = "initial_resource_limit_memory"
	AppStartTime                    appField = "start_time"
	AppEndTime                      appField = "end_time"
	AppTotalCost                    appField = "total_cost"
	AppApplyPlanningNow             appField = "apply_planning_now"
)

const (
	AppMetricKindLimit       = "limit"
	AppMetricKindRequest     = "request"
	AppMetricKindInitLimit   = "initLimit"
	AppMetricKindInitRequest = "initRequest"
)

var (
	AppMetricKinds = []string{
		AppMetricKindLimit,
		AppMetricKindRequest,
		AppMetricKindInitLimit,
		AppMetricKindInitRequest,
	}
)

var (
	AppTags = []appTag{
		AppPlanningId,
		AppPlanningType,
		AppTime,
		AppNamespace,
		AppName,
		AppGranularity,
	}

	AppFields = []appField{
		AppResourceRequestCPU,
		AppResourceRequestMemory,
		AppResourceLimitCPU,
		AppResourceLimitMemory,
		AppInitialResourceRequestCPU,
		AppInitialResourceRequestMemory,
		AppInitialResourceLimitCPU,
		AppInitialResourceLimitMemory,
		AppStartTime,
		AppEndTime,
		AppTotalCost,
		AppApplyPlanningNow,
	}
)
