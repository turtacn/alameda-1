package clusterstatus

import (
	"fmt"
	EntityInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/cluster_status"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"strings"
	"time"
)

type ControllerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewControllerRepository(influxDBCfg *InternalInflux.Config) *ControllerRepository {
	scope.Infof("influxdb-NewControllerRepository input %v", influxDBCfg)
	return &ControllerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *ControllerRepository) CreateControllers(controllers []*datahub_v1alpha1.Controller) error {
	scope.Infof("influxdb-CreateControllers input %d %v", len(controllers), controllers)
	points := make([]*InfluxClient.Point, 0)
	for _, controller := range controllers {
		controllerNamespace := controller.GetControllerInfo().GetNamespacedName().GetNamespace()
		controllerName := controller.GetControllerInfo().GetNamespacedName().GetName()
		controllerKind := controller.GetControllerInfo().GetKind().String()
		controllerExecution := controller.GetEnableRecommendationExecution()
		controllerPolicy := controller.GetPolicy().String()

		ownerNamespace := ""
		ownerName := ""
		ownerKind := ""

		if len(controller.GetOwnerInfo()) > 0 {
			ownerNamespace = controller.GetOwnerInfo()[0].GetNamespacedName().GetNamespace()
			ownerName = controller.GetOwnerInfo()[0].GetNamespacedName().GetName()
			ownerKind = controller.GetOwnerInfo()[0].GetKind().String()
		}

		tags := map[string]string{
			string(EntityInfluxClusterStatus.ControllerNamespace):      controllerNamespace,
			string(EntityInfluxClusterStatus.ControllerName):           controllerName,
			string(EntityInfluxClusterStatus.ControllerOwnerNamespace): ownerNamespace,
			string(EntityInfluxClusterStatus.ControllerOwnerName):      ownerName,
		}

		fields := map[string]interface{}{
			string(EntityInfluxClusterStatus.ControllerKind):            controllerKind,
			string(EntityInfluxClusterStatus.ControllerOwnerKind):       ownerKind,
			string(EntityInfluxClusterStatus.ControllerReplicas):        controller.GetReplicas(),
			string(EntityInfluxClusterStatus.ControllerEnableExecution): strconv.FormatBool(controllerExecution),
			string(EntityInfluxClusterStatus.ControllerPolicy):          controllerPolicy,
			string(EntityInfluxClusterStatus.ControllerSpecReplicas):    controller.GetSpecReplicas(),
		}

		pt, err := InfluxClient.NewPoint(string(Controller), tags, fields, time.Unix(0, 0))
		if err != nil {
			scope.Error(err.Error())
		}
		points = append(points, pt)
	}

	err := c.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})

	if err != nil {
		scope.Error(err.Error())
	}

	return nil
}

func (c *ControllerRepository) ListControllers(in *datahub_v1alpha1.ListControllersRequest) ([]*datahub_v1alpha1.Controller, error) {
	scope.Infof("influxdb-ListControllers input %s", in.String())
	namespace := in.GetNamespacedName().GetNamespace()
	name := in.GetNamespacedName().GetName()

	whereStr := c.convertQueryCondition(namespace, name)

	influxdbStatement := InternalInflux.Statement{
		Measurement: Controller,
		WhereClause: whereStr,
		GroupByTags: []string{EntityInfluxClusterStatus.ControllerNamespace, EntityInfluxClusterStatus.ControllerName},
	}

	cmd := influxdbStatement.BuildQueryCmd()

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		scope.Infof("influxdb-ListControllers cmd %s, error %v", cmd, err)
		return make([]*datahub_v1alpha1.Controller, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)

	controllerList := c.getControllersFromInfluxRows(influxdbRows)
	scope.Infof("influxdb-ListControllers return %d %v", len(controllerList), controllerList)
	return controllerList, nil
}

