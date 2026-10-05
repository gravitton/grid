package grid

import (
	"strconv"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

func TestGrid_Size(t *testing.T) {
	g := NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.Size(), geom.Sz(10, 5))
	assert.Equal(t, g.Width(), 10)
	assert.Equal(t, g.Height(), 5)
}

func TestGrid_Layout(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))

	t.Run("maps as the grid does", func(t *testing.T) {
		layout := g.Layout()

		assert.Equal(t, layout.Size(), g.Size())
		geomtest.AssertPoint(t, layout.ToPoint(geom.Pt(2, 3)), g.Cell(geom.Pt(2, 3)).Center())
		assert.Equal(t, layout.FromPoint(geom.Pt(70.0, 40.0)), g.IndexAt(geom.Pt(70.0, 40.0)))
	})
	t.Run("moving the copy leaves the grid where it is", func(t *testing.T) {
		moved := g.Layout().MoveTo(geom.Pt(500.0, 500.0))

		geomtest.AssertPoint(t, moved.ToPoint(geom.Pt(0, 0)), geom.Pt(500.0, 500.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(32.0, 16.0))
	})
}

func TestGrid_CellSize(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))

	geomtest.AssertSize(t, g.CellSize(), geom.Sz(64.0, 32.0))
	assert.Equal(t, g.CellWidth(), 64.0)
	assert.Equal(t, g.CellHeight(), 32.0)
}

func TestGrid_CellSpacing(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))

	geomtest.AssertSize(t, g.CellSpacing(), geom.Sz(64.0, 32.0))
}

func TestGrid_Cell(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	g.Set(geom.Pt(1, 2), 42)

	t.Run("is a handle on the index", func(t *testing.T) {
		assert.Equal(t, g.Cell(geom.Pt(1, 2)).Index(), geom.Pt(1, 2))
		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
	t.Run("outside the grid has no value", func(t *testing.T) {
		assert.Nil(t, g.Cell(geom.Pt(99, 99)).Value())
	})
}

func TestGrid_IndexAt(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.IndexAt(geom.Pt(64.0, 96.0)), geom.Pt(2, 3))
}

func TestGrid_CellAt(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.CellAt(geom.Pt(64.0, 96.0)).Index(), geom.Pt(2, 3))
}

