// 执行函数操作
// author: baoqiang
// time: 2019-05-19 22:00
package funcpool

const num = 10000

type Pool struct {
	jobChan    chan interface{}
	workerPool chan chan interface{}
	exitChan   chan bool
	retChan    chan interface{}
	fn         HandleFunc

	poolSize int
}

func NewPool(poolSize int, fn HandleFunc) *Pool {
	pool := &Pool{
		jobChan:    make(chan interface{}, num),
		workerPool: make(chan chan interface{}, poolSize),
		exitChan:   make(chan bool),
		poolSize:   poolSize,
		retChan:    make(chan interface{}, num),
		fn:         fn,
	}

	pool.start()

	return pool
}

func (p *Pool) Submit(req interface{}) {
	p.jobChan <- req
}

func (p *Pool) Close() {
	close(p.workerPool)
	close(p.jobChan)
	close(p.exitChan)
}

func (p *Pool) start() {
	// go run workers
	for i := 0; i < p.poolSize; i++ {
		w := NewWorker(p.workerPool, p.exitChan, p.retChan, p.fn)
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
