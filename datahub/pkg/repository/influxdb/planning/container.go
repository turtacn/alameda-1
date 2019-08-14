package planning

import (
	"fmt"
	EntityInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/planning"
	EntityInfluxUtilsEnum "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/utils/enumconv"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"math"
	"strconv"
	"time"
)

var (
	scope = Log.RegisterScope("planning_db_measurement", "planning DB measurement", 0)
)

// ContainerRepository is used to operate container measurement of planning database
type ContainerRepository struct {
	influxDB *RepoInflux.InfluxDBRepository
}

// NewContainerRepository creates the ContainerRepository instance
func NewContainerRepository(influxDBCfg *RepoInflux.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &RepoInflux.InfluxDBRepository{
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
func (c *ContainerRepository) CreateContainerPlannings(in *DatahubV1alpha1.CreatePodPlanningsRequest) error {
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

		podNS := podPlanning.GetNamespacedName().GetNamespace()
		podName := podPlanning.GetNamespacedName().GetName()
		podTotalCost := podPlanning.GetTotalCost()
		containerPlannings := podPlanning.GetContainerPlannings()
		topController := podPlanning.GetTopController()

		podPolicy := podPlanning.GetAssignPodPolicy().GetPolicy()
		podPolicyValue := ""
		switch podPolicy.(type) {
		case *DatahubV1alpha1.AssignPodPolicy_NodeName:
			podPolicyValue = podPolicy.(*DatahubV1alpha1.AssignPodPolicy_NodeName).NodeName
		case *DatahubV1alpha1.AssignPodPolicy_NodePriority:
			nodeList := podPolicy.(*DatahubV1alpha1.AssignPodPolicy_NodePriority).NodePriority.GetNodes()
			if len(nodeList) > 0 {
				podPolicyValue = nodeList[0]
			}
			podPolicyValue = podPolicy.(*DatahubV1alpha1.AssignPodPolicy_NodePriority).NodePriority.GetNodes()[0]
		case *DatahubV1alpha1.AssignPodPolicy_NodeSelector:
			nodeMap := podPolicy.(*DatahubV1alpha1.AssignPodPolicy_NodeSelector).NodeSelector.Selector
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
				EntityInfluxPlanning.ContainerTopControllerName: topController.GetNamespacedName().GetName(),
				EntityInfluxPlanning.ContainerTopControllerKind: EntityInfluxUtilsEnum.KindDisp[(topController.GetKind())],
				EntityInfluxPlanning.ContainerPolicy:            podPolicyValue,
				EntityInfluxPlanning.ContainerPolicyTime:        podPlanning.GetAssignPodPolicy().GetTime().GetSeconds(),
				EntityInfluxPlanning.ContainerPodTotalCost:      podTotalCost,
			}

			initialLimitPlanning := make(map[DatahubV1alpha1.MetricType]interface{})
			if containerPlanning.GetInitialLimitPlannings() != nil {
				for _, rec := range containerPlanning.GetInitialLimitPlannings() {
					// One and only one record in initial limit recommendation
					initialLimitPlanning[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}
			initialRequestPlanning := make(map[DatahubV1alpha1.MetricType]interface{})
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
						case DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ContainerResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitPlanning[DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ContainerInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ContainerInitialResourceLimitCPU] = float64(0)
							}
						case DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ContainerResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitPlanning[DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES]; ok {
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
						case DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxPlanning.ContainerResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestPlanning[DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxPlanning.ContainerInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxPlanning.ContainerInitialResourceRequestCPU] = float64(0)
							}
						case DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxPlanning.ContainerResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestPlanning[DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES]; ok {
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
func (c *ContainerRepository) ListContainerPlannings(in *DatahubV1alpha1.ListPodPlanningsRequest) ([]*DatahubV1alpha1.PodPlanning, error) {
	kind := in.GetKind()
	granularity := in.GetGranularity()

	podPlannings := make([]*DatahubV1alpha1.PodPlanning, 0)

	influxdbStatement := RepoInflux.StatementNew{
		Measurement:    Container,
		QueryCondition: in.GetQueryCondition(),
		GroupByTags:    []string{EntityInfluxPlanning.ContainerName, EntityInfluxPlanning.ContainerNamespace, EntityInfluxPlanning.ContainerPodName},
	}

	nameCol := ""
	switch kind {
	case DatahubV1alpha1.Kind_POD:
		nameCol = string(EntityInfluxPlanning.ContainerPodName)
	case DatahubV1alpha1.Kind_DEPLOYMENT:
		nameCol = string(EntityInfluxPlanning.ContainerTopControllerName)
	case DatahubV1alpha1.Kind_DEPLOYMENTCONFIG:
		nameCol = string(EntityInfluxPlanning.ContainerTopControllerName)
	case DatahubV1alpha1.Kind_STATEFULSET:
		nameCol = string(EntityInfluxPlanning.ContainerTopControllerName)
	default:
		return podPlannings, errors.Errorf("no matching kind for Datahub Kind, received Kind: %s", DatahubV1alpha1.Kind_name[int32(kind)])
	}
	influxdbStatement.AppendWhereCondition(EntityInfluxPlanning.ContainerNamespace, "=", in.GetNamespacedName().GetNamespace())
	influxdbStatement.AppendWhereCondition(nameCol, "=", in.GetNamespacedName().GetName())

	influxdbStatement.AppendTimeConditionFromQueryCondition()

	if kind != DatahubV1alpha1.Kind_POD {
		kindConditionStr := fmt.Sprintf("\"%s\"='%s'", EntityInfluxPlanning.ContainerTopControllerKind, EntityInfluxUtilsEnum.KindDisp[kind])
		influxdbStatement.AppendWhereCondition(EntityInfluxPlanning.ContainerTopControllerKind, "=", kindConditionStr)
	}

	if granularity == 0 || granularity == 30 {
		tempCondition := fmt.Sprintf("(\"%s\"='' OR \"%s\"='30')", EntityInfluxPlanning.ContainerGranularity, EntityInfluxPlanning.ContainerGranularity)
		influxdbStatement.AppendWhereConditionDirect(tempCondition)
	} else {
		influxdbStatement.AppendWhereCondition(EntityInfluxPlanning.ContainerGranularity, "=", strconv.FormatInt(granularity, 10))
	}

	influxdbStatement.AppendOrderClauseFromQueryCondition()
	influxdbStatement.AppendLimitClauseFromQueryCondition()

	cmd := influxdbStatement.BuildQueryCmd()
	scope.Debugf(fmt.Sprintf("ListContainerPlannings: %s", cmd))

	podPlannings, err := c.queryPlannings(cmd, granularity)
	if err != nil {
		return podPlannings, err
	}

	return podPlannings, nil
}

func (c *ContainerRepository) queryPlannings(cmd string, granularity int64) ([]*DatahubV1alpha1.PodPlanning, error) {
	podPlannings := make([]*DatahubV1alpha1.PodPlanning, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Planning))
	if err != nil {
		return podPlannings, err
	}

	rows := RepoInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			podPlanning := &DatahubV1alpha1.PodPlanning{}
			podPlanning.PlanningType = DatahubV1alpha1.PlanningType(DatahubV1alpha1.PlanningType_value[data[EntityInfluxPlanning.ContainerPlanningType]])
			podPlanning.NamespacedName = &DatahubV1alpha1.NamespacedName{
				Namespace: data[EntityInfluxPlanning.ContainerNamespace],
				Name:      data[EntityInfluxPlanning.ContainerPodName],
			}

			tempTopControllerKind := data[EntityInfluxPlanning.ContainerTopControllerKind]
			var topControllerKind DatahubV1alpha1.Kind
			if val, ok := EntityInfluxUtilsEnum.KindEnum[tempTopControllerKind]; ok {
				topControllerKind = val
			}

			podPlanning.TopController = &DatahubV1alpha1.TopController{
				NamespacedName: &DatahubV1alpha1.NamespacedName{
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
			podPlanning.AssignPodPolicy = &DatahubV1alpha1.AssignPodPolicy{
				Time: &timestamp.Timestamp{
					Seconds: policyTime,
				},
				Policy: &DatahubV1alpha1.AssignPodPolicy_NodeName{
					NodeName: data[EntityInfluxPlanning.ContainerPolicy],
				},
			}

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxPlanning.ContainerPodTotalCost], 64)
			podPlanning.TotalCost = tempTotalCost

			containerPlanning := &DatahubV1alpha1.ContainerPlanning{}
			containerPlanning.Name = data[EntityInfluxPlanning.ContainerName]

			metricTypeList := []DatahubV1alpha1.MetricType{DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES}
			sampleTime := &timestamp.Timestamp{
				Seconds: startTime,
			}
			sampleEndTime := &timestamp.Timestamp{
				Seconds: endTime,
			}

			//
			for _, metricType := range metricTypeList {
				metricDataList := make([]*DatahubV1alpha1.MetricData, 0)
				for a := 0; a < 4; a++ {
					sample := &DatahubV1alpha1.Sample{
						Time:    sampleTime,
						EndTime: sampleEndTime,
					}

					metricData := &DatahubV1alpha1.MetricData{
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
