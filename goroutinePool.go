package goroutinePool

import (
	"errors"
	"log"
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
	currentWorkerCount int
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
	var taskCh *taskChan
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if len(pool.taskChans) > 0 {
		taskCh = pool.taskChans[0]
		pool.taskChans = pool.taskChans[1:]
	} else {
		taskCh = &taskChan{
			ch: make(chan Task),
		}
	}

	return taskCh
}

func (pool *Pool) consume(taskCh *taskChan) {
	// 为这个 taskChan 开启一个对应的 goroutine
	go func() {
		// 防止 task 运行时发生 panic 导致整个程序崩溃
		defer func() {
			if r := recover(); r != nil {
				log.Printf("task panic: %s\n", r)
			}
		}()

		for {
			if task, ok := <-taskCh.ch; ok {
				taskCh.lastUsedTime = time.Now()
				task.Handler(task.Params...)
			} else {
				break
			}
		}
	}()
}

func (pool *Pool) Put(task Task) error {
	if pool.stop {
		return errors.New("THE POOL HAS BEEN STOPPED")
	}

	taskCh := pool.getTaskChan()
	if len(pool.taskChans) < pool.MaxWorkerCount {
		pool.consume(taskCh)
		taskCh.ch <-task
		pool.taskChans = append(pool.taskChans, taskCh)
		return nil
	}

	return errors.New("TOO MANY TASKCHAN")
}

func (pool *Pool) Stop() {
	pool.stop = true

	for i := range pool.taskChans {
		close(pool.taskChans[i].ch)
	}
}
