package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

func TestCell_Valid(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(4, 4), geom.Sz(32.0, 32.0))
	t.Run("in bounds", func(t *testing.T) {
		assert.True(t, g.Get(geom.Pt(0, 0)).Valid())
		assert.True(t, g.Get(geom.Pt(3, 3)).Valid())
	})
	t.Run("out of bounds", func(t *testing.T) {
		assert.False(t, g.Get(geom.Pt(10, 10)).Valid())
	})
}

func TestCell_Index(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	assert.Equal(t, g.Get(geom.Pt(1, 3)).Index(), geom.Pt(1, 3))
}

func TestCell_Center(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	// With AlignTopLeft, cell (0,0) center is at (16, 16) for 32×32 cells.
	geom.AssertPoint(t, g.Get(geom.Pt(0, 0)).Center(), 16.0, 16.0)
}

func TestCell_DistanceTo(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
	assert.Equal(t, g.Get(geom.Pt(0, 0)).DistanceTo(geom.Pt(3, 4)), 7)
}

func TestCell_Range(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
	all := func(c *Cell[int]) bool { return true }
	assert.Equal(t, len(g.Get(geom.Pt(5, 5)).Range(1, all)), 5)
}

func TestCell_PathTo(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	assert.Equal(t, len(g.Get(geom.Pt(0, 0)).PathTo(geom.Pt(4, 0), nil, nil)), 5)
}
