package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

func TestCell_Index(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.Cell(geom.Pt(1, 3)).Index(), geom.Pt(1, 3))
}

func TestCell_Center(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(16.0, 16.0))
}

func TestCell_Neighbors(t *testing.T) {
	cardinal := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	diagonal := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0), withDiagonalMovement)

	t.Run("cardinal", func(t *testing.T) {
		t.Run("center", func(t *testing.T) {
			assert.Length(t, cardinal.Cell(geom.Pt(2, 2)).Neighbors(), 4)
		})
		t.Run("corner", func(t *testing.T) {
			assert.Length(t, cardinal.Cell(geom.Pt(0, 0)).Neighbors(), 2)
		})
		t.Run("edge", func(t *testing.T) {
			assert.Length(t, cardinal.Cell(geom.Pt(1, 0)).Neighbors(), 3)
		})
	})
	t.Run("diagonal", func(t *testing.T) {
		t.Run("center", func(t *testing.T) {
			assert.Length(t, diagonal.Cell(geom.Pt(2, 2)).Neighbors(), 8)
		})
		t.Run("corner", func(t *testing.T) {
			assert.Length(t, diagonal.Cell(geom.Pt(0, 0)).Neighbors(), 3)
		})
	})
}

func TestCell_Range(t *testing.T) {
	g := NewRect[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))

	assert.Length(t, g.Cell(geom.Pt(5, 5)).Range(1, nil), 5)
}

func TestCell_DistanceTo(t *testing.T) {
	g := NewRect[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.Cell(geom.Pt(0, 0)).DistanceTo(geom.Pt(3, 4)), 7)
}

func TestCell_PathTo(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	assert.Length(t, g.Cell(geom.Pt(0, 0)).PathTo(geom.Pt(4, 0), nil, nil), 5)
}

func TestCell_Valid(t *testing.T) {
	g := NewRect[int](geom.Sz(4, 4), geom.Sz(32.0, 32.0))

	t.Run("in bounds", func(t *testing.T) {
		assert.True(t, g.Cell(geom.Pt(0, 0)).Valid())
		assert.True(t, g.Cell(geom.Pt(3, 3)).Valid())
	})
	t.Run("out of bounds", func(t *testing.T) {
		assert.False(t, g.Cell(geom.Pt(10, 10)).Valid())
	})
}
