package monitor

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-redis/redis"

	"github.com/ericDeliot/saas-interview-challenge1/common"
)

// Monitor is an interface defining how a Monitor can listen to recorded and published tasks statuses
// Defining interfaces helps in testing
// Note that redis.Message is a struct made of 3 strings and could be easily abstracted away in a non-redis
// dependent type for generality's sake, todo
// Also note that Monitor could expose further methods to let other components learn the status of tasks,
// for example if the producer wants to manage task sequencing
type Monitor interface {
	Subscribe() (<-chan *redis.Message, error)
	ProcessChannel(taskChan <-chan *redis.Message, wg *sync.WaitGroup)
}

// TaskMonitor is a redisdb based incarnation of Monitor
type TaskMonitor struct {
	redisdb *redis.Client
}

// NewMonitor is a helper function used to create a new instance of TaskMonitor
func NewMonitor(redisdb *redis.Client) Monitor {
	ret := new(TaskMonitor)
	ret.redisdb = redisdb
	return ret
}

// Subscribe allows the Monitor to get a handle on the channel that is receiving task statuses updates
func (m *TaskMonitor) Subscribe() (<-chan *redis.Message, error) {
	pubsub := m.redisdb.Subscribe(common.TaskRecordTopic)
	_, err := pubsub.Receive()
	if err != nil {
		return nil, err
	}

	// Go channel which receives messages.
	return pubsub.Channel(), nil
}

// ProcessChannel dequeues a message from the channel as it arrives and simply logs it.
// Further processing could be to store the data in redis itself
func (m *TaskMonitor) ProcessChannel(taskChan <-chan *redis.Message, wg *sync.WaitGroup) {
	for msg := range taskChan {
		var task common.Task
		err := json.Unmarshal([]byte(msg.Payload), &task)
		if err != nil {
			fmt.Printf("[Monitor] Abort as cannot unmarshall received message: %v \n", err)
			return
		}
		fmt.Printf("[Monitor] Task %s is %s \n", task.Name, common.TaskStatusStr[task.Status])
	}
	wg.Done()
}
