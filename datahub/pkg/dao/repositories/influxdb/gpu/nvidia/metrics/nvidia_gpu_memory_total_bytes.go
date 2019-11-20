package metrics

import (
	EntityInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/gpu/nvidia/metrics"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	"github.com/pkg/errors"
)

type MemoryTotalBytesRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewMemoryTotalBytesRepositoryWithConfig(cfg InternalInflux.Config) *MemoryTotalBytesRepository {
	return &MemoryTotalBytesRepository{
		influxDB: InternalInflux.NewClient(&cfg),
	}
}

func (r *MemoryTotalBytesRepository) ListMemoryTotalBytes(host, minorNumber string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuMetric.MemoryTotalBytesEntity, error) {
	entities := make([]*EntityInfluxGpuMetric.MemoryTotalBytesEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    MemoryTotalBytes,
		GroupByTags:    []string{"host", "uuid"},
	}

	influxdbStatement.AppendWhereClause("AND", "host", "=", host)
	influxdbStatement.AppendWhereClause("AND", "minor_number", "=", minorNumber)
	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Gpu))
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu memory total bytes")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				entity := EntityInfluxGpuMetric.NewMemoryTotalBytesEntityFromMap(group.GetRow(j))
				entities = append(entities, &entity)
			}
		}
	}

	return entities, nil
}
