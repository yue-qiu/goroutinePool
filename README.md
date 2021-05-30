## <p align="center">goroutinePool</p>

正如项目名称所示，这是一个 goroutine 池，它避免了在高并发条件下大量创建和销毁协程的消耗。

## Installtion

安装这个包所需要做的就是设置你的 go space 并执行下面的命令行:

```shell
go get github.com/yue-qiu/goroutinePool
```

## Example

```go
package main

import (
    "github.com/yue-qiu/goroutinePool"
    "time"
    "log"
)

func main() {
    pool := goroutinePool.NewGoroutinePool(20, 3 * time.Second)  // 创建最多有 20 个 goroutine 的协程池，这个协程池每 3 秒回收一次不用的协程
    err := pool.Serve()  // 开始服务
    if err != nil {
        log.Fatalln(err)
    }

    // 准备一个测试函数
    var sum uint64 = 0
    var handler = func(params ...interface{}) {
        for i := 0; i < 100; i++ {
            atomic.AddUint64(&sum, 1)
        }
    }

    // 协程的行为被封装到一个 Task 结构体中
    task := goroutinePool.Task{
        Handler: handler,
        Params: []interface{}{},
    }

    for i := 0; i < 100000; i++ {
        err = pool.Put(task)
        if err != nil {
            log.Println(err)
        }

    }

    pool.Stop()
}
```

## BenchMark

CPU: Intel(R) Core(TM) i5-7300HQ CPU @ 2.50GHz

RAM: 16G

go version: 1.15.6 windows/amd64

command:

```go
$ go test -benchmem -run=none -bench=. -benchtime=3s goroutinePool_test.go goroutinePool.go
```

协程池的内存占用大约是原生的 `1/4000`，动态分配次数大约是原生的 `1/1700`。

```
goos: windows
goarch: amd64
BenchmarkPool-4   	       1	1942008200 ns/op	   16024 B/op	      93 allocs/op
BenchmarkWithoutPool-4         2        2374031550 ns/op        69073600 B/op     161406 allocs/op
```
