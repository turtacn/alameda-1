package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

type ApplicationDAO interface {
	CreateApplications([]*Application) error
	ListApplications(ListApplicationsRequest) ([]*Application, error)
}

type Application struct {
	ObjectMeta             metadata.ObjectMeta
	AlamedaApplicationSpec *AlamedaApplicationSpec
	Controllers            []*Controller
}

type ListApplicationsRequest struct {
	common.QueryCondition
	ObjectMeta []metadata.ObjectMeta
}

type AlamedaApplicationSpec struct {
	ScalingTool string
}

func NewApplication() *Application {
	application := Application{}
	application.Controllers = make([]*Controller, 0)
	return &application
}

func NewListApplicationsRequest() ListApplicationsRequest {
	request := ListApplicationsRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}
