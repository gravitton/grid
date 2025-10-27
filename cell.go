package grid

import (
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

type Cell[T any] struct {
	grid  *Grid[T]
	index ints.Point
}

func (c Cell[T]) Index() ints.Point {
	return c.index
}

func (c Cell[T]) Size() floats.Size {
	return c.grid.cellSize
}

func (c Cell[T]) Get() *T {
	return c.grid.getData(c.index)
}

func (c Cell[T]) Set(value T) {
	c.grid.setData(c.index, value)
}

func (c Cell[T]) Valid() bool {
	return c.grid.valid(c.index)
}

func (c Cell[T]) Center() floats.Point {
	// TODO: cache
	return c.grid.cellCenter(c.index)
}

func (c Cell[T]) Bounds() floats.Rectangle {
	return c.grid.cellBounds(c.index)
}

func (c Cell[T]) Polygon() floats.RegularPolygon {
	return c.grid.cellPolygon(c.index)
}

func (c Cell[T]) Neighbours() []ints.Point {
	return c.grid.cellNeighbours(c.index)
}

func (c Cell[T]) DistanceTo(to ints.Point) int {
	return c.grid.distance(c.index, to)
}

func (c Cell[T]) Range(n int, valid ValidFunc[T]) []ints.Point {
	return c.grid.Range(c.index, n, valid)
}

func (c Cell[T]) PathTo(to ints.Point, valid ValidFunc[T], cost CostFunc[T]) []*Cell[T] {
	return c.grid.Path(c.index, to, valid, cost)
}
