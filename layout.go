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
// hexagon where hex.FractionalHex.Round puts it. It converts through [geom.Cast] on every
// layout, so it panics for a point with a NaN or infinite coordinate, and a point too far for
// an index to count gives a platform-dependent one.
func (g *Grid[T]) IndexAt(point geom.Point[float64]) geom.Point[int] {
	return g.layout.transform.fromPoint(point)
}

// CellAt returns the cell the point of the world lies in, the cell of IndexAt.
func (g *Grid[T]) CellAt(point geom.Point[float64]) Cell[T] {
	return g.Cell(g.IndexAt(point))
}

// MoveTo places the top-left corner of the bounds of the grid at origin. Every cell keeps its
// value and its size. It panics for an origin with a NaN or infinite coordinate, and for one
// so far that the grid would reach past the largest number or its cells could not be told
// apart.
func (g *Grid[T]) MoveTo(origin geom.Point[float64]) {
	g.layout = g.layout.moveTo(origin)
}

// Resize replaces the scale the constructor of the grid took: the size of a square or
// isometric cell, the circumradius per axis of a hexagon. The origin stays where it is and
// every cell keeps its value. It panics for a scale that is not positive and finite on both
// axes, for one so small that its cells could not be told apart where the grid is, and for one
// so large that the grid would reach past the largest number.
func (g *Grid[T]) Resize(cellScale geom.Size[float64]) {
	g.layout = g.layout.resize(cellScale)
}

// layout is where a grid is in the world: the transform its cells are placed by and what is
// measured from it once. Its origin is the top-left corner of its bounds, on every kind of
// grid.
type layout struct {
	transform transform

	extent      geom.Box[int]
	bounds      geom.Box[float64]
	cellSize    geom.Size[float64]
	cellSpacing geom.Size[float64]
	polygon     geom.RegularPolygon[float64]
}

// newLayout returns the layout of a grid of the given size in cells of the given kind, with
// the origin at the world origin. It panics for a negative size, and for a cell scale resize
// panics for.
func newLayout(kind kind, size geom.Size[int], cellScale geom.Size[float64]) layout {
	if size.Width < 0 || size.Height < 0 {
		panic(fmt.Sprintf("grid: grid of negative size %s", size))
	}

	l := layout{
		transform: transform{kind: kind},
		extent:    kind.extent(size),
	}

	return l.resize(cellScale)
}

// moveTo returns a copy of the layout with the top-left corner of its bounds at origin. It
// panics for an origin the layout does not resolve at: one with a NaN or infinite coordinate,
// or so far that the far corner is past the largest number or a step is lost in the gap
// between two floats.
func (l layout) moveTo(origin geom.Point[float64]) layout {
	l.transform.origin = origin
	l.bounds = geom.BoxFromMinMax(origin, l.transform.at(l.extent.Max))
	if !l.resolves() {
		panic(fmt.Sprintf("grid: grid cannot be laid out at origin %s", origin))
	}

	return l
}

// resize returns a copy of the layout with the scale its constructor took replaced. The origin
// stays where it is. It panics for a scale the layout does not resolve at, one that is not
// positive and finite on both axes or so small that a step of its lattice is nothing or is
// lost at the origin, and for one the size of a cell is not finite at.
func (l layout) resize(cellScale geom.Size[float64]) layout {
	k := l.transform.kind

	step := k.step().ScaleXY(cellScale.XY())

	l.transform = transform{k, l.transform.origin, l.extent.Min, step}
	l.bounds = geom.BoxFromMinMax(l.transform.origin, l.transform.at(l.extent.Max))
	l.cellSize = l.transform.span(k.reach().Multiply(2))
	l.cellSpacing = l.transform.span(k.spacing())
	l.polygon = geom.RegularPolygonWithOrientation(geom.Point[float64]{}, k.radius().ScaleXY(cellScale.XY()), k.sides(), k.orientation())

	if !l.resolves() || !l.cellSize.IsFinite() {
		panic(fmt.Sprintf("grid: cell of size %s cannot be laid out", cellScale))
	}

	return l
}

