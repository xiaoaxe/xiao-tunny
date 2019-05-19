// worker
// author: baoqiang
// time: 2019-05-19 20:43
package pool

type Runnable func()

type Worker struct {
	jobChan    chan interface{}
	workerPool chan chan interface{}
	closeChan  chan bool
}

func NewWorker(workerPool chan chan interface{}, closeChan chan bool) *Worker {
	worker := &Worker{
		workerPool: workerPool,
		jobChan:    make(chan interface{}),
		closeChan:  closeChan,
	}

	return worker
}

func (w Worker) Start() {
	go func() {
		for {
			//put the jobChan to worker pool
			w.workerPool <- w.jobChan

			select {
			case job := <-w.jobChan:
				w.Run(job)
			case <-w.closeChan:
				return
			}
		}
	}()
}

func (w Worker) Run(job interface{}) {
	switch task := job.(type) {
	case Runnable:
		task()
	}
}
