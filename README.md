# ForgeRock SaaS Software Engineer Coding Challenge

## Requirements

The challenge description asks to:

- build a worker based system: this implies building a queue and having workers dequeue tasks so that each task is porcessed by a worker and many workers can process tasks in parallel
- the tasks must report status to the system
- Redis is suggested as a useful technology to use
- 3rd party code can be used

## Design overview

Not having used Redis before, I spent a bit of time reading about it and the way it provides a list feature that can be used to build a work queue. Several open source projects have been developed that provide a queue abstraction on top of Redis.
I picked https://github.com/adjust/rmq (MIT licensed) to provide a task queue.
My code has 4 components:

- Producer: code pushing tasks into the queue
- Worker: code starting a configurable number of workers that dequeue and process tasks
- Recorder: code updating the status of each task and publishing this update into a well known pubsub topic
- Monitor: code subscribing to the pubsub topic to track the status of each task.

Building the code leads to two executable:

- providersvc: this creates N tasks where N is a command line argument. It also starts a Monitor that logs the tasks statuses.
- workersvc: this is the worker defined above and takes as argument the number of subWorkers it starts.

Tasks are created with a random time.Duration (between 1 and 5s) and the workers simply sleep this amount of time to mimic work.

My objective was to try to write the code so that it was easily testable (using Interfaces when needed) and flexible: A worker process can be configured to run several subWorkers within the same process and many worker processes can also be started in parallel. So, just to be clear, if you start one worker with `workersvc 2` and then add another worker with `workersvc 8`, you should end up with `10` goroutines servicing the same task queue and therefore emptying the queue faster.

## How to build

The code has only one dependency (rmq) and this is handled by the provided Makefile. Running `make` should:

- install the dependdency
- build the code and place the binaries in `$GOPATH/bin`
- run the unit tests
- build the docker image for the worker service

## How to run

A docker-compose file is supplied that should:

- start Redis
- start one intance of the worker service configured to run 2 subWorkers

After running `make` to make sure the `worker` docker image is created, run `docker-compose up` and this should create a redis and a worker container.

once this is up, from a different terminal, run:

`producersvc 20` this will generate 20 tasks that will be pushed to the redis queue and processed by the Worker service.

You should see some logs in the terminal for the Monitor listing the task statuses. For example:

```
[Monitor] Task task_12 is Done
[Monitor] Task task_16 is InProgress
[Monitor] Task task_18 is Done
[Monitor] Task task_21 is InProgress
[Monitor] Task task_10 is Done
[Monitor] Task task_17 is InProgress
[Monitor] Task task_19 is Done
[Monitor] Task task_22 is InProgress
[Monitor] Task task_14 is Done
[Monitor] Task task_20 is InProgress
[Monitor] Task task_16 is Done
[Monitor] Task task_15 is Done
[Monitor] Task task_17 is Done
```

and from docker-compose output:

```
worker_1  | [worker_77090_1] START consuming task task_18 - sleeping for 1.534s:
worker_1  | Recording task task_18 status InProgress
worker_1  | Recording task task_9 status Done
worker_1  | [worker_77090_0] STOP consuming task task_9:
worker_1  | [worker_77090_0] START consuming task task_19 - sleeping for 1.599s:
worker_1  | Recording task task_19 status InProgress
worker_1  | Recording task task_18 status Done
worker_1  | [worker_77090_1] STOP consuming task task_18:
worker_1  | [worker_77090_1] START consuming task task_21 - sleeping for 1.831s:
worker_1  | Recording task task_21 status InProgress
worker_1  | Recording task task_19 status Done
worker_1  | [worker_77090_0] STOP consuming task task_19:
worker_1  | [worker_77090_0] START consuming task task_22 - sleeping for 4.876s:
worker_1  | Recording task task_22 status InProgress
```

As mentioned before, you can run `producersvc 200` say and start, from an additional terminal, a new worker process with more subWorkers, `workersvc 20` say.

This has been developed and tested on an Ubuntu16.04 VM running:
- go version: 1.11.2
- docker version: 17.09.1-ce
- docker-compose version: 1.24.1

## Comments

### Testing
I provided some unit test for the worker code as an example of how it works and how using interfaces make using test objects convenient. Avoiding dependencies is the objective here. I am used to using testify with its test suite and mocking features but I didn't use it here to keep this simpler.

The rmq library I am using has thought a bit about testing since it provides an in-memory mock for Redis which would prove useful for tests running in Jenkins, as opposed to have to run a Redis sidecar container.

### Code shortcomings
- No graceful shutdown: workers should catch a signal and finish processing the tasks they have dequeued before exiting.
- No failed tasks handling: if a task were to fail, the Monitor could track this and resubmit it assuming retrying makes any sense.
- No Task cancellation: once a task has been pushed to the queue, it cannot be removed from it.
- No proper logging: only prints to stdout for now.
- No proper front end: instead of a simple providersvc app, a REST front end could be provided to allow for a user to sumbit tasks.
- No integration testing.
- No corner case testing: the code may well fall over...

### Scale
The number of workers can be scaled out already with the current code. A limiting factor will then be how many tasks can be pushed in the redis back queue. You could imagine creating a system with multiple queues, priority based maybe, with workers dedicated to a given queue.
This kind of system would be nicely managed by K8s where the queue length could be monitored and a new worker container could be started as the queue fills up.

### Sequential tasks
Once you have a solid parallel franework, imposing a sequential constraint could done by having the Monitor expose and interface where a callback could be registered so that it is called when a given task has reached a given status. The callback would then push a new task in the queue. If the Monitor logic is made more complex (trigger the callback when many tasks have completed succesfully) then this would provide the mechanisms to process task flows that can combine sequential and parallel processing.
