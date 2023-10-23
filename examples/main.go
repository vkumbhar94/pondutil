package main

import (
	"context"
	"fmt"
	"stash.logicmonitor.com/dev/log-intelligence.git/pkg/pool"
	"sync"
)

func main() {
	p := pool.New(2, 3)
	result, err := pool.RunTasksWithSupplierFunc[int64](p, context.Background(), func(ch chan<- pool.Task) {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for i := 1; i <= 10; i++ {
				v := i
				ch <- func() (any, error) {
					return int64(v), nil
				}
			}
			wg.Done()
		}()
		go func() {
			for i := 11; i <= 20; i++ {
				v := i
				ch <- func() (any, error) {
					return int64(v), nil
				}
			}
			wg.Done()
		}()
		wg.Wait()
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
