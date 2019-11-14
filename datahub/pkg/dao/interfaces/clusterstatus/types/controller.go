package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"strconv"
)

type ControllerDAO interface {
	CreateControllers([]*Controller) error
	ListControllers(ListControllersRequest) ([]*Controller, error)
	DeleteControllers(*resources.DeleteControllersRequest) error
}

type Controller struct {
	ObjectMeta            metadata.ObjectMeta
	Kind                  string
	Replicas              int32
	SpecReplicas          int32
	AlamedaControllerSpec AlamedaControllerSpec
}

type ListControllersRequest struct {
	common.QueryCondition
	ObjectMeta []metadata.ObjectMeta
	Kind       string // Valid values: POD, DEPLOYMENT, DEPLOYMENTCONFIG, ALAMEDASCALER, STATEFULSET,
}

type AlamedaControllerSpec struct {
	AlamedaScaler   metadata.ObjectMeta
	ScalingTool     string
	Policy          string
	EnableExecution bool
}

func NewController() *Controller {
	controller := Controller{}
	return &controller
}

func NewListControllersRequest() ListControllersRequest {
	request := ListControllersRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}

func (p *Controller) Populate(values map[string]string) {
	p.Kind = values[string(clusterstatus.ControllerKind)]

	replicas, _ := strconv.ParseInt(values[string(clusterstatus.ControllerReplicas)], 10, 32)
	p.Replicas = int32(replicas)

	specReplicas, _ := strconv.ParseInt(values[string(clusterstatus.ControllerSpecReplicas)], 10, 32)
	p.SpecReplicas = int32(specReplicas)

	p.AlamedaControllerSpec.Initialize(values)
}

func (p *AlamedaControllerSpec) Initialize(values map[string]string) {
	p.AlamedaScaler.Initialize(values)

	if value, ok := values[string(clusterstatus.ControllerAlamedaSpecScalingTool)]; ok {
		p.ScalingTool = value
	}

	if value, ok := values[string(clusterstatus.ControllerAlamedaSpecPolicy)]; ok {
		p.Policy = value
	}

	if value, ok := values[string(clusterstatus.ControllerAlamedaSpecEnableExecution)]; ok {
		enableExecution, _ := strconv.ParseBool(value)
		p.EnableExecution = enableExecution
	}
}
