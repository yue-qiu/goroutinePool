package goroutinePool

import (
	"errors"
	"log"
	"math/rand"
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
	currentWorkerCount int64
	taskChans []*taskChan
	MaxIdleWorkerTime time.Duration
	stop bool
	lock sync.Locker
}

func NewGoroutinePool(maxWorkerCount int, maxIdleWorkerTime time.Duration) *Pool {
	return &Pool{
		MaxIdleWorkerTime: maxIdleWorkerTime,
		MaxWorkerCount: maxWorkerCount,
		taskChans: make([]*taskChan, maxWorkerCount),
		stop: false,
		lock: &sync.Mutex{},
	}
}

func (pool *Pool) getCurrentWorkerCount() int {
	return int(pool.currentWorkerCount)
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
	var tmp = make([]*taskChan, pool.getCurrentWorkerCount())

	for _, taskCn := range pool.taskChans {
		if taskCn != nil {
			if now.Sub(taskCn.lastUsedTime) >= pool.MaxIdleWorkerTime {
				pool.currentWorkerCount--
				close(taskCn.ch)
			} else {
				tmp[cnt] = taskCn
				cnt++
			}
		}
	}

	m := copy(pool.taskChans, tmp[:cnt])
	for i := m; i < pool.MaxWorkerCount; i++ {
		pool.taskChans[i] = nil
	}
}

func (pool *Pool) consume(taskCh *taskChan) {
	// create a goroutine for this taskChan
	go func() {
		// in case of panic
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

func (pool *Pool) getTaskChan() *taskChan {
	var taskCh *taskChan

	if pool.getCurrentWorkerCount() < pool.MaxWorkerCount {
		// create a new worker
		taskCh = &taskChan{
			ch: make(chan Task, 0),
		}
		pool.taskChans[pool.getCurrentWorkerCount()] = taskCh
		pool.currentWorkerCount++
		pool.consume(taskCh)
	} else { // random choose a old worker
		taskCh = pool.taskChans[rand.Intn(pool.getCurrentWorkerCount())]
	}

	return taskCh
}

func (pool *Pool) Put(task Task) error {
	pool.lock.Lock()
	if pool.stop {
		return errors.New("THE POOL HAS BEEN STOPPED")
	}

	taskCh := pool.getTaskChan()
	taskCh.ch <-task
	pool.lock.Unlock()

	return nil
}

func (pool *Pool) Stop() {
	pool.stop = true

	for _, taskCh := range pool.taskChans {
		if taskCh != nil {
			close(taskCh.ch)
		}
	}
}
