package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
	hex "github.com/gravitton/hexagon"
)

// NewHexagonPointyTopGrid creates a new hexagon grid with a pointy top layout.
func NewHexagonPointyTopGrid[T any](grid ints.Size, hexSize floats.Size, odd bool) *Grid[T] {
	layout := hex.LayoutPointyTop(hexSize, geom.Pt(0.0, 0.0))
	layout.Origin = layout.Origin.Add(layout.Bounds().Scale(0.5).Vector())

	system := hex.OffsetOddR
	if !odd {
		system = hex.OffsetEvenR
	}

	return newHexagonGrid[T](grid, layout, system)
}

// NewHexagonFlatTopGrid creates a new hexagon grid with a flat top layout.
func NewHexagonFlatTopGrid[T any](grid ints.Size, hexSize floats.Size, odd bool) *Grid[T] {
	layout := hex.LayoutFlatTop(hexSize, geom.Pt(0.0, 0.0))
	layout.Origin = layout.Origin.Add(layout.Bounds().Scale(0.5).Vector())

	system := hex.OffsetOddQ
	if !odd {
		system = hex.OffsetEvenQ
	}

	return newHexagonGrid[T](grid, layout, system)
}

func newHexagonGrid[T any](grid ints.Size, layout hex.Layout, system hex.CoordinateSystem) *Grid[T] {
	return &Grid[T]{
		cells:       Arr[T](grid),
		cellSize:    layout.Bounds(),
		cellSpacing: layout.Spacing(),
		bounds:      geom.RectFromMin(layout.Origin, layout.Spacing().ScaleXY(float64(grid.Width), float64(grid.Height))),
		toPoint: func(index ints.Point) floats.Point {
			return layout.ToPoint(hex.From(index, system))
		},
		fromPoint: func(point floats.Point) ints.Point {
			return hex.To(layout.FromPoint(point).Round(), system)
		},
		polygon: func(index ints.Point) floats.RegularPolygon {
			return layout.Hexagon(hex.From(index, system))
		},
		distance: func(from, to ints.Point) int {
			f, t := hex.From(from, system), hex.From(to, system)
			return f.DistanceTo(t)
		},
		toRange: func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point {
			h := hex.From(index, system)
			r := h.Range(n)

			blocking := []hex.Hex{}
			for _, i := range r {
				if !valid(hex.To(i, system)) {
					blocking = append(blocking, i)
				}
			}

			fov := h.FieldOfView(r, blocking)

			results := make([]ints.Point, len(fov))
			for i, fovHex := range fov {
				results[i] = hex.To(fovHex, system)
			}

			return results
		},
		neighbors: func(index ints.Point) []ints.Vector {
			return hex.NeighborOffsets(index, system)
		},
	}
}
