package clusterstatus

import (
	"errors"
	"fmt"
	"strings"

	cluster_status_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/cluster_status"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
)

var (
	containerScope = log.RegisterScope("cluster_status_db_container_measurement", "cluster_status DB container measurement", 0)
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
					AlamedaScaler: &datahub_v1alpha1.NamespacedName{
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
			tags := map[string]string{
				string(cluster_status_entity.ContainerNamespace): podNS,
				string(cluster_status_entity.ContainerPodName):   podName,
				string(cluster_status_entity.ContainerNodeName):  pod.GetNodeName(),
				string(cluster_status_entity.ContainerName):      container.GetName(),
			}
			fields := map[string]interface{}{
				string(cluster_status_entity.ContainerIsDeleted): false,
				string(cluster_status_entity.ContainerIsAlameda): isAlamedaPod,
				string(cluster_status_entity.ContainerPolicy):    pod.GetPolicy(),
			}
			if isAlamedaPod {
				tags[string(cluster_status_entity.ContainerAlamedaScalerNamespace)] = pod.GetAlamedaScaler().GetNamespace()
				tags[string(cluster_status_entity.ContainerAlamedaScalerName)] = pod.GetAlamedaScaler().GetName()
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

			if pt, err := influxdb_client.NewPoint(string(Container), tags, fields, influxdb.ZeroTime); err == nil {
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

// DeleteContainers set containers' field is_deleted to true into container measurement
func (containerRepository *ContainerRepository) DeleteContainers(pods []*datahub_v1alpha1.Pod) error {

	var (
		err error

		containersEntityBeforeDelete = make([]*cluster_status_entity.ContainerEntity, 0)

		pointsToDelete = make([]*influxdb_client.Point, 0)
	)

	containersEntityBeforeDelete, err = containerRepository.ListPodsContainers(pods)
	if err != nil {
		return errors.New("delete containers failed: " + err.Error())
	}
	for _, containerEntity := range containersEntityBeforeDelete {
		entity := *containerEntity

		trueString := string("true")
		entity.IsDeleted = &trueString
		point, err := entity.InfluxDBPoint(string(Container))
		if err != nil {
			return errors.New("delete containers failed: " + err.Error())
		}

		pointsToDelete = append(pointsToDelete, point)
	}

	containerRepository.influxDB.WritePoints(pointsToDelete, influxdb_client.BatchPointsConfig{
		Database: string(influxdb.ClusterStatus),
	})
	return nil
}

// ListPodsContainers list containers information container measurement
func (containerRepository *ContainerRepository) ListPodsContainers(pods []*datahub_v1alpha1.Pod) ([]*cluster_status_entity.ContainerEntity, error) {

	var (
		cmd                 = ""
		cmdSelectString     = ""
		cmdTagsFilterString = ""
		containerEntities   = make([]*cluster_status_entity.ContainerEntity, 0)
	)

	if len(pods) == 0 {
		return containerEntities, nil
	}

	cmdSelectString = fmt.Sprintf(`select * from "%s" `, Container)
	for _, pod := range pods {

		var (
			namespace = ""
			podName   = ""
		)

		if pod.GetNamespacedName() != nil {
			namespace = pod.GetNamespacedName().GetNamespace()
			podName = pod.GetNamespacedName().GetName()
		}

		cmdTagsFilterString += fmt.Sprintf(`("%s" = '%s' and "%s" = '%s') or `,
			cluster_status_entity.ContainerNamespace, namespace,
			cluster_status_entity.ContainerPodName, podName,
		)
	}
	cmdTagsFilterString = strings.TrimSuffix(cmdTagsFilterString, "or ")

	cmd = fmt.Sprintf("%s where %s", cmdSelectString, cmdTagsFilterString)
	results, err := containerRepository.influxDB.QueryDB(cmd, string(influxdb.ClusterStatus))
	if err != nil {
		return containerEntities, errors.New("list containers' entity failed: " + err.Error())
	}

	rows := influxdb.PackMap(results)
	for _, row := range rows {
		for _, data := range row.Data {
			entity := cluster_status_entity.NewContainerEntityFromMap(data)
			containerEntities = append(containerEntities, &entity)
		}
	}

	return containerEntities, nil
}
