package types

import (
	"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
)

type ContainerPredictionSample struct {
	ContainerName metadata.ContainerName
	PodName       metadata.PodName
	Namespace     metadata.NamespaceName
	NodeName      metadata.NodeName
	ClusterName   metadata.ClusterName
	MetricType    enumconv.MetricType
	MetricKind    enumconv.MetricKind
	Predictions   *types.PredictionMetricData
}

// ContainerPrediction Prediction model to represent one container Prediction
type ContainerPrediction struct {
	ContainerName        metadata.ContainerName
	PodName              metadata.PodName
	Namespace            metadata.NamespaceName
	NodeName             metadata.NodeName
	ClusterName          metadata.ClusterName
	PredictionRaw        map[enumconv.MetricType]*types.PredictionMetricData
	PredictionUpperBound map[enumconv.MetricType]*types.PredictionMetricData
	PredictionLowerBound map[enumconv.MetricType]*types.PredictionMetricData
}

// ContainersPredictionMap Containers Prediction map
type ContainerPredictionMap struct {
	MetricMap map[metadata.NamespacePodContainerName]*ContainerPrediction
}

func NewContainerPrediction() *ContainerPrediction {
	containerPrediction := &ContainerPrediction{}
	containerPrediction.PredictionRaw = make(map[enumconv.MetricType]*types.PredictionMetricData)
	containerPrediction.PredictionUpperBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	containerPrediction.PredictionLowerBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	return containerPrediction
}

func NewContainerPredictionMap() ContainerPredictionMap {
	containerPredictionMap := ContainerPredictionMap{}
	containerPredictionMap.MetricMap = make(map[metadata.NamespacePodContainerName]*ContainerPrediction)
	return containerPredictionMap
}

func (c *ContainerPrediction) AddRawSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := c.PredictionRaw[metricType]; !exist {
		c.PredictionRaw[metricType] = types.NewPredictionMetricData()
		c.PredictionRaw[metricType].Granularity = granularity
	}
	c.PredictionRaw[metricType].Data = append(c.PredictionRaw[metricType].Data, sample)
}

func (c *ContainerPrediction) AddUpperBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := c.PredictionUpperBound[metricType]; !exist {
		c.PredictionUpperBound[metricType] = types.NewPredictionMetricData()
		c.PredictionUpperBound[metricType].Granularity = granularity
	}
	c.PredictionUpperBound[metricType].Data = append(c.PredictionUpperBound[metricType].Data, sample)
}

func (c *ContainerPrediction) AddLowerBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := c.PredictionLowerBound[metricType]; !exist {
		c.PredictionLowerBound[metricType] = types.NewPredictionMetricData()
		c.PredictionLowerBound[metricType].Granularity = granularity
	}
	c.PredictionLowerBound[metricType].Data = append(c.PredictionLowerBound[metricType].Data, sample)
}

func (c *ContainerPrediction) Merge(in *ContainerPrediction) {
	// Handle predicted raw data
	for metricType, predictions := range in.PredictionRaw {
		c.PredictionRaw[metricType].Data = append(c.PredictionRaw[metricType].Data, predictions.Data...)
	}

	// Handle predicted upper bound data
	for metricType, predictions := range in.PredictionUpperBound {
		c.PredictionUpperBound[metricType].Data = append(c.PredictionUpperBound[metricType].Data, predictions.Data...)
	}

	// Handle predicted lower bound data
	for metricType, predictions := range in.PredictionLowerBound {
		c.PredictionLowerBound[metricType].Data = append(c.PredictionLowerBound[metricType].Data, predictions.Data...)
	}
}

// BuildPodPrediction Build PodPrediction consist of the receiver in ContainersPredictionMap.
func (c *ContainerPrediction) BuildPodPrediction() *PodPrediction {
	containerPredictionMap := NewContainerPredictionMap()
	containerPredictionMap.MetricMap[c.NamespacePodContainerName()] = c

	return &PodPrediction{
		ObjectMeta: metadata.ObjectMeta{
			Name:        c.PodName,
			Namespace:   c.Namespace,
			NodeName:    c.NodeName,
			ClusterName: c.ClusterName,
		},
		ContainerPredictionMap: containerPredictionMap,
	}
}

// NamespacePodContainerName Return identity of the container Prediction.
func (c *ContainerPrediction) NamespacePodContainerName() metadata.NamespacePodContainerName {
	return metadata.NamespacePodContainerName(fmt.Sprintf("%s/%s/%s", c.Namespace, c.PodName, c.ContainerName))
}

func (c *ContainerPredictionMap) AddContainerPrediction(containerPrediction *ContainerPrediction) {
	namespaceContainerName := containerPrediction.NamespacePodContainerName()
	if existContainerPrediction, exist := c.MetricMap[namespaceContainerName]; exist {
		existContainerPrediction.Merge(containerPrediction)
	} else {
		c.MetricMap[namespaceContainerName] = containerPrediction
	}
}

// BuildPodsPredictionMap Build PodsPredictionMap base on current ContainersPredictionMap
func (c *ContainerPredictionMap) BuildPodsPredictionMap() PodPredictionMap {
	podPredictionMap := NewPodPredictionMap()

	for _, containerPrediction := range c.MetricMap {
		podPredictionMap.AddContainerPrediction(containerPrediction)
	}

	return podPredictionMap
}
