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

func (c *NodeRepository) CreatePlannings(in *ApiPlannings.CreateNodePlanningsRequest) error {
	nodePlannings := in.GetNodePlannings()
	granularity := in.GetGranularity()
	if granularity == 0 {
		granularity = 30
	}

	points := make([]*InfluxClient.Point, 0)
	for _, nodePlanning := range nodePlannings {
		if nodePlanning.GetApplyPlanningNow() {
			//TODO
		}

		planningId := nodePlanning.GetPlanningId()
		planningType := nodePlanning.GetPlanningType().String()
		name := nodePlanning.GetObjectMeta().GetName()
		totalCost := nodePlanning.GetTotalCost()
		applyPlanningNow := nodePlanning.GetApplyPlanningNow()

		plannings := nodePlanning.GetPlannings()
		for _, planning := range plannings {
			tags := map[string]string{
				EntityInfluxPlanning.NodePlanningId:   planningId,
				EntityInfluxPlanning.NodePlanningType: planningType,
				EntityInfluxPlanning.NodeName:         name,
				EntityInfluxPlanning.NodeGranularity:  strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				EntityInfluxPlanning.NodeTotalCost:        totalCost,
				EntityInfluxPlanning.NodeApplyPlanningNow: applyPlanningNow,
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
						newFields[EntityInfluxPlanning.NodeStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.NodeEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.NodeResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.NodeInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.NodeInitialResourceLimitCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.NodeResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.NodeInitialResourceLimitMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.NodeInitialResourceLimitMemory] = float64(0)
							}
						}

						if pt, err := InfluxClient.NewPoint(string(Node), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
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
						newFields[EntityInfluxPlanning.NodeStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.NodeEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.NodeResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.NodeInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.NodeInitialResourceRequestCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.NodeResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.NodeInitialResourceRequestMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.NodeInitialResourceRequestMemory] = float64(0)
							}
						}
						if pt, err := InfluxClient.NewPoint(string(Node),
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

func (c *NodeRepository) ListPlannings(in *ApiPlannings.ListNodePlanningsRequest) ([]*ApiPlannings.NodePlanning, error) {
	plannings := make([]*ApiPlannings.NodePlanning, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Node,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxPlanning.NodeName},
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
			EntityInfluxPlanning.NodeName,
			EntityInfluxPlanning.NodeGranularity,
		}
		valueList := []string{name, strconv.FormatInt(granularity, 10)}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.NodePlanningType)
			valueList = append(valueList, planningType)
		}

		tempCondition = influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	if influxdbStatement.WhereClause == "" {
		tempCondition := ""

		keyList := []string{
			EntityInfluxPlanning.NodeGranularity,
		}
		valueList := []string{
			strconv.FormatInt(granularity, 10),
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.NodePlanningType)
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

func (c *NodeRepository) queryPlannings(cmd string, granularity int64) ([]*ApiPlannings.NodePlanning, error) {
	ret := make([]*ApiPlannings.NodePlanning, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return ret, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			nodePlanning := &ApiPlannings.NodePlanning{}
			nodePlanning.PlanningId = data[EntityInfluxPlanning.NodePlanningId]
			nodePlanning.ObjectMeta = &ApiResources.ObjectMeta{
				Name: data[EntityInfluxPlanning.NodeName],
			}

			var planningType ApiPlannings.PlanningType
			if tempPlanningType, exist := data[EntityInfluxPlanning.NodePlanningType]; exist {
				if value, ok := ApiPlannings.PlanningType_value[tempPlanningType]; ok {
					planningType = ApiPlannings.PlanningType(value)
				}
			}
			nodePlanning.PlanningType = planningType

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.NodeTotalCost], 64)
			nodePlanning.TotalCost = tempTotalCost

			tempApplyPlanningNow, _ := strconv.ParseBool(data[EntityInfluxPlanning.NodeApplyPlanningNow])
			nodePlanning.ApplyPlanningNow = tempApplyPlanningNow

			startTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.NodeStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.NodeEndTime], 10, 64)

			nodePlanning.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			nodePlanning.EndTime = &timestamp.Timestamp{
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

			tempPlanning.LimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NodeResourceLimitCPU]
			tempPlanning.LimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NodeResourceLimitMemory]

			tempPlanning.RequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NodeResourceRequestCPU]
			tempPlanning.RequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NodeResourceRequestMemory]

			tempPlanning.InitialLimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NodeInitialResourceLimitCPU]
			tempPlanning.InitialLimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NodeInitialResourceLimitMemory]

			tempPlanning.InitialRequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.NodeInitialResourceRequestCPU]
			tempPlanning.InitialRequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.NodeInitialResourceRequestMemory]

			nodePlanning.Plannings = append(nodePlanning.Plannings, tempPlanning)

			ret = append(ret, nodePlanning)
		}
	}

	return ret, nil
}
