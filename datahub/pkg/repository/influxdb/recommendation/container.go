package recommendation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	recommendation_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/recommendation"
	"github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/utils/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
)

var (
	scope = log.RegisterScope("recommendation_db_measurement", "recommendation DB measurement", 0)
)

// ContainerRepository is used to operate node measurement of recommendation database
type ContainerRepository struct {
	influxDB *influxdb.InfluxDBRepository
}

// IsTag checks the column is tag or not
func (containerRepository *ContainerRepository) IsTag(column string) bool {
	for _, tag := range recommendation_entity.ContainerTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

// NewContainerRepository creates the ContainerRepository instance
func NewContainerRepository(influxDBCfg *influxdb.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &influxdb.InfluxDBRepository{
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

	points := make([]*influxdb_client.Point, 0)
	for _, podRecommendation := range podRecommendations {
		if podRecommendation.GetApplyRecommendationNow() {
			//TODO
		}

		podNS := podRecommendation.GetNamespacedName().GetNamespace()
		podName := podRecommendation.GetNamespacedName().GetName()
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
				recommendation_entity.ContainerNamespace:   podNS,
				recommendation_entity.ContainerPodName:     podName,
				recommendation_entity.ContainerName:        containerRecommendation.GetName(),
				recommendation_entity.ContainerGranularity: strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				//TODO
				//string(recommendation_entity.ContainerPolicy):            "",
				recommendation_entity.ContainerTopControllerName: topController.GetNamespacedName().GetName(),
				recommendation_entity.ContainerTopControllerKind: enumconv.KindDisp[(topController.GetKind())],
				recommendation_entity.ContainerPolicy:            podPolicyValue,
				recommendation_entity.ContainerPolicyTime:        podRecommendation.GetAssignPodPolicy().GetTime().GetSeconds(),
			}

			for _, metricData := range containerRecommendation.GetInitialLimitRecommendations() {
				for _, datum := range metricData.GetData() {
					newFields := map[string]interface{}{}
					for key, value := range fields {
						newFields[key] = value
					}
					newFields[recommendation_entity.ContainerStartTime] = datum.GetTime().GetSeconds()
					newFields[recommendation_entity.ContainerEndTime] = datum.GetEndTime().GetSeconds()

					switch metricData.GetMetricType() {
					case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerInitialResourceLimitCPU] = numVal
						}
					case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerInitialResourceLimitMemory] = numVal
						}
					}

					if pt, err := influxdb_client.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
						points = append(points, pt)
					} else {
						scope.Error(err.Error())
					}
				}
			}

			for _, metricData := range containerRecommendation.GetInitialRequestRecommendations() {
				for _, datum := range metricData.GetData() {
					newFields := map[string]interface{}{}
					for key, value := range fields {
						newFields[key] = value
					}
					newFields[recommendation_entity.ContainerStartTime] = datum.GetTime().GetSeconds()
					newFields[recommendation_entity.ContainerEndTime] = datum.GetEndTime().GetSeconds()

					switch metricData.GetMetricType() {
					case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerInitialResourceRequestCPU] = numVal
						}
					case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerInitialResourceRequestMemory] = numVal
						}
					}

					if pt, err := influxdb_client.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
						points = append(points, pt)
					} else {
						scope.Error(err.Error())
					}
				}
			}

			for _, metricData := range containerRecommendation.GetLimitRecommendations() {
				for _, datum := range metricData.GetData() {
					newFields := map[string]interface{}{}
					for key, value := range fields {
						newFields[key] = value
					}
					newFields[recommendation_entity.ContainerStartTime] = datum.GetTime().GetSeconds()
					newFields[recommendation_entity.ContainerEndTime] = datum.GetEndTime().GetSeconds()

					switch metricData.GetMetricType() {
					case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerResourceLimitCPU] = numVal
						}
					case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerResourceLimitMemory] = numVal
						}
					}

					if pt, err := influxdb_client.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
						points = append(points, pt)
					} else {
						scope.Error(err.Error())
					}
				}
			}

			for _, metricData := range containerRecommendation.GetRequestRecommendations() {
				for _, datum := range metricData.GetData() {
					newFields := map[string]interface{}{}
					for key, value := range fields {
						newFields[key] = value
					}
					newFields[recommendation_entity.ContainerStartTime] = datum.GetTime().GetSeconds()
					newFields[recommendation_entity.ContainerEndTime] = datum.GetEndTime().GetSeconds()

					switch metricData.GetMetricType() {
					case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerResourceRequestCPU] = numVal
						}
					case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
						if numVal, err := utils.StringToFloat64(strings.Split(datum.GetNumValue(), ".")[0]); err == nil {
							newFields[recommendation_entity.ContainerResourceRequestMemory] = numVal
						}
					}
					if pt, err := influxdb_client.NewPoint(string(Container),
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
	err := c.influxDB.WritePoints(points, influxdb_client.BatchPointsConfig{
		Database: string(influxdb.Recommendation),
	})

	if err != nil {
		return err
	}
	return nil
}

