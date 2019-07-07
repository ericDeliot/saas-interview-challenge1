package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/adjust/rmq"
	"github.com/go-redis/redis"

	"github.com/ericDeliot/saas-interview-challenge1/recorder"
	"github.com/ericDeliot/saas-interview-challenge1/worker"
)

// This program will run a Worker receiving tasks from redis
// processing them and recording their status.
// It relies on rmq to offer a queue abstraction built on top of redis lists.
func main() {
	if len(os.Args) > 2 {
		fmt.Println("You must pass argument: \n worker [workerNumber]")
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

	// default value
	workerNumber := 1
	if len(os.Args) == 2 {
		wkrNbr, err := strconv.Atoi(os.Args[1])
		if err == nil {
			workerNumber = wkrNbr
		}
	}
	recorder := recorder.NewRecorder(redisdb)
	rand.Seed(time.Now().Unix())
	randName := strconv.Itoa(rand.Intn(100000))
	worker := worker.NewWorker("worker_"+randName, taskQueue, recorder)
	err := worker.Start(workerNumber)
	if err != nil {
		fmt.Printf("Error starting worker: %s \n", err.Error())
		os.Exit(1)
	}
	// wait for for more tasks to turn up
	// TODO catch signal to close worker gracefully
	select {}

}
