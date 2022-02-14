package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	// Your definitions here.

	// mutex is to protect the data of Coordinator from concurrent
	mut sync.Mutex

	// cond: cond is used for waiting for the excution of map or reduce tasks,
	// when all map or reduce tasks are assigned but not return within a reasonable time.
	// Hence, use cond.wait() to wait.
	cond sync.Cond

	// nMaps is the number of map tasks
	nMaps int
	// nReduces is the number of reduce tasks
	nReduce int

	// MapTask
	mapTaskFilename []string
	// Map task finish status
	mapFinished []bool
	// MapTaskIssuedTime
	mapTaskIssuedTime []time.Time
	// reduce task finish status
	reduceFinished []bool
	// Reduce Task issued Time, will be use to check if workers processe too long.
	reduceTaskIssuedTime []time.Time

	// isDone indicate whether all tasks are done.
	isDone bool
}

// Your code here -- RPC handlers for the worker to call.

// GetTask assigned a map or reduce(or done) task to a worker.
//
// If mapTask still has new tasks unassigned, it will assign it first.
// when all map tasks are assigned, it will wait until the work finish their tasks and
// send the request to the FinishedNotification. Or awake by the sechdule goroutine to check
// if some tasks are timeout, it will be reschudle to a new wait goroutine.
//
// If all map tasks are finished, the same logic is used to assign reduce tasks.
// After that, isdone is set to true.
func (c *Coordinator) GetTask(args *GetTaskRequestArgs, reply *GetTaskReply) error {
	// lock the coordinator to protect
	c.mut.Lock()
	defer c.mut.Unlock()

	reply.NumsMap = c.nMaps
	reply.NumsReduce = c.nReduce

	// issue the map task
	for {
		mapDone := true
		// loop through all map tasks and find the task available
		for index, isDone := range c.mapFinished {
			if !isDone {
				if c.mapTaskIssuedTime[index].IsZero() ||
					time.Since(c.mapTaskIssuedTime[index]).Seconds() > 10 {
					reply.Filename = c.mapTaskFilename[index]
					reply.TaskNum = index
					reply.Type = Map

					return nil
				} else { // all map tasks are issued but not completed during a reasonable time.
					isDone = false
				}
			}
		}

		if mapDone {
			break
		} else {
			c.cond.Wait()
		}
	}

	// maptask has completed, issue the reduce tasks
	for {
		reduceDone := true
		// loop through all reduce tasks and find the task available
		for index, isDone := range c.reduceFinished {
			if !isDone {
				if c.reduceTaskIssuedTime[index].IsZero() ||
					time.Since(c.reduceTaskIssuedTime[index]).Seconds() > 10 {

					reply.Filename = ""
					reply.TaskNum = index
					reply.Type = Reduce

					return nil
				} else {
					isDone = false
				}
			}
		}

		if reduceDone {
			break
		} else {
			c.cond.Wait()
		}
	}

	// If all map and reduce tasks are done, new request will get the type done task
	// and set the isDone flag to true, and
	reply.Type = Done
	c.isDone = true

	return nil
}

// FinishNotify
func (c *Coordinator) FinishNotify(args *FinishedNotificationArgs, reply *FinishedNotificationResponse) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	//
	switch args.Type {
	case Map:
		//set the status to false
		c.mapFinished[args.TaskNum] = false
	case Reduce:
		// set the status to false
		c.reduceFinished[args.TaskNum] = false
	default:
		fmt.Println("Unknown task type:", args.Type)
	}
	
	//
	return nil
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
	ret := false

	// Your code here.

	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	// init the coordinator
	c := Coordinator{}

	// Your code here.

	// init the coordinator struct
	c.cond = *sync.NewCond(&c.mut)

	c.nMaps = len(files)
	c.nReduce = nReduce

	c.mapTaskFilename = files
	c.mapFinished = make([]bool, len(files))
	c.mapTaskIssuedTime = make([]time.Time, len(files))

	c.reduceFinished = make([]bool, nReduce)
	c.reduceTaskIssuedTime = make([]time.Time, nReduce)

	c.isDone = false

	// go routine to awaken other routines that wait all map or reduce tasks finished
	// using sync.cond
	go func() {
		c.mut.Lock()
		c.cond.Broadcast()
		c.mut.Unlock()
		time.Sleep(1 * time.Second)
	}()

	c.server()
	return &c
}
