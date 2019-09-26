package prediction

import (
	"fmt"
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	EntityInfluxPredictionContainer "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/prediction/container"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	Utils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
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

func (r *ContainerRepository) CreateContainerPrediction(in *datahub_v1alpha1.CreatePodPredictionsRequest) error {

	points := make([]*InfluxClient.Point, 0)

	for _, podPrediction := range in.GetPodPredictions() {
		podNamespace := podPrediction.GetNamespacedName().GetNamespace()
		podName := podPrediction.GetNamespacedName().GetName()
		modelId := podPrediction.GetModelId()
		predictionId := podPrediction.GetPredictionId()

		for _, containerPrediction := range podPrediction.GetContainerPredictions() {
			containerName := containerPrediction.GetName()

			r.appendMetricDataToPoints(Metric.ContainerMetricKindRaw, containerPrediction.GetPredictedRawData(), &points, podNamespace, podName, containerName, modelId, predictionId)
			r.appendMetricDataToPoints(Metric.ContainerMetricKindUpperbound, containerPrediction.GetPredictedUpperboundData(), &points, podNamespace, podName, containerName, modelId, predictionId)
			r.appendMetricDataToPoints(Metric.ContainerMetricKindLowerbound, containerPrediction.GetPredictedLowerboundData(), &points, podNamespace, podName, containerName, modelId, predictionId)
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

func (r *ContainerRepository) appendMetricDataToPoints(kind Metric.ContainerMetricKind, metricDataList []*datahub_v1alpha1.MetricData, points *[]*InfluxClient.Point, podNamespace, podName, containerName, modelId, predictionId string) error {
	for _, metricData := range metricDataList {
		metricType := ""
		switch metricData.GetMetricType() {
		case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
			metricType = Metric.TypeContainerCPUUsageSecondsPercentage
		case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
			metricType = Metric.TypeContainerMemoryUsageBytes
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
			valueInFloat64, err := Utils.StringToFloat64(value)
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}

			tags := map[string]string{
				EntityInfluxPredictionContainer.Namespace:   podNamespace,
				EntityInfluxPredictionContainer.PodName:     podName,
				EntityInfluxPredictionContainer.Name:        containerName,
				EntityInfluxPredictionContainer.Metric:      metricType,
				EntityInfluxPredictionContainer.Kind:        kind,
				EntityInfluxPredictionContainer.Granularity: strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				EntityInfluxPredictionContainer.ModelId:      modelId,
				EntityInfluxPredictionContainer.PredictionId: predictionId,
				EntityInfluxPredictionContainer.Value:        valueInFloat64,
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
func (r *ContainerRepository) ListContainerPredictionsByRequest(request DaoPrediction.ListPodPredictionsRequest) ([]*datahub_v1alpha1.PodPrediction, error) {
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
		GroupByTags:    []string{EntityInfluxPredictionContainer.Namespace, EntityInfluxPredictionContainer.PodName, EntityInfluxPredictionContainer.Name, EntityInfluxPredictionContainer.Metric, EntityInfluxPredictionContainer.Kind, EntityInfluxPredictionContainer.Granularity},
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return []*datahub_v1alpha1.PodPrediction{}, errors.Wrap(err, "list container prediction failed")
	}

	rows := InternalInflux.PackMap(results)
	podPredictions := r.getPodPredictionsFromInfluxRows(rows)

	return podPredictions, nil
}

func (r *ContainerRepository) getPodPredictionsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*datahub_v1alpha1.PodPrediction {
	podMap := map[string]*datahub_v1alpha1.PodPrediction{}
	podContainerMap := map[string]*datahub_v1alpha1.ContainerPrediction{}
	podContainerKindMetricMap := map[string]*datahub_v1alpha1.MetricData{}
	podContainerKindMetricSampleMap := map[string][]*datahub_v1alpha1.Sample{}

	for _, row := range rows {
		namespace := row.Tags[EntityInfluxPredictionContainer.Namespace]
		podName := row.Tags[EntityInfluxPredictionContainer.PodName]
		name := row.Tags[EntityInfluxPredictionContainer.Name]
		metricType := row.Tags[EntityInfluxPredictionContainer.Metric]

		metricValue := datahub_v1alpha1.MetricType(datahub_v1alpha1.MetricType_value[metricType])
		switch metricType {
		case Metric.TypeContainerCPUUsageSecondsPercentage:
			metricValue = datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE
		case Metric.TypeContainerMemoryUsageBytes:
			metricValue = datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES
		}

		kind := Metric.ContainerMetricKindRaw
		if val, ok := row.Tags[EntityInfluxPredictionContainer.Kind]; ok {
			if val != "" {
				kind = val
			}
		}

		granularity := int64(30)
		if val, ok := row.Tags[EntityInfluxPredictionContainer.Granularity]; ok {
			if val != "" {
				granularity, _ = strconv.ParseInt(val, 10, 64)
			}
		}

		for _, data := range row.Data {
			modelId := data[EntityInfluxPredictionContainer.ModelId]
			predictionId := data[EntityInfluxPredictionContainer.PredictionId]

			podKey := namespace + "|" + podName + "|" + modelId + "|" + predictionId
			podMap[podKey] = &datahub_v1alpha1.PodPrediction{}
			podMap[podKey].NamespacedName = &datahub_v1alpha1.NamespacedName{
				Namespace: namespace,
				Name:      podName,
			}
			podMap[namespace+"|"+podName+"|"+modelId+"|"+predictionId].ModelId = modelId
			podMap[namespace+"|"+podName+"|"+modelId+"|"+predictionId].PredictionId = predictionId

			podContainerKey := podKey + "|" + name
			podContainerMap[podContainerKey] = &datahub_v1alpha1.ContainerPrediction{}
			podContainerMap[podContainerKey].Name = name

			metricKey := podContainerKey + "|" + kind + "|" + metricType
			podContainerKindMetricMap[metricKey] = &datahub_v1alpha1.MetricData{}
			podContainerKindMetricMap[metricKey].MetricType = metricValue
			podContainerKindMetricMap[metricKey].Granularity = granularity

			t, _ := time.Parse(time.RFC3339, data[EntityInfluxPredictionContainer.Time])
			value := data[EntityInfluxPredictionContainer.Value]

			googleTimestamp, _ := ptypes.TimestampProto(t)

			tempSample := &datahub_v1alpha1.Sample{
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

		if kind == Metric.ContainerMetricKindUpperbound {
			podContainerMap[podContainerKey].PredictedUpperboundData = append(podContainerMap[podContainerKey].PredictedUpperboundData, podContainerKindMetricMap[metricKey])
		} else if kind == Metric.ContainerMetricKindLowerbound {
			podContainerMap[podContainerKey].PredictedLowerboundData = append(podContainerMap[podContainerKey].PredictedLowerboundData, podContainerKindMetricMap[metricKey])
		} else {
			podContainerMap[podContainerKey].PredictedRawData = append(podContainerMap[podContainerKey].PredictedRawData, podContainerKindMetricMap[metricKey])
		}
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

	podList := make([]*datahub_v1alpha1.PodPrediction, 0)
	for k := range podMap {
		podList = append(podList, podMap[k])
	}

	return podList
}

func (r *ContainerRepository) buildInfluxQLWhereClauseFromRequest(request DaoPrediction.ListPodPredictionsRequest) string {

	var (
		whereClause string
		conditions  string
	)

	if request.Namespace != "" {
		conditions += fmt.Sprintf(`"%s"='%s'`, EntityInfluxPredictionContainer.Namespace, request.Namespace)
	}
	if request.PodName != "" {
		if conditions != "" {
			conditions += fmt.Sprintf(` AND "%s"='%s'`, EntityInfluxPredictionContainer.PodName, request.PodName)
		} else {
			conditions += fmt.Sprintf(`"%s"='%s'`, EntityInfluxPredictionContainer.PodName, request.PodName)
		}
	}

	if conditions != "" {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(` AND ("%s"='' OR "%s"='%d')`, EntityInfluxPredictionContainer.Granularity, EntityInfluxPredictionContainer.Granularity, request.Granularity)
		} else {
			conditions += fmt.Sprintf(` AND "%s"='%d'`, EntityInfluxPredictionContainer.Granularity, request.Granularity)
		}
	} else {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(`("%s"='' OR "%s"='%d')`, EntityInfluxPredictionContainer.Granularity, EntityInfluxPredictionContainer.Granularity, request.Granularity)
		} else {
			conditions += fmt.Sprintf(`"%s"='%d'`, EntityInfluxPredictionContainer.Granularity, request.Granularity)
		}
	}

	if conditions != "" {
		whereClause = fmt.Sprintf("WHERE %s", conditions)
	}

	return whereClause
}
