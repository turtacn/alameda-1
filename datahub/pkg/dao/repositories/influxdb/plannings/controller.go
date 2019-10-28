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

type ControllerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewControllerRepository(influxDBCfg *InternalInflux.Config) *ControllerRepository {
	return &ControllerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *ControllerRepository) CreateControllerPlannings(controllerPlannings []*ApiPlannings.ControllerPlanning) error {
	points := make([]*InfluxClient.Point, 0)

	for _, controllerPlanning := range controllerPlannings {
		ctlPlanningType := controllerPlanning.GetCtlPlanningType()

		if ctlPlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
			planningSpec := controllerPlanning.GetCtlPlanningSpec()

			tags := map[string]string{
				EntityInfluxPlanning.ControllerPlanningType: controllerPlanning.GetPlanningType().String(),
				EntityInfluxPlanning.ControllerNamespace:    planningSpec.GetNamespacedName().GetNamespace(),
				EntityInfluxPlanning.ControllerName:         planningSpec.GetNamespacedName().GetName(),
				EntityInfluxPlanning.ControllerType:         ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxPlanning.ControllerCurrentReplicas: planningSpec.GetCurrentReplicas(),
				EntityInfluxPlanning.ControllerDesiredReplicas: planningSpec.GetDesiredReplicas(),
				EntityInfluxPlanning.ControllerCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
				EntityInfluxPlanning.ControllerKind:            planningSpec.GetKind().String(),

				EntityInfluxPlanning.ControllerCurrentCPURequest: planningSpec.GetCurrentCpuRequests(),
				EntityInfluxPlanning.ControllerCurrentMEMRequest: planningSpec.GetCurrentMemRequests(),
				EntityInfluxPlanning.ControllerCurrentCPULimit:   planningSpec.GetCurrentCpuLimits(),
				EntityInfluxPlanning.ControllerCurrentMEMLimit:   planningSpec.GetCurrentMemLimits(),
				EntityInfluxPlanning.ControllerDesiredCPULimit:   planningSpec.GetDesiredCpuLimits(),
				EntityInfluxPlanning.ControllerDesiredMEMLimit:   planningSpec.GetDesiredMemLimits(),
				EntityInfluxPlanning.ControllerTotalCost:         planningSpec.GetTotalCost(),
			}

			pt, err := InfluxClient.NewPoint(string(Controller), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)

		} else if ctlPlanningType == ApiPlannings.ControllerPlanningType_CPT_K8S {
			planningSpec := controllerPlanning.GetCtlPlanningSpecK8S()

			tags := map[string]string{
				EntityInfluxPlanning.ControllerPlanningType: controllerPlanning.GetPlanningType().String(),
				EntityInfluxPlanning.ControllerNamespace:    planningSpec.GetNamespacedName().GetNamespace(),
				EntityInfluxPlanning.ControllerName:         planningSpec.GetNamespacedName().GetName(),
				EntityInfluxPlanning.ControllerType:         ApiPlannings.ControllerPlanningType_CPT_K8S.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxPlanning.ControllerCurrentReplicas: planningSpec.GetCurrentReplicas(),
				EntityInfluxPlanning.ControllerDesiredReplicas: planningSpec.GetDesiredReplicas(),
				EntityInfluxPlanning.ControllerCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
				EntityInfluxPlanning.ControllerKind:            planningSpec.GetKind().String(),
			}

			pt, err := InfluxClient.NewPoint(string(Controller), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
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

func (c *ControllerRepository) ListControllerPlannings(in *ApiPlannings.ListControllerPlanningsRequest) ([]*ApiPlannings.ControllerPlanning, error) {
	namespace := in.GetNamespacedName().GetNamespace()
	name := in.GetNamespacedName().GetName()
	ctlPlanningType := in.GetCtlPlanningType()

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Controller,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	influxdbStatement.AppendWhereClause(EntityInfluxPlanning.ControllerNamespace, "=", namespace)
	influxdbStatement.AppendWhereClause(EntityInfluxPlanning.ControllerName, "=", name)
	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()

	if ctlPlanningType != ApiPlannings.ControllerPlanningType_CPT_UNDEFINED {
		influxdbStatement.AppendWhereClause(EntityInfluxPlanning.ControllerType, "=", ctlPlanningType.String())
	}

	cmd := influxdbStatement.BuildQueryCmd()

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return make([]*ApiPlannings.ControllerPlanning, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	recommendations := c.getControllersPlanningsFromInfluxRows(influxdbRows)

	return recommendations, nil
}

func (c *ControllerRepository) getControllersPlanningsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiPlannings.ControllerPlanning {
	plannings := make([]*ApiPlannings.ControllerPlanning, 0)

	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			currentReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.ControllerCurrentReplicas], 10, 64)
			desiredReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.ControllerDesiredReplicas], 10, 64)
			createTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ControllerCreateTime], 10, 64)

			t, _ := time.Parse(time.RFC3339, data[EntityInfluxPlanning.ControllerTime])
			tempTime, _ := ptypes.TimestampProto(t)

			currentCpuRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerCurrentCPURequest], 64)
			currentMemRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerCurrentMEMRequest], 64)
			currentCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerCurrentCPULimit], 64)
			currentMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerCurrentMEMLimit], 64)
			desiredCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerDesiredCPULimit], 64)
			desiredMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerDesiredMEMLimit], 64)
			totalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerTotalCost], 64)

			var ctlPlanningType ApiPlannings.ControllerPlanningType
			if tempType, exist := data[EntityInfluxPlanning.ControllerType]; exist {
				if value, ok := ApiPlannings.ControllerPlanningType_value[tempType]; ok {
					ctlPlanningType = ApiPlannings.ControllerPlanningType(value)
				}
			}

			var planningKind ApiResources.Kind
			if tempKind, exist := data[EntityInfluxPlanning.ControllerKind]; exist {
				if value, ok := ApiResources.Kind_value[tempKind]; ok {
					planningKind = ApiResources.Kind(value)
				}
			}

			if ctlPlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
				tempPlanning := &ApiPlannings.ControllerPlanning{
					PlanningType:    ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.ControllerPlanningType)]]),
					CtlPlanningType: ctlPlanningType,
					CtlPlanningSpec: &ApiPlannings.ControllerPlanningSpec{
						NamespacedName: &ApiResources.NamespacedName{
							Namespace: data[string(EntityInfluxPlanning.ControllerNamespace)],
							Name:      data[string(EntityInfluxPlanning.ControllerName)],
						},
						CurrentReplicas: int32(currentReplicas),
						DesiredReplicas: int32(desiredReplicas),
						Time:            tempTime,
						CreateTime: &timestamp.Timestamp{
							Seconds: createTime,
						},
						Kind:               planningKind,
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
				tempPlanning := &ApiPlannings.ControllerPlanning{
					PlanningType:    ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.ControllerPlanningType)]]),
					CtlPlanningType: ctlPlanningType,
					CtlPlanningSpecK8S: &ApiPlannings.ControllerPlanningSpecK8S{
						NamespacedName: &ApiResources.NamespacedName{
							Namespace: data[string(EntityInfluxPlanning.ControllerNamespace)],
							Name:      data[string(EntityInfluxPlanning.ControllerName)],
						},
						CurrentReplicas: int32(currentReplicas),
						DesiredReplicas: int32(desiredReplicas),
						Time:            tempTime,
						CreateTime: &timestamp.Timestamp{
							Seconds: createTime,
						},
						Kind: planningKind,
					},
				}

				plannings = append(plannings, tempPlanning)
			}
		}
	}

	return plannings
}
