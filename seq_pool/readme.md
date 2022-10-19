# seq_pool

## 背景介绍

假设我们有一个编程模型如下，其中dispatch被串行调用，
produce根据job参数生产结果且produce占用较多cpu，
consume消费结果且consume被要求按job顺序执行。

```go
func dispatch(job) {
    result := produce(job) // cpu intensive
    consume(result) // result must be consume in sequence
}
```

那么一种优化模型为:

```go
func dispatch(job) {
    var result
    var done

    p4.Go(func() { // 4个goroutine执行
        result = produce(job)
        done = true
    })

    p1.Go(func()) { // 1个goroutine执行
        while !done {
            time.Sleep(time.MilliSecond)    
        }
        consume(result)
    })
}
```

实际上，我们目前的 land sender 正好是是这样的功能：

```go
func dispatch(msg Msg) {
    data := pack(msg)   // cpu intensive
    send(data)          // 被要求串行发送
}
```

## 性能优化

上述模型的一个简单实现方式是使用 `gopool`:

```go
import "github.com/bytedance/gopkg/util/gopool"
p4 := gopool.NewPool("parallel", 4, gopool.NewConfig())
p1 := gopool.NewPool("sequence", 1, gopool.NewConfig())
```

gopool 使用链表存放`task`队列，多个 goroutine 通过锁来竞争访问队列。
单是很可惜，这种实现方式在我们的环境下不太理想，原因在于提升效率比。

当 produce cpu开销逐渐降低时，锁开销以及链表操作开销就会逐渐变大，直到
并行程序相对于串行没有提升，不仅如此，次数进程还会耗费 4~5 倍的cpu。
我们称引起并行程序没有提升的这个临界点(produce开销)为效率阀。上述模型的简单实现
的效率阀为 300ns，即当produce的平均开销为 300ns及以下时，并行程序不如串行程序。

所以，我们需要尽可能降低效率阀，以提升并行程序提升效率。
我们的优化方式为：
* 去掉链表改用数组，降低memory deference
* 去掉锁/channel，改用lock free队列，降低开销。

## 代码详细解释

文档请配合[代码](https://github.com/LeGamerDc/lib_chaos/blob/master/seq_pool/pool.go)食用

先看一下代码：

```
var a string
var done bool // variable done is synchronize variable

go func() {
    a = "hello"
    done = true
}

go func() {
    for !done {}
    fmt.Println(a) // a can be empty string
}
```

上面代码中 done 的功能是一个同步变量，但是这里的用法是错的，原因在于Go
可能在编译期 reorder code，如何避免 Go reorder 的影响呢，方案在于
对 done 做原子操作， Go 不会对原子操作的语句进行 reorder。并且注意：
编译后的 atomic.LoadInt32/atomic.StoreInt32 与直接读写的汇编代码
完全一样，只不过不会对代码 reorder。

sender_pool 有2个地方使用到了同步变量。
1. prepareIdx 同步 queue[prepareIdx]
2. Msg.done 同步了 Msg.ptr

同时要注意，Dispatch 可能被多个goroutine 同时调用，所以做了加锁操作。
在这个情况下，操作 prepareIdx 仍然需要使用atomic 操作，因为这里跟
produce/consume 里的 prepareIdx queue 构成了同步代码。
