package grid

import (
	"iter"

	geom "github.com/gravitton/geometry"
)

// Cell is a handle on one cell of a [Grid]: its index, its value and its place in the world,
// with the short forms of the queries the grid answers from it. It is passed by value, and the
// cell of an index outside the grid is held like any other. The zero value belongs to no grid,
// and its methods are not to be called.
type Cell[T any] struct {
	grid  *Grid[T]
	index geom.Point[int]
}

// Index returns the index of the cell.
func (c Cell[T]) Index() geom.Point[int] {
	return c.index
}

// Size returns the size of the cell as drawn, the CellSize of its grid.
func (c Cell[T]) Size() geom.Size[float64] {
	return c.grid.CellSize()
}

// Center returns the center of the cell in the world.
func (c Cell[T]) Center() geom.Point[float64] {
	return c.grid.layout.toWorld(c.index)
}

// Bounds returns the box around the cell as drawn, in the world.
func (c Cell[T]) Bounds() geom.Box[float64] {
	return c.grid.layout.cellBounds(c.index)
}

// Polygon returns the polygon of the cell in the world: a square, a diamond or a hexagon.
func (c Cell[T]) Polygon() geom.RegularPolygon[float64] {
	polygon := c.grid.layout.cellPolygon
	polygon.Center = c.Center()

	return polygon
}

// Value returns a pointer into the grid to the value of the cell.
func (c Cell[T]) Value() *T {
	if !c.grid.Contains(c.index) {
		return nil
	}

	return c.grid.cells.Get(c.index)
}

// Set stores value in the cell.
func (c Cell[T]) Set(value T) {
	if !c.grid.Contains(c.index) {
		return
	}

	c.grid.cells.Set(c.index, value)
}

// Neighbors returns the indices of Grid.Neighbors from the cell.
func (c Cell[T]) Neighbors() []geom.Point[int] {
	return c.grid.Neighbors(c.index)
}

// AppendNeighbors appends the indices of Neighbors to dst and returns the extended slice.
func (c Cell[T]) AppendNeighbors(dst []geom.Point[int]) []geom.Point[int] {
	return c.grid.AppendNeighbors(dst, c.index)
}

// Range returns the region of Grid.Range from the cell.
func (c Cell[T]) Range(radius int, passable func(Cell[T]) bool) Region {
	return c.grid.Range(c.index, radius, passable)
}

// AppendRange appends the indices of Grid.AppendRange from the cell to dst and returns the
// extended slice.
func (c Cell[T]) AppendRange(dst []geom.Point[int], radius int, passable func(Cell[T]) bool) []geom.Point[int] {
	return c.grid.AppendRange(dst, c.index, radius, passable)
}

// RangeSeq returns an iterator over the indices of AppendRange.
func (c Cell[T]) RangeSeq(radius int, passable func(Cell[T]) bool) iter.Seq[geom.Point[int]] {
	return c.grid.RangeSeq(c.index, radius, passable)
}

// DistanceTo returns the number of steps of Grid.Distance from the cell to the index.
func (c Cell[T]) DistanceTo(to geom.Point[int]) int {
	return c.grid.Distance(c.index, to)
}

// PathTo returns the path of Grid.Path from the cell to the index.
func (c Cell[T]) PathTo(to geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return c.grid.Path(c.index, to, passable, cost)
}

// IsValid reports whether the cell is one of the grid, which Grid.Contains reports of its
// index.
func (c Cell[T]) IsValid() bool {
	return c.grid.Contains(c.index)
}

// String returns the cell as the call that gives it: Cell(x,y).
func (c Cell[T]) String() string {
	return "Cell" + c.index.String()
}
