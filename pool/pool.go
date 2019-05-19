// 20个线程池，并行执行任务
// author: baoqiang
// time: 2019-05-19 20:37
package pool

const num = 10000

type Pool struct {
	jobChan    chan interface{}
	workerPool chan chan interface{}
	exitChan   chan bool

	poolSize int
}

func NewPool(poolSize int) *Pool {
	pool := &Pool{
		jobChan:    make(chan interface{}, num),
		workerPool: make(chan chan interface{}, poolSize),
		exitChan:   make(chan bool),
		poolSize:   poolSize,
	}

	pool.start()

	return pool
}

func (p *Pool) Submit(f Runnable) {
	p.jobChan <- f
}

func (p *Pool) Close() {
	close(p.workerPool)
	close(p.jobChan)
	close(p.exitChan)
}

func (p *Pool) start() {
	// go run workers
	for i := 0; i < p.poolSize; i++ {
		w := NewWorker(p.workerPool, p.exitChan)
		w.Start()
	}

	go p.dispatch()
}

func (p *Pool) dispatch() {
	for {
		select {
		case job := <-p.jobChan:
			// got a worker
			w := <-p.workerPool
			// send job
			w <- job
		case <-p.exitChan:
			return
		}
	}

}