// ListContainerRecommendations list container recommendations
func (c *ContainerRepository) ListContainerRecommendations(in *datahub_v1alpha1.ListPodRecommendationsRequest) ([]*datahub_v1alpha1.PodRecommendation, error) {
	podNamespacedName := in.GetNamespacedName()
	queryCondition := in.GetQueryCondition()
	kind := in.GetKind()
	granularity := in.GetGranularity()

	podRecommendations := make([]*datahub_v1alpha1.PodRecommendation, 0)
	reqNS := podNamespacedName.GetNamespace()
	reqName := podNamespacedName.GetName()

	var (
		reqStartTime *timestamp.Timestamp
		reqEndTime   *timestamp.Timestamp
	)
	timeRange := queryCondition.GetTimeRange()
	if timeRange != nil {
		reqStartTime = timeRange.GetStartTime()
		reqEndTime = timeRange.GetEndTime()
	}

	whereStr := ""
	fieldToCompareRequestName := ""
	switch kind {
	case datahub_v1alpha1.Kind_POD:
		fieldToCompareRequestName = recommendation_entity.ContainerPodName
	case datahub_v1alpha1.Kind_DEPLOYMENT:
		fieldToCompareRequestName = recommendation_entity.ContainerTopControllerName
	case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
		fieldToCompareRequestName = recommendation_entity.ContainerTopControllerName
	default:
		return podRecommendations, errors.Errorf("no matching kind for Datahub Kind, received Kind: %s", datahub_v1alpha1.Kind_name[int32(kind)])
	}

	if reqNS != "" && reqName == "" {
		//whereStr = fmt.Sprintf("WHERE \"%s\"='%s'", string(recommendation_entity.ContainerNamespace), reqNS)
		c.influxDB.AddWhereCondition(&whereStr, recommendation_entity.ContainerNamespace, "=", reqNS)
	} else if reqNS == "" && reqName != "" {
		//whereStr = fmt.Sprintf("WHERE \"%s\"='%s'", fieldToCompareRequestName, reqName)
		c.influxDB.AddWhereCondition(&whereStr, fieldToCompareRequestName, "=", reqName)
	} else if reqNS != "" && reqName != "" {
		//whereStr = fmt.Sprintf("WHERE \"%s\"='%s' AND \"%s\"='%s'", string(recommendation_entity.ContainerNamespace), reqNS, fieldToCompareRequestName, reqName)
		c.influxDB.AddWhereCondition(&whereStr, recommendation_entity.ContainerNamespace, "=", reqNS)
		c.influxDB.AddWhereCondition(&whereStr, fieldToCompareRequestName, "=", reqName)
	}

	if reqStartTime != nil {
		c.influxDB.AddTimeCondition(&whereStr, ">=", reqStartTime.Seconds)
	}
	if reqEndTime != nil {
		c.influxDB.AddTimeCondition(&whereStr, "<=", reqEndTime.Seconds)
	}

	if kind != datahub_v1alpha1.Kind_POD {
		kindConditionStr := fmt.Sprintf("\"%s\"='%s'", recommendation_entity.ContainerTopControllerKind, enumconv.KindDisp[kind])
		c.influxDB.AddWhereCondition(&whereStr, recommendation_entity.ContainerTopControllerKind, "=", kindConditionStr)
	}

	if granularity == 0 || granularity == 30 {
		tempCondition := fmt.Sprintf("(\"%s\"='' OR \"%s\"='30')", recommendation_entity.ContainerGranularity, recommendation_entity.ContainerGranularity)
		c.influxDB.AddWhereConditionDirect(&whereStr, tempCondition)
	} else {
		c.influxDB.AddWhereCondition(&whereStr, recommendation_entity.ContainerGranularity, "=", strconv.FormatInt(granularity, 10))
	}

	orderStr := c.buildOrderClause(queryCondition)
	limitStr := c.buildLimitClause(queryCondition)

	cmd := fmt.Sprintf("SELECT * FROM %s %s GROUP BY \"%s\",\"%s\",\"%s\" %s %s",
		string(Container), whereStr, recommendation_entity.ContainerName,
		recommendation_entity.ContainerNamespace, recommendation_entity.ContainerPodName, orderStr, limitStr)
	scope.Debugf(fmt.Sprintf("ListContainerRecommendations: %s", cmd))

	podRecommendations, err := c.queryRecommendationNew(cmd, granularity)
	//podRecommendations, err := c.queryRecommendation(cmd)
	if err != nil {
		return podRecommendations, err
	}

	return podRecommendations, nil

}

