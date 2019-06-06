package recommendation

import (
	"fmt"
	recommendation_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/recommendation"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"strings"
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
	points := make([]*influxdb_client.Point, 0)
	for _, conrollerRecommendation := range controllerRecommendations {
		recommendedType := conrollerRecommendation.GetRecommendedType()

		if recommendedType == datahub_v1alpha1.ControllerRecommendedType_CRT_Primitive {
			recommendedSpec := conrollerRecommendation.GetRecommendedSpec()

			tags := map[string]string{
				recommendation_entity.ControllerNamespace: recommendedSpec.GetNamespacedName().GetNamespace(),
				recommendation_entity.ControllerName:      recommendedSpec.GetNamespacedName().GetName(),
			}

			fields := map[string]interface{}{
				recommendation_entity.ControllerType:            datahub_v1alpha1.ControllerRecommendedType_CRT_Primitive,
				recommendation_entity.ControllerCurrentReplicas: recommendedSpec.GetCurrentReplicas(),
				recommendation_entity.ControllerDesiredReplicas: recommendedSpec.GetDesiredReplicas(),
				recommendation_entity.ControllerCreateTime:      recommendedSpec.GetCreateTime().GetSeconds(),
				recommendation_entity.ControllerKind:            recommendedSpec.GetKind().String(),

				recommendation_entity.ControllerCurrentCPURequest: recommendedSpec.GetCurrentCpuRequests(),
				recommendation_entity.ControllerCurrentMEMRequest: recommendedSpec.GetCurrentMemRequests(),
				recommendation_entity.ControllerCurrentCPULimit:   recommendedSpec.GetCurrentCpuLimits(),
				recommendation_entity.ControllerCurrentMEMLimit:   recommendedSpec.GetCurrentMemLimits(),
				recommendation_entity.ControllerDesiredCPULimit:   recommendedSpec.GetDesiredCpuLimits(),
				recommendation_entity.ControllerDesiredMEMLimit:   recommendedSpec.GetDesiredMemLimits(),
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

	namespace := controllerNamespacedName.GetNamespace()
	name := controllerNamespacedName.GetName()

	whereStr := c.convertQueryCondition(namespace, name, queryCondition)
	cmd := fmt.Sprintf("SELECT * FROM %s %s", string(Controller), whereStr)

	results, err := c.influxDB.QueryDB(cmd, string(influxdb.Recommendation))
	if err != nil {
		return make([]*datahub_v1alpha1.ControllerRecommendation, 0), err
	}

	influxdbRows := influxdb.PackMap(results)
	recommendations := c.getControllersRecommendationsFromInfluxRows(influxdbRows)

	return recommendations, nil
}

func (c *ControllerRepository) convertQueryCondition(namespace string, name string, queryCondition *datahub_v1alpha1.QueryCondition) string {
	ret := ""

	if namespace != "" {
		ret += fmt.Sprintf("\"namespace\"='%s' ", namespace)
	}

	if name != "" {
		ret += fmt.Sprintf("AND \"name\"='%s' ", name)
	}

	if queryCondition == nil {
		ret = strings.TrimPrefix(ret, "AND")
		if ret != "" {
			ret = "WHERE " + ret
		}
		return ret
	}

	start := queryCondition.GetTimeRange().GetStartTime().GetSeconds()
	end := queryCondition.GetTimeRange().GetEndTime().GetSeconds()

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

	ret = strings.TrimPrefix(ret, "AND")
	if ret != "" {
		ret = "WHERE " + ret
	}
	return ret
}

func (c *ControllerRepository) getControllersRecommendationsFromInfluxRows(rows []*influxdb.InfluxDBRow) []*datahub_v1alpha1.ControllerRecommendation {
	recommendations := make([]*datahub_v1alpha1.ControllerRecommendation, 0)
	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			currentReplicas, _ := strconv.ParseInt(data[recommendation_entity.ControllerCurrentReplicas], 10, 64)
			desiredReplicas, _ := strconv.ParseInt(data[recommendation_entity.ControllerDesiredReplicas], 10, 64)
			createTime, _ := strconv.ParseInt(data[recommendation_entity.ControllerCreateTime], 10, 64)

			currentCpuRequests, _ := strconv.ParseFloat(data[recommendation_entity.ControllerCurrentCPURequest], 32)
			currentMemRequests, _ := strconv.ParseFloat(data[recommendation_entity.ControllerCurrentMEMRequest], 32)
			currentCpuLimits, _ := strconv.ParseFloat(data[recommendation_entity.ControllerCurrentCPULimit], 32)
			currentMemLimits, _ := strconv.ParseFloat(data[recommendation_entity.ControllerCurrentMEMLimit], 32)
			desiredCpuLimits, _ := strconv.ParseFloat(data[recommendation_entity.ControllerDesiredCPULimit], 32)
			desiredMemLimits, _ := strconv.ParseFloat(data[recommendation_entity.ControllerDesiredMEMLimit], 32)

			var commendationType datahub_v1alpha1.ControllerRecommendedType
			if tempType, exist := data[recommendation_entity.ControllerType]; exist {
				if value, ok := datahub_v1alpha1.ControllerRecommendedType_value[tempType]; ok {
					commendationType = datahub_v1alpha1.ControllerRecommendedType(value)
				}
			}

			var commendationKind datahub_v1alpha1.Kind
			if tempKind, exist := data[recommendation_entity.ControllerKind]; exist {
				if value, ok := datahub_v1alpha1.Kind_value[tempKind]; ok {
					commendationKind = datahub_v1alpha1.Kind(value)
				}
			}

			tempRecommendation := &datahub_v1alpha1.ControllerRecommendation{
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
					Kind:               commendationKind,
					CurrentCpuRequests: float32(currentCpuRequests),
					CurrentMemRequests: float32(currentMemRequests),
					CurrentCpuLimits:   float32(currentCpuLimits),
					CurrentMemLimits:   float32(currentMemLimits),
					DesiredCpuLimits:   float32(desiredCpuLimits),
					DesiredMemLimits:   float32(desiredMemLimits),
				},
			}

			recommendations = append(recommendations, tempRecommendation)
		}
	}

	return recommendations
}
