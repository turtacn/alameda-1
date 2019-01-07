package recommendation

import (
	"fmt"
	"time"

	recommendation_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/recommendation"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
)

var (
	containerScope = log.RegisterScope("recommendation_db_container_measurement", "recommendation DB container measurement", 0)
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
func (containerRepository *ContainerRepository) CreateContainerRecommendations(podRecommendations []*datahub_v1alpha1.PodRecommendation) error {
	points := []*influxdb_client.Point{}
	for _, podRecommendation := range podRecommendations {
		if podRecommendation.GetApplyRecommendationNow() {
			//TODO
		}

		podNS := podRecommendation.GetNamespacedName().GetNamespace()
		podName := podRecommendation.GetNamespacedName().GetName()
		containerRecommendations := podRecommendation.GetContainerRecommendations()
		for _, containerRecommendation := range containerRecommendations {
			tags := map[string]string{
				string(recommendation_entity.ContainerNamespace): podNS,
				string(recommendation_entity.ContainerPodName):   podName,
				string(recommendation_entity.ContainerName):      containerRecommendation.GetName(),
			}
			fields := map[string]interface{}{
				//TODO
				string(recommendation_entity.ContainerPolicy): "",
			}

			for _, metricData := range containerRecommendation.GetLimitRecommendations() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						switch metricData.GetMetricType() {
						case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							newFields[string(recommendation_entity.ContainerResourceLimitCPU)] = datum.NumValue
						case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
							newFields[string(recommendation_entity.ContainerResourceLimitMemory)] = datum.NumValue
						}
						if pt, err := influxdb_client.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
							points = append(points, pt)
						} else {
							containerScope.Error(err.Error())
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
						switch metricData.GetMetricType() {
						case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							newFields[string(recommendation_entity.ContainerResourceRequestCPU)] = datum.NumValue
						case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
							newFields[string(recommendation_entity.ContainerResourceRequestMemory)] = datum.NumValue
						}
						if pt, err := influxdb_client.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
							points = append(points, pt)
						} else {
							containerScope.Error(err.Error())
						}
					}
				}
			}
			for _, metricData := range containerRecommendation.GetInitialLimitRecommendations() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						switch metricData.GetMetricType() {
						case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							newFields[string(recommendation_entity.ContainerInitialResourceLimitCPU)] = datum.NumValue
						case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
							newFields[string(recommendation_entity.ContainerInitialResourceLimitMemory)] = datum.NumValue
						}
						if pt, err := influxdb_client.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
							points = append(points, pt)
						} else {
							containerScope.Error(err.Error())
						}
					}
				}
			}
			for _, metricData := range containerRecommendation.GetInitialRequestRecommendations() {
				if data := metricData.GetData(); len(data) > 0 {
					for _, datum := range data {
						newFields := map[string]interface{}{}
						for key, value := range fields {
							newFields[key] = value
						}
						switch metricData.GetMetricType() {
						case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
							newFields[string(recommendation_entity.ContainerInitialResourceRequestCPU)] = datum.NumValue
						case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
							newFields[string(recommendation_entity.ContainerInitialResourceRequestMemory)] = datum.NumValue
						}
						if pt, err := influxdb_client.NewPoint(string(Container), tags, newFields, time.Unix(datum.GetTime().GetSeconds(), 0)); err == nil {
							points = append(points, pt)
						} else {
							containerScope.Error(err.Error())
						}
					}
				}
			}
		}
	}
	containerRepository.influxDB.WritePoints(points, influxdb_client.BatchPointsConfig{
		Database: string(influxdb.Recommendation),
	})
	return nil
}