func (c *ContainerRepository) buildOrderClause(queryCondition *datahub_v1alpha1.QueryCondition) string {
	if queryCondition == nil {
		return "ORDER BY time ASC"
	}
	if queryCondition.GetOrder() == datahub_v1alpha1.QueryCondition_DESC {
		return "ORDER BY time DESC"
	} else if queryCondition.GetOrder() == datahub_v1alpha1.QueryCondition_ASC {
		return "ORDER BY time ASC"
	}
	return "ORDER BY time ASC"
}

func (c *ContainerRepository) buildLimitClause(queryCondition *datahub_v1alpha1.QueryCondition) string {
	if queryCondition == nil {
		return ""
	}
	limit := queryCondition.GetLimit()
	if queryCondition.GetLimit() > 0 {
		return fmt.Sprintf("LIMIT %v", limit)
	}
	return ""
}

func (c *ContainerRepository) ListAvailablePodRecommendations(in *datahub_v1alpha1.ListPodRecommendationsRequest) ([]*datahub_v1alpha1.PodRecommendation, error) {
	//podRecommendations := make([]*datahub_v1alpha1.PodRecommendation, 0)
	granularity := in.GetGranularity()

	whereStrName := c.buildNameClause(in)
	whereStrKind := c.buildKindClause(in)
	whereStrTime := c.buildApplyTimeClause(in)

	whereStr := c.combineClause([]string{whereStrName, whereStrKind, whereStrTime})

	if granularity == 0 || granularity == 30 {
		tempCondition := fmt.Sprintf("(\"%s\"='' OR \"%s\"='30')", recommendation_entity.ContainerGranularity, recommendation_entity.ContainerGranularity)
		c.influxDB.AddWhereConditionDirect(&whereStr, tempCondition)
	} else {
		c.influxDB.AddWhereCondition(&whereStr, recommendation_entity.ContainerGranularity, "=", strconv.FormatInt(granularity, 10))
	}

	orderStr := c.buildOrderClause(in.QueryCondition)
	limitStr := c.buildLimitClause(in.QueryCondition)

	cmd := fmt.Sprintf("SELECT * FROM %s %s GROUP BY \"%s\",\"%s\",\"%s\" %s %s",
		string(Container), whereStr, recommendation_entity.ContainerName,
		recommendation_entity.ContainerNamespace, recommendation_entity.ContainerPodName, orderStr, limitStr)

	podRecommendations, err := c.queryRecommendationNew(cmd, granularity)
	if err != nil {
		return podRecommendations, err
	}

	return podRecommendations, nil
}

