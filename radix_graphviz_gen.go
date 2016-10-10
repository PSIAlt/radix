package radix

import (
	"fmt"
	"io"
	"sync/atomic"
	"unsafe"

	"github.com/google/btree"
)

type id struct {
	c int64
	n map[*node]int64
	l map[*leaf]int64
}

func SearchNode(t *Trie, path Path) *node { return searchNode(t, path) }

func SiftUp(n *node) *node { return siftUp(n) }

func newID() *id {
	return &id{
		n: make(map[*node]int64),
		l: make(map[*leaf]int64),
	}
}

func (id *id) node(n *node) int64 {
	if _, ok := id.n[n]; !ok {
		id.n[n] = id.next()
	}
	return id.n[n]
}

func (id *id) leaf(l *leaf) int64 {
	if _, ok := id.l[l]; !ok {
		id.l[l] = id.next()
	}
	return id.l[l]
}

func (id *id) next() int64 {
	return atomic.AddInt64(&id.c, 1)
}

var markup = map[*node]bool{}

func MarkNode(n *node)   { markup[n] = true }
func UnmarkNode(n *node) { delete(markup, n) }

func Graphviz(w io.Writer, label string, t *Trie) {
	fmt.Fprintf(w, `digraph G {graph[label="%s(%vb)"]; node[style=filled];`, label, t.root.size())
	graphvizLeaf(w, "root", t.root, newID())
	fmt.Fprint(w, `}`)
}

func graphvizNode(w io.Writer, n *node, id *id) int64 {
	i := id.node(n)
	if markup[n] {
		fmt.Fprintf(w, `"%v"[label="%v" fillcolor="#ffffff" color="red"];`, i, n.key)
	} else {
		fmt.Fprintf(w, `"%v"[label="%v" fillcolor="#ffffff"];`, i, n.key)
	}
	for key, l := range n.values {
		lid := graphvizLeaf(w, key, l, id)
		child(w, i, lid)
	}
	if n.parent != nil {
		cid := id.leaf(n.parent)
		parent(w, i, cid)
	}
	return i
}

func graphvizLeaf(w io.Writer, value string, l *leaf, id *id) int64 {
	i := id.leaf(l)
	fmt.Fprintf(w, `"%v"[label="%v" fillcolor="#bef1cf"];`, i, value)
	if l.data.Len() > 0 {
		d := graphvizData(w, l.data, id)
		fmt.Fprintf(w, `"%v"->"%v"[style=dashed,dir=none];`, i, d)
	}
	l.ascendChildren(func(c *node) bool {
		cid := graphvizNode(w, c, id)
		child(w, i, cid)
		return true
	})
	if l.parent != nil {
		cid := id.node(l.parent)
		parent(w, i, cid)
	}
	return i
}

func parent(w io.Writer, a, b int64) {
	fmt.Fprintf(w, `"%v"->"%v"[dir=forward style=dashed color="#cccccc"];`, a, b)
}

func child(w io.Writer, a, b int64) {
	fmt.Fprintf(w, `"%v"->"%v"[dir=forward];`, a, b)
}

func (l *leaf) size() (s uintptr) {
	s += unsafe.Sizeof(l)
	s += unsafe.Sizeof(l.data)
	l.data.Ascend(func(i btree.Item) bool {
		s += unsafe.Sizeof(int(i.(btree.Int)))
		return true
	})
	s += unsafe.Sizeof(l.children)
	l.ascendChildren(func(c *node) bool {
		s += c.size()
		return true
	})
	s += unsafe.Sizeof(l.parent)
	return
}

func (n *node) size() (s uintptr) {
	s += unsafe.Sizeof(n)
	s += unsafe.Sizeof(n.key)
	s += unsafe.Sizeof(n.parent)
	s += uintptr(len(n.val)) + unsafe.Sizeof(n.val)
	s += unsafe.Sizeof(n.values)
	for val, leaf := range n.values {
		s += uintptr(len(val)) + unsafe.Sizeof(val)
		s += leaf.size()
	}
	return
}

func graphvizData(w io.Writer, data *btree.BTree, id *id) int64 {
	var str string
	data.Ascend(func(i btree.Item) bool {
		str += fmt.Sprintf("%v;", int(i.(btree.Int)))
		return true
	})
	n := id.next()
	fmt.Fprintf(w, `"%v"[label="%v" fillcolor="#cccccc" shape=polygon];`, n, str)
	return n
}

func nextID(id *int64) int64 { return atomic.AddInt64(id, 1) }
