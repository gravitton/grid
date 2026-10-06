package grid

import (
	"strconv"

	geom "github.com/gravitton/geometry"
)

// Cell is a handle on one cell of a [Grid]: its index, its value and its place in the world,
// with the short forms of the queries the grid answers from it. It is passed by value. The cell
// of an index outside the grid can be held like any other: its IsValid reports false, its Value
// is nil and its Set stores nothing.
type Cell[T any] struct {
	grid  *Grid[T]
	index geom.Point[int]
}

// Index returns the index of the cell, its column and row in the grid.
func (c Cell[T]) Index() geom.Point[int] {
	return c.index
}

// Size returns the size of the cell as drawn, the same for every cell, as [Grid.CellSize]
// does.
func (c Cell[T]) Size() geom.Size[float64] {
	return c.grid.CellSize()
}

// Center returns the center of the cell in world coordinates.
func (c Cell[T]) Center() geom.Point[float64] {
	return c.grid.layout.ToPoint(c.index)
}

// Bounds returns the box around the cell as drawn, in world coordinates.
func (c Cell[T]) Bounds() geom.Box[float64] {
	return geom.Rect(c.Center(), c.Size()).Bounds()
}

// Polygon returns the shape of the cell in world coordinates: a square, a diamond or a
// hexagon.
func (c Cell[T]) Polygon() geom.RegularPolygon[float64] {
	return c.grid.layout.CellPolygon(c.index)
}

// Value returns a pointer to the value of the cell, into the grid, or nil for a cell outside
// it.
func (c Cell[T]) Value() *T {
	if !c.grid.Contains(c.index) {
		return nil
	}

	return c.grid.cells.at(c.index)
}

// Set stores value in the cell. A cell outside the grid stores nothing.
func (c Cell[T]) Set(value T) {
	if !c.grid.Contains(c.index) {
		return
	}

	*c.grid.cells.at(c.index) = value
}

// Neighbors returns the indices adjacent to the cell that lie inside the grid, as
// [Grid.Neighbors] does.
func (c Cell[T]) Neighbors() []geom.Point[int] {
	return c.grid.Neighbors(c.index)
}

// AppendNeighbors appends the indices Neighbors returns to dst and returns the extended slice.
func (c Cell[T]) AppendNeighbors(dst []geom.Point[int]) []geom.Point[int] {
	return c.grid.AppendNeighbors(dst, c.index)
}

// Range returns the cells within radius of the cell that are passable and in sight of it, as
// [Grid.Range] does.
func (c Cell[T]) Range(radius int, passable func(Cell[T]) bool) Region {
	return c.grid.Range(c.index, radius, passable)
}

// DistanceTo returns the number of steps from the cell to the given index, as [Grid.Distance]
// does.
func (c Cell[T]) DistanceTo(to geom.Point[int]) int {
	return c.grid.Distance(c.index, to)
}

// PathTo returns the cheapest path from the cell to the given index, both included, or nil
// where there is none, as [Grid.Path] does.
func (c Cell[T]) PathTo(to geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return c.grid.Path(c.index, to, passable, cost)
}

// IsValid reports whether the cell is one of the grid, as [Grid.Contains] does for its index.
func (c Cell[T]) IsValid() bool {
	return c.grid.Contains(c.index)
}

// String returns the cell as the call that gives it, Cell and its index.
func (c Cell[T]) String() string {
	buf := make([]byte, 0, stringSize)
	buf = append(buf, "Cell("...)
	buf = strconv.AppendInt(buf, int64(c.index.X), 10)
	buf = append(buf, ',')
	buf = strconv.AppendInt(buf, int64(c.index.Y), 10)
	buf = append(buf, ')')

	return string(buf)
}
