package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

// --- Construction ---

func TestNewGrid_Size(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0))
	assert.Equal(t, g.Size(), geom.Sz(10, 5))
	assert.Equal(t, g.Width(), 10)
	assert.Equal(t, g.Height(), 5)
}

func TestNewIsometricGrid_Size(t *testing.T) {
	g := NewIsometricRectGrid[int](geom.Sz(8, 6), geom.Sz(64.0, 32.0))
	assert.Equal(t, g.Width(), 8)
	assert.Equal(t, g.Height(), 6)
}

func TestNewGrid_WithDiagonal(t *testing.T) {
	cardinal := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	diagonal := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0), RectGridOpts.DiagonalMovement())
	centre := geom.Pt(2, 2)
	assert.Equal(t, len(cardinal.Get(centre).Neighbors()), 4)
	assert.Equal(t, len(diagonal.Get(centre).Neighbors()), 8)
}

// --- CellBounds / CellSpacing ---

func TestGrid_CellBounds(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))
	geom.AssertSize(t, g.CellBounds(), 64.0, 32.0)
	assert.Equal(t, g.CellWidth(), 64.0)
	assert.Equal(t, g.CellHeight(), 32.0)
}

func TestGrid_CellSpacing(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))
	geom.AssertSize(t, g.CellSpacing(), 64.0, 32.0)
}

// --- Has ---

func TestGrid_Has(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(4, 3), geom.Sz(32.0, 32.0))
	t.Run("valid", func(t *testing.T) {
		assert.True(t, g.Has(geom.Pt(0, 0)))
		assert.True(t, g.Has(geom.Pt(3, 2)))
	})
	t.Run("negative coords", func(t *testing.T) {
		assert.False(t, g.Has(geom.Pt(-1, 0)))
		assert.False(t, g.Has(geom.Pt(0, -1)))
	})
	t.Run("beyond size", func(t *testing.T) {
		assert.False(t, g.Has(geom.Pt(4, 0)))
		assert.False(t, g.Has(geom.Pt(0, 3)))
	})
}

// --- Set / Get ---

func TestGrid_SetGet(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	g.Set(geom.Pt(1, 2), 42)
	assert.Equal(t, *g.Get(geom.Pt(1, 2)).Get(), 42)
	// Out-of-bounds Set is a no-op.
	g.Set(geom.Pt(99, 99), 7)
	assert.Equal(t, *g.Get(geom.Pt(1, 2)).Get(), 42)
}

// --- Fill / Clear ---

func TestGrid_Fill(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	g.Fill(5)
	for y := range 3 {
		for x := range 3 {
			assert.Equal(t, *g.Get(geom.Pt(x, y)).Get(), 5)
		}
	}
}

func TestGrid_Clear(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	g.Fill(5)
	g.Clear()
	assert.Equal(t, *g.Get(geom.Pt(0, 0)).Get(), 0)
	assert.Equal(t, *g.Get(geom.Pt(2, 2)).Get(), 0)
}

// --- Clone ---

func TestGrid_Clone(t *testing.T) {
	g1 := NewRectGrid[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	g1.Fill(1)
	g2 := g1.Clone()
	g2.Set(geom.Pt(0, 0), 99)
	// Original is unchanged.
	assert.Equal(t, *g1.Get(geom.Pt(0, 0)).Get(), 1)
	assert.Equal(t, *g2.Get(geom.Pt(0, 0)).Get(), 99)
}

// --- IndexAt / At ---

func TestGrid_IndexAt(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	// Cell (2, 3) center is at (64, 96); IndexAt should round back.
	assert.Equal(t, g.IndexAt(geom.Pt(64.0, 96.0)), geom.Pt(2, 3))
}

func TestGrid_At(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	assert.Equal(t, g.At(geom.Pt(64.0, 96.0)).Index(), geom.Pt(2, 3))
}

// --- Neighbors ---

func TestGrid_Neighbors(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
		t.Run("center", func(t *testing.T) {
			assert.Equal(t, len(g.Get(geom.Pt(2, 2)).Neighbors()), 4)
		})
		t.Run("corner", func(t *testing.T) {
			assert.Equal(t, len(g.Get(geom.Pt(0, 0)).Neighbors()), 2)
		})
		t.Run("edge", func(t *testing.T) {
			assert.Equal(t, len(g.Get(geom.Pt(1, 0)).Neighbors()), 3)
		})
	})
	t.Run("diagonal", func(t *testing.T) {
		g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0), RectGridOpts.DiagonalMovement())
		t.Run("center", func(t *testing.T) {
			assert.Equal(t, len(g.Get(geom.Pt(2, 2)).Neighbors()), 8)
		})
		t.Run("corner", func(t *testing.T) {
			assert.Equal(t, len(g.Get(geom.Pt(0, 0)).Neighbors()), 3)
		})
	})
}

// --- Distance ---

func TestGrid_Distance(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		g := NewRectGrid[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
		assert.Equal(t, g.Distance(geom.Pt(0, 0), geom.Pt(0, 0)), 0)
		assert.Equal(t, g.Distance(geom.Pt(0, 0), geom.Pt(3, 4)), 7)
		assert.Equal(t, g.Distance(geom.Pt(3, 4), geom.Pt(0, 0)), 7)
	})
	t.Run("diagonal", func(t *testing.T) {
		g := NewRectGrid[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0), RectGridOpts.DiagonalMovement())
		assert.Equal(t, g.Distance(geom.Pt(0, 0), geom.Pt(3, 4)), 4)
	})
}

// --- Range ---

func TestGrid_Range(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
	g.Fill(0)
	all := func(c *Cell[int]) bool { return true }
	r0 := g.Range(geom.Pt(5, 5), 0, all)
	assert.Equal(t, len(r0), 1)
	assert.Equal(t, r0[0], geom.Pt(5, 5))

	r1 := g.Range(geom.Pt(5, 5), 1, all)
	// Range uses Euclidean distance: only cardinal neighbors at exactly distance 1
	assert.Equal(t, len(r1), 5)
}

func TestGrid_Range_Negative(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	// Negative n short-circuits in square.Range; FieldOfView returns empty.
	all := func(c *Cell[int]) bool { return true }
	assert.Equal(t, len(g.Range(geom.Pt(2, 2), -1, all)), 0)
}
