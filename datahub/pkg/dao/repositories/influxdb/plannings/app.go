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

func (c *AppRepository) CreatePlannings(plannings []*ApiPlannings.ApplicationPlanning) error {
	points := make([]*InfluxClient.Point, 0)

	for _, planning := range plannings {
		appPlanningType := planning.GetApplicationPlanningType()

		if appPlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
			planningSpec := planning.GetApplicationPlanningSpec()

			tags := map[string]string{
				EntityInfluxPlanning.AppPlanningType: planning.GetPlanningType().String(),
				EntityInfluxPlanning.AppNamespace:    planning.GetObjectMeta().GetNamespace(),
				EntityInfluxPlanning.AppName:         planning.GetObjectMeta().GetName(),
				EntityInfluxPlanning.AppType:         ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxPlanning.AppCurrentReplicas: planningSpec.GetCurrentReplicas(),
				EntityInfluxPlanning.AppDesiredReplicas: planningSpec.GetDesiredReplicas(),
				EntityInfluxPlanning.AppCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
				EntityInfluxPlanning.AppKind:            planning.GetKind().String(),

				EntityInfluxPlanning.AppCurrentCPURequest: planningSpec.GetCurrentCpuRequests(),
				EntityInfluxPlanning.AppCurrentMEMRequest: planningSpec.GetCurrentMemRequests(),
				EntityInfluxPlanning.AppCurrentCPULimit:   planningSpec.GetCurrentCpuLimits(),
				EntityInfluxPlanning.AppCurrentMEMLimit:   planningSpec.GetCurrentMemLimits(),
				EntityInfluxPlanning.AppDesiredCPULimit:   planningSpec.GetDesiredCpuLimits(),
				EntityInfluxPlanning.AppDesiredMEMLimit:   planningSpec.GetDesiredMemLimits(),
				EntityInfluxPlanning.AppTotalCost:         planningSpec.GetTotalCost(),
			}

			pt, err := InfluxClient.NewPoint(string(App), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
			if err != nil {
				scope.Error(err.Error())
			}

			points = append(points, pt)

		} else if appPlanningType == ApiPlannings.ControllerPlanningType_CPT_K8S {
			planningSpec := planning.GetApplicationPlanningSpecK8S()

			tags := map[string]string{
				EntityInfluxPlanning.AppPlanningType: planning.GetPlanningType().String(),
				EntityInfluxPlanning.AppNamespace:    planning.GetObjectMeta().GetNamespace(),
				EntityInfluxPlanning.AppName:         planning.GetObjectMeta().GetName(),
				EntityInfluxPlanning.AppType:         ApiPlannings.ControllerPlanningType_CPT_K8S.String(),
			}

			fields := map[string]interface{}{
				EntityInfluxPlanning.AppCurrentReplicas: planningSpec.GetCurrentReplicas(),
				EntityInfluxPlanning.AppDesiredReplicas: planningSpec.GetDesiredReplicas(),
				EntityInfluxPlanning.AppCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
				EntityInfluxPlanning.AppKind:            planning.GetKind().String(),
			}

			pt, err := InfluxClient.NewPoint(string(App), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
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

func (c *AppRepository) ListPlannings(in *ApiPlannings.ListApplicationPlanningsRequest) ([]*ApiPlannings.ApplicationPlanning, error) {
	influxdbStatement := InternalInflux.Statement{
		Measurement:    App,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	planningType := in.GetPlanningType().String()
	ctlPlanningType := in.GetCtlPlanningType().String()
	kind := in.GetKind().String()

	for _, objMeta := range in.GetObjectMeta() {
		namespace := objMeta.GetNamespace()
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxPlanning.AppNamespace,
			EntityInfluxPlanning.AppName,
			EntityInfluxPlanning.AppKind,
		}
		valueList := []string{namespace, name, kind}

		if ctlPlanningType != ApiPlannings.ControllerPlanningType_CPT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.AppType)
			valueList = append(valueList, ctlPlanningType)
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.AppPlanningType)
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
		return make([]*ApiPlannings.ApplicationPlanning, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	plannings := c.getPlanningsFromInfluxRows(influxdbRows)

	return plannings, nil
}

func (c *AppRepository) getPlanningsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiPlannings.ApplicationPlanning {
	plannings := make([]*ApiPlannings.ApplicationPlanning, 0)

	for _, influxdbRow := range rows {
		for _, data := range influxdbRow.Data {
			currentReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.AppCurrentReplicas], 10, 64)
			desiredReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.AppDesiredReplicas], 10, 64)
			createTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.AppCreateTime], 10, 64)

			t, _ := time.Parse(time.RFC3339, data[EntityInfluxPlanning.AppTime])
			tempTime, _ := ptypes.TimestampProto(t)

			currentCpuRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppCurrentCPURequest], 64)
			currentMemRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppCurrentMEMRequest], 64)
			currentCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppCurrentCPULimit], 64)
			currentMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppCurrentMEMLimit], 64)
			desiredCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppDesiredCPULimit], 64)
			desiredMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppDesiredMEMLimit], 64)
			totalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppTotalCost], 64)

			var ctlPlanningType ApiPlannings.ControllerPlanningType
			if tempType, exist := data[EntityInfluxPlanning.AppType]; exist {
				if value, ok := ApiPlannings.ControllerPlanningType_value[tempType]; ok {
					ctlPlanningType = ApiPlannings.ControllerPlanningType(value)
				}
			}

			var planningKind ApiResources.Kind
			if tempKind, exist := data[EntityInfluxPlanning.AppKind]; exist {
				if value, ok := ApiResources.Kind_value[tempKind]; ok {
					planningKind = ApiResources.Kind(value)
				}
			}

			if ctlPlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
				tempPlanning := &ApiPlannings.ApplicationPlanning{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name:      data[string(EntityInfluxPlanning.AppName)],
						Namespace: data[string(EntityInfluxPlanning.AppNamespace)],
					},
					Kind:                    planningKind,
					PlanningType:            ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.AppPlanningType)]]),
					ApplicationPlanningType: ctlPlanningType,
					ApplicationPlanningSpec: &ApiPlannings.ControllerPlanningSpec{
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
				tempPlanning := &ApiPlannings.ApplicationPlanning{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name:      data[string(EntityInfluxPlanning.AppName)],
						Namespace: data[string(EntityInfluxPlanning.AppNamespace)],
					},
					Kind:                    planningKind,
					PlanningType:            ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.AppPlanningType)]]),
					ApplicationPlanningType: ctlPlanningType,
					ApplicationPlanningSpecK8S: &ApiPlannings.ControllerPlanningSpecK8S{
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
