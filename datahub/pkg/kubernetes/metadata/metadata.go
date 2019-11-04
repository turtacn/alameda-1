package metadata

type name = string

// NamespaceName Type alias
type NamespaceName = name

// PodName Type alias
type PodName = name

// ContainerName Type alias
type ContainerName = name

// NodeName Type alias
type NodeName = name

// NamespacePodName Type alias
type NamespacePodName = name

// NamespacePodContainerName Type alias
type NamespacePodContainerName = name

type ObjectMeta struct {
	Name        string
	Namespace   string
	NodeName    string
	ClusterName string
	Uid         string
}
