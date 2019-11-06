package recommendations

import (
	EntityInfluxRecommend "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/recommendations"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
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

func (c *AppRepository) CreateRecommendations(recommendations []*ApiRecommendations.ApplicationRecommendation) error {
	points := make([]*InfluxClient.Point, 0)
	for _, recommendation := range recommendations {
		tags := map[string]string{
			EntityInfluxRecommend.AppNamespace: recommendation.GetObjectMeta().GetNamespace(),
			EntityInfluxRecommend.AppName:      recommendation.GetObjectMeta().GetName(),
		}

		fields := map[string]interface{}{
			EntityInfluxRecommend.AppValue: 0,
		}

		pt, err := InfluxClient.NewPoint(string(App), tags, fields, time.Unix(time.Now().UTC().Unix(), 0))
		if err != nil {
			scope.Error(err.Error())
		}

		points = append(points, pt)

	}

	err := c.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Recommendation),
	})

	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (c *AppRepository) ListRecommendations(in *ApiRecommendations.ListApplicationRecommendationsRequest) ([]*ApiRecommendations.ApplicationRecommendation, error) {
	namespace := in.GetObjectMeta()[0].GetNamespace()
	name := in.GetObjectMeta()[0].GetName()

	influxdbStatement := InternalInflux.Statement{
		Measurement:    App,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	influxdbStatement.AppendWhereClause("AND", EntityInfluxRecommend.AppNamespace, "=", namespace)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxRecommend.AppName, "=", name)
	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()

	cmd := influxdbStatement.BuildQueryCmd()

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Recommendation))
	if err != nil {
		return make([]*ApiRecommendations.ApplicationRecommendation, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	recommendations := c.getRecommendationsFromInfluxRows(influxdbRows)

	return recommendations, nil
}

func (c *AppRepository) getRecommendationsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiRecommendations.ApplicationRecommendation {
	recommendations := make([]*ApiRecommendations.ApplicationRecommendation, 0)
	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			tempRecommendation := &ApiRecommendations.ApplicationRecommendation{
				ObjectMeta: &ApiResources.ObjectMeta{
					Namespace: data[string(EntityInfluxRecommend.AppNamespace)],
					Name:      data[string(EntityInfluxRecommend.AppName)],
				},
			}

			recommendations = append(recommendations, tempRecommendation)
		}
	}

	return recommendations
}
