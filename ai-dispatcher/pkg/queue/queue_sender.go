package queue

import "github.com/containers-ai/alameda/pkg/utils/log"

type QueueSender interface {
	SendJsonString(queueName, jsonStr string) error
	getRetry() *retry
}

var scope = log.RegisterScope("queue", "job queue", 0)

const (
	DEFAULT_PUBLISH_RETRY_TIME              = 3
	DEFAULT_PUBLISH_RETRY_INTERVAL_MS int64 = 500
)
