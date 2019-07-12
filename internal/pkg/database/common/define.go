package common

import (
	Common "github.com/containers-ai/api/common"
)

// Order enumerator
type Order = int

// Aggregate function enumerator
type AggregateFunction = int

// Sort order definition
const (
	// Represent ascending order
	Asc Order = 0
	// Represent descending order
	Desc Order = 1
)

// Aggregation function definition
const (
	None        AggregateFunction = 0
	MaxOverTime AggregateFunction = 1
)

var (
	AggregationOverTime = map[AggregateFunction]string{
		MaxOverTime: "max_over_time",
	}

	TimeRange2AggregationOverTime = map[Common.TimeRange_AggregateFunction]AggregateFunction{
		Common.TimeRange_NONE: None,
		Common.TimeRange_MAX:  MaxOverTime,
	}
)
