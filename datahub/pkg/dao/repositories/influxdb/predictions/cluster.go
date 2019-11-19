package predictions

import (
	EntityInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/predictions"
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
)

type ClusterRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewClusterRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ClusterRepository {
	return &ClusterRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ClusterRepository) CreatePredictions(predictions DaoPredictionTypes.ClusterPredictionMap) error {
	points := make([]*InfluxClient.Point, 0)

	for _, prediction := range predictions.MetricMap {
		r.appendPoints(FormatEnum.MetricKindRaw, prediction.ObjectMeta, prediction.PredictionRaw, &points)
		r.appendPoints(FormatEnum.MetricKindUpperBound, prediction.ObjectMeta, prediction.PredictionUpperBound, &points)
		r.appendPoints(FormatEnum.MetricKindLowerBound, prediction.ObjectMeta, prediction.PredictionLowerBound, &points)
	}

	// Batch write influxdb data points
	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		return errors.Wrap(err, "failed to batch write cluster prediction in influxdb")
	}

	return nil
}

func (r *ClusterRepository) ListPredictions(request DaoPredictionTypes.ListClusterPredictionsRequest) ([]*DaoPredictionTypes.ClusterPrediction, error) {
	clusterPredictionList := make([]*DaoPredictionTypes.ClusterPrediction, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Cluster,
		GroupByTags:    []string{string(EntityInfluxPrediction.ClusterName)},
	}

	for _, objectMeta := range request.ObjectMeta {
		keyList := objectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxPrediction.ClusterGranularity))
		keyList = append(keyList, string(EntityInfluxPrediction.ClusterModelId))
		keyList = append(keyList, string(EntityInfluxPrediction.ClusterPredictionId))

		valueList := objectMeta.GenerateValueList()
		valueList = append(valueList, strconv.FormatInt(request.Granularity, 10))
		valueList = append(valueList, request.ModelId)
		valueList = append(valueList, request.PredictionId)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	if len(request.ObjectMeta) == 0 {
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.ClusterGranularity), "=", strconv.FormatInt(request.Granularity, 10))
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.ApplicationModelId), "=", request.ModelId)
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.ApplicationPredictionId), "=", request.PredictionId)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetLimitClauseFromQueryCondition()
	statement.SetOrderClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return make([]*DaoPredictionTypes.ClusterPrediction, 0), errors.Wrap(err, "failed to list cluster prediction")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			clusterPrediction := DaoPredictionTypes.NewClusterPrediction()
			clusterPrediction.ObjectMeta.Initialize(group.GetRow(0))
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewClusterEntity(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.MetricType {
					case FormatEnum.MetricKindRaw:
						clusterPrediction.AddRawSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindUpperBound:
						clusterPrediction.AddUpperBoundSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindLowerBound:
						clusterPrediction.AddLowerBoundSample(*entity.Metric, granularity, sample)
					}
				}
			}
			clusterPredictionList = append(clusterPredictionList, clusterPrediction)
		}
	}

	return clusterPredictionList, nil
}

func (r *ClusterRepository) appendPoints(kind FormatEnum.MetricKind, objectMeta Metadata.ObjectMeta, predictions map[FormatEnum.MetricType]*FormatTypes.PredictionMetricData, points *[]*InfluxClient.Point) error {
	for metricType, metricData := range predictions {
		granularity := metricData.Granularity
		if granularity == 0 {
			granularity = 30
		}

		for _, sample := range metricData.Data {
			// Parse float string to value
			valueInFloat64, err := DatahubUtils.StringToFloat64(sample.Value)
			if err != nil {
				return errors.Wrap(err, "failed to parse string to float64")
			}

			// Pack influx tags
			tags := map[string]string{
				string(EntityInfluxPrediction.ClusterName):        objectMeta.Name,
				string(EntityInfluxPrediction.ClusterMetric):      metricType,
				string(EntityInfluxPrediction.ClusterMetricType):  kind,
				string(EntityInfluxPrediction.ClusterGranularity): strconv.FormatInt(granularity, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.ClusterModelId):      sample.ModelId,
				string(EntityInfluxPrediction.ClusterPredictionId): sample.PredictionId,
				string(EntityInfluxPrediction.ClusterValue):        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(Cluster), tags, fields, sample.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			*points = append(*points, point)
		}
	}

	return nil
}
