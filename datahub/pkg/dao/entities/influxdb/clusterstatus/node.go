package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"time"
)

const (
	NodeName        influxdb.Tag = "name" // NodeName is the name of node
	NodeClusterName influxdb.Tag = "cluster_name"
	NodeUid         influxdb.Tag = "uid"

	NodeCreateTime     influxdb.Field = "create_time"
	NodeCPUCores       influxdb.Field = "node_cpu_cores"    // NodeCPUCores is the amount of cores in node
	NodeMemoryBytes    influxdb.Field = "node_memory_bytes" // NodeMemoryBytes is the amount of memory bytes in node
	NodeNetworkMbps    influxdb.Field = "node_network_mbps" // NodeNetworkMbps is mega bits per second
	NodeIOProvider     influxdb.Field = "io_provider"       // Cloud service provider
	NodeIOInstanceType influxdb.Field = "io_instance_type"
	NodeIORegion       influxdb.Field = "io_region"
	NodeIOZone         influxdb.Field = "io_zone"
	NodeIOOS           influxdb.Field = "io_os"
	NodeIORole         influxdb.Field = "io_role"
	NodeIOInstanceID   influxdb.Field = "io_instance_id"
	NodeIOStorageSize  influxdb.Field = "io_storage_size"
)

var (
	// NodeTags list tags of node measurement
	NodeTags = []influxdb.Tag{
		NodeName,
		NodeClusterName,
		NodeUid,
	}

	// NodeFields list fields of node measurement
	NodeFields = []influxdb.Field{
		NodeCreateTime,
		NodeCPUCores,
		NodeMemoryBytes,
		NodeNetworkMbps,
		NodeIOProvider,
		NodeIOInstanceType,
		NodeIORegion,
		NodeIOZone,
		NodeIOOS,
		NodeIORole,
		NodeIOInstanceID,
		NodeIOStorageSize,
	}
)

// NodeEntity is entity in database
type NodeEntity struct {
	Time        time.Time
	Name        string
	ClusterName string
	Uid         string

	CreateTime     int64
	CPUCores       int64
	MemoryBytes    int64
	NetworkMbps    int64
	IOProvider     string
	IOInstanceType string
	IORegion       string
	IOZone         string
	IOOS           string
	IORole         string
	IOInstanceID   string
	IOStorageSize  int64
}

// NewNodeEntityFromMap Build entity from map
func NewNodeEntity(data map[string]string) *NodeEntity {
	entity := NodeEntity{}

	tempTimestamp, _ := utils.ParseTime(data["time"])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(NodeName)]; exist {
		entity.Name = value
	}
	if value, exist := data[string(NodeClusterName)]; exist {
		entity.ClusterName = value
	}
	if value, exist := data[string(NodeUid)]; exist {
		entity.Uid = value
	}

	// InfluxDB fields
	if value, exist := data[string(NodeCreateTime)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.CreateTime = valueInt64
	}
	if value, exist := data[string(NodeCPUCores)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.CPUCores = valueInt64
	}
	if value, exist := data[string(NodeMemoryBytes)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.MemoryBytes = valueInt64
	}
	if value, exist := data[string(NodeNetworkMbps)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.NetworkMbps = valueInt64
	}
	if value, exist := data[string(NodeIOProvider)]; exist {
		entity.IOProvider = value
	}
	if value, exist := data[string(NodeIOInstanceType)]; exist {
		entity.IOInstanceType = value
	}
	if value, exist := data[string(NodeIORegion)]; exist {
		entity.IORegion = value
	}
	if value, exist := data[string(NodeIOZone)]; exist {
		entity.IOZone = value
	}
	if value, exist := data[string(NodeIOOS)]; exist {
		entity.IOOS = value
	}
	if value, exist := data[string(NodeIORole)]; exist {
		entity.IORole = value
	}
	if value, exist := data[string(NodeIOInstanceID)]; exist {
		entity.IOInstanceID = value
	}
	if value, exist := data[string(NodeIOStorageSize)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.IOStorageSize = valueInt64
	}

	return &entity
}

func (p *NodeEntity) BuildInfluxPoint(measurement string) (*InfluxClient.Point, error) {
	// Pack influx tags
	tags := map[string]string{
		string(NodeName):        p.Name,
		string(NodeClusterName): p.ClusterName,
		string(NodeUid):         p.Uid,
	}

	// Pack influx fields
	fields := map[string]interface{}{
		string(NodeCreateTime):     p.CreateTime,
		string(NodeCPUCores):       p.CPUCores,
		string(NodeMemoryBytes):    p.MemoryBytes,
		string(NodeNetworkMbps):    p.NetworkMbps,
		string(NodeIOProvider):     p.IOProvider,
		string(NodeIOInstanceType): p.IOInstanceType,
		string(NodeIORegion):       p.IORegion,
		string(NodeIOZone):         p.IOZone,
		string(NodeIOOS):           p.IOOS,
		string(NodeIORole):         p.IORole,
		string(NodeIOInstanceID):   p.IOInstanceID,
		string(NodeIOStorageSize):  p.IOStorageSize,
	}

	return InfluxClient.NewPoint(measurement, tags, fields, p.Time)
}
