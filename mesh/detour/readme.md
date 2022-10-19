## detour pathfinding

`detour pathfinding` 是一个非常快速的 suboptimal pathfinding A* 算法，使用多边形mesh 作为A* 的寻路节点，在推导A*时，使用一个节点多边形
与当前节点多边形邻边的中点为参考点来计算`f(v)=g(v)+h(v)`。其中，g(v)为上一个节点的g(v)加上上一个节点的参考点到该节点参考点的直线距离；h(v)为
参考点到终点t的直线距离。

由于g(v)大于从s到参考点的寻路距离，因此，`detour pathfinding` 不保证算法首次弹出的路径为最短距离。`detour pathfinding` 采用去重节点（
每个多边形仅映射为A*的一个节点，多次入堆时仅更新新的参考点和f(v)）来提升寻路效率，因此也无法通过迭代寻路的方法来求最短距离。如果用户对寻路精度
有要求，请使用 `tra pathfinding`