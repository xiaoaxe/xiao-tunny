// worker struct定义
// author: baoqiang
// time: 2019-05-06 18:40
package xiao_tunny

// 代表一个独立的单一请求
type workerRequest struct {
	// 往里面放要运行的任务的参数数据
	jobChan chan<- interface{}

	// 接收任务运行的结果
	retChan <-chan interface{}

	// 用于中断正在运行的任务
	interruptFunc func()
}

// 负责管理多个workerRequest的状态
type workerWrapper struct {
	worker Worker

	// 接收中断信号
	interruptChan chan struct{}

	//往里面放单个请求
	reqChan chan<- workerRequest

	//接收关闭信号
	closeChan chan struct{}

	//关闭函数执行完的结果
	closedChan chan struct{}
}

//新建一个workerWrapper
func newWorkerWrapper(reqChan chan<- workerRequest, worker Worker) *workerWrapper {
	w := workerWrapper{
		worker:        worker,
		reqChan:       reqChan,
		interruptChan: make(chan struct{}),
		closeChan:     make(chan struct{}),
		closedChan:    make(chan struct{}),
	}

	go w.run()

	return &w
}

// 一些workerWrapper的封装函数

func (w *workerWrapper) interrupt() {
	close(w.interruptChan)
	w.worker.Interrupt()
}

func (w *workerWrapper) run() {
	jobChan := make(chan interface{})
	retChan := make(chan interface{})

	// 运行结束的时候进行的函数
	defer func() {
		w.worker.Terminate()
		close(retChan)
		close(w.closedChan)
	}()

	for {
		//init
		w.worker.BlockUntilReady()

		select {
		// 如果能够向请求的chan发送单条的请求
		case w.reqChan <- workerRequest{
			jobChan:       jobChan,
			retChan:       retChan,
			interruptFunc: w.interrupt,
		}:
			select {
			// 工作流里面可以拿到单条的请求数据
			case payload := <-jobChan:
				// 调用函数获取结果
				result := w.worker.Process(payload)
				select {
				// 写返回的数据的结果
				case retChan <- result:
				// do nothing
				// 如果说接收到中断信号的话
				case <-w.interruptChan:
					w.interruptChan = make(chan struct{})
				}

			// 如果读到单条的中断信号
			case _, _ = <-w.interruptChan:
				w.interruptChan = make(chan struct{})
			}

		// 收到关闭的信号
		case <-w.closeChan:
			return

		}

	}
}

// 中断任务
func (w *workerWrapper) stop() {
	close(w.closeChan)
}

// 等待关闭的结果返回
func (w *workerWrapper) join() {
	<-w.closedChan
}
