package common

import (
	"time"

	"github.com/adjust/rmq"
)

// Task represents the work that needs to be done. In this test, work is simply
// represented by the time the worker will sleep when processing the task
type Task struct {
	Name     string
	Duration time.Duration
	Status   TaskStatus
}

// QueueWrapper is an Interface enabling tests to only mock the methods of rmq.Queue that are actually used
type QueueWrapper interface {
	PublishBytes(payload []byte) bool
	StartConsuming(prefetchLimit int, pollDuration time.Duration) bool
	AddConsumer(tag string, consumer rmq.Consumer) string
}

// TaskStatus is an Enum like const used to record the task status
type TaskStatus int

const (
	New TaskStatus = iota
	InProgress
	Done
	Failed
)

// TaskStatusStr is a helper list used to log human readable statuses
var TaskStatusStr = [...]string{
	"New",
	"InProgress",
	"Done",
	"Failed",
}

// TaskRecordTopic is the name used by both tasks recorder and monitor to exchange tasks statuses
const TaskRecordTopic = "recordTopic"
