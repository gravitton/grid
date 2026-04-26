package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
	sq "github.com/gravitton/grid/square"
)

// NewGrid constructs a new rectangular grid with 4-directional (cardinal) movement.
func NewGrid[T any](grid ints.Size, size floats.Size) *Grid[T] {
	layout := sq.LayoutFlat(size, geom.Pt(0.0, 0.0))
	layout.Origin = layout.Origin.Add(layout.Bounds().Scale(0.5).Vector())
	return newRectGrid[T](grid, layout, sq.Cardinal)
}

// NewIsometricGrid constructs a new isometric grid with 4-directional movement.
// The origin is positioned so that the left edge of the grid starts at x=0
// and the top edge starts at y=0.
func NewIsometricGrid[T any](grid ints.Size, size floats.Size) *Grid[T] {
	layout := sq.LayoutIsometric(size, geom.Pt(0.0, 0.0))
	layout.Origin = layout.Origin.Add(layout.Bounds().Scale(0.5).Vector())
	return newRectGrid[T](grid, layout, sq.Cardinal)
}

func newRectGrid[T any](grid ints.Size, layout sq.Layout, system sq.System) *Grid[T] {
	return &Grid[T]{
		cells:       Arr[T](grid),
		cellSize:    layout.Bounds(),
		cellSpacing: layout.Spacing(),
		bounds:      geom.RectFromMin(layout.Origin, layout.Spacing().ScaleXY(grid.Float().XY())),
		toPoint: func(index ints.Point) floats.Point {
			return layout.ToPoint(index)
		},
		fromPoint: func(point floats.Point) ints.Point {
			return layout.FromPoint(point).Round().Int()
		},
		polygon: func(index ints.Point) floats.RegularPolygon {
			return layout.CellPolygon(index)
		},
		distance: func(from, to ints.Point) int {
			return sq.DistanceTo(from, to, system)
		},
		toRange: func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point {
			candidates := sq.Range(index, n)

			var blocking []ints.Point
			for _, i := range candidates {
				if !valid(i) {
					blocking = append(blocking, i)
				}
			}

			return sq.FieldOfView(index, candidates, blocking)
		},
		neighbors: func(index ints.Point) []ints.Vector {
			return sq.NeighborOffsets(system)
		},
	}
}
