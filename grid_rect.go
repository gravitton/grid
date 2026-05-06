package grid

import (
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

// NewRectGrid constructs a new rectangular grid with 4-directional (cardinal) movement.
// Pass RectGridOpts.DiagonalMovement() to enable 8-directional movement.
func NewRectGrid[T any](grid ints.Size, size floats.Size, opts ...RectGridOption) *Grid[T] {
	return newRectGrid[T](grid, size, SquareFlat, opts)
}

// NewIsometricRectGrid constructs a new isometric (diamond-projection) grid with 4-directional movement.
// Pass RectGridOpts.DiagonalMovement() to enable 8-directional movement.
func NewIsometricRectGrid[T any](grid ints.Size, size floats.Size, opts ...RectGridOption) *Grid[T] {
	return newRectGrid[T](grid, size, SquareIsometric, opts)
}

func newRectGrid[T any](grid ints.Size, size floats.Size, transform *Transform, opts []RectGridOption) *Grid[T] {
	o := applyRectGridOptions(opts)

	system := o.movement
	layout := NewLayout(transform,
		LayoutOpts.Size(grid),
		LayoutOpts.CellSize(size),
	).AlignTopLeft()

	return &Grid[T]{
		cells:       Arr[T](grid),
		cellSize:    layout.CellBounds(),
		cellSpacing: layout.CellSpacing(),
		bounds:      layout.Bounds(),
		toPoint:     layout.ToPoint,
		fromPoint:   layout.FromPoint,
		polygon:     layout.CellPolygon,
		distance: func(from, to ints.Point) int {
			return DistanceTo(from, to, system)
		},
		toRange: func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point {
			p := Pt(index.XY())
			candidates := p.Range(n)

			var blocking []ints.Point
			for _, i := range candidates {
				if !valid(i) {
					blocking = append(blocking, i)
				}
			}

			return p.FieldOfView(candidates, blocking)
		},
		neighbors: func(index ints.Point) []ints.Vector {
			return NeighborOffsets(system)
		},
	}
}

// RectGridOption configures a rectangular grid constructor.
type RectGridOption func(*rectGridOptions)

type rectGridOptions struct {
	movement System
}

// RectGridOpts is the namespace for rectangular grid options.
var RectGridOpts rectGridOptions

// DiagonalMovement returns an option that enables 8-directional (Chebyshev) movement.
func (o rectGridOptions) DiagonalMovement() RectGridOption {
	return func(o *rectGridOptions) {
		o.movement = Diagonal
	}
}

func applyRectGridOptions(opts []RectGridOption) *rectGridOptions {
	o := &rectGridOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return o
}
