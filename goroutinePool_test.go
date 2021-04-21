package goroutinePool

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestPoolPut(t *testing.T) {
	pool := NewGoroutinePool(100, time.Second)
	err := pool.Serve()
	if err != nil {
		t.Error()
	}

	lock := sync.Mutex{}
	handler := func(params ...interface{}) {
		a, b := 0, 1
		for i :=0; i < (params[0]).(int); i++ {
			a, b = b, a + b
		}

		lock.Lock()
		res := (params[1]).(*[]int)
		*res = append(*res, b)
		lock.Unlock()
	}
	res := make([]int, 0)

	for i := 0; i < 100; i++ {
		task := Task{
			Handler: handler,
			Params: []interface{}{i, &res},
		}

		err = pool.Put(task)
		if err!= nil {
			t.Error()
		}
	}

	log.Printf("%v\n", res)
}

func BenchmarkPool_Put(b *testing.B) {
	pool := NewGoroutinePool(100, time.Second)
	pool.Serve()
	lock := sync.Mutex{}
	handler := func(params ...interface{}) {
		a, b := 0, 1
		for i :=0; i < (params[0]).(int); i++ {
			a, b = b, a + b
		}

		lock.Lock()
		res := (params[1]).(*[]int)
		*res = append(*res, b)
		lock.Unlock()
	}
	res := make([]int, 0)

	for i := 0; i < b.N; i++ {
		task := Task{
			Handler: handler,
			Params: []interface{}{i, &res},
		}

		err := pool.Put(task)
		if err != nil {
			log.Println(err)
		}
	}
}

func BenchmarkNormal(b *testing.B) {
	lock := sync.Mutex{}
	handler := func(params ...interface{}) {
		a, b := 0, 1
		for i :=0; i < (params[0]).(int); i++ {
			a, b = b, a + b
		}

		lock.Lock()
		res := (params[1]).(*[]int)
		*res = append(*res, b)
		lock.Unlock()
	}
	res := make([]int, 0)

	for i := 0; i < b.N; i++ {
		go handler([]interface{}{i, &res}...)
	}
}