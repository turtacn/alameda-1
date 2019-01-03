package clusterstatus

// Node provides node measurement operations
type Node interface {
	RegisterAlamedaNodes([]*AlamedaNode) error
	DeregisterAlamedaNodes([]*AlamedaNode) error
	ListAlamedaNodes() ([]*AlamedaNode, error)
}

// AlamedaNode is predicted node in cluster
type AlamedaNode struct {
	Name string
}