// ListContainerRecommendations list container recommendations
func (containerRepository *ContainerRepository) ListContainerRecommendations(podNamespacedName *datahub_v1alpha1.NamespacedName, timeRange *datahub_v1alpha1.TimeRange) ([]*datahub_v1alpha1.PodRecommendation, error) {
	podRecommendations := []*datahub_v1alpha1.PodRecommendation{}
	reqPodNS := podNamespacedName.GetNamespace()
	reqPodName := podNamespacedName.GetName()
	reqStartTime := timeRange.GetStartTime()
	reqEndTime := timeRange.GetEndTime()
	whereStr := ""
	if reqPodNS != "" && reqPodName == "" {
		whereStr = fmt.Sprintf("WHERE \"%s\"='%s'", string(recommendation_entity.ContainerNamespace), reqPodNS)
	} else if reqPodNS == "" && reqPodName != "" {
		whereStr = fmt.Sprintf("WHERE \"%s\"='%s'", string(recommendation_entity.ContainerPodName), reqPodName)
	} else if reqPodNS != "" && reqPodName != "" {
		whereStr = fmt.Sprintf("WHERE \"%s\"='%s' AND \"%s\"='%s'", string(recommendation_entity.ContainerNamespace), reqPodNS, string(recommendation_entity.ContainerPodName), reqPodName)
	}

	timeConditionStr := ""
	if reqStartTime != nil && reqEndTime != nil {
		timeConditionStr = fmt.Sprintf("time >= %v AND time <= %v", utils.TimeStampToNanoSecond(reqStartTime), utils.TimeStampToNanoSecond(reqEndTime))
	} else if reqStartTime != nil && reqEndTime == nil {
		timeConditionStr = fmt.Sprintf("time >= %v", utils.TimeStampToNanoSecond(reqStartTime))
	} else if reqStartTime == nil && reqEndTime != nil {
		timeConditionStr = fmt.Sprintf("time <= %v", utils.TimeStampToNanoSecond(reqEndTime))
	}

	if whereStr == "" && timeConditionStr != "" {
		whereStr = fmt.Sprintf("WHERE %s", timeConditionStr)
	} else if whereStr != "" && timeConditionStr != "" {
		whereStr = fmt.Sprintf("%s AND %s", whereStr, timeConditionStr)
	}

	cmd := fmt.Sprintf("SELECT * FROM %s %s GROUP BY \"%s\",\"%s\",\"%s\" ORDER BY time DESC",
		string(Container), whereStr, recommendation_entity.ContainerName,
		recommendation_entity.ContainerNamespace, recommendation_entity.ContainerPodName)
	containerScope.Infof(fmt.Sprintf("ListContainerRecommendations: %s", cmd))
	if results, err := containerRepository.influxDB.QueryDB(cmd, string(influxdb.Recommendation)); err == nil {
		for _, result := range results {
			//container recommendation time series data
			for _, ser := range result.Series {
				podName := ser.Tags[string(recommendation_entity.ContainerPodName)]
				podNS := ser.Tags[string(recommendation_entity.ContainerNamespace)]
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
				for _, val := range ser.Values {
					timeColIdx := utils.GetTimeIdxFromColumns(ser.Columns)
					timeObj, _ := utils.ParseTime(val[timeColIdx].(string))
					for columnIdx, column := range ser.Columns {
						colVal := val[columnIdx].(string)
						sampleObj := utils.GetSampleInstance(&timeObj, colVal)
						if column == string(recommendation_entity.ContainerInitialResourceLimitCPU) {
							initialResourceLimitCPUData = append(initialResourceLimitCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerInitialResourceRequestCPU) {
							initialResourceRequestCPUData = append(initialResourceRequestCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceLimitCPU) {
							resourceLimitCPUData = append(resourceLimitCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceRequestCPU) {
							resourceRequestCPUData = append(resourceRequestCPUData, sampleObj)
						} else if column == string(recommendation_entity.ContainerInitialResourceLimitMemory) {
							initialResourceLimitMemoryData = append(initialResourceLimitMemoryData, sampleObj)
						} else if column == string(recommendation_entity.ContainerInitialResourceRequestMemory) {
							initialResourceRequestMemoryData = append(initialResourceRequestMemoryData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceLimitMemory) {
							resourceLimitMemoryData = append(resourceLimitMemoryData, sampleObj)
						} else if column == string(recommendation_entity.ContainerResourceRequestMemory) {
							resourceRequestMemoryData = append(resourceRequestMemoryData, sampleObj)
						}
					}
					containerRecommendation.InitialLimitRecommendations = append(containerRecommendation.InitialLimitRecommendations,
						&datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
							Data:       initialResourceLimitCPUData,
						}, &datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
							Data:       initialResourceLimitMemoryData,
						})
					containerRecommendation.InitialRequestRecommendations = append(containerRecommendation.InitialLimitRecommendations,
						&datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
							Data:       initialResourceRequestCPUData,
						},
						&datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
							Data:       initialResourceRequestMemoryData,
						})
					containerRecommendation.LimitRecommendations = append(containerRecommendation.InitialLimitRecommendations,
						&datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
							Data:       resourceLimitCPUData,
						},
						&datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
							Data:       resourceLimitMemoryData,
						})
					containerRecommendation.RequestRecommendations = append(containerRecommendation.InitialLimitRecommendations,
						&datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
							Data:       resourceRequestCPUData,
						},
						&datahub_v1alpha1.MetricData{
							MetricType: datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
							Data:       resourceRequestMemoryData,
						})
				}

				foundPodRec := false
				for podRecommendationIdx, podRecommendation := range podRecommendations {
					if podRecommendation.GetNamespacedName().GetNamespace() == podNS && podRecommendation.GetNamespacedName().GetName() == podName {
						foundPodRec = true
						podRecommendations[podRecommendationIdx].ContainerRecommendations = append(podRecommendations[podRecommendationIdx].ContainerRecommendations, containerRecommendation)
					}
				}
				if !foundPodRec {
					podRecommendations = append(podRecommendations, &datahub_v1alpha1.PodRecommendation{
						ContainerRecommendations: []*datahub_v1alpha1.ContainerRecommendation{
							containerRecommendation,
						},
					})
				}
			}

		}
		return podRecommendations, nil
	} else {
		return podRecommendations, err
	}
}
