package mr

const (

	/*
	 Status translation:
	 Task: 0, 1, 2
	 Worker: 0, 1, 2, 3

	 For Task : 0 -> 1 -> 2 -> 0
	 For Worker : 0 -> 1 -> 2 -> 0 & if somethings bad happend, 0、1、2 -> 3

	 Woker will init a new gorountine and notify the coordinator during the heartbeat machnism. The worker processor will create a new
	 worker to ask a new task. Task status will be reset to 0 and waitting to be assigned.
	*/

	Idle      = 0 // nothing to do
	InProcess = 1 // Process a task
	Completed = 2 // has completed
	Error     = 3 // only happend in the goroutines

	/*
	 TaskType For worker to choose run mapFunc or reduceFunc
	*/
	MapTask    = 1
	ReduceTask = 2

	MaxRetryTimes = 3

	/*
	 Coordinator RPC Function
	*/
	AssignWorks = "Coordinator.AssignWorks"

	/*
	 some useful defination
	*/
	EmptyPath           = ""
	TaskNotFound  int64 = -1
	FatalTaskType int64 = -1
	FileWriterKey       = "syncWriter"
)