package grid

import (
	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// HexOptions configures a grid of hexagonal cells. The zero value is the default: the odd rows
// or columns shifted, and the top-left corner of the grid at the world origin.
type HexOptions struct {
	// EvenOffset shifts the even rows of a pointy-top grid, or the even columns of a flat-top
	// one, where the odd ones are shifted without it.
	EvenOffset bool

	// Origin is the world position of the top-left corner of the bounds of the grid.
	Origin geom.Point[float64]
}

// NewHexFlatTop creates a new hexagon grid with a flat-top layout.
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
func NewHexFlatTop[T any](size geom.Size[int], hexSize geom.Size[float64], opts ...func(*HexOptions)) *Grid[T] {
	return newHex[T](size, hexSize, hexFlatTop, hex.OffsetOddQ, opts)
}

// NewHexPointyTop creates a new hexagon grid with a pointy-top layout.
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
func NewHexPointyTop[T any](size geom.Size[int], hexSize geom.Size[float64], opts ...func(*HexOptions)) *Grid[T] {
	return newHex[T](size, hexSize, hexPointyTop, hex.OffsetOddR, opts)
}

func newHex[T any](size geom.Size[int], hexSize geom.Size[float64], transform *transform, system hex.CoordinateSystem, opts []func(*HexOptions)) *Grid[T] {
	var options HexOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.EvenOffset {
		switch system {
		case hex.OffsetOddR:
			system = hex.OffsetEvenR
		case hex.OffsetOddQ:
			system = hex.OffsetEvenQ
		}
	}

	return newGrid[T](
		newLayout(transform, size, hexSize, system).AlignTopLeft().Translate(options.Origin.Vector()),
		newHexTopology(system),
	)
}

// HexFlatTopCellSize returns the circumradius for NewHexFlatTop where
// each tile is exactly width pixels wide and width·√3/2 pixels tall (2:√3 ratio).
func HexFlatTopCellSize(width float64) geom.Size[float64] {
	return geom.SzU(width).Scale(0.5)
}

// HexPointyTopCellSize returns the circumradius for NewHexPointyTop where
// each tile is exactly width pixels wide and width·2/√3 pixels tall (√3:2 ratio).
func HexPointyTopCellSize(width float64) geom.Size[float64] {
	return geom.SzU(width * 2 / geom.Sqrt3).Scale(0.5)
}

// HexFlatTopIsometricPixelPerfectCellSize returns the circumradius for
// NewHexFlatTop where each tile is exactly width pixels wide and
// width·3/4 pixels tall (4:3 ratio), suitable for pixel-perfect isometric rendering.
func HexFlatTopIsometricPixelPerfectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width, width*geom.Sqrt3/2).Scale(0.5)
}

// HexPointyTopIsometricPixelPerfectCellSize returns the circumradius for
// NewHexPointyTop where each tile is exactly width pixels wide and
// width pixels tall (1:1 ratio), suitable for pixel-perfect isometric rendering.
func HexPointyTopIsometricPixelPerfectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width*2/geom.Sqrt3, width).Scale(0.5)
}

// hexTopology is the neighbor rule of hexagonal cells indexed in an offset coordinate system,
// the methods of [hex.Hex] reached through the system.
type hexTopology struct {
	system hex.CoordinateSystem

	// steps holds the neighbor offsets of an index of even and of odd parity, since the shifted
	// rows or columns of an offset system step differently from the others.
	steps [2][6]geom.Vector[int]

	// byColumn is true where the parity is that of the column, as in a flat-top grid.
	byColumn bool
}

func newHexTopology(system hex.CoordinateSystem) *hexTopology {
	return &hexTopology{
		system:   system,
		steps:    [2][6]geom.Vector[int]{system.Offsets(geom.Pt(0, 0)), system.Offsets(geom.Pt(1, 1))},
		byColumn: system == hex.OffsetOddQ || system == hex.OffsetEvenQ,
	}
}

func (t *hexTopology) offsets(index geom.Point[int]) []geom.Vector[int] {
	if t.byColumn {
		return t.steps[index.X&1][:]
	}

	return t.steps[index.Y&1][:]
}

func (t *hexTopology) distance(from, to geom.Point[int]) int {
	return t.system.From(from).DistanceTo(t.system.From(to))
}

func (t *hexTopology) rangeOf(index geom.Point[int], n int, passable func(index geom.Point[int]) bool) []geom.Point[int] {
	h := t.system.From(index)

	var candidates, blocking []hex.Hex
	for _, candidate := range h.Range(n) {
		if passable(t.system.To(candidate)) {
			candidates = append(candidates, candidate)
		} else {
			blocking = append(blocking, candidate)
		}
	}

	visible := h.FieldOfView(candidates, blocking)

	indices := make([]geom.Point[int], len(visible))
	for i, seen := range visible {
		indices[i] = t.system.To(seen)
	}

	return indices
}