func (c *ControllerRepository) DeleteControllers(in *datahub_v1alpha1.DeleteControllersRequest) error {
	scope.Infof("influxdb-DeleteControllers input %s %+v", in.String(), in)
	controllers := in.GetControllers()
	whereStr := ""

	for _, controller := range controllers {
		namespace := controller.GetControllerInfo().GetNamespacedName().GetNamespace()
		name := controller.GetControllerInfo().GetNamespacedName().GetName()
		whereStr += fmt.Sprintf(" (\"name\"='%s' AND \"namespace\"='%s') OR", name, namespace)
	}

	whereStr = strings.TrimSuffix(whereStr, "OR")

	if whereStr != "" {
		whereStr = "WHERE" + whereStr
	}
	cmd := fmt.Sprintf("DROP SERIES FROM %s %s", string(Controller), whereStr)

	_, err := c.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		scope.Infof("influxdb-DeleteControllers cmd %s, error %v", cmd, err)
		return err
	}

	return nil
}

func (c *ControllerRepository) convertQueryCondition(namespace string, name string) string {
	ret := ""

	if namespace != "" {
		ret += fmt.Sprintf("\"namespace\"='%s' ", namespace)
	}

	if name != "" {
		ret += fmt.Sprintf("AND \"name\"='%s' ", name)
	}

	ret = strings.TrimPrefix(ret, "AND")
	if ret != "" {
		ret = "WHERE " + ret
	}
	return ret
}

func (c *ControllerRepository) getControllersFromInfluxRows(rows []*InternalInflux.InfluxRow) []*datahub_v1alpha1.Controller {
	controllerList := make([]*datahub_v1alpha1.Controller, 0)
	for _, row := range rows {
		namespace := row.Tags[EntityInfluxClusterStatus.ControllerNamespace]
		name := row.Tags[EntityInfluxClusterStatus.ControllerName]

		tempController := &datahub_v1alpha1.Controller{
			ControllerInfo: &datahub_v1alpha1.ResourceInfo{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: namespace,
					Name:      name,
				},
			},
		}

		ownerInfoList := make([]*datahub_v1alpha1.ResourceInfo, 0)
		for _, data := range row.Data {
			ownerNamespace := data[EntityInfluxClusterStatus.ControllerOwnerNamespace]
			ownerName := data[EntityInfluxClusterStatus.ControllerOwnerName]
			tempOwnerKind := data[EntityInfluxClusterStatus.ControllerOwnerKind]
			var ownerKind datahub_v1alpha1.Kind

			if val, found := datahub_v1alpha1.Kind_value[tempOwnerKind]; found {
				ownerKind = datahub_v1alpha1.Kind(val)
			}

			tempOwner := &datahub_v1alpha1.ResourceInfo{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: ownerNamespace,
					Name:      ownerName,
				},
				Kind: ownerKind,
			}

			ownerInfoList = append(ownerInfoList, tempOwner)

			//------
			tempKind := data[EntityInfluxClusterStatus.ControllerKind]
			var kind datahub_v1alpha1.Kind
			if val, found := datahub_v1alpha1.Kind_value[tempKind]; found {
				kind = datahub_v1alpha1.Kind(val)
			}
			tempController.ControllerInfo.Kind = kind

			tempReplicas, _ := strconv.ParseInt(data[string(EntityInfluxClusterStatus.ControllerReplicas)], 10, 32)
			tempController.Replicas = int32(tempReplicas)

			enableExecution, _ := strconv.ParseBool(data[EntityInfluxClusterStatus.ControllerEnableExecution])
			tempController.EnableRecommendationExecution = enableExecution

			tempPolicy := data[EntityInfluxClusterStatus.ControllerPolicy]
			var policy datahub_v1alpha1.RecommendationPolicy
			if val, found := datahub_v1alpha1.RecommendationPolicy_value[tempPolicy]; found {
				policy = datahub_v1alpha1.RecommendationPolicy(val)
			}
			tempController.Policy = policy
		}

		tempController.OwnerInfo = ownerInfoList
		controllerList = append(controllerList, tempController)
	}

	return controllerList
}
