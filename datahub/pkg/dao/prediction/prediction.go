package prediction

import (
	"fmt"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
)

// IsScheduled Specified if the node prediction is scheduled
type IsScheduled = bool

// DAO DAO interface of prediction
type DAO interface {
	ListPodPredictions(ListPodPredictionsRequest) (PodsPredictionMap, error)
	ListNodePredictions(ListNodePredictionsRequest) (NodesPredictionMap, error)
	CreateContainerPredictions([]*ContainerPrediction) error
	CreateNodePredictions([]*NodePrediction) error
}

// ListPodPredictionsRequest ListPodPredictionsRequest
type ListPodPredictionsRequest struct {
	Namespace string
	PodName   string
	StartTime *time.Time
	EndTime   *time.Time
}

// ListNodePredictionsRequest ListNodePredictionsRequest
type ListNodePredictionsRequest struct {
	NodeNames []string
	StartTime *time.Time
	EndTime   *time.Time
}

// ContainerPrediction Prediction model to represent one container Prediction
type ContainerPrediction struct {
	Namespace         metadata.NamespaceName
	PodName           metadata.PodName
	ContainerName     metadata.ContainerName
	CPUPredictions    []metric.Sample
	MemoryPredictions []metric.Sample
}

// BuildPodPrediction Build PodPrediction consist of the receiver in ContainersPredictionMap.
func (c ContainerPrediction) BuildPodPrediction() PodPrediction {

	containersPredictionMap := ContainersPredictionMap{}
	containersPredictionMap[c.NamespacePodContainerName()] = c

	return PodPrediction{
		Namespace:               c.Namespace,
		PodName:                 c.PodName,
		ContainersPredictionMap: containersPredictionMap,
	}
}

// NamespacePodContainerName Return identity of the container Prediction.
func (c ContainerPrediction) NamespacePodContainerName() metadata.NamespacePodContainerName {
	return metadata.NamespacePodContainerName(fmt.Sprintf("%s/%s/%s", c.Namespace, c.PodName, c.ContainerName))
}

// ContainersPredictionMap Containers Prediction map
type ContainersPredictionMap map[metadata.NamespacePodContainerName]ContainerPrediction

// BuildPodsPredictionMap Build PodsPredictionMap base on current ContainersPredictionMap
func (c ContainersPredictionMap) BuildPodsPredictionMap() PodsPredictionMap {

	var (
		podsPredictionMap = &PodsPredictionMap{}
	)

	for _, containerPrediction := range c {
		podsPredictionMap.AddContainerPrediction(containerPrediction)
	}

	return *podsPredictionMap
}

// Merge Merge current ContainersPredictionMap with input ContainersPredictionMap
func (c ContainersPredictionMap) Merge(in ContainersPredictionMap) ContainersPredictionMap {

	var (
		newContainersPredictionMap = c
	)

	for namespacePodContainerName, containerPrediction := range in {
		if existedContainerPrediction, exist := newContainersPredictionMap[namespacePodContainerName]; exist {
			existedContainerPrediction.CPUPredictions = append(existedContainerPrediction.CPUPredictions, containerPrediction.CPUPredictions...)
			existedContainerPrediction.MemoryPredictions = append(existedContainerPrediction.MemoryPredictions, containerPrediction.MemoryPredictions...)
			newContainersPredictionMap[namespacePodContainerName] = existedContainerPrediction
		} else {
			newContainersPredictionMap[namespacePodContainerName] = containerPrediction
		}
	}

	return newContainersPredictionMap
}

// PodPrediction Prediction model to represent one pod's Prediction
type PodPrediction struct {
	Namespace               metadata.NamespaceName
	PodName                 metadata.PodName
	ContainersPredictionMap ContainersPredictionMap
}

// NamespacePodName Return identity of the pod Prediction
func (p PodPrediction) NamespacePodName() metadata.NamespacePodName {
	return metadata.NamespacePodName(fmt.Sprintf("%s/%s", p.Namespace, p.PodName))
}

// Merge Merge current PodPrediction with input PodPrediction
func (p PodPrediction) Merge(in PodPrediction) PodPrediction {

	var (
		currentContainerMetircMap   = p.ContainersPredictionMap
		mergeWithContainerMetircMap = in.ContainersPredictionMap
		newPodPrediction            = PodPrediction{
			Namespace:               p.Namespace,
			PodName:                 p.PodName,
			ContainersPredictionMap: currentContainerMetircMap.Merge(mergeWithContainerMetircMap),
		}
	)

	return newPodPrediction
}

// PodsPredictionMap Pods' Prediction map
type PodsPredictionMap map[metadata.NamespacePodName]PodPrediction

// AddContainerPrediction Add container Prediction into PodsPredictionMap
func (p *PodsPredictionMap) AddContainerPrediction(c ContainerPrediction) {

	podPrediction := c.BuildPodPrediction()
	namespacePodName := podPrediction.NamespacePodName()
	if existedPodPrediction, exist := (*p)[namespacePodName]; exist {
		(*p)[namespacePodName] = existedPodPrediction.Merge(podPrediction)
	} else {
		(*p)[namespacePodName] = podPrediction
	}
}

// NodePrediction Prediction model to represent one node Prediction
type NodePrediction struct {
	NodeName               metadata.NodeName
	IsScheduled            bool
	CPUUsagePredictions    []metric.Sample
	MemoryUsagePredictions []metric.Sample
}

// Merge Merge current NodePrediction with input NodePrediction
func (n NodePrediction) Merge(in NodePrediction) NodePrediction {

	var (
		newNodePrediction = NodePrediction{
			NodeName:               n.NodeName,
			IsScheduled:            n.IsScheduled,
			CPUUsagePredictions:    append(n.CPUUsagePredictions, in.CPUUsagePredictions...),
			MemoryUsagePredictions: append(n.MemoryUsagePredictions, in.MemoryUsagePredictions...),
		}
	)

	return newNodePrediction
}

// IsScheduledNodePredictionMap Nodes' Prediction map
type IsScheduledNodePredictionMap map[IsScheduled]NodePrediction

// NodesPredictionMap Nodes' Prediction map
type NodesPredictionMap map[metadata.NodeName]IsScheduledNodePredictionMap

// AddNodePrediction Add node Prediction into NodesPredictionMap
func (n *NodesPredictionMap) AddNodePrediction(nodePrediction NodePrediction) {

	nodeName := nodePrediction.NodeName
	isScheduled := nodePrediction.IsScheduled

	if existIsScheduledNodePredictionMap, exist := (*n)[nodeName]; exist {
		if existNodePrediction, exist := existIsScheduledNodePredictionMap[isScheduled]; exist {
			(*n)[nodeName][isScheduled] = existNodePrediction.Merge(nodePrediction)
		} else {
			(*n)[nodeName][isScheduled] = nodePrediction
		}
	} else {
		(*n)[nodeName] = make(IsScheduledNodePredictionMap)
		(*n)[nodeName][isScheduled] = nodePrediction
	}
}
