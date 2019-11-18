package plannings

type controllerTag = string
type controllerField = string

const (
	ControllerPlanningId   controllerTag = "planning_id"
	ControllerPlanningType controllerTag = "planning_type"
	ControllerTime         controllerTag = "time"
	ControllerNamespace    controllerTag = "namespace"
	ControllerName         controllerTag = "name"
	ControllerGranularity  controllerTag = "granularity"
	ControllerKind         controllerTag = "kind"

	ControllerResourceRequestCPU           controllerField = "resource_request_cpu"
	ControllerResourceRequestMemory        controllerField = "resource_request_memory"
	ControllerResourceLimitCPU             controllerField = "resource_limit_cpu"
	ControllerResourceLimitMemory          controllerField = "resource_limit_memory"
	ControllerInitialResourceRequestCPU    controllerField = "initial_resource_request_cpu"
	ControllerInitialResourceRequestMemory controllerField = "initial_resource_request_memory"
	ControllerInitialResourceLimitCPU      controllerField = "initial_resource_limit_cpu"
	ControllerInitialResourceLimitMemory   controllerField = "initial_resource_limit_memory"
	ControllerStartTime                    controllerField = "start_time"
	ControllerEndTime                      controllerField = "end_time"
	ControllerTotalCost                    controllerField = "total_cost"
	ControllerApplyPlanningNow             controllerField = "apply_planning_now"
)

var (
	// ControllerTags is list of tags of alameda_controller_recommendation measurement
	ControllerTags = []controllerTag{
		ControllerPlanningId,
		ControllerPlanningType,
		ControllerTime,
		ControllerNamespace,
		ControllerName,
		ControllerGranularity,
		ControllerKind,
	}
	// ControllerFields is list of fields of alameda_controller_recommendation measurement
	ControllerField = []controllerField{
		ControllerResourceRequestCPU,
		ControllerResourceRequestMemory,
		ControllerResourceLimitCPU,
		ControllerResourceLimitMemory,
		ControllerInitialResourceRequestCPU,
		ControllerInitialResourceRequestMemory,
		ControllerInitialResourceLimitCPU,
		ControllerInitialResourceLimitMemory,
		ControllerStartTime,
		ControllerEndTime,
		ControllerTotalCost,
		ControllerApplyPlanningNow,
	}
)
