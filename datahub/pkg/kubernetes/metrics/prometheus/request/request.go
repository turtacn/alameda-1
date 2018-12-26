package request

import (
	"net/url"
	"strconv"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics"
)

type QueryRequest interface{}

func GetQueryEndpointByTimeSelector(t metrics.TimeSelector) string {

	var ep string

	switch t.(type) {
	case nil:
		ep = apiPrefix + epQuery
	case *metrics.Timestamp:
		ep = apiPrefix + epQuery
	case *metrics.TimeRange:
		ep = apiPrefix + epQueryRange
	case *metrics.Since:
		ep = apiPrefix + epQuery
	}

	return ep
}

func GetQueryParametersByTimeSelector(t metrics.TimeSelector) url.Values {

	var v = make(url.Values)

	switch t.(type) {
	case *metrics.Timestamp:
		t := t.(*metrics.Timestamp)
		tStr := strconv.FormatInt(int64(t.T.Unix()), 10)
		v.Set("time", tStr)

	case *metrics.TimeRange:
		t := t.(*metrics.TimeRange)
		startTime := t.StartTime
		endTime := t.EndTime
		step := t.Step
		startTimeString := strconv.FormatInt(int64(startTime.Unix()), 10)
		endTimeString := strconv.FormatInt(int64(endTime.Unix()), 10)
		stepString := strconv.FormatFloat(step.Seconds(), 'f', 0, 64)

		v.Set("start", startTimeString)
		v.Set("end", endTimeString)
		v.Set("step", stepString)
	}

	return v
}
