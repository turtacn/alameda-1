package v1alpha1

import (
	DaoCluster "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreateAlamedaNodes add node information to database
func (s *ServiceV1alpha1) CreateNodes(ctx context.Context, in *ApiResources.CreateNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateNodesRequestExtended{CreateNodesRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	nodeDAO := DaoCluster.NewNodeDAO(*s.Config)
	if err := nodeDAO.CreateNodes(requestExtended.ProduceNodes()); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNodes(ctx context.Context, in *ApiResources.ListNodesRequest) (*ApiResources.ListNodesResponse, error) {
	scope.Debug("Request received from ListNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListNodesRequestExtended{ListNodesRequest: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiResources.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	nodeDAO := DaoCluster.NewNodeDAO(*s.Config)
	ns, err := nodeDAO.ListNodes(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListNodes failed: %+v", err)
		return &ApiResources.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	nodes := make([]*ApiResources.Node, 0)
	for _, n := range ns {
		nodeExtended := FormatResponse.NodeExtended{Node: n}
		node := nodeExtended.ProduceNode()
		nodes = append(nodes, node)
	}

	response := ApiResources.ListNodesResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Nodes: nodes,
	}

	return &response, nil
}

// DeleteAlamedaNodes remove node information to database
func (s *ServiceV1alpha1) DeleteNodes(ctx context.Context, in *ApiResources.DeleteNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	/*nodeList := make([]*ApiResources.Node, 0)
	for _, objectMeta := range in.GetObjectMeta() {
		nodeList = append(nodeList, &ApiResources.Node{
			ObjectMeta: &ApiResources.ObjectMeta{
				Name: objectMeta.GetName(),
			},
		})
	}

	nodeDAO := DaoCluster.NewNodeDAO(*s.Config)
	if err := nodeDAO.DeregisterAlamedaNodes(nodeList); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}*/

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

/*
// DeleteAlamedaNodes remove node information to database
func (s *ServiceV1alpha1) DeleteAlamedaNodes(ctx context.Context, in *ApiResources.DeleteAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusInflux.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	alamedaNodeList := []*ApiResources.Node{}
	for _, alamedaNode := range in.GetAlamedaNodes() {
		alamedaNodeList = append(alamedaNodeList, &ApiResources.Node{
			Name: alamedaNode.GetName(),
		})
	}
	if err := nodeDAO.DeregisterAlamedaNodes(alamedaNodeList); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
*/
