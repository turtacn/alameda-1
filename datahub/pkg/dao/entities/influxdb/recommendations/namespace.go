package recommendations

type namespaceTag = string
type namespaceField = string

const (
	NamespaceTime namespaceTag = "time"
	NamespaceName namespaceTag = "name"

	NamespaceValue namespaceField = "value"
)

var (
	NamespaceTags = []namespaceTag{
		NamespaceTime,
		NamespaceName,
	}

	NamespaceFields = []namespaceField{
		NamespaceValue,
	}
)
