package radix

import (
	"fmt"
	"sync"

	"github.com/google/btree"
)

const degree = 128

type Leaf struct {
	parent *Node

	dmu  sync.RWMutex
	data *btree.BTree

	children nodeArray
}

func newLeaf(parent *Node) *Leaf {
	return &Leaf{
		data:   btree.New(degree),
		parent: parent,
	}
}

func (l *Leaf) Parent() *Node {
	return l.parent
}

func (l *Leaf) HasChild(key uint) bool {
	return l.children.Has(key)
}

func (l *Leaf) AddChild(n *Node) {
	prev := l.children.Upsert(n)
	n.parent = l
	if prev != nil {
		panic(fmt.Sprintf("leaf already has child with key %v", n.key))
	}
}

func (l *Leaf) GetChild(key uint) *Node {
	return l.children.Get(key)
}

func (l *Leaf) GetsertChild(key uint) *Node {
	return l.children.Getsert(&Node{key: key})
}

func (l *Leaf) RemoveChild(key uint) *Node {
	return l.children.Delete(key)
}

func (l *Leaf) AscendChildren(cb func(*Node) bool) (ok bool) {
	return l.children.Ascend(cb)
}

func (l *Leaf) AscendChildrenRange(a, b uint, cb func(*Node) bool) (ok bool) {
	return l.children.AscendRange(a, b, cb)
}

func (l *Leaf) Data() []uint {
	l.dmu.RLock()
	defer l.dmu.RUnlock()

	ret := make([]uint, l.data.Len())
	var i int
	l.data.Ascend(func(x btree.Item) bool {
		ret[i] = uint(x.(btreeUint))
		i++
		return true
	})
	return ret
}

func (l *Leaf) Empty() bool {
	if l.children.Len() > 0 {
		return false
	}
	l.dmu.RLock()
	dl := l.data.Len()
	l.dmu.RUnlock()
	return dl == 0
}

func (l *Leaf) Append(v uint) {
	l.dmu.Lock()
	l.data.ReplaceOrInsert(btreeUint(v))
	l.dmu.Unlock()
}

// todo use store
func (l *Leaf) Remove(v uint) (ok bool) {
	l.dmu.Lock()
	ok = l.data.Delete(btreeUint(v)) != nil
	l.dmu.Unlock()
	return
}

func (l *Leaf) Ascend(it Iterator) (ok bool) {
	ok = true
	l.dmu.RLock()
	l.data.Ascend(func(i btree.Item) bool {
		ok = it(uint(i.(btreeUint)))
		return ok
	})
	l.dmu.RUnlock()
	return
}

func LeafInsert(l *Leaf, path Path, value uint, cb nodeIndexer) {
	for {
		if path.Len() == 0 {
			l.Append(value)
			return
		}
		var has bool
		min, max := path.Min(), path.Max()
		// TODO(s.kamardin): use heap sort here to detect max miss factored node.
		// TODO(s.kamardin): when n.key is not in path, maybe get next element in Path
		//                   and seek children to ascend after the index of next pair
		l.AscendChildrenRange(min.Key, max.Key, func(n *Node) bool {
			if v, ok := path.Get(n.key); ok {
				l = n.GetsertLeaf(v)
				path = path.Without(n.key)
				has = true
				return false
			}
			return true
		})
		if !has {
			l.AddChild(makeTree(path, value, cb))
			return
		}
	}
}

// Int implements the Item interface for integers.
type btreeUint uint

func (a btreeUint) Less(b btree.Item) bool {
	return a < b.(btreeUint)
}
