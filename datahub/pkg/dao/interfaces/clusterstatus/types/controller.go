package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
)

type ControllerDAO interface {
	CreateControllers([]*Controller) error
	ListControllers(*ListControllersRequest) ([]*Controller, error)
	DeleteControllers(*DeleteControllersRequest) error
}

type Controller struct {
	ObjectMeta            *metadata.ObjectMeta
	Kind                  string
	Replicas              int32
	SpecReplicas          int32
	AlamedaControllerSpec *AlamedaControllerSpec
}

type ListControllersRequest struct {
	common.QueryCondition
	ControllerObjectMeta []*ControllerObjectMeta
}

type DeleteControllersRequest struct {
	ControllerObjectMeta []*ControllerObjectMeta
}

type ControllerObjectMeta struct {
	ObjectMeta    *metadata.ObjectMeta
	AlamedaScaler *metadata.ObjectMeta
	Kind          string // Valid values: DEPLOYMENT, DEPLOYMENTCONFIG, STATEFULSET
	ScalingTool   string // Valid values: NONE, VPA, HPA
}

type AlamedaControllerSpec struct {
	AlamedaScaler   *metadata.ObjectMeta
	ScalingTool     string
	Policy          string
	EnableExecution bool
}

func NewController(entity *clusterstatus.ControllerEntity) *Controller {
	controller := Controller{}
	controller.ObjectMeta = &metadata.ObjectMeta{}
	controller.ObjectMeta.Name = entity.Name
	controller.ObjectMeta.Namespace = entity.Namespace
	controller.ObjectMeta.ClusterName = entity.ClusterName
	controller.ObjectMeta.Uid = entity.Uid
	controller.Kind = entity.Kind
	controller.Replicas = entity.Replicas
	controller.SpecReplicas = entity.SpecReplicas
	controller.AlamedaControllerSpec = NewAlamedaControllerSpec(entity)
	return &controller
}

func NewListControllersRequest() *ListControllersRequest {
	request := ListControllersRequest{}
	request.ControllerObjectMeta = make([]*ControllerObjectMeta, 0)
	return &request
}

func NewDeleteControllersRequest() *DeleteControllersRequest {
	request := DeleteControllersRequest{}
	request.ControllerObjectMeta = make([]*ControllerObjectMeta, 0)
	return &request
}

func NewControllerObjectMeta(objectMeta, alamedaScaler *metadata.ObjectMeta, kind, scalingTool string) *ControllerObjectMeta {
	controllerObjectMeta := ControllerObjectMeta{}
	controllerObjectMeta.ObjectMeta = objectMeta
	controllerObjectMeta.AlamedaScaler = alamedaScaler
	controllerObjectMeta.Kind = kind
	controllerObjectMeta.ScalingTool = scalingTool
	return &controllerObjectMeta
}

func NewAlamedaControllerSpec(entity *clusterstatus.ControllerEntity) *AlamedaControllerSpec {
	spec := AlamedaControllerSpec{}
	spec.AlamedaScaler = &metadata.ObjectMeta{}
	spec.AlamedaScaler.Name = entity.AlamedaSpecScalerName
	spec.ScalingTool = entity.AlamedaSpecScalingTool
	spec.Policy = entity.AlamedaSpecPolicy
	enableExecution, _ := strconv.ParseBool(entity.AlamedaSpecEnableExecution)
	spec.EnableExecution = enableExecution
	return &spec
}

func (p *Controller) BuildEntity() *clusterstatus.ControllerEntity {
	entity := clusterstatus.ControllerEntity{
		// InfluxDB tags
		Time:                   influxdb.ZeroTime,
		Name:                   p.ObjectMeta.Name,
		Namespace:              p.ObjectMeta.Namespace,
		ClusterName:            p.ObjectMeta.ClusterName,
		Uid:                    p.ObjectMeta.Uid,
		Kind:                   p.Kind,
		AlamedaSpecScalerName:  p.AlamedaControllerSpec.AlamedaScaler.Name,
		AlamedaSpecScalingTool: p.AlamedaControllerSpec.ScalingTool,

		// InfluxDB fields
		Replicas:                   p.Replicas,
		SpecReplicas:               p.SpecReplicas,
		AlamedaSpecPolicy:          p.AlamedaControllerSpec.Policy,
		AlamedaSpecEnableExecution: strconv.FormatBool(p.AlamedaControllerSpec.EnableExecution),
	}

	return &entity
}
