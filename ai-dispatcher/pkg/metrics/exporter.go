package metrics

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
	podNS, podName, dataGranularity string, val float64) {
	exporter.podMetric.setPodMetricModelTime(podNS,
		podName, dataGranularity, val)
	exporter.podMetric.addPodMetricModelTimeTotal(podNS,
		podName, dataGranularity, val)
}

func (exporter *Exporter) SetContainerMetricMAPE(
	podNS, podName, name, metricType, dataGranularity string, val float64) {
	exporter.podMetric.setContainerMetricMAPE(podNS,
		podName, name, metricType, dataGranularity, val)
}

func (exporter *Exporter) SetContainerMetricRMSE(
	podNS, podName, name, metricType, dataGranularity string, val float64) {
	exporter.podMetric.setContainerMetricRMSE(podNS,
		podName, name, metricType, dataGranularity, val)
}

func (exporter *Exporter) AddPodMetricDrift(
	podNS, podName, dataGranularity string, val float64) {
	exporter.podMetric.addPodMetricDrift(podNS,
		podName, dataGranularity, val)
}

// Node
func (exporter *Exporter) ExportNodeMetricModelTime(
	name, dataGranularity string, val float64) {
	exporter.nodeMetric.setNodeMetricModelTime(name,
		dataGranularity, val)
	exporter.nodeMetric.addNodeMetricModelTimeTotal(name,
		dataGranularity, val)
}

func (exporter *Exporter) SetNodeMetricMAPE(
	name, metricType, dataGranularity string, val float64) {
	exporter.nodeMetric.setNodeMetricMAPE(name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) SetNodeMetricRMSE(
	name, metricType, dataGranularity string, val float64) {
	exporter.nodeMetric.setNodeMetricRMSE(name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) AddNodeMetricDrift(
	name, dataGranularity string, val float64) {
	exporter.nodeMetric.addNodeMetricDrift(name,
		dataGranularity, val)
}

// GPU
func (exporter *Exporter) ExportGPUMetricModelTime(host, minor_number,
	dataGranularity string, val float64) {
	exporter.gpuMetric.setGPUMetricModelTime(host, minor_number,
		dataGranularity, val)
	exporter.gpuMetric.addGPUMetricModelTimeTotal(host, minor_number,
		dataGranularity, val)
}

func (exporter *Exporter) SetGPUMetricMAPE(host, minor_number,
	metricType, dataGranularity string, val float64) {
	exporter.gpuMetric.setGPUMetricMAPE(host, minor_number,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) SetGPUMetricRMSE(host, minor_number,
	metricType, dataGranularity string, val float64) {
	exporter.gpuMetric.setGPUMetricRMSE(host, minor_number,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) AddGPUMetricDrift(host, minor_number,
	dataGranularity string, val float64) {
	exporter.gpuMetric.addGPUMetricDrift(host, minor_number,
		dataGranularity, val)
}

// Application
func (exporter *Exporter) ExportApplicationMetricModelTime(ns, name,
	dataGranularity string, val float64) {
	exporter.applicationMetric.setApplicationMetricModelTime(ns, name,
		dataGranularity, val)
	exporter.applicationMetric.addApplicationMetricModelTimeTotal(ns, name,
		dataGranularity, val)
}

func (exporter *Exporter) SetApplicationMetricMAPE(ns, name,
	metricType, dataGranularity string, val float64) {
	exporter.applicationMetric.setApplicationMetricMAPE(ns, name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) SetApplicationMetricRMSE(ns, name,
	metricType, dataGranularity string, val float64) {
	exporter.applicationMetric.setApplicationMetricRMSE(ns, name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) AddApplicationMetricDrift(ns, name,
	dataGranularity string, val float64) {
	exporter.applicationMetric.addApplicationMetricDrift(ns, name,
		dataGranularity, val)
}

// Namespace
func (exporter *Exporter) ExportNamespaceMetricModelTime(
	name, dataGranularity string, val float64) {
	exporter.namespaceMetric.setNamespaceMetricModelTime(name,
		dataGranularity, val)
	exporter.namespaceMetric.addNamespaceMetricModelTimeTotal(name,
		dataGranularity, val)
}

func (exporter *Exporter) SetNamespaceMetricMAPE(
	name, metricType, dataGranularity string, val float64) {
	exporter.namespaceMetric.setNamespaceMetricMAPE(name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) SetNamespaceMetricRMSE(
	name, metricType, dataGranularity string, val float64) {
	exporter.namespaceMetric.setNamespaceMetricRMSE(name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) AddNamespaceMetricDrift(
	name, dataGranularity string, val float64) {
	exporter.namespaceMetric.addNamespaceMetricDrift(name,
		dataGranularity, val)
}

// Cluster
func (exporter *Exporter) ExportClusterMetricModelTime(
	name, dataGranularity string, val float64) {
	exporter.clusterMetric.setClusterMetricModelTime(name,
		dataGranularity, val)
	exporter.clusterMetric.addClusterMetricModelTimeTotal(name,
		dataGranularity, val)
}

func (exporter *Exporter) SetClusterMetricMAPE(
	name, metricType, dataGranularity string, val float64) {
	exporter.clusterMetric.setClusterMetricMAPE(name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) SetClusterMetricRMSE(
	name, metricType, dataGranularity string, val float64) {
	exporter.clusterMetric.setClusterMetricRMSE(name,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) AddClusterMetricDrift(
	name, dataGranularity string, val float64) {
	exporter.clusterMetric.addClusterMetricDrift(name,
		dataGranularity, val)
}

// Controller
func (exporter *Exporter) ExportControllerMetricModelTime(ns, name, kind,
	dataGranularity string, val float64) {
	exporter.controllerMetric.setControllerMetricModelTime(ns, name,
		kind, dataGranularity, val)
	exporter.controllerMetric.addControllerMetricModelTimeTotal(ns, name,
		kind, dataGranularity, val)
}

func (exporter *Exporter) SetControllerMetricMAPE(ns, name, kind,
	metricType, dataGranularity string, val float64) {
	exporter.controllerMetric.setControllerMetricMAPE(ns, name, kind,
		metricType, dataGranularity, val)
}

func (exporter *Exporter) SetControllerMetricRMSE(ns, name, kind,
	metricType, dataGranularity string, val float64) {
	exporter.controllerMetric.setControllerMetricRMSE(ns, name,
		kind, metricType, dataGranularity, val)
}

func (exporter *Exporter) AddControllerMetricDrift(ns, name, kind,
	dataGranularity string, val float64) {
	exporter.controllerMetric.addControllerMetricDrift(ns, name,
		kind, dataGranularity, val)
}
