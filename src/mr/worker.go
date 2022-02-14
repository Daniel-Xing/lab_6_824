package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
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

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// loop
	for {
		// get the task from the map
		args := GetTaskRequestArgs{}
		reply := GetTaskReply{}

		call("Coordinator.GetTask", &args, &reply)

		switch reply.Type {
		case Map: // if task is map, process map tasks
			MapWorker(&reply, mapf)
		case Reduce: // if task is reduce, process reduce tasks
			ReduceWorker(&reply)
		case Done: // if task is Done, exit
			os.Exit(0)
		default:
			fmt.Println("Unknow task type: ", reply.Type)
		}

		// After process the tasks, sene the finished tasks to the coordinator.b
		finishArgs := FinishedNotificationArgs{
			TaskNum: reply.TaskNum,
			Type: reply.Type,
		}

		finishReply := FinishedNotificationResponse{}
		call("Coordinator.FinishNotify", &finishArgs, &finishReply)
	}
}

func MapWorker(getReply *GetTaskReply, mapf func(string, string) []KeyValue) {
	// 
	file, err := os.Open(getReply.Filename)
	if err != nil {
		log.Fatalf("cannot open %v", getReply.Filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", getReply.Filename)
	}
	file.Close()
	kva := mapf(getReply.Filename, string(content))

	// create the temp file
	tempFile, err := ioutil.TempFile("", "map")
	tempFileName := tempFile.Name()
	if err != nil {
		log.Fatal("cannot create the temp file, task id: %d", getReply.TaskNum)
	}

	// encode the kv value into the json
	enc := json.NewEncoder(tempFile)
	for _, kv := range kva {
		enc.Encode(&kv)
	}
	tempFile.Close()

	// rename the file
	os.Rename(tempFileName, getFinalMapName(getReply.TaskNum))
}

func ReduceWorker(getReply *GetTaskReply) {

}


//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
