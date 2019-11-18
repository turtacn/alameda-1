package plannings

import (
	"math"
	"strconv"
	"time"

	//"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"

	EntityInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/plannings"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
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

func (c *ControllerRepository) CreateControllerPlannings(in *ApiPlannings.CreateControllerPlanningsRequest) error {
	controllerPlannings := in.GetControllerPlannings()
	granularity := in.GetGranularity()
	if granularity == 0 {
		granularity = 30
	}

	points := make([]*InfluxClient.Point, 0)
	for _, controllerPlanning := range controllerPlannings {
		if controllerPlanning.GetApplyPlanningNow() {
			//TODO
		}

		planningId := controllerPlanning.GetPlanningId()
		planningType := controllerPlanning.GetPlanningType().String()
		namespace := controllerPlanning.GetObjectMeta().GetNamespace()
		name := controllerPlanning.GetObjectMeta().GetName()
		totalCost := controllerPlanning.GetTotalCost()
		applyPlanningNow := controllerPlanning.GetApplyPlanningNow()
		kind := controllerPlanning.GetKind().String()

		plannings := controllerPlanning.GetPlannings()
		for _, planning := range plannings {
			tags := map[string]string{
				EntityInfluxPlanning.ControllerPlanningId:   planningId,
				EntityInfluxPlanning.ControllerPlanningType: planningType,
				EntityInfluxPlanning.ControllerNamespace:    namespace,
				EntityInfluxPlanning.ControllerName:         name,
				EntityInfluxPlanning.ControllerGranularity:  strconv.FormatInt(granularity, 10),
				EntityInfluxPlanning.ControllerKind:         kind,
			}
			fields := map[string]interface{}{
				EntityInfluxPlanning.ControllerTotalCost:        totalCost,
				EntityInfluxPlanning.ControllerApplyPlanningNow: applyPlanningNow,
			}

			initialLimitPlanning := make(map[ApiCommon.MetricType]interface{})
			if planning.GetInitialLimitPlannings() != nil {
				for _, rec := range planning.GetInitialLimitPlannings() {
					// One and only one record in initial limit recommendation
					initialLimitPlanning[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}
			initialRequestPlanning := make(map[ApiCommon.MetricType]interface{})
			if planning.GetInitialRequestPlannings() != nil {
				for _, rec := range planning.GetInitialRequestPlannings() {
					// One and only one record in initial request recommendation
					initialRequestPlanning[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}

			for _, metricData := range planning.GetLimitPlannings() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						newFields[EntityInfluxPlanning.ControllerStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.ControllerEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ControllerResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ControllerInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ControllerInitialResourceLimitCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ControllerResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.ControllerInitialResourceLimitMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.ControllerInitialResourceLimitMemory] = float64(0)
							}
						}

						if pt, err := InfluxClient.NewPoint(string(Controller), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
							points = append(points, pt)
						} else {
							scope.Error(err.Error())
						}
					}
				}
			}

			for _, metricData := range planning.GetRequestPlannings() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						newFields[EntityInfluxPlanning.ControllerStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.ControllerEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ControllerResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ControllerInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ControllerInitialResourceRequestCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ControllerResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.ControllerInitialResourceRequestMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.ControllerInitialResourceRequestMemory] = float64(0)
							}
						}
						if pt, err := InfluxClient.NewPoint(string(Controller),
							tags, newFields,
							time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
							points = append(points, pt)
						} else {
							scope.Error(err.Error())
						}
					}
				}
			}
		}
	}
	err := c.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Planning),
	})

	if err != nil {
		return err
	}
	return nil
}

func (c *ControllerRepository) ListControllerPlannings(in *ApiPlannings.ListControllerPlanningsRequest) ([]*ApiPlannings.ControllerPlanning, error) {
	plannings := make([]*ApiPlannings.ControllerPlanning, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Controller,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxPlanning.ControllerName},
	}

	planningType := in.GetPlanningType().String()
	granularity := in.GetGranularity()
	kind := in.GetKind().String()

	if granularity == 0 {
		granularity = 30
	}

	for _, objMeta := range in.GetObjectMeta() {
		tempCondition := ""
		namespace := objMeta.GetNamespace()
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxPlanning.ControllerNamespace,
			EntityInfluxPlanning.ClusterName,
			EntityInfluxPlanning.ClusterGranularity,
		}
		valueList := []string{namespace, name, strconv.FormatInt(granularity, 10)}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ClusterPlanningType)
			valueList = append(valueList, planningType)
		}

		if kind != ApiResources.Kind_KIND_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ControllerKind)
			valueList = append(valueList, kind)
		}

		tempCondition = influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	if influxdbStatement.WhereClause == "" {
		tempCondition := ""

		keyList := []string{
			EntityInfluxPlanning.ClusterGranularity,
		}
		valueList := []string{
			strconv.FormatInt(granularity, 10),
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ClusterPlanningType)
			valueList = append(valueList, planningType)
		}

		if kind != ApiResources.Kind_KIND_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ControllerKind)
			valueList = append(valueList, kind)
		}

		tempCondition = influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	plannings, err := c.queryPlannings(cmd, granularity)
	if err != nil {
		return plannings, err
	}

	return plannings, nil
}

