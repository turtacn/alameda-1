package metrics

import (
	"fmt"
	EntityInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/gpu/nvidia/metrics"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	Client "github.com/influxdata/influxdb/client/v2"
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
	steps := int(condition.StepTime.Seconds())
	if steps == 0 || steps == 30 {
		return r.read(host, minorNumber, condition)
	} else {
		return r.steps(host, minorNumber, condition)
	}
}

func (r *DutyCycleRepository) read(host, minorNumber string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuMetric.DutyCycleEntity, error) {
	entities := make([]*EntityInfluxGpuMetric.DutyCycleEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    DutyCycle,
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
		return entities, errors.Wrap(err, "failed to list nvidia gpu duty cycle")
	}

	entities = r.genEntities(response)

	return entities, nil
}

func (r *DutyCycleRepository) steps(host, minorNumber string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuMetric.DutyCycleEntity, error) {
	entities := make([]*EntityInfluxGpuMetric.DutyCycleEntity, 0)

	response, err := r.last(host, minorNumber, condition)
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu duty cycle with last")
	}
	lastEntities := r.genEntities(response)

	response, err = r.max(host, minorNumber, condition)
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu duty cycle with max")
	}
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			entityPtr := &EntityInfluxGpuMetric.DutyCycleEntity{}
			group := result.GetGroup(i)
			gpuId := group.Tags["uuid"]
			found := false

			for _, entityPtr = range lastEntities {
				if *entityPtr.Uuid == gpuId {
					found = true
					break
				}
			}

			if found {
				for j := 0; j < group.GetRowNum(); j++ {
					row := group.GetRow(j)
					if row["max_value"] != "" {
						entityMap := make(map[string]string)
						entityMap[EntityInfluxGpuMetric.DutyCycleTime] = row["time"]
						entityMap[EntityInfluxGpuMetric.DutyCycleHost] = *entityPtr.Host
						entityMap[EntityInfluxGpuMetric.DutyCycleInstance] = *entityPtr.Instance
						entityMap[EntityInfluxGpuMetric.DutyCycleJob] = *entityPtr.Job
						entityMap[EntityInfluxGpuMetric.DutyCycleName] = *entityPtr.Name
						entityMap[EntityInfluxGpuMetric.DutyCycleUuid] = *entityPtr.Uuid

						entityMap[EntityInfluxGpuMetric.DutyCycleMinorNumber] = *entityPtr.MinorNumber
						entityMap[EntityInfluxGpuMetric.DutyCycleValue] = row["max_value"]

						entity := EntityInfluxGpuMetric.NewDutyCycleEntityFromMap(entityMap)
						entities = append(entities, &entity)
					}
				}
			}
		}
	}

	return entities, nil
}

func (r *DutyCycleRepository) last(host, minorNumber string, condition *DBCommon.QueryCondition) ([]Client.Result, error) {
	queryCondition := *condition
	queryCondition.Limit = 1

	statement := InternalInflux.Statement{
		QueryCondition: &queryCondition,
		Measurement:    DutyCycle,
		GroupByTags:    []string{"uuid"},
	}

	statement.AppendWhereClause("AND", "host", "=", host)
	statement.AppendWhereClause("AND", "minor_number", "=", minorNumber)
	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	return r.influxDB.QueryDB(cmd, string(RepoInflux.Gpu))
}

func (r *DutyCycleRepository) max(host, minorNumber string, condition *DBCommon.QueryCondition) ([]Client.Result, error) {
	seconds := int(condition.StepTime.Seconds())
	groupTag := fmt.Sprintf("time(%ds)", seconds)

	statement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    DutyCycle,
		GroupByTags:    []string{"uuid", groupTag},
	}

	statement.AppendWhereClause("AND", "host", "=", host)
	statement.AppendWhereClause("AND", "minor_number", "=", minorNumber)
	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	statement.SetFunction(InternalInflux.Select, "MAX", "")
	cmd := statement.BuildQueryCmd()

	return r.influxDB.QueryDB(cmd, string(RepoInflux.Gpu))
}

func (r *DutyCycleRepository) genEntities(response []Client.Result) []*EntityInfluxGpuMetric.DutyCycleEntity {
	entities := make([]*EntityInfluxGpuMetric.DutyCycleEntity, 0)

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

	return entities
}
