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
func (exporter *Exporter) ExportContainerMetricModelTime(clusterID,
	podNS, podName, ctname, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.podMetric.setContainerMetricModelTime(clusterID, podNS,
		podName, ctname, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.podMetric.addContainerMetricModelTimeTotal(clusterID, podNS,
		podName, ctname, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetContainerMetricMAPE(clusterID,
	podNS, podName, name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.podMetric.setContainerMetricMAPE(clusterID, podNS,
		podName, name, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetContainerMetricRMSE(clusterID,
	podNS, podName, name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.podMetric.setContainerMetricRMSE(clusterID, podNS,
		podName, name, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddContainerMetricDrift(clusterID,
	podNS, podName, name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.podMetric.addPodMetricDrift(clusterID, podNS,
		podName, name, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

// Node
func (exporter *Exporter) ExportNodeMetricModelTime(
	clusterID, name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.setNodeMetricModelTime(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.nodeMetric.addNodeMetricModelTimeTotal(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNodeMetricMAPE(clusterID,
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.setNodeMetricMAPE(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNodeMetricRMSE(clusterID,
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.setNodeMetricRMSE(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddNodeMetricDrift(clusterID,
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.nodeMetric.addNodeMetricDrift(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

// GPU
func (exporter *Exporter) ExportGPUMetricModelTime(clusterID, host, minor_number,
	dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.setGPUMetricModelTime(clusterID, host, minor_number,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.gpuMetric.addGPUMetricModelTimeTotal(clusterID, host, minor_number,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetGPUMetricMAPE(clusterID, host, minor_number, dataGranularity,
	metricType string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.setGPUMetricMAPE(clusterID, host, minor_number,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetGPUMetricRMSE(clusterID, host, minor_number, dataGranularity,
	metricType string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.setGPUMetricRMSE(clusterID, host, minor_number,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddGPUMetricDrift(clusterID, host, minor_number,
	dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.gpuMetric.addGPUMetricDrift(clusterID, host, minor_number,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

// Application
func (exporter *Exporter) ExportApplicationMetricModelTime(clusterID, ns, name,
	dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.setApplicationMetricModelTime(clusterID, ns, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.applicationMetric.addApplicationMetricModelTimeTotal(clusterID, ns, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetApplicationMetricMAPE(clusterID, ns, name, dataGranularity,
	metricType string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.setApplicationMetricMAPE(clusterID, ns, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetApplicationMetricRMSE(clusterID, ns, name, dataGranularity,
	metricType string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.setApplicationMetricRMSE(clusterID, ns, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddApplicationMetricDrift(clusterID, ns, name,
	dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.applicationMetric.addApplicationMetricDrift(clusterID, ns, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

// Namespace
func (exporter *Exporter) ExportNamespaceMetricModelTime(clusterID,
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.setNamespaceMetricModelTime(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.namespaceMetric.addNamespaceMetricModelTimeTotal(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNamespaceMetricMAPE(clusterID,
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.setNamespaceMetricMAPE(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetNamespaceMetricRMSE(clusterID,
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.setNamespaceMetricRMSE(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddNamespaceMetricDrift(clusterID,
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.namespaceMetric.addNamespaceMetricDrift(clusterID, name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

// Cluster
func (exporter *Exporter) ExportClusterMetricModelTime(
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.setClusterMetricModelTime(name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.clusterMetric.addClusterMetricModelTimeTotal(name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetClusterMetricMAPE(
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.setClusterMetricMAPE(name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetClusterMetricRMSE(
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.setClusterMetricRMSE(name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddClusterMetricDrift(
	name, dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.clusterMetric.addClusterMetricDrift(name,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

// Controller
func (exporter *Exporter) ExportControllerMetricModelTime(clusterID, ns, name, kind,
	dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.setControllerMetricModelTime(clusterID, ns, name,
		kind, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
	exporter.controllerMetric.addControllerMetricModelTimeTotal(clusterID, ns, name,
		kind, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetControllerMetricMAPE(clusterID, ns, name, kind, dataGranularity,
	metricType string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.setControllerMetricMAPE(clusterID, ns, name, kind,
		dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) SetControllerMetricRMSE(clusterID, ns, name, kind, dataGranularity,
	metricType string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.setControllerMetricRMSE(clusterID, ns, name,
		kind, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}

func (exporter *Exporter) AddControllerMetricDrift(clusterID, ns, name, kind,
	dataGranularity, metricType string, exportTimestamp int64, val float64) {
	exporter.controllerMetric.addControllerMetricDrift(clusterID, ns, name,
		kind, dataGranularity, metricType, strconv.FormatInt(exportTimestamp, 10), val)
}
