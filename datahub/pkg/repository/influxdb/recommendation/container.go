package recommendation

import (
	"fmt"
	EntityInfluxRecommend "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/recommendation"
	"github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/utils/enumconv"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	Utils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"math"
	"strconv"
	"time"
)

var (
	scope = log.RegisterScope("recommendation_db_measurement", "recommendation DB measurement", 0)
)

// ContainerRepository is used to operate node measurement of recommendation database
type ContainerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

// IsTag checks the column is tag or not
func (containerRepository *ContainerRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxRecommend.ContainerTags {
		if column == string(tag) {
			return true
		}
	}
	return false
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

// CreateContainerRecommendations add containers information container measurement
func (c *ContainerRepository) CreateContainerRecommendations(in *datahub_v1alpha1.CreatePodRecommendationsRequest) error {
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

		podNS := podRecommendation.GetNamespacedName().GetNamespace()
		podName := podRecommendation.GetNamespacedName().GetName()
		podTotalCost := podRecommendation.GetTotalCost()
		containerRecommendations := podRecommendation.GetContainerRecommendations()
		topController := podRecommendation.GetTopController()

		podPolicy := podRecommendation.GetAssignPodPolicy().GetPolicy()
		podPolicyValue := ""
		switch podPolicy.(type) {
		case *datahub_v1alpha1.AssignPodPolicy_NodeName:
			podPolicyValue = podPolicy.(*datahub_v1alpha1.AssignPodPolicy_NodeName).NodeName
		case *datahub_v1alpha1.AssignPodPolicy_NodePriority:
			nodeList := podPolicy.(*datahub_v1alpha1.AssignPodPolicy_NodePriority).NodePriority.GetNodes()
			if len(nodeList) > 0 {
				podPolicyValue = nodeList[0]
			}
			podPolicyValue = podPolicy.(*datahub_v1alpha1.AssignPodPolicy_NodePriority).NodePriority.GetNodes()[0]
		case *datahub_v1alpha1.AssignPodPolicy_NodeSelector:
			nodeMap := podPolicy.(*datahub_v1alpha1.AssignPodPolicy_NodeSelector).NodeSelector.Selector
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
				EntityInfluxRecommend.ContainerTopControllerName: topController.GetNamespacedName().GetName(),
				EntityInfluxRecommend.ContainerTopControllerKind: enumconv.KindDisp[(topController.GetKind())],
				EntityInfluxRecommend.ContainerPolicy:            podPolicyValue,
				EntityInfluxRecommend.ContainerPolicyTime:        podRecommendation.GetAssignPodPolicy().GetTime().GetSeconds(),
				EntityInfluxRecommend.ContainerPodTotalCost:      podTotalCost,
			}

			initialLimitRecommendation := make(map[datahub_v1alpha1.MetricType]interface{})
			if containerRecommendation.GetInitialLimitRecommendations() != nil {
				for _, rec := range containerRecommendation.GetInitialLimitRecommendations() {
					// One and only one record in initial limit recommendation
					initialLimitRecommendation[rec.GetMetricType()] = rec.Data[0].NumValue
				}
			}
			initialRequestRecommendation := make(map[datahub_v1alpha1.MetricType]interface{})
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
						case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := Utils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxRecommend.ContainerResourceLimitCPU] = numVal
							}
							if value, ok := initialLimitRecommendation[datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := Utils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxRecommend.ContainerInitialResourceLimitCPU] = numVal
								}
							} else {
								newFields[EntityInfluxRecommend.ContainerInitialResourceLimitCPU] = float64(0)
							}
						case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := Utils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxRecommend.ContainerResourceLimitMemory] = memoryBytes
							}
							if value, ok := initialLimitRecommendation[datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := Utils.StringToFloat64(value.(string)); err == nil {
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
						case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							if numVal, err := Utils.StringToFloat64(datum.NumValue); err == nil {
								newFields[EntityInfluxRecommend.ContainerResourceRequestCPU] = numVal
							}
							if value, ok := initialRequestRecommendation[datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE]; ok {
								if numVal, err := Utils.StringToFloat64(value.(string)); err == nil {
									newFields[EntityInfluxRecommend.ContainerInitialResourceRequestCPU] = numVal
								}
							} else {
								newFields[EntityInfluxRecommend.ContainerInitialResourceRequestCPU] = float64(0)
							}
						case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
							if numVal, err := Utils.StringToFloat64(datum.NumValue); err == nil {
								memoryBytes := math.Floor(numVal)
								newFields[EntityInfluxRecommend.ContainerResourceRequestMemory] = memoryBytes
							}
							if value, ok := initialRequestRecommendation[datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES]; ok {
								if numVal, err := Utils.StringToFloat64(value.(string)); err == nil {
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
func (c *ContainerRepository) ListContainerRecommendations(in *datahub_v1alpha1.ListPodRecommendationsRequest) ([]*datahub_v1alpha1.PodRecommendation, error) {
	kind := in.GetKind()
	granularity := in.GetGranularity()

	podRecommendations := make([]*datahub_v1alpha1.PodRecommendation, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Container,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxRecommend.ContainerName, EntityInfluxRecommend.ContainerNamespace, EntityInfluxRecommend.ContainerPodName},
	}

	nameCol := ""
	switch kind {
	case datahub_v1alpha1.Kind_POD:
		nameCol = string(EntityInfluxRecommend.ContainerPodName)
	case datahub_v1alpha1.Kind_DEPLOYMENT:
		nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
	case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
		nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
	case datahub_v1alpha1.Kind_STATEFULSET:
		nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
	default:
		return podRecommendations, errors.Errorf("no matching kind for Datahub Kind, received Kind: %s", datahub_v1alpha1.Kind_name[int32(kind)])
	}
	influxdbStatement.AppendWhereClause(EntityInfluxRecommend.ContainerNamespace, "=", in.GetNamespacedName().GetNamespace())
	influxdbStatement.AppendWhereClause(nameCol, "=", in.GetNamespacedName().GetName())

	influxdbStatement.AppendWhereClauseFromTimeCondition()

	if kind != datahub_v1alpha1.Kind_POD {
		kindConditionStr := fmt.Sprintf("\"%s\"='%s'", EntityInfluxRecommend.ContainerTopControllerKind, enumconv.KindDisp[kind])
		influxdbStatement.AppendWhereClause(EntityInfluxRecommend.ContainerTopControllerKind, "=", kindConditionStr)
	}

	if granularity == 0 || granularity == 30 {
		tempCondition := fmt.Sprintf("(\"%s\"='' OR \"%s\"='30')", EntityInfluxRecommend.ContainerGranularity, EntityInfluxRecommend.ContainerGranularity)
		influxdbStatement.AppendWhereClauseDirectly(tempCondition)
	} else {
		influxdbStatement.AppendWhereClause(EntityInfluxRecommend.ContainerGranularity, "=", strconv.FormatInt(granularity, 10))
	}

	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()

	cmd := influxdbStatement.BuildQueryCmd()
	scope.Debugf(fmt.Sprintf("ListContainerRecommendations: %s", cmd))

	podRecommendations, err := c.queryRecommendationNew(cmd, granularity)
	if err != nil {
		return podRecommendations, err
	}

	return podRecommendations, nil
}

func (c *ContainerRepository) ListAvailablePodRecommendations(in *datahub_v1alpha1.ListPodRecommendationsRequest) ([]*datahub_v1alpha1.PodRecommendation, error) {
	kind := in.GetKind()
	granularity := in.GetGranularity()

	podRecommendations := make([]*datahub_v1alpha1.PodRecommendation, 0)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    Container,
		QueryCondition: DBCommon.BuildQueryConditionV1(in.GetQueryCondition()),
		GroupByTags:    []string{EntityInfluxRecommend.ContainerName, EntityInfluxRecommend.ContainerNamespace, EntityInfluxRecommend.ContainerPodName},
	}

	nameCol := ""
	switch kind {
	case datahub_v1alpha1.Kind_POD:
		nameCol = string(EntityInfluxRecommend.ContainerPodName)
	case datahub_v1alpha1.Kind_DEPLOYMENT:
		nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
	case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
		nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
	case datahub_v1alpha1.Kind_STATEFULSET:
		nameCol = string(EntityInfluxRecommend.ContainerTopControllerName)
	default:
		return podRecommendations, errors.Errorf("no matching kind for Datahub Kind, received Kind: %s", datahub_v1alpha1.Kind_name[int32(kind)])
	}
	influxdbStatement.AppendWhereClause(EntityInfluxRecommend.ContainerNamespace, "=", in.GetNamespacedName().GetNamespace())
	influxdbStatement.AppendWhereClause(nameCol, "=", in.GetNamespacedName().GetName())

	if granularity == 0 || granularity == 30 {
		tempCondition := fmt.Sprintf("(\"%s\"='' OR \"%s\"='30')", EntityInfluxRecommend.ContainerGranularity, EntityInfluxRecommend.ContainerGranularity)
		influxdbStatement.AppendWhereClauseDirectly(tempCondition)
	} else {
		influxdbStatement.AppendWhereClause(EntityInfluxRecommend.ContainerGranularity, "=", strconv.FormatInt(granularity, 10))
	}

	whereStrTime := ""
	applyTime := in.GetQueryCondition().GetTimeRange().GetApplyTime().GetSeconds()
	if applyTime > 0 {
		whereStrTime = fmt.Sprintf(" \"end_time\">=%d AND \"start_time\"<=%d", applyTime, applyTime)
	}
	influxdbStatement.AppendWhereClauseDirectly(whereStrTime)

	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()

	cmd := influxdbStatement.BuildQueryCmd()
	scope.Debugf(fmt.Sprintf("ListContainerRecommendations: %s", cmd))

	podRecommendations, err := c.queryRecommendationNew(cmd, granularity)
	if err != nil {
		return podRecommendations, err
	}

	return podRecommendations, nil
}

func (c *ContainerRepository) queryRecommendationNew(cmd string, granularity int64) ([]*datahub_v1alpha1.PodRecommendation, error) {
	podRecommendations := make([]*datahub_v1alpha1.PodRecommendation, 0)

	results, err := c.influxDB.QueryDB(cmd, string(RepoInflux.Recommendation))
	if err != nil {
		return podRecommendations, err
	}

	rows := InternalInflux.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			podRecommendation := &datahub_v1alpha1.PodRecommendation{}
			podRecommendation.NamespacedName = &datahub_v1alpha1.NamespacedName{
				Namespace: data[EntityInfluxRecommend.ContainerNamespace],
				Name:      data[EntityInfluxRecommend.ContainerPodName],
			}

			tempTopControllerKind := data[EntityInfluxRecommend.ContainerTopControllerKind]
			var topControllerKind datahub_v1alpha1.Kind
			if val, ok := enumconv.KindEnum[tempTopControllerKind]; ok {
				topControllerKind = val
			}

			podRecommendation.TopController = &datahub_v1alpha1.TopController{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
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
			podRecommendation.AssignPodPolicy = &datahub_v1alpha1.AssignPodPolicy{
				Time: &timestamp.Timestamp{
					Seconds: policyTime,
				},
				Policy: &datahub_v1alpha1.AssignPodPolicy_NodeName{
					NodeName: data[EntityInfluxRecommend.ContainerPolicy],
				},
			}

			tempTotalCost, _ := strconv.ParseFloat(data[EntityInfluxRecommend.ContainerPodTotalCost], 64)
			podRecommendation.TotalCost = tempTotalCost

			containerRecommendation := &datahub_v1alpha1.ContainerRecommendation{}
			containerRecommendation.Name = data[EntityInfluxRecommend.ContainerName]

			metricTypeList := []datahub_v1alpha1.MetricType{datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES}
			sampleTime := &timestamp.Timestamp{
				Seconds: startTime,
			}
			sampleEndTime := &timestamp.Timestamp{
				Seconds: endTime,
			}

			//
			for _, metricType := range metricTypeList {
				metricDataList := make([]*datahub_v1alpha1.MetricData, 0)
				for a := 0; a < 4; a++ {
					sample := &datahub_v1alpha1.Sample{
						Time:    sampleTime,
						EndTime: sampleEndTime,
					}

					metricData := &datahub_v1alpha1.MetricData{
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

			podRecommendation.ContainerRecommendations = append(podRecommendation.ContainerRecommendations, containerRecommendation)

			podRecommendations = append(podRecommendations, podRecommendation)
		}
	}

	return podRecommendations, nil
}
