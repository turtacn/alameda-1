package metric

import (
	EntityInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/gpu/nvidia/metric"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	"github.com/pkg/errors"
)

type PowerUsageMilliWattsRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewPowerUsageMilliWattsRepositoryWithConfig(cfg InternalInflux.Config) *PowerUsageMilliWattsRepository {
	return &PowerUsageMilliWattsRepository{
		influxDB: InternalInflux.NewClient(&cfg),
	}
}

func (r *PowerUsageMilliWattsRepository) ListMetrics(host, minorNumber string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuMetric.PowerUsageMilliWattsEntity, error) {
	entities := make([]*EntityInfluxGpuMetric.PowerUsageMilliWattsEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    PowerUsageMilliWatts,
		GroupByTags:    []string{"host"},
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.AppendWhereClause("host", "=", host)
	influxdbStatement.AppendWhereClause("minor_number", "=", minorNumber)
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Gpu))
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu power usage milli watts")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				entity := EntityInfluxGpuMetric.NewPowerUsageMilliWattsEntityFromMap(group.GetRow(j))
				entities = append(entities, &entity)
			}
		}
	}

	return entities, nil
}
