package metrics

type Exporter struct {
	podMetric  *podMetric
	nodeMetric *nodeMetric
	gpuMetric  *gpuMetric
}

func NewExporter() *Exporter {
	return &Exporter{
		podMetric:  newPodMetric(),
		nodeMetric: newNodeMetric(),
		gpuMetric:  newGPUMetric(),
	}
}

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
