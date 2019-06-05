package recommendation

type controllerTag = string
type controllerField = string

const (
	ControllerTime      controllerTag = "time"
	ControllerNamespace controllerTag = "namespace"
	ControllerName      controllerTag = "name"

	ControllerKind            controllerField = "kind"
	ControllerType            controllerField = "type"
	ControllerCurrentReplicas controllerField = "current_replicas"
	ControllerDesiredReplicas controllerField = "desired_replicas"
	ControllerCreateTime      controllerField = "create_time"
	ControllerCPURequest      controllerField = "cpu_requests"
	ControllerMEMRequest      controllerField = "mem_requests"
	ControllerCPULimit        controllerField = "cpu_limits"
	ControllerMEMLimit        controllerField = "mem_limits"
	ControllerTotalCPULimit   controllerField = "total_cpu_limits"
	ControllerTotalMEMLimit   controllerField = "total_mem_limits"
)

var (
	// ControllerTags is list of tags of alameda_controller_recommendation measurement
	ControllerTags = []controllerTag{
		ControllerTime,
		ControllerNamespace,
		ControllerName,
	}
	// ControllerFields is list of fields of alameda_controller_recommendation measurement
	ControllerField = []controllerField{
		ControllerCurrentReplicas,
		ControllerDesiredReplicas,
		ControllerCreateTime,
		ControllerType,
		ControllerKind,

		ControllerCPURequest,
		ControllerMEMRequest,
		ControllerCPULimit,
		ControllerMEMLimit,
		ControllerTotalCPULimit,
		ControllerTotalMEMLimit,
	}
)
