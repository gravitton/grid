package grid

import (
	"iter"

	geom "github.com/gravitton/geometry"
)

// Grid is a 2D grid.
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

// Size returns the size of the grid.
func (g *Grid[T]) Size() geom.Size[int] {
	return g.cells.Size()
}

// Width returns the width of the grid.
func (g *Grid[T]) Width() int {
	return g.cells.Width()
}

// Height returns the height of the grid.
func (g *Grid[T]) Height() int {
	return g.cells.Height()
}

// Layout returns the mapping between the indices of the grid and the world, for a caller who
// places or measures cells without the grid. It is a copy: moving or resizing it leaves the
// grid where it is.
func (g *Grid[T]) Layout() Layout {
	return g.layout
}

// Bounds returns the bounds of the grid.
func (g *Grid[T]) Bounds() geom.Box[float64] {
	return g.bounds
}

// CellSize returns the size of a cell as drawn: the width and height of the box around it,
// the same for every cell and what [Cell.Size] returns.
func (g *Grid[T]) CellSize() geom.Size[float64] {
	return g.layout.CellSize()
}

// CellSpacing returns the spacing between cell centers.
func (g *Grid[T]) CellSpacing() geom.Size[float64] {
	return g.layout.CellSpacing()
}

// CellWidth returns the width of a cell.
func (g *Grid[T]) CellWidth() float64 {
	return g.layout.CellSize().Width
}

// CellHeight returns the height of a cell.
func (g *Grid[T]) CellHeight() float64 {
	return g.layout.CellSize().Height
}

// Cell returns a cell.
func (g *Grid[T]) Cell(index geom.Point[int]) Cell[T] {
	return g.cell(index)
}

// IndexAt returns the index of the cell at a given point.
func (g *Grid[T]) IndexAt(point geom.Point[float64]) geom.Point[int] {
	return g.layout.FromPoint(point)
}

// CellAt returns the cell at a given point.
func (g *Grid[T]) CellAt(point geom.Point[float64]) Cell[T] {
	return g.cell(g.layout.FromPoint(point))
}

// Set stores value at the given index. Out-of-bounds writes are no-ops.
func (g *Grid[T]) Set(index geom.Point[int], value T) {
	g.setData(index, value)
}

// Fill sets every cell to value.
func (g *Grid[T]) Fill(value T) {
	g.cells.Fill(value)
}

// Clear resets every cell to the zero value.
func (g *Grid[T]) Clear() {
	g.cells.Clear()
}

// Clone returns a deep copy of the grid with the same layout and cell values.
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
			if !yield(g.cell(index)) {
				return
			}
		}
	}
}

func (g *Grid[T]) cell(index geom.Point[int]) Cell[T] {
	return Cell[T]{g, index}
}

func (g *Grid[T]) getData(index geom.Point[int]) *T {
	if !g.cells.valid(index) {
		return nil
	}
	return g.cells.at(index)
}

func (g *Grid[T]) setData(index geom.Point[int], value T) {
	if !g.cells.valid(index) {
		return
	}
	*g.cells.at(index) = value
}

func (g *Grid[T]) cellCenter(index geom.Point[int]) geom.Point[float64] {
	return g.layout.ToPoint(index)
}

func (g *Grid[T]) cellBounds(index geom.Point[int]) geom.Box[float64] {
	return geom.Rect(g.cellCenter(index), g.layout.CellSize()).Bounds()
}

func (g *Grid[T]) cellPolygon(index geom.Point[int]) geom.RegularPolygon[float64] {
	return g.layout.CellPolygon(index)
}

// Range returns the indices within n steps of index, under the movement of the grid, that are
// passable and in sight of it. A nil passable admits every cell.
func (g *Grid[T]) Range(index geom.Point[int], n int, passable func(Cell[T]) bool) []geom.Point[int] {
	passable = passableOrAny(passable)

	return g.topology.rangeOf(index, n, func(index geom.Point[int]) bool {
		return g.valid(index) && passable(g.cell(index))
	})
}

func (g *Grid[T]) cellNeighbors(index geom.Point[int]) []geom.Point[int] {
	offsets := g.topology.offsets(index)

	neighbors := make([]geom.Point[int], 0, len(offsets))
	for _, offset := range offsets {
		next := index.Add(offset)
		if g.valid(next) {
			neighbors = append(neighbors, next)
		}
	}

	return neighbors
}

// Distance returns the distance between two points.
func (g *Grid[T]) Distance(from, to geom.Point[int]) int {
	return g.topology.distance(from, to)
}

// Contains reports whether the index is within the grid bounds.
func (g *Grid[T]) Contains(index geom.Point[int]) bool {
	return g.valid(index)
}

func (g *Grid[T]) valid(index geom.Point[int]) bool {
	return g.cells.valid(index)
}

// topology is the neighbor rule of one kind of cell: the steps to the cells adjacent to an
// index, the number of steps between two, and the indices within a number of steps that are
// passable and in sight.
type topology interface {
	// offsets returns the steps from the index to its neighbors, in the order of the directions
	// of the kind. The slice is the topology's own and is never written to.
	offsets(index geom.Point[int]) []geom.Vector[int]
	distance(from, to geom.Point[int]) int
	rangeOf(index geom.Point[int], n int, passable func(index geom.Point[int]) bool) []geom.Point[int]
}
