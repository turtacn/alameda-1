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

type NodeRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNodeRepository(influxDBCfg *InternalInflux.Config) *NodeRepository {
	return &NodeRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *NodeRepository) CreateRecommendations(recommendations []*ApiRecommendations.NodeRecommendation) error {
	points := make([]*InfluxClient.Point, 0)
	for _, recommendation := range recommendations {
		recommendedType := recommendation.GetRecommendedType()

		if recommendedType == ApiRecommendations.ControllerRecommendedType_PRIMITIVE {
			recommendedSpec := recommendation.GetRecommendedSpec()

			tags := map[string]string{
				EntityInfluxRecommend.NodeName: recommendation.GetObjectMeta().GetName(),
				EntityInfluxRecommend.NodeType: ApiRecommendations.ControllerRecommendedType_PRIMITIVE.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxRecommend.NodeCurrentReplicas: recommendedSpec.GetCurrentReplicas(),
				EntityInfluxRecommend.NodeDesiredReplicas: recommendedSpec.GetDesiredReplicas(),
				EntityInfluxRecommend.NodeCreateTime:      recommendedSpec.GetCreateTime().GetSeconds(),
				EntityInfluxRecommend.NodeKind:            recommendation.GetKind().String(),

				EntityInfluxRecommend.NodeCurrentCPURequest: recommendedSpec.GetCurrentCpuRequests(),
				EntityInfluxRecommend.NodeCurrentMEMRequest: recommendedSpec.GetCurrentMemRequests(),
				EntityInfluxRecommend.NodeCurrentCPULimit:   recommendedSpec.GetCurrentCpuLimits(),
				EntityInfluxRecommend.NodeCurrentMEMLimit:   recommendedSpec.GetCurrentMemLimits(),
				EntityInfluxRecommend.NodeDesiredCPULimit:   recommendedSpec.GetDesiredCpuLimits(),
				EntityInfluxRecommend.NodeDesiredMEMLimit:   recommendedSpec.GetDesiredMemLimits(),
				EntityInfluxRecommend.NodeTotalCost:         recommendedSpec.GetTotalCost(),
			}

			pt, err := InfluxClient.NewPoint(string(Node), tags, fields, time.Unix(recommendedSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)

		} else if recommendedType == ApiRecommendations.ControllerRecommendedType_K8S {
			recommendedSpec := recommendation.GetRecommendedSpecK8S()

			tags := map[string]string{
				EntityInfluxRecommend.NodeName: recommendation.GetObjectMeta().GetName(),
				EntityInfluxRecommend.NodeType: ApiRecommendations.ControllerRecommendedType_K8S.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxRecommend.NodeCurrentReplicas: recommendedSpec.GetCurrentReplicas(),
				EntityInfluxRecommend.NodeDesiredReplicas: recommendedSpec.GetDesiredReplicas(),
				EntityInfluxRecommend.NodeCreateTime:      recommendedSpec.GetCreateTime().GetSeconds(),
				EntityInfluxRecommend.NodeKind:            recommendation.GetKind().String(),
			}

			pt, err := InfluxClient.NewPoint(string(Node), tags, fields, time.Unix(recommendedSpec.GetTime().GetSeconds(), 0))
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

func (c *NodeRepository) ListRecommendations(in *ApiRecommendations.ListNodeRecommendationsRequest) ([]*ApiRecommendations.NodeRecommendation, error) {
	influxdbStatement := InternalInflux.Statement{
		Measurement:    Node,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	recommendationType := in.GetRecommendedType().String()
	kind := in.GetKind().String()

	for _, objMeta := range in.GetObjectMeta() {
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxRecommend.NodeName,
			EntityInfluxRecommend.NodeKind,
		}
		valueList := []string{name, kind}

		if recommendationType != ApiRecommendations.ControllerRecommendedType_CRT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxRecommend.NodeType)
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
		return make([]*ApiRecommendations.NodeRecommendation, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	recommendations := c.getRecommendationsFromInfluxRows(influxdbRows)

	return recommendations, nil
}

func (c *NodeRepository) getRecommendationsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiRecommendations.NodeRecommendation {
	recommendations := make([]*ApiRecommendations.NodeRecommendation, 0)
	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			currentReplicas, _ := strconv.ParseInt(data[EntityInfluxRecommend.NodeCurrentReplicas], 10, 64)
			desiredReplicas, _ := strconv.ParseInt(data[EntityInfluxRecommend.NodeDesiredReplicas], 10, 64)
			createTime, _ := strconv.ParseInt(data[EntityInfluxRecommend.NodeCreateTime], 10, 64)

			t, _ := time.Parse(time.RFC3339, data[EntityInfluxRecommend.NodeTime])
			tempTime, _ := ptypes.TimestampProto(t)

			currentCpuRequests, _ := strconv.ParseFloat(data[EntityInfluxRecommend.NodeCurrentCPURequest], 64)
			currentMemRequests, _ := strconv.ParseFloat(data[EntityInfluxRecommend.NodeCurrentMEMRequest], 64)
			currentCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.NodeCurrentCPULimit], 64)
			currentMemLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.NodeCurrentMEMLimit], 64)
			desiredCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.NodeDesiredCPULimit], 64)
			desiredMemLimits, _ := strconv.ParseFloat(data[EntityInfluxRecommend.NodeDesiredMEMLimit], 64)
			totalCost, _ := strconv.ParseFloat(data[EntityInfluxRecommend.NodeTotalCost], 64)

			var commendationType ApiRecommendations.ControllerRecommendedType
			if tempType, exist := data[EntityInfluxRecommend.NodeType]; exist {
				if value, ok := ApiRecommendations.ControllerRecommendedType_value[tempType]; ok {
					commendationType = ApiRecommendations.ControllerRecommendedType(value)
				}
			}

			var commendationKind ApiResources.Kind
			if tempKind, exist := data[EntityInfluxRecommend.NodeKind]; exist {
				if value, ok := ApiResources.Kind_value[tempKind]; ok {
					commendationKind = ApiResources.Kind(value)
				}
			}

			if commendationType == ApiRecommendations.ControllerRecommendedType_PRIMITIVE {
				tempRecommendation := &ApiRecommendations.NodeRecommendation{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name: data[string(EntityInfluxRecommend.NodeName)],
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
				tempRecommendation := &ApiRecommendations.NodeRecommendation{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name: data[string(EntityInfluxRecommend.NodeName)],
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
