package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

func withDiagonalMovement(options *RectOptions) {
	options.Movement = MovementDiagonal
}

func TestNewRect(t *testing.T) {
	g := NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0))

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, g.Size(), geom.Sz(10, 5))
	})
	t.Run("connects four neighbors by default", func(t *testing.T) {
		assert.Length(t, g.Cell(geom.Pt(2, 2)).Neighbors(), 4)
	})
	t.Run("sits with its top-left corner at the world origin by default", func(t *testing.T) {
		geomtest.AssertPoint(t, g.Bounds().Min, geom.Pt(0.0, 0.0))
	})
	t.Run("Movement chooses the neighbors", func(t *testing.T) {
		diagonal := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0), withDiagonalMovement)

		assert.Length(t, diagonal.Cell(geom.Pt(2, 2)).Neighbors(), 8)
	})
	t.Run("Origin places the top-left corner", func(t *testing.T) {
		moved := NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0), func(options *RectOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		geomtest.AssertPoint(t, moved.Bounds().Min, geom.Pt(100.0, 50.0))
		geomtest.AssertPoint(t, moved.Cell(geom.Pt(0, 0)).Center(), geom.Pt(116.0, 66.0))
		assert.Equal(t, moved.IndexAt(geom.Pt(116.0, 66.0)), geom.Pt(0, 0))
	})
	t.Run("every callback is applied, in order", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0), withDiagonalMovement, func(options *RectOptions) {
			options.Movement = MovementCardinal
		})

		assert.Length(t, g.Cell(geom.Pt(2, 2)).Neighbors(), 4)
	})
	t.Run("panics for a movement outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0), func(options *RectOptions) {
				options.Movement = MovementNone
			})
		}, "grid: unknown movement -1")
	})
}

func TestNewIsometric(t *testing.T) {
	g := NewIsometric[int](geom.Sz(8, 6), geom.Sz(64.0, 32.0))

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, g.Size(), geom.Sz(8, 6))
	})
	t.Run("Origin places the top-left corner of the first cell", func(t *testing.T) {
		moved := NewIsometric[int](geom.Sz(8, 6), geom.Sz(64.0, 32.0), func(options *RectOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		geomtest.AssertPoint(t, moved.Cell(geom.Pt(0, 0)).Center(), g.Cell(geom.Pt(0, 0)).Center().Add(geom.Vec(100.0, 50.0)))
	})
}

func TestRectCellSize(t *testing.T) {
	size := RectCellSize(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(64.0, 64.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewRect[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 64.0))
	})
}

func TestIsometricRectCellSize(t *testing.T) {
	size := IsometricRectCellSize(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(64.0, 36.9504172))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewIsometric[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 36.9504172))
	})
}

func TestIsometricPixelPerfectRectCellSize(t *testing.T) {
	size := IsometricPixelPerfectRectCellSize(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(64.0, 32.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewIsometric[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 32.0))
	})
}
