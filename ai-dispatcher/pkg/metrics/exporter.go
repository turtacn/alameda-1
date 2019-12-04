package metrics

import (
	"strconv"
)

type Exporter struct {
	podMetric         *podMetric
	nodeMetric        *nodeMetric
	gpuMetric         *gpuMetric
	applicationMetric *applicationMetric
	namespaceMetric   *namespaceMetric
	clusterMetric     *clusterMetric
	controllerMetric  *controllerMetric
}

func NewExporter() *Exporter {
	return &Exporter{
		podMetric:         newPodMetric(),
		nodeMetric:        newNodeMetric(),
		gpuMetric:         newGPUMetric(),
		applicationMetric: newApplicationMetric(),
		namespaceMetric:   newNamespaceMetric(),
		controllerMetric:  newControllerMetric(),
	}
}

// Pod
func (exporter *Exporter) ExportPodMetricModelTime(
	podNS, podName, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.podMetric.setPodMetricModelTime(podNS,
		podName, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.podMetric.addPodMetricModelTimeTotal(podNS,
		podName, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetContainerMetricMAPE(
	podNS, podName, name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.podMetric.setContainerMetricMAPE(podNS,
		podName, name, metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetContainerMetricRMSE(
	podNS, podName, name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.podMetric.setContainerMetricRMSE(podNS,
		podName, name, metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddPodMetricDrift(
	podNS, podName, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.podMetric.addPodMetricDrift(podNS,
		podName, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

// Node
func (exporter *Exporter) ExportNodeMetricModelTime(
	name, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.setNodeMetricModelTime(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.nodeMetric.addNodeMetricModelTimeTotal(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNodeMetricMAPE(
	name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.setNodeMetricMAPE(name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNodeMetricRMSE(
	name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.setNodeMetricRMSE(name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddNodeMetricDrift(
	name, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.addNodeMetricDrift(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

// GPU
func (exporter *Exporter) ExportGPUMetricModelTime(host, minor_number,
	dataGranularity string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.setGPUMetricModelTime(host, minor_number,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.gpuMetric.addGPUMetricModelTimeTotal(host, minor_number,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetGPUMetricMAPE(host, minor_number,
	metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.setGPUMetricMAPE(host, minor_number,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetGPUMetricRMSE(host, minor_number,
	metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.setGPUMetricRMSE(host, minor_number,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddGPUMetricDrift(host, minor_number,
	dataGranularity string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.addGPUMetricDrift(host, minor_number,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

// Application
func (exporter *Exporter) ExportApplicationMetricModelTime(ns, name,
	dataGranularity string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.setApplicationMetricModelTime(ns, name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.applicationMetric.addApplicationMetricModelTimeTotal(ns, name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetApplicationMetricMAPE(ns, name,
	metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.setApplicationMetricMAPE(ns, name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetApplicationMetricRMSE(ns, name,
	metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.setApplicationMetricRMSE(ns, name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddApplicationMetricDrift(ns, name,
	dataGranularity string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.addApplicationMetricDrift(ns, name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

// Namespace
func (exporter *Exporter) ExportNamespaceMetricModelTime(
	name, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.setNamespaceMetricModelTime(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.namespaceMetric.addNamespaceMetricModelTimeTotal(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNamespaceMetricMAPE(
	name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.setNamespaceMetricMAPE(name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNamespaceMetricRMSE(
	name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.setNamespaceMetricRMSE(name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddNamespaceMetricDrift(
	name, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.addNamespaceMetricDrift(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

// Cluster
func (exporter *Exporter) ExportClusterMetricModelTime(
	name, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.setClusterMetricModelTime(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.clusterMetric.addClusterMetricModelTimeTotal(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetClusterMetricMAPE(
	name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.setClusterMetricMAPE(name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetClusterMetricRMSE(
	name, metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.setClusterMetricRMSE(name,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddClusterMetricDrift(
	name, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.addClusterMetricDrift(name,
		dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

// Controller
func (exporter *Exporter) ExportControllerMetricModelTime(ns, name, kind,
	dataGranularity string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.setControllerMetricModelTime(ns, name,
		kind, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.controllerMetric.addControllerMetricModelTimeTotal(ns, name,
		kind, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetControllerMetricMAPE(ns, name, kind,
	metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.setControllerMetricMAPE(ns, name, kind,
		metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetControllerMetricRMSE(ns, name, kind,
	metricType, dataGranularity string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.setControllerMetricRMSE(ns, name,
		kind, metricType, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddControllerMetricDrift(ns, name, kind,
	dataGranularity string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.addControllerMetricDrift(ns, name,
		kind, dataGranularity, strconv.FormatInt(exportTimestamp, 10), val)
}
