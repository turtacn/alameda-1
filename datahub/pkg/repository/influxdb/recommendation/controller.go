package recommendation

import (
	recommendation_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/recommendation"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
	"time"
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

func (c *ControllerRepository) CreateControllerRecommendations(controllerRecommendations []*datahub_v1alpha1.ControllerRecommendation) error {
	points := []*influxdb_client.Point{}
	for _, conrollerRecommendation := range controllerRecommendations {
		recommendedType := conrollerRecommendation.GetRecommendedType()

		if recommendedType == datahub_v1alpha1.ControllerRecommendation_primitive {
			recommendedSpec := conrollerRecommendation.GetRecommendedSpec()

			tags := map[string]string{
				string(recommendation_entity.ControllerNamespace): recommendedSpec.NamespacedName.Namespace,
				string(recommendation_entity.ControllerName):      recommendedSpec.NamespacedName.Name,
			}

			fields := map[string]interface{}{
				string(recommendation_entity.ControllerCurrentReplicas): recommendedSpec.CurrentReplicas,
				string(recommendation_entity.ControllerDesiredReplicas): recommendedSpec.DesiredReplicas,
				string(recommendation_entity.ControllerCreateTime):      recommendedSpec.CreateTime.Seconds,
			}

			pt, err := influxdb_client.NewPoint(string(Controller), tags, fields, time.Now())
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)
		}
	}

	err := c.influxDB.WritePoints(points, influxdb_client.BatchPointsConfig{
		Database: string(influxdb.Recommendation),
	})

	if err != nil {
		scope.Error(err.Error())
	}

	return nil
}
