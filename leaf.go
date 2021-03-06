package radix

//go:generate ppgo

import (
	"fmt"
	"sync"

	"github.com/google/btree"
)

const (
	degree = 128
)

type Leaf struct {
	parent *Node
	value  string

	// dmu holds mutex for data manipulation.
	dmu sync.RWMutex

	// If leaf data is at most array.Cap(), uint sorted array is used.
	// Otherwise BTree will hold the data.
	array uintArray
	btree *btree.BTree

	children *nodeSyncSlice
}

// newLeaf creates leaf with parent node.
func NewLeaf(parent *Node, value string) *Leaf {
	return &Leaf{
		parent:   parent,
		value:    value,
		children: &nodeSyncSlice{},
	}
}

func (l *Leaf) Parent() *Node {
	return l.parent
}

func (l *Leaf) Value() string {
	return l.value
}

func (l *Leaf) HasChild(key uint) bool {
	return l.children.Has(key)
}

func (l *Leaf) AddChild(n *Node) {
	prev, _ := l.children.Upsert(n)
	n.parent = l
	if prev != nil {
		panic(fmt.Sprintf("leaf already has child with key %v", n.key))
	}
}

func (l *Leaf) GetChild(key uint) *Node {
	n, _ := l.children.Get(key)
	return n
}

func (l *Leaf) ChildrenCount() int {
	return l.children.Len()
}

func (l *Leaf) GetsertChild(key uint) (node *Node, inserted bool) {
	node = l.children.GetsertFn(key, func() *Node {
		inserted = true
		return &Node{
			key:    key,
			parent: l,
		}
	})
	return
}

func (l *Leaf) RemoveChild(key uint) *Node {
	prev, _ := l.children.Delete(key)
	return prev
}

func (l *Leaf) RemoveEmptyChild(key uint) (*Node, bool) {
	return l.children.DeleteCond(key, (*Node).Empty)
}

func (l *Leaf) AscendChildren(cb func(*Node) bool) (ok bool) {
	return l.children.Ascend(cb)
}

func (l *Leaf) AscendChildrenRange(a, b uint, cb func(*Node) bool) (ok bool) {
	return l.children.AscendRange(a, b, cb)
}

func (l *Leaf) GetAny(it func() (uint, bool)) (*Node, bool) {
	return l.children.GetAny(it)
}

func (l *Leaf) GetsertAny(it func() (uint, bool), add func() *Node) *Node {
	return l.children.GetsertAnyFn(it, add)
}

func (l *Leaf) AppendTo(p []uint) []uint {
	l.dmu.RLock()
	if l.btree != nil {
		l.btree.Ascend(func(x btree.Item) bool {
			p = append(p, uint(x.(btreeUint)))
			return true
		})
		l.dmu.RUnlock()
		return p
	}
	array := l.array
	l.dmu.RUnlock()

	return array.AppendTo(p)
}

func (l *Leaf) Empty() bool {
	if l.children.Len() > 0 {
		return false
	}
	return l.ItemCount() == 0
}

func (l *Leaf) ItemCount() int {
	l.dmu.RLock()
	var n int
	if l.btree != nil {
		n = l.btree.Len()
	} else {
		n = l.array.Len()
	}
	l.dmu.RUnlock()
	return n
}

// Append appends v to leaf values.
// It returns true if v was not present there.
// Note that zero v will not report correct ok value.
func (l *Leaf) Append(v uint) (ok bool) {
	l.dmu.Lock()
	switch {
	case l.array.Len() == l.array.Cap():
		l.btree = btree.New(degree)
		l.array.Ascend(func(v uint) bool {
			l.btree.ReplaceOrInsert(btreeUint(v))
			return true
		})
		l.array = l.array.Reset()
		fallthrough

	case l.btree != nil:
		prev := l.btree.ReplaceOrInsert(btreeUint(v))
		ok = prev == nil

	default:
		var replaced bool
		l.array, _, replaced = l.array.Upsert(v)
		ok = !replaced
	}
	l.dmu.Unlock()

	return
}

// Remove removes v from leafs values. It returns true if v was present there.
func (l *Leaf) Remove(v uint) (ok bool) {
	l.dmu.Lock()
	if l.btree != nil {
		ok = l.btree.Delete(btreeUint(v)) != nil
		if l.btree.Len() == 0 {
			l.btree = nil
		}
	} else {
		l.array, _, ok = l.array.Delete(v)
	}
	l.dmu.Unlock()
	return
}