func (c *ContainerRepository) queryRecommendationNew(cmd string, granularity int64) ([]*datahub_v1alpha1.PodRecommendation, error) {
	podRecommendations := make([]*datahub_v1alpha1.PodRecommendation, 0)

	results, err := c.influxDB.QueryDB(cmd, string(influxdb.Recommendation))
	if err != nil {
		return podRecommendations, err
	}

	rows := influxdb.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			podRecommendation := &datahub_v1alpha1.PodRecommendation{}
			podRecommendation.NamespacedName = &datahub_v1alpha1.NamespacedName{
				Namespace: data[recommendation_entity.ContainerNamespace],
				Name:      data[recommendation_entity.ContainerPodName],
			}

			tempTopControllerKind := data[recommendation_entity.ContainerTopControllerKind]
			var topControllerKind datahub_v1alpha1.Kind
			if val, ok := enumconv.KindEnum[tempTopControllerKind]; ok {
				topControllerKind = val
			}

			podRecommendation.TopController = &datahub_v1alpha1.TopController{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: data[recommendation_entity.ContainerNamespace],
					Name:      data[recommendation_entity.ContainerTopControllerName],
				},
				Kind: topControllerKind,
			}

			startTime, _ := strconv.ParseInt(data[recommendation_entity.ContainerStartTime], 10, 64)
			endTime, _ := strconv.ParseInt(data[recommendation_entity.ContainerEndTime], 10, 64)

			podRecommendation.StartTime = &timestamp.Timestamp{
				Seconds: startTime,
			}

			podRecommendation.EndTime = &timestamp.Timestamp{
				Seconds: endTime,
			}

			policyTime, _ := strconv.ParseInt(data[recommendation_entity.ContainerPolicyTime], 10, 64)
			podRecommendation.AssignPodPolicy = &datahub_v1alpha1.AssignPodPolicy{
				Time: &timestamp.Timestamp{
					Seconds: policyTime,
				},
				Policy: &datahub_v1alpha1.AssignPodPolicy_NodeName{
					NodeName: data[recommendation_entity.ContainerPolicy],
				},
			}

			containerRecommendation := &datahub_v1alpha1.ContainerRecommendation{}
			containerRecommendation.Name = data[recommendation_entity.ContainerName]

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

			containerRecommendation.LimitRecommendations[0].Data[0].NumValue = data[recommendation_entity.ContainerResourceLimitCPU]
			containerRecommendation.LimitRecommendations[1].Data[0].NumValue = data[recommendation_entity.ContainerResourceLimitMemory]

			containerRecommendation.RequestRecommendations[0].Data[0].NumValue = data[recommendation_entity.ContainerResourceRequestCPU]
			containerRecommendation.RequestRecommendations[1].Data[0].NumValue = data[recommendation_entity.ContainerResourceRequestMemory]

			containerRecommendation.InitialLimitRecommendations[0].Data[0].NumValue = data[recommendation_entity.ContainerInitialResourceLimitCPU]
			containerRecommendation.InitialLimitRecommendations[1].Data[0].NumValue = data[recommendation_entity.ContainerInitialResourceLimitMemory]

			containerRecommendation.InitialRequestRecommendations[0].Data[0].NumValue = data[recommendation_entity.ContainerInitialResourceRequestCPU]
			containerRecommendation.InitialRequestRecommendations[1].Data[0].NumValue = data[recommendation_entity.ContainerInitialResourceRequestMemory]

			podRecommendation.ContainerRecommendations = append(podRecommendation.ContainerRecommendations, containerRecommendation)

			podRecommendations = append(podRecommendations, podRecommendation)
		}
	}

	return podRecommendations, nil
}

