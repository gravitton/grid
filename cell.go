package grid

import (
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

// Cell is a single cell in a Grid, identified by its index.
// It provides access to the cell's value, spatial properties, and graph operations.
type Cell[T any] struct {
	grid  *Grid[T]
	index ints.Point
}

// Index returns the grid index of the cell.
func (c Cell[T]) Index() ints.Point {
	return c.index
}

// Size returns the size of the cell.
func (c Cell[T]) Size() floats.Size {
	return c.grid.cellSize
}

// Get returns a pointer to the cell's value.
func (c Cell[T]) Get() *T {
	return c.grid.getData(c.index)
}

// Set sets the cell's value.
func (c Cell[T]) Set(value T) {
	c.grid.setData(c.index, value)
}

// Valid reports whether the cell's index is within the grid bounds.
func (c Cell[T]) Valid() bool {
	return c.grid.valid(c.index)
}

// Center returns the world-space center point of the cell.
func (c Cell[T]) Center() floats.Point {
	// TODO: cache
	return c.grid.cellCenter(c.index)
}

// Bounds returns the world-space bounding rectangle of the cell.
func (c Cell[T]) Bounds() floats.Rectangle {
	return c.grid.cellBounds(c.index)
}

// Polygon returns the regular polygon representing the cell's shape.
func (c Cell[T]) Polygon() floats.RegularPolygon {
	return c.grid.cellPolygon(c.index)
}

// Neighbors returns the grid indices of all valid neighbors of the cell.
func (c Cell[T]) Neighbors() []ints.Point {
	return c.grid.cellNeighbors(c.index)
}

// DistanceTo returns the grid distance from the cell to the given index.
func (c Cell[T]) DistanceTo(to ints.Point) int {
	return c.grid.distance(c.index, to)
}

// Range returns all cell indices within n steps that satisfy valid.
func (c Cell[T]) Range(n int, valid ValidFunc[T]) []ints.Point {
	return c.grid.Range(c.index, n, valid)
}

// PathTo returns a path from the cell to the given index using A*.
// valid and cost are forwarded to Grid.Path.
func (c Cell[T]) PathTo(to ints.Point, valid ValidFunc[T], cost CostFunc[T]) []*Cell[T] {
	return c.grid.Path(c.index, to, valid, cost)
}
