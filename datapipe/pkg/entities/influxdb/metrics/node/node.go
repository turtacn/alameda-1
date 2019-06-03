package node

import (
	commonAPI "github.com/containers-ai/api/common"
)

type Field = string
type Tag = string

const (
	Name       Tag = "name"
	MetricType Tag = "metric_type"

	Value Field = "value"
)

var (
	Tags   = []Tag{Name, MetricType}
	Fields = []Field{Value}

	MetricDatabaseName    = "alameda_metric"
	MetricMeasurementName = "node"
	MetricColumns         = []string{
		Name,
		MetricType,
		Value}

	MetricColumnsTypes = []commonAPI.ColumnType{
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_FIELD}

	MetricDataTypes = []commonAPI.DataType{
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_FLOAT32}
)
