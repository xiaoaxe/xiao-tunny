// 工作
// author: baoqiang
// time: 2019-05-19 22:02
package funcpool

type HandleFunc func(req interface{}) interface{}

type Worker struct {
	jobChan    chan interface{}
	workerPool chan chan interface{}
	closeChan  chan bool
	retChan    chan interface{}
	fn         HandleFunc
}

func NewWorker(workerPool chan chan interface{}, closeChan chan bool, retChan chan interface{}, fn HandleFunc) *Worker {
	worker := &Worker{
		workerPool: workerPool,
		jobChan:    make(chan interface{}),
		closeChan:  closeChan,
		fn:         fn,
		retChan:    retChan,
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

func (w Worker) Run(req interface{}) {
	w.retChan <- map[string]interface{}{
		"req":  req,
		"resp": w.fn(req),
	}
}
