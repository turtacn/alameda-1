package clusterstatus

type Node interface {
	RegisterAlamedaNodes([]*AlamedaNode) error
	DeregisterAlamedaNodes([]*AlamedaNode) error
	ListAlamedaNodes() ([]*AlamedaNode, error)
}

type AlamedaNode struct {
	Name string
}