func (l *Leaf) Ascend(it Iterator) bool {
	var (
		ok = true
	)
	l.dmu.RLock()
	if l.btree != nil {
		l.btree.Ascend(func(i btree.Item) bool {
			ok = it(uint(i.(btreeUint)))
			return ok
		})
		l.dmu.RUnlock()
		return ok
	}
	array := l.array
	l.dmu.RUnlock()

	return array.Ascend(it)
}

// Inserter contains options for inserting values into the tree.
type Inserter struct {
	// IndexNode is a callback that will be called on every newly created Node.
	IndexNode func(*Node)

	// NodeOrder is an order of node keys, that should be kept during insertion.
	// That is, when we insert path {1:a;2:b;3:c} and NodeOrder is [2,3],
	// the tree will looks like 2:b -> 3:c -> 1:a.
	NodeOrder []uint
}

// Insert inserts value to the leaf that exists (or not and will be created) at
// the given path starting with the leaf as root.
//
// It first inserts/creates nodes from the Inserter's NodeOrder field.
// Then it takes first node for which there are key and value in the path.
// If at the current level there are no such nodes, it creates one with some
// key from the path.
//
// It returns true if value was not present in target leaf's values.
func (c Inserter) Insert(leaf *Leaf, path Path, value uint) bool {
	_, ok := c.insert(leaf, path, value, true)
	return ok
}

// GetLeaf returns Leaf after given root by given path.
// If path is empty root leaf is returned.
func (c Inserter) GetLeaf(leaf *Leaf, path Path) *Leaf {
	leaf, _ = c.insert(leaf, path, 0, false)
	return leaf
}

func (c Inserter) insert(leaf *Leaf, path Path, value uint, insert bool) (*Leaf, bool) {
	// First we should save the fixed order of nodes.
	for _, key := range c.NodeOrder {
		if val, ok := path.Get(key); ok {
			n, inserted := leaf.GetsertChild(key)
			if inserted && c.IndexNode != nil {
				c.IndexNode(n)
			}
			leaf = n.GetsertLeaf(val)
			path = path.Without(key)
		}
	}

	for path.Len() > 0 {
		// Get the cursor of path begining.
		cur := path.Begin()

		// First try to find an existance of any path key in leaf children.
		// Due to the trie usage pattern, it is probably exists already.
		// If we do just lookup, leaf will not be locked for other goroutine lookups.
		// When we decided to insert new node to the leaf, we do the same thing, except
		// the locking leaf for lookups and other writes.
		n, ok := leaf.GetAny(func() (key uint, ok bool) {
			cur, key, ok = path.NextKey(cur)
			return
		})
		if !ok {
			cur = path.Begin() // Reset cursor.

			var bottomLeaf *Leaf
			n = leaf.GetsertAny(
				func() (key uint, ok bool) {
					cur, key, ok = path.NextKey(cur)
					return
				},
				func() (n *Node) {
					n, bottomLeaf = c.makeTree(path, value, insert)
					n.parent = leaf
					return n
				},
			)
			if bottomLeaf != nil {
				return bottomLeaf, true
			}
		}
		v, ok := path.Get(n.key)
		if !ok {
			panic("inconsistent path state")
		}
		leaf = n.GetsertLeaf(v)
		path = path.Without(n.key)
	}

	var ok bool
	if insert {
		ok = leaf.Append(value)
	}
	return leaf, ok
}

// ForceInsert inserts value to the leaf that exists (or not and will be
// created) at the given path starting with the leaf as root.
//
// Note that path is inserted as is, without any optimizations.
func (c Inserter) ForceInsert(leaf *Leaf, pairs []Pair, value uint) {
	cb := c.IndexNode
	for _, pair := range pairs {
		n, inserted := leaf.GetsertChild(pair.Key)
		if inserted && cb != nil {
			cb(n)
		}
		leaf = n.GetsertLeaf(pair.Value)
	}
	leaf.Append(value)
}

func (c Inserter) makeTree(p Path, v uint, insert bool) (topNode *Node, bottomLeaf *Leaf) {
	cur, last, ok := p.Last()
	if !ok {
		panic("could not make tree with empty path")
	}
	cn := &Node{key: last.Key}
	cl := cn.GetsertLeaf(last.Value)
	if insert {
		cl.Append(v)
	}
	bottomLeaf = cl

	cb := c.IndexNode
	if cb != nil {
		cb(cn)
	}

	p.Descend(cur, func(p Pair) bool {
		n := &Node{key: p.Key}
		l := n.GetsertLeaf(p.Value)
		l.AddChild(cn)

		if cb != nil {
			cb(cn)
		}
		cn, cl = n, l
		return true
	})

	return cn, bottomLeaf
}

// Int implements the Item interface for integers.
type btreeUint uint

func (a btreeUint) Less(b btree.Item) bool {
	return a < b.(btreeUint)
}
