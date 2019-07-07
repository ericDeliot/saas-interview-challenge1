package worker

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/adjust/rmq"

	"github.com/ericDeliot/saas-interview-challenge1/common"
)

/*********** Test Types ***************/
type TestRecorder struct {
	statusList []common.TaskStatus
}

func (tr *TestRecorder) RecordTask(task common.Task) error {
	tr.statusList = append(tr.statusList, task.Status)
	return nil
}

func newTestRecorder() *TestRecorder {
	ret := new(TestRecorder)
	ret.statusList = make([]common.TaskStatus, 0)
	return ret
}

type TestDelivery struct {
	acked    bool
	rejected bool
	payload  string
}

func (d *TestDelivery) Payload() string {
	return d.payload
}

func (d *TestDelivery) Ack() bool {
	d.acked = true
	return true
}

func (d *TestDelivery) Reject() bool {
	d.rejected = true
	return true
}

func (d *TestDelivery) Push() bool {
	return true
}

type TestQueue struct {
	consList []string
	subNbr   int
}

func newTestQueue() *TestQueue {
	ret := new(TestQueue)
	ret.consList = make([]string, 0)
	return ret
}

func (q *TestQueue) PublishBytes(payload []byte) bool {
	return true
}

func (q *TestQueue) StartConsuming(prefetchLimit int, pollDuration time.Duration) bool {
	if prefetchLimit == 666 {
		return false
	}
	q.subNbr = prefetchLimit
	return true
}

func (q *TestQueue) AddConsumer(tag string, consumer rmq.Consumer) string {
	q.consList = append(q.consList, tag)
	return "none"
}

/****** Tests start here ******************/
func TestConsumeOK(t *testing.T) {
	rec := newTestRecorder()
	worker := NewWorker("wk1", nil, rec)
	subWorker := newSubWorker("sub1", worker)
	task := common.Task{Name: "task_0", Duration: time.Millisecond, Status: common.New}
	taskBytes, _ := json.Marshal(task)
	td := TestDelivery{payload: string(taskBytes)}
	//exercise function under test
	subWorker.Consume(&td)
	// check calls were done as expected
	if len(rec.statusList) != 2 {
		t.Errorf("Expected 2 statuses - got %d", len(rec.statusList))
		return
	}
	if rec.statusList[0] != common.InProgress {
		t.Errorf("Expected first status to be InProgress - got %v", rec.statusList[0])
	}
	if rec.statusList[1] != common.Done {
		t.Errorf("Expected second status to be Done - got %v", rec.statusList[1])
	}
	if !td.acked {
		t.Error("Expected delivery to be acked")
	}
}

func TestConsumeFailNoTask(t *testing.T) {
	rec := newTestRecorder()
	worker := NewWorker("wk1", nil, rec)
	subWorker := newSubWorker("sub1", worker)

	td := TestDelivery{payload: "wrong format"}
	//exercise function under test
	subWorker.Consume(&td)
	// check calls were done as expected
	if len(rec.statusList) != 0 {
		t.Errorf("Expected 0 statuses - got %d", len(rec.statusList))
	}
	if td.acked {
		t.Error("Expected delivery to be NOT acked")
	}
	if !td.rejected {
		t.Error("Expected delivery to be rejected")
	}
}

func TestStartOK(t *testing.T) {
	queue := newTestQueue()
	worker := NewWorker("wk1", queue, nil)
	//exercise function under test
	err := worker.Start(2)
	//check outcome
	if err != nil {
		t.Errorf("Expected nil error - got %v", err)
	}
	if queue.subNbr != 2 {
		t.Errorf("Expected subNbr to be 2 - got %d", queue.subNbr)
	}
	if len(queue.consList) != 2 {
		t.Errorf("Expected 2 subWorkers - got %d", len(queue.consList))
		return
	}
	if queue.consList[0] != "wk1_0" {
		t.Errorf("Expected first subWorker to be wk1_0 - got %s", queue.consList[0])
	}
	if queue.consList[1] != "wk1_1" {
		t.Errorf("Expected first subWorker to be wk1_1 - got %s", queue.consList[1])
	}
}

func TestStartFailWrongSubNbr(t *testing.T) {
	worker := NewWorker("wk1", nil, nil)
	//exercise function under test
	err := worker.Start(0)
	//check outcomes
	if err == nil {
		t.Error("Expected error - got nil")
	}
}

func TestStartFailStartConsumingError(t *testing.T) {
	queue := newTestQueue()
	worker := NewWorker("wk1", queue, nil)
	//exercise function under test
	err := worker.Start(666)
	//check outcome
	if err == nil {
		t.Error("Expected error - got nil")
	}
}
