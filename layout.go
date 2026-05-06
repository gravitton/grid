package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

// Layout describes the mapping between grid coordinates and pixel space.
// Create one via NewLayout using a predefined transform (SquareFlat, HexFlatTop, etc.).
type Layout struct {
	transform       *Transform
	size            ints.Size
	origin          floats.Point
	cellSize        floats.Size
	toPointMapper   func(ints.Point) floats.Point
	fromPointMapper func(floats.Point) ints.Point

	computed computedTransform
}

// LayoutOption configures a Layout.
type LayoutOption func(*Layout)
type layoutOptions struct{}

var LayoutOpts layoutOptions

// AlignTopLeft shifts the layout origin so cell (0,0) sits at the world origin.
func (o layoutOptions) AlignTopLeft() LayoutOption {
	return func(l *Layout) {
		*l = l.AlignTopLeft()
	}
}

// Origin sets the pixel position of the grid origin.
func (o layoutOptions) Origin(origin floats.Point) LayoutOption {
	return func(l *Layout) {
		l.origin = origin
	}
}

// Size stores the grid dimensions so Bounds can be called without arguments.
func (o layoutOptions) Size(size ints.Size) LayoutOption {
	return func(l *Layout) {
		l.size = size
	}
}

// CellSize sets the circumradius of each cell (half the bounding-box dimension).
func (o layoutOptions) CellSize(size floats.Size) LayoutOption {
	return func(l *Layout) {
		l.cellSize = size
	}
}

// ToPointMapper stores a function that converts offset grid indices to the
// coordinate system expected by the layout (e.g. axial hex coords).
// ToPoint and CellPolygon apply it automatically when set.
func (o layoutOptions) ToPointMapper(fn func(ints.Point) floats.Point) LayoutOption {
	return func(l *Layout) {
		l.toPointMapper = fn
	}
}

// FromPointMapper stores a function that converts fractional layout coordinates
// (returned by FromPoint) back to offset grid indices.
// IndexFromPoint uses it when set; defaults to rounding to the nearest integer point.
func (o layoutOptions) FromPointMapper(fn func(floats.Point) ints.Point) LayoutOption {
	return func(l *Layout) {
		l.fromPointMapper = fn
	}
}

// NewLayout creates a Layout with the given transform and cell size, with the origin at (0, 0).
func NewLayout(t *Transform, opts ...LayoutOption) Layout {
	l := Layout{transform: t}
	for _, opt := range opts {
		opt(&l)
	}

	return l.recalculate()
}

// With returns a copy of the layout with the given options applied.
func (l Layout) With(opts ...LayoutOption) Layout {
	for _, opt := range opts {
		opt(&l)
	}

	return l.recalculate()
}

// Resize returns a copy of the layout with the cell size replaced.
func (l Layout) Resize(size floats.Size) Layout {
	l.cellSize = size

	return l.recalculate()
}

// Add returns a copy of the layout with origin shifted by change.
func (l Layout) Add(change floats.Vector) Layout {
	l.origin = l.origin.Add(change)

	return l.recalculate()
}

// MoveTo returns a copy of the layout with origin set to the given point.
func (l Layout) MoveTo(origin floats.Point) Layout {
	l.origin = origin

	return l.recalculate()
}

func (l Layout) recalculate() Layout {
	l.computed = calculateTransform(l.transform, l.cellSize, l.origin)

	return l
}

// AlignTopLeft shifts the origin so the top-left corner of cell (0, 0) sits at the world origin.
func (l Layout) AlignTopLeft() Layout {
	return l.Add(l.computed.bounds.Scale(0.5).Vector())
}

// Size returns the grid dimensions.
func (l Layout) Size() ints.Size {
	return l.size
}

// CellBounds returns the pixel bounding box of a single cell.
func (l Layout) CellBounds() floats.Size {
	return l.computed.bounds
}

// CellSpacing returns the pixel distance between adjacent cell centers along each axis.
func (l Layout) CellSpacing() floats.Size {
	return l.computed.spacing
}

// ToPoint converts offset grid coordinates to the pixel center of the cell.
// If an index mapper was set via WithIndexMapper, it is applied first.
func (l Layout) ToPoint(index ints.Point) floats.Point {
	return l.mapToPoint(index).Transform(l.computed.toPoint)
}

func (l Layout) mapToPoint(index ints.Point) floats.Point {
	if l.toPointMapper != nil {
		return l.toPointMapper(index)
	}

	return index.Float()
}

// FromPoint converts a pixel coordinate to the nearest grid index.
func (l Layout) FromPoint(point floats.Point) ints.Point {
	return l.mapFromPoint(point.Transform(l.computed.fromPoint))
}

func (l Layout) mapFromPoint(point floats.Point) ints.Point {
	if l.fromPointMapper != nil {
		return l.fromPointMapper(point)
	}

	return point.Round().Int()
}

// CellPolygon returns the regular polygon representing the cell at the given grid coordinates.
func (l Layout) CellPolygon(index ints.Point) floats.RegularPolygon {
	return geom.RegularPolygonWithOrientation(l.ToPoint(index), l.computed.polygon, l.transform.sides, l.transform.orientation)
}

// Bounds returns the pixel bounding box for the whole grid of cells.
// Grid size must be set via WithGridSize. If an index mapper was set via
// WithIndexMapper it is applied via ToPoint automatically.
func (l Layout) Bounds() floats.Rectangle {
	return l.cellsBounds(l.borderCells(l.size))
}