func TestGrid_Set(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	t.Run("stores the value", func(t *testing.T) {
		g.Set(geom.Pt(1, 2), 42)

		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
	t.Run("a write outside the grid does nothing", func(t *testing.T) {
		g.Set(geom.Pt(99, 99), 7)

		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
}

func TestGrid_Fill(t *testing.T) {
	g := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	g.Fill(5)

	for cell := range g.All() {
		assert.Equal(t, *cell.Value(), 5)
	}
}

func TestGrid_Clear(t *testing.T) {
	g := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	g.Fill(5)
	g.Clear()

	for cell := range g.All() {
		assert.Equal(t, *cell.Value(), 0)
	}
}

func TestGrid_Clone(t *testing.T) {
	original := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	original.Fill(1)

	clone := original.Clone()
	clone.Set(geom.Pt(0, 0), 99)

	t.Run("holds the same values", func(t *testing.T) {
		assert.Equal(t, clone.Size(), original.Size())
		assert.Equal(t, *clone.Cell(geom.Pt(2, 2)).Value(), 1)
	})
	t.Run("a write to the clone leaves the original alone", func(t *testing.T) {
		assert.Equal(t, *original.Cell(geom.Pt(0, 0)).Value(), 1)
		assert.Equal(t, *clone.Cell(geom.Pt(0, 0)).Value(), 99)
	})
}

func TestGrid_All(t *testing.T) {
	g := NewHexFlatTop[int](geom.Sz(3, 2), geom.SzU(16.0))

	t.Run("yields every cell once, row by row", func(t *testing.T) {
		var indices []geom.Point[int]
		for cell := range g.All() {
			indices = append(indices, cell.Index())
		}

		assert.Equal(t, indices, []geom.Point[int]{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(2, 0), geom.Pt(0, 1), geom.Pt(1, 1), geom.Pt(2, 1)})
	})
	t.Run("a cell writes through to the grid", func(t *testing.T) {
		for cell := range g.All() {
			cell.Set(cell.Index().X)
		}

		assert.Equal(t, *g.Cell(geom.Pt(2, 1)).Value(), 2)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range g.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
}

func TestGrid_Range(t *testing.T) {
	center := geom.Pt(5, 5)
	cardinal := NewRect[int](geom.Sz(11, 11), geom.Sz(32.0, 32.0))
	diagonal := NewRect[int](geom.Sz(11, 11), geom.Sz(32.0, 32.0), withDiagonalMovement)
	hexagonal := NewHexFlatTop[int](geom.Sz(11, 11), geom.Sz(32.0, 32.0))
	grids := map[string]*Grid[int]{"cardinal": cardinal, "diagonal": diagonal, "hexagonal": hexagonal}

	t.Run("a negative radius is empty", func(t *testing.T) {
		assert.Length(t, cardinal.Range(center, -1, nil), 0)
	})
	t.Run("a zero radius is the center alone", func(t *testing.T) {
		assert.Equal(t, cardinal.Range(center, 0, nil), []geom.Point[int]{center})
	})
	t.Run("cardinal is the diamond of Manhattan distance", func(t *testing.T) {
		assert.Length(t, cardinal.Range(center, 1, nil), 5)
		assert.Length(t, cardinal.Range(center, 3, nil), 25)
	})
	t.Run("diagonal is the square of Chebyshev distance", func(t *testing.T) {
		assert.Length(t, diagonal.Range(center, 1, nil), 9)
		assert.Length(t, diagonal.Range(center, 3, nil), 49)
	})
	t.Run("is exactly the cells within the distance", func(t *testing.T) {
		for name, g := range grids {
			for n := range 5 {
				var expected []geom.Point[int]
				for cell := range g.All() {
					if g.Distance(center, cell.Index()) <= n {
						expected = append(expected, cell.Index())
					}
				}

				assert.True(t, Reg(g.Range(center, n, nil)).Equal(Reg(expected)), name+" at "+strconv.Itoa(n)+": ")
			}
		}
	})
	t.Run("a nil passable admits every cell", func(t *testing.T) {
		all := func(cell Cell[int]) bool {
			return true
		}

		assert.Equal(t, cardinal.Range(center, 2, nil), cardinal.Range(center, 2, all))
	})
	t.Run("a cell that is not passable is left out and hides what is behind it", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return cell.Index() != geom.Pt(5, 6)
		}

		for name, g := range grids {
			reach := Reg(g.Range(center, 2, open))

			assert.False(t, reach.Contains(geom.Pt(5, 6)), name+": ")
			assert.False(t, reach.Contains(geom.Pt(5, 7)), name+": ")
			assert.True(t, reach.Contains(geom.Pt(5, 4)), name+": ")
		}
	})
	t.Run("holds no index outside the grid", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range g.Range(geom.Pt(0, 0), 2, nil) {
				assert.True(t, g.Contains(index), name+" at "+index.String()+": ")
			}
		}
	})
}

func TestGrid_Distance(t *testing.T) {
	t.Run("cardinal is the Manhattan distance", func(t *testing.T) {
		g := NewRect[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))

		assert.Equal(t, g.Distance(geom.Pt(0, 0), geom.Pt(0, 0)), 0)
		assert.Equal(t, g.Distance(geom.Pt(0, 0), geom.Pt(3, 4)), 7)
		assert.Equal(t, g.Distance(geom.Pt(3, 4), geom.Pt(0, 0)), 7)
	})
	t.Run("diagonal is the Chebyshev distance", func(t *testing.T) {
		g := NewRect[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0), withDiagonalMovement)

		assert.Equal(t, g.Distance(geom.Pt(0, 0), geom.Pt(3, 4)), 4)
	})
}

func TestGrid_Contains(t *testing.T) {
	g := NewRect[int](geom.Sz(4, 3), geom.Sz(32.0, 32.0))

	t.Run("the corners are inside", func(t *testing.T) {
		assert.True(t, g.Contains(geom.Pt(0, 0)))
		assert.True(t, g.Contains(geom.Pt(3, 2)))
	})
	t.Run("a negative coordinate is outside", func(t *testing.T) {
		assert.False(t, g.Contains(geom.Pt(-1, 0)))
		assert.False(t, g.Contains(geom.Pt(0, -1)))
	})
	t.Run("the size itself is outside", func(t *testing.T) {
		assert.False(t, g.Contains(geom.Pt(4, 0)))
		assert.False(t, g.Contains(geom.Pt(0, 3)))
	})
}
