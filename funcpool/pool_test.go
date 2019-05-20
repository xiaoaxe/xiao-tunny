// test func pool
// author: baoqiang
// time: 2019-05-19 22:06
package funcpool

import (
	"fmt"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	fn := func(req interface{}) interface{} {
		i := req.(int)
		return i + 1
	}

	count := 10000000

	pool := NewPool(5, fn)

	var done = make(chan bool)

	go func() {
		for i := 0; i <= count; i++ {
			pool.Submit(i)
		}
	}()

	//go func() {
	//	for i := 0; i <= count; i++ {
	//		ret := <-pool.retChan
	//		fmt.Println(ret)
	//	}
	//	done <- true
	//}()
	//

	//timeout := time.NewTicker(3 * time.Second)
	timeout := time.NewTicker(100 * time.Millisecond)

	ok := 0

	go func() {
	OUTER:
		for i := 0; i <= count; i++ {
			select {
			//case res := <-pool.retChan:
			//fmt.Println(res)
			case <-pool.retChan:
				ok += 1
			case <-timeout.C:
				fmt.Println("timeout")
				break OUTER
			}
		}

		done <- true

	}()

	<-done

	fmt.Printf("ok len: %d\n", ok)
	fmt.Println("exit pool")
}
