package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
	hex "github.com/gravitton/hexagon"
	"github.com/gravitton/x/slices"
)

// NewHexagonPointyTopGrid creates a new hexagon grid with a pointy-top layout.
//
// hexSize is the circumradius of each hex cell — the distance from the center
// to any vertex, which equals the edge length for a regular hexagon.
// Use equal W and H for an undistorted hex; different values skew the cell
// independently along each axis (e.g. Sz(1, 0.5) produces a hex that is
// compressed vertically). The pixel bounding box of each cell is
// hexSize.W*√3 wide and hexSize.H*2 tall.
//
// Grid cell (0, 0) is placed at the top-left: even rows are unshifted and
// odd rows are offset half a cell to the right, so no cell center has a
// negative x coordinate.
func NewHexagonPointyTopGrid[T any](size ints.Size, hexSize floats.Size, opts ...HexGridOption) *Grid[T] {
	return newHexagonGrid[T](size, hexSize, HexPointyTop, hex.OffsetOddR, opts)
}

// NewHexagonFlatTopGrid creates a new hexagon grid with a flat-top layout.
//
// hexSize is the circumradius of each hex cell — the distance from the center
// to any vertex, which equals the edge length for a regular hexagon.
// Use equal W and H for an undistorted hex; different values skew the cell
// independently along each axis (e.g. Sz(0.5, 1) produces a hex that is
// compressed horizontally). The pixel bounding box of each cell is
// hexSize.W*2 wide and hexSize.H*√3 tall.
//
// Grid cell (0, 0) is placed at the top-left: even columns are unshifted and
// odd columns are offset half a cell downward, so no cell center has a
// negative y coordinate.
func NewHexagonFlatTopGrid[T any](size ints.Size, hexSize floats.Size, opts ...HexGridOption) *Grid[T] {
	return newHexagonGrid[T](size, hexSize, HexFlatTop, hex.OffsetOddQ, opts)
}

// HexFlatTopCellSize returns the circumradius for NewHexagonFlatTopGrid where
// each tile is exactly width pixels wide and width·√3/2 pixels tall (2:√3 ratio).
func HexFlatTopCellSize(width float64) floats.Size {
	return geom.SzU(width).Scale(0.5)
}

// HexPointyTopCellSize returns the circumradius for NewHexagonPointyTopGrid where
// each tile is exactly width pixels wide and width·2/√3 pixels tall (√3:2 ratio).
func HexPointyTopCellSize(width float64) floats.Size {
	return geom.SzU(width * 2 / geom.Sqrt3).Scale(0.5)
}

// HexFlatTopIsometricPixelPerfectCellSize returns the circumradius for
// NewHexagonFlatTopGrid where each tile is exactly width pixels wide and
// width·3/4 pixels tall (4:3 ratio), suitable for pixel-perfect isometric rendering.
func HexFlatTopIsometricPixelPerfectCellSize(width float64) floats.Size {
	return geom.Sz(width, width*geom.Sqrt3/2).Scale(0.5)
}

// HexPointyTopIsometricPixelPerfectCellSize returns the circumradius for
// NewHexagonPointyTopGrid where each tile is exactly width pixels wide and
// width pixels tall (1:1 ratio), suitable for pixel-perfect isometric rendering.
func HexPointyTopIsometricPixelPerfectCellSize(width float64) floats.Size {
	return geom.Sz(width*2/geom.Sqrt3, width).Scale(0.5)
}

func newHexagonGrid[T any](size ints.Size, hexSize floats.Size, transform *Transform, system hex.CoordinateSystem, opts []HexGridOption) *Grid[T] {
	o := applyHexGridOptions(opts)

	if o.even {
		switch system {
		case hex.OffsetOddR:
			system = hex.OffsetEvenR
		case hex.OffsetOddQ:
			system = hex.OffsetEvenQ
		default:
			// noop
		}
	}

	layout := NewLayout(transform,
		LayoutOpts.Size(size),
		LayoutOpts.CellSize(hexSize),
		LayoutOpts.ToPointMapper(func(index ints.Point) floats.Point {
			return hex.From(index, system).Point().Float()
		}),
		LayoutOpts.FromPointMapper(func(pixel floats.Point) ints.Point {
			return hex.To(hex.FracPt(pixel.XY()).Round(), system)
		}),
	).AlignTopLeft()

	return NewGrid[T](
		layout,
		func(from, to ints.Point) int {
			return hex.From(from, system).DistanceTo(hex.From(to, system))
		},
		func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point {
			h := hex.From(index, system)
			candidates := h.Range(n)

			var blocking []hex.Hex
			for _, i := range candidates {
				if !valid(hex.To(i, system)) {
					blocking = append(blocking, i)
				}
			}

			return slices.Map(h.FieldOfView(candidates, blocking), func(h hex.Hex) ints.Point {
				return hex.To(h, system)
			})
		},
		func(index ints.Point) []ints.Vector {
			return hex.NeighborOffsets(index, system)
		},
	)
}

// HexGridOption configures a hexagonal grid constructor.
type HexGridOption func(*hexGridOptions)

type hexGridOptions struct {
	even bool
}

// HexGridOpts is the namespace for hexagonal grid options.
var HexGridOpts hexGridOptions

// EvenSystem returns an option that switches the offset row/column convention
// from odd-offset (default) to even-offset.
func (o hexGridOptions) EvenSystem() HexGridOption {
	return func(o *hexGridOptions) {
		o.even = true
	}
}

func applyHexGridOptions(opts []HexGridOption) *hexGridOptions {
	o := &hexGridOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return o
}
