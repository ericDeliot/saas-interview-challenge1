package recorder

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"

	"github.com/ericDeliot/saas-interview-challenge1/common"
)

// Recorder is an interface helping to keep track of the status of each task by publishing it to a well know topic
// Who listens to this topic and what do they do with the information is not the Recorder's concern.
// Defining an interface helps with testing as Recorder is a dependency of Worker.
type Recorder interface {
	RecordTask(task common.Task) error
}

// TaskRecorder is redisdb based incarnation of Recorder
type TaskRecorder struct {
	redisdb *redis.Client
}

// NewRecorder is a helper method creating an instance of Recorder
func NewRecorder(redisdb *redis.Client) Recorder {
	ret := new(TaskRecorder)
	ret.redisdb = redisdb
	return ret
}

// RecordTask uses a well know topic to publish task statuses
func (t *TaskRecorder) RecordTask(task common.Task) error {
	fmt.Printf("Recording task %s status %s \n", task.Name, common.TaskStatusStr[task.Status])
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}
	cmd := t.redisdb.Publish(common.TaskRecordTopic, string(taskBytes))
	if cmd != nil && cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}
