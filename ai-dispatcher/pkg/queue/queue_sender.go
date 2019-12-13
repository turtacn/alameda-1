package queue

import "github.com/containers-ai/alameda/pkg/utils/log"

type QueueSender interface {
	SendJsonString(queueName, jsonStr, msgID string, timeout int64) error
	getRetry() *retry
}

var scope = log.RegisterScope("queue", "job queue", 0)

const (
	DEFAULT_PUBLISH_RETRY_TIME              = 3
	DEFAULT_PUBLISH_RETRY_INTERVAL_MS int64 = 500
	DEFAULT_CONSUME_RETRY_TIME              = 3
	DEFAULT_CONSUME_RETRY_INTERVAL_MS int64 = 500
	DEFAULT_ACK_TIMEOUT_SEC                 = 3
)
