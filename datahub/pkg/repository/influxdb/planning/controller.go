package planning

import (
	EntityInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/planning"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"time"
)

type ControllerRepository struct {
	influxDB *RepoInflux.InfluxDBRepository
}

func NewControllerRepository(influxDBCfg *RepoInflux.Config) *ControllerRepository {
	return &ControllerRepository{
		influxDB: &RepoInflux.InfluxDBRepository{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *ControllerRepository) CreateControllerPlannings(controllerPlannings []*DatahubV1alpha1.ControllerPlanning) error {
	points := make([]*InfluxClient.Point, 0)

	for _, controllerPlanning := range controllerPlannings {
		ctlPlanningType := controllerPlanning.GetCtlPlanningType()

		if ctlPlanningType == DatahubV1alpha1.ControllerPlanningType_CPT_PRIMITIVE {
			planningSpec := controllerPlanning.GetCtlPlanningSpec()

			tags := map[string]string{
				EntityInfluxPlanning.ControllerPlanningType: controllerPlanning.GetPlanningType().String(),
				EntityInfluxPlanning.ControllerNamespace:    planningSpec.GetNamespacedName().GetNamespace(),
				EntityInfluxPlanning.ControllerName:         planningSpec.GetNamespacedName().GetName(),
				EntityInfluxPlanning.ControllerType:         DatahubV1alpha1.ControllerPlanningType_CPT_PRIMITIVE.String(),
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

		} else if ctlPlanningType == DatahubV1alpha1.ControllerPlanningType_CPT_K8S {
			planningSpec := controllerPlanning.GetCtlPlanningSpecK8S()

			tags := map[string]string{
				EntityInfluxPlanning.ControllerPlanningType: controllerPlanning.GetPlanningType().String(),
				EntityInfluxPlanning.ControllerNamespace:    planningSpec.GetNamespacedName().GetNamespace(),
				EntityInfluxPlanning.ControllerName:         planningSpec.GetNamespacedName().GetName(),
				EntityInfluxPlanning.ControllerType:         DatahubV1alpha1.ControllerPlanningType_CPT_K8S.String(),
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

func (c *ControllerRepository) ListControllerPlannings(in *DatahubV1alpha1.ListControllerPlanningsRequest) ([]*DatahubV1alpha1.ControllerPlanning, error) {
	namespace := in.GetNamespacedName().GetNamespace()
	name := in.GetNamespacedName().GetName()
	ctlPlanningType := in.GetCtlPlanningType()

	influxdbStatement := RepoInflux.StatementNew{
		Measurement:    Controller,
		QueryCondition: in.GetQueryCondition(),
	}

	influxdbStatement.AppendWhereCondition(EntityInfluxPlanning.ControllerNamespace, "=", namespace)
	influxdbStatement.AppendWhereCondition(EntityInfluxPlanning.ControllerName, "=", name)
	influxdbStatement.AppendTimeConditionFromQueryCondition()
	influxdbStatement.AppendLimitClauseFromQueryCondition()
	influxdbStatement.AppendOrderClauseFromQueryCondition()

	if ctlPlanningType != DatahubV1alpha1.ControllerPlanningType_CPT_UNDEFINED {
		influxdbStatement.AppendWhereCondition(EntityInfluxPlanning.ControllerType, "=", ctlPlanningType.String())
	}

	cmd := influxdbStatement.BuildQueryCmd()

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return make([]*DatahubV1alpha1.ControllerPlanning, 0), err
	}

	influxdbRows := RepoInflux.PackMap(results)
	recommendations := c.getControllersPlanningsFromInfluxRows(influxdbRows)

	return recommendations, nil
}

func (c *ControllerRepository) getControllersPlanningsFromInfluxRows(rows []*RepoInflux.InfluxDBRow) []*DatahubV1alpha1.ControllerPlanning {
	plannings := make([]*DatahubV1alpha1.ControllerPlanning, 0)

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

			var ctlPlanningType DatahubV1alpha1.ControllerPlanningType
			if tempType, exist := data[EntityInfluxPlanning.ControllerType]; exist {
				if value, ok := DatahubV1alpha1.ControllerPlanningType_value[tempType]; ok {
					ctlPlanningType = DatahubV1alpha1.ControllerPlanningType(value)
				}
			}

			var planningKind DatahubV1alpha1.Kind
			if tempKind, exist := data[EntityInfluxPlanning.ControllerKind]; exist {
				if value, ok := DatahubV1alpha1.Kind_value[tempKind]; ok {
					planningKind = DatahubV1alpha1.Kind(value)
				}
			}

			if ctlPlanningType == DatahubV1alpha1.ControllerPlanningType_CPT_PRIMITIVE {
				tempPlanning := &DatahubV1alpha1.ControllerPlanning{
					PlanningType:    DatahubV1alpha1.PlanningType(DatahubV1alpha1.PlanningType_value[data[string(EntityInfluxPlanning.ControllerPlanningType)]]),
					CtlPlanningType: ctlPlanningType,
					CtlPlanningSpec: &DatahubV1alpha1.ControllerPlanningSpec{
						NamespacedName: &DatahubV1alpha1.NamespacedName{
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

			} else if ctlPlanningType == DatahubV1alpha1.ControllerPlanningType_CPT_K8S {
				tempPlanning := &DatahubV1alpha1.ControllerPlanning{
					PlanningType:    DatahubV1alpha1.PlanningType(DatahubV1alpha1.PlanningType_value[data[string(EntityInfluxPlanning.ControllerPlanningType)]]),
					CtlPlanningType: ctlPlanningType,
					CtlPlanningSpecK8S: &DatahubV1alpha1.ControllerPlanningSpecK8S{
						NamespacedName: &DatahubV1alpha1.NamespacedName{
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