func (c *ControllerRepository) queryPlannings(cmd string, granularity int64) ([]*ApiPlannings.ControllerPlanning, error) {
	ret := make([]*ApiPlannings.ControllerPlanning, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return ret, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			controllerPlanning := &ApiPlannings.ControllerPlanning{}
			controllerPlanning.PlanningId = data[EntityInfluxPlanning.ControllerPlanningId]
			controllerPlanning.ObjectMeta = &ApiResources.ObjectMeta{
				Namespace: data[EntityInfluxPlanning.ControllerNamespace],
				Name:      data[EntityInfluxPlanning.ControllerName],
			}

			var kind ApiResources.Kind
			if tempKind, exist := data[EntityInfluxPlanning.ControllerKind]; exist {
				if value, ok := ApiResources.Kind_value[tempKind]; ok {
					kind = ApiResources.Kind(value)
				}
			}
			controllerPlanning.Kind = kind

			var planningType ApiPlannings.PlanningType
			if tempPlanningType, exist := data[EntityInfluxPlanning.ControllerPlanningType]; exist {
				if value, ok := ApiPlannings.PlanningType_value[tempPlanningType]; ok {
					planningType = ApiPlannings.PlanningType(value)
				}
			}
			controllerPlanning.PlanningType = planningType

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ControllerTotalCost], 64)
			controllerPlanning.TotalCost = tempTotalCost

			tempApplyPlanningNow, _ := strconv.ParseBool(data[EntityInfluxPlanning.ControllerApplyPlanningNow])
			controllerPlanning.ApplyPlanningNow = tempApplyPlanningNow

			startTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ControllerStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ControllerEndTime], 10, 64)

			controllerPlanning.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			controllerPlanning.EndTime = &timestamp.Timestamp{
				Seconds: endTime,
			}

			//
			tempPlanning := &ApiPlannings.Planning{}

			metricTypeList := []ApiCommon.MetricType{ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, ApiCommon.MetricType_MEMORY_USAGE_BYTES}
			sampleTime := &timestamp.Timestamp{
				Seconds: startTime,
			}
			sampleEndTime := &timestamp.Timestamp{
				Seconds: endTime,
			}

			//
			for _, metricType := range metricTypeList {
				metricDataList := make([]*ApiCommon.MetricData, 0)
				for a := 0; a < 4; a++ {
					sample := &ApiCommon.Sample{
						Time:    sampleTime,
						EndTime: sampleEndTime,
					}

					metricData := &ApiCommon.MetricData{
						MetricType:  metricType,
						Granularity: granularity,
					}
					metricData.Data = append(metricData.Data, sample)
					metricDataList = append(metricDataList, metricData)
				}

				tempPlanning.LimitPlannings = append(tempPlanning.LimitPlannings, metricDataList[0])
				tempPlanning.RequestPlannings = append(tempPlanning.RequestPlannings, metricDataList[1])
				tempPlanning.InitialLimitPlannings = append(tempPlanning.InitialLimitPlannings, metricDataList[2])
				tempPlanning.InitialRequestPlannings = append(tempPlanning.InitialRequestPlannings, metricDataList[3])
			}

			tempPlanning.LimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ControllerResourceLimitCPU]
			tempPlanning.LimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ControllerResourceLimitMemory]

			tempPlanning.RequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ControllerResourceRequestCPU]
			tempPlanning.RequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ControllerResourceRequestMemory]

			tempPlanning.InitialLimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ControllerInitialResourceLimitCPU]
			tempPlanning.InitialLimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ControllerInitialResourceLimitMemory]

			tempPlanning.InitialRequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ControllerInitialResourceRequestCPU]
			tempPlanning.InitialRequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ControllerInitialResourceRequestMemory]

			controllerPlanning.Plannings = append(controllerPlanning.Plannings, tempPlanning)

			ret = append(ret, controllerPlanning)
		}
	}

	return ret, nil
}
