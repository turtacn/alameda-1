package common

import (
	"time"
)

var (
	DefaultStepTime, _ = time.ParseDuration("30s")
)

type Options struct {
	StartTime             *time.Time
	EndTime               *time.Time
	Timeout               *time.Time
	StepTime              *time.Duration
	AggregateOverTimeFunc AggregateFunction
}

type Option func(*Options)

func NewDefaultOptions() Options {
	copyDefaultStepTime := DefaultStepTime

	return Options{
		StepTime:              &copyDefaultStepTime,
		AggregateOverTimeFunc: None,
	}
}

func StartTime(t *time.Time) Option {
	return func(o *Options) {
		o.StartTime = t
	}
}

func EndTime(t *time.Time) Option {
	return func(o *Options) {
		o.EndTime = t
	}
}

func Timeout(t *time.Time) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

func StepTime(d *time.Duration) Option {
	return func(o *Options) {
		o.StepTime = d
	}
}

func AggregateOverTimeFunc(f AggregateFunction) Option {
	return func(o *Options) {
		o.AggregateOverTimeFunc = f
	}
}
