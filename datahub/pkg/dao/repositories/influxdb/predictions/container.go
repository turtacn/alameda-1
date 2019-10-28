package predictions

import (
	"fmt"
	EntityInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/predictions"
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/ptypes"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

// ContainerRepository Repository to access containers' prediction data
type ContainerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

// NewContainerRepositoryWithConfig New container repository with influxDB configuration
func NewContainerRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ContainerRepository) CreatePredictions(predictions []*DaoPredictionTypes.ContainerPredictionSample) error {
	points := make([]*InfluxClient.Point, 0)

	for _, predictionSample := range predictions {
		granularity := int64(30)
		if predictionSample.Predictions.Granularity != 0 {
			granularity = predictionSample.Predictions.Granularity
		}

		for _, sample := range predictionSample.Predictions.Data {
			// Parse float string to value
			valueInFloat64, err := DatahubUtils.StringToFloat64(sample.Value)
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}

			// Pack influx tags
			tags := map[string]string{
				string(EntityInfluxPrediction.ContainerNamespace):   predictionSample.Namespace,
				string(EntityInfluxPrediction.ContainerPodName):     predictionSample.PodName,
				string(EntityInfluxPrediction.ContainerName):        predictionSample.ContainerName,
				string(EntityInfluxPrediction.ContainerMetric):      predictionSample.MetricType,
				string(EntityInfluxPrediction.ContainerKind):        predictionSample.MetricKind,
				string(EntityInfluxPrediction.ContainerGranularity): strconv.FormatInt(granularity, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.ContainerModelId):      sample.ModelId,
				string(EntityInfluxPrediction.ContainerPredictionId): sample.PredictionId,
				string(EntityInfluxPrediction.ContainerValue):        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(Container), tags, fields, sample.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			points = append(points, point)
		}
	}

	// Batch write influxdb data points
	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

// ListContainerPredictionsByRequest list containers' prediction from influxDB
func (r *ContainerRepository) ListPredictions(request DaoPredictionTypes.ListPodPredictionsRequest) ([]*DaoPredictionTypes.ContainerPrediction, error) {
	containerPredictionList := make([]*DaoPredictionTypes.ContainerPrediction, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Container,
		GroupByTags:    []string{string(EntityInfluxPrediction.ContainerNamespace), string(EntityInfluxPrediction.ContainerPodName), string(EntityInfluxPrediction.ContainerName)},
	}

	whereClause := ""
	if request.Granularity == 0 || request.Granularity == 30 {
		whereClause = fmt.Sprintf("(\"%s\"='' OR \"%s\"='30')", string(EntityInfluxPrediction.ContainerGranularity), string(EntityInfluxPrediction.ContainerGranularity))
	} else {
		whereClause = fmt.Sprintf("\"%s\"='%s'", string(EntityInfluxPrediction.ContainerGranularity), strconv.FormatInt(request.Granularity, 10))
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.AppendWhereClause(string(EntityInfluxPrediction.ContainerNamespace), "=", request.Namespace)
	statement.AppendWhereClause(string(EntityInfluxPrediction.ContainerPodName), "=", request.PodName)
	statement.AppendWhereClause(string(EntityInfluxPrediction.ContainerModelId), "=", request.ModelId)
	statement.AppendWhereClause(string(EntityInfluxPrediction.ContainerPredictionId), "=", request.PredictionId)
	statement.AppendWhereClauseDirectly(whereClause)
	statement.SetLimitClauseFromQueryCondition()
	statement.SetOrderClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return make([]*DaoPredictionTypes.ContainerPrediction, 0), errors.Wrap(err, "failed to list container prediction")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			containerPrediction := DaoPredictionTypes.NewContainerPrediction()
			containerPrediction.Namespace = group.Tags[string(EntityInfluxPrediction.ContainerNamespace)]
			containerPrediction.PodName = group.Tags[string(EntityInfluxPrediction.ContainerPodName)]
			containerPrediction.ContainerName = group.Tags[string(EntityInfluxPrediction.ContainerName)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewContainerEntityFromMap(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.Kind {
					case FormatEnum.MetricKindRaw:
						containerPrediction.AddRawSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindUpperBound:
						containerPrediction.AddUpperBoundSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindLowerBound:
						containerPrediction.AddLowerBoundSample(*entity.Metric, granularity, sample)
					}
				}
			}
			containerPredictionList = append(containerPredictionList, containerPrediction)
		}
	}

	return containerPredictionList, nil
}

