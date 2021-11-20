package mr

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

var (
	NReduceTask int64 // init when mrcoordinator call MkCoordinator and pass nReduce
	NMapTask    int64 // Map
)

type Coordinator struct {
	MapTasks         sync.Map // MapTaskID and filePath
	ReduceTasks      sync.Map // ReduceTask
	MapTaskStatus    sync.Map // Task status
	ReduceTaskStatus sync.Map //
	WorkerStatus     sync.Map // worker status
	IdleWorker       []*int64 // idle worker queue
	TotalWorkers     *int64

	WorkerMapFiles sync.Map

	AllMapOk    bool //
	AllReduceOk bool //
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	if c.AllMapOk && c.AllReduceOk {
		return true
	}
	return false
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// init some status
	NReduceTask = int64(nReduce)

	for _, file := range files {

	}

	c.server()
	return &c
}

// Your code here -- RPC handlers for the worker to call.

// AssigneWorks traverses all maptasks first and then all reduce tasks, try to find a idle task to assigned
// if found, change the worker status and map staus
func (c *Coordinator) AssignWorks(args *AskTaskRequest, reply *AskTaskResponse) error {

	// if worker is new, assign worker ID and init status
	if args.WorkerID == 0 {
		if len(c.IdleWorker) == 0 {
			log.Fatalf("Init Error: no new workerIds, init worker failed")
			return errors.New("no new workers, init failed")
		}

		// get the first ID from idle worker queue
		reply.WorkerID = *c.IdleWorker[0]
		// if everything ok, change the staus when return
		defer func() {
			c.IdleWorker = c.IdleWorker[1:]
			c.WorkerStatus.Store(reply.WorkerID, InProcess)
		}()
	}

	// TODO: try to assign work and update the status
	taskID, taskType := c.FindTheTask()
	if taskID == TaskNotFound || taskType == FatalTaskType {
		log.Fatal("All Task Has Been Done")
		reply.TaskPath = EmptyPath
		reply.TaskType = FatalTaskType
		return nil
	}

	reply.TaskType = taskType
	if taskType == MapTask {
		taskPath, ok := c.MapTasks.Load(taskID)
		if !ok {
			log.Printf("System Error: Load key error - %v", taskID)
		}

		reply.TaskPath, ok = taskPath.(string)
		if !ok {
			log.Printf("System Error: type assertion error - %v", taskPath)
		}

	} else if taskType == ReduceTask {
		taskPath, ok := c.ReduceTasks.Load(taskID)
		if !ok {
			log.Printf("System Error: Load key error - %v", taskID)
		}

		reply.TaskPath, ok = taskPath.(string)
		if !ok {
			log.Printf("System Error: type assertion error - %v", taskPath)
		}
	}

	return nil
}

// TaskACk is a func that worker return the result by this func.
func (c *Coordinator) TaskAck(args *TaskAckRequest, reply *TaskAckResponse) error {
	if args.WorkerID == 0 {
		return errors.New("Params Error: WorkerID equals zero!")
	}

	if args.TaskType == MapTask {

	} else if args.TaskType == ReduceTask {

	} else {
		return errors.New("Params Error: Task Type Error")
	}

	return nil
}

func (c *Coordinator) FindTheTask() (int64, int64) {
	if taskID := c.FindIdleMapTasks(); taskID != TaskNotFound {
		return taskID, MapTask
	}

	if taskID := c.FindIdleReduceTasks(); taskID != -1 {

		return taskID, ReduceTask
	}

	return TaskNotFound, FatalTaskType
}

/*
 Heartbeat module
*/

// TODO：Implement Heartbeat module

/*
 Status check module
*/
func (c *Coordinator) IsFinished() bool {
	if c.FindIdleMapTasks() == TaskNotFound && c.FindIdleReduceTasks() == TaskNotFound {
		return true
	}

	return false
}

func (c *Coordinator) FindIdleMapTasks() int64 {

	var IdleKey int64

	c.MapTaskStatus.Range(func(key, value interface{}) bool {
		intVal := value.(int64)
		if intVal == Idle {
			IdleKey = intVal
			return false
		}
		return true
	})

	if IdleKey != 0 {

	}

	// update the MapStatus
	c.AllMapOk = true
	return TaskNotFound
}

func (c *Coordinator) FindIdleReduceTasks() int64 {
	for key, val := range c.ReduceTaskStatus {
		if val == Idle {
			return key
		}
	}

	// update the reduceStatus
	c.AllReduceOk = true
	return TaskNotFound
}
