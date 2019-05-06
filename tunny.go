// pool实现
// author: baoqiang
// time: 2019-05-06 18:40
package xiao_tunny

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// errors
var (
	ErrPoolNotRunning = errors.New("the pool is not running")
	ErrJobNotFunc     = errors.New("generic worker not given a func()")
	ErrWorkerClosed   = errors.New("worker was closed")
	ErrJobTimeOut     = errors.New("job request timed out")
)

type Worker interface {
	// 实际需要运行的函数
	Process(interface{}) interface{}

	// 任务的初始化函数
	BlockUntilReady()

	// 任务被中断的时候运行的函数
	Interrupt()

	// 任务接收的时候运行的函数
	Terminate()
}

// simple implement of worker
type closureWorker struct {
	processor func(interface{}) interface{}
}

func (w *closureWorker) Process(payload interface{}) interface{} {
	return w.processor(payload)
}
func (w *closureWorker) BlockUntilReady() {}
func (w *closureWorker) Interrupt()       {}
func (w *closureWorker) Terminate()       {}

var _ Worker = &closureWorker{}

//callback implement of worker
type callbackWorker struct {
}

func (w *callbackWorker) Process(payload interface{}) interface{} {
	f, ok := payload.(func())
	if !ok {
		return ErrJobNotFunc
	}
	f()
	return nil
}
func (w *callbackWorker) BlockUntilReady() {}
func (w *callbackWorker) Interrupt()       {}
func (w *callbackWorker) Terminate()       {}

var _ Worker = &callbackWorker{}

// pool manager the workerWrapper
type Pool struct {
	// 任务数量
	queuedJobs int64
	// 创建worker的构造器
	ctor func() Worker
	// 所有的worker
	workers []*workerWrapper
	// 接收负责请求的worker
	reqChan chan workerRequest
	//锁
	workerMut sync.Mutex
}

// 创建一个线程池
func New(n int, ctor func() Worker) *Pool {
	p := &Pool{
		ctor:    ctor,
		reqChan: make(chan workerRequest),
	}

	p.SetSize(n)

	return p
}

//新建一个处理任务的线程池
func NewFunc(n int, f func(interface{}) interface{}) *Pool {
	return New(n, func() Worker {
		return &closureWorker{
			processor: f,
		}
	})
}

//新建一个回调函数的线程池
func NewCallback(n int) *Pool {
	return New(n, func() Worker {
		return &callbackWorker{}
	})
}

// 主运行函数
func (p *Pool) Process(payload interface{}) interface{} {
	atomic.AddInt64(&p.queuedJobs, 1)

	// 读取一个请求
	request, open := <-p.reqChan
	if !open {
		panic(ErrPoolNotRunning)
	}

	//放入当前的请求数据
	request.jobChan <- payload

	//读取返回数据
	payload, open = <-request.retChan
	if !open {
		panic(ErrWorkerClosed)
	}

	atomic.AddInt64(&p.queuedJobs, -1)

	return payload
}

// 带超时的运行主协程函数
func (p *Pool) ProcessTimed(payload interface{}, timeout time.Duration) (interface{}, error) {
	atomic.AddInt64(&p.queuedJobs, 1)
	defer atomic.AddInt64(&p.queuedJobs, -1)

	// 超时时间
	tout := time.NewTimer(timeout)

	// 请求队列定义变量
	var request workerRequest
	var open bool

	// 接收任务的超时
	select {
	case request, open = <-p.reqChan:
		if !open {
			return nil, ErrPoolNotRunning
		}
	case <-tout.C:
		return nil, ErrJobTimeOut
	}

	// 放入任务的超时
	select {
	case request.jobChan <- payload:
		// do nothing
	case <-tout.C:
		request.interruptFunc()
		return nil, ErrJobTimeOut
	}

	// 获取结果的超时
	select {
	case payload, open = <-request.retChan:
		if !open {
			return nil, ErrWorkerClosed
		}
	case <-tout.C:
		request.interruptFunc()
		return nil, ErrJobTimeOut
	}

	// 停止ticker
	tout.Stop()

	// 返回数据
	return payload, nil
}

// 当前运行的任务总数
func (p *Pool) QueueLength() int64 {
	return atomic.LoadInt64(&p.queuedJobs)
}

// 设置协程池的大小
func (p *Pool) SetSize(n int) {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()

	// 当时的worker的数量
	lWorkers := len(p.workers)
	if lWorkers == n {
		return
	}

	//如果目前worker数量不够的话，添加新的worker
	for i := lWorkers; i < n; i++ {
		p.workers = append(p.workers, newWorkerWrapper(p.reqChan, p.ctor()))
	}

	// 如果目前worker数量过多的话，停止一些
	for i := n; i < lWorkers; i++ {
		p.workers[i].stop()
	}

	// 停止的这些等待他们运行结束
	for i := n; i < lWorkers; i++ {
		p.workers[i].join()
	}

	// 设置为新的worker的数量
	p.workers = p.workers[:n]
}

// 获取协程池的大小
func (p *Pool) GetSize() int {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()

	return len(p.workers)
}

// 关闭协程池
func (p *Pool) Close() {
	p.SetSize(0)
	close(p.reqChan)
}
