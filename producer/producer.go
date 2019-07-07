package producer

import (
	"encoding/json"
	"fmt"

	"github.com/ericDeliot/saas-interview-challenge1/common"
)

// Producer is a simple helper object that wraps the publishing functionality
// Producer for now is not a dependency in other objects so defining an interface for it is not necessary
type Producer struct {
	taskQueue common.QueueWrapper
}

// NewProducer is a helper method creating a new instance of Producer
func NewProducer(taskQueue common.QueueWrapper) *Producer {
	ret := new(Producer)
	ret.taskQueue = taskQueue
	return ret
}

// Publish simply takes the given tasks and inserts them in the task queue.
func (p *Producer) Publish(tasks []common.Task) error {
	for _, task := range tasks {
		fmt.Printf("Publishing task %s \n", task.Name)
		taskBytes, err := json.Marshal(task)
		if err != nil {
			return err
		}
		p.taskQueue.PublishBytes(taskBytes)
	}
	return nil
}
