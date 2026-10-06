package grid

import (
	"iter"

	geom "github.com/gravitton/geometry"
)

// Grid holds a value of type T for every cell of a rectangular, isometric or hexagonal layout,
// maps its cells to the world, and answers the neighbor, range and path queries between them.
// An index is the column and row of a cell, on every layout. A grid is built by [NewRect],
// [NewIsometric], [NewHexFlatTop] or [NewHexPointyTop].
type Grid[T any] struct {
	cells    Array[T]
	layout   Layout
	topology topology
	bounds   geom.Box[float64]
}

// newGrid returns a grid stored for the size of the layout, with the neighbor rule of its kind
// of cell.
func newGrid[T any](layout Layout, topology topology) *Grid[T] {
	return &Grid[T]{
		cells:    Arr[T](layout.Size()),
		layout:   layout,
		topology: topology,
		bounds:   layout.Bounds(),
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

// Layout returns the mapping between the indices of the grid and the world, for a caller who
// places or measures cells without the grid. It is a copy: moving or resizing it leaves the
// grid where it is.
func (g *Grid[T]) Layout() Layout {
	return g.layout
}

// Bounds returns the box around every cell of the grid, in world coordinates.
func (g *Grid[T]) Bounds() geom.Box[float64] {
	return g.bounds
}

// CellSize returns the size of a cell as drawn: the width and height of the box around it,
// the same for every cell and what [Cell.Size] returns.
func (g *Grid[T]) CellSize() geom.Size[float64] {
	return g.layout.CellSize()
}

// CellSpacing returns the distance between the centers of adjacent cells along each axis.
func (g *Grid[T]) CellSpacing() geom.Size[float64] {
	return g.layout.CellSpacing()
}

// CellWidth returns the width of a cell as drawn, the width of CellSize.
func (g *Grid[T]) CellWidth() float64 {
	return g.CellSize().Width
}

// CellHeight returns the height of a cell as drawn, the height of CellSize.
func (g *Grid[T]) CellHeight() float64 {
	return g.CellSize().Height
}

// Cell returns the cell at index, a handle on its value and its place in the world. An index
// outside the grid gives a cell whose IsValid reports false.
func (g *Grid[T]) Cell(index geom.Point[int]) Cell[T] {
	return Cell[T]{g, index}
}

// IndexAt returns the index of the cell the point of the world lies in. A point outside the
// grid gives an index outside it, where Contains reports false.
func (g *Grid[T]) IndexAt(point geom.Point[float64]) geom.Point[int] {
	return g.layout.FromPoint(point)
}

// CellAt returns the cell the point of the world lies in, the cell of IndexAt.
func (g *Grid[T]) CellAt(point geom.Point[float64]) Cell[T] {
	return g.Cell(g.IndexAt(point))
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
// itself: nothing is seen from where nothing can stand.
func (g *Grid[T]) Range(index geom.Point[int], radius int, passable func(Cell[T]) bool) Region {
	if radius < 0 || !g.passes(index, passable) {
		return Region{}
	}

	return setOf(g.appendRange(make([]geom.Point[int], 0, g.topology.rangeLen(radius)), index, radius, passable))
}

// AppendRange appends the indices of the cells Range returns to dst, in no order to rely on,
// and returns the extended slice, for a caller who wants them without a region.
func (g *Grid[T]) AppendRange(dst []geom.Point[int], index geom.Point[int], radius int, passable func(Cell[T]) bool) []geom.Point[int] {
	if !g.passes(index, passable) {
		return dst
	}

	return g.appendRange(dst, index, radius, passable)
}

// appendRange appends the indices in range of an index the caller knows to pass.
func (g *Grid[T]) appendRange(dst []geom.Point[int], index geom.Point[int], radius int, passable func(Cell[T]) bool) []geom.Point[int] {
	g.walkRangeFrom(index, radius, passable, func(candidate geom.Point[int]) bool {
		dst = append(dst, candidate)

		return true
	})

	return dst
}

// RangeSeq returns an iterator over the indices AppendRange appends, in its order.
func (g *Grid[T]) RangeSeq(index geom.Point[int], radius int, passable func(Cell[T]) bool) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		g.walkRange(index, radius, passable, yield)
	}
}

// walkRange yields the indices of the cells Range returns and reports whether it yielded them
// all. It checks the index and then calls walkRangeFrom. The check is made here, inside the
// closure of RangeSeq, and not before it: RangeSeq then does nothing but build the closure and
// stays small enough to inline.
func (g *Grid[T]) walkRange(index geom.Point[int], radius int, passable func(Cell[T]) bool, yield func(geom.Point[int]) bool) bool {
	if !g.passes(index, passable) {
		return true
	}

	return g.walkRangeFrom(index, radius, passable, yield)
}

// walkRangeFrom yields the indices in range of an index the caller knows to pass, which is not
// asked about again, and reports whether it yielded them all.
func (g *Grid[T]) walkRangeFrom(index geom.Point[int], radius int, passable func(Cell[T]) bool, yield func(geom.Point[int]) bool) bool {
	return g.topology.walkRange(index, radius, func(candidate geom.Point[int]) bool {
		return candidate == index || g.passes(candidate, passable)
	}, yield)
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
// index, the number of steps between two, and the indices within a number of steps that are
// passable and in sight.
type topology interface {
	// offsets returns the steps from the index to its neighbors, in the order of the directions
	// of the kind. The slice is the topology's own and is never written to.
	offsets(index geom.Point[int]) []geom.Vector[int]

	// distance returns the number of steps between two indices, whether or not they lie inside
	// a grid.
	distance(from, to geom.Point[int]) int

	// rangeLen returns the number of indices within radius of an index, the most walkRange
	// yields.
	rangeLen(radius int) int

	// walkRange yields the indices within radius of index that are passable and in sight of
	// it, in no order the kinds share, and reports whether it yielded them all.
	walkRange(index geom.Point[int], radius int, passable func(index geom.Point[int]) bool, yield func(geom.Point[int]) bool) bool
}
