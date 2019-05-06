// 测试函数
// author: baoqiang
// time: 2019-05-06 20:03
package xiao_tunny

import "testing"

func TestPoolSizeAdjustment(t *testing.T) {

}

func TestFuncJob(t *testing.T) {
	pool := NewFunc(3, func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	for i := 0; i < 10; i++ {
		ret := pool.Process(i)
		if exp, act := i*2, ret.(int); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestFuncJobTimed(t *testing.T) {

}

func TestCallbackJob(t *testing.T) {

}

func TestTimeout(t *testing.T) {

}

func TestTimedJobsAfterClose(t *testing.T) {

}

func TestJobsAfterClose(t *testing.T) {

}

func TestParallelJobs(t *testing.T) {

}

func TestCustomWorker(t *testing.T) {

}

func BenchmarkFuncJob(b *testing.B) {

}

func BenchmarkFuncTimedJob(b *testing.B) {

}