// resolves reports whether the layout tells a step of its lattice from the next at both corners
// of its bounds, on both axes: the first step in from each corner is a point strictly inside.
// It is false for a step that is zero, negative or not a number, for an origin or a far corner
// that is not finite, and for an origin so far that a step is less than the gap between two
// floats there, where the cells would fall on one another.
func (l layout) resolves() bool {
	origin, corner := l.transform.origin, l.transform.at(l.extent.Max)
	first := l.transform.at(l.extent.Min.AddXY(1, 1))
	last := l.transform.at(l.extent.Max.AddXY(-1, -1))

	return first.X > origin.X && first.Y > origin.Y && last.X < corner.X && last.Y < corner.Y
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

// step returns the distance between two points of the lattice of the kind along each axis, for
// a cell of unit scale. A step is as short as the corner of the box around a cell needs to be
// a point of the lattice too: half a square or a diamond, and for a hexagon half its height
// and a quarter of its width, or the other way for a pointy-top one.
func (k kind) step() geom.Size[float64] {
	switch k {
	case kindHexFlatTop, kindHexFlatTopEven:
		return geom.Sz(0.5, geom.Sqrt3/2)
	case kindHexPointyTop, kindHexPointyTopEven:
		return geom.Sz(geom.Sqrt3/2, 0.5)
	default:
		return geom.Sz(0.5, 0.5)
	}
}

// radius returns the size of the polygon of a cell of the kind at unit scale: the circumradius
// per axis, which on a square is half the diagonal of the cell.
func (k kind) radius() geom.Size[float64] {
	switch k {
	case kindIsometric:
		return geom.Sz(0.5, 0.5)
	case kindHexFlatTop, kindHexFlatTopEven, kindHexPointyTop, kindHexPointyTopEven:
		return geom.Sz(1.0, 1.0)
	default:
		return geom.Sz(geom.Sqrt2/2, geom.Sqrt2/2)
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

// extent returns the box around every cell of a grid of the kind and the given size in cells,
// on its lattice, in whole steps from the center of the first cell. The cells that can reach
// an edge are the four corners and, where every other row or column of hexagons is shifted,
// their neighbors along the edges. A grid of no cells has an empty box there.
func (k kind) extent(size geom.Size[int]) geom.Box[int] {
	w, h := size.XY()
	if w <= 0 || h <= 0 {
		return geom.Box[int]{}
	}

	cells := [8]geom.Point[int]{
		geom.Pt(0, 0), geom.Pt(w-1, 0), geom.Pt(0, h-1), geom.Pt(w-1, h-1),
		geom.Pt(min(1, w-1), 0), geom.Pt(min(1, w-1), h-1), geom.Pt(0, min(1, h-1)), geom.Pt(w-1, min(1, h-1)),
	}

	low := k.lattice(cells[0])
	high := low
	for _, index := range cells[1:] {
		center := k.lattice(index)
		low = geom.Pt(min(low.X, center.X), min(low.Y, center.Y))
		high = geom.Pt(max(high.X, center.X), max(high.Y, center.Y))
	}

	reach := k.reach()

	return geom.BoxFromMinMax(low.Add(reach.Negate()), high.Add(reach))
}

// lattice returns where the center of the cell at the index is on the lattice of the kind, in
// whole steps from the center of the first cell: the index doubled for a square, the diagonals
// through it for a diamond, and for a hexagon its axial coordinate, which its coordinate system turns the index
// into, doubled along the axis its neighbors interlock on.
func (k kind) lattice(index geom.Point[int]) geom.Point[int] {
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

// index returns the index of the cell a position on the lattice of the kind falls in, the
// inverse of lattice for a position between its points. A square and a diamond own the edges
// their position rounds up from, and a hexagon the ones hex.FractionalHex.Round gives it. The
// position is a number on both axes, which its caller sees to.
func (k kind) index(position geom.Point[float64]) geom.Point[int] {
	x, y := position.XY()

	switch k {
	case kindIsometric:
		return geom.Pt(geom.Cast[int](geom.RoundHalfUp((x+y)/2)), geom.Cast[int](geom.RoundHalfUp((y-x)/2)))
	case kindHexFlatTop, kindHexFlatTopEven:
		q := x / 3

		return k.system().To(hex.FracPt(q, (y-q)/2).Round())
	case kindHexPointyTop, kindHexPointyTopEven:
		r := y / 3

		return k.system().To(hex.FracPt((x-r)/2, r).Round())
	default:
		return geom.Pt(geom.Cast[int](geom.RoundHalfUp(x/2)), geom.Cast[int](geom.RoundHalfUp(y/2)))
	}
}

// transform is the mapping between the indices of a kind and the world: the kind, where the
// corner of the bounds is in the world and on the lattice of the kind, and the step between
// two points of the lattice along each axis. It is all a cell is placed by, and small, so the
// grid and its cells read it alone.
//
// Every center, edge and corner of the box around a cell is the origin and a whole number of
// steps, so a point goes to the world in one product and one sum per axis and comes back in
// one difference and one quotient, each exact where its result is a number a float holds.
// That keeps an edge on the same side in every cell and the corner of the bounds in the first
// one; a matrix each way, as it was until 2026-10-07, folds the origin into a constant and
// rounds twice. The arithmetic is written out, each product rounded before it is added, where
// it went through the vectors of geom until 2026-10-08 at twice the time.
type transform struct {
	kind   kind
	origin geom.Point[float64]
	corner geom.Point[int]
	step   geom.Size[float64]
}

// span returns the extent in the world of the given number of steps along each axis.
func (t transform) span(steps geom.Vector[int]) geom.Size[float64] {
	return t.step.ScaleXY(steps.Float().XY())
}

// at returns the point of the world at the given point of the lattice.
func (t transform) at(lattice geom.Point[int]) geom.Point[float64] {
	return geom.Pt(
		t.origin.X+float64(float64(lattice.X-t.corner.X)*t.step.Width),
		t.origin.Y+float64(float64(lattice.Y-t.corner.Y)*t.step.Height),
	)
}

// toPoint returns the center of the cell at the given index in the world.
func (t transform) toPoint(index geom.Point[int]) geom.Point[float64] {
	return t.at(t.kind.lattice(index))
}

// fromPoint returns the index of the cell the given point of the world falls in. It is
// the arithmetic alone and checks nothing: kind.index converts through geom.Cast, which
// panics for what a point that is not finite or a step of zero comes to.
func (t transform) fromPoint(point geom.Point[float64]) geom.Point[int] {
	return t.kind.index(geom.Pt(
		(point.X-t.origin.X)/t.step.Width+float64(t.corner.X),
		(point.Y-t.origin.Y)/t.step.Height+float64(t.corner.Y),
	))
}

// cellBox returns the box around the cell at the given index, its corners points of the
// lattice as its center is, so the boxes of neighbors meet in the same numbers and the box
// around them all is the bounds.
func (t transform) cellBox(index geom.Point[int]) geom.Box[float64] {
	center := t.kind.lattice(index)
	reach := t.kind.reach()

	return geom.BoxFromMinMax(t.at(center.Add(reach.Negate())), t.at(center.Add(reach)))
}
