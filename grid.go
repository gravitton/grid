package grid

import (
	"iter"

	geom "github.com/gravitton/geometry"
)

// Grid is a 2D grid.
type Grid[T any] struct {
	cells       Array[T]
	layout      Layout
	cellSize    geom.Size[float64]
	cellSpacing geom.Size[float64]
	bounds      geom.Box[float64]
	kind        kind

	toPoint   func(index geom.Point[int]) geom.Point[float64]
	fromPoint func(point geom.Point[float64]) geom.Point[int]
	polygon   func(index geom.Point[int]) geom.RegularPolygon[float64]
	distance  func(from, to geom.Point[int]) int
	toRange   func(index geom.Point[int], n int, passable func(index geom.Point[int]) bool) []geom.Point[int]
	neighbors func(index geom.Point[int]) []geom.Vector[int]
}

// newGrid returns a grid stored for the size of the layout, with the distance, range and
// neighbors of its kind of cell.
func newGrid[T any](
	layout Layout,
	distance func(from, to geom.Point[int]) int,
	toRange func(index geom.Point[int], n int, passable func(index geom.Point[int]) bool) []geom.Point[int],
	neighbors func(index geom.Point[int]) []geom.Vector[int],
) *Grid[T] {
	return &Grid[T]{
		cells:       Arr[T](layout.Size()),
		layout:      layout,
		cellSize:    layout.CellSize(),
		cellSpacing: layout.CellSpacing(),
		bounds:      layout.Bounds(),
		kind:        layout.kind(),
		toPoint:     layout.ToPoint,
		fromPoint:   layout.FromPoint,
		polygon:     layout.CellPolygon,
		distance:    distance,
		toRange:     toRange,
		neighbors:   neighbors,
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
	return g.cellSize
}

// CellSpacing returns the spacing between cell centers.
func (g *Grid[T]) CellSpacing() geom.Size[float64] {
	return g.cellSpacing
}

// CellWidth returns the width of a cell.
func (g *Grid[T]) CellWidth() float64 {
	return g.cellSize.Width
}

// CellHeight returns the height of a cell.
func (g *Grid[T]) CellHeight() float64 {
	return g.cellSize.Height
}

// Cell returns a cell.
func (g *Grid[T]) Cell(index geom.Point[int]) *Cell[T] {
	return g.cell(index)
}

// IndexAt returns the index of the cell at a given point.
func (g *Grid[T]) IndexAt(point geom.Point[float64]) geom.Point[int] {
	return g.fromPoint(point)
}

// CellAt returns the cell at a given point.
func (g *Grid[T]) CellAt(point geom.Point[float64]) *Cell[T] {
	return g.cell(g.fromPoint(point))
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
func (g *Grid[T]) All() iter.Seq[*Cell[T]] {
	return func(yield func(*Cell[T]) bool) {
		for index := range g.cells.Keys() {
			if !yield(g.cell(index)) {
				return
			}
		}
	}
}

func (g *Grid[T]) cell(index geom.Point[int]) *Cell[T] {
	return &Cell[T]{g, index}
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
	return g.toPoint(index)
}

func (g *Grid[T]) cellBounds(index geom.Point[int]) geom.Box[float64] {
	return geom.Rect(g.cellCenter(index), g.cellSize).Bounds()
}

func (g *Grid[T]) cellPolygon(index geom.Point[int]) geom.RegularPolygon[float64] {
	return g.polygon(index)
}

// Range returns the indices within n steps of index, under the movement of the grid, that are
// passable and in sight of it. A nil passable admits every cell.
func (g *Grid[T]) Range(index geom.Point[int], n int, passable func(*Cell[T]) bool) []geom.Point[int] {
	passable = passableOrAny(passable)

	return g.toRange(index, n, func(index geom.Point[int]) bool {
		return g.valid(index) && passable(g.cell(index))
	})
}

func (g *Grid[T]) cellNeighbors(index geom.Point[int]) []geom.Point[int] {
	offsets := g.neighbors(index)

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
	return g.distance(from, to)
}

// Contains reports whether the index is within the grid bounds.
func (g *Grid[T]) Contains(index geom.Point[int]) bool {
	return g.valid(index)
}

func (g *Grid[T]) valid(index geom.Point[int]) bool {
	return g.cells.valid(index)
}
