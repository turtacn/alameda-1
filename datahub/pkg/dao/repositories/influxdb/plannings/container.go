package plannings

import (
	"fmt"
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
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

var (
	scope = Log.RegisterScope("planning_db_measurement", "planning DB measurement", 0)
)

// ContainerRepository is used to operate container measurement of planning database
type ContainerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

// NewContainerRepository creates the ContainerRepository instance
func NewContainerRepository(influxDBCfg *InternalInflux.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

// IsTag checks the column is tag or not
func (c *ContainerRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxPlanning.ContainerTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

// CreateContainerPlannings add containers plannings
func (c *ContainerRepository) CreateContainerPlannings(in *ApiPlannings.CreatePodPlanningsRequest) error {
	podPlannings := in.GetPodPlannings()
	granularity := in.GetGranularity()
	if granularity == 0 {
		granularity = 30
	}

	points := make([]*InfluxClient.Point, 0)
	for _, podPlanning := range podPlannings {
		if podPlanning.GetApplyPlanningNow() {
			//TODO
		}

		podNS := podPlanning.GetObjectMeta().GetNamespace()
		podName := podPlanning.GetObjectMeta().GetName()
		podTotalCost := podPlanning.GetTotalCost()
		containerPlannings := podPlanning.GetContainerPlannings()
		topController := podPlanning.GetTopController()

		podPolicy := podPlanning.GetAssignPodPolicy().GetPolicy()
		podPolicyValue := ""
		switch podPolicy.(type) {
		case *ApiResources.AssignPodPolicy_NodeName:
			podPolicyValue = podPolicy.(*ApiResources.AssignPodPolicy_NodeName).NodeName
		case *ApiResources.AssignPodPolicy_NodePriority:
			nodeList := podPolicy.(*ApiResources.AssignPodPolicy_NodePriority).NodePriority.GetNodes()
			if len(nodeList) > 0 {
				podPolicyValue = nodeList[0]
			}
			podPolicyValue = podPolicy.(*ApiResources.AssignPodPolicy_NodePriority).NodePriority.GetNodes()[0]
		case *ApiResources.AssignPodPolicy_NodeSelector:
			nodeMap := podPolicy.(*ApiResources.AssignPodPolicy_NodeSelector).NodeSelector.Selector
			for _, value := range nodeMap {
				podPolicyValue = value
				break
			}
		}

		for _, containerPlanning := range containerPlannings {
			tags := map[string]string{
				EntityInfluxPlanning.ContainerPlanningType: podPlanning.GetPlanningType().String(),
				EntityInfluxPlanning.ContainerNamespace:    podNS,
				EntityInfluxPlanning.ContainerPodName:      podName,
				EntityInfluxPlanning.ContainerName:         containerPlanning.GetName(),
				EntityInfluxPlanning.ContainerGranularity:  strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				//TODO
				//string(EntityInfluxRecommend.ContainerPolicy):            "",
				EntityInfluxPlanning.ContainerTopControllerName: topController.GetObjectMeta().GetName(),
				EntityInfluxPlanning.ContainerTopControllerKind: topController.GetKind().String(),
				EntityInfluxPlanning.ContainerPolicy:            podPolicyValue,
				EntityInfluxPlanning.ContainerPolicyTime:        podPlanning.GetAssignPodPolicy().GetTime().GetSeconds(),
				EntityInfluxPlanning.ContainerPodTotalCost:      podTotalCost,
			}

			initialLimitPlanning := make(map[ApiCommon.MetricType]interface{})
			if containerPlanning.GetInitialLimitPlannings() != nil {
				for _, rec := range containerPlanning.GetInitialLimitPlannings() {
					// One and only one record in initial limit recommendation
					initialLimitPlanning[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}
			initialRequestPlanning := make(map[ApiCommon.MetricType]interface{})
			if containerPlanning.GetInitialRequestPlannings() != nil {
				for _, rec := range containerPlanning.GetInitialRequestPlannings() {
					// One and only one record in initial request recommendation
					initialRequestPlanning[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}

			for _, metricData := range containerPlanning.GetLimitPlannings() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						newFields[EntityInfluxPlanning.ContainerStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.ContainerEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ContainerResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ContainerInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ContainerInitialResourceLimitCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ContainerResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.ContainerInitialResourceLimitMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.ContainerInitialResourceLimitMemory] = float64(0)
							}
						}

						if pt, err := InfluxClient.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
							points = append(points, pt)
						} else {
							scope.Error(err.Error())
						}
					}
				}
			}

			for _, metricData := range containerPlanning.GetRequestPlannings() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						newFields[EntityInfluxPlanning.ContainerStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxPlanning.ContainerEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ContainerResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ContainerInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ContainerInitialResourceRequestCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ContainerResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestPlanning[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxPlanning.ContainerInitialResourceRequestMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxPlanning.ContainerInitialResourceRequestMemory] = float64(0)
							}
						}
						if pt, err := InfluxClient.NewPoint(string(Container),
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

// ListContainerPlannings list container plannings
func (c *ContainerRepository) ListContainerPlannings(in *ApiPlannings.ListPodPlanningsRequest) ([]*ApiPlannings.PodPlanning, error) {
	podPlannings := make([]*ApiPlannings.PodPlanning, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Container,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxPlanning.ContainerName, EntityInfluxPlanning.ContainerNamespace, EntityInfluxPlanning.ContainerPodName},
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
			EntityInfluxPlanning.ContainerNamespace,
			EntityInfluxPlanning.ContainerPodName,
			EntityInfluxPlanning.ContainerGranularity,
		}
		valueList := []string{
			namespace,
			name,
			strconv.FormatInt(granularity, 10),
		}

		if planningType != ApiPlannings.PlanningType_PT_UNDEFINED.String() {
			keyList = append(keyList, EntityInfluxPlanning.ControllerPlanningType)
			valueList = append(valueList, planningType)
		}

		tempCondition = influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()
	scope.Debugf(fmt.Sprintf("ListContainerPlannings: %s", cmd))

	podPlannings, err := c.queryPlannings(cmd, granularity)
	if err != nil {
		return podPlannings, err
	}

	return podPlannings, nil
}

func (c *ContainerRepository) queryPlannings(cmd string, granularity int64) ([]*ApiPlannings.PodPlanning, error) {
	podPlannings := make([]*ApiPlannings.PodPlanning, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return podPlannings, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			podPlanning := &ApiPlannings.PodPlanning{}
			podPlanning.PlanningType = ApiPlannings.PlanningType(ApiPlannings.PlanningType_value[data[EntityInfluxPlanning.ContainerPlanningType]])
			podPlanning.ObjectMeta = &ApiResources.ObjectMeta{
				Namespace: data[EntityInfluxPlanning.ContainerNamespace],
				Name:      data[EntityInfluxPlanning.ContainerPodName],
			}

			tempTopControllerKind := data[EntityInfluxPlanning.ContainerTopControllerKind]
			var topControllerKind ApiResources.Kind
			if val, ok := ApiResources.Kind_value[tempTopControllerKind]; ok {
				topControllerKind = ApiResources.Kind(val)
			}

			podPlanning.TopController = &ApiResources.Controller{
				ObjectMeta: &ApiResources.ObjectMeta{
					Namespace: data[EntityInfluxPlanning.ContainerNamespace],
					Name:      data[EntityInfluxPlanning.ContainerTopControllerName],
				},
				Kind: topControllerKind,
			}

			startTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ContainerStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ContainerEndTime], 10, 64)

			podPlanning.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			podPlanning.EndTime = &timestamp.Timestamp{
				Seconds: endTime,
			}

			policyTime, _ := strconv.ParseInt(data[EntityInfluxPlanning.ContainerPolicyTime], 10, 64)
			podPlanning.AssignPodPolicy = &ApiResources.AssignPodPolicy{
				Time: &timestamp.Timestamp{
					Seconds: policyTime,
				},
				Policy: &ApiResources.AssignPodPolicy_NodeName{
					NodeName: data[EntityInfluxPlanning.ContainerPolicy],
				},
			}

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ContainerPodTotalCost], 64)
			podPlanning.TotalCost = tempTotalCost

			containerPlanning := &ApiPlannings.ContainerPlanning{}
			containerPlanning.Name = data[EntityInfluxPlanning.ContainerName]

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

				containerPlanning.LimitPlannings = append(containerPlanning.LimitPlannings, metricDataList[0])
				containerPlanning.RequestPlannings = append(containerPlanning.RequestPlannings, metricDataList[1])
				containerPlanning.InitialLimitPlannings = append(containerPlanning.InitialLimitPlannings, metricDataList[2])
				containerPlanning.InitialRequestPlannings = append(containerPlanning.InitialRequestPlannings, metricDataList[3])
			}

			containerPlanning.LimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ContainerResourceLimitCPU]
			containerPlanning.LimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ContainerResourceLimitMemory]

			containerPlanning.RequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ContainerResourceRequestCPU]
			containerPlanning.RequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ContainerResourceRequestMemory]

			containerPlanning.InitialLimitPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ContainerInitialResourceLimitCPU]
			containerPlanning.InitialLimitPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ContainerInitialResourceLimitMemory]

			containerPlanning.InitialRequestPlannings[0].Data[0].NumValue = data[EntityInfluxPlanning.ContainerInitialResourceRequestCPU]
			containerPlanning.InitialRequestPlannings[1].Data[0].NumValue = data[EntityInfluxPlanning.ContainerInitialResourceRequestMemory]

			podPlanning.ContainerPlannings = append(podPlanning.ContainerPlannings, containerPlanning)

			podPlannings = append(podPlannings, podPlanning)
		}
	}

	return podPlannings, nil
}