func (r *ContainerRepository) CreateContainerPrediction(in *ApiPredictions.CreatePodPredictionsRequest) error {

	points := make([]*InfluxClient.Point, 0)

	for _, podPrediction := range in.GetPodPredictions() {
		podNamespace := podPrediction.GetNamespacedName().GetNamespace()
		podName := podPrediction.GetNamespacedName().GetName()
		//modelId := podPrediction.GetModelId()
		//predictionId := podPrediction.GetPredictionId()
		modelId := ""
		predictionId := ""

		for _, containerPrediction := range podPrediction.GetContainerPredictions() {
			containerName := containerPrediction.GetName()

			r.appendMetricDataToPoints(FormatEnum.MetricKindRaw, containerPrediction.GetPredictedRawData(), &points, podNamespace, podName, containerName, modelId, predictionId)
			r.appendMetricDataToPoints(FormatEnum.MetricKindUpperBound, containerPrediction.GetPredictedUpperboundData(), &points, podNamespace, podName, containerName, modelId, predictionId)
			r.appendMetricDataToPoints(FormatEnum.MetricKindLowerBound, containerPrediction.GetPredictedLowerboundData(), &points, podNamespace, podName, containerName, modelId, predictionId)
		}
	}

	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		return errors.Wrap(err, "create container prediction failed")
	}

	return nil
}

func (r *ContainerRepository) appendMetricDataToPoints(kind FormatEnum.MetricKind, metricDataList []*ApiPredictions.MetricData, points *[]*InfluxClient.Point, podNamespace, podName, containerName, modelId, predictionId string) error {
	for _, metricData := range metricDataList {
		metricType := ""
		switch metricData.GetMetricType() {
		case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
			metricType = FormatEnum.MetricTypeCPUUsageSecondsPercentage
		case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
			metricType = FormatEnum.MetricTypeMemoryUsageBytes
		}

		if metricType == "" {
			return errors.New("No corresponding metricType")
		}

		granularity := metricData.GetGranularity()
		if granularity == 0 {
			granularity = 30
		}

		for _, data := range metricData.GetData() {
			tempTimeSeconds := data.GetTime().Seconds
			value := data.GetNumValue()
			valueInFloat64, err := DatahubUtils.StringToFloat64(value)
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}

			tags := map[string]string{
				string(EntityInfluxPrediction.ContainerNamespace):   podNamespace,
				string(EntityInfluxPrediction.ContainerPodName):     podName,
				string(EntityInfluxPrediction.ContainerName):        containerName,
				string(EntityInfluxPrediction.ContainerMetric):      metricType,
				string(EntityInfluxPrediction.ContainerKind):        kind,
				string(EntityInfluxPrediction.ContainerGranularity): strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.ContainerModelId):      modelId,
				string(EntityInfluxPrediction.ContainerPredictionId): predictionId,
				string(EntityInfluxPrediction.ContainerValue):        valueInFloat64,
			}
			point, err := InfluxClient.NewPoint(string(Container), tags, fields, time.Unix(tempTimeSeconds, 0))
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}
			*points = append(*points, point)
		}
	}

	return nil
}

