package grid

import (
	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// Layout is the mapping between the indices of a grid and the world: where a cell is drawn and
// which cell a point falls in. A grid hands out its own as [Grid.Layout]; every method returns
// a new layout.
type Layout struct {
	transform *transform
	size      geom.Size[int]
	origin    geom.Point[float64]
	cellScale geom.Size[float64]

	// system is the offset coordinate system the indices of a hexagonal grid are in, and unused
	// on a square one, whose indices are the coordinates the transform works in.
	system hex.CoordinateSystem

	computed computedTransform
}

// newLayout returns the layout of a grid of the given size in cells under the transform, with
// the origin at the world origin.
func newLayout(t *transform, size geom.Size[int], cellScale geom.Size[float64], system hex.CoordinateSystem) Layout {
	l := Layout{
		transform: t,
		size:      size,
		cellScale: cellScale,
		system:    system,
	}

	return l.recalculate()
}

// Size returns the grid dimensions.
func (l Layout) Size() geom.Size[int] {
	return l.size
}

// CellSize returns the size of a cell as drawn: the width and height of the box around it.
func (l Layout) CellSize() geom.Size[float64] {
	return l.computed.bounds
}

// CellSpacing returns the pixel distance between adjacent cell centers along each axis.
func (l Layout) CellSpacing() geom.Size[float64] {
	return l.computed.spacing
}

// Bounds returns the pixel bounding box for the whole grid of cells.
// It is the box around every cell.
func (l Layout) Bounds() geom.Box[float64] {
	return l.cellsBounds(l.borderCells())
}

func (l Layout) kind() kind {
	switch {
	case l.transform.sides == 4 && l.transform.orientation == geom.OrientationFlatTop:
		return kindDefault
	case l.transform.sides == 4 && l.transform.orientation == geom.OrientationPointyTop:
		return kindIsometric
	case l.transform.sides == 6 && l.transform.orientation == geom.OrientationFlatTop:
		return kindHexagonalFlatTop
	case l.transform.sides == 6 && l.transform.orientation == geom.OrientationPointyTop:
		return kindHexagonalPointyTop
	default:
		return kindDefault
	}
}

// borderCells returns the grid indices of cells that may define the pixel
// extent of the grid. For hex grids the extremes can fall in the second or
// second-to-last row/column (due to the alternating half-step shift), so
// those are included in addition to the four corners.
func (l Layout) borderCells() []geom.Point[int] {
	w, h := l.size.XY()
	cells := []geom.Point[int]{
		geom.Pt(0, 0), geom.Pt(w-1, 0),
		geom.Pt(0, h-1), geom.Pt(w-1, h-1),
	}

	if l.transform.sides == 6 {
		if h >= 2 {
			cells = append(cells, geom.Pt(0, 1), geom.Pt(w-1, 1))
			if h >= 3 {
				cells = append(cells, geom.Pt(0, h-2), geom.Pt(w-1, h-2))
			}
		}
		if w >= 2 {
			cells = append(cells, geom.Pt(1, 0), geom.Pt(1, h-1))
			if w >= 3 {
				cells = append(cells, geom.Pt(w-2, 0), geom.Pt(w-2, h-1))
			}
		}
	}

	return cells
}

// cellsBounds returns the box that completely encloses the given
// cell centers, expanded outward by half the cell bounding box on all sides.
func (l Layout) cellsBounds(cells []geom.Point[int]) geom.Box[float64] {
	minX, minY := l.ToPoint(cells[0]).XY()
	maxX, maxY := minX, minY
	for _, idx := range cells[1:] {
		x, y := l.ToPoint(idx).XY()
		minX, minY = min(minX, x), min(minY, y)
		maxX, maxY = max(maxX, x), max(maxY, y)
	}
	hw, hh := l.computed.bounds.Scale(0.5).XY()

	return geom.BoxFromMinMax(
		geom.Pt(minX-hw, minY-hh),
		geom.Pt(maxX+hw, maxY+hh),
	)
}

// Translate returns a copy of the layout with origin shifted by change.
func (l Layout) Translate(change geom.Vector[float64]) Layout {
	l.origin = l.origin.Add(change)

	return l.recalculate()
}

// MoveTo returns a copy of the layout with origin set to the given point.
func (l Layout) MoveTo(origin geom.Point[float64]) Layout {
	l.origin = origin

	return l.recalculate()
}

// Resize returns a copy of the layout with the cell size replaced.
func (l Layout) Resize(size geom.Size[float64]) Layout {
	l.cellScale = size

	return l.recalculate()
}

// AlignTopLeft shifts the origin so the top-left corner of cell (0, 0) sits at the world origin.
func (l Layout) AlignTopLeft() Layout {
	return l.Translate(l.computed.bounds.Scale(0.5).Vector())
}

func (l Layout) recalculate() Layout {
	l.computed = calculateTransform(l.transform, l.cellScale, l.origin)

	return l
}

// ToPoint converts offset grid coordinates to the pixel center of the cell.
// On a hexagonal grid the index is an offset coordinate.
func (l Layout) ToPoint(index geom.Point[int]) geom.Point[float64] {
	return l.mapToPoint(index).Transform(l.computed.toPoint)
}

// FromPoint converts a pixel coordinate to the nearest grid index.
func (l Layout) FromPoint(point geom.Point[float64]) geom.Point[int] {
	return l.mapFromPoint(point.Transform(l.computed.fromPoint))
}

// CellPolygon returns the regular polygon representing the cell at the given grid coordinates.
func (l Layout) CellPolygon(index geom.Point[int]) geom.RegularPolygon[float64] {
	return geom.RegularPolygonWithOrientation(l.ToPoint(index), l.computed.polygon, l.transform.sides, l.transform.orientation)
}

// mapToPoint returns the index in the coordinates the transform works in: axial for a hexagon,
// the index itself for a square.
func (l Layout) mapToPoint(index geom.Point[int]) geom.Point[float64] {
	if l.transform.sides == 6 {
		return l.system.From(index).Point().Float()
	}

	return index.Float()
}

// mapFromPoint returns the index of the cell the coordinates of the transform fall in.
func (l Layout) mapFromPoint(point geom.Point[float64]) geom.Point[int] {
	if l.transform.sides == 6 {
		return l.system.To(hex.FracPt(point.XY()).Round())
	}

	return point.Round().Int()
}

type kind int

const (
	kindDefault kind = iota
	kindIsometric
	kindHexagonalFlatTop
	kindHexagonalPointyTop
)

// transform is the mapping of one kind of cell between the coordinates of the grid and the
// world, for a cell of unit size: the regular polygon the cell is, the matrices both ways, and
// the bounds, spacing and polygon size a cell size scales.
type transform struct {
	sides       int
	orientation geom.Orientation
	toPoint     geom.Matrix[float64]
	fromPoint   geom.Matrix[float64]
	bounds      geom.Vector[float64]
	spacing     geom.Vector[float64]
	polygon     geom.Vector[float64]
}

// The transforms of the four layouts.
var (
	// squareFlat maps grid (col, row) directly to screen (x, y). Each cell is a square
	// with circumradius = size * sqrt2/2 (the half-diagonal of the bounding box).
	squareFlat = &transform{
		sides:       4,
		orientation: geom.OrientationFlatTop,
		toPoint:     geom.IdentityMatrix[float64](),
		fromPoint:   geom.IdentityMatrix[float64](),
		bounds:      geom.Vec(1.0, 1.0),
		spacing:     geom.Vec(1.0, 1.0),
		polygon:     geom.Vec(geom.Sqrt2/2, geom.Sqrt2/2),
	}

	// squareIsometric maps grid (col, row) to an isometric (diamond) projection:
	//   screen_x = (col - row) * 0.5 * size.W
	//   screen_y = (col + row) * 0.5 * size.H
	// For the classic ~30° game-iso look use size.W = 2 * size.H.
	squareIsometric = &transform{
		sides:       4,
		orientation: geom.OrientationPointyTop,
		toPoint:     geom.Mat[float64](0.5, -0.5, 0, 0.5, 0.5, 0),
		fromPoint:   geom.Mat[float64](1, 1, 0, -1, 1, 0),
		bounds:      geom.Vec(1.0, 1.0),
		spacing:     geom.Vec(0.5, 0.5),
		polygon:     geom.Vec(0.5, 0.5),
	}

	// hexFlatTop maps axial hex (q, r) to pixel for flat-top hexagons.
	hexFlatTop = &transform{
		sides:       6,
		orientation: geom.OrientationFlatTop,
		toPoint:     geom.Mat[float64](3.0/2, 0, 0, geom.Sqrt3/2, geom.Sqrt3, 0),
		fromPoint:   geom.Mat[float64](2.0/3, 0, 0, -1.0/3, geom.Sqrt3/3, 0),
		bounds:      geom.Vec(2.0, geom.Sqrt3),
		spacing:     geom.Vec(3.0/2, geom.Sqrt3),
		polygon:     geom.Vec(1.0, 1.0),
	}

	// hexPointyTop maps axial hex (q, r) to pixel for pointy-top hexagons.
	hexPointyTop = &transform{
		sides:       6,
		orientation: geom.OrientationPointyTop,
		toPoint:     geom.Mat[float64](geom.Sqrt3, geom.Sqrt3/2, 0, 0, 3.0/2, 0),
		fromPoint:   geom.Mat[float64](geom.Sqrt3/3, -1.0/3, 0, 0, 2.0/3, 0),
		bounds:      geom.Vec(geom.Sqrt3, 2.0),
		spacing:     geom.Vec(geom.Sqrt3, 3.0/2),
		polygon:     geom.Vec(1.0, 1.0),
	}
)

type computedTransform struct {
	toPoint   geom.Matrix[float64]
	fromPoint geom.Matrix[float64]
	bounds    geom.Size[float64]
	spacing   geom.Size[float64]
	polygon   geom.Size[float64]
}

func calculateTransform(t *transform, size geom.Size[float64], origin geom.Point[float64]) computedTransform {
	factorX, factorY := size.Float().XY()
	deltaX, deltaY := origin.Float().XY()

	return computedTransform{
		toPoint:   t.toPoint.PreScale(factorX, factorY).PreTranslate(deltaX, deltaY),
		fromPoint: t.fromPoint.Unscale(factorX, factorY).Untranslate(deltaX, deltaY),
		bounds:    t.bounds.MultiplyXY(factorX, factorY).Size(),
		spacing:   t.spacing.MultiplyXY(factorX, factorY).Size(),
		polygon:   t.polygon.MultiplyXY(factorX, factorY).Size(),
	}
}
