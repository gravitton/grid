package grid

// passableOrAny returns passable, or a function admitting every cell where it is nil.
func passableOrAny[T any](passable func(Cell[T]) bool) func(Cell[T]) bool {
	if passable != nil {
		return passable
	}

	return func(Cell[T]) bool {
		return true
	}
}

// costOrUnit returns cost, or a function charging one per step where it is nil.
func costOrUnit[T any](cost func(current, next Cell[T]) float64) func(current, next Cell[T]) float64 {
	if cost != nil {
		return cost
	}

	return func(_, _ Cell[T]) float64 {
		return 1
	}
}

// frontier is the cells a search has reached and not yet expanded, a binary heap that hands
// out the one of the least priority first. It holds its items by value, and a cell reached
// again more cheaply is pushed again rather than moved.
type frontier struct {
	items []frontierItem
}

type frontierItem struct {
	slot     int32
	priority float64
}

func (f *frontier) len() int {
	return len(f.items)
}

// push adds a slot with its priority and sifts it up to its place.
func (f *frontier) push(slot int32, priority float64) {
	f.items = append(f.items, frontierItem{slot, priority})

	for i := len(f.items) - 1; i > 0; {
		parent := (i - 1) / 2
		if f.items[parent].priority <= f.items[i].priority {
			break
		}

		f.items[parent], f.items[i] = f.items[i], f.items[parent]
		i = parent
	}
}

// pop removes the slot of the least priority and sifts the last item down from the top.
func (f *frontier) pop() int32 {
	top := f.items[0].slot

	last := len(f.items) - 1
	f.items[0] = f.items[last]
	f.items = f.items[:last]

	for i := 0; ; {
		least := i
		for child := 2*i + 1; child <= 2*i+2 && child < last; child++ {
			if f.items[child].priority < f.items[least].priority {
				least = child
			}
		}
		if least == i {
			break
		}

		f.items[i], f.items[least] = f.items[least], f.items[i]
		i = least
	}

	return top
}
