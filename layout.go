package grid

import (
	"fmt"

	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// Layout is the mapping between the indices of a grid and the world: where a cell is drawn and
// which cell a point falls in. A grid hands out its own as [Grid.Layout]; every method returns
// a new layout. Its origin is the top-left corner of its bounds, on every kind of grid. The
// zero value is a layout of no cells, every index at its origin.
type Layout struct {
	kind      kind
	size      geom.Size[int]
	transform transform

	// bounds, cellSize and cellSpacing are read off the transform where the layout is resized
	// or placed, so that reading them costs nothing.
	bounds      geom.Box[float64]
	cellSize    geom.Size[float64]
	cellSpacing geom.Size[float64]
}

// newLayout returns the layout of a grid of the given size in cells of the given kind, with
// the origin at the world origin. It panics for a cell scale [Layout.Resize] panics for.
func newLayout(kind kind, size geom.Size[int], cellScale geom.Size[float64]) Layout {
	l := Layout{
		kind: kind,
		size: size,
	}

	return l.Resize(cellScale)
}

// Size returns the grid dimensions.
func (l Layout) Size() geom.Size[int] {
	return l.size
}

// Bounds returns the box around every cell of the grid. Its top-left corner is the origin of
// the layout, and a grid of no cells has an empty box there.
func (l Layout) Bounds() geom.Box[float64] {
	return l.bounds
}

// CellSize returns the size of a cell as drawn: the width and height of the box around it.
func (l Layout) CellSize() geom.Size[float64] {
	return l.cellSize
}

// CellSpacing returns the distance between adjacent cell centers along each axis.
func (l Layout) CellSpacing() geom.Size[float64] {
	return l.cellSpacing
}

// Translate returns a copy of the layout with the origin shifted by change. It panics where
// [Layout.MoveTo] of the shifted origin does.
func (l Layout) Translate(change geom.Vector[float64]) Layout {
	return l.place(l.bounds.Min.Add(change))
}

// MoveTo returns a copy of the layout with the top-left corner of its bounds at origin. It
// panics for an origin with a NaN or infinite coordinate.
func (l Layout) MoveTo(origin geom.Point[float64]) Layout {
	return l.place(origin)
}

// Resize returns a copy of the layout with the scale its constructor took replaced: the size of
// a square or isometric cell, the circumradius per axis of a hexagon. The origin stays where
// it is. It panics for a scale that is not positive and finite on both axes, and for one so
// small that a step of its lattice is nothing.
func (l Layout) Resize(cellScale geom.Size[float64]) Layout {
	scaled := l.kind.transform().scale(cellScale)

	if !scaled.step.IsPositive() || !scaled.step.IsFinite() {
		panic(fmt.Sprintf("grid: cell of size %s cannot be laid out", cellScale))
	}

	l.transform = scaled
	l.cellSize = scaled.span(l.kind.reach().Multiply(2))
	l.cellSpacing = scaled.span(l.kind.spacing())

	return l.place(l.bounds.Min)
}

// place returns the layout with its transform placed so the top-left corner of the bounds is
// origin. The zero layout has no cells and no step, and keeps every index at its origin.
func (l Layout) place(origin geom.Point[float64]) Layout {
	if !origin.IsFinite() {
		panic(fmt.Sprintf("grid: origin %s is not finite", origin))
	}

	extent := l.extent()

	l.transform = l.transform.place(origin, extent.Min)
	l.bounds = geom.BoxFromMinMax(origin, l.transform.toPoint(extent.Max))

	return l
}

// extent returns the box around every cell of the layout on the lattice of its kind, in whole
// steps from the center of its first cell. The cells that can reach an edge are the four
// corners and, where every other row or column of hexagons is shifted, their neighbors along
// the edges. A layout of no cells has an empty box there.
func (l Layout) extent() geom.Box[int] {
	w, h := l.size.XY()
	if w <= 0 || h <= 0 {
		return geom.Box[int]{}
	}

	cells := [8]geom.Point[int]{
		geom.Pt(0, 0), geom.Pt(w-1, 0), geom.Pt(0, h-1), geom.Pt(w-1, h-1),
		geom.Pt(min(1, w-1), 0), geom.Pt(min(1, w-1), h-1), geom.Pt(0, min(1, h-1)), geom.Pt(w-1, min(1, h-1)),
	}

	low := l.kind.lattice(cells[0])
	high := low
	for _, index := range cells[1:] {
		center := l.kind.lattice(index)
		low = geom.Pt(min(low.X, center.X), min(low.Y, center.Y))
		high = geom.Pt(max(high.X, center.X), max(high.Y, center.Y))
	}

	reach := l.kind.reach()

	return geom.BoxFromMinMax(low.Add(reach.Negate()), high.Add(reach))
}

// ToPoint returns the center of the cell at the given index in the world.
func (l Layout) ToPoint(index geom.Point[int]) geom.Point[float64] {
	return l.transform.toPoint(l.kind.lattice(index))
}

// FromPoint returns the index of the cell the given point of the world falls in. A point on an
// edge goes to the same side in every cell, where the edge is a number a float holds: on a
// square to the cell after it, to the right or below; on a diamond a vertex to the cell it is
// the top of; on a hexagon where hex.FractionalHex.Round puts it.
func (l Layout) FromPoint(point geom.Point[float64]) geom.Point[int] {
	return l.kind.index(l.transform.fromPoint(point))
}

// CellPolygon returns the regular polygon the cell at the given index is drawn as.
func (l Layout) CellPolygon(index geom.Point[int]) geom.RegularPolygon[float64] {
	return geom.RegularPolygonWithOrientation(l.ToPoint(index), l.transform.polygon, l.kind.sides(), l.kind.orientation())
}

// cellBox returns the box around the cell at the given index, its corners points of the
// lattice as its center is, so the boxes of neighbors meet in the same numbers and the box
// around them all is the bounds.
func (l Layout) cellBox(index geom.Point[int]) geom.Box[float64] {
	center := l.kind.lattice(index)
	reach := l.kind.reach()

	return geom.BoxFromMinMax(l.transform.toPoint(center.Add(reach.Negate())), l.transform.toPoint(center.Add(reach)))
}

// kind is one of the layouts a grid is built in, a hexagonal one for each row or column that
// can be the shifted one. The zero value is the rectangular one.
type kind int

const (
	kindRect kind = iota
	kindIsometric
	kindHexFlatTop
	kindHexFlatTopEven
	kindHexPointyTop
	kindHexPointyTopEven
)

// sides returns the number of sides of a cell of the kind.
func (k kind) sides() int {
	switch k {
	case kindHexFlatTop, kindHexFlatTopEven, kindHexPointyTop, kindHexPointyTopEven:
		return 6
	default:
		return 4
	}
}

// orientation returns how the polygon of a cell of the kind is turned.
func (k kind) orientation() geom.Orientation {
	switch k {
	case kindIsometric, kindHexPointyTop, kindHexPointyTopEven:
		return geom.OrientationPointyTop
	default:
		return geom.OrientationFlatTop
	}
}

// system returns the offset coordinate system the indices of a hexagonal kind are in, and
// hex.CoordinateSystemNone for a square one, whose indices need no conversion to reach its
// lattice.
func (k kind) system() hex.CoordinateSystem {
	switch k {
	case kindHexFlatTop:
		return hex.OffsetOddQ
	case kindHexFlatTopEven:
		return hex.OffsetEvenQ
	case kindHexPointyTop:
		return hex.OffsetOddR
	case kindHexPointyTopEven:
		return hex.OffsetEvenR
	default:
		return hex.CoordinateSystemNone
	}
}

// lattice returns where the center of the cell at the index is on the lattice of the kind, in
// whole steps from the center of the first cell. A step is as short as the corner of the box
// around a cell needs to be a point of the lattice too: half a square or a diamond, and for a
// hexagon half its height and a quarter of its width, or the other way for a pointy-top one.
func (k kind) lattice(index geom.Point[int]) geom.Point[int] {
	switch k {
	case kindIsometric:
		return geom.Pt(index.X-index.Y, index.X+index.Y)
	case kindHexFlatTop, kindHexFlatTopEven:
		h := k.system().From(index)

		return geom.Pt(3*h.Q, 2*h.R+h.Q)
	case kindHexPointyTop, kindHexPointyTopEven:
		h := k.system().From(index)

		return geom.Pt(2*h.Q+h.R, 3*h.R)
	default:
		return geom.Pt(2*index.X, 2*index.Y)
	}
}

// reach returns how far the box around a cell of the kind reaches from its center, in steps of
// the lattice.
func (k kind) reach() geom.Vector[int] {
	switch k {
	case kindHexFlatTop, kindHexFlatTopEven:
		return geom.Vec(2, 1)
	case kindHexPointyTop, kindHexPointyTopEven:
		return geom.Vec(1, 2)
	default:
		return geom.Vec(1, 1)
	}
}

// spacing returns how far the centers of two adjacent cells of the kind are apart along each
// axis, in steps of the lattice.
func (k kind) spacing() geom.Vector[int] {
	switch k {
	case kindIsometric:
		return geom.Vec(1, 1)
	case kindHexFlatTop, kindHexFlatTopEven:
		return geom.Vec(3, 2)
	case kindHexPointyTop, kindHexPointyTopEven:
		return geom.Vec(2, 3)
	default:
		return geom.Vec(2, 2)
	}
}

// index returns the index of the cell a position on the lattice of the kind falls in, the
// inverse of lattice for a position between its points. A square and a diamond own the edges
// their position rounds up from, and a hexagon the ones hex.FractionalHex.Round gives it.
func (k kind) index(position geom.Point[float64]) geom.Point[int] {
	x, y := position.XY()

	switch k {
	case kindIsometric:
		return geom.Pt(geom.RoundHalfUp((x+y)/2), geom.RoundHalfUp((y-x)/2)).Int()
	case kindHexFlatTop, kindHexFlatTopEven:
		q := x / 3

		return k.system().To(hex.FracPt(q, (y-q)/2).Round())
	case kindHexPointyTop, kindHexPointyTopEven:
		r := y / 3

		return k.system().To(hex.FracPt((x-r)/2, r).Round())
	default:
		return geom.Pt(geom.RoundHalfUp(x/2), geom.RoundHalfUp(y/2)).Int()
	}
}

// transform returns the mapping of a cell of the kind at unit scale, not placed yet.
func (k kind) transform() transform {
	switch k {
	case kindIsometric:
		return transform{
			step:    geom.Sz(0.5, 0.5),
			polygon: geom.Sz(0.5, 0.5),
		}
	case kindHexFlatTop, kindHexFlatTopEven:
		return transform{
			step:    geom.Sz(0.5, geom.Sqrt3/2),
			polygon: geom.Sz(1.0, 1.0),
		}
	case kindHexPointyTop, kindHexPointyTopEven:
		return transform{
			step:    geom.Sz(geom.Sqrt3/2, 0.5),
			polygon: geom.Sz(1.0, 1.0),
		}
	default:
		return transform{
			step:    geom.Sz(0.5, 0.5),
			polygon: geom.Sz(geom.Sqrt2/2, geom.Sqrt2/2),
		}
	}
}

// transform is the mapping between the lattice of a kind and the world: the step between two
// points of the lattice along each axis, where the corner of the bounds is in the world and on
// the lattice, and the size of the polygon. A kind gives its own at unit scale, and a layout
// scales and places it. The size of a cell and the spacing of two are whole steps of the kind,
// its reach doubled and its spacing, which a layout reads off once.
//
// Every center, edge and corner of the box around a cell is the origin and a whole number of
// steps, so a point goes to the world in one product and one sum per axis and comes back in
// one difference and one quotient, each exact where its result is a number a float holds.
// That keeps an edge on the same side in every cell and the corner of the bounds in the first
// one; a matrix each way, as it was until 2026-10-07, folds the origin into a constant and
// rounds twice. The polygon of a square is measured by its circumradius, half the diagonal of
// the cell.
type transform struct {
	origin  geom.Point[float64]
	corner  geom.Point[int]
	step    geom.Size[float64]
	polygon geom.Size[float64]
}

// scale returns the transform of a cell the given factor larger along each axis.
func (t transform) scale(factor geom.Size[float64]) transform {
	x, y := factor.XY()

	t.step = t.step.ScaleXY(x, y)
	t.polygon = t.polygon.ScaleXY(x, y)

	return t
}

// place returns the transform with the given point of the lattice, the corner of the bounds,
// at origin in the world.
func (t transform) place(origin geom.Point[float64], corner geom.Point[int]) transform {
	t.origin = origin
	t.corner = corner

	return t
}

// span returns the extent in the world of the given number of steps along each axis.
func (t transform) span(steps geom.Vector[int]) geom.Size[float64] {
	return t.step.ScaleXY(steps.Float().XY())
}

// toPoint returns the point of the world at the given point of the lattice.
func (t transform) toPoint(lattice geom.Point[int]) geom.Point[float64] {
	return t.origin.Add(lattice.Subtract(t.corner).Float().MultiplyXY(t.step.XY()))
}

// fromPoint returns the position on the lattice of the given point of the world, in steps
// along each axis. A transform of no scale has every point at the first point of the lattice.
func (t transform) fromPoint(point geom.Point[float64]) geom.Point[float64] {
	if !t.step.IsPositive() {
		return geom.Point[float64]{}
	}

	return point.Subtract(t.origin).DivideXY(t.step.XY()).Add(t.corner.Vector().Float()).Point()
}
