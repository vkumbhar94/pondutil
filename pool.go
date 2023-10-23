package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/alitto/pond"
)

// Task defines unit work function
type Task func() (any, error)

// Pool extends the pond pool with more features
type Pool struct {
	*pond.WorkerPool
}

// New creates a new pool (with maxWorkers, maxCapacity, opts) and returns created pool
func New(maxWorkers, maxCapacity int, opts ...pond.Option) *Pool {
	pool := Pool{
		pond.New(maxWorkers, maxCapacity, opts...),
	}
	return &pool
}

func RunTasks[R any](p *Pool, ctx context.Context, tasks ...Task) ([]R, error) {
	return RunTasksWithSupplierFunc[R](p, ctx, func(outch chan<- Task) {
		for _, task := range tasks {
			outch <- task
		}
	})
}
func RunTasksWithSupplierFunc[R any](p *Pool, ctx context.Context, fn func(chan<- Task)) ([]R, error) {
	inch := make(chan Task)
	var result []R
	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		result, err = RunTasksWithSupplierChan[R](p, ctx, inch)
		wg.Done()
	}()
	fn(inch)
	close(inch)
	wg.Wait()
	return result, err
}

func RunTasksWithSupplierChan[R any](p *Pool, ctx context.Context, inch <-chan Task) ([]R, error) {
	tg, _ := p.GroupContext(ctx)

	var result []R
	ch := make(chan R)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for chanResult := range ch {
			result = append(result, chanResult)
		}

		wg.Done()
	}()

	for task := range inch {
		cpy := task
		tg.Submit(func() error {
			r, err := cpy()
			if err != nil {
				return err
			}
			if v, ok := r.(R); ok {
				ch <- v
			} else {
				panic(fmt.Sprintf("result is not of expected type: %T", r))
			}
			return nil
		})
	}
	err := tg.Wait()
	if err != nil {
		return nil, err
	} // wait to complete task execution
	close(ch) // close result chan
	wg.Wait() // wait to collect all results

	return result, nil
}

////
////

func RunTaskGroupWithSupplierFunc[R any](tg *pond.TaskGroupWithContext, fn func(chan<- Task)) ([]R, error) {
	inch := make(chan Task)
	var result []R
	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		result, err = RunTaskGroupWithSupplierChan[R](tg, inch)
		wg.Done()
	}()
	fn(inch)
	close(inch)
	wg.Wait()
	return result, err
}

func RunTaskGroupWithSupplierChan[R any](tg *pond.TaskGroupWithContext, inch <-chan Task) ([]R, error) {

	var result []R
	ch := make(chan R)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for chanResult := range ch {
			result = append(result, chanResult)
		}

		wg.Done()
	}()

	for task := range inch {
		cpy := task
		tg.Submit(func() error {
			r, err := cpy()
			if err != nil {
				return err
			}
			if v, ok := r.(R); ok {
				ch <- v
			} else {
				panic(fmt.Sprintf("result is not of expected type: %T", r))
			}
			return nil
		})
	}
	err := tg.Wait()
	if err != nil {
		return nil, err
	} // wait to complete task execution
	close(ch) // close result chan
	wg.Wait() // wait to collect all results

	return result, nil
}