func (c *ContainerRepository) queryRecommendation(cmd string) ([]*datahub_v1alpha1.PodRecommendation, error) {
	podRecommendations := []*datahub_v1alpha1.PodRecommendation{}

	if results, err := c.influxDB.QueryDB(cmd, string(influxdb.Recommendation)); err == nil {
		for _, result := range results {
			//individual containers
			for _, ser := range result.Series {
				podName := ser.Tags[string(recommendation_entity.ContainerPodName)]
				podNS := ser.Tags[string(recommendation_entity.ContainerNamespace)]
				topControllerName := ""
				topControllerKind := datahub_v1alpha1.Kind_POD

				var startTime int64 = 0
				var endTime int64 = 0
				// per container time series data
				for _, val := range ser.Values {
					timeColIdx := utils.GetTimeIdxFromColumns(ser.Columns)
					timeObj, _ := utils.ParseTime(val[timeColIdx].(string))

					endTimeColIdx := utils.GetEndTimeIdxFromColumns(ser.Columns)
					ts, _ := val[endTimeColIdx].(json.Number).Int64()
					endTimeObj := time.Unix(ts, 0)

					containerRecommendation := &datahub_v1alpha1.ContainerRecommendation{
						Name:                          ser.Tags[string(recommendation_entity.ContainerName)],
						InitialLimitRecommendations:   []*datahub_v1alpha1.MetricData{},
						InitialRequestRecommendations: []*datahub_v1alpha1.MetricData{},
						LimitRecommendations:          []*datahub_v1alpha1.MetricData{},
						RequestRecommendations:        []*datahub_v1alpha1.MetricData{},
					}
					initialResourceLimitCPUData := []*datahub_v1alpha1.Sample{}
					initialResourceRequestCPUData := []*datahub_v1alpha1.Sample{}
					resourceLimitCPUData := []*datahub_v1alpha1.Sample{}
					resourceRequestCPUData := []*datahub_v1alpha1.Sample{}
					initialResourceLimitMemoryData := []*datahub_v1alpha1.Sample{}
					initialResourceRequestMemoryData := []*datahub_v1alpha1.Sample{}
					resourceLimitMemoryData := []*datahub_v1alpha1.Sample{}
					resourceRequestMemoryData := []*datahub_v1alpha1.Sample{}

					for columnIdx, column := range ser.Columns {
						if val[columnIdx] == nil {
							continue
						}

						if column == string(recommendation_entity.ContainerInitialResourceLimitCPU) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							initialResourceLimitCPUData = append(initialResourceLimitCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerInitialResourceRequestCPU) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							initialResourceRequestCPUData = append(initialResourceRequestCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceLimitCPU) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							resourceLimitCPUData = append(resourceLimitCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceRequestCPU) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							resourceRequestCPUData = append(resourceRequestCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerInitialResourceLimitMemory) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							initialResourceLimitMemoryData = append(initialResourceLimitMemoryData, sampleObj)
						} else if column == string(recommendation_entity.ContainerInitialResourceRequestMemory) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							initialResourceRequestMemoryData = append(initialResourceRequestMemoryData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceLimitMemory) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							resourceLimitMemoryData = append(resourceLimitMemoryData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceRequestMemory) {
							colVal := val[columnIdx].(json.Number).String()
							sampleObj := utils.GetSampleInstance(&timeObj, &endTimeObj, colVal)
							resourceRequestMemoryData = append(resourceRequestMemoryData, sampleObj)
						} else if column == string(recommendation_entity.ContainerStartTime) {
							startTime, _ = val[columnIdx].(json.Number).Int64()
						} else if column == string(recommendation_entity.ContainerEndTime) {
							endTime, _ = val[columnIdx].(json.Number).Int64()
						} else if column == string(recommendation_entity.ContainerTopControllerName) {
							topControllerName = val[columnIdx].(string)
						} else if column == string(recommendation_entity.ContainerTopControllerKind) {
							topControllerKind = enumconv.KindEnum[val[columnIdx].(string)]
						}
					}
					if len(initialResourceLimitCPUData) > 0 {
						containerRecommendation.InitialLimitRecommendations = append(containerRecommendation.InitialLimitRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
								Data:       initialResourceLimitCPUData,
							})
					}
					if len(initialResourceLimitMemoryData) > 0 {
						containerRecommendation.InitialLimitRecommendations = append(containerRecommendation.InitialLimitRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
								Data:       initialResourceLimitMemoryData,
							})
					}
					if len(initialResourceRequestCPUData) > 0 {
						containerRecommendation.InitialRequestRecommendations = append(containerRecommendation.InitialRequestRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
								Data:       initialResourceRequestCPUData,
							})
					}
					if len(initialResourceRequestMemoryData) > 0 {
						containerRecommendation.InitialRequestRecommendations = append(containerRecommendation.InitialRequestRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
								Data:       initialResourceRequestMemoryData,
							})
					}
					if len(resourceLimitCPUData) > 0 {
						containerRecommendation.LimitRecommendations = append(containerRecommendation.LimitRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
								Data:       resourceLimitCPUData,
							})
					}
					if len(resourceLimitMemoryData) > 0 {
						containerRecommendation.LimitRecommendations = append(containerRecommendation.LimitRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
								Data:       resourceLimitMemoryData,
							})
					}
					if len(resourceRequestCPUData) > 0 {
						containerRecommendation.RequestRecommendations = append(containerRecommendation.RequestRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
								Data:       resourceRequestCPUData,
							})
					}
					if len(resourceRequestMemoryData) > 0 {
						containerRecommendation.RequestRecommendations = append(containerRecommendation.RequestRecommendations,
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
								Data:       resourceRequestMemoryData,
							})
					}

					foundPodRec := false
					for podRecommendationIdx, podRecommendation := range podRecommendations {
						if podRecommendation.GetStartTime() != nil && startTime != 0 && podRecommendation.GetStartTime().GetSeconds() == startTime &&
							podRecommendation.GetEndTime() != nil && endTime != 0 && podRecommendation.GetEndTime().GetSeconds() == endTime &&
							podRecommendation.GetNamespacedName().GetNamespace() == podNS && podRecommendation.GetNamespacedName().GetName() == podName {
							foundPodRec = true
							podRecommendations[podRecommendationIdx].ContainerRecommendations = append(podRecommendations[podRecommendationIdx].ContainerRecommendations, containerRecommendation)
							if startTime != 0 {
								podRecommendations[podRecommendationIdx].StartTime = &timestamp.Timestamp{
									Seconds: startTime,
								}
							}
							if endTime != 0 {
								podRecommendations[podRecommendationIdx].EndTime = &timestamp.Timestamp{
									Seconds: endTime,
								}
							}
						}
					}
					if !foundPodRec {
						podRec := &datahub_v1alpha1.PodRecommendation{
							NamespacedName: &datahub_v1alpha1.NamespacedName{
								Namespace: podNS,
								Name:      podName,
							},
							ContainerRecommendations: []*datahub_v1alpha1.ContainerRecommendation{
								containerRecommendation,
							},
							TopController: &datahub_v1alpha1.TopController{
								NamespacedName: &datahub_v1alpha1.NamespacedName{
									Namespace: podNS,
									Name:      topControllerName,
								},
								Kind: topControllerKind,
							},
						}
						if startTime != 0 {
							podRec.StartTime = &timestamp.Timestamp{
								Seconds: startTime,
							}
						}
						if endTime != 0 {
							podRec.EndTime = &timestamp.Timestamp{
								Seconds: endTime,
							}
						}
						podRecommendations = append(podRecommendations, podRec)
					}
				}
			}
		}
		return podRecommendations, nil
	} else {
		return podRecommendations, err
	}
}