// ListContainerPredictionsByRequest list containers' prediction from influxDB
func (r *ContainerRepository) ListContainerPredictionsByRequest(request DaoPredictionTypes.ListPodPredictionsRequest) ([]*ApiPredictions.PodPrediction, error) {
	whereClause := r.buildInfluxQLWhereClauseFromRequest(request)

	queryCondition := DBCommon.QueryCondition{
		StartTime:      request.QueryCondition.StartTime,
		EndTime:        request.QueryCondition.EndTime,
		StepTime:       request.QueryCondition.StepTime,
		TimestampOrder: request.QueryCondition.TimestampOrder,
		Limit:          request.QueryCondition.Limit,
	}

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: &queryCondition,
		Measurement:    Container,
		WhereClause:    whereClause,
		GroupByTags:    []string{string(EntityInfluxPrediction.ContainerNamespace), string(EntityInfluxPrediction.ContainerPodName), string(EntityInfluxPrediction.ContainerName), string(EntityInfluxPrediction.ContainerMetric), string(EntityInfluxPrediction.ContainerKind), string(EntityInfluxPrediction.ContainerGranularity)},
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.AppendWhereClause(string(EntityInfluxPrediction.ContainerModelId), "=", request.ModelId)
	influxdbStatement.AppendWhereClause(string(EntityInfluxPrediction.ContainerPredictionId), "=", request.PredictionId)
	influxdbStatement.SetLimitClauseFromQueryCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return []*ApiPredictions.PodPrediction{}, errors.Wrap(err, "list container prediction failed")
	}

	rows := InternalInflux.PackMap(results)
	podPredictions := r.getPodPredictionsFromInfluxRows(rows)

	return podPredictions, nil
}

func (r *ContainerRepository) getPodPredictionsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiPredictions.PodPrediction {
	podMap := map[string]*ApiPredictions.PodPrediction{}
	podContainerMap := map[string]*ApiPredictions.ContainerPrediction{}
	podContainerKindMetricMap := map[string]*ApiCommon.MetricData{}
	podContainerKindMetricSampleMap := map[string][]*ApiCommon.Sample{}

	for _, row := range rows {
		namespace := row.Tags[string(EntityInfluxPrediction.ContainerNamespace)]
		podName := row.Tags[string(EntityInfluxPrediction.ContainerPodName)]
		name := row.Tags[string(EntityInfluxPrediction.ContainerName)]
		metricType := row.Tags[string(EntityInfluxPrediction.ContainerMetric)]

		metricValue := ApiCommon.MetricType(ApiCommon.MetricType_value[metricType])
		switch metricType {
		case FormatEnum.MetricTypeCPUUsageSecondsPercentage:
			metricValue = ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE
		case FormatEnum.MetricTypeMemoryUsageBytes:
			metricValue = ApiCommon.MetricType_MEMORY_USAGE_BYTES
		}

		kind := FormatEnum.MetricKindRaw
		if val, ok := row.Tags[string(EntityInfluxPrediction.ContainerKind)]; ok {
			if val != "" {
				kind = val
			}
		}

		granularity := int64(30)
		if val, ok := row.Tags[string(EntityInfluxPrediction.ContainerGranularity)]; ok {
			if val != "" {
				granularity, _ = strconv.ParseInt(val, 10, 64)
			}
		}

		for _, data := range row.Data {
			modelId := data[string(EntityInfluxPrediction.ContainerModelId)]
			predictionId := data[string(EntityInfluxPrediction.ContainerPredictionId)]

			podKey := namespace + "|" + podName + "|" + modelId + "|" + predictionId
			podMap[podKey] = &ApiPredictions.PodPrediction{}
			podMap[podKey].NamespacedName = &ApiResources.NamespacedName{
				Namespace: namespace,
				Name:      podName,
			}
			//podMap[namespace+"|"+podName+"|"+modelId+"|"+predictionId].ModelId = modelId
			//podMap[namespace+"|"+podName+"|"+modelId+"|"+predictionId].PredictionId = predictionId

			podContainerKey := podKey + "|" + name
			podContainerMap[podContainerKey] = &ApiPredictions.ContainerPrediction{}
			podContainerMap[podContainerKey].Name = name

			metricKey := podContainerKey + "|" + kind + "|" + metricType
			podContainerKindMetricMap[metricKey] = &ApiCommon.MetricData{}
			podContainerKindMetricMap[metricKey].MetricType = metricValue
			podContainerKindMetricMap[metricKey].Granularity = granularity

			t, _ := time.Parse(time.RFC3339, data[string(EntityInfluxPrediction.ContainerTime)])
			value := data[string(EntityInfluxPrediction.ContainerValue)]

			googleTimestamp, _ := ptypes.TimestampProto(t)

			tempSample := &ApiCommon.Sample{
				Time:     googleTimestamp,
				NumValue: value,
			}
			podContainerKindMetricSampleMap[metricKey] = append(podContainerKindMetricSampleMap[metricKey], tempSample)
		}
	}

	for k := range podContainerKindMetricMap {
		namespace := strings.Split(k, "|")[0]
		podName := strings.Split(k, "|")[1]
		modelId := strings.Split(k, "|")[2]
		predictionId := strings.Split(k, "|")[3]
		name := strings.Split(k, "|")[4]
		kind := strings.Split(k, "|")[5]
		metricType := strings.Split(k, "|")[6]

		podKey := namespace + "|" + podName + "|" + modelId + "|" + predictionId
		podContainerKey := podKey + "|" + name
		metricKey := podContainerKey + "|" + kind + "|" + metricType

		podContainerKindMetricMap[metricKey].Data = podContainerKindMetricSampleMap[metricKey]

		/*if kind == Metric.ContainerMetricKindUpperbound {
			podContainerMap[podContainerKey].PredictedUpperboundData = append(podContainerMap[podContainerKey].PredictedUpperboundData, podContainerKindMetricMap[metricKey])
		} else if kind == Metric.ContainerMetricKindLowerbound {
			podContainerMap[podContainerKey].PredictedLowerboundData = append(podContainerMap[podContainerKey].PredictedLowerboundData, podContainerKindMetricMap[metricKey])
		} else {
			podContainerMap[podContainerKey].PredictedRawData = append(podContainerMap[podContainerKey].PredictedRawData, podContainerKindMetricMap[metricKey])
		}*/
	}

	for k := range podContainerMap {
		namespace := strings.Split(k, "|")[0]
		podName := strings.Split(k, "|")[1]
		modelId := strings.Split(k, "|")[2]
		predictionId := strings.Split(k, "|")[3]
		name := strings.Split(k, "|")[4]

		podKey := namespace + "|" + podName + "|" + modelId + "|" + predictionId
		podContainerKey := podKey + "|" + name

		podMap[podKey].ContainerPredictions = append(podMap[podKey].ContainerPredictions, podContainerMap[podContainerKey])
	}

	podList := make([]*ApiPredictions.PodPrediction, 0)
	for k := range podMap {
		podList = append(podList, podMap[k])
	}

	return podList
}

