## triangulation reduce pathfinding

`tra pathfinding` 是较为快速的 optimal pathfinding A* 算法，其参考了[TRA*](https://www.aaai.org/Papers/AAAI/2006/AAAI06-148.pdf)
和 [Polyanya](https://www.ijcai.org/proceedings/2017/0070.pdf) 算法的优点，使用多边形边上的一条线段+一个点的方式来作为A* 的节点。
每个多边形可以用不同的线段+点的方式入堆，因此对寻路效率有所影响，但是保证了 g(v) == cost(s, v)即保证了结果为最短路径。

`tra pathfinding` 还使用mesh 分组的方案来加速寻路速度。



[further reading](http://www.aamircheema.com/research/ESPP_AIJ2021.pdf)