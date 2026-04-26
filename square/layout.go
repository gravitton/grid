package square

import (
	"math"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

// Layout describes the mapping between square grid coordinates and pixel space.
// It holds the grid orientation (flat or isometric), the cell Size, and the pixel
// Origin that corresponds to grid cell (0,0) center.
type Layout struct {
	transform transform
	Size      floats.Size
	Origin    floats.Point
}

// LayoutFlat constructs a Layout for a standard orthographic (flat) grid.
//
// size is the bounding box of each cell in pixels: size.W is the cell width,
// size.H is the cell height. For square cells use equal W and H.
// origin is the pixel position of the center of grid cell (0, 0).
func LayoutFlat(size floats.Size, origin floats.Point) Layout {
	return Layout{transformFlat, size, origin}
}

// LayoutIsometric constructs a Layout for an isometric (diamond) grid.
//
// size is the bounding box of the diamond tile in pixels: size.W is the full
// horizontal extent (left tip to right tip) and size.H is the full vertical
// extent (top tip to bottom tip). Adjacent cell centers are size.W/2 apart
// horizontally and size.H/2 apart vertically.
//
// For the classic ~30° game isometric look use size.W = 2 * size.H
// (e.g. Sz(64, 32)). For a symmetric diamond use equal W and H.
//
// origin is the pixel position of the center of grid cell (0, 0).
func LayoutIsometric(size floats.Size, origin floats.Point) Layout {
	return Layout{transformIsometric, size, origin}
}

// Bounds returns the pixel width/height of a single cell in this layout.
func (l Layout) Bounds() floats.Size {
	return l.Size.ScaleXY(l.transform.bounds.XY())
}

// Spacing returns the pixel distance between adjacent cell centers along each axis.
func (l Layout) Spacing() floats.Size {
	return l.Size.ScaleXY(l.transform.spacing.XY())
}

// ToPoint converts a grid index to the pixel coordinates of its center.
func (l Layout) ToPoint(index ints.Point) floats.Point {
	return l.Origin.Add(floats.Vec(index.XY()).Transform(l.transform.toPoint).MultiplyXY(l.Size.XY()))
}

// FromPoint converts a pixel point to fractional grid coordinates.
// Call .Round().Int() on the result to obtain the grid index.
func (l Layout) FromPoint(point floats.Point) floats.Point {
	return point.Subtract(l.Origin).DivideXY(l.Size.XY()).Transform(l.transform.fromPoint).Point()
}

// CellPolygon returns the polygon representing the cell at the given grid index.
func (l Layout) CellPolygon(index ints.Point) floats.RegularPolygon {
	return geom.Square(l.ToPoint(index), l.Size.ScaleXY(l.transform.polygon.XY()), l.transform.orientation)
}

type transform struct {
	toPoint     geom.Matrix
	fromPoint   geom.Matrix
	orientation geom.Orientation
	bounds      floats.Vector
	spacing     floats.Vector
	polygon     floats.Vector
}

// transformFlat maps grid (col, row) directly to screen (x, y).
// Inverse is identity: screen / size → grid.
//
// The polygon circumradius is size * sqrt2/2: for a FlatTop square (45° vertices),
// the circumradius equals half the diagonal of the cell's bounding box.
var transformFlat = transform{
	toPoint:     geom.IdentityMatrix(),
	fromPoint:   geom.IdentityMatrix(),
	orientation: geom.FlatTop,
	bounds:      geom.Vec(1.0, 1.0),
	spacing:     geom.Vec(1.0, 1.0),
	polygon:     geom.Vec(math.Sqrt2/2, math.Sqrt2/2),
}

// transformIsometric maps grid (col, row) to screen using a 2D isometric projection:
//
//	screen_x = (col - row) * 0.5 * size.W
//	screen_y = (col + row) * 0.5 * size.H
//
// size.W and size.H are the full pixel width and height of the diamond tile.
// For classic isometric (2:1 diamond, ~30° view angle) use size.W = 2 * size.H.
//
// The polygon is a PointyTop diamond whose vertices sit at the four edges of the
// tile bounding box: half-width = size.W/2, half-height = size.H/2.
//
// Inverse (let nx = screen_x/size.W, ny = screen_y/size.H):
//
//	col = nx + ny
//	row = ny - nx
var transformIsometric = transform{
	toPoint:     geom.Mat(0.5, -0.5, 0, 0.5, 0.5, 0),
	fromPoint:   geom.Mat(1, 1, 0, -1, 1, 0),
	orientation: geom.PointyTop,
	bounds:      geom.Vec(1.0, 1.0),
	spacing:     geom.Vec(0.5, 0.5),
	polygon:     geom.Vec(0.5, 0.5),
}
