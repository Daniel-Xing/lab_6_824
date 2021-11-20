package mr

// GoPool is a routine pool, you should use like this :
// 1. pool := NewGoPool(nums)
// 2. go pool.Run()
// 3. pool.Put(Task)
type GoPool struct {
	MaxRoutine    int64
	Task          chan *Task
	ControlSignal chan int64
}

func (g *GoPool) Put(t *Task) {
	g.ControlSignal <- 1
	g.Task <- t
}

func (g *GoPool) Worker(t *Task) {
	t.Run()
	<-g.ControlSignal
}

func (g *GoPool) Run() {
	for {
		select {
		case t := <-g.Task:
			go g.Worker(t)
		}
	}
}

func NewGoPool(maxNum int64) *GoPool {
	Task := make(chan *Task, maxNum)
	ControlSignal := make(chan int64, maxNum)

	return &GoPool{
		MaxRoutine:    maxNum,
		Task:          Task,
		ControlSignal: ControlSignal,
	}
}