// borderCells returns the grid indices of cells that may define the pixel
// extent of the grid. For hex grids the extremes can fall in the second or
// second-to-last row/column (due to the alternating half-step shift), so
// those are included in addition to the four corners.
func (l Layout) borderCells(grid ints.Size) []ints.Point {
	w, h := grid.XY()
	cells := []ints.Point{
		ints.Pt(0, 0), ints.Pt(w-1, 0),
		ints.Pt(0, h-1), ints.Pt(w-1, h-1),
	}

	if l.transform.sides == 6 {
		if h >= 2 {
			cells = append(cells, ints.Pt(0, 1), ints.Pt(w-1, 1))
			if h >= 3 {
				cells = append(cells, ints.Pt(0, h-2), ints.Pt(w-1, h-2))
			}
		}
		if w >= 2 {
			cells = append(cells, ints.Pt(1, 0), ints.Pt(1, h-1))
			if w >= 3 {
				cells = append(cells, ints.Pt(w-2, 0), ints.Pt(w-2, h-1))
			}
		}
	}

	return cells
}

// cellsBounds returns the pixel rectangle that completely encloses the given
// cell centers, expanded outward by half the cell bounding box on all sides.
func (l Layout) cellsBounds(cells []ints.Point) floats.Rectangle {
	minX, minY := l.ToPoint(cells[0]).XY()
	maxX, maxY := minX, minY
	for _, idx := range cells[1:] {
		x, y := l.ToPoint(idx).XY()
		minX, minY = min(minX, x), min(minY, y)
		maxX, maxY = max(maxX, x), max(maxY, y)
	}
	hw, hh := l.computed.bounds.Scale(0.5).XY()

	return geom.RectFromMinMax(
		geom.Pt(minX-hw, minY-hh),
		geom.Pt(maxX+hw, maxY+hh),
	)
}

// Transform holds the predefined coordinate mapping data for a grid layout type.
// Use the package-level variables (SquareFlat, SquareIsometric, HexFlatTop, HexPointyTop)
// or NewTransform to construct a Layout.
type Transform struct {
	sides       int
	orientation geom.Orientation
	toPoint     floats.Matrix
	fromPoint   floats.Matrix
	bounds      floats.Vector
	spacing     floats.Vector
	polygon     floats.Vector
}

// NewTransform creates a custom transform for use with NewLayout.
// toPoint and fromPoint must be inverses of each other.
// bounds, spacing, and polygon are unit-space vectors scaled by the cell size at layout creation.
func NewTransform(sides int, orientation geom.Orientation, toPoint, fromPoint floats.Matrix, bounds, spacing, polygon floats.Vector) *Transform {
	return &Transform{sides, orientation, toPoint, fromPoint, bounds, spacing, polygon}
}

// Predefined grid layout transforms.
var (
	// SquareFlat maps grid (col, row) directly to screen (x, y). Each cell is a square
	// with circumradius = size * sqrt2/2 (the half-diagonal of the bounding box).
	SquareFlat = &Transform{
		sides:       4,
		orientation: geom.FlatTop,
		toPoint:     floats.IdentityMatrix(),
		fromPoint:   floats.IdentityMatrix(),
		bounds:      floats.Vec(1.0, 1.0),
		spacing:     floats.Vec(1.0, 1.0),
		polygon:     floats.Vec(geom.Sqrt2/2, geom.Sqrt2/2),
	}

	// SquareIsometric maps grid (col, row) to an isometric (diamond) projection:
	//   screen_x = (col - row) * 0.5 * size.W
	//   screen_y = (col + row) * 0.5 * size.H
	// For the classic ~30° game-iso look use size.W = 2 * size.H.
	SquareIsometric = &Transform{
		sides:       4,
		orientation: geom.PointyTop,
		toPoint:     floats.Mat(0.5, -0.5, 0, 0.5, 0.5, 0),
		fromPoint:   floats.Mat(1, 1, 0, -1, 1, 0),
		bounds:      floats.Vec(1.0, 1.0),
		spacing:     floats.Vec(0.5, 0.5),
		polygon:     floats.Vec(0.5, 0.5),
	}

	// HexPointyTop maps axial hex (q, r) to pixel for pointy-top hexagons.
	HexPointyTop = &Transform{
		sides:       6,
		orientation: geom.PointyTop,
		toPoint:     floats.Mat(geom.Sqrt3, geom.Sqrt3/2, 0, 0, 3.0/2, 0),
		fromPoint:   floats.Mat(geom.Sqrt3/3, -1.0/3, 0, 0, 2.0/3, 0),
		bounds:      floats.Vec(geom.Sqrt3, 2.0),
		spacing:     floats.Vec(geom.Sqrt3, 3.0/2),
		polygon:     floats.Vec(1.0, 1.0),
	}

	// HexFlatTop maps axial hex (q, r) to pixel for flat-top hexagons.
	HexFlatTop = &Transform{
		sides:       6,
		orientation: geom.FlatTop,
		toPoint:     floats.Mat(3.0/2, 0, 0, geom.Sqrt3/2, geom.Sqrt3, 0),
		fromPoint:   floats.Mat(2.0/3, 0, 0, -1.0/3, geom.Sqrt3/3, 0),
		bounds:      floats.Vec(2.0, geom.Sqrt3),
		spacing:     floats.Vec(3.0/2, geom.Sqrt3),
		polygon:     floats.Vec(1.0, 1.0),
	}
)

type computedTransform struct {
	toPoint   floats.Matrix
	fromPoint floats.Matrix
	bounds    floats.Size
	spacing   floats.Size
	polygon   floats.Size
}

func calculateTransform(t *Transform, size floats.Size, origin floats.Point) computedTransform {
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
