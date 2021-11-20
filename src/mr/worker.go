package mr

import (
	"context"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

/*
 Go pool - The number of goroutine is controled by the NumWorker, which is larger than 1.
 Worker wiill init the pool and when a goroutine exit unexceptedly, a new goroutine will be created.

 TODO: This Go Pool rely on a infinite loop. So as for good use, it should provide a useful way to exit the pool.
		But may not used in this project.
*/

// Task contains params and function and should pass to GoPool
type Task struct {
	Worker func(context.Context)
	Params context.Context
}

func (t *Task) Run() {
	t.Worker(t.Params)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {

	// init goroutines pools and start pool
	pool := NewGoPool(10)
	go pool.Run()

	// build context to pass params
	ctx := context.Background()
	ctx = context.WithValue(ctx, MapTask, mapf)
	ctx = context.WithValue(ctx, ReduceTask, reducef)
	ctx = context.WithValue(ctx, FileWriterKey, &SyncWriter{
		m: sync.Mutex{},
	})

	// use infinite loop to try to create a  go routine
	for {
		pool.Put(
			&Task{
				Params: ctx,
				Worker: work,
			},
		)
	}
}

// work is real process function in order to talk with coordinator
func work(ctx context.Context) {
	var workerID int64

	// Make sure the number of reduce task has been assigned the value.
	for retry := 0; NReduceTask == 0 && retry < MaxRetryTimes; retry++ {
		time.Sleep(2 * time.Second)
	}

	if NReduceTask == 0 {
		return
	}

	// infinite loop to ask for task and do the work

	for {
		var mapf func(string, string) []KeyValue
		var reducef func(string, []string) string
		var syncWriter *SyncWriter
		var ok bool

		// Map function and Reduce function are packed in the context, get it by the key and assert the tyoe.
		if mapf, ok = ctx.Value(MapTask).(func(string, string) []KeyValue); !ok || mapf == nil {
			log.Error("Params Error: get map function from ctx error")
			return
		}

		if reducef, ok = ctx.Value(ReduceTask).(func(string, []string) string); !ok || reducef == nil {
			log.Error("Params Error: get reduce function from ctx error")
			return
		}

		if syncWriter, ok = ctx.Value(FileWriterKey).(*SyncWriter); !ok || syncWriter == nil {
			log.Error("Params Error: get syncWriter from ctx error")
		}

		// ask task from master, try to get a test. If failed, try some times and if all failed, exit.
		taskRequest, taskResponse := &AskTaskRequest{}, &AskTaskResponse{}

		// if workerID == 0, which means this is the first loop
		if workerID != 0 {
			taskRequest.WorkerID = workerID
		}

		ok = rpcCaller(AssignWorks, taskRequest, taskResponse)
		for times := 0; !ok && times < MaxRetryTimes; times++ {
			log.Error("Rpc Caller Error: AssignWorks Error, Start Retry")
			ok = rpcCaller(AssignWorks, taskRequest, taskResponse)
		}

		// retry 3 times and still fails
		if !ok {
			return
		}

		// update workerID
		workerID = taskResponse.WorkerID

		// Do the work according to Type
		if taskResponse.TaskType == MapTask {
			err := mapWorker(mapf, taskResponse)
			if err != nil {
				log.Errorf("Worker Error: MapWorker Error, WorkerID - %d - %v", taskResponse.WorkerID, err)
			}

		} else if taskResponse.TaskType == ReduceTask {
			err := reduceWorker(reducef, taskResponse)
			if err != nil {
				log.Errorf("Worker Error: ReduceWorker Error, WorkerID - %d - %v", taskResponse.WorkerID, err)
			}

		} else {
			// 消息类型错误
			log.Error("TaskType Error: Unknown Type!")
		}
	}
}

// mapWorker is
func mapWorker(mapf func(string, string) []KeyValue, params *AskTaskResponse) error {

	// 读取文件
	file, err := os.Open(params.TaskPath)
	if err != nil {
		log.Errorf("Cannot open %v", params.TaskPath)
		return fmt.Errorf("Cannot open %v", params.TaskPath)
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Cannot read %v", params.TaskPath)
		return fmt.Errorf("Cannot read %v", params.TaskPath)
	}

	// 调用Map函数
	kva := mapf(params.TaskPath, string(content))
	if len(kva) == 0 {
		log.Errorf("Task Error: file contains no content or the mapf exit unexceptedly")
		return fmt.Errorf("Task Error: file contains no content or the mapf exit unexceptedly")
	}

	fdJsonEncoder := make(map[string]*json.Encoder, 0)

	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("System Error: get work path error!")
		return fmt.Errorf("System Error: get work path error!")
	}

	for _, v := range kva {
		// calculate the target ID
		reduceId := int64(ihash(v.Key)) % NReduceTask
		// TODO: Pass intermediate filename to master
		intermediateFileName := "mr-" + strconv.Itoa(int(params.WorkerID)) + "-" + strconv.Itoa(int(reduceId))
		filePath := pwd + "/" + intermediateFileName + ".json"

		// As the flag is os.O_APPEND, it is atonimic opetration and it is safe: https://blog.hotwill.cn/Linux%E5%B9%B6%E5%8F%91%E6%96%87%E4%BB%B6%E8%AF%BB%E5%86%99%E6%80%BB%E7%BB%93.html
		if encoder, ok := fdJsonEncoder[filePath]; !ok || encoder == nil {
			tempFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				log.Errorf("System Error: Open File Error, %v", filePath)
				return fmt.Errorf("System Error: Open File Error, %v", filePath)
			}
			defer tempFile.Close() // not sure whether working

			fdJsonEncoder[filePath] = json.NewEncoder(tempFile)
		}

		// 写入文件
		encoder := fdJsonEncoder[filePath]
		encoder.Encode(&v)
	}

	// 包装已经打开的文件
	var filePaths []string
	for path, _ := range fdJsonEncoder {
		filePaths = append(filePaths, path)
	}

	// call rpc function to return file modified.

	ok := rpcCaller("TaskAck")

	return nil
}

func reduceWorker(reducef func(string, []string) string, params *AskTaskResponse) error {

	//

	return nil
}

// rpcCalller 参数调用封装
func rpcCaller(rpcFunc string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	// type assertion
	var ok bool
	switch args.(type) {
	case *AskTaskRequest:
		args, ok = args.(*AskTaskRequest)
		if !ok {
			log.Error("type assertion error")
		}

		reply, ok = reply.(*AskTaskResponse)
		if !ok {
			log.Error("type assertion error")
		}
	}

	err = c.Call(rpcFunc, args, reply)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

/*
 Utils contains some useful methonds.
*/

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
