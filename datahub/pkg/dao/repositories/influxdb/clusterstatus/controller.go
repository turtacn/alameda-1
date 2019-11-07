package clusterstatus

import (
	"fmt"
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type ControllerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewControllerRepository(influxDBCfg *InternalInflux.Config) *ControllerRepository {
	return &ControllerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *ControllerRepository) CreateControllers(controllers []*DaoClusterTypes.Controller) error {
	points := make([]*InfluxClient.Point, 0)

	for _, controller := range controllers {
		for _, ownerRef := range controller.OwnerReferences {
			// Pack influx tags
			tags := map[string]string{
				string(EntityInfluxCluster.ControllerName):           controller.ObjectMeta.Name,
				string(EntityInfluxCluster.ControllerNamespace):      controller.ObjectMeta.Namespace,
				string(EntityInfluxCluster.ControllerClusterName):    controller.ObjectMeta.ClusterName,
				string(EntityInfluxCluster.ControllerUid):            controller.ObjectMeta.Uid,
				string(EntityInfluxCluster.ControllerOwnerName):      ownerRef.ObjectMeta.Name,
				string(EntityInfluxCluster.ControllerOwnerNamespace): ownerRef.ObjectMeta.Namespace,
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxCluster.ControllerOwnerKind):                  ownerRef.Kind,
				string(EntityInfluxCluster.ControllerKind):                       controller.Kind,
				string(EntityInfluxCluster.ControllerReplicas):                   controller.Replicas,
				string(EntityInfluxCluster.ControllerSpecReplicas):               controller.SpecReplicas,
				string(EntityInfluxCluster.ControllerAlamedaSpecName):            controller.AlamedaControllerSpec.AlamedaScaler.Name,
				string(EntityInfluxCluster.ControllerAlamedaSpecNamespace):       controller.AlamedaControllerSpec.AlamedaScaler.Namespace,
				string(EntityInfluxCluster.ControllerAlamedaSpecPolicy):          controller.AlamedaControllerSpec.Policy,
				string(EntityInfluxCluster.ControllerAlamedaSpecEnableExecution): strconv.FormatBool(controller.AlamedaControllerSpec.EnableExecution),
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(Controller), tags, fields, time.Unix(0, 0))
			if err != nil {
				scope.Error(err.Error())
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			points = append(points, point)
		}
	}

	// Batch write influxdb data points
	err := c.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

func (c *ControllerRepository) ListControllers(request DaoClusterTypes.ListControllersRequest) ([]*DaoClusterTypes.Controller, error) {
	controllers := make([]*DaoClusterTypes.Controller, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Controller,
		GroupByTags:    []string{string(EntityInfluxCluster.ControllerNamespace), string(EntityInfluxCluster.ControllerName), string(EntityInfluxCluster.ControllerClusterName)},
	}

	// Build influx query command
	for _, objectMeta := range request.ObjectMeta {
		keyList := objectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxCluster.ControllerKind))

		valueList := objectMeta.GenerateValueList()
		valueList = append(valueList, request.Kind)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	if len(request.ObjectMeta) == 0 {
		statement.AppendWhereClause("AND", string(EntityInfluxCluster.ControllerKind), "=", request.Kind)
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := c.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Controller, 0), errors.Wrap(err, "failed to list controllers")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			controller := DaoClusterTypes.NewController()
			controller.ObjectMeta.Initialize(group.Tags)
			controller.Populate(group.GetRow(0))
			for j := 0; j < group.GetRowNum(); j++ {
				ownerReference := DaoClusterTypes.OwnerReference{}
				ownerReference.Initialize(group.GetRow(j))
				controller.OwnerReferences = append(controller.OwnerReferences, ownerReference)
			}
			controllers = append(controllers, controller)
		}
	}

	return controllers, nil
}

func (c *ControllerRepository) DeleteControllers(in *ApiResources.DeleteControllersRequest) error {
	whereStr := ""

	for _, objectMeta := range in.GetObjectMeta() {
		namespace := objectMeta.GetNamespace()
		name := objectMeta.GetName()
		whereStr += fmt.Sprintf(" (\"name\"='%s' AND \"namespace\"='%s') OR", name, namespace)
	}

	whereStr = strings.TrimSuffix(whereStr, "OR")

	if whereStr != "" {
		whereStr = "WHERE" + whereStr
	}
	cmd := fmt.Sprintf("DROP SERIES FROM %s %s", string(Controller), whereStr)

	_, err := c.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return err
	}

	return nil
}
