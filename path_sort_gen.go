package radix

func pairPartition(data []Pair, l, r int) int {
	// Inlined partition algorithm
	var j int
	{
		// Let x be a pivot
		x := data[l]
		j = l
		for i := l + 1; i < r; i++ {
			if data[i].Key <= x.Key {
				j++
				data[j], data[i] = data[i], data[j]
			}
		}
		data[j], data[l] = data[l], data[j]
	}
	return j
}

func pairQuickSort(data []Pair, lo, hi int) {
	if lo >= hi {
		return
	}
	// Inlined partition algorithm
	var p int
	{
		// Let x be a pivot
		x := data[lo]
		p = lo
		for i := lo + 1; i < hi; i++ {
			if data[i].Key <= x.Key {
				p++
				data[p], data[i] = data[i], data[p]
			}
		}
		data[p], data[lo] = data[lo], data[p]
	}
	pairQuickSort(data, lo, p)
	pairQuickSort(data, p+1, hi)
}

func pairInsertionSort(data []Pair, l, r int) {
	// Inlined insertion sort
	for i := l + 1; i < r; i++ {
		for j := i; j > l && data[j-1].Key > data[j].Key; j-- {
			data[j], data[j-1] = data[j-1], data[j]
		}
	}
}

func pairSort(data []Pair, l, r int) {
	if r-l > 12 {
		pairQuickSort(data, l, r)
		return
	}
	// Inlined insertion sort
	for i := l + 1; i < r; i++ {
		for j := i; j > l && data[j-1].Key > data[j].Key; j-- {
			data[j], data[j-1] = data[j-1], data[j]
		}
	}
}

func pairSearch(data []Pair, key uint) (int, bool) {
	// Inlined binary search
	var ok bool
	i := len(data)
	{
		l := 0
		for !ok && l < i {
			m := l + (i-l)/2
			switch {
			case data[m].Key == key:
				ok = true
				i = m
			case data[m].Key < key:
				l = m + 1
			case data[m].Key > key:
				i = m
			}
		}
	}
	return i, ok
}
