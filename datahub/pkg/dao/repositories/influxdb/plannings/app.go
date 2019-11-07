package plannings

import (
	EntityInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/plannings"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	//"github.com/golang/protobuf/ptypes"
	//"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	//"strconv"
	"time"
)

type AppRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewAppRepository(influxDBCfg *InternalInflux.Config) *AppRepository {
	return &AppRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *AppRepository) CreatePlannings(plannings []*ApiPlannings.ApplicationPlanning) error {
	points := make([]*InfluxClient.Point, 0)
	for _, planning := range plannings {
		tags := map[string]string{
			EntityInfluxPlanning.AppNamespace: planning.GetObjectMeta().GetNamespace(),
			EntityInfluxPlanning.AppName:      planning.GetObjectMeta().GetName(),
		}

		fields := map[string]interface{}{
			EntityInfluxPlanning.AppValue: 0,
		}

		pt, err := InfluxClient.NewPoint(string(App), tags, fields, time.Unix(time.Now().UTC().Unix(), 0))
		if err != nil {
			scope.Error(err.Error())
		}

		points = append(points, pt)

	}

	err := c.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Planning),
	})

	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (c *AppRepository) ListPlannings(in *ApiPlannings.ListApplicationPlanningsRequest) ([]*ApiPlannings.ApplicationPlanning, error) {
	influxdbStatement := InternalInflux.Statement{
		Measurement:    App,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	for _, objMeta := range in.GetObjectMeta() {
		namespace := objMeta.GetNamespace()
		name := objMeta.GetName()

		if namespace == "" && name == "" {
			influxdbStatement.WhereClause = ""
			break
		}

		keyList := []string{EntityInfluxPlanning.AppNamespace, EntityInfluxPlanning.AppName}
		valueList := []string{namespace, name}

		tempCondition := influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()

	cmd := influxdbStatement.BuildQueryCmd()

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return make([]*ApiPlannings.ApplicationPlanning, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	plannings := c.getPlanningsFromInfluxRows(influxdbRows)

	return plannings, nil
}

func (c *AppRepository) getPlanningsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiPlannings.ApplicationPlanning {
	plannings := make([]*ApiPlannings.ApplicationPlanning, 0)
	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			tempPlanning := &ApiPlannings.ApplicationPlanning{
				ObjectMeta: &ApiResources.ObjectMeta{
					Namespace: data[string(EntityInfluxPlanning.AppNamespace)],
					Name:      data[string(EntityInfluxPlanning.AppName)],
				},
			}

			plannings = append(plannings, tempPlanning)
		}
	}

	return plannings
}
