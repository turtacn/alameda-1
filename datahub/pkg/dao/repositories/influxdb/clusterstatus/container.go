package clusterstatus

import (
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strings"
)

type ContainerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewContainerRepository(influxDBCfg InternalInflux.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (p *ContainerRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxCluster.ContainerTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

// CreateContainers add containers information container measurement
func (p *ContainerRepository) CreateContainers(containers map[string][]*DaoClusterTypes.Container) error {
	points := make([]*InfluxClient.Point, 0)

	for _, cnts := range containers {
		for _, cnt := range cnts {
			entity := cnt.BuildEntity()

			// Add to influx point list
			if pt, err := entity.BuildInfluxPoint(string(Container)); err == nil {
				points = append(points, pt)
			} else {
				scope.Error(err.Error())
			}
		}
	}

	// Batch write influxdb data points
	err := p.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

// ListContainers list predicted containers have relation with arguments
func (p *ContainerRepository) ListContainers(request DaoClusterTypes.ListContainersRequest) (map[string][]*DaoClusterTypes.Container, error) {
	containerMap := make(map[string][]*DaoClusterTypes.Container, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Container,
		GroupByTags:    []string{string(EntityInfluxCluster.ContainerPodName), string(EntityInfluxCluster.ContainerNamespace), string(EntityInfluxCluster.ContainerNodeName), string(EntityInfluxCluster.ContainerClusterName)},
	}

	// Build influx query command
	for _, containerMeta := range request.ContainerObjectMeta {
		keyList := make([]string, 0)
		valueList := make([]string, 0)

		keyList = append(keyList, string(EntityInfluxCluster.ContainerName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerPodName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerNamespace))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerNodeName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerClusterName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerTopControllerName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerAlamedaScalerName))

		valueList = append(valueList, containerMeta.Name)
		valueList = append(valueList, containerMeta.PodName)
		valueList = append(valueList, containerMeta.Namespace)
		valueList = append(valueList, containerMeta.NodeName)
		valueList = append(valueList, containerMeta.ClusterName)
		valueList = append(valueList, containerMeta.TopControllerName)
		valueList = append(valueList, containerMeta.AlamedaScalerName)

		if containerMeta.TopControllerKind != "" && containerMeta.TopControllerKind != ApiResources.Kind_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ContainerTopControllerKind))
			valueList = append(valueList, containerMeta.TopControllerKind)
		}

		if containerMeta.AlamedaScalerScalingTool != "" && containerMeta.AlamedaScalerScalingTool != ApiResources.ScalingTool_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ContainerAlamedaScalerScalingTool))
			valueList = append(valueList, containerMeta.AlamedaScalerScalingTool)
		}

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make(map[string][]*DaoClusterTypes.Container, 0), errors.Wrap(err, "failed to list containers")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			row := group.GetRow(0)
			clusterNamespacePodName := p.ClusterNamespacePodName(row)
			containerMap[clusterNamespacePodName] = make([]*DaoClusterTypes.Container, 0)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				container := DaoClusterTypes.NewContainer()
				container.Initialize(EntityInfluxCluster.NewContainerEntity(row))
				containerMap[clusterNamespacePodName] = append(containerMap[clusterNamespacePodName], container)
			}
		}
	}

	return containerMap, nil
}

// DeleteContainers set containers' field is_deleted to true into container measurement
func (p *ContainerRepository) DeleteContainers(request DaoClusterTypes.DeleteContainersRequest) error {
	statement := InternalInflux.Statement{
		Measurement: Container,
	}

	if !p.influxDB.MeasurementExist(string(RepoInflux.ClusterStatus), string(Container)) {
		return nil
	}

	// Build influx drop command
	for _, containerMeta := range request.ContainerObjectMeta {
		keyList := make([]string, 0)
		valueList := make([]string, 0)

		keyList = append(keyList, string(EntityInfluxCluster.ContainerName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerPodName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerNamespace))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerNodeName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerClusterName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerTopControllerName))
		keyList = append(keyList, string(EntityInfluxCluster.ContainerAlamedaScalerName))

		valueList = append(valueList, containerMeta.Name)
		valueList = append(valueList, containerMeta.PodName)
		valueList = append(valueList, containerMeta.Namespace)
		valueList = append(valueList, containerMeta.NodeName)
		valueList = append(valueList, containerMeta.ClusterName)
		valueList = append(valueList, containerMeta.TopControllerName)
		valueList = append(valueList, containerMeta.AlamedaScalerName)

		if containerMeta.TopControllerKind != "" && containerMeta.TopControllerKind != ApiResources.Kind_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ContainerTopControllerKind))
			valueList = append(valueList, containerMeta.TopControllerKind)
		}

		if containerMeta.AlamedaScalerScalingTool != "" && containerMeta.AlamedaScalerScalingTool != ApiResources.ScalingTool_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ContainerAlamedaScalerScalingTool))
			valueList = append(valueList, containerMeta.AlamedaScalerScalingTool)
		}

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	cmd := statement.BuildDropCmd()

	_, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return errors.Wrap(err, "failed to delete containers")
	}

	return nil
}

func (p *ContainerRepository) ClusterNamespacePodName(values map[string]string) string {
	valueList := make([]string, 0)

	if value, ok := values[string(EntityInfluxCluster.ContainerClusterName)]; ok {
		if value != "" {
			valueList = append(valueList, value)
		}
	}
	if value, ok := values[string(EntityInfluxCluster.ContainerNamespace)]; ok {
		if value != "" {
			valueList = append(valueList, value)
		}
	}
	if value, ok := values[string(EntityInfluxCluster.ContainerPodName)]; ok {
		if value != "" {
			valueList = append(valueList, value)
		}
	}

	if len(valueList) > 0 {
		return strings.Join(valueList, "/")
	}

	return ""
}
