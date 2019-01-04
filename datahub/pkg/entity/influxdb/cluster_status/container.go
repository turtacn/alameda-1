package clusterstatus

type containerTag string
type containerField string

const (
	// ContainerTime is the time that container information is saved to the measurement
	ContainerTime containerTag = "time"
	// ContainerNamespace is the container namespace
	ContainerNamespace containerTag = "namespace"
	// ContainerPodName is the name of pod that container is running in
	ContainerPodName containerTag = "pod_name"
	// ContainerAlamedaScalerNamespace is the namespace of AlamedaScaler that container belongs to
	ContainerAlamedaScalerNamespace containerTag = "alameda_scaler_namespace"
	// ContainerAlamedaScalerName is the name of AlamedaScaler that container belongs to
	ContainerAlamedaScalerName containerTag = "alameda_scaler_name"
	// ContainerNodeName is the name of node that container is running in
	ContainerNodeName containerTag = "node_name"

	// ContainerName is the container name
	ContainerName containerField = "name"
	// ContainerResourceRequestCPU is CPU request of the container
	ContainerResourceRequestCPU containerField = "resource_request_cpu"
	// ContainerResourceRequestMemory is memory request of the container
	ContainerResourceRequestMemory containerField = "resource_request_memroy"
	// ContainerResourceLimitCPU is CPU limit of the container
	ContainerResourceLimitCPU containerField = "resource_limit_cpu"
	// ContainerResourceLimitMemory is memory limit of the container
	ContainerResourceLimitMemory containerField = "resource_limit_memory"
	// ContainerIsAlameda is the state that container is predicted or not
	ContainerIsAlameda containerField = "is_alameda"
	// ContainerIsDeleted is the state that container is deleted or not
	ContainerIsDeleted containerField = "is_deleted"
	// ContainerPolicy is the prediction policy of container
	ContainerPolicy containerField = "policy"
)

var (
	// ContainerTags is the list of container measurement tags
	ContainerTags = []containerTag{
		ContainerTime, ContainerNamespace, ContainerPodName,
		ContainerAlamedaScalerNamespace, ContainerAlamedaScalerName,
		ContainerNodeName,
	}
	// ContainerFields is the list of container measurement fields
	ContainerFields = []containerField{
		ContainerName, ContainerResourceRequestCPU, ContainerResourceRequestMemory,
		ContainerResourceLimitCPU, ContainerResourceLimitMemory,
		ContainerIsAlameda, ContainerIsDeleted, ContainerPolicy,
	}
)
