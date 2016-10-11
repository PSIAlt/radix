package radix

// THIS FILE WAS AUTOGENERATED.
// DO NOT EDIT.

type nodeArray struct {
	data []*node
}

func (a nodeArray) Has(x uint) bool {
	// Inlined binary search
	var ok bool
	var i int
	{
		l := 0
		r := len(a.data)
		for !ok && l < r {
			m := l + (r-l)/2
			switch {
			case a.data[m].key == x:
				ok = true
				r = m
			case a.data[m].key < x:
				l = m + 1
			case a.data[m].key > x:
				r = m
			}
		}
		i = r
		_ = i // in case when i not being used
	}
	return ok
}

func (a nodeArray) Get(x uint) *node {
	// Inlined binary search
	var ok bool
	var i int
	{
		l := 0
		r := len(a.data)
		for !ok && l < r {
			m := l + (r-l)/2
			switch {
			case a.data[m].key == x:
				ok = true
				r = m
			case a.data[m].key < x:
				l = m + 1
			case a.data[m].key > x:
				r = m
			}
		}
		i = r
		_ = i // in case when i not being used
	}
	if !ok {
		return nil
	}
	return a.data[i]
}

func (a nodeArray) Upsert(x *node) (cp nodeArray, prev *node) {
	var with []*node
	// Inlined binary search
	var has bool
	var i int
	{
		l := 0
		r := len(a.data)
		for !has && l < r {
			m := l + (r-l)/2
			switch {
			case a.data[m].key == x.key:
				has = true
				r = m
			case a.data[m].key < x.key:
				l = m + 1
			case a.data[m].key > x.key:
				r = m
			}
		}
		i = r
		_ = i // in case when i not being used
	}
	if has {
		with = make([]*node, len(a.data))
		copy(with, a.data)
		a.data[i], prev = x, a.data[i]
	} else {
		with = make([]*node, len(a.data)+1)
		copy(with[:i], a.data[:i])
		copy(with[i+1:], a.data[i:])
		with[i] = x
	}
	return nodeArray{with}, prev
}

func (a nodeArray) Delete(x uint) (cp nodeArray, prev *node) {
	// Inlined binary search
	var has bool
	var i int
	{
		l := 0
		r := len(a.data)
		for !has && l < r {
			m := l + (r-l)/2
			switch {
			case a.data[m].key == x:
				has = true
				r = m
			case a.data[m].key < x:
				l = m + 1
			case a.data[m].key > x:
				r = m
			}
		}
		i = r
		_ = i // in case when i not being used
	}
	if !has {
		return a, nil
	}
	without := make([]*node, len(a.data)-1)
	copy(without[:i], a.data[:i])
	copy(without[i:], a.data[i+1:])
	return nodeArray{without}, a.data[i]
}

func (a nodeArray) Ascend(cb func(x *node) bool) bool {
	for _, x := range a.data {
		if !cb(x) {
			return false
		}
	}
	return true
}

func (a nodeArray) AscendRange(x, y uint, cb func(x *node) bool) bool {
	// Inlined binary search
	var ok0 bool
	var i int
	{
		l := 0
		r := len(a.data)
		for !ok0 && l < r {
			m := l + (r-l)/2
			switch {
			case a.data[m].key == x:
				ok0 = true
				r = m
			case a.data[m].key < x:
				l = m + 1
			case a.data[m].key > x:
				r = m
			}
		}
		i = r
		_ = i // in case when i not being used
	}
	// Inlined binary search
	var ok1 bool
	var j int
	{
		l := 0
		r := len(a.data)
		for !ok1 && l < r {
			m := l + (r-l)/2
			switch {
			case a.data[m].key == y:
				ok1 = true
				r = m
			case a.data[m].key < y:
				l = m + 1
			case a.data[m].key > y:
				r = m
			}
		}
		j = r
		_ = j // in case when j not being used
	}
	for ; i < len(a.data) && i <= j; i++ {
		if !cb(a.data[i]) {
			return false
		}
	}
	return true
}

func (a nodeArray) Len() int {
	return len(a.data)
}
