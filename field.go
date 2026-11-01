package grid

import (
	"iter"
	"math"
	"slices"

	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

// Field is what a search from one index found: every index it reached, with the step back
// toward the start and the cost of getting there. A field searched from a destination
// serves every unit heading there, its path running from the destination to the unit, where a
// step costs the same both ways. It does not change with the grid. The zero value is the empty
// field.
type Field struct {
	size  geom.Size[int]
	start geom.Point[int]

	// from holds, for the slot of every index, one more than the slot stepped from, so
	// zero marks an index the search did not reach and the start holds its own slot.
	from []int32

	// costs holds the cost of reaching every index, and is nil for a search that counts steps.
	costs []float64

	reached int
}

// newField returns a field over a grid of the given size, holding the start alone, with room
// for costs where the search weighs its steps.
func newField(size geom.Size[int], start geom.Point[int], weighted bool) Field {
	f := Field{
		size:    size,
		start:   start,
		from:    make([]int32, size.Width*size.Height),
		reached: 1,
	}
	if weighted {
		f.costs = make([]float64, len(f.from))
	}

	f.from[f.slot(start)] = f.slot(start) + 1

	return f
}

// Start returns the index the search started from, and the zero index for the empty field.
func (f Field) Start() geom.Point[int] {
	return f.start
}

// Len returns the number of indices the search reached, the start included.
func (f Field) Len() int {
	return f.reached
}

// Next returns the index one step closer to the start, and false for the start itself and for
// an index the search did not reach.
func (f Field) Next(index geom.Point[int]) (geom.Point[int], bool) {
	if !f.Contains(index) || index == f.start {
		return geom.Point[int]{}, false
	}

	return f.index(f.from[f.slot(index)] - 1), true
}

// CostTo returns what the search paid to reach the index, the number of steps where it did not
// weigh them, and false for an index it did not reach.
func (f Field) CostTo(index geom.Point[int]) (float64, bool) {
	if !f.Contains(index) {
		return 0, false
	}
	if f.costs != nil {
		return f.costs[f.slot(index)], true
	}

	return float64(f.steps(index)), true
}

// PathTo returns the path the search found from the start to the index, both included, and nil
// for an index it did not reach.
func (f Field) PathTo(index geom.Point[int]) []geom.Point[int] {
	if !f.Contains(index) {
		return nil
	}

	return f.AppendPathTo(make([]geom.Point[int], 0, f.steps(index)+1), index)
}

// AppendPathTo appends the indices of PathTo to dst and returns the extended slice.
func (f Field) AppendPathTo(dst []geom.Point[int], index geom.Point[int]) []geom.Point[int] {
	if !f.Contains(index) {
		return dst
	}

	start, length := len(dst), f.steps(index)+1
	dst = slices.Grow(dst, length)[:start+length]
	for i := len(dst) - 1; i > start; i-- {
		dst[i] = index
		index = f.index(f.from[f.slot(index)] - 1)
	}
	dst[start] = index

	return dst
}

// steps returns the number of steps from the start to an index the search reached.
func (f Field) steps(index geom.Point[int]) int {
	steps := 0
	for slot := f.slot(index); slot != f.slot(f.start); slot = f.from[slot] - 1 {
		steps++
	}

	return steps
}

// Contains reports whether the search reached the index.
func (f Field) Contains(index geom.Point[int]) bool {
	if index.X < 0 || index.X >= f.size.Width || index.Y < 0 || index.Y >= f.size.Height {
		return false
	}

	return f.from[f.slot(index)] != 0
}

// IsEmpty reports whether the search reached nothing, not even its start.
func (f Field) IsEmpty() bool {
	return f.reached == 0
}

// All returns an iterator over the indices the search reached, in the order of a region.
func (f Field) All() iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		if f.reached == 0 {
			return
		}

		for x := range f.size.Width {
			for y := range f.size.Height {
				if index := geom.Pt(x, y); f.from[f.slot(index)] != 0 && !yield(index) {
					return
				}
			}
		}
	}
}

// Region returns the indices the search reached.
func (f Field) Region() Region {
	if f.reached == 0 {
		return Region{}
	}

	return container.SetFrom(slices.AppendSeq(make([]geom.Point[int], 0, f.reached), f.All()))
}

// improve records that index is reached from the slot at the given cost, where the cost is
// finite and the search had not reached it or had paid more, and reports whether it did.
func (f *Field) improve(index geom.Point[int], from int32, cost float64) bool {
	slot := f.slot(index)

	switch {
	case math.IsInf(cost, 1):
		return false
	case f.from[slot] == 0:
		f.reached++
	case cost >= f.costs[slot]:
		return false
	}

	f.from[slot] = from + 1
	f.costs[slot] = cost

	return true
}

// slot returns the place of an index inside the grid in the slices of the field.
func (f Field) slot(index geom.Point[int]) int32 {
	return int32(index.X + index.Y*f.size.Width)
}

// index returns the index at a slot, the inverse of slot.
func (f Field) index(slot int32) geom.Point[int] {
	return geom.Pt(int(slot)%f.size.Width, int(slot)/f.size.Width)
}
