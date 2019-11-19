package metadata

type name = string

// ContainerName Type alias
type ContainerName = name

// PodName Type alias
type PodName = name

// NamespaceName Type alias
type NamespaceName = name

// NodeName Type alias
type NodeName = name

// NodeName Type alias
type ClusterName = name

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

func (p *ObjectMeta) IsEmpty() bool {
	if p.Name == "" && p.Namespace == "" && p.NodeName == "" && p.ClusterName == "" && p.Uid == "" {
		return true
	}
	return false
}

func (p *ObjectMeta) Initialize(values map[string]string) {
	if value, ok := values["name"]; ok {
		p.Name = value
	}
	if value, ok := values["namespace"]; ok {
		p.Namespace = value
	}
	if value, ok := values["node_name"]; ok {
		p.NodeName = value
	}
	if value, ok := values["cluster_name"]; ok {
		p.ClusterName = value
	}
	if value, ok := values["uid"]; ok {
		p.Uid = value
	}
}

func (p *ObjectMeta) GenerateKeyList() []string {
	keyList := make([]string, 0)
	if p.ClusterName != "" {
		keyList = append(keyList, "cluster_name")
	}
	if p.NodeName != "" {
		keyList = append(keyList, "node_name")
	}
	if p.Namespace != "" {
		keyList = append(keyList, "namespace")
	}
	if p.Name != "" {
		keyList = append(keyList, "name")
	}
	if p.Uid != "" {
		keyList = append(keyList, "uid")
	}
	return keyList
}

func (p *ObjectMeta) GenerateValueList() []string {
	valueList := make([]string, 0)
	if p.ClusterName != "" {
		valueList = append(valueList, p.ClusterName)
	}
	if p.NodeName != "" {
		valueList = append(valueList, p.NodeName)
	}
	if p.Namespace != "" {
		valueList = append(valueList, p.Namespace)
	}
	if p.Name != "" {
		valueList = append(valueList, p.Name)
	}
	if p.Uid != "" {
		valueList = append(valueList, p.Uid)
	}
	return valueList
}
