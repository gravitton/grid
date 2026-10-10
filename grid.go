package grid

import (
	"iter"

	geom "github.com/gravitton/geometry"
)

// Grid holds a value of type T for every cell of a rectangular, isometric or hexagonal layout,
// maps its cells to the world, and answers the neighbor, range and path queries between them.
// An index is the column and row of a cell, on every layout. A grid is built by [NewRect],
// [NewIsometric], [NewHexFlatTop] or [NewHexPointyTop]; the zero value is no grid, and its
// methods are not to be called.
type Grid[T any] struct {
	cells    Array[T]
	layout   layout
	topology topology
}

// newGrid returns a grid of the given size in cells, placed by the layout, with the neighbor
// rule of its kind of cell.
func newGrid[T any](size geom.Size[int], layout layout, topology topology) *Grid[T] {
	return &Grid[T]{
		cells:    Arr[T](size),
		layout:   layout,
		topology: topology,
	}
}

// Size returns the number of columns and rows of the grid.
func (g *Grid[T]) Size() geom.Size[int] {
	return g.cells.Size()
}

// Width returns the number of columns of the grid.
func (g *Grid[T]) Width() int {
	return g.cells.Width()
}

// Height returns the number of rows of the grid.
func (g *Grid[T]) Height() int {
	return g.cells.Height()
}

// Cell returns the cell at index, a handle on its value and its place in the world. An index
// outside the grid gives a cell whose IsValid reports false.
func (g *Grid[T]) Cell(index geom.Point[int]) Cell[T] {
	return Cell[T]{g, index}
}

// Fill sets every cell to value.
func (g *Grid[T]) Fill(value T) {
	g.cells.Fill(value)
}

// Clear resets every cell to the zero value.
func (g *Grid[T]) Clear() {
	g.cells.Clear()
}

// Clone returns a copy of the grid with storage of its own: the same layout and the same
// values, copied as an assignment copies them, so a write to one grid leaves the other alone.
func (g *Grid[T]) Clone() *Grid[T] {
	clone := *g
	clone.cells = g.cells.Clone()

	return &clone
}

// All returns an iterator over every cell of the grid, row by row. For the order to draw them
// in, see [Grid.DrawOrder].
func (g *Grid[T]) All() iter.Seq[Cell[T]] {
	return func(yield func(Cell[T]) bool) {
		for index := range g.cells.Keys() {
			if !yield(g.Cell(index)) {
				return
			}
		}
	}
}

// Neighbors returns the indices adjacent to index under the movement of the grid that lie
// inside it, in the order of the neighbor rule. It returns nil for an index outside the grid.
func (g *Grid[T]) Neighbors(index geom.Point[int]) []geom.Point[int] {
	if !g.Contains(index) {
		return nil
	}

	return g.appendNeighbors(make([]geom.Point[int], 0, len(g.topology.offsets(index))), index)
}

// AppendNeighbors appends the indices Neighbors returns to dst and returns the extended slice.
func (g *Grid[T]) AppendNeighbors(dst []geom.Point[int], index geom.Point[int]) []geom.Point[int] {
	if !g.Contains(index) {
		return dst
	}

	return g.appendNeighbors(dst, index)
}

// appendNeighbors appends the neighbors of an index the caller knows to be inside the grid.
func (g *Grid[T]) appendNeighbors(dst []geom.Point[int], index geom.Point[int]) []geom.Point[int] {
	for _, offset := range g.topology.offsets(index) {
		next := index.Add(offset)
		if g.Contains(next) {
			dst = append(dst, next)
		}
	}

	return dst
}

// Range returns the cells within radius of index, under the movement of the grid, that are
// passable and in sight of it: what a unit there sees or a blast from there hits. A cell that
// is not passable is left out and hides what lies behind it. A nil passable admits every cell.
// It is empty for a negative radius, and for an index that is outside the grid or not passable
// itself: nothing is seen from where nothing can stand. A radius beyond the grid costs no more
// than one that reaches the cell farthest from index.
func (g *Grid[T]) Range(index geom.Point[int], radius int, passable func(Cell[T]) bool) Region {
	if radius < 0 || !g.passes(index, passable) {
		return Region{}
	}

	radius = g.clampRadius(index, radius)

	return Region{g.appendRange(make([]geom.Point[int], 0, g.topology.rangeLen(radius)), index, radius, passable)}
}

// AppendRange appends the indices of the cells Range returns to dst, in no order to rely on,
// and returns the extended slice, for a caller who wants them without a region.
func (g *Grid[T]) AppendRange(dst []geom.Point[int], index geom.Point[int], radius int, passable func(Cell[T]) bool) []geom.Point[int] {
	if radius < 0 || !g.passes(index, passable) {
		return dst
	}

	return g.appendRange(dst, index, g.clampRadius(index, radius), passable)
}

// appendRange appends the indices in range of an index the caller knows to pass. It asks
// passable about every cell within radius once and keeps the cells that pass; only where one
// was refused can anything be hidden, and only then is a sightline walked, to each cell kept,
// past the cells refused. A nil passable refuses nothing.
func (g *Grid[T]) appendRange(dst []geom.Point[int], index geom.Point[int], radius int, passable func(Cell[T]) bool) []geom.Point[int] {
	start := len(dst)
	dst = g.topology.appendWithin(dst, index, radius)

	passed := dst[:start]
	refused := refusals{corner: geom.Pt(index.X-radius-1, index.Y-radius-1), side: 2*radius + 3}
	for _, candidate := range dst[start:] {
		switch {
		case !g.Contains(candidate):
		case candidate == index || passable == nil || passable(g.Cell(candidate)):
			passed = append(passed, candidate)
		default:
			refused.add(candidate)
		}
	}

	if refused.isEmpty() {
		return passed
	}

	return g.appendSeen(passed[:start], index, passed[start:], refused)
}

