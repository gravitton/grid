package grid

import (
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
	return c.grid.layout.cellSize
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

	return &c.grid.cells.Data()[c.grid.slot(c.index)]
}

// Set stores value in the cell.
func (c Cell[T]) Set(value T) {
	if target := c.Value(); target != nil {
		*target = value
	}
}

// Neighbors returns the indices of Grid.Neighbors from the cell.
func (c Cell[T]) Neighbors() []geom.Point[int] {
	return c.grid.Neighbors(c.index)
}

// Range returns the region of Grid.Range from the cell.
func (c Cell[T]) Range(radius int) Region {
	return c.grid.Range(c.index, radius)
}

// Ring returns the indices of Grid.Ring from the cell.
func (c Cell[T]) Ring(radius int) []geom.Point[int] {
	return c.grid.Ring(c.index, radius)
}

// Spiral returns the indices of Grid.Spiral from the cell.
func (c Cell[T]) Spiral(radius int) []geom.Point[int] {
	return c.grid.Spiral(c.index, radius)
}

// DistanceTo returns the number of steps of Grid.Distance from the cell to the index.
func (c Cell[T]) DistanceTo(index geom.Point[int]) int {
	return c.grid.Distance(c.index, index)
}

// LineTo returns the indices of Grid.Line from the cell to the index.
func (c Cell[T]) LineTo(index geom.Point[int]) []geom.Point[int] {
	return c.grid.Line(c.index, index)
}

// HasLineOfSight reports whether the cell sees target by Grid.HasLineOfSight.
func (c Cell[T]) HasLineOfSight(target geom.Point[int], blocked func(Cell[T]) bool) bool {
	return c.grid.HasLineOfSight(c.index, target, blocked)
}

// FieldOfView returns the region of Grid.FieldOfView from the cell.
func (c Cell[T]) FieldOfView(radius int, blocked func(Cell[T]) bool) Region {
	return c.grid.FieldOfView(c.index, radius, blocked)
}

// PathTo returns the path of Grid.Path from the cell to the index.
func (c Cell[T]) PathTo(index geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return c.grid.Path(c.index, index, passable, cost)
}

// Field returns the field of Grid.Field from the cell.
func (c Cell[T]) Field(passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	return c.grid.Field(c.index, passable, cost)
}

// Reachable returns the field of Grid.Reachable from the cell.
func (c Cell[T]) Reachable(budget float64, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	return c.grid.Reachable(c.index, budget, passable, cost)
}

// IsValid reports whether the grid contains the cell, as Grid.Contains does of its index.
func (c Cell[T]) IsValid() bool {
	return c.grid.Contains(c.index)
}

// String returns the cell and its index: Cell(x,y).
func (c Cell[T]) String() string {
	return "Cell" + c.index.String()
}
