package recommendations

import (
	EntityInfluxRecommend "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/recommendations"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"time"
)

type ClusterRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewClusterRepository(influxDBCfg *InternalInflux.Config) *ClusterRepository {
	return &ClusterRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *ClusterRepository) CreateRecommendations(recommendations []*ApiRecommendations.ClusterRecommendation) error {
	points := make([]*InfluxClient.Point, 0)
	for _, recommendation := range recommendations {
		recommendedType := recommendation.GetRecommendedType()

		if recommendedType == ApiRecommendations.ControllerRecommendedType_PRIMITIVE {
			recommendedSpec := recommendation.GetRecommendedSpec()

			tags := map[string]string{
				EntityInfluxRecommend.ClusterName: recommendation.GetObjectMeta().GetName(),
				EntityInfluxRecommend.ClusterType: ApiRecommendations.ControllerRecommendedType_PRIMITIVE.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxRecommend.ClusterCurrentReplicas: recommendedSpec.GetCurrentReplicas(),
				EntityInfluxRecommend.ClusterDesiredReplicas: recommendedSpec.GetDesiredReplicas(),
				EntityInfluxRecommend.ClusterCreateTime:      recommendedSpec.GetCreateTime().GetSeconds(),
				EntityInfluxRecommend.ClusterKind:            recommendation.GetKind().String(),

				EntityInfluxRecommend.ClusterCurrentCPURequest: recommendedSpec.GetCurrentCpuRequests(),
				EntityInfluxRecommend.ClusterCurrentMEMRequest: recommendedSpec.GetCurrentMemRequests(),
				EntityInfluxRecommend.ClusterCurrentCPULimit:   recommendedSpec.GetCurrentCpuLimits(),
				EntityInfluxRecommend.ClusterCurrentMEMLimit:   recommendedSpec.GetCurrentMemLimits(),
				EntityInfluxRecommend.ClusterDesiredCPULimit:   recommendedSpec.GetDesiredCpuLimits(),
				EntityInfluxRecommend.ClusterDesiredMEMLimit:   recommendedSpec.GetDesiredMemLimits(),
				EntityInfluxRecommend.ClusterTotalCost:         recommendedSpec.GetTotalCost(),
			}

			pt, err := InfluxClient.NewPoint(string(Cluster), tags, fields, time.Unix(recommendedSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)

		} else if recommendedType == ApiRecommendations.ControllerRecommendedType_K8S {
			recommendedSpec := recommendation.GetRecommendedSpecK8S()

			tags := map[string]string{
				EntityInfluxRecommend.ClusterName: recommendation.GetObjectMeta().GetName(),
				EntityInfluxRecommend.ClusterType: ApiRecommendations.ControllerRecommendedType_K8S.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxRecommend.ClusterCurrentReplicas: recommendedSpec.GetCurrentReplicas(),
				EntityInfluxRecommend.ClusterDesiredReplicas: recommendedSpec.GetDesiredReplicas(),
				EntityInfluxRecommend.ClusterCreateTime:      recommendedSpec.GetCreateTime().GetSeconds(),
				EntityInfluxRecommend.ClusterKind:            recommendation.GetKind().String(),
			}

			pt, err := InfluxClient.NewPoint(string(Cluster), tags, fields, time.Unix(recommendedSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)
		}
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

func (c *ClusterRepository) ListRecommendations(in *ApiRecommendations.ListClusterRecommendationsRequest) ([]*ApiRecommendations.ClusterRecommendation, error) {
	influxdbStatement := InternalInflux.Statement{
		Measurement:    Cluster,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	recommendationType := in.GetRecommendedType().String()
	kind := in.GetKind().String()

	for _, objMeta := range in.GetObjectMeta() {
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxRecommend.ClusterName,
			EntityInfluxRecommend.ClusterKind,
		}
		valueList := []string{name, kind}

		if recommendationType != ApiRecommendations.ControllerRecommendedType_CRT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxRecommend.ControllerType)
			valueList = append(valueList, recommendationType)
		}

		tempCondition := influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Recommendation))
	if err != nil {
		return make([]*ApiRecommendations.ClusterRecommendation, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	recommendations := c.getRecommendationsFromInfluxRows(influxdbRows)

	return recommendations, nil
}

func (c *ClusterRepository) getRecommendationsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiRecommendations.ClusterRecommendation {
	recommendations := make([]*ApiRecommendations.ClusterRecommendation, 0)
	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			currentReplicas, _ := strconv.ParseInt(data[EntityInfluxRecommend.ClusterCurrentReplicas], 10, 64)
			desiredReplicas, _ := strconv.ParseInt(data[EntityInfluxRecommend.ClusterDesiredReplicas], 10, 64)
			createTime, _ := strconv.ParseInt(data[EntityInfluxRecommend.ClusterCreateTime], 10, 64)

			t, _ := time.Parse(time.RFC3339, data[EntityInfluxRecommend.ClusterTime])
			tempTime, _ := ptypes.TimestampProto(t)

			currentCpuRequests, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ClusterCurrentCPURequest], 64)
			currentMemRequests, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ClusterCurrentMEMRequest], 64)
			currentCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ClusterCurrentCPULimit], 64)
			currentMemLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ClusterCurrentMEMLimit], 64)
			desiredCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ClusterDesiredCPULimit], 64)
			desiredMemLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ClusterDesiredMEMLimit], 64)
			totalCost, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ClusterTotalCost], 64)

			var commendationType ApiRecommendations.ControllerRecommendedType
			if tempType, exist := data[EntityInfluxRecommend.ClusterType]; exist {
				if value, ok := ApiRecommendations.ControllerRecommendedType_value[tempType]; ok {
					commendationType = ApiRecommendations.ControllerRecommendedType(value)
				}
			}

			var commendationKind ApiResources.Kind
			if tempKind, exist := data[EntityInfluxRecommend.ClusterKind]; exist {
				if value, ok := ApiResources.Kind_value[tempKind]; ok {
					commendationKind = ApiResources.Kind(value)
				}
			}

			if commendationType == ApiRecommendations.ControllerRecommendedType_PRIMITIVE {
				tempRecommendation := &ApiRecommendations.ClusterRecommendation{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name: data[string(EntityInfluxRecommend.ClusterName)],
					},
					Kind:            commendationKind,
					RecommendedType: commendationType,
					RecommendedSpec: &ApiRecommendations.ControllerRecommendedSpec{
						CurrentReplicas: int32(currentReplicas),
						DesiredReplicas: int32(desiredReplicas),
						Time:            tempTime,
						CreateTime: &timestamp.Timestamp{
							Seconds: createTime,
						},
						CurrentCpuRequests: currentCpuRequests,
						CurrentMemRequests: currentMemRequests,
						CurrentCpuLimits:   currentCpuLimits,
						CurrentMemLimits:   currentMemLimits,
						DesiredCpuLimits:   desiredCpuLimits,
						DesiredMemLimits:   desiredMemLimits,
						TotalCost:          totalCost,
					},
				}

				recommendations = append(recommendations, tempRecommendation)

			} else if commendationType == ApiRecommendations.ControllerRecommendedType_K8S {
				tempRecommendation := &ApiRecommendations.ClusterRecommendation{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name: data[string(EntityInfluxRecommend.ClusterName)],
					},
					Kind:            commendationKind,
					RecommendedType: commendationType,
					RecommendedSpecK8S: &ApiRecommendations.ControllerRecommendedSpecK8S{
						CurrentReplicas: int32(currentReplicas),
						DesiredReplicas: int32(desiredReplicas),
						Time:            tempTime,
						CreateTime: &timestamp.Timestamp{
							Seconds: createTime,
						},
					},
				}

				recommendations = append(recommendations, tempRecommendation)
			}
		}
	}

	return recommendations
}
