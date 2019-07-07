package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/adjust/rmq"

	"github.com/ericDeliot/saas-interview-challenge1/common"
	"github.com/ericDeliot/saas-interview-challenge1/recorder"
)

// Worker is a manager type structure supervising subworkers
type Worker struct {
	name      string
	taskQueue common.QueueWrapper
	recorder  recorder.Recorder
}

// SubWorker is the object actually processing the task
type SubWorker struct {
	name   string
	worker *Worker
}

// NewWorker is helper function creating an instance of Worker
func NewWorker(name string, taskQueue common.QueueWrapper, recorder recorder.Recorder) *Worker {
	ret := new(Worker)
	ret.name = name
	ret.taskQueue = taskQueue
	ret.recorder = recorder
	return ret
}

func newSubWorker(name string, worker *Worker) *SubWorker {
	ret := new(SubWorker)
	ret.name = name
	ret.worker = worker
	return ret
}

// Consume is the implementation of the callback when a task is dequeued
func (sw *SubWorker) Consume(delivery rmq.Delivery) {
	var task common.Task
	err := json.Unmarshal([]byte(delivery.Payload()), &task)
	if err != nil {
		delivery.Reject()
		return
	}
	fmt.Printf("[%s] START consuming task %s - sleeping for %v: \n", sw.name, task.Name, task.Duration)
	task.Status = common.InProgress
	err = sw.worker.recorder.RecordTask(task)
	if err != nil {
		fmt.Printf("Error in recording task %s -> %v\n", task.Name, err)
	}
	time.Sleep(task.Duration)
	//todo: return failure from time to time
	task.Status = common.Done
	sw.worker.recorder.RecordTask(task)
	fmt.Printf("[%s] STOP consuming task %s: \n", sw.name, task.Name)
	delivery.Ack()
}

// Start initializes the consuming process and registers as many SubWorkers as requested
func (w *Worker) Start(subNbr int) error {
	if subNbr <= 0 {
		return errors.New("subNbr must be an integer >= 1")
	}
	ok := w.taskQueue.StartConsuming(subNbr, 500*time.Millisecond)
	if !ok {
		return errors.New("StartConsuming returned false")
	}
	fmt.Printf("Worker %s starting \n", w.name)
	for i := 0; i < subNbr; i++ {
		subwName := w.name + "_" + strconv.Itoa(i)
		subw := newSubWorker(subwName, w)
		fmt.Printf("SubWorker %s starting \n", subw.name)
		w.taskQueue.AddConsumer(subwName, subw)
	}
	return nil
}
