package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// define a new type for task
type TaskTyppe int

const (
	// Map task: read the words from file and generate the temporary file
	Map TaskTyppe = 1
	// Reduce task: read the word count from the temporary file and generate the final files.
	Reduce TaskTyppe = 2
	// Done task: indicate that all tasks has been done, workers should exit.
	Done TaskTyppe = 3
)

type GetTaskRequestArgs struct {
}

type GetTaskReply struct {
	// NumsMap: total numbers of map tasks, used by Reduce task to read the tempfile.
	// for example, reduce task number is 2, so it should read the file from temp-0-2 to temp-NumsMap-2.
	NumsMap int
	// NumsReduce: total number of reduce tasks, used by Map task to generate the tempfile.
	// for example, Map task number is 2, so it will generate the file temp-2-<ishash(key)%NumsReduce>.
	NumsReduce int

	// TaskNum indicate the which task it is , relate to the Type.
	TaskNum int
	// Type indictae which type it is, include the map, reduce and done.
	Type TaskTyppe
	// filename will be used in Map task to read the original file.
	Filename string
}

type FinishedNotificationArgs struct {
	// TaskNum indicate the which task it is , relate to the Type.
	TaskNum int
	// Type indictae which type it is, include the map, reduce and done.
	Type TaskTyppe
}

type FinishedNotificationResponse struct {
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
