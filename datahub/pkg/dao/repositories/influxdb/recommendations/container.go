package recommendations

import (
	"fmt"
	EntityInfluxRecommend "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/recommendations"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"math"
	"strconv"
	"time"
)

var (
	scope = Log.RegisterScope("recommendation_db_measurement", "recommendation DB measurement", 0)
)

// ContainerRepository is used to operate node measurement of recommendation database
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
	for _, tag := range EntityInfluxRecommend.ContainerTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

// CreateContainerRecommendations add containers information container measurement
func (c *ContainerRepository) CreateContainerRecommendations(in *ApiRecommendations.CreatePodRecommendationsRequest) error {
	podRecommendations := in.GetPodRecommendations()
	granularity := in.GetGranularity()
	if granularity == 0 {
		granularity = 30
	}

	points := make([]*InfluxClient.Point, 0)
	for _, podRecommendation := range podRecommendations {
		if podRecommendation.GetApplyRecommendationNow() {
			//TODO
		}

		podNS := podRecommendation.GetObjectMeta().GetNamespace()
		podName := podRecommendation.GetObjectMeta().GetName()
		podTotalCost := podRecommendation.GetTotalCost()
		containerRecommendations := podRecommendation.GetContainerRecommendations()
		topController := podRecommendation.GetTopController()

		podPolicy := podRecommendation.GetAssignPodPolicy().GetPolicy()
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

		for _, containerRecommendation := range containerRecommendations {
			tags := map[string]string{
				EntityInfluxRecommend.ContainerNamespace:   podNS,
				EntityInfluxRecommend.ContainerPodName:     podName,
				EntityInfluxRecommend.ContainerName:        containerRecommendation.GetName(),
				EntityInfluxRecommend.ContainerGranularity: strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				//TODO
				//string(EntityInfluxRecommend.ContainerPolicy):            "",
				EntityInfluxRecommend.ContainerTopControllerName: topController.GetObjectMeta().GetName(),
				EntityInfluxRecommend.ContainerTopControllerKind: topController.GetKind().String(),
				EntityInfluxRecommend.ContainerPolicy:            podPolicyValue,
				EntityInfluxRecommend.ContainerPolicyTime:        podRecommendation.GetAssignPodPolicy().GetTime().GetSeconds(),
				EntityInfluxRecommend.ContainerPodTotalCost:      podTotalCost,
			}

			initialLimitRecommendation := make(map[ApiCommon.MetricType]interface{})
			if containerRecommendation.GetInitialLimitRecommendations() != nil {
				for _, rec := range containerRecommendation.GetInitialLimitRecommendations() {
					// One and only one record in initial limit recommendation
					initialLimitRecommendation[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}
			initialRequestRecommendation := make(map[ApiCommon.MetricType]interface{})
			if containerRecommendation.GetInitialRequestRecommendations() != nil {
				for _, rec := range containerRecommendation.GetInitialRequestRecommendations() {
					// One and only one record in initial request recommendation
					initialRequestRecommendation[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}

			for _, metricData := range containerRecommendation.GetLimitRecommendations() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						newFields[EntityInfluxRecommend.ContainerStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxRecommend.ContainerEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxRecommend.ContainerResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitRecommendation[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxRecommend.ContainerInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxRecommend.ContainerInitialResourceLimitCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxRecommend.ContainerResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitRecommendation[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxRecommend.ContainerInitialResourceLimitMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxRecommend.ContainerInitialResourceLimitMemory] = float64(0)
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

			for _, metricData := range containerRecommendation.GetRequestRecommendations() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						newFields[EntityInfluxRecommend.ContainerStartTime] = datum.GetTime().GetSeconds()
						newFields[EntityInfluxRecommend.ContainerEndTime] = datum.GetEndTime().GetSeconds()

						switch metricData.GetMetricType() {
						case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxRecommend.ContainerResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestRecommendation[ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxRecommend.ContainerInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxRecommend.ContainerInitialResourceRequestCPU] = float64(0)
							}
						case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := DatahubUtils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxRecommend.ContainerResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestRecommendation[ApiCommon.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := DatahubUtils.StringToFloat64(value.(string)); err == nil {
									memoryBytes := math.Floor(numVal)
									newFields[EntityInfluxRecommend.ContainerInitialResourceRequestMemory] = memoryBytes
								}
							} else {
								newFields[EntityInfluxRecommend.ContainerInitialResourceRequestMemory] = float64(0)
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
		Database: string(RepoInflux.Recommendation),
	})

	if err != nil {
		return err
	}
	return nil
}

// ListContainerRecommendations list container recommendations
func (c *ContainerRepository) ListContainerRecommendations(in *ApiRecommendations.ListPodRecommendationsRequest) ([]*ApiRecommendations.PodRecommendation, error) {
	podRecommendations := make([]*ApiRecommendations.PodRecommendation, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Container,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxRecommend.ContainerName, EntityInfluxRecommend.ContainerNamespace, EntityInfluxRecommend.ContainerPodName},
	}

	kind := in.GetKind()
	granularity := in.GetGranularity()

	if granularity == 0 {
		granularity = 30
	}

	for _, objMeta := range in.GetObjectMeta() {
		tempCondition := ""
		name := objMeta.GetName()

		nameCol := ""
		switch kind {
		case ApiResources.Kind_KIND_UNDEFINED:
			nameCol = string(EntityInfluxRecommend.ContainerPodName)
		case ApiResources.Kind_DEPLOYMENT:
			nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
		case ApiResources.Kind_DEPLOYMENTCONFIG:
			nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
		case ApiResources.Kind_STATEFULSET:
			nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
		default:
			return podRecommendations, errors.Errorf("no matching kind for Datahub Kind, received Kind: %s", ApiResources.Kind_name[int32(kind)])
		}

		keyList := []string{nameCol, EntityInfluxRecommend.ContainerGranularity}
		valueList := []string{name, strconv.FormatInt(granularity, 10)}

		if kind != ApiResources.Kind_KIND_UNDEFINED {
			keyList = append(keyList, EntityInfluxRecommend.ContainerTopControllerKind)
			valueList = append(valueList, kind.String())
		}

		tempCondition = influxdbStatement.GenerateCondition(keyList, valueList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()
	scope.Debugf(fmt.Sprintf("ListContainerRecommendations: %s", cmd))

	podRecommendations, err := c.queryRecommendation(cmd, granularity)
	if err != nil {
		return podRecommendations, err
	}

	return podRecommendations, nil
}

func (c *ContainerRepository) ListAvailablePodRecommendations(in *ApiRecommendations.ListPodRecommendationsRequest) ([]*ApiRecommendations.PodRecommendation, error) {
	kind := in.GetKind()
	granularity := in.GetGranularity()

	if granularity == 0 {
		granularity = 30
	}

	podRecommendations := make([]*ApiRecommendations.PodRecommendation, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Container,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxRecommend.ContainerName, EntityInfluxRecommend.ContainerNamespace, EntityInfluxRecommend.ContainerPodName},
	}

	for _, objMeta := range in.GetObjectMeta() {
		tempCondition := ""
		name := objMeta.GetName()

		nameCol := ""
		switch kind {
		case ApiResources.Kind_KIND_UNDEFINED:
			nameCol = string(EntityInfluxRecommend.ContainerPodName)
		case ApiResources.Kind_DEPLOYMENT:
			nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
		case ApiResources.Kind_DEPLOYMENTCONFIG:
			nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
		case ApiResources.Kind_STATEFULSET:
			nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
		default:
			return podRecommendations, errors.Errorf("no matching kind for Datahub Kind, received Kind: %s", ApiResources.Kind_name[int32(kind)])
		}

		applyTime := in.GetQueryCondition().GetTimeRange().GetApplyTime().GetSeconds()

		conditionList := []string{
			fmt.Sprintf("\"%s\"='%s'", nameCol, name),
			fmt.Sprintf("\"%s\"='%d'", EntityInfluxRecommend.ContainerGranularity, granularity),
			fmt.Sprintf("\"%s\">=%d", EntityInfluxRecommend.ContainerStartTime, applyTime),
			fmt.Sprintf("\"%s\"<=%d", EntityInfluxRecommend.ContainerEndTime, applyTime),
		}

		if kind != ApiResources.Kind_KIND_UNDEFINED {
			kindCondition := fmt.Sprintf("\"%s\"='%s'", EntityInfluxRecommend.ContainerTopControllerKind, kind.String())
			conditionList = append(conditionList, kindCondition)
		}

		tempCondition = influxdbStatement.GenerateConditionByList(conditionList, "AND")
		influxdbStatement.AppendWhereClauseDirectly("OR", tempCondition)
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()
	scope.Debugf(fmt.Sprintf("ListContainerRecommendations: %s", cmd))

	podRecommendations, err := c.queryRecommendation(cmd, granularity)
	if err != nil {
		return podRecommendations, err
	}

	return podRecommendations, nil
}

func (c *ContainerRepository) queryRecommendation(cmd string, granularity int64) ([]*ApiRecommendations.PodRecommendation, error) {
	podRecommendations := make([]*ApiRecommendations.PodRecommendation, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Recommendation))
	if err != nil {
		return podRecommendations, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			podRecommendation := &ApiRecommendations.PodRecommendation{}
			podRecommendation.ObjectMeta = &ApiResources.ObjectMeta{
				Namespace: data[EntityInfluxRecommend.ContainerNamespace],
				Name:      data[EntityInfluxRecommend.ContainerPodName],
			}

			tempTopControllerKind := data[EntityInfluxRecommend.ContainerTopControllerKind]
			var topControllerKind ApiResources.Kind
			if val, ok := ApiResources.Kind_value[tempTopControllerKind]; ok {
				topControllerKind = ApiResources.Kind(val)
			}

			podRecommendation.TopController = &ApiResources.Controller{
				ObjectMeta: &ApiResources.ObjectMeta{
					Namespace: data[EntityInfluxRecommend.ContainerNamespace],
					Name:      data[EntityInfluxRecommend.ContainerTopControllerName],
				},
				Kind: topControllerKind,
			}

			startTime, _ := strconv.ParseInt(data[EntityInfluxRecommend.ContainerStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[EntityInfluxRecommend.ContainerEndTime], 10, 64)

			podRecommendation.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			podRecommendation.EndTime = &timestamp.Timestamp{
				Seconds: endTime,
			}

			policyTime, _ := strconv.ParseInt(data[EntityInfluxRecommend.ContainerPolicyTime], 10, 64)
			podRecommendation.AssignPodPolicy = &ApiResources.AssignPodPolicy{
				Time: &timestamp.Timestamp{
					Seconds: policyTime,
				},
				Policy: &ApiResources.AssignPodPolicy_NodeName{
					NodeName: data[EntityInfluxRecommend.ContainerPolicy],
				},
			}

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ContainerPodTotalCost], 64)
			podRecommendation.TotalCost = tempTotalCost

			containerRecommendation := &ApiRecommendations.ContainerRecommendation{}
			containerRecommendation.Name = data[EntityInfluxRecommend.ContainerName]

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

				containerRecommendation.LimitRecommendations = append(containerRecommendation.LimitRecommendations, metricDataList[0])
				containerRecommendation.RequestRecommendations = append(containerRecommendation.RequestRecommendations, metricDataList[1])
				containerRecommendation.InitialLimitRecommendations = append(containerRecommendation.InitialLimitRecommendations, metricDataList[2])
				containerRecommendation.InitialRequestRecommendations = append(containerRecommendation.InitialRequestRecommendations, metricDataList[3])
			}

			containerRecommendation.LimitRecommendations[0].Data[0].NumValue = data[EntityInfluxRecommend.ContainerResourceLimitCPU]
			containerRecommendation.LimitRecommendations[1].Data[0].NumValue = data[EntityInfluxRecommend.ContainerResourceLimitMemory]

			containerRecommendation.RequestRecommendations[0].Data[0].NumValue = data[EntityInfluxRecommend.ContainerResourceRequestCPU]
			containerRecommendation.RequestRecommendations[1].Data[0].NumValue = data[EntityInfluxRecommend.ContainerResourceRequestMemory]

			containerRecommendation.InitialLimitRecommendations[0].Data[0].NumValue = data[EntityInfluxRecommend.ContainerInitialResourceLimitCPU]
			containerRecommendation.InitialLimitRecommendations[1].Data[0].NumValue = data[EntityInfluxRecommend.ContainerInitialResourceLimitMemory]

			containerRecommendation.InitialRequestRecommendations[0].Data[0].NumValue = data[EntityInfluxRecommend.ContainerInitialResourceRequestCPU]
			containerRecommendation.InitialRequestRecommendations[1].Data[0].NumValue = data[EntityInfluxRecommend.ContainerInitialResourceRequestMemory]

			isPodInList := false
			for index := range podRecommendations {
				if podRecommendations[index].ObjectMeta.Name == data[EntityInfluxRecommend.ContainerPodName] && podRecommendations[index].ObjectMeta.Namespace == data[EntityInfluxRecommend.ContainerNamespace] {
					if podRecommendations[index].StartTime.Seconds == startTime && podRecommendations[index].EndTime.Seconds == endTime {
						podRecommendations[index].ContainerRecommendations = append(podRecommendations[index].ContainerRecommendations, containerRecommendation)
						isPodInList = true
						break
					}
				}
			}

			if isPodInList == false {
				podRecommendation.ContainerRecommendations = append(podRecommendation.ContainerRecommendations, containerRecommendation)
				podRecommendations = append(podRecommendations, podRecommendation)
			}
		}
	}

	return podRecommendations, nil
}
