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

type NamespaceRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNamespaceRepositoryWithConfig(influxDBCfg InternalInflux.Config) *NamespaceRepository {
	return &NamespaceRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *NamespaceRepository) CreatePredictions(predictions DaoPredictionTypes.NamespacePredictionMap) error {
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
		return errors.Wrap(err, "failed to batch write namespace prediction in influxdb")
	}

	return nil
}

func (r *NamespaceRepository) ListPredictions(request DaoPredictionTypes.ListNamespacePredictionsRequest) ([]*DaoPredictionTypes.NamespacePrediction, error) {
	namespacePredictionList := make([]*DaoPredictionTypes.NamespacePrediction, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Namespace,
		GroupByTags:    []string{string(EntityInfluxPrediction.NamespaceName)},
	}

	for _, objectMeta := range request.ObjectMeta {
		keyList := objectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxPrediction.NamespaceGranularity))
		keyList = append(keyList, string(EntityInfluxPrediction.NamespaceModelId))
		keyList = append(keyList, string(EntityInfluxPrediction.NamespacePredictionId))

		valueList := objectMeta.GenerateValueList()
		valueList = append(valueList, strconv.FormatInt(request.Granularity, 10))
		valueList = append(valueList, request.ModelId)
		valueList = append(valueList, request.PredictionId)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	if len(request.ObjectMeta) == 0 {
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NamespaceGranularity), "=", strconv.FormatInt(request.Granularity, 10))
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NamespaceModelId), "=", request.ModelId)
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NamespacePredictionId), "=", request.PredictionId)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetLimitClauseFromQueryCondition()
	statement.SetOrderClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return make([]*DaoPredictionTypes.NamespacePrediction, 0), errors.Wrap(err, "failed to list namespace prediction")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			namespacePrediction := DaoPredictionTypes.NewNamespacePrediction()
			namespacePrediction.ObjectMeta.Initialize(group.GetRow(0))
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewNamespaceEntity(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.MetricType {
					case FormatEnum.MetricKindRaw:
						namespacePrediction.AddRawSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindUpperBound:
						namespacePrediction.AddUpperBoundSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindLowerBound:
						namespacePrediction.AddLowerBoundSample(*entity.Metric, granularity, sample)
					}
				}
			}
			namespacePredictionList = append(namespacePredictionList, namespacePrediction)
		}
	}

	return namespacePredictionList, nil
}

func (r *NamespaceRepository) appendPoints(kind FormatEnum.MetricKind, objectMeta Metadata.ObjectMeta, predictions map[FormatEnum.MetricType]*FormatTypes.PredictionMetricData, points *[]*InfluxClient.Point) error {
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
				string(EntityInfluxPrediction.NamespaceName):        objectMeta.Name,
				string(EntityInfluxPrediction.NamespaceClusterName): objectMeta.ClusterName,
				string(EntityInfluxPrediction.NamespaceMetric):      metricType,
				string(EntityInfluxPrediction.NamespaceMetricType):  kind,
				string(EntityInfluxPrediction.NamespaceGranularity): strconv.FormatInt(granularity, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.NamespaceModelId):      sample.ModelId,
				string(EntityInfluxPrediction.NamespacePredictionId): sample.PredictionId,
				string(EntityInfluxPrediction.NamespaceValue):        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(Namespace), tags, fields, sample.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			*points = append(*points, point)
		}
	}

	return nil
}
