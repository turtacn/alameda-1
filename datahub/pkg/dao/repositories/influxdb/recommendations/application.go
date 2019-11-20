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
		recommendedType := recommendation.GetRecommendedType()

		if recommendedType == ApiRecommendations.ControllerRecommendedType_PRIMITIVE {
			recommendedSpec := recommendation.GetRecommendedSpec()

			tags := map[string]string{
				EntityInfluxRecommend.AppNamespace: recommendation.GetObjectMeta().GetNamespace(),
				EntityInfluxRecommend.AppName:      recommendation.GetObjectMeta().GetName(),
				EntityInfluxRecommend.AppType:      ApiRecommendations.ControllerRecommendedType_PRIMITIVE.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxRecommend.AppCurrentReplicas: recommendedSpec.GetCurrentReplicas(),
				EntityInfluxRecommend.AppDesiredReplicas: recommendedSpec.GetDesiredReplicas(),
				EntityInfluxRecommend.AppCreateTime:      recommendedSpec.GetCreateTime().GetSeconds(),
				EntityInfluxRecommend.AppKind:            recommendation.GetKind().String(),

				EntityInfluxRecommend.AppCurrentCPURequest: recommendedSpec.GetCurrentCpuRequests(),
				EntityInfluxRecommend.AppCurrentMEMRequest: recommendedSpec.GetCurrentMemRequests(),
				EntityInfluxRecommend.AppCurrentCPULimit:   recommendedSpec.GetCurrentCpuLimits(),
				EntityInfluxRecommend.AppCurrentMEMLimit:   recommendedSpec.GetCurrentMemLimits(),
				EntityInfluxRecommend.AppDesiredCPULimit:   recommendedSpec.GetDesiredCpuLimits(),
				EntityInfluxRecommend.AppDesiredMEMLimit:   recommendedSpec.GetDesiredMemLimits(),
				EntityInfluxRecommend.AppTotalCost:         recommendedSpec.GetTotalCost(),
			}

			pt, err := InfluxClient.NewPoint(string(Application), tags, fields, time.Unix(recommendedSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)

		} else if recommendedType == ApiRecommendations.ControllerRecommendedType_K8S {
			recommendedSpec := recommendation.GetRecommendedSpecK8S()

			tags := map[string]string{
				EntityInfluxRecommend.AppNamespace: recommendation.GetObjectMeta().GetNamespace(),
				EntityInfluxRecommend.AppName:      recommendation.GetObjectMeta().GetName(),
				EntityInfluxRecommend.AppType:      ApiRecommendations.ControllerRecommendedType_K8S.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxRecommend.AppCurrentReplicas: recommendedSpec.GetCurrentReplicas(),
				EntityInfluxRecommend.AppDesiredReplicas: recommendedSpec.GetDesiredReplicas(),
				EntityInfluxRecommend.AppCreateTime:      recommendedSpec.GetCreateTime().GetSeconds(),
				EntityInfluxRecommend.AppKind:            recommendation.GetKind().String(),
			}

			pt, err := InfluxClient.NewPoint(string(Application), tags, fields, time.Unix(recommendedSpec.GetTime().GetSeconds(), 0))
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

func (c *AppRepository) ListRecommendations(in *ApiRecommendations.ListApplicationRecommendationsRequest) ([]*ApiRecommendations.ApplicationRecommendation, error) {
	influxdbStatement := InternalInflux.Statement{
		Measurement:    Application,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	recommendationType := in.GetRecommendedType().String()
	kind := in.GetKind().String()

	for _, objMeta := range in.GetObjectMeta() {
		namespace := objMeta.GetNamespace()
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxRecommend.AppNamespace,
			EntityInfluxRecommend.AppName,
			EntityInfluxRecommend.AppKind,
		}
		valueList := []string{namespace, name, kind}

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
			currentReplicas, _ := strconv.ParseInt(data[EntityInfluxRecommend.AppCurrentReplicas], 10, 64)
			desiredReplicas, _ := strconv.ParseInt(data[EntityInfluxRecommend.AppDesiredReplicas], 10, 64)
			createTime, _ := strconv.ParseInt(data[EntityInfluxRecommend.AppCreateTime], 10, 64)

			t, _ := time.Parse(time.RFC3339, data[EntityInfluxRecommend.AppTime])
			tempTime, _ := ptypes.TimestampProto(t)

			currentCpuRequests, _ := strconv.ParseFloat(data[EntityInfluxRecommend.AppCurrentCPURequest], 64)
			currentMemRequests, _ := strconv.ParseFloat(data[EntityInfluxRecommend.AppCurrentMEMRequest], 64)
			currentCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.AppCurrentCPULimit], 64)
			currentMemLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.AppCurrentMEMLimit], 64)
			desiredCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.AppDesiredCPULimit], 64)
			desiredMemLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.AppDesiredMEMLimit], 64)
			totalCost, _ := strconv.ParseFloat(data[EntityInfluxRecommend.AppTotalCost], 64)

			var commendationType ApiRecommendations.ControllerRecommendedType
			if tempType, exist := data[EntityInfluxRecommend.AppType]; exist {
				if value, ok := ApiRecommendations.ControllerRecommendedType_value[tempType]; ok {
					commendationType = ApiRecommendations.ControllerRecommendedType(value)
				}
			}

			var commendationKind ApiResources.Kind
			if tempKind, exist := data[EntityInfluxRecommend.AppKind]; exist {
				if value, ok := ApiResources.Kind_value[tempKind]; ok {
					commendationKind = ApiResources.Kind(value)
				}
			}

			if commendationType == ApiRecommendations.ControllerRecommendedType_PRIMITIVE {
				tempRecommendation := &ApiRecommendations.ApplicationRecommendation{
					ObjectMeta: &ApiResources.ObjectMeta{
						Namespace: data[string(EntityInfluxRecommend.AppNamespace)],
						Name:      data[string(EntityInfluxRecommend.AppName)],
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
				tempRecommendation := &ApiRecommendations.ApplicationRecommendation{
					ObjectMeta: &ApiResources.ObjectMeta{
						Namespace: data[string(EntityInfluxRecommend.AppNamespace)],
						Name:      data[string(EntityInfluxRecommend.AppName)],
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
