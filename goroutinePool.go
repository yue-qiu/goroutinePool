package goroutinePool

import (
	"errors"
	"sync"
	"time"
)

type Task struct {
	Handler func(...interface{})
	Params []interface{}
}

type taskChan struct {
	lastUsedTime time.Time
	ch chan Task
}

type Pool struct {
	MaxWorkerCount int
	taskChans []taskChan
	MaxIdleWorkerTime time.Duration
	stop bool
	lock sync.Locker
}

func NewGoroutinePool(maxWorkerCount int, maxIdleWorkerTime time.Duration) *Pool {
	return &Pool{
		MaxIdleWorkerTime: maxIdleWorkerTime,
		MaxWorkerCount: maxWorkerCount,
		taskChans: make([]taskChan, 0),
		stop: false,
		lock: &sync.Mutex{},
	}
}

func (pool *Pool) Serve() error {
	if pool.stop {
		return errors.New("THE POOL HAS BEEN STOPPED")
	}

	go func() {
		for !pool.stop {
			pool.clean()
			time.Sleep(pool.MaxIdleWorkerTime)
		}
	}()

	return nil
}

// FILO
func (pool *Pool) clean() {
	now := time.Now()
	cnt := 0
	for _, taskChan := range pool.taskChans {
		if now.Sub(taskChan.lastUsedTime) > pool.MaxIdleWorkerTime {
			cnt++
		}
	}

	for _, taskChan := range pool.taskChans[: cnt] {
		close(taskChan.ch)
	}

	pool.taskChans = pool.taskChans[cnt:]
}

