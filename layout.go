package grid

import (
	"fmt"

	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// Bounds returns the box around every cell of the grid, in world coordinates. Its top-left
// corner is the origin of the grid, and a grid of no cells has an empty box there.
func (g *Grid[T]) Bounds() geom.Box[float64] {
	return g.layout.bounds
}

// CellSize returns the size of a cell as drawn: the width and height of the box around it,
// the same for every cell and what [Cell.Size] returns.
func (g *Grid[T]) CellSize() geom.Size[float64] {
	return g.layout.cellSize
}

// CellSpacing returns the distance between the centers of adjacent cells along each axis.
func (g *Grid[T]) CellSpacing() geom.Size[float64] {
	return g.layout.cellSpacing
}

// CellWidth returns the width of a cell as drawn, the width of CellSize.
func (g *Grid[T]) CellWidth() float64 {
	return g.layout.cellSize.Width
}

// CellHeight returns the height of a cell as drawn, the height of CellSize.
func (g *Grid[T]) CellHeight() float64 {
	return g.layout.cellSize.Height
}

// IndexAt returns the index of the cell the point of the world lies in. A point outside the
// grid gives an index outside it, where Contains reports false. A point on an edge goes to the
// same side in every cell, where the edge is a number a float holds: on a square to the cell
// after it, to the right or below; on a diamond a vertex to the cell it is the top of; on a
// hexagon where [hex.FractionalHex.Round] puts it. It converts through [geom.Cast] on every
// layout, so it panics for a point with a NaN or infinite coordinate, and a point too far for
// an index to count gives a platform-dependent one.
func (g *Grid[T]) IndexAt(point geom.Point[float64]) geom.Point[int] {
	return g.layout.fromWorld(point)
}

// CellAt returns the cell the point of the world lies in, the cell of IndexAt.
func (g *Grid[T]) CellAt(point geom.Point[float64]) Cell[T] {
	return g.Cell(g.IndexAt(point))
}

// MoveTo places the top-left corner of the bounds of the grid at origin. Every cell keeps its
// value and its size. It panics for an origin with a NaN or infinite coordinate.
func (g *Grid[T]) MoveTo(origin geom.Point[float64]) {
	g.layout.moveTo(origin)
}

// Resize replaces the scale the constructor of the grid took: the size of a square or
// isometric cell, the circumradius per axis of a hexagon. The origin stays where it is and
// every cell keeps its value. It panics for a scale that is not positive and finite on both
// axes.
func (g *Grid[T]) Resize(cellScale geom.Size[float64]) {
	g.layout.resize(cellScale)
}

// layout is where a grid is in the world. Every center, edge and corner of the box around a
// cell is a point of one lattice: the origin, the top-left corner of the bounds, and a whole
// number of steps along each axis. A cell goes to the world in one product and one sum per
// axis and a point comes back in one difference and one quotient, each exact where its result
// is a number a float holds, so an edge is on the same side in every cell and the boxes of
// neighbors meet in the same numbers.
//
// Its methods take a pointer, so the grid and its cells read it in place.
type layout struct {
	kind          kind
	origin        geom.Point[float64]
	latticeStep   geom.Size[float64]
	latticeBounds geom.Box[int]

	// derived from the fields above, by moveTo and resize
	bounds      geom.Box[float64]
	cellSize    geom.Size[float64]
	cellSpacing geom.Size[float64]
	cellPolygon geom.RegularPolygon[float64]
}

// newLayout returns the layout of a grid of the given kind and size in cells, its cells at the
// given scale and the top-left corner of its bounds at origin. It panics where resize and
// moveTo do.
func newLayout(kind kind, size geom.Size[int], cellScale geom.Size[float64], origin geom.Point[float64]) layout {
	l := layout{
		kind:          kind,
		latticeBounds: kind.latticeBounds(size),
	}
	l.resize(cellScale)
	l.moveTo(origin)

	return l
}

// moveTo places the top-left corner of the bounds at origin. It panics for an origin with a
// NaN or infinite coordinate.
func (l *layout) moveTo(origin geom.Point[float64]) {
	if !origin.IsFinite() {
		panic(fmt.Sprintf("grid: origin %s is not finite", origin))
	}

	l.origin = origin
	l.bounds = geom.BoxFromMinMax(origin, l.latticeToWorld(l.latticeBounds.Max))
}

// resize replaces the scale the constructor took and leaves the origin where it is. It panics
// for a scale that is not positive and finite on both axes.
func (l *layout) resize(cellScale geom.Size[float64]) {
	if !cellScale.IsPositive() || !cellScale.IsFinite() {
		panic(fmt.Sprintf("grid: cell scale %s cannot be laid out", cellScale))
	}

	l.latticeStep = l.kind.latticeStep().ScaleXY(cellScale.XY())
	l.bounds = geom.BoxFromMinMax(l.origin, l.latticeToWorld(l.latticeBounds.Max))
	l.cellSize = l.worldSize(l.kind.sizeSteps())
	l.cellSpacing = l.worldSize(l.kind.spacingSteps())
	l.cellPolygon = l.kind.cellPolygon(cellScale)
}

// fromWorld returns the index of the cell the given point of the world falls in. It checks
// nothing: every kind converts through geom.Cast, which panics for a point that is not finite.
func (l *layout) fromWorld(point geom.Point[float64]) geom.Point[int] {
	return l.kind.fromLattice(l.latticeFromWorld(point))
}

// toWorld returns the center of the cell at the given index in the world.
func (l *layout) toWorld(index geom.Point[int]) geom.Point[float64] {
	return l.latticeToWorld(l.kind.toLattice(index))
}

// cellBounds returns the box around the cell at the given index, its corners read off the
// lattice as its center is.
func (l *layout) cellBounds(index geom.Point[int]) geom.Box[float64] {
	center := l.kind.toLattice(index)
	sizeSteps := l.kind.sizeSteps()
	half := geom.Vec(sizeSteps.X/2, sizeSteps.Y/2)

	return geom.BoxFromMinMax(l.latticeToWorld(center.Add(half.Negate())), l.latticeToWorld(center.Add(half)))
}

// latticeFromWorld returns the position on the lattice of the given point of the world, a
// fractional one where the point lies between the points of the lattice. It is the inverse of
// latticeToWorld.
func (l *layout) latticeFromWorld(point geom.Point[float64]) geom.Point[float64] {
	return geom.Pt(
		(point.X-l.origin.X)/l.latticeStep.Width+float64(l.latticeBounds.Min.X),
		(point.Y-l.origin.Y)/l.latticeStep.Height+float64(l.latticeBounds.Min.Y),
	)
}

// latticeToWorld returns the point of the world at the given position on the lattice. Each
// product is rounded before it is added, never fused with the sum.
func (l *layout) latticeToWorld(position geom.Point[int]) geom.Point[float64] {
	return geom.Pt(
		l.origin.X+float64(float64(position.X-l.latticeBounds.Min.X)*l.latticeStep.Width),
		l.origin.Y+float64(float64(position.Y-l.latticeBounds.Min.Y)*l.latticeStep.Height),
	)
}

// worldSize returns the size in the world of the given number of steps along each axis.
func (l *layout) worldSize(steps geom.Vector[int]) geom.Size[float64] {
	return l.latticeStep.ScaleXY(steps.Float().XY())
}

// kind is one of the layouts a grid is built in, a hexagonal one counted twice: with its odd
// columns or rows shifted, and with its even ones. The zero value is the rectangular one.
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

// cellRadius returns the circumradius of the polygon of a cell of the kind, in the world, for
// a cell of unit scale: half the diagonal of a square, half the width of a diamond, and the
// scale itself for a hexagon, which is sized by its radius.
func (k kind) cellRadius() float64 {
	switch k {
	case kindIsometric:
		return 0.5
	case kindHexFlatTop, kindHexFlatTopEven, kindHexPointyTop, kindHexPointyTopEven:
		return 1
	default:
		return geom.Sqrt2 / 2
	}
}

// cellPolygon returns the polygon a cell of the kind is drawn as at the given scale, centered
// on the world origin, for a cell to move to its own center.
func (k kind) cellPolygon(cellScale geom.Size[float64]) geom.RegularPolygon[float64] {
	return geom.RegularPolygonWithOrientation(geom.Point[float64]{}, cellScale.Scale(k.cellRadius()), k.sides(), k.orientation())
}

// sizeSteps returns the size of the box around a cell of the kind, in steps of the lattice. It
// is even along each axis, so the box reaches a whole number of steps from the center of the
// cell.
func (k kind) sizeSteps() geom.Vector[int] {
	switch k {
	case kindHexFlatTop, kindHexFlatTopEven:
		return geom.Vec(4, 2)
	case kindHexPointyTop, kindHexPointyTopEven:
		return geom.Vec(2, 4)
	default:
		return geom.Vec(2, 2)
	}
}

// spacingSteps returns how far the centers of two adjacent cells of the kind are apart along
// each axis, in steps of the lattice.
func (k kind) spacingSteps() geom.Vector[int] {
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

// latticeStep returns the distance between two points of the lattice of the kind along each
// axis, in the world, for a cell of unit scale. A step is short enough for the corners of the
// box around a cell to be points of the lattice, as its center is: half a square or a diamond,
// and for a hexagon half its height and a quarter of its width, or the other way for a
// pointy-top one.
func (k kind) latticeStep() geom.Size[float64] {
	switch k {
	case kindHexFlatTop, kindHexFlatTopEven:
		return geom.Sz(0.5, geom.Sqrt3/2)
	case kindHexPointyTop, kindHexPointyTopEven:
		return geom.Sz(geom.Sqrt3/2, 0.5)
	default:
		return geom.Sz(0.5, 0.5)
	}
}

// latticeBounds returns the box around every cell of a grid of the kind and the given size in
// cells, on its lattice, in whole steps from the center of the first cell. Only the first two
// and the last of the columns and of the rows can reach an edge, the second where every other
// one is shifted. A grid of no cells has an empty box.
func (k kind) latticeBounds(size geom.Size[int]) geom.Box[int] {
	w, h := size.XY()
	if w <= 0 || h <= 0 {
		return geom.Box[int]{}
	}

	var low, high geom.Point[int]
	for _, x := range [3]int{0, min(1, w-1), w - 1} {
		for _, y := range [3]int{0, min(1, h-1), h - 1} {
			center := k.toLattice(geom.Pt(x, y))
			low = geom.Pt(min(low.X, center.X), min(low.Y, center.Y))
			high = geom.Pt(max(high.X, center.X), max(high.Y, center.Y))
		}
	}

	sizeSteps := k.sizeSteps()
	half := geom.Vec(sizeSteps.X/2, sizeSteps.Y/2)

	return geom.BoxFromMinMax(low.Add(half.Negate()), high.Add(half))
}

// fromLattice returns the index of the cell a position on the lattice of the kind falls in,
// the inverse of toLattice for a position between the points of the lattice too. A square and
// a diamond own the edges that round up to them, and a hexagon the ones
// hex.FractionalHex.Round gives it.
func (k kind) fromLattice(position geom.Point[float64]) geom.Point[int] {
	x, y := position.XY()

	switch k {
	case kindIsometric:
		column := geom.RoundHalfUp((x + y) / 2)
		row := geom.RoundHalfUp((y - x) / 2)

		return geom.Pt(geom.Cast[int](column), geom.Cast[int](row))
	case kindHexFlatTop, kindHexFlatTopEven:
		q := x / 3

		return k.system().To(hex.FracPt(q, (y-q)/2).Round())
	case kindHexPointyTop, kindHexPointyTopEven:
		r := y / 3

		return k.system().To(hex.FracPt((x-r)/2, r).Round())
	default:
		column := geom.RoundHalfUp(x / 2)
		row := geom.RoundHalfUp(y / 2)

		return geom.Pt(geom.Cast[int](column), geom.Cast[int](row))
	}
}

// toLattice returns where the center of the cell at the index is on the lattice of the kind,
// in whole steps from the center of the first cell: the index doubled for a square, the
// diagonals through it for a diamond, and for a hexagon its axial coordinate, which its
// coordinate system turns the index into, at the spacing of the kind, with half a row more for
// each column on a flat-top one and half a column more for each row on a pointy-top one.
func (k kind) toLattice(index geom.Point[int]) geom.Point[int] {
	x, y := index.XY()

	switch k {
	case kindIsometric:
		return geom.Pt(x-y, x+y)
	case kindHexFlatTop, kindHexFlatTopEven:
		h := k.system().From(index)

		return geom.Pt(3*h.Q, 2*h.R+h.Q)
	case kindHexPointyTop, kindHexPointyTopEven:
		h := k.system().From(index)

		return geom.Pt(2*h.Q+h.R, 3*h.R)
	default:
		return geom.Pt(2*x, 2*y)
	}
}
