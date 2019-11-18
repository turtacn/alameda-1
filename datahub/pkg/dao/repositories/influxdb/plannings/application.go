package plannings

import (
	"math"
	"strconv"
	"time"

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

func (c *AppRepository) CreatePlannings(in *ApiPlannings.CreateApplicationPlanningsRequest) error {
	appPlannings := in.GetApplicationPlannings()
	granularity := in.GetGranularity()
	if granularity == 0 {
		granularity = 30
	}

	points := make([]*InfluxClient.Point, 0)
	for _, appPlanning := range appPlannings {
		if appPlanning.GetApplyPlanningNow() {
			//TODO
		}

		planningId := appPlanning.GetPlanningId()
		planningType := appPlanning.GetPlanningType().String()
		namespace := appPlanning.GetObjectMeta().GetNamespace()
		name := appPlanning.GetObjectMeta().GetName()
		totalCost := appPlanning.GetTotalCost()
		applyPlanningNow := appPlanning.GetApplyPlanningNow()

		plannings := appPlanning.GetPlannings()
		for _, planning := range plannings {
			tags := map[string]string{
				EntityInfluxPlanning.AppPlanningId:   planningId,
				EntityInfluxPlanning.AppPlanningType: planningType,
				EntityInfluxPlanning.AppNamespace:    namespace,
				EntityInfluxPlanning.AppName:         name,
				EntityInfluxPlanning.AppGranularity:  strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				EntityInfluxPlanning.AppTotalCost:        totalCost,
				EntityInfluxPlanning.AppApplyPlanningNow: applyPlanningNow,
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
						newFields[EntityInfluxPlanning.AppStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.AppEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.AppResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.AppInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.AppInitialResourceLimitCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.AppResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.AppInitialResourceLimitMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.AppInitialResourceLimitMemory] = float64(0)
							}
						}

						if pt, err := InfluxClient.NewPoint(string(Application), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
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
						newFields[EntityInfluxPlanning.AppStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.AppEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.AppResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.AppInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.AppInitialResourceRequestCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.AppResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.AppInitialResourceRequestMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.AppInitialResourceRequestMemory] = float64(0)
							}
						}
						if pt, err := InfluxClient.NewPoint(string(Application),
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

func (c *AppRepository) ListPlannings(in *ApiPlannings.ListApplicationPlanningsRequest) ([]*ApiPlannings.ApplicationPlanning, error) {
	plannings := make([]*ApiPlannings.ApplicationPlanning, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Application,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxPlanning.AppNamespace, EntityInfluxPlanning.AppName},
	}

	planningType := in.GetPlanningType().String()
	granularity := in.GetGranularity()

	if granularity == 0 {
		granularity = 30
	}

	for _, objMeta := range in.GetObjectMeta() {
		tempCondition := ""
		namespace := objMeta.GetNamespace()
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxPlanning.AppNamespace,
			EntityInfluxPlanning.AppName,
			EntityInfluxPlanning.AppGranularity,
		}
		valueList := []string{namespace, name, strconv.FormatInt(granularity, 10)}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.AppPlanningType)
			valueList = append(valueList, planningType)
		}

		tempCondition = influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	if influxdbStatement.WhereClause == "" {
		tempCondition := ""

		keyList := []string{
			EntityInfluxPlanning.AppGranularity,
		}
		valueList := []string{
			strconv.FormatInt(granularity, 10),
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.AppPlanningType)
			valueList = append(valueList, planningType)
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

func (c *AppRepository) queryPlannings(cmd string, granularity int64) ([]*ApiPlannings.ApplicationPlanning, error) {
	ret := make([]*ApiPlannings.ApplicationPlanning, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return ret, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			appPlanning := &ApiPlannings.ApplicationPlanning{}
			appPlanning.PlanningId = data[EntityInfluxPlanning.AppPlanningId]
			appPlanning.ObjectMeta = &ApiResources.ObjectMeta{
				Namespace: data[EntityInfluxPlanning.AppNamespace],
				Name:      data[EntityInfluxPlanning.AppName],
			}

			var planningType ApiPlannings.PlanningType
			if tempPlanningType, exist := data[EntityInfluxPlanning.AppPlanningType]; exist {
				if value, ok := ApiPlannings.PlanningType_value[tempPlanningType]; ok {
					planningType = ApiPlannings.PlanningType(value)
				}
			}
			appPlanning.PlanningType = planningType

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.AppTotalCost], 64)
			appPlanning.TotalCost = tempTotalCost

			tempApplyPlanningNow, _ := strconv.ParseBool(data[EntityInfluxPlanning.AppApplyPlanningNow])
			appPlanning.ApplyPlanningNow = tempApplyPlanningNow

			startTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.AppStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.AppEndTime], 10, 64)

			appPlanning.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			appPlanning.EndTime = &timestamp.Timestamp{
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

			tempPlanning.LimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.AppResourceLimitCPU]
			tempPlanning.LimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.AppResourceLimitMemory]

			tempPlanning.RequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.AppResourceRequestCPU]
			tempPlanning.RequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.AppResourceRequestMemory]

			tempPlanning.InitialLimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.AppInitialResourceLimitCPU]
			tempPlanning.InitialLimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.AppInitialResourceLimitMemory]

			tempPlanning.InitialRequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.AppInitialResourceRequestCPU]
			tempPlanning.InitialRequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.AppInitialResourceRequestMemory]

			appPlanning.Plannings = append(appPlanning.Plannings, tempPlanning)

			ret = append(ret, appPlanning)
		}
	}

	return ret, nil
}
