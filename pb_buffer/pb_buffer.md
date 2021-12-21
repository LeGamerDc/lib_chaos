---
layout: default
title: pb buffer
---

## pb buffer

***简介***：

​	游戏服务器每时每刻都序列化大量的数据发送给玩家客户端及其他周边服务器，这些数据存在数据量大，对象多，生存周期短的特征，对服务器的gc带来了严重的负担。如果能接管这些对象的分配，销毁，不通过 *golang runtime* 的 *gc* ，那么势必将大幅降低服务器的负。不过要么受限于代码量，要么受限于简洁性，我们一直都无法实现这个库。不过随着 *golang generics* 的诞生，我们终于可以一劳永逸地解决这个问题了。

我们先来看一下 *pb* 生成的数据主要有哪些内容：

1. 基本类型，不包括string。如 int，byte，float64
2. string
3. 嵌入结构体
4. 指针结构体
5. slice of plain data
6. slice of pointer

所有这些对象可以分为3种分配方式：

1. 指针分配，包括基本类型的指针和结构体指针
2. slice 分配，包括元素为结构体，基本类型，指针类型的slice
3. string 分配，仅string

所以 *PbBuffer* 支持的 *function* 为：

```go
// 指针分配
func Malloc[T any](buf *PbBuffer) *T

// slice 分配
func MallocSlice[T any](buf *PbBuffer, l, c int) []T

// string 分配，并拷贝
func CopyString(buf *PbBuffer, s string) string

// 销毁 回收
func (buf *PbBuffer) Destroy()
```

***实现***：

​	参考[pb buffer](https://github.com/LeGamerDc/lib_chaos/blob/master/pb_buffer/pb_buffer.go)

***示例代码***：

```go
package main

import (
	"fmt"
	"lib_chaos/pb_buffer"
)

type Person struct {
	name  string
	age   int
	house *House
}

type House struct {
	addr string
	size int
}

func (p *Person) String() string {
	if p.house != nil {
		return fmt.Sprintf("name: %s, age %d, house:[%s]", p.name, p.age, p.house)
	} else {
		return fmt.Sprintf("name: %s, age %d, homeless", p.name, p.age)
	}
}

func (h *House) String() string {
	return fmt.Sprintf("addr: %s, size: %d", h.addr, h.size)
}

func main() {
	var buf = new(pbBuffer.PbBuffer)
	var p = pbBuffer.Malloc[Person](buf)
	p.name = pbBuffer.CopyString(buf, "john")
	p.age = 22
	p.house = pbBuffer.Malloc[House](buf)
	p.house.addr = pbBuffer.CopyString(buf, "天府新区-兴隆湖-xx街-21-1")
	p.house.size = 169
	fmt.Println(buf.Explain())
	fmt.Println(p)
	buf.Destroy()
}
```

---

this lib belongs to *project 8 tick*