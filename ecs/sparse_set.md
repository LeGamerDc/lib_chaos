## sparse set

***背景***：

​	游戏服务器需要一种数据结构存储大量同类对象（比如 *castle，square* ），根据以往的游戏开发经验，对该数据结构提出以下基本抽象：

1. *set(obj)*: 存放一个对象，给出唯一的标志id
2. *get(id)*: 给定id，定位到与该id绑定的对象地址
3. *delete(id)*: 给定id，执行移除操作，并对之后的 *get(id)* 操作返回 *nil*

目前的 *lmap* 数据结构是一个以上抽象的实现。为了应对对象数量增加对服务器的负担，我们需要对该对象管理数据结构提出新的需求：

1. *no gc*: 我们希望对对象的 *set(obj)*，*delete(id)* 操作均不导致 *gc*，并且我们希望对象的存在本身也不导致 *gc scan*（why? 考虑到单种对象几w-几十w，所有种类对象上百万的数量，并且存在频繁的创建销毁操作）
2. *O(1) 操作*: 我们希望 *set(obj)*，*get(id)*，*delete(id)* 都是常数操作；考虑到 *go map* 的性能问题，我们希望操作中尽可能不使用 *map* 操作。
3. *continuous layout*: 考虑到，我们经常对对象使用 *for each* 操作，我们希望对象能在内存上连续分布。

***思路***：

1. 用 *slice* 来存储对象，用数组下标 *index* 作为对象id，用一个额外的数组记录空闲的位置

   优点是对象连续存储，且访问对象直接寻址非常快。缺点是地址空间被重用后会导致以往的id失效(系统其他地方记录了本对象id，比如说target，在本对象被删除又在原地创建新对象后，该id会指向错误的对象)。

2. id分为两部分，*tick* 和 *index* ，使用 *index* 找到对象后判断 *tick* 相同才报告找到对象。

   如果系统保证单个tick内不会删除，创建phase执行多次，那么使用tick来标记对象；如果无法保证，可以使用其他的自增id来标记对象。还存在的问题是使用单一数组不方便系统扩缩容，并且对内存分配不友好。

3. 模仿 *go mspan* 将 *slice* 分割成等大小的多个 *slice* 用于分配。

   在 *array* 数据结构上创建一个 *sparse set* 数据结构来管理多个 *array*，避免一次分配大量内存。

4. 考虑到我们的 *id* 还有其他的用途，预留 8 个bit的 tag字段来支持 *id* 的其他用于

***实现***：

​	参考[sparse set](https://github.com/LeGamerDc/lib_chaos/blob/master/ecs/sparse_set.go)

**示例代码**：

```go
package main

import (
	"lib_chaos/ecs"
	"math/rand"
)

const (
	SquareSuffix = 3
)

type Square struct {
	id       int64
	troopCnt int64
}

func main() {
	var (
		sqs  = ecs.MakeSparseArray[Square](1024)
		tick int32
		m    = map[int64]int64{}
	)
	for tick = 0; tick < 10; tick++ {
		for i := 0; i < 999; i++ {
			id, sq := sqs.Place(tick, SquareSuffix)
			sq.troopCnt = rand.Int63n(1000)
			sq.id = id
			m[id] = sq.troopCnt
		}
	}
	for id, cnt := range m {
		sq := sqs.Get(id)
		if sq == nil || sq.troopCnt != cnt {
			panic("not equal")
		}
	}
	sqs.Foreach(func(sq *Square) {
		if cnt, ok := m[sq.id]; !ok || cnt != sq.troopCnt {
			panic("not equal")
		}
	})
}
```

---

this lib belongs to *project 8 tick*

