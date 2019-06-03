package container

import (
	commonAPI "github.com/containers-ai/api/common"
)

type Field = string
type Tag = string

const (
	PodNamespace Tag = "pod_namespace"
	PodName      Tag = "pod_name"
	Name         Tag = "name"
	MetricType   Tag = "metric_type"

	Value Field = "value"
)

var (
	Tags   = []Tag{PodNamespace, PodName, Name, MetricType}
	Fields = []Field{Value}

	MetricDatabaseName    = "alameda_metric"
	MetricMeasurementName = "container"
	MetricColumns         = []string{
		PodNamespace,
		PodName,
		Name,
		MetricType,
		Value}

	MetricColumnsTypes = []commonAPI.ColumnType{
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_FIELD}

	MetricDataTypes = []commonAPI.DataType{
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_FLOAT32}
)
