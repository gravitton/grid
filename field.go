package grid

import (
	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

// Field is what a search from one index found: every index it reached, with the step back
// toward the start and the cost of getting there, so one field serves every unit heading to
// the same place. It does not change with the grid. The zero value is the empty field.
type Field struct {
	size  geom.Size[int]
	start geom.Point[int]

	// from holds, for the position of every index, one more than the position stepped from, so
	// zero marks an index the search did not reach and the start holds its own position.
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

	f.from[f.position(start)] = f.position(start) + 1

	return f
}

// improve records that index is reached from the position at the given cost, where the search
// had not reached it or had paid more, and reports whether it did.
func (f *Field) improve(index geom.Point[int], from int32, cost float64) bool {
	position := f.position(index)

	switch {
	case f.from[position] == 0:
		f.reached++
	case cost >= f.costs[position]:
		return false
	}

	f.from[position] = from + 1
	f.costs[position] = cost

	return true
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

	return f.index(f.from[f.position(index)] - 1), true
}

// CostTo returns what the search paid to reach the index, the number of steps where it did not
// weigh them, and false for an index it did not reach.
func (f Field) CostTo(index geom.Point[int]) (float64, bool) {
	if !f.Contains(index) {
		return 0, false
	}
	if f.costs != nil {
		return f.costs[f.position(index)], true
	}

	return float64(f.steps(index)), true
}

// PathTo returns the path the search found from the start to the index, both included, and nil
// for an index it did not reach.
func (f Field) PathTo(index geom.Point[int]) []geom.Point[int] {
	if !f.Contains(index) {
		return nil
	}

	path := make([]geom.Point[int], f.steps(index)+1)
	for i := len(path) - 1; i > 0; i-- {
		path[i] = index
		index = f.index(f.from[f.position(index)] - 1)
	}
	path[0] = index

	return path
}

// steps returns the number of steps from the start to an index the search reached.
func (f Field) steps(index geom.Point[int]) int {
	steps := 0
	for position := f.position(index); position != f.position(f.start); position = f.from[position] - 1 {
		steps++
	}

	return steps
}

// position returns the place of an index inside the grid in the slices of the field.
func (f Field) position(index geom.Point[int]) int32 {
	return int32(index.X + index.Y*f.size.Width)
}

// index returns the index at a position, the inverse of position.
func (f Field) index(position int32) geom.Point[int] {
	return geom.Pt(int(position)%f.size.Width, int(position)/f.size.Width)
}

// Contains reports whether the search reached the index.
func (f Field) Contains(index geom.Point[int]) bool {
	if index.X < 0 || index.X >= f.size.Width || index.Y < 0 || index.Y >= f.size.Height {
		return false
	}

	return f.from[f.position(index)] != 0
}

// IsEmpty reports whether the search reached nothing, not even its start.
func (f Field) IsEmpty() bool {
	return f.reached == 0
}

// Region returns the indices the search reached.
func (f Field) Region() Region {
	if f.reached == 0 {
		return Region{}
	}

	points := make([]geom.Point[int], 0, f.reached)
	for x := range f.size.Width {
		for y := range f.size.Height {
			if index := geom.Pt(x, y); f.from[f.position(index)] != 0 {
				points = append(points, index)
			}
		}
	}

	return container.SetFrom(points)
}
