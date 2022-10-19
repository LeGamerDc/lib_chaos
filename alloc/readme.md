## pb allocator

***背景：***

golang 系统中，一个比较大的gc开销是构造 pb 数据结构，我们首先分析一下 pb 对象的特点：

1. 对象存在时间短，被分配后很快就会序列化然后回收；
2. 对象引用关系简单，一个 pb msg 内部虽然存在大量的子对象，但是 pb msg间的没有引用关系，也不会引用非pb msg的对象，当 pb msg 被销毁时，其内部递归引用的对象都会被销毁。
3. pb 对象只有3中类型，结构体(包括raw类型)/slice/string。

考虑到在游戏中大量的消息传输需要构造pb数据结构，我们希望构造pb数据结构的对象不通过内存分配机制，因此不对gc产生压力。

***思路：***

1. 构造一个连续的内存区域，当有内存分配需求时，直接紧凑地从内存区域上分配。类似的有 go arena 库。这里做一下解释：go 开发者长期饱受 go gc 性能地困扰，特别是当系统同时有大量常驻对象和大量临时对象这种特征的时候（go gc 的频率受临时对象增加速度控制，go gc 单次开销受总对象数控制），因此社群常常有人做 unmanaged memory 的库。最终官方尝试在 go 2 开放 arena 库使用，相比玩家自己开发的 unmanaged 库，arena库可以在unmaged memory 中持有正常内存中对象的引用，这个是第三方库无法做到的，但是这个特征存在一定开销，且在我们的这个场景下并不需要。
   *问题：* 内存区域有几个问题需要解决： 1. 什么时候销毁，多少pb msg 共享一个内存区域
2. 有三种方案处理内存区域与 pb msg的对应关系 1. 每个 pb msg 对应一个内存区域，当pb msg 销毁时，内存区域里的内存回收；2. 较多pb msg对应一个内存区域，内存区域设置一个存活时限，到达时限后内存区域被销毁回收；3. 多个 pb msg 对应一个内存区域并记录引用计数，当引用技术为0时销毁回收。这里选择第三种方案，原因不在赘述，主要考虑内存利用效率的因素。

***细节：***

1. 内存区域尾部分配时可能无法满足一个 pb msg，因此一个pb msg 会引用至多2个引用计数。
2. 为了防止内存区域在运行过程中出现没有引用而被回收的情况，引用计数初始为1，当内存区域分配完毕后扣1
3. 为了防治开发者错误使用，数据结构隐藏了引用计数的实现，开发者只需要在一个函数 `CreateMsg` 里创建对象即可。
4. 创建出来的 `Msg` 结构体包含 pb msg，并且实现了 Marshal (调用 msg.Marshal)，因此可以平替原来的 pb msg使用；实现了 Close() 函数（dec 引用），因此在打包完毕后调用Close即可
5. 分配对象过程中使用了内存对齐等技术，保证程序正确运行。
6. 即使pb msg没有正确 dec 引用，内存区域也会因为没有指针引用而被销毁，

***how to use:***

```go
type PbHome struct {  
    Addr string  
    Size int  
}  
  
type PbPerson struct {  
    Name string  
    Age  *int  
    Home *PbHome  
}  
  
var ac alloc.Allocator  
ac.Init()  

msg := ac.CreateMsg(func(buf *alloc.Buf) interface{} {  
	p := alloc.Malloc[PbPerson](buf)  
	p.Name = "john"  
	p.Age = alloc.Malloc[int](buf)  
	*p.Age = 18  
	p.Home = alloc.Malloc[PbHome](buf)  
	*p.Home = PbHome{  
		Addr: "xxx-xxx-xx",  
		Size: 125,  
	}  
	return p  
})  
// do marshal with msg   
_ = msg.Close()  
```

