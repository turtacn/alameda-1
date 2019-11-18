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

func (c *NamespaceRepository) CreatePlannings(in *ApiPlannings.CreateNamespacePlanningsRequest) error {
	namesapcePlannings := in.GetNamespacePlannings()
	granularity := in.GetGranularity()
	if granularity == 0 {
		granularity = 30
	}

	points := make([]*InfluxClient.Point, 0)
	for _, namespacePlanning := range namesapcePlannings {
		if namespacePlanning.GetApplyPlanningNow() {
			//TODO
		}

		planningId := namespacePlanning.GetPlanningId()
		planningType := namespacePlanning.GetPlanningType().String()
		name := namespacePlanning.GetObjectMeta().GetName()
		totalCost := namespacePlanning.GetTotalCost()
		applyPlanningNow := namespacePlanning.GetApplyPlanningNow()

		plannings := namespacePlanning.GetPlannings()
		for _, planning := range plannings {
			tags := map[string]string{
				EntityInfluxPlanning.NamespacePlanningId:   planningId,
				EntityInfluxPlanning.NamespacePlanningType: planningType,
				EntityInfluxPlanning.NamespaceName:         name,
				EntityInfluxPlanning.NamespaceGranularity:  strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				EntityInfluxPlanning.NamespaceTotalCost:        totalCost,
				EntityInfluxPlanning.NamespaceApplyPlanningNow: applyPlanningNow,
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
						newFields[EntityInfluxPlanning.NamespaceStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.NamespaceEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.NamespaceResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.NamespaceInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.NamespaceInitialResourceLimitCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.NamespaceResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.NamespaceInitialResourceLimitMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.NamespaceInitialResourceLimitMemory] = float64(0)
							}
						}

						if pt, err := InfluxClient.NewPoint(string(Namespace), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
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
						newFields[EntityInfluxPlanning.NamespaceStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.NamespaceEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.NamespaceResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.NamespaceInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.NamespaceInitialResourceRequestCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.NamespaceResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.NamespaceInitialResourceRequestMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.NamespaceInitialResourceRequestMemory] = float64(0)
							}
						}
						if pt, err := InfluxClient.NewPoint(string(Namespace),
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

func (c *NamespaceRepository) ListPlannings(in *ApiPlannings.ListNamespacePlanningsRequest) ([]*ApiPlannings.NamespacePlanning, error) {
	plannings := make([]*ApiPlannings.NamespacePlanning, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Namespace,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxPlanning.NamespaceName},
	}

	planningType := in.GetPlanningType().String()
	granularity := in.GetGranularity()

	if granularity == 0 {
		granularity = 30
	}

	for _, objMeta := range in.GetObjectMeta() {
		tempCondition := ""
		name := objMeta.GetName()

		keyList := []string{
			EntityInfluxPlanning.NamespaceName,
			EntityInfluxPlanning.NamespaceGranularity,
		}
		valueList := []string{name, strconv.FormatInt(granularity, 10)}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.NamespacePlanningType)
			valueList = append(valueList, planningType)
		}

		tempCondition = influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	if influxdbStatement.WhereClause == "" {
		tempCondition := ""

		keyList := []string{
			EntityInfluxPlanning.NamespaceGranularity,
		}
		valueList := []string{
			strconv.FormatInt(granularity, 10),
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.NamespacePlanningType)
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

func (c *NamespaceRepository) queryPlannings(cmd string, granularity int64) ([]*ApiPlannings.NamespacePlanning, error) {
	ret := make([]*ApiPlannings.NamespacePlanning, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return ret, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			namespacePlanning := &ApiPlannings.NamespacePlanning{}
			namespacePlanning.PlanningId = data[EntityInfluxPlanning.NamespacePlanningId]
			namespacePlanning.ObjectMeta = &ApiResources.ObjectMeta{
				Name: data[EntityInfluxPlanning.NamespaceName],
			}

			var planningType ApiPlannings.PlanningType
			if tempPlanningType, exist := data[EntityInfluxPlanning.NamespacePlanningType]; exist {
				if value, ok := ApiPlannings.PlanningType_value[tempPlanningType]; ok {
					planningType = ApiPlannings.PlanningType(value)
				}
			}
			namespacePlanning.PlanningType = planningType

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NamespaceTotalCost], 64)
			namespacePlanning.TotalCost = tempTotalCost

			tempApplyPlanningNow, _ := strconv.ParseBool(data[EntityInfluxPlanning.NamespaceApplyPlanningNow])
			namespacePlanning.ApplyPlanningNow = tempApplyPlanningNow

			startTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.NamespaceStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.NamespaceEndTime], 10, 64)

			namespacePlanning.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			namespacePlanning.EndTime = &timestamp.Timestamp{
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

			tempPlanning.LimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceResourceLimitCPU]
			tempPlanning.LimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceResourceLimitMemory]

			tempPlanning.RequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceResourceRequestCPU]
			tempPlanning.RequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceResourceRequestMemory]

			tempPlanning.InitialLimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceInitialResourceLimitCPU]
			tempPlanning.InitialLimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceInitialResourceLimitMemory]

			tempPlanning.InitialRequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceInitialResourceRequestCPU]
			tempPlanning.InitialRequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NamespaceInitialResourceRequestMemory]

			namespacePlanning.Plannings = append(namespacePlanning.Plannings, tempPlanning)

			ret = append(ret, namespacePlanning)
		}
	}

	return ret, nil
}
