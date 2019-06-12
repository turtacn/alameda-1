package metric

import (
	"fmt"
	container_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/metric/container"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"strings"
	"time"
)

var (
	scope = log.RegisterScope("cluster_status_db_measurement", "cluster_status DB measurement", 0)
)

// ContainerRepository is used to operate node measurement of cluster_status database
type ContainerRepository struct {
	influxDB *influxdb.InfluxDBRepository
}

// NewContainerRepositoryWithConfig New container repository with influxDB configuration
func NewContainerRepositoryWithConfig(influxDBCfg influxdb.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &influxdb.InfluxDBRepository{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

// ListContainerPredictionsByRequest list containers' prediction from influxDB
func (r *ContainerRepository) ListContainerMetrics(in *datahub_v1alpha1.ListPodMetricsRequest) ([]*datahub_v1alpha1.PodMetric, error) {
	podMetricList := make([]*datahub_v1alpha1.PodMetric, 0)

	whereClause := r.buildInfluxQLWhereClauseFromRequest(in)
	influxdbStatement := influxdb.Statement{
		Measurement: influxdb.Measurement(container_entity.MetricMeasurementName),
		WhereClause: whereClause,
		GroupByTags: []string{container_entity.PodNamespace, container_entity.PodName, container_entity.Name, container_entity.MetricType},
	}

	queryCondition := influxdb.QueryCondition{
		//StartTime:      in.GetQueryCondition().GetTimeRange().GetStartTime(),
		//EndTime:        request.QueryCondition.EndTime,
		//StepTime:       request.QueryCondition.StepTime,
		TimestampOrder: influxdb.Order(in.GetQueryCondition().GetOrder()),
		Limit:          int(in.GetQueryCondition().GetLimit()),
	}
	//influxdbStatement.AppendTimeConditionIntoWhereClause(queryCondition)
	influxdbStatement.SetLimitClauseFromQueryCondition(queryCondition)
	influxdbStatement.SetOrderClauseFromQueryCondition(queryCondition)
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := r.influxDB.QueryDB(cmd, container_entity.MetricDatabaseName)
	if err != nil {
		return podMetricList, errors.Wrap(err, "list container prediction failed")
	}

	rows := influxdb.PackMap(results)

	podMetricList = r.getPodMetricsFromInfluxRows(rows)
	return podMetricList, nil
}

func (r *ContainerRepository) getPodMetricsFromInfluxRows(rows []*influxdb.InfluxDBRow) []*datahub_v1alpha1.PodMetric {
	podMap := map[string]*datahub_v1alpha1.PodMetric{}

	podContainerMap := map[string]*datahub_v1alpha1.ContainerMetric{}
	podContainerMetricMap := map[string]*datahub_v1alpha1.MetricData{}
	podContainerMetricSampleMap := map[string][]*datahub_v1alpha1.Sample{}

	for _, row := range rows {
		podNamespace := row.Tags[container_entity.PodNamespace]
		podName := row.Tags[container_entity.PodName]
		containerName := row.Tags[container_entity.Name]
		metricType := row.Tags[container_entity.MetricType]

		metricValue := datahub_v1alpha1.MetricType(datahub_v1alpha1.MetricType_value[metricType])
		switch metricType {
		case metric.TypeContainerCPUUsageSecondsPercentage:
			metricValue = datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE
		case metric.TypeContainerMemoryUsageBytes:
			metricValue = datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES
		}

		podMap[podNamespace+"|"+podName] = &datahub_v1alpha1.PodMetric{}
		podMap[podNamespace+"|"+podName].NamespacedName = &datahub_v1alpha1.NamespacedName{
			Namespace: podNamespace,
			Name:      podName,
		}

		podContainerMap[podNamespace+"|"+podName+"|"+containerName] = &datahub_v1alpha1.ContainerMetric{}
		podContainerMap[podNamespace+"|"+podName+"|"+containerName].Name = containerName

		metricKey := podNamespace + "|" + podName + "|" + containerName + "|" + metricType
		podContainerMetricMap[metricKey] = &datahub_v1alpha1.MetricData{}
		podContainerMetricMap[metricKey].MetricType = metricValue

		for _, data := range row.Data {
			t, _ := time.Parse(time.RFC3339, data[container_entity.PodTime])
			value := data[container_entity.Value]

			googleTimestamp, _ := ptypes.TimestampProto(t)

			tempSample := &datahub_v1alpha1.Sample{
				Time:     googleTimestamp,
				NumValue: value,
			}
			podContainerMetricSampleMap[metricKey] = append(podContainerMetricSampleMap[metricKey], tempSample)
		}
	}

	for k := range podContainerMetricMap {
		podNamespace := strings.Split(k, "|")[0]
		podName := strings.Split(k, "|")[1]
		containerName := strings.Split(k, "|")[2]
		metricType := strings.Split(k, "|")[3]

		containerKey := podNamespace + "|" + podName + "|" + containerName
		metricKey := podNamespace + "|" + podName + "|" + containerName + "|" + metricType

		podContainerMetricMap[metricKey].Data = podContainerMetricSampleMap[metricKey]
		podContainerMap[containerKey].MetricData = append(podContainerMap[containerKey].MetricData, podContainerMetricMap[metricKey])
	}

	for k := range podContainerMap {
		podNamespace := strings.Split(k, "|")[0]
		podName := strings.Split(k, "|")[1]
		containerName := strings.Split(k, "|")[2]

		podKey := podNamespace + "|" + podName
		containerKey := podNamespace + "|" + podName + "|" + containerName

		podMap[podKey].ContainerMetrics = append(podMap[podKey].ContainerMetrics, podContainerMap[containerKey])
	}

	podList := make([]*datahub_v1alpha1.PodMetric, 0)
	for k := range podMap {
		podList = append(podList, podMap[k])
	}

	return podList
}

func (r *ContainerRepository) buildInfluxQLWhereClauseFromRequest(in *datahub_v1alpha1.ListPodMetricsRequest) string {
	whereClause := ""

	podNamespace := in.GetNamespacedName().GetNamespace()
	podName := in.GetNamespacedName().GetName()
	startTime := in.GetQueryCondition().GetTimeRange().GetStartTime().GetSeconds()
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	r.influxDB.AddWhereCondition(&whereClause, container_entity.PodNamespace, "=", podNamespace)
	r.influxDB.AddWhereCondition(&whereClause, container_entity.PodName, "=", podName)

	r.influxDB.AddTimeCondition(&whereClause, ">=", startTime)
	r.influxDB.AddTimeCondition(&whereClause, "<=", endTime)

	return whereClause
}
