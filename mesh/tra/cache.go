package tra

import "fmt"

type treeNode struct {
	depth  int32
	father int32
	big    bool
}

type Tr struct {
	tree []treeNode
}

// BuildTr will build nav.MTri.GroupId and return a cache of group's connectivity
func BuildTr(nav *NavMesh) *Tr {
	var (
		f   set
		big = make(map[int32]struct{})
	)
	f.init(len(nav.MTri))
	// build group
	var dfs func(int32, int32)
	dfs = func(i, from int32) {
		var (
			t     = &nav.MTri[i]
			merge = nav.countEdge(i) < 3
		)
		t.GroupId = -1
		for l := t.Link; l != -1; l = nav.MLink[l].Next {
			var (
				ni = nav.MLink[l].ToRef
				nt = &nav.MTri[ni]
			)
			if ni == from { // skip father
				continue
			}
			if nt.GroupId != 0 { // already visited -> find big group
				f.union(i, ni)
				continue
			}
			if merge && nav.countEdge(ni) < 3 { // follow group
				f.union(i, ni)
			}
			dfs(ni, i)
		}
	}
	for i := range nav.MTri {
		var t = &nav.MTri[i]
		if t.GroupId == 0 {
			dfs(int32(i), -1)
		}
	}
	//for i := range nav.MTri {
	//	fmt.Printf("%d ", f.find(int32(i)))
	//}
	// build tree
	var (
		dfs2 func(int32, int32)
		inc  int32
		m    = make(map[int32]int32)
	)
	dfs2 = func(i, from int32) {
		var (
			fi = f.find(i)
			t  = &nav.MTri[i]
		)
		if g, ok := m[fi]; ok {
			t.GroupId = g
		} else {
			m[fi] = inc
			t.GroupId = inc
			inc++
		}
		for l := nav.MTri[i].Link; l != -1; l = nav.MLink[l].Next {
			var (
				ni  = nav.MLink[l].ToRef
				nfi = f.find(ni)
				nt  = &nav.MTri[ni]
			)
			if ni == from {
				continue
			}
			if nt.GroupId >= 0 { // visited
				if nfi != fi { // bugs !!!
					fmt.Println("tree bugs")
				}
				big[t.GroupId] = struct{}{}
				//fmt.Println(i, "->", ni)
				continue
			}
			dfs2(ni, i)
		}
	}
	for i := range nav.MTri {
		var t = &nav.MTri[i]
		if t.GroupId < 0 {
			dfs2(int32(i), -1)
		}
	}
	var tree = make([]treeNode, inc)
	for i := range tree {
		tree[i].father = -1
	}
	for i := range nav.MTri {
		var t = &nav.MTri[i]
		for l := nav.MTri[i].Link; l != -1; l = nav.MLink[l].Next {
			var nt = &nav.MTri[nav.MLink[l].ToRef]
			if t.GroupId != nt.GroupId {
				if tree[t.GroupId].father == -1 {
					tree[t.GroupId] = treeNode{
						depth:  tree[nt.GroupId].depth + 1,
						father: nt.GroupId,
					}
				} else if tree[nt.GroupId].father == -1 {
					tree[nt.GroupId] = treeNode{
						depth:  tree[t.GroupId].depth + 1,
						father: t.GroupId,
					}
				}
			}
		}
	}
	for i := range big {
		tree[i].big = true
	}
	var tr = &Tr{tree: tree}
	nav.cache = tr
	return tr
}

type set struct {
	f []int32
}

func (s *set) init(n int) {
	s.f = make([]int32, n)
	for i := 0; i < n; i++ {
		s.f[i] = int32(i)
	}
}
func (s *set) find(a int32) int32 {
	for {
		var fa = s.f[a]
		if fa == a {
			return a
		}
		a = fa
	}
}
func (s *set) union(a, b int32) {
	var (
		fa = s.find(a)
		fb = s.find(b)
	)
	if fa < fb {
		fa, fb = fb, fa
	}
	s.f[fa] = fb
}
