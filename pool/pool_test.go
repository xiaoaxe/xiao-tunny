// 协程池测试
// author: baoqiang
// time: 2019-05-19 20:38
package pool

import (
	"fmt"
	"github.com/githubao/xiao-gogo/gls"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	start := time.Now()

	pool := NewPool(5)

	f := func() {
		fmt.Printf("gid:%d, current: %d\n", gls.GetGoid(), time.Now().UnixNano())
	}

	for i := 0; i <= 100; i++ {
		pool.Submit(f)
	}

	for {
		if time.Since(start) > 5*time.Second {
			break
		}

	}

	fmt.Println("exit pool")
}
