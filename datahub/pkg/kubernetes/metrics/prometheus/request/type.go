package request

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics"
)

const (
	apiPrefix    = "/api/v1"
	epQuery      = "/query"
	epQueryRange = "/query_range"
)

var (
	StringOperatorLiteral = map[metrics.StringOperator]string{
		metrics.StringOperatorEqueal:    "=",
		metrics.StringOperatorNotEqueal: "!=",
	}
)
