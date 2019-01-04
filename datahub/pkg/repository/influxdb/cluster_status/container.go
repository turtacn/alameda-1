package clusterstatus

import (
	"fmt"
	"time"

	cluster_status_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/cluster_status"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
)

var (
	containerScope = log.RegisterScope("influxdb_repo_container_measurement", "InfluxDB repository container measurement", 0)
)

// ContainerRepository is used to operate node measurement of cluster_status database
type ContainerRepository struct {
	influxDB *influxdb.InfluxDBRepository
}

// IsTag checks the column is tag or not
func (containerRepository *ContainerRepository) IsTag(column string) bool {
	for _, tag := range cluster_status_entity.ContainerTags {
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

// ListAlamedaContainers list predicted containers
func (containerRepository *ContainerRepository) ListAlamedaContainers() ([]*datahub_v1alpha1.Pod, error) {
	podList := []*datahub_v1alpha1.Pod{}
	// SELECT * FROM container WHERE is_deleted=false AND is_alameda=true GROUP BY namespace,pod_name,alameda_scaler_namespace,alameda_scaler_name
	cmd := fmt.Sprintf("SELECT * FROM %s WHERE \"%s\"=%s AND \"%s\"=%s GROUP BY \"%s\",\"%s\",\"%s\",\"%s\"",
		string(Container), string(cluster_status_entity.ContainerIsAlameda), "true",
		string(cluster_status_entity.ContainerIsDeleted), "false",
		string(cluster_status_entity.ContainerNamespace), string(cluster_status_entity.ContainerPodName),
		string(cluster_status_entity.ContainerAlamedaScalerNamespace), string(cluster_status_entity.ContainerAlamedaScalerName))

	if results, err := containerRepository.influxDB.QueryDB(cmd, string(influxdb.ClusterStatus)); err == nil {
		for _, result := range results {
			for _, ser := range result.Series {
				podName := ser.Tags[string(cluster_status_entity.ContainerPodName)]
				contanerNS := ser.Tags[string(cluster_status_entity.ContainerNamespace)]
				podList = append(podList, &datahub_v1alpha1.Pod{
					NamespacedName: &datahub_v1alpha1.NamespacedName{
						Name:      podName,
						Namespace: contanerNS,
					},
					IsAlameda: true,
					AlamedaResource: &datahub_v1alpha1.NamespacedName{
						Name:      podName,
						Namespace: contanerNS,
					},
				})
			}
			containerScope.Infof(fmt.Sprintf(""))
		}
		return podList, nil
	} else {
		return podList, err
	}
}

// CreateContainers add containers information container measurement
func (containerRepository *ContainerRepository) CreateContainers(pods []*datahub_v1alpha1.Pod) error {
	points := []*influxdb_client.Point{}
	for _, pod := range pods {
		podNS := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()
		containers := pod.GetContainers()
		isAlamedaPod := pod.GetIsAlameda()

		for _, container := range containers {
			// due to containers in pod may have same tags as the following, sleep for a short while to prevent data point overridden
			time.Sleep(1 * time.Microsecond)
			tags := map[string]string{
				string(cluster_status_entity.ContainerNamespace): podNS,
				string(cluster_status_entity.ContainerPodName):   podName,
				string(cluster_status_entity.ContainerNodeName):  pod.GetNodeName(),
			}
			fields := map[string]interface{}{
				string(cluster_status_entity.ContainerName):      container.GetName(),
				string(cluster_status_entity.ContainerIsDeleted): false,
				string(cluster_status_entity.ContainerIsAlameda): isAlamedaPod,
				string(cluster_status_entity.ContainerPolicy):    pod.GetPolicy(),
			}
			if isAlamedaPod {
				tags[string(cluster_status_entity.ContainerAlamedaScalerNamespace)] = pod.GetAlamedaResource().GetNamespace()
				tags[string(cluster_status_entity.ContainerAlamedaScalerName)] = pod.GetAlamedaResource().GetName()
			}
			for _, metricData := range container.GetLimitResource() {
				if data := metricData.GetData(); len(data) == 1 {
					switch metricData.GetMetricType() {
					case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
						fields[string(cluster_status_entity.ContainerResourceLimitCPU)] = data[0].NumValue
					case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
						fields[string(cluster_status_entity.ContainerResourceLimitMemory)] = data[0].NumValue
					}
				}
			}
			for _, metricData := range container.GetRequestResource() {
				if data := metricData.GetData(); len(data) == 1 {
					switch metricData.GetMetricType() {
					case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
						fields[string(cluster_status_entity.ContainerResourceRequestCPU)] = data[0].NumValue
					case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
						fields[string(cluster_status_entity.ContainerResourceRequestMemory)] = data[0].NumValue
					}
				}
			}

			if pt, err := influxdb_client.NewPoint(string(Container), tags, fields, time.Now()); err == nil {
				points = append(points, pt)
			} else {
				scope.Error(err.Error())
			}
		}
	}
	containerRepository.influxDB.WritePoints(points, influxdb_client.BatchPointsConfig{
		Database: string(influxdb.ClusterStatus),
	})
	return nil
}

// UpdateContainers updates containers information container measurement
func (containerRepository *ContainerRepository) UpdateContainers(pods []*datahub_v1alpha1.Pod) error {
	return nil
}
