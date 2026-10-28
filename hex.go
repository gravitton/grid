package grid

import (
	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// HexOptions configures a grid of hexagonal cells. The zero value is the default: the odd
// columns or rows shifted, and the top-left corner of the bounds at the world origin.
type HexOptions struct {
	// EvenOffset shifts the even columns of a flat-top grid, or the even rows of a pointy-top
	// one, in place of the odd ones.
	EvenOffset bool

	// Origin is where the top-left corner of the bounds of the grid is in the world.
	Origin geom.Point[float64]
}

// NewHexFlatTop returns a grid of flat-top hexagons of the given size in cells, each of the
// given circumradius per axis: a regular hexagon where the two are equal, one stretched along
// an axis where they are not. The box around a cell is twice the radius wide and √3 times it
// tall. The odd columns are shifted half a cell down, or the even ones with EvenOffset. Each
// opts function sets the fields of the [HexOptions] it is given. It panics for a radius that is
// not positive and finite on both axes, and for an origin with a NaN or infinite coordinate.
func NewHexFlatTop[T any](size geom.Size[int], hexSize geom.Size[float64], opts ...func(*HexOptions)) *Grid[T] {
	return newHex[T](size, hexSize, kindHexFlatTop, kindHexFlatTopEven, opts)
}

// NewHexPointyTop returns a grid of pointy-top hexagons of the given size in cells, each of the
// given circumradius per axis: a regular hexagon where the two are equal, one stretched along
// an axis where they are not. The box around a cell is √3 times the radius wide and twice it
// tall. The odd rows are shifted half a cell right, or the even ones with EvenOffset. Each opts
// function sets the fields of the [HexOptions] it is given. It panics for a radius that is not
// positive and finite on both axes, and for an origin with a NaN or infinite coordinate.
func NewHexPointyTop[T any](size geom.Size[int], hexSize geom.Size[float64], opts ...func(*HexOptions)) *Grid[T] {
	return newHex[T](size, hexSize, kindHexPointyTop, kindHexPointyTopEven, opts)
}

func newHex[T any](size geom.Size[int], hexSize geom.Size[float64], odd, even kind, opts []func(*HexOptions)) *Grid[T] {
	var options HexOptions
	for _, opt := range opts {
		opt(&options)
	}

	kind := odd
	if options.EvenOffset {
		kind = even
	}

	return newGrid[T](
		size,
		newLayout(kind, size, hexSize, options.Origin),
		newHexTopology(kind.system()),
	)
}

// HexFlatTopCellSize returns the circumradius of NewHexFlatTop for a hexagon of the given width
// and √3/2 of it tall.
func HexFlatTopCellSize(width float64) geom.Size[float64] {
	return geom.SzU(width).Scale(0.5)
}

// HexPointyTopCellSize returns the circumradius of NewHexPointyTop for a hexagon of the given
// width and 2/√3 of it tall.
func HexPointyTopCellSize(width float64) geom.Size[float64] {
	return geom.SzU(width * 2 / geom.Sqrt3).Scale(0.5)
}

// HexFlatTopIsometricPixelPerfectCellSize returns the circumradius of NewHexFlatTop for a
// hexagon of the given width and three quarters of it tall, the hexagon of pixel-art isometry.
func HexFlatTopIsometricPixelPerfectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width, width*geom.Sqrt3/2).Scale(0.5)
}

// HexPointyTopIsometricPixelPerfectCellSize returns the circumradius of NewHexPointyTop for a
// hexagon as tall as the given width, the hexagon of pixel-art isometry.
func HexPointyTopIsometricPixelPerfectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width*2/geom.Sqrt3, width).Scale(0.5)
}

// hexTopology is the neighbor rule of hexagonal cells indexed in an offset coordinate system,
// the methods of hex.Hex reached through the system.
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

func (t *hexTopology) rangeLen(radius int) int {
	return hex.RangeLen(radius)
}

func (t *hexTopology) appendWithin(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	center := t.system.From(index)
	if t.byColumn {
		for candidate := range center.RangeSeq(radius) {
			dst = append(dst, t.system.To(candidate))
		}

		return dst
	}

	for x := index.X - radius; x <= index.X+radius; x++ {
		for y := index.Y - radius; y <= index.Y+radius; y++ {
			if candidate := geom.Pt(x, y); center.DistanceTo(t.system.From(candidate)) <= radius {
				dst = append(dst, candidate)
			}
		}
	}

	return dst
}

func (t *hexTopology) ringLen(radius int) int {
	return hex.RingLen(radius)
}

func (t *hexTopology) appendRing(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	for candidate := range t.system.From(index).RingSeq(radius) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) appendSpiral(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	for candidate := range t.system.From(index).SpiralSeq(radius) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) appendLine(dst []geom.Point[int], from, to geom.Point[int]) []geom.Point[int] {
	for candidate := range t.system.From(from).LineSeq(t.system.From(to)) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) sight(center geom.Point[int], radius int, _ []geom.Point[int], marks *sightMarks) {
	sight := t.system.From(center).FieldOfViewRangeSeq(radius, func(candidate hex.Hex) bool {
		return marks.blocks(t.system.To(candidate))
	})
	for visible := range sight {
		marks.see(t.system.To(visible))
	}
}
