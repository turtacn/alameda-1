package plannings

import (
	EntityInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/plannings"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	//ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	//"github.com/golang/protobuf/ptypes"
	//"github.com/golang/protobuf/ptypes/timestamp"
	//InfluxClient "github.com/influxdata/influxdb/client/v2"
	//"strconv"
	//"time"
)

type NamespaceRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNamespaceRepository(influxDBCfg *InternalInflux.Config) *NamespaceRepository {
	return &NamespaceRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *NamespaceRepository) CreatePlannings(plannings []*ApiPlannings.NamespacePlanning) error {
	/*
		points := make([]*InfluxClient.Point, 0)

		for _, planning := range plannings {
			namespacePlanningType := planning.GetNamespacePlanningType()

			if namespacePlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
				planningSpec := planning.GetNamespacePlanningSpec()

				tags := map[string]string{
					EntityInfluxPlanning.NamespacePlanningType: planning.GetPlanningType().String(),
					EntityInfluxPlanning.NamespaceName:         planning.GetObjectMeta().GetName(),
					EntityInfluxPlanning.NamespaceType:         ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE.String(),
				}

				fields := map[string]interface{}{
					EntityInfluxPlanning.NamespaceCurrentReplicas: planningSpec.GetCurrentReplicas(),
					EntityInfluxPlanning.NamespaceDesiredReplicas: planningSpec.GetDesiredReplicas(),
					EntityInfluxPlanning.NamespaceCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
					EntityInfluxPlanning.NamespaceKind:            planning.GetKind().String(),

					EntityInfluxPlanning.NamespaceCurrentCPURequest: planningSpec.GetCurrentCpuRequests(),
					EntityInfluxPlanning.NamespaceCurrentMEMRequest: planningSpec.GetCurrentMemRequests(),
					EntityInfluxPlanning.NamespaceCurrentCPULimit:   planningSpec.GetCurrentCpuLimits(),
					EntityInfluxPlanning.NamespaceCurrentMEMLimit:   planningSpec.GetCurrentMemLimits(),
					EntityInfluxPlanning.NamespaceDesiredCPULimit:   planningSpec.GetDesiredCpuLimits(),
					EntityInfluxPlanning.NamespaceDesiredMEMLimit:   planningSpec.GetDesiredMemLimits(),
					EntityInfluxPlanning.NamespaceTotalCost:         planningSpec.GetTotalCost(),
				}

				pt, err := InfluxClient.NewPoint(string(Namespace), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
				if err != nil {
					scope.Error(err.Error())
				}

				points = append(points, pt)

			} else if namespacePlanningType == ApiPlannings.ControllerPlanningType_CPT_K8S {
				planningSpec := planning.GetNamespacePlanningSpecK8S()

				tags := map[string]string{
					EntityInfluxPlanning.NamespacePlanningType: planning.GetPlanningType().String(),
					EntityInfluxPlanning.NamespaceName:         planning.GetObjectMeta().GetName(),
					EntityInfluxPlanning.NamespaceType:         ApiPlannings.ControllerPlanningType_CPT_K8S.String(),
				}

				fields := map[string]interface{}{
					EntityInfluxPlanning.NamespaceCurrentReplicas: planningSpec.GetCurrentReplicas(),
					EntityInfluxPlanning.NamespaceDesiredReplicas: planningSpec.GetDesiredReplicas(),
					EntityInfluxPlanning.NamespaceCreateTime:      planningSpec.GetCreateTime().GetSeconds(),
					EntityInfluxPlanning.NamespaceKind:            planning.GetKind().String(),
				}

				pt, err := InfluxClient.NewPoint(string(Namespace), tags, fields, time.Unix(planningSpec.GetTime().GetSeconds(), 0))
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
	*/

	return nil
}

func (c *NamespaceRepository) ListPlannings(in *ApiPlannings.ListNamespacePlanningsRequest) ([]*ApiPlannings.NamespacePlanning, error) {
	influxdbStatement := InternalInflux.Statement{
		Measurement:    Namespace,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
	}

	planningType := in.GetPlanningType().String()
	ctlPlanningType := in.GetCtlPlanningType().String()
	kind := in.GetKind().String()

	for _, objMeta := range in.GetObjectMeta() {
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxPlanning.NamespaceName,
			EntityInfluxPlanning.NamespaceKind,
		}
		valueList := []string{name, kind}

		if ctlPlanningType != ApiPlannings.ControllerPlanningType_CPT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.NamespaceType)
			valueList = append(valueList, ctlPlanningType)
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.NamespacePlanningType)
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
		return make([]*ApiPlannings.NamespacePlanning, 0), err
	}

	influxdbRows := InternalInflux.PackMap(results)
	plannings := c.getPlanningsFromInfluxRows(influxdbRows)

	return plannings, nil
}

func (c *NamespaceRepository) getPlanningsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiPlannings.NamespacePlanning {
	plannings := make([]*ApiPlannings.NamespacePlanning, 0)

	/*
		for _, influxdbRow := range rows {
			for _, data := range influxdbRow.Data {
				currentReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.NamespaceCurrentReplicas], 10, 64)
				desiredReplicas, _ := strconv.ParseInt(data[EntityInfluxPlanning.NamespaceDesiredReplicas], 10, 64)
				createTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.NamespaceCreateTime], 10, 64)

				t, _ := time.Parse(time.RFC3339, data[EntityInfluxPlanning.NamespaceTime])
				tempTime, _ := ptypes.TimestampProto(t)

				currentCpuRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceCurrentCPURequest], 64)
				currentMemRequests, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceCurrentMEMRequest], 64)
				currentCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceCurrentCPULimit], 64)
				currentMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceCurrentMEMLimit], 64)
				desiredCpuLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceDesiredCPULimit], 64)
				desiredMemLimits, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceDesiredMEMLimit], 64)
				totalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceTotalCost], 64)

				var ctlPlanningType ApiPlannings.ControllerPlanningType
				if tempType, exist := data[EntityInfluxPlanning.NamespaceType]; exist {
					if value, ok := ApiPlannings.ControllerPlanningType_value[tempType]; ok {
						ctlPlanningType = ApiPlannings.ControllerPlanningType(value)
					}
				}

				var planningKind ApiResources.Kind
				if tempKind, exist := data[EntityInfluxPlanning.NamespaceKind]; exist {
					if value, ok := ApiResources.Kind_value[tempKind]; ok {
						planningKind = ApiResources.Kind(value)
					}
				}

				if ctlPlanningType == ApiPlannings.ControllerPlanningType_CPT_PRIMITIVE {
					tempPlanning := &ApiPlannings.NamespacePlanning{
						ObjectMeta: &ApiResources.ObjectMeta{
							Name: data[string(EntityInfluxPlanning.NamespaceName)],
						},
						Kind:                  planningKind,
						PlanningType:          ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.NamespacePlanningType)]]),
						NamespacePlanningType: ctlPlanningType,
						NamespacePlanningSpec: &ApiPlannings.ControllerPlanningSpec{
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
					tempPlanning := &ApiPlannings.NamespacePlanning{
						ObjectMeta: &ApiResources.ObjectMeta{
							Name: data[string(EntityInfluxPlanning.NamespaceName)],
						},
						Kind:                  planningKind,
						PlanningType:          ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[string(EntityInfluxPlanning.NamespacePlanningType)]]),
						NamespacePlanningType: ctlPlanningType,
						NamespacePlanningSpecK8S: &ApiPlannings.ControllerPlanningSpecK8S{
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
	*/

	return plannings
}
