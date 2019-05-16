package clusterstatus

import (
	"fmt"
	"strconv"
	"time"

	controller_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/cluster_status"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	datahub_api "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
)

type ControllerRepository struct {
	influxDB *influxdb.InfluxDBRepository
}

func NewControllerRepository(influxDBCfg *influxdb.Config) *ControllerRepository {
	return &ControllerRepository{
		influxDB: &influxdb.InfluxDBRepository{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *ControllerRepository) CreateControllers(controllers []*datahub_api.Controller) error {
	points := make([]*influxdb_client.Point, 0)
	for _, controller := range controllers {
		controllerNamespace := controller.GetControllerInfo().GetNamespacedName().GetNamespace()
		controllerName := controller.GetControllerInfo().GetNamespacedName().GetName()
		controllerKind := controller.GetControllerInfo().GetKind()

		ownerNamespace := ""
		ownerName := ""
		var ownerKind datahub_api.Kind

		if len(controller.GetOwnerInfo()) > 0 {
			ownerNamespace = controller.GetOwnerInfo()[0].GetNamespacedName().GetNamespace()
			ownerName = controller.GetOwnerInfo()[0].GetNamespacedName().GetName()
			ownerKind = controller.GetOwnerInfo()[0].GetKind()
		}

		tags := map[string]string{
			string(controller_entity.ControllerNamespace):      controllerNamespace,
			string(controller_entity.ControllerName):           controllerName,
			string(controller_entity.ControllerOwnerNamespace): ownerNamespace,
			string(controller_entity.ControllerOwnerName):      ownerName,
		}

		fields := map[string]interface{}{
			string(controller_entity.ControllerKind):      strconv.Itoa(int(controllerKind)),
			string(controller_entity.ControllerOwnerKind): strconv.Itoa(int(ownerKind)),
			string(controller_entity.ControllerReplicas):  controller.GetReplicas(),
		}

		pt, err := influxdb_client.NewPoint(string(Controller), tags, fields, time.Unix(0, 0))
		if err != nil {
			scope.Error(err.Error())
		}
		points = append(points, pt)
	}

	err := c.influxDB.WritePoints(points, influxdb_client.BatchPointsConfig{
		Database: string(influxdb.ClusterStatus),
	})

	if err != nil {
		scope.Error(err.Error())
	}

	return nil
}

func (c *ControllerRepository) ListControllers(in *datahub_api.ListControllersRequest) ([]*datahub_api.Controller, error) {
	controllers := make([]*datahub_api.Controller, 0)

	namespace := in.GetNamespacedName().GetNamespace()
	name := in.GetNamespacedName().GetName()

	//conditionStr := c.convertQueryCondition(queryCondition)
	whereStr := fmt.Sprintf("WHERE \"name\"='%s' AND \"namespace\"='%s'", name, namespace)

	cmd := fmt.Sprintf("SELECT * FROM %s %s", string(Controller), whereStr)

	results, err := c.influxDB.QueryDB(cmd, string(influxdb.ClusterStatus))
	if err != nil {
		return controllers, err
	}

	kind := datahub_api.Kind_POD
	replicas := 0
	ownerinfoList := make([]*datahub_api.ResourceInfo, 0)

	influxdbRows := influxdb.PackMap(results)
	for _, influxdbRow := range influxdbRows {
		for _, data := range influxdbRow.Data {
			tempOwner := c.NewOwnerInfoFromMap(data)
			ownerinfoList = append(ownerinfoList, tempOwner)

			tempKind, _ := strconv.ParseInt(data[string(controller_entity.ControllerKind)], 10, 32)
			kind = datahub_api.Kind(tempKind)

			tempReplicas, _ := strconv.ParseInt(data[string(controller_entity.ControllerReplicas)], 10, 32)
			replicas = int(tempReplicas)
		}
	}

	if len(ownerinfoList) > 0 {
		tempController := datahub_api.Controller{
			ControllerInfo: &datahub_api.ResourceInfo{
				NamespacedName: &datahub_api.NamespacedName{
					Namespace: namespace,
					Name:      name,
				},
				Kind: kind,
			},
			OwnerInfo: ownerinfoList,
			Replicas:  int32(replicas),
		}

		controllers = append(controllers, &tempController)
	}

	return controllers, nil
}

// NewEntityFromMap Build entity from map
func (c *ControllerRepository) NewOwnerInfoFromMap(data map[string]string) *datahub_api.ResourceInfo {

	ownerNamespace := data[string(controller_entity.ControllerOwnerNamespace)]
	ownerName := data[string(controller_entity.ControllerOwnerName)]

	tempKind, _ := strconv.ParseInt(data[string(controller_entity.ControllerOwnerKind)], 10, 32)
	ownerKind := datahub_api.Kind(tempKind)

	tempOwner := datahub_api.ResourceInfo{
		NamespacedName: &datahub_api.NamespacedName{
			Namespace: ownerNamespace,
			Name:      ownerName,
		},
		Kind: ownerKind,
	}

	return &tempOwner
}
