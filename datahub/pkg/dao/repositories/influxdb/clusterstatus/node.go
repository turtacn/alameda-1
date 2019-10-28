package clusterstatus

import (
	"fmt"
	EntityInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strings"
)

type NodeRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func (nodeRepository *NodeRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxClusterStatus.NodeTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

func NewNodeRepository(influxDBCfg *InternalInflux.Config) *NodeRepository {
	return &NodeRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

// AddAlamedaNodes add node information to database
func (nodeRepository *NodeRepository) AddAlamedaNodes(alamedaNodes []*ApiResources.Node) error {
	points := []*InfluxClient.Point{}
	for _, alamedaNode := range alamedaNodes {
		isInCluster := true
		startTime := alamedaNode.StartTime.GetSeconds()
		entity := EntityInfluxClusterStatus.NodeEntity{
			Time:        InternalInflux.ZeroTime,
			Name:        &alamedaNode.Name,
			IsInCluster: &isInCluster,
			CreatedTime: &startTime,
		}
		if nodeCapacity := alamedaNode.GetCapacity(); nodeCapacity != nil {
			entity.CPUCores = &nodeCapacity.CpuCores
			entity.MemoryBytes = &nodeCapacity.MemoryBytes
		}
		if nodeProvider := alamedaNode.GetProvider(); nodeProvider != nil {
			entity.IOProvider = &nodeProvider.Provider
			entity.IOInstanceType = &nodeProvider.InstanceType
			entity.IORegion = &nodeProvider.Region
			entity.IOZone = &nodeProvider.Zone
			entity.IOOS = &nodeProvider.Os
			entity.IORole = &nodeProvider.Role
			entity.IOInstanceID = &nodeProvider.InstanceId
			entity.IOStorageSize = &nodeProvider.StorageSize
		}
		if pt, err := entity.InfluxDBPoint(string(Node)); err == nil {
			points = append(points, pt)
		} else {
			scope.Error(err.Error())
		}
	}
	err := nodeRepository.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		return errors.Wrapf(err, "add alameda nodes failed: %s", err.Error())
	}
	return nil
}

func (nodeRepository *NodeRepository) RemoveAlamedaNodes(alamedaNodes []*ApiResources.Node) error {
	hasErr := false
	errMsg := ""
	for _, alamedaNode := range alamedaNodes {
		cmd := fmt.Sprintf("DROP SERIES FROM %s WHERE \"%s\"='%s'",
			string(Node), string(EntityInfluxClusterStatus.NodeName), alamedaNode.Name)
		_, err := nodeRepository.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
		if err != nil {
			hasErr = true
			errMsg += errMsg + err.Error()
		}
	}
	if hasErr {
		return fmt.Errorf(errMsg)
	}
	return nil
}

func (nodeRepository *NodeRepository) ListAlamedaNodes(timeRange *ApiCommon.TimeRange) ([]*EntityInfluxClusterStatus.NodeEntity, error) {

	nodeEntities := []*EntityInfluxClusterStatus.NodeEntity{}
	nodeCreatePeriodCondition := nodeRepository.getNodeCreatePeriodCondition(timeRange)

	cmd := fmt.Sprintf("SELECT * FROM %s WHERE \"%s\"=%s %s",
		string(Node), string(EntityInfluxClusterStatus.NodeInCluster), "true", nodeCreatePeriodCondition)

	scope.Debug(fmt.Sprintf("Query nodes in cluster: %s", cmd))
	results, err := nodeRepository.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return nodeEntities, errors.Wrap(err, "list alameda nodes from influxdb failed")
	}

	if len(results) == 1 && len(results[0].Series) > 0 {

		influxdbRows := InternalInflux.PackMap(results)
		for _, influxdbRow := range influxdbRows {
			for _, data := range influxdbRow.Data {
				nodeEntity := EntityInfluxClusterStatus.NewNodeEntityFromMap(data)
				nodeEntities = append(nodeEntities, &nodeEntity)
			}
		}
	}

	return nodeEntities, nil
}

func (nodeRepository *NodeRepository) ListNodes(request DaoClusterStatus.ListNodesRequest) ([]*EntityInfluxClusterStatus.NodeEntity, error) {

	nodeEntities := []*EntityInfluxClusterStatus.NodeEntity{}

	whereClause := nodeRepository.buildInfluxQLWhereClauseFromRequest(request)
	cmd := fmt.Sprintf("SELECT * FROM %s %s", string(Node), whereClause)
	scope.Debug(fmt.Sprintf("Query nodes in cluster: %s", cmd))
	results, err := nodeRepository.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return nodeEntities, errors.Wrap(err, "list nodes from influxdb failed")
	}

	influxdbRows := InternalInflux.PackMap(results)
	for _, influxdbRow := range influxdbRows {
		for _, data := range influxdbRow.Data {
			nodeEntity := EntityInfluxClusterStatus.NewNodeEntityFromMap(data)
			nodeEntities = append(nodeEntities, &nodeEntity)
		}
	}

	return nodeEntities, nil
}

func (nodeRepository *NodeRepository) buildInfluxQLWhereClauseFromRequest(request DaoClusterStatus.ListNodesRequest) string {

	var (
		whereClause string
		conditions  string
	)

	conditions += fmt.Sprintf("\"%s\" = %t", EntityInfluxClusterStatus.NodeInCluster, request.InCluster)

	statementFilteringNodes := ""
	for _, nodeName := range request.NodeNames {
		statementFilteringNodes += fmt.Sprintf(`"%s" = '%s' OR `, EntityInfluxClusterStatus.NodeName, nodeName)
	}
	statementFilteringNodes = strings.TrimSuffix(statementFilteringNodes, "OR ")
	if statementFilteringNodes != "" {
		conditions = fmt.Sprintf("(%s) AND (%s)", conditions, statementFilteringNodes)
	}

	whereClause = fmt.Sprintf("WHERE %s", conditions)

	return whereClause
}

func (nodeRepository *NodeRepository) getNodeCreatePeriodCondition(timeRange *ApiCommon.TimeRange) string {
	if timeRange == nil {
		return ""
	}

	var start int64 = 0
	var end int64 = 0

	if timeRange.StartTime != nil {
		start = timeRange.StartTime.Seconds
	}

	if timeRange.EndTime != nil {
		end = timeRange.EndTime.Seconds
	}

	if start == 0 && end == 0 {
		return ""
	} else if start == 0 && end != 0 {
		period := fmt.Sprintf(`AND "create_time" < %d`, end)
		return period
	} else if start != 0 && end == 0 {
		period := fmt.Sprintf(`AND "create_time" >= %d`, start)
		return period
	} else if start != 0 && end != 0 {
		period := fmt.Sprintf(`AND "create_time" >= %d AND "create_time" < %d`, start, end)
		return period
	}

	return ""
}
