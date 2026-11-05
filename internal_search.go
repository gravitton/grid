package grid

import (
	"fmt"
	"math"
	"sync"
)

// orAny returns the callback, or one admitting every cell where it is nil.
func orAny[T any](admitted func(Cell[T]) bool) func(Cell[T]) bool {
	if admitted != nil {
		return admitted
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

// checkedCost returns the cost of the step, a cost with no meaning a panic: a negative one or a
// NaN would keep a search improving on what it found and never let it end.
func checkedCost[T any](cost func(current, next Cell[T]) float64, current, next Cell[T]) float64 {
	step := cost(current, next)
	if step < 0 || math.IsNaN(step) {
		panic(fmt.Errorf("%w %v from %s to %s", ErrInvalidCost, step, current, next))
	}

	return step
}

// A field keeps what a search found in square chunks of the grid, room made for a chunk when
// the search first reaches a cell in it, so a field holds as much as the search touched.
const (
	chunkShift = 4
	chunkSide  = 1 << chunkShift
	chunkMask  = chunkSide - 1
	chunkCells = chunkSide * chunkSide
)

// link is what a field holds for one slot: one more than the slot stepped from, zero where the
// search did not reach it and the slot itself at the start, and the steps from the start.
type link struct {
	from  int32
	steps int32
}

// node is a cell a search reached, its index with its slot in the field.
type node struct {
	index Index
	slot  int32
}

// scratch is what a search needs only while it runs, borrowed from scratches and handed back
// when it ends, so searches from several goroutines never share one.
type scratch struct {
	queue []node
	open  frontier
}

var scratches = sync.Pool{
	New: func() any {
		return new(scratch)
	},
}

// takeScratch returns a scratch no other search holds, its slices kept from its last search.
func takeScratch() *scratch {
	return scratches.Get().(*scratch)
}

// frontier returns the frontier of the scratch, emptied.
func (s *scratch) frontier() *frontier {
	s.open.items = s.open.items[:0]

	return &s.open
}

// release hands the scratch back for another search.
func (s *scratch) release() {
	scratches.Put(s)
}

// frontier is the cells a search has reached and not yet expanded, a binary heap that hands
// out the one of the least priority first, and of two equal ones the one of the lesser tie. It
// holds its items by value, and a cell reached again more cheaply is pushed again rather than
// moved.
type frontier struct {
	items []frontierItem
}

type frontierItem struct {
	slot     int32
	tie      int32
	priority float64
}

func (f *frontier) len() int {
	return len(f.items)
}

// push adds a slot with its priority and tie and sifts it up to its place.
func (f *frontier) push(slot int32, priority float64, tie int32) {
	f.items = append(f.items, frontierItem{slot, tie, priority})

	for i := len(f.items) - 1; i > 0; {
		parent := (i - 1) / 2
		if !f.before(i, parent) {
			break
		}

		f.items[parent], f.items[i] = f.items[i], f.items[parent]
		i = parent
	}
}

// pop removes the slot of the least priority, returning it with its priority, and sifts the
// last item down from the top.
func (f *frontier) pop() (int32, float64) {
	top := f.items[0]

	last := len(f.items) - 1
	f.items[0] = f.items[last]
	f.items = f.items[:last]

	for i := 0; ; {
		least := i
		for child := 2*i + 1; child <= 2*i+2 && child < last; child++ {
			if f.before(child, least) {
				least = child
			}
		}
		if least == i {
			break
		}

		f.items[i], f.items[least] = f.items[least], f.items[i]
		i = least
	}

	return top.slot, top.priority
}

// before reports whether the item at i is handed out before the one at j.
func (f *frontier) before(i, j int) bool {
	a, b := f.items[i], f.items[j]

	return a.priority < b.priority || (a.priority == b.priority && a.tie < b.tie)
}
