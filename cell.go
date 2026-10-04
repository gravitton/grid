package grid

import (
	geom "github.com/gravitton/geometry"
)

// Cell is a single cell in a Grid, identified by its index.
// It provides access to the cell's value, spatial properties, and graph operations.
type Cell[T any] struct {
	grid  *Grid[T]
	index geom.Point[int]
}

// Index returns the grid index of the cell.
func (c Cell[T]) Index() geom.Point[int] {
	return c.index
}

// Size returns the size of the cell.
func (c Cell[T]) Size() geom.Size[float64] {
	return c.grid.cellSize
}

// Center returns the world-space center point of the cell.
func (c Cell[T]) Center() geom.Point[float64] {
	// TODO: cache
	return c.grid.cellCenter(c.index)
}

// Bounds returns the world-space bounding box of the cell.
func (c Cell[T]) Bounds() geom.Box[float64] {
	return c.grid.cellBounds(c.index)
}

// Polygon returns the regular polygon representing the cell's shape.
func (c Cell[T]) Polygon() geom.RegularPolygon[float64] {
	return c.grid.cellPolygon(c.index)
}

// Value returns a pointer to the cell's value.
func (c Cell[T]) Value() *T {
	return c.grid.getData(c.index)
}

// Set sets the cell's value.
func (c Cell[T]) Set(value T) {
	c.grid.setData(c.index, value)
}

// Neighbors returns the grid indices of all valid neighbors of the cell.
func (c Cell[T]) Neighbors() []geom.Point[int] {
	return c.grid.cellNeighbors(c.index)
}

// Range returns the indices within n steps of the cell that are passable and in sight of it.
func (c Cell[T]) Range(n int, passable func(*Cell[T]) bool) []geom.Point[int] {
	return c.grid.Range(c.index, n, passable)
}

// DistanceTo returns the grid distance from the cell to the given index.
func (c Cell[T]) DistanceTo(to geom.Point[int]) int {
	return c.grid.distance(c.index, to)
}

// PathTo returns a path from the cell to the given index using A*.
// passable and cost are forwarded to Grid.Path.
func (c Cell[T]) PathTo(to geom.Point[int], passable func(*Cell[T]) bool, cost func(current, next *Cell[T]) float64) []*Cell[T] {
	return c.grid.Path(c.index, to, passable, cost)
}

// Valid reports whether the cell's index is within the grid bounds.
func (c Cell[T]) Valid() bool {
	return c.grid.valid(c.index)
}
