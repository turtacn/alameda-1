package metric

import (
	EntityInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/gpu/nvidia/metric"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	"github.com/pkg/errors"
)

type DutyCycleRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewDutyCycleRepositoryWithConfig(cfg InternalInflux.Config) *DutyCycleRepository {
	return &DutyCycleRepository{
		influxDB: InternalInflux.NewClient(&cfg),
	}
}

func (r *DutyCycleRepository) ListMetrics(host, minorNumber string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuMetric.DutyCycleEntity, error) {
	entities := make([]*EntityInfluxGpuMetric.DutyCycleEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    DutyCycle,
		GroupByTags:    []string{"host", "uuid"},
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.AppendWhereClause("host", "=", host)
	influxdbStatement.AppendWhereClause("minor_number", "=", minorNumber)
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Gpu))
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu duty cycle")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				entity := EntityInfluxGpuMetric.NewDutyCycleEntityFromMap(group.GetRow(j))
				entities = append(entities, &entity)
			}
		}
	}

	return entities, nil
}