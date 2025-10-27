package grid

import (
	"iter"
	"math"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
)

// IterConfig is the configuration for the grid iterator.
type IterConfig struct {
	Bounds     floats.Rectangle
	ColsDouble bool
	RowsDouble bool
}

// Iter returns an iterator over the cells in the grid.
func (g *Grid[T]) Iter(config *IterConfig) iter.Seq[*Cell[T]] {
	if config == nil {
		return func(yield func(*Cell[T]) bool) {
			for index := range g.cells.Iter() {
				if !yield(g.cell(index)) {
					return
				}
			}
		}
	}

	bounds := config.Bounds
	index := g.IndexAt(bounds.Min())

	colsOdd := index.X%2 != 0
	rowsOdd := index.Y%2 != 0

	gridX, gridY := bounds.Size.Width/g.cellSpacing.Width, bounds.Size.Height/g.cellSpacing.Height
	cols := int(math.Round(gridX))
	rows := int(math.Round(gridY))

	return func(yield func(*Cell[T]) bool) {
		for row := range g.iter(rows, config.RowsDouble, rowsOdd) {
			for col := range g.iter(cols, config.ColsDouble, colsOdd) {
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
}

func (g *Grid[T]) iter(max int, double bool, odd bool) iter.Seq[int] {
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
