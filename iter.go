package grid

import (
	"iter"
	"math"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

// IterOptions configures bounds-based grid iteration.
// Pass a Bounds to restrict iteration to cells near the visible rectangle; one
// extra row and column are included on every edge so partially visible tiles
// are not culled. Border cells may lie outside the grid (cell.Valid() == false).
//
// Draw order is resolved automatically from the grid layout:
//   - Rectangular (SquareFlat): row-major, top-to-bottom.
//   - Isometric (SquareIsometric): diagonal depth order (painter's algorithm) —
//     depths (col+row) are visited in ascending screen-Y order.
//   - Hex flat-top (HexFlatTop): two-pass column iteration — even columns then
//     odd — so tiles in the same visual row are drawn back-to-front.
//   - Hex pointy-top (HexPointyTop): row-major; the row offset is along X only,
//     so a single pass is sufficient for correct draw order.
type IterOptions struct {
	Bounds floats.Rectangle
}

// Iter returns an iterator over grid cells in draw order (painter's algorithm).
// Pass nil to iterate every cell using the full grid bounds.
// Pass an IterOptions to restrict iteration to a viewport subset.
// The iteration strategy is resolved automatically from the grid layout, so
// tiles are always yielded back-to-front regardless of grid type.
// Cells at the boundary of the bounds may lie outside the grid (cell.Valid() == false).
func (g *Grid[T]) Iter(options *IterOptions) iter.Seq[*Cell[T]] {
	if options == nil {
		options = &IterOptions{}
	}

	if options.Bounds.IsZero() {
		options.Bounds = g.Bounds()
	}

	return func(yield func(*Cell[T]) bool) {
		switch g.kind {
		case kindDefault:
			g.iterDefault(options.Bounds, yield)
		case kindIsometric:
			g.iterIsometric(options.Bounds, yield)
		case kindHexagonalFlatTop:
			g.iterHexagonal(options.Bounds, yield, false, true)
		case kindHexagonalPointyTop:
			g.iterHexagonal(options.Bounds, yield, false, false)
		default:
			g.iterDefault(options.Bounds, yield)
		}
	}

}

func (g *Grid[T]) iterDefault(bounds floats.Rectangle, yield func(*Cell[T]) bool) {
	index := g.IndexAt(bounds.Min())

	cols, rows := bounds.Size.Vector().DivideXY(g.cellSpacing.XY()).Round().Int().XY()

	for row := 0; row <= rows; row++ {
		for col := 0; col <= cols; col++ {
			cell := g.Get(index.AddXY(col, row))

			if !yield(cell) {
				return
			}
		}
	}
}

// iterIsometric iterates visible cells for an isometric grid using row-culling
// and painter's algorithm (back-to-front by screen Y).
//
// All cells sharing depth d = col+row project to the same screen Y, so depths
// are iterated in ascending order (larger d = lower on screen = drawn later).
// For each depth only the column range whose screen X falls within bounds is
// scanned, with a per-cell CollisionRectangles check for the boundaries.
//
// Screen-space relationships for SquareIsometric (spacingW = cellWidth/2,
// spacingH = cellHeight/2, origin = center of cell (0,0)):
//
//	screen_y = d * spacingH + origin.Y
//	screen_x = (2*col - d) * spacingW + origin.X
func (g *Grid[T]) iterIsometric(bounds floats.Rectangle, yield func(*Cell[T]) bool) {
	origin := g.cellCenter(ints.Pt(0, 0))
	spacingW := g.cellSpacing.Width
	spacingH := g.cellSpacing.Height
	halfCellH := g.cellSize.Height * 0.5

	// depth d = col + row; screen_y = d*spacingH + origin.Y
	// ascending d → ascending screen Y → painter's algorithm back-to-front
	dMin := int(math.Floor((bounds.Min().Y-origin.Y)/spacingH)) - 1
	dMax := int(math.Ceil((bounds.Max().Y-origin.Y)/spacingH)) + 1

	for d := dMin; d <= dMax; d++ {
		screenY := float64(d)*spacingH + origin.Y

		// row-cull: skip depths whose tile strip doesn't intersect bounds vertically
		if screenY+halfCellH <= bounds.Min().Y || screenY-halfCellH >= bounds.Max().Y {
			continue
		}

		// screen_x = (2*col - d)*spacingW + origin.X
		// → col = (screen_x - origin.X) / (2*spacingW) + d/2
		colLeft := int(math.Floor((bounds.Min().X-origin.X)/(2*spacingW)+float64(d)*0.5)) - 1
		colRight := int(math.Ceil((bounds.Max().X-origin.X)/(2*spacingW)+float64(d)*0.5)) + 1

		for col := colLeft; col <= colRight; col++ {
			row := d - col
			cell := g.Get(ints.Pt(col, row))
			if !geom.CollisionRectangles(cell.Bounds(), bounds) {
				continue
			}
			if !yield(cell) {
				return
			}
		}
	}
}

// iterHexagonal iterates visible cells for a hexagonal offset grid.
//
// Hex offset grids stagger every other column (flat-top) or row (pointy-top)
// by half a cell. Two tiles that appear in the same visual row or column can
// therefore have different parities, so a naïve single pass would draw them in
// the wrong order. When doubleRows or doubleCols is true the iterator runs two
// sub-passes along that axis — one for even-indexed tiles, one for odd-indexed
// tiles — so all tiles in the same visual strip are drawn back-to-front.
//
//   - doubleCols = true for flat-top hex (columns are staggered along Y).
//   - doubleRows = true for pointy-top hex (rows are staggered along X);
//     however, the row stagger is purely horizontal, so a single pass is
//     sufficient for correct draw order and doubleRows is always false.
//
// One extra cell is included on each edge of bounds so partially visible tiles
// at the boundary are not culled. Cells outside the grid are skipped via a
// CollisionRectangles check.
func (g *Grid[T]) iterHexagonal(bounds floats.Rectangle, yield func(*Cell[T]) bool, doubleRows, doubleCols bool) {
	index := g.IndexAt(bounds.Min())

	colsOdd := index.X%2 != 0
	rowsOdd := index.Y%2 != 0

	cols, rows := bounds.Size.Vector().DivideXY(g.cellSpacing.XY()).Round().Int().XY()

	for row := range g.iterHexagonalDirection(rows, doubleRows, rowsOdd) {
		for col := range g.iterHexagonalDirection(cols, doubleCols, colsOdd) {
			cell := g.Get(index.AddXY(col, row))

			if (row < 0 || col < 0 || row > rows || col > cols) && !geom.CollisionRectangles(cell.Bounds(), bounds) {
				continue
			}

			if !yield(cell) {
				return
			}

		}
	}
}

// iterHexagonalDirection returns an index sequence for one axis of hex iteration.
// When double is false it emits a single pass from -1 to max+1 (inclusive),
// adding one cell of padding on each edge for partial visibility.
// When double is true it emits two interleaved passes — even indices then odd
// (or odd then even when the starting index is odd) — so tiles in each parity
// group are visited before the next group, giving correct back-to-front draw
// order within a staggered hex strip.
func (g *Grid[T]) iterHexagonalDirection(max int, double bool, odd bool) iter.Seq[int] {
	if !double {
		return func(yield func(int) bool) {
			for index := -1; index < max+2; index++ {
				if !yield(index) {
					return
				}
			}
		}
	}

	i, j := 0, -1
	if odd {
		i, j = -1, 0
	}

	return func(yield func(int) bool) {
		for index := i; index < max+2; index += 2 {
			if !yield(index) {
				return
			}
		}

		for index := j; index < max+2; index += 2 {
			if !yield(index) {
				return
			}
		}

	}
}
