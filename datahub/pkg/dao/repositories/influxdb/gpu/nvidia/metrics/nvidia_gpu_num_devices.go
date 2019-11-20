package metrics

import (
	EntityInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/gpu/nvidia/metrics"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	"github.com/pkg/errors"
)

type NumDevicesRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNumDevicesRepositoryWithConfig(cfg InternalInflux.Config) *NumDevicesRepository {
	return &NumDevicesRepository{
		influxDB: InternalInflux.NewClient(&cfg),
	}
}

func (r *NumDevicesRepository) ListNumDevices(host, instance string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuMetric.NumDevicesEntity, error) {
	entities := make([]*EntityInfluxGpuMetric.NumDevicesEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    NumDevices,
		GroupByTags:    []string{"host"},
	}

	influxdbStatement.AppendWhereClause("AND", "host", "=", host)
	influxdbStatement.AppendWhereClause("AND", "instance", "=", instance)
	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Gpu))
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu num devices")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				entity := EntityInfluxGpuMetric.NewNumDevicesEntityFromMap(group.GetRow(j))
				entities = append(entities, &entity)
			}
		}
	}

	return entities, nil
}
