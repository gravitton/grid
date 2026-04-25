package grid

import (
	"math"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

// NewGrid constructs a new grid with the given size.
func NewGrid[T any](grid ints.Size, size floats.Size) *Grid[T] {
	return newRectGrid[T](grid, rectLayout{
		origin:      geom.Pt(0.0, 0.0),
		size:        size,
		orientation: geom.FlatTop,
	})
}

// TODO: support isometric grid

type rectLayout struct {
	origin      floats.Point
	size        floats.Size
	orientation geom.Orientation
}

func newRectGrid[T any](grid ints.Size, layout rectLayout) *Grid[T] {
	origin := layout.origin
	size := layout.size
	offset := geom.Vec(size.Scale(0.5).XY())
	directions := []ints.Vector{
		geom.Vec(0, 1),
		geom.Vec(1, 0),
		geom.Vec(0, -1),
		geom.Vec(-1, 0),
	}
	toPoint := func(index ints.Point) floats.Point {
		return geom.Pt(size.Width*float64(index.X), size.Height*float64(index.Y)).Add(offset)
	}

	return &Grid[T]{
		cells:       Arr[T](grid),
		cellSize:    size,
		cellSpacing: size,
		bounds:      geom.RectFromMin(origin, geom.Sz(size.Width*float64(grid.Width), size.Height*float64(grid.Height))),
		toPoint:     toPoint,
		fromPoint: func(point floats.Point) ints.Point {
			return point.DivideXY(size.XY()).Floor().Int()
		},
		polygon: func(index ints.Point) floats.RegularPolygon {
			return geom.Square(toPoint(index), size.ScaleXY(math.Sqrt2/2.0, math.Sqrt2/2.0), layout.orientation)
		},
		distance: func(from, to ints.Point) int {
			return from.ManhattanDistanceTo(to)
		},
		toRange: func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point {
			if n < 0 {
				return nil
			}

			results := make([]ints.Point, 0)
			radiusSquared := n * n

			for dx := -n; dx <= n; dx++ {
				for dy := -n; dy <= n; dy++ {
					if dx*dx+dy*dy <= radiusSquared {
						pt := geom.Pt(index.X+dx, index.Y+dy)
						if valid(pt) {
							results = append(results, pt)
						}
					}
				}
			}
			return results

		},
		neighbors: func(index ints.Point) []ints.Vector {
			return directions
		},
	}
}