// appendSeen appends the candidates in sight of center past the cells refused and the edge of
// the grid. The candidates may lie in dst past its length: a candidate is read before the place
// it is kept at is written.
func (g *Grid[T]) appendSeen(dst []geom.Point[int], center geom.Point[int], candidates []geom.Point[int], refused refusals) []geom.Point[int] {
	blocked := func(index geom.Point[int]) bool {
		return !g.Contains(index) || refused.contains(index)
	}

	for _, candidate := range candidates {
		if g.topology.sees(center, candidate, blocked) {
			dst = append(dst, candidate)
		}
	}

	return dst
}

// RangeSeq returns an iterator over the indices AppendRange appends, in its order.
func (g *Grid[T]) RangeSeq(index geom.Point[int], radius int, passable func(Cell[T]) bool) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		g.walkRange(index, radius, passable, yield)
	}
}

// walkRange yields the indices of the cells Range returns and reports whether it yielded them
// all. The index is checked here, inside the closure of RangeSeq, and not before it: RangeSeq
// then does nothing but build the closure and stays small enough to inline. The range is
// gathered before any of it is yielded, since a cell is known to be in sight only once every
// cell around it was asked about.
func (g *Grid[T]) walkRange(index geom.Point[int], radius int, passable func(Cell[T]) bool, yield func(geom.Point[int]) bool) bool {
	if radius < 0 || !g.passes(index, passable) {
		return true
	}

	radius = g.clampRadius(index, radius)

	for _, found := range g.appendRange(make([]geom.Point[int], 0, g.topology.rangeLen(radius)), index, radius, passable) {
		if !yield(found) {
			return false
		}
	}

	return true
}

// clampRadius returns the radius cut to the distance from the index to the farthest cell of
// the grid, one of its four corners: a range gathers every index within its radius before it
// asks which lie in the grid, so a larger one would only cost more. A radius that stops short
// of the nearest edge needs no cut, a step moving one column and one row at most.
func (g *Grid[T]) clampRadius(index geom.Point[int], radius int) int {
	x, y := index.XY()
	right, bottom := g.Width()-1, g.Height()-1
	if radius <= min(x, y, right-x, bottom-y) {
		return radius
	}

	return min(radius, max(
		g.topology.distance(index, geom.Pt(0, 0)),
		g.topology.distance(index, geom.Pt(right, 0)),
		g.topology.distance(index, geom.Pt(0, bottom)),
		g.topology.distance(index, geom.Pt(right, bottom)),
	))
}

// passes reports whether the index is inside the grid and its cell passable.
func (g *Grid[T]) passes(index geom.Point[int], passable func(Cell[T]) bool) bool {
	return g.Contains(index) && (passable == nil || passable(g.Cell(index)))
}

// Distance returns the number of steps between two indices under the movement of the grid,
// whether or not they lie inside it.
func (g *Grid[T]) Distance(from, to geom.Point[int]) int {
	return g.topology.distance(from, to)
}

// Contains reports whether the index is that of a cell of the grid.
func (g *Grid[T]) Contains(index geom.Point[int]) bool {
	return g.cells.Contains(index)
}

// topology is the neighbor rule of one kind of cell: the steps to the cells adjacent to an
// index, the number of steps between two, the indices within a number of steps, and whether
// one index is in sight of another.
type topology interface {
	// offsets returns the steps from the index to its neighbors, in the order of the directions
	// of the kind. The slice is the topology's own and is never written to.
	offsets(index geom.Point[int]) []geom.Vector[int]

	// distance returns the number of steps between two indices, whether or not they lie inside
	// a grid.
	distance(from, to geom.Point[int]) int

	// rangeLen returns the number of indices within radius of an index, what appendWithin
	// appends.
	rangeLen(radius int) int

	// appendWithin appends every index within radius of index, inside a grid or not, in the
	// order of a region, so a range kept in that order is a region with nothing to sort.
	appendWithin(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int]

	// sees reports whether to is in sight of from past the indices blocked reports, neither end
	// blocking.
	sees(from, to geom.Point[int], blocked func(index geom.Point[int]) bool) bool
}

// refusals is the set of the cells a range was refused: one mark for every index of the square
// around its center, from the corner, made at the first refusal, so a range that is refused
// nothing makes none.
type refusals struct {
	corner geom.Point[int]
	side   int
	marks  []bool
}

// add marks the cell at the index as refused.
func (r *refusals) add(index geom.Point[int]) {
	if r.marks == nil {
		r.marks = make([]bool, r.side*r.side)
	}

	r.marks[index.X-r.corner.X+(index.Y-r.corner.Y)*r.side] = true
}

// contains reports whether the cell at the index was refused. An index outside the square
// was not.
func (r *refusals) contains(index geom.Point[int]) bool {
	x, y := index.X-r.corner.X, index.Y-r.corner.Y

	return x >= 0 && x < r.side && y >= 0 && y < r.side && r.marks[x+y*r.side]
}

// isEmpty reports whether no cell was refused.
func (r *refusals) isEmpty() bool {
	return r.marks == nil
}
