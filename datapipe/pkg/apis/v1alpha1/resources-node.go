package v1alpha1

import (
	entity "github.com/containers-ai/alameda/datapipe/pkg/entities/influxdb/cluster_status/node"
	apiServer "github.com/containers-ai/alameda/datapipe/pkg/repositories/apiserver"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	commonAPI "github.com/containers-ai/api/common"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/rpc/status"
	"strconv"
)

func (s *ServiceV1alpha1) CreateNodesImpl(in *datahub_v1alpha1.CreateAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	rowDataList := make([]*commonAPI.WriteRawdata, 0)

	rowData := &commonAPI.WriteRawdata{
		Database:    entity.NodeDatabaseName,
		Table:       entity.NodeMeasurementName,
		Columns:     entity.NodeColumns,
		Rows:        make([]*commonAPI.Row, 0),
		ColumnTypes: entity.NodeColumnsTypes,
		DataTypes:   entity.NodeDataTypes,
	}

	indexMap := entity.NodeColIndexMap

	for _, node := range in.GetAlamedaNodes() {
		nodeName := node.GetName()
		capacity := node.GetCapacity()
		startTime := node.GetStartTime().GetSeconds()

		values := make([]string, len(entity.NodeColumns))
		values[indexMap[entity.NodeName]] = nodeName
		values[indexMap[entity.NodeGroup]] = ""
		values[indexMap[entity.NodeInCluster]] = ""
		values[indexMap[entity.NodeCPUCores]] = strconv.FormatInt(capacity.GetCpuCores(), 10)
		values[indexMap[entity.NodeMemoryBytes]] = strconv.FormatInt(capacity.GetMemoryBytes(), 10)
		values[indexMap[entity.NodeCreateTime]] = strconv.FormatInt(startTime, 10)

		row := &commonAPI.Row{
			Time:   &timestamp.Timestamp{Seconds: 0},
			Values: values,
		}

		rowData.Rows = append(rowData.Rows, row)
	}

	rowDataList = append(rowDataList, rowData)

	retStatus, err := apiServer.WriteRawdata(s.Target, commonAPI.DatabaseType_INFLUXDB, rowDataList)
	return retStatus, err
}
