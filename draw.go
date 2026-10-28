package grid

import (
	"iter"
	"math"

	geom "github.com/gravitton/geometry"
)

// DrawOrder returns an iterator over the cells in view of bounds, back to front, so a cell is
// never drawn over one in front of it: row by row on a rectangular or pointy-top grid, by the
// sum of column and row on an isometric one, and on a flat-top one row by row with the higher
// columns of each first. A zero bounds is the bounds of the grid. Cells beyond the edge of
// bounds are included where one may be partly in view, and so are cells outside the grid,
// whose IsValid reports false.
func (g *Grid[T]) DrawOrder(bounds geom.Box[float64]) iter.Seq[Cell[T]] {
	if bounds.IsZero() {
		bounds = g.Bounds()
	}

	return func(yield func(Cell[T]) bool) {
		switch g.layout.kind {
		case kindIsometric:
			g.walkIsometric(bounds, yield)
		case kindHexFlatTop, kindHexFlatTopEven:
			g.walkHexagonal(bounds, yield, false, true)
		case kindHexPointyTop, kindHexPointyTopEven:
			g.walkHexagonal(bounds, yield, false, false)
		default:
			g.walkRect(bounds, yield)
		}
	}
}

// walkRect yields the cells of a rectangular grid row by row from the one at the corner of
// bounds.
func (g *Grid[T]) walkRect(bounds geom.Box[float64], yield func(Cell[T]) bool) {
	index := g.IndexAt(bounds.Min)

	cols, rows := bounds.Size().Vector().DivideXY(g.CellSpacing().XY()).Round().Int().XY()

	for row := 0; row <= rows; row++ {
		for col := 0; col <= cols; col++ {
			cell := g.Cell(index.AddXY(col, row))

			if !yield(cell) {
				return
			}
		}
	}
}

// walkIsometric yields the cells of an isometric grid in view of bounds by depth, the sum of
// column and row, which every cell of one screen row shares. A cell of depth d and column c is
// centered at (2c-d) spacings across and d spacings down from the center of the first cell.
func (g *Grid[T]) walkIsometric(bounds geom.Box[float64], yield func(Cell[T]) bool) {
	origin := g.Cell(geom.Pt(0, 0)).Center()
	spacingW := g.CellSpacing().Width
	spacingH := g.CellSpacing().Height
	halfCellH := float64(g.CellSize().Height * 0.5)

	// depth d = col + row; screen_y = d*spacingH + origin.Y
	// ascending d → ascending screen Y → painter's algorithm back-to-front
	dMin := int(math.Floor((bounds.Min.Y-origin.Y)/spacingH)) - 1
	dMax := int(math.Ceil((bounds.Max.Y-origin.Y)/spacingH)) + 1

	for d := dMin; d <= dMax; d++ {
		screenY := float64(float64(d)*spacingH) + origin.Y

		// row-cull: skip depths whose tile strip doesn't intersect bounds vertically
		if screenY+halfCellH <= bounds.Min.Y || screenY-halfCellH >= bounds.Max.Y {
			continue
		}

		// screen_x = (2*col - d)*spacingW + origin.X
		// → col = (screen_x - origin.X) / (2*spacingW) + d/2
		colLeft := int(math.Floor((bounds.Min.X-origin.X)/(2*spacingW)+float64(float64(d)*0.5))) - 1
		colRight := int(math.Ceil((bounds.Max.X-origin.X)/(2*spacingW)+float64(float64(d)*0.5))) + 1

		for col := colLeft; col <= colRight; col++ {
			row := d - col
			cell := g.Cell(geom.Pt(col, row))
			if !cell.Bounds().IntersectsBox(bounds) {
				continue
			}
			if !yield(cell) {
				return
			}
		}
	}
}

// walkHexagonal yields the cells of a hexagonal grid in view of bounds, one cell beyond it
// on each side where it is in view. Where the columns are staggered, two cells of one screen
// row differ in parity, so doubleCols walks each row in two passes, one parity each; a
// pointy-top grid staggers its rows across only, and one pass is in order.
func (g *Grid[T]) walkHexagonal(bounds geom.Box[float64], yield func(Cell[T]) bool, doubleRows, doubleCols bool) {
	index := g.IndexAt(bounds.Min)

	colsOdd := index.X%2 != 0
	rowsOdd := index.Y%2 != 0

	cols, rows := bounds.Size().Vector().DivideXY(g.CellSpacing().XY()).Round().Int().XY()

	for row := range g.hexagonalAxisOrder(rows, doubleRows, rowsOdd) {
		for col := range g.hexagonalAxisOrder(cols, doubleCols, colsOdd) {
			cell := g.Cell(index.AddXY(col, row))

			if (row < 0 || col < 0 || row > rows || col > cols) && !cell.Bounds().IntersectsBox(bounds) {
				continue
			}

			if !yield(cell) {
				return
			}

		}
	}
}

// hexagonalAxisOrder returns the offsets along one axis from -1 to count+1, in one pass, or
// where double is set in two, the offsets that land on an even index first.
func (g *Grid[T]) hexagonalAxisOrder(count int, double bool, odd bool) iter.Seq[int] {
	if !double {
		return func(yield func(int) bool) {
			for index := -1; index < count+2; index++ {
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
		for index := i; index < count+2; index += 2 {
			if !yield(index) {
				return
			}
		}

		for index := j; index < count+2; index += 2 {
			if !yield(index) {
				return
			}
		}
	}
}
