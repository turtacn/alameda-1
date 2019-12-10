package clusterstatus

import (
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	Utils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
)

type ControllerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewControllerRepository(influxDBCfg InternalInflux.Config) *ControllerRepository {
	return &ControllerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (p *ControllerRepository) CreateControllers(controllers []*DaoClusterTypes.Controller) error {
	points := make([]*InfluxClient.Point, 0)

	for _, controller := range controllers {
		entity := controller.BuildEntity()

		// Add to influx point list
		pt, err := entity.BuildInfluxPoint(string(Controller))
		if err != nil {
			scope.Error(err.Error())
			return errors.Wrap(err, "failed to instance influxdb data point")
		}
		points = append(points, pt)
	}

	// Batch write influxdb data points
	err := p.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

func (p *ControllerRepository) ListControllers(request *DaoClusterTypes.ListControllersRequest) ([]*DaoClusterTypes.Controller, error) {
	controllers := make([]*DaoClusterTypes.Controller, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Controller,
		GroupByTags:    []string{string(EntityInfluxCluster.ControllerNamespace), string(EntityInfluxCluster.ControllerClusterName)},
	}

	// Build influx query command
	for _, controllerObjectMeta := range request.ControllerObjectMeta {
		keyList := make([]string, 0)
		valueList := make([]string, 0)

		if controllerObjectMeta.ObjectMeta != nil {
			keyList = controllerObjectMeta.ObjectMeta.GenerateKeyList()
			valueList = controllerObjectMeta.ObjectMeta.GenerateValueList()
		}

		if controllerObjectMeta.AlamedaScaler != nil {
			keyList = append(keyList, string(EntityInfluxCluster.ControllerAlamedaSpecScalerName))
			valueList = append(valueList, controllerObjectMeta.AlamedaScaler.Name)

			if !Utils.SliceContains(keyList, string(EntityInfluxCluster.ControllerNamespace)) {
				keyList = append(keyList, string(EntityInfluxCluster.ControllerNamespace))
				valueList = append(valueList, controllerObjectMeta.AlamedaScaler.Namespace)
			}

			if !Utils.SliceContains(keyList, string(EntityInfluxCluster.ControllerClusterName)) {
				keyList = append(keyList, string(EntityInfluxCluster.ControllerClusterName))
				valueList = append(valueList, controllerObjectMeta.AlamedaScaler.ClusterName)
			}
		}

		if controllerObjectMeta.Kind != "" && controllerObjectMeta.Kind != ApiResources.Kind_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ControllerKind))
			valueList = append(valueList, controllerObjectMeta.Kind)
		}

		if controllerObjectMeta.ScalingTool != "" && controllerObjectMeta.ScalingTool != ApiResources.ScalingTool_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ControllerAlamedaSpecScalerScalingTool))
			valueList = append(valueList, controllerObjectMeta.ScalingTool)
		}

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Controller, 0), errors.Wrap(err, "failed to list controllers")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				controller := DaoClusterTypes.NewController(EntityInfluxCluster.NewControllerEntity(row))
				controllers = append(controllers, controller)
			}
		}
	}

	return controllers, nil
}

func (p *ControllerRepository) DeleteControllers(request *DaoClusterTypes.DeleteControllersRequest) error {
	statement := InternalInflux.Statement{
		Measurement: Controller,
	}

	if !p.influxDB.MeasurementExist(string(RepoInflux.ClusterStatus), string(Controller)) {
		return nil
	}

	// Build influx drop command
	for _, controllerObjectMeta := range request.ControllerObjectMeta {
		keyList := make([]string, 0)
		valueList := make([]string, 0)

		if controllerObjectMeta.ObjectMeta != nil {
			keyList = controllerObjectMeta.ObjectMeta.GenerateKeyList()
			valueList = controllerObjectMeta.ObjectMeta.GenerateValueList()
		}

		if controllerObjectMeta.AlamedaScaler != nil {
			keyList = append(keyList, string(EntityInfluxCluster.ControllerAlamedaSpecScalerName))
			valueList = append(valueList, controllerObjectMeta.AlamedaScaler.Name)

			if !Utils.SliceContains(keyList, string(EntityInfluxCluster.ControllerNamespace)) {
				keyList = append(keyList, string(EntityInfluxCluster.ControllerNamespace))
				valueList = append(valueList, controllerObjectMeta.AlamedaScaler.Namespace)
			}

			if !Utils.SliceContains(keyList, string(EntityInfluxCluster.ControllerClusterName)) {
				keyList = append(keyList, string(EntityInfluxCluster.ControllerClusterName))
				valueList = append(valueList, controllerObjectMeta.AlamedaScaler.ClusterName)
			}
		}

		if controllerObjectMeta.Kind != "" && controllerObjectMeta.Kind != ApiResources.Kind_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ControllerKind))
			valueList = append(valueList, controllerObjectMeta.Kind)
		}

		if controllerObjectMeta.ScalingTool != "" && controllerObjectMeta.ScalingTool != ApiResources.ScalingTool_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ControllerAlamedaSpecScalerScalingTool))
			valueList = append(valueList, controllerObjectMeta.ScalingTool)
		}

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	cmd := statement.BuildDropCmd()

	_, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to delete controllers")
	}

	return nil
}
