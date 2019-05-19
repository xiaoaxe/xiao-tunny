// test func pool
// author: baoqiang
// time: 2019-05-19 22:06
package funcpool

import (
	"fmt"
	"testing"
)

func TestPool(t *testing.T) {
	fn := func(req interface{}) interface{} {
		i := req.(int)
		return i + 1
	}

	pool := NewPool(5, fn)

	for i := 0; i <= 100; i++ {
		pool.Submit(i)
	}

	for i := 0; i <= 100; i++ {
		res := <-pool.retChan
		fmt.Println(res)
	}

	fmt.Println("exit pool")
}
