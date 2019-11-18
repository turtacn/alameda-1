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

func (c *ClusterRepository) CreatePlannings(in *ApiPlannings.CreateClusterPlanningsRequest) error {
	clusterPlannings := in.GetClusterPlannings()
	granularity := in.GetGranularity()
	if granularity == 0 {
		granularity = 30
	}

	points := make([]*InfluxClient.Point, 0)
	for _, clusterPlanning := range clusterPlannings {
		if clusterPlanning.GetApplyPlanningNow() {
			//TODO
		}

		planningId := clusterPlanning.GetPlanningId()
		planningType := clusterPlanning.GetPlanningType().String()
		name := clusterPlanning.GetObjectMeta().GetName()
		totalCost := clusterPlanning.GetTotalCost()
		applyPlanningNow := clusterPlanning.GetApplyPlanningNow()

		plannings := clusterPlanning.GetPlannings()
		for _, planning := range plannings {
			tags := map[string]string{
				EntityInfluxPlanning.ClusterPlanningId:   planningId,
				EntityInfluxPlanning.ClusterPlanningType: planningType,
				EntityInfluxPlanning.ClusterName:         name,
				EntityInfluxPlanning.ClusterGranularity:  strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				EntityInfluxPlanning.ClusterTotalCost:        totalCost,
				EntityInfluxPlanning.ClusterApplyPlanningNow: applyPlanningNow,
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
						newFields[EntityInfluxPlanning.ClusterStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.ClusterEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ClusterResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ClusterInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ClusterInitialResourceLimitCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ClusterResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.ClusterInitialResourceLimitMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.ClusterInitialResourceLimitMemory] = float64(0)
							}
						}

						if pt, err := InfluxClient.NewPoint(string(Cluster), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
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
						newFields[EntityInfluxPlanning.ClusterStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.ClusterEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ClusterResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ClusterInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ClusterInitialResourceRequestCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ClusterResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.ClusterInitialResourceRequestMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.ClusterInitialResourceRequestMemory] = float64(0)
							}
						}
						if pt, err := InfluxClient.NewPoint(string(Cluster),
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

func (c *ClusterRepository) ListPlannings(in *ApiPlannings.ListClusterPlanningsRequest) ([]*ApiPlannings.ClusterPlanning, error) {
	plannings := make([]*ApiPlannings.ClusterPlanning, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Cluster,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxPlanning.ClusterName},
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
			EntityInfluxPlanning.ClusterName,
			EntityInfluxPlanning.ClusterGranularity,
		}
		valueList := []string{name, strconv.FormatInt(granularity, 10)}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ClusterPlanningType)
			valueList = append(valueList, planningType)
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

func (c *ClusterRepository) queryPlannings(cmd string, granularity int64) ([]*ApiPlannings.ClusterPlanning, error) {
	ret := make([]*ApiPlannings.ClusterPlanning, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return ret, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			clusterPlanning := &ApiPlannings.ClusterPlanning{}
			clusterPlanning.PlanningId = data[EntityInfluxPlanning.ClusterPlanningId]
			clusterPlanning.ObjectMeta = &ApiResources.ObjectMeta{
				Name: data[EntityInfluxPlanning.ClusterName],
			}

			var planningType ApiPlannings.PlanningType
			if tempPlanningType, exist := data[EntityInfluxPlanning.ClusterPlanningType]; exist {
				if value, ok := ApiPlannings.PlanningType_value[tempPlanningType]; ok {
					planningType = ApiPlannings.PlanningType(value)
				}
			}
			clusterPlanning.PlanningType = planningType

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ClusterTotalCost], 64)
			clusterPlanning.TotalCost = tempTotalCost

			tempApplyPlanningNow, _ := strconv.ParseBool(data[EntityInfluxPlanning.ClusterApplyPlanningNow])
			clusterPlanning.ApplyPlanningNow = tempApplyPlanningNow

			startTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ClusterStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ClusterEndTime], 10, 64)

			clusterPlanning.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			clusterPlanning.EndTime = &timestamp.Timestamp{
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

			tempPlanning.LimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ClusterResourceLimitCPU]
			tempPlanning.LimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ClusterResourceLimitMemory]

			tempPlanning.RequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ClusterResourceRequestCPU]
			tempPlanning.RequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ClusterResourceRequestMemory]

			tempPlanning.InitialLimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ClusterInitialResourceLimitCPU]
			tempPlanning.InitialLimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ClusterInitialResourceLimitMemory]

			tempPlanning.InitialRequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ClusterInitialResourceRequestCPU]
			tempPlanning.InitialRequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ClusterInitialResourceRequestMemory]

			clusterPlanning.Plannings = append(clusterPlanning.Plannings, tempPlanning)

			ret = append(ret, clusterPlanning)
		}
	}

	return ret, nil
}
