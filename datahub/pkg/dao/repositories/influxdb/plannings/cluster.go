package plannings

import (
	EntityInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/plannings"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
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

func (c *ClusterRepository) CreatePlannings(plannings []*ApiPlannings.ClusterPlanning) error {
	points := make([]*InfluxClient.Point, 0)

	for _, planning := range plannings {
		clusterPlanningType := planning.GetClusterPlanningType()

		if clusterPlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
			planningSpec := planning.GetClusterPlanningSpec()

			tags := map[string]string{
				EntityInfluxPlanning.ClusterPlanningType: planning.GetPlanningType().String(),
				EntityInfluxPlanning.ClusterName:         planning.GetObjectMeta().GetName(),
				EntityInfluxPlanning.ClusterType:         ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxPlanning.ClusterCurrentReplicas: planningSpec.GetCurrentReplicas(),
				EntityInfluxPlanning.ClusterDesiredReplicas: planningSpec.GetDesiredReplicas(),
				EntityInfluxPlanning.ClusterCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
				EntityInfluxPlanning.ClusterKind:            planning.GetKind().String(),

				EntityInfluxPlanning.ClusterCurrentCPURequest: planningSpec.GetCurrentCpuRequests(),
				EntityInfluxPlanning.ClusterCurrentMEMRequest: planningSpec.GetCurrentMemRequests(),
				EntityInfluxPlanning.ClusterCurrentCPULimit:   planningSpec.GetCurrentCpuLimits(),
				EntityInfluxPlanning.ClusterCurrentMEMLimit:   planningSpec.GetCurrentMemLimits(),
				EntityInfluxPlanning.ClusterDesiredCPULimit:   planningSpec.GetDesiredCpuLimits(),
				EntityInfluxPlanning.ClusterDesiredMEMLimit:   planningSpec.GetDesiredMemLimits(),
				EntityInfluxPlanning.ClusterTotalCost:         planningSpec.GetTotalCost(),
			}

			pt, err := InfluxClient.NewPoint(string(Cluster), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)

		} else if clusterPlanningType == ApiPlannings.ControllerPlanningType_CPT_K8S {
			planningSpec := planning.GetClusterPlanningSpecK8S()

			tags := map[string]string{
				EntityInfluxPlanning.ClusterPlanningType: planning.GetPlanningType().String(),
				EntityInfluxPlanning.ClusterName:         planning.GetObjectMeta().GetName(),
				EntityInfluxPlanning.ClusterType:         ApiPlannings.ControllerPlanningType_CPT_K8S.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxPlanning.ClusterCurrentReplicas: planningSpec.GetCurrentReplicas(),
				EntityInfluxPlanning.ClusterDesiredReplicas: planningSpec.GetDesiredReplicas(),
				EntityInfluxPlanning.ClusterCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
				EntityInfluxPlanning.ClusterKind:            planning.GetKind().String(),
			}

			pt, err := InfluxClient.NewPoint(string(Cluster), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)
		}
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

func (c *ClusterRepository) ListPlannings(in *ApiPlannings.ListClusterPlanningsRequest) ([]*ApiPlannings.ClusterPlanning, error) {
	influxdbStatement := InternalInflux.Statement{
		Measurement:    Cluster,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	planningType := in.GetPlanningType().String()
	ctlPlanningType := in.GetCtlPlanningType().String()
	kind := in.GetKind().String()

	for _, objMeta := range in.GetObjectMeta() {
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxPlanning.ClusterName,
			EntityInfluxPlanning.ClusterKind,
		}
		valueList := []string{name, kind}

		if ctlPlanningType != ApiPlannings.ControllerPlanningType_CPT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ClusterType)
			valueList = append(valueList, ctlPlanningType)
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ClusterPlanningType)
			valueList = append(valueList, planningType)
		}

		tempCondition := influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()

	cmd := influxdbStatement.BuildQueryCmd()

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return make([]*ApiPlannings.ClusterPlanning, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	plannings := c.getPlanningsFromInfluxRows(influxdbRows)

	return plannings, nil
}

func (c *ClusterRepository) getPlanningsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiPlannings.ClusterPlanning {
	plannings := make([]*ApiPlannings.ClusterPlanning, 0)

	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			currentReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.ClusterCurrentReplicas], 10, 64)
			desiredReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.ClusterDesiredReplicas], 10, 64)
			createTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ClusterCreateTime], 10, 64)

			t, _ := time.Parse(time.RFC3339, data[EntityInfluxPlanning.ClusterTime])
			tempTime, _ := ptypes.TimestampProto(t)

			currentCpuRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterCurrentCPURequest], 64)
			currentMemRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterCurrentMEMRequest], 64)
			currentCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterCurrentCPULimit], 64)
			currentMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterCurrentMEMLimit], 64)
			desiredCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterDesiredCPULimit], 64)
			desiredMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterDesiredMEMLimit], 64)
			totalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterTotalCost], 64)

			var ctlPlanningType ApiPlannings.ControllerPlanningType
			if tempType, exist := data[EntityInfluxPlanning.ClusterType]; exist {
				if value, ok := ApiPlannings.ControllerPlanningType_value[tempType]; ok {
					ctlPlanningType = ApiPlannings.ControllerPlanningType(value)
				}
			}

			var planningKind ApiResources.Kind
			if tempKind, exist := data[EntityInfluxPlanning.ClusterKind]; exist {
				if value, ok := ApiResources.Kind_value[tempKind]; ok {
					planningKind = ApiResources.Kind(value)
				}
			}

			if ctlPlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
				tempPlanning := &ApiPlannings.ClusterPlanning{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name: data[string(EntityInfluxPlanning.ClusterName)],
					},
					Kind:                planningKind,
					PlanningType:        ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.ClusterPlanningType)]]),
					ClusterPlanningType: ctlPlanningType,
					ClusterPlanningSpec: &ApiPlannings.ControllerPlanningSpec{
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

				plannings = append(plannings, tempPlanning)

			} else if ctlPlanningType == ApiPlannings.ControllerPlanningType_CPT_K8S {
				tempPlanning := &ApiPlannings.ClusterPlanning{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name: data[string(EntityInfluxPlanning.ClusterName)],
					},
					Kind:                planningKind,
					PlanningType:        ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.ClusterPlanningType)]]),
					ClusterPlanningType: ctlPlanningType,
					ClusterPlanningSpecK8S: &ApiPlannings.ControllerPlanningSpecK8S{
						CurrentReplicas: int32(currentReplicas),
						DesiredReplicas: int32(desiredReplicas),
						Time:            tempTime,
						CreateTime: &timestamp.Timestamp{
							Seconds: createTime,
						},
					},
				}

				plannings = append(plannings, tempPlanning)
			}
		}
	}

	return plannings
}
