package grid

import (
	"math"

	geom "github.com/gravitton/geometry"
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

// RectCellSize returns the cell size for NewRectGrid where each tile is exactly
// width pixels wide and width pixels tall (1:1 ratio).
func RectCellSize(width float64) floats.Size {
	return geom.SzU(width)
}

// IsometricRectCellSize returns the cell size for NewIsometricRectGrid where each
// diamond tile is width pixels wide at a geometrically accurate 30° isometric angle
// (height = width·tan30° ≈ width·0.577).
// For a pixel-art-friendly 2:1 ratio use IsometricPixelPerfectRectCellSize.
func IsometricRectCellSize(width float64) floats.Size {
	return floats.Sz(width, width*math.Tan(geom.ToRadians(30)))
}

// IsometricPixelPerfectRectCellSize returns the cell size for NewIsometricRectGrid
// where each diamond tile is width pixels wide and width/2 pixels tall (2:1 ratio).
// The 2:1 ratio is the classic pixel-art isometric convention — tile edges land on
// exact pixel boundaries for crisp rendering.
// For a geometrically accurate 30° angle use IsometricRectCellSize.
func IsometricPixelPerfectRectCellSize(width float64) floats.Size {
	return floats.Sz(width, width*0.5)
}

func newRectGrid[T any](grid ints.Size, size floats.Size, transform *Transform, opts []RectGridOption) *Grid[T] {
	o := applyRectGridOptions(opts)

	system := o.movement
	layout := NewLayout(transform,
		LayoutOpts.Size(grid),
		LayoutOpts.CellSize(size),
	).AlignTopLeft()

	return NewGrid[T](
		layout,
		func(from, to ints.Point) int {
			return DistanceTo(from, to, system)
		},
		func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point {
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
		func(index ints.Point) []ints.Vector {
			return NeighborOffsets(system)
		},
	)
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
