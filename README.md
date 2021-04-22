## <p align="center">goroutinePool</p>
As the project name indicates, this is a goroutine pool, which avoid the large amount of consumption of creation and destruction under heigh concurrency.

## Installtion

All you need to do for installing this package is to setup your go space and execute the command line below:

```
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
    pool := goroutinePool.NewGoroutinePool(20, 3 * time.Second)
    err := pool.Serve()
    if err != nil {
        log.Fatalln(err)
    }

    var sum uint64 = 0

    var handler = func(params ...interface{}) {
        for i := 0; i < 100; i++ {
            atomic.AddUint64(&sum, 1)
        }
    }

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
```
goos: windows
goarch: amd64
BenchmarkPool-4                        1        2120999000 ns/op           67160 B/op        415 allocs/op
BenchmarkWithoutPool-4                 1        2090998000 ns/op        148155008 B/op    347748 allocs/op
```
