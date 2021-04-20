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
	taskChans []*taskChan
	MaxIdleWorkerTime time.Duration
	stop bool
	lock sync.Locker
}

func NewGoroutinePool(maxWorkerCount int, maxIdleWorkerTime time.Duration) *Pool {
	return &Pool{
		MaxIdleWorkerTime: maxIdleWorkerTime,
		MaxWorkerCount: maxWorkerCount,
		taskChans: make([]*taskChan, 0),
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

func (pool *Pool) clean() {
	now := time.Now()
	cnt := 0
	pool.lock.Lock()
	defer pool.lock.Unlock()

	for _, taskChan := range pool.taskChans {
		if now.Sub(taskChan.lastUsedTime) > pool.MaxIdleWorkerTime {
			cnt++
			close(taskChan.ch)
		}
	}

	pool.taskChans = pool.taskChans[cnt:]
}

// FIFO
func (pool *Pool) getTaskChan() *taskChan {
	var ch *taskChan
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if len(pool.taskChans) > 0 {
		ch = pool.taskChans[0]
		pool.taskChans = pool.taskChans[1:]
	} else {
		ch = &taskChan{
			ch: make(chan Task),
			lastUsedTime: time.Now(),
		}
		pool.taskChans = append(pool.taskChans, ch)
	}

	return ch
}
