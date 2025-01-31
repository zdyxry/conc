package pool

import (
	"context"
)

// ResultErrorPool is a pool that executes tasks that return a generic result
// type and an error. Tasks are executed in the pool with Go(), then the
// results of the tasks are returned by Wait().
//
// The order of the results is not guaranteed to be the same as the order the
// tasks were submitted. If your use case requires consistent ordering,
// consider using the `stream` package or `Map` from the `iter` package.
type ResultErrorPool[T any] struct {
	errorPool      ErrorPool
	agg            resultAggregator[T]
	collectErrored bool
}

// Go submits a task to the pool.
func (p *ResultErrorPool[T]) Go(f func() (T, error)) {
	p.errorPool.Go(func() error {
		res, err := f()
		if err == nil || p.collectErrored {
			p.agg.add(res)
		}
		return err
	})
}

// Wait cleans up any spawned goroutines, propagating any panics and
// returning the results and any errors from tasks.
func (p *ResultErrorPool[T]) Wait() ([]T, error) {
	err := p.errorPool.Wait()
	return p.agg.results, err
}

// WithCollectErrored configures the pool to still collect the result of a task
// even if the task returned an error. By default, the result of tasks that errored
// are ignored and only the error is collected.
func (p *ResultErrorPool[T]) WithCollectErrored() *ResultErrorPool[T] {
	p.collectErrored = true
	return p
}

// WithContext converts the pool to a ResultContextPool for tasks that should
// be canceled on first error.
func (p *ResultErrorPool[T]) WithContext(ctx context.Context) *ResultContextPool[T] {
	return &ResultContextPool[T]{
		contextPool: *p.errorPool.WithContext(ctx),
	}
}

// WithFirstError configures the pool to only return the first error
// returned by a task. By default, Wait() will return a combined error.
func (p *ResultErrorPool[T]) WithFirstError() *ResultErrorPool[T] {
	p.errorPool.WithFirstError()
	return p
}

// WithMaxGoroutines limits the number of goroutines in a pool.
// Defaults to unlimited. Panics if n < 1.
func (p *ResultErrorPool[T]) WithMaxGoroutines(n int) *ResultErrorPool[T] {
	p.errorPool.WithMaxGoroutines(n)
	return p
}
