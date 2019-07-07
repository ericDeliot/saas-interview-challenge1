package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/adjust/rmq"
	"github.com/go-redis/redis"

	"github.com/ericDeliot/saas-interview-challenge1/common"
	"github.com/ericDeliot/saas-interview-challenge1/monitor"
	"github.com/ericDeliot/saas-interview-challenge1/producer"
)

// This program will run a Task producer
// It relies on rmq to offer a queue abstraction built on top of redis lists.
func main() {
	if len(os.Args) > 2 {
		fmt.Println("You must pass argument: \n producer [taskNumber]")
		os.Exit(1)
	}
	redisLoc := "localhost:6379"
	connection := rmq.OpenConnection("producer", "tcp", redisLoc, 0)
	taskQueue := connection.OpenQueue("tasks")
	redisdb := redis.NewClient(&redis.Options{
		Addr:     redisLoc,
		Password: "",
		DB:       0,
	})

	// setup a monitor before producing anything
	monitor := monitor.NewMonitor(redisdb)
	taskChan, err := monitor.Subscribe()
	if err != nil {
		fmt.Printf("Could not subscribe to monitor tasks - abort")
		os.Exit(1)
	}
	// start the processing of messages on a goroutine
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go monitor.ProcessChannel(taskChan, wg)

	// now create some tasks
	prod := producer.NewProducer(taskQueue)

	// default value
	taskNumber := 10
	if len(os.Args) == 2 {
		taskNbr, err := strconv.Atoi(os.Args[1])
		if err == nil {
			taskNumber = taskNbr
		}
	}

	tasks := make([]common.Task, taskNumber)
	// seed random number to get different times every run
	rand.Seed(time.Now().Unix())
	for i := 0; i < taskNumber; i++ {
		// get a random nbr between 1000 and 5000 (used as milliseconds)
		randNbr := rand.Intn(4000) + 1000
		tasks[i] = common.Task{Name: "task_" + strconv.Itoa(i), Duration: time.Duration(randNbr) * time.Millisecond,
			Status: common.New}
	}
	err = prod.Publish(tasks)
	if err != nil {
		fmt.Printf("Error publishing tasks: %s \n", err.Error())
		os.Exit(1)
	}
	// wait for monitor to exit
	wg.Wait()

}
