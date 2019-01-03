package prediction

import (
	"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"time"
)

// NamespaceName Type alias
type NamespaceName = string

// PodName Type alias
type PodName = string

// ContainerName Type alias
type ContainerName = string

// NodeName Type alias
type NodeName = string

// NamespacePodName Type alias
type NamespacePodName = string

// NamespacePodContainerName Type alias
type NamespacePodContainerName = string

// Sample Data struct representing timestamp and Prediction value of Prediction data point
type Sample = metric.Sample

// DAO DAO interface of prediction
type DAO interface {
	ListPodPredictions(ListPodPredictionsRequest) (PodsPredictionMap, error)
	CreateContainerPredictions([]*ContainerPrediction) error
}

// ListPodPredictionsRequest ListPodPredictionsRequest interface of prediction
type ListPodPredictionsRequest struct {
	Namespace string
	PodName   string
	StartTime *time.Time
	EndTime   *time.Time
}

// ContainerPrediction Prediction model to represent one container Prediction
type ContainerPrediction struct {
	Namespace         NamespaceName
	PodName           PodName
	ContainerName     ContainerName
	CPUPredictions    []Sample
	MemoryPredictions []Sample
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
func (c ContainerPrediction) NamespacePodContainerName() NamespacePodContainerName {
	return NamespacePodContainerName(fmt.Sprintf("%s/%s/%s", c.Namespace, c.PodName, c.ContainerName))
}

// ContainersPredictionMap Containers Prediction map
type ContainersPredictionMap map[NamespacePodContainerName]ContainerPrediction

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
	Namespace               NamespaceName
	PodName                 PodName
	ContainersPredictionMap ContainersPredictionMap
}

// NamespacePodName Return identity of the pod Prediction
func (p PodPrediction) NamespacePodName() NamespacePodName {
	return NamespacePodName(fmt.Sprintf("%s/%s", p.Namespace, p.PodName))
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
type PodsPredictionMap map[NamespacePodName]PodPrediction

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
