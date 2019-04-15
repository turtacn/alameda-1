package recommendation

import (
	"fmt"
	recommendation_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/recommendation"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
	"strconv"
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

		if recommendedType == datahub_v1alpha1.ControllerRecommendedType_CRT_Primitive {
			recommendedSpec := conrollerRecommendation.GetRecommendedSpec()

			tags := map[string]string{
				string(recommendation_entity.ControllerNamespace): recommendedSpec.NamespacedName.Namespace,
				string(recommendation_entity.ControllerName):      recommendedSpec.NamespacedName.Name,
			}

			fields := map[string]interface{}{
				string(recommendation_entity.ControllerType):            strconv.Itoa(int(datahub_v1alpha1.ControllerRecommendedType_CRT_Primitive)),
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

// ListContainerRecommendations list container recommendations
func (c *ControllerRepository) ListControllerRecommendations(controllerNamespacedName *datahub_v1alpha1.NamespacedName,
	queryCondition *datahub_v1alpha1.QueryCondition) ([]*datahub_v1alpha1.ControllerRecommendation, error) {

	response := make([]*datahub_v1alpha1.ControllerRecommendation, 0)

	namespace := controllerNamespacedName.GetNamespace()
	name := controllerNamespacedName.GetName()

	conditionStr := c.convertQueryCondition(queryCondition)
	whereStr := fmt.Sprintf("WHERE \"name\"='%s' AND \"namespace\"='%s' %s", name, namespace, conditionStr)

	cmd := fmt.Sprintf("SELECT * FROM %s %s", string(Controller), whereStr)

	results, err := c.influxDB.QueryDB(cmd, string(influxdb.Recommendation))
	if err != nil {
		return response, err
	}

	influxdbRows := influxdb.PackMap(results)
	for _, influxdbRow := range influxdbRows {
		for _, data := range influxdbRow.Data {
			tempRecommendation := c.NewRecommendationFromMap(data)
			response = append(response, tempRecommendation)
		}
	}

	return response, nil
}

func (c *ControllerRepository) convertQueryCondition(queryCondition *datahub_v1alpha1.QueryCondition) string {
	ret := ""

	if queryCondition == nil {
		return ret
	}

	start := queryCondition.GetTimeRange().StartTime.Seconds
	end := queryCondition.GetTimeRange().EndTime.Seconds

	order := queryCondition.GetOrder()
	limit := queryCondition.GetLimit()

	if start > 0 {
		tm := time.Unix(int64(start), 0)
		ret += fmt.Sprintf("AND time>'%s' ", tm.UTC().Format(time.RFC3339))
	}

	if end > 0 {
		tm := time.Unix(int64(end), 0)
		ret += fmt.Sprintf("AND time<'%s' ", tm.UTC().Format(time.RFC3339))
	}

	if order == 0 {
		ret += "ORDER BY time ASC "
	} else {
		ret += "ORDER BY time DESC "
	}

	if limit > 0 {
		ret += fmt.Sprintf("LIMIT %d ", limit)
	}

	return ret
}

// NewEntityFromMap Build entity from map
func (c *ControllerRepository) NewRecommendationFromMap(data map[string]string) *datahub_v1alpha1.ControllerRecommendation {

	currentReplicas, _ := strconv.ParseInt(data[string(recommendation_entity.ControllerCurrentReplicas)], 10, 64)
	desiredReplicas, _ := strconv.ParseInt(data[string(recommendation_entity.ControllerDesiredReplicas)], 10, 64)
	createTime, _ := strconv.ParseInt(data[string(recommendation_entity.ControllerCreateTime)], 10, 64)

	var commendationType datahub_v1alpha1.ControllerRecommendedType

	if tempType, exist := data[string(recommendation_entity.ControllerType)]; exist {
		i, _ := strconv.ParseInt(tempType, 10, 32)
		commendationType = datahub_v1alpha1.ControllerRecommendedType(i)
	} else {
		commendationType = datahub_v1alpha1.ControllerRecommendedType_CRT_Primitive
	}

	tempRecommendation := datahub_v1alpha1.ControllerRecommendation{
		RecommendedType: commendationType,
		RecommendedSpec: &datahub_v1alpha1.ControllerRecommendedSpec{
			NamespacedName: &datahub_v1alpha1.NamespacedName{
				Namespace: data[string(recommendation_entity.ControllerNamespace)],
				Name:      data[string(recommendation_entity.ControllerName)],
			},
			CurrentReplicas: int32(currentReplicas),
			DesiredReplicas: int32(desiredReplicas),
			CreateTime: &timestamp.Timestamp{
				Seconds: createTime,
			},
		},
	}

	return &tempRecommendation
}
