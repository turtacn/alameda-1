package clusterstatus

import (
	"fmt"
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	InternalCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strings"
)

type NodeRepository struct {
	influxDB *InternalInflux.InfluxClient
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

func (p *NodeRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxCluster.NodeTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

func (p *NodeRepository) CreateNodes(nodes []*DaoClusterTypes.Node) error {
	points := make([]*InfluxClient.Point, 0)

	for _, node := range nodes {
		entity := EntityInfluxCluster.NodeEntity{
			Time:        InternalInflux.ZeroTime,
			Name:        node.ObjectMeta.Name,
			ClusterName: node.ObjectMeta.ClusterName,
			Uid:         node.ObjectMeta.Uid,
			CreateTime:  node.CreateTime.GetSeconds(),
		}
		if node.Capacity != nil {
			entity.CPUCores = node.Capacity.CpuCores
			entity.MemoryBytes = node.Capacity.MemoryBytes
			entity.NetworkMbps = node.Capacity.NetworkMegabitsPerSecond
		}
		if nodeSpec := node.AlamedaNodeSpec; nodeSpec != nil {
			if nodeSpec.Provider != nil {
				entity.IOProvider = node.AlamedaNodeSpec.Provider.Provider
				entity.IOInstanceType = node.AlamedaNodeSpec.Provider.InstanceType
				entity.IORegion = node.AlamedaNodeSpec.Provider.Region
				entity.IOZone = node.AlamedaNodeSpec.Provider.Zone
				entity.IOOS = node.AlamedaNodeSpec.Provider.Os
				entity.IORole = node.AlamedaNodeSpec.Provider.Role
				entity.IOInstanceID = node.AlamedaNodeSpec.Provider.InstanceId
				entity.IOStorageSize = node.AlamedaNodeSpec.Provider.StorageSize
			}
		}

		// Add to influx point list
		if pt, err := entity.BuildInfluxPoint(string(Node)); err == nil {
			points = append(points, pt)
		} else {
			scope.Error(err.Error())
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

func (p *NodeRepository) ListNodes(request DaoClusterTypes.ListNodesRequest) ([]*DaoClusterTypes.Node, error) {
	nodes := make([]*DaoClusterTypes.Node, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Node,
		GroupByTags:    []string{string(EntityInfluxCluster.NodeClusterName)},
	}

	// Build influx query command
	for _, objectMeta := range request.ObjectMeta {
		conditionList := make([]string, 0)

		metaCondition := p.genObjectMetaCondition(objectMeta)
		if metaCondition != "" {
			conditionList = append(conditionList, metaCondition)
		}

		createCondition := p.genCreatePeriodCondition(request.QueryCondition)
		if createCondition != "" {
			conditionList = append(conditionList, createCondition)
		}

		condition := strings.Join(conditionList, " AND ")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	if len(request.ObjectMeta) == 0 {
		statement.AppendWhereClauseDirectly("AND", p.genCreatePeriodCondition(request.QueryCondition))
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Node, 0), errors.Wrap(err, "failed to list nodes")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				node := DaoClusterTypes.NewNode()
				node.Initialize(row)
				nodes = append(nodes, node)
			}
		}
	}

	return nodes, nil
}

func (p *NodeRepository) DeleteNodes(alamedaNodes []*ApiResources.Node) error {
	hasErr := false
	errMsg := ""
	for _, alamedaNode := range alamedaNodes {
		cmd := fmt.Sprintf("DROP SERIES FROM %s WHERE \"%s\"='%s'",
			string(Node), string(EntityInfluxCluster.NodeName), alamedaNode.ObjectMeta.Name)
		_, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
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

func (p *NodeRepository) genObjectMetaCondition(objectMeta Metadata.ObjectMeta) string {
	condition := ""
	keyList := objectMeta.GenerateKeyList()
	valueList := objectMeta.GenerateValueList()
	for i := 0; i < len(keyList); i++ {
		if valueList[i] != "" {
			condition += fmt.Sprintf("\"%s\"='%s' AND ", keyList[i], valueList[i])
		}
	}
	condition = strings.TrimSuffix(condition, "AND ")
	return condition
}

func (p *NodeRepository) genCreatePeriodCondition(query InternalCommon.QueryCondition) string {
	if query.StartTime != nil && query.EndTime != nil {
		return fmt.Sprintf("\"%s\">=%d AND \"%s\"<%d", EntityInfluxCluster.NodeCreateTime, query.StartTime.Unix(), EntityInfluxCluster.NodeCreateTime, query.EndTime.Unix())
	} else if query.StartTime != nil && query.EndTime == nil {
		return fmt.Sprintf("\"%s\">=%d", EntityInfluxCluster.NodeCreateTime, query.StartTime.Unix())
	} else if query.StartTime == nil && query.EndTime != nil {
		return fmt.Sprintf("\"%s\"<%d", EntityInfluxCluster.NodeCreateTime, query.EndTime.Unix())
	} else {
		return ""
	}
}