func (r *ContainerRepository) buildInfluxQLWhereClauseFromRequest(request DaoPredictionTypes.ListPodPredictionsRequest) string {
	var (
		whereClause string
		conditions  string
	)

	if request.Namespace != "" {
		conditions += fmt.Sprintf(`"%s"='%s'`, string(EntityInfluxPrediction.ContainerNamespace), request.Namespace)
	}
	if request.PodName != "" {
		if conditions != "" {
			conditions += fmt.Sprintf(` AND "%s"='%s'`, string(EntityInfluxPrediction.ContainerPodName), request.PodName)
		} else {
			conditions += fmt.Sprintf(`"%s"='%s'`, string(EntityInfluxPrediction.ContainerPodName), request.PodName)
		}
	}

	if conditions != "" {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(` AND ("%s"='' OR "%s"='%d')`, string(EntityInfluxPrediction.ContainerGranularity), string(EntityInfluxPrediction.ContainerGranularity), request.Granularity)
		} else {
			conditions += fmt.Sprintf(` AND "%s"='%d'`, string(EntityInfluxPrediction.ContainerGranularity), request.Granularity)
		}
	} else {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(`("%s"='' OR "%s"='%d')`, string(EntityInfluxPrediction.ContainerGranularity), string(EntityInfluxPrediction.ContainerGranularity), request.Granularity)
		} else {
			conditions += fmt.Sprintf(`"%s"='%d'`, string(EntityInfluxPrediction.ContainerGranularity), request.Granularity)
		}
	}

	if conditions != "" {
		whereClause = fmt.Sprintf("WHERE %s", conditions)
	}

	return whereClause
}
