package goroutinePool

import (
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var sum uint64 = 0

var wg = sync.WaitGroup{}

var handler = func(params ...interface{}) {
	wg.Done()
	for i := 0; i < 1000; i++ {
		atomic.AddUint64(&sum, 1)
	}
}

var runtimes = 1000000

func TestPoolPut(t *testing.T) {
	pool := NewGoroutinePool(100, time.Second)
	err := pool.Serve()
	if err != nil {
		t.Error()
	}

	wg.Add(runtimes)
	for i := 0; i < runtimes; i++ {
		task := Task{
			Handler: handler,
			Params: []interface{}{},
		}

		err = pool.Put(task)
		if err!= nil {
			t.Error()
		}
	}
	wg.Wait()

	if sum != uint64(100 * runtimes) {
		t.Error()
	}
}

func BenchmarkPool_Put(b *testing.B) {
	pool := NewGoroutinePool(100, time.Second)
	pool.Serve()
	defer pool.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runtimes)
		for j := 0; j < runtimes; j++ {
			task := Task{
				Handler: handler,
				Params: []interface{}{},
			}

			err := pool.Put(task)
			if err != nil {
				log.Println(err)
			}
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkNormal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg.Add(runtimes)
		for j := 0; j < runtimes; j++ {
			go handler([]interface{}{}...)
		}
		wg.Wait()
	}
}