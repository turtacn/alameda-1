package prediction

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	prediction_dao "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	container_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/prediction/container"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
)

// ContainerRepository Repository to access containers' prediction data
type ContainerRepository struct {
	influxDB *influxdb.InfluxDBRepository
}

// NewContainerRepositoryWithConfig New container repository with influxDB configuration
func NewContainerRepositoryWithConfig(influxDBCfg influxdb.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &influxdb.InfluxDBRepository{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

// CreateContainerPrediction Create containers' prediction into influxDB
func (r *ContainerRepository) CreateContainerPrediction(containersPrediction []*prediction_dao.ContainerPrediction) error {

	var (
		err error

		points []*influxdb_client.Point
	)

	for _, containerPrediction := range containersPrediction {

		namespace := containerPrediction.Namespace
		podName := containerPrediction.PodName
		containerName := containerPrediction.ContainerName

		for _, memoryPrediction := range containerPrediction.MemoryPredictions {

			tags := map[string]string{
				container_entity.Namespace: namespace,
				container_entity.PodName:   podName,
				container_entity.Name:      containerName,
				container_entity.Metric:    container_entity.MetricTypeMemoryUsage,
			}
			fields := map[string]interface{}{
				container_entity.Value: memoryPrediction.Value,
			}
			point, err := influxdb_client.NewPoint(container_entity.Measurement, tags, fields, memoryPrediction.Timestamp)
			if err != nil {
				return errors.New("new influxdb datapoint failed: " + err.Error())
			}
			points = append(points, point)
		}

		for _, cpuPrediction := range containerPrediction.CPUPredictions {

			tags := map[string]string{
				container_entity.Namespace: namespace,
				container_entity.PodName:   podName,
				container_entity.Name:      containerName,
				container_entity.Metric:    container_entity.MetricTypeCPUUsage,
			}
			fields := map[string]interface{}{
				container_entity.Value: cpuPrediction.Value,
			}
			point, err := influxdb_client.NewPoint(container_entity.Measurement, tags, fields, cpuPrediction.Timestamp)
			if err != nil {
				return errors.New("new influxdb datapoint failed: " + err.Error())
			}
			points = append(points, point)
		}
	}

	err = r.influxDB.WritePoints(points, influxdb_client.BatchPointsConfig{
		Database: container_entity.Database,
	})
	if err != nil {
		return errors.New("write influxdb datapoint failed: " + err.Error())
	}

	return nil
}

// ListContainerPredictionsByRequest list containers' prediction from influxDB
func (r *ContainerRepository) ListContainerPredictionsByRequest(request prediction_dao.ListPodPredictionsRequest) ([]*container_entity.Entity, error) {

	var (
		err error

		results  []influxdb_client.Result
		rows     []*influxdb.InfluxDBRow
		entities []*container_entity.Entity
	)

	whereClause := buildInfluxQLWhereClauseFromRequest(request)
	lastClause, whereTimeClause := buildTimeClause(request)
	if whereTimeClause != "" {
		if whereClause != "" {
			whereClause += whereTimeClause
		} else {
			whereClause = "where " + whereTimeClause
		}
	}
	whereClause = strings.TrimSuffix(whereClause, "and ")

	cmd := fmt.Sprintf("SELECT * FROM %s %s %s", container_entity.Measurement, whereClause, lastClause)

	results, err = r.influxDB.QueryDB(cmd, string(container_entity.Database))
	if err != nil {
		return entities, err

	}

	rows = influxdb.PackMap(results)

	for _, row := range rows {
		for _, data := range row.Data {
			tempTimestamp, _ := time.Parse("2006-01-02T15:04:05.999999Z07:00", data[container_entity.Time])

			entity := container_entity.Entity{
				Timestamp: tempTimestamp,
				Namespace: data[container_entity.Namespace],
				PodName:   data[container_entity.PodName],
				Name:      data[container_entity.Name],
				Metric:    data[container_entity.Metric],
				Value:     data[container_entity.Value],
			}
			entities = append(entities, &entity)
		}
	}

	return entities, nil
}

func buildInfluxQLWhereClauseFromRequest(request prediction_dao.ListPodPredictionsRequest) string {

	var (
		whereClause string
		conditions  string
	)

	if request.Namespace != "" {
		conditions += fmt.Sprintf(`"%s" = '%s' and `, container_entity.Namespace, request.Namespace)
	}
	if request.PodName != "" {
		conditions += fmt.Sprintf(`"%s" = '%s' and `, container_entity.PodName, request.PodName)
	}

	if conditions != "" {
		whereClause = fmt.Sprintf("where %s", conditions)
	}

	return whereClause
}

func buildTimeClause(request prediction_dao.ListPodPredictionsRequest) (string, string) {

	var (
		lastClause      string
		whereTimeClause string

		startTime = time.Now()
	)

	if request.StartTime == nil && request.EndTime == nil {
		lastClause = "order by time desc limit 1"
	} else {

		if request.StartTime != nil {
			startTime = *request.StartTime
		}

		nanoTimestampInString := strconv.FormatInt(int64(startTime.UnixNano()), 10)
		whereTimeClause = fmt.Sprintf("time > %s and ", nanoTimestampInString)

		if request.EndTime != nil {
			nanoTimestampInString := strconv.FormatInt(int64(request.EndTime.UnixNano()), 10)
			whereTimeClause += fmt.Sprintf("time <= %s and ", nanoTimestampInString)
		}
	}

	return lastClause, whereTimeClause
}