func (c *ContainerRepository) combineClause(strList []string) string {
	ret := ""
	whereFlag := false

	for _, value := range strList {
		if value != "" && whereFlag == false {
			ret = fmt.Sprintf("WHERE %s ", value)
			whereFlag = true
		} else if value != "" {
			ret += fmt.Sprintf("AND %s ", value)
		}
	}

	return ret
}

func (c *ContainerRepository) buildNameClause(in *datahub_v1alpha1.ListPodRecommendationsRequest) string {
	ret := ""
	namespace := in.GetNamespacedName().GetNamespace()
	if namespace == "" {
		return ret
	}

	ret = fmt.Sprintf(" \"namespace\"='%s'", namespace)
	return ret
}

func (c *ContainerRepository) buildKindClause(in *datahub_v1alpha1.ListPodRecommendationsRequest) string {
	ret := ""
	col := ""

	name := in.GetNamespacedName().GetName()
	if name == "" {
		return ret
	}

	kind := in.GetKind()

	switch kind {
	case datahub_v1alpha1.Kind_POD:
		col = string(recommendation_entity.ContainerPodName)
	case datahub_v1alpha1.Kind_DEPLOYMENT:
		col = string(recommendation_entity.ContainerTopControllerName)
	case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
		col = string(recommendation_entity.ContainerTopControllerName)
	default:
		return ""
	}

	ret = fmt.Sprintf(" \"%s\"='%s'", col, name)
	return ret
}

func (c *ContainerRepository) buildApplyTimeClause(in *datahub_v1alpha1.ListPodRecommendationsRequest) string {
	ret := ""

	applyTime := in.GetQueryCondition().GetTimeRange().GetApplyTime().GetSeconds()
	if applyTime > 0 {
		ret = fmt.Sprintf(" \"end_time\">=%d AND \"start_time\"<=%d", applyTime, applyTime)
	}

	return ret
}
