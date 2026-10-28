package grid

import (
	"strconv"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

func withDiagonalMovement(options *RectOptions) {
	options.Movement = MovementDiagonal
}

// testNewRect runs the sections the two square constructors share.
func testNewRect(t *testing.T, construct func(opts ...func(*RectOptions)) *Grid[int]) {
	g := construct()
	inner := geom.Pt(2, 2)

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, g.Size(), geom.Sz(10, 5))
	})
	t.Run("connects four neighbors by default", func(t *testing.T) {
		assert.Length(t, g.Cell(inner).Neighbors(), 4)
	})
	t.Run("every center maps back to its index", func(t *testing.T) {
		for cell := range g.All() {
			assert.Equal(t, g.IndexAt(cell.Center()), cell.Index(), cell.Index().String()+": ")
		}
	})
	t.Run("Movement chooses the neighbors", func(t *testing.T) {
		diagonal := construct(withDiagonalMovement)

		assert.Length(t, diagonal.Cell(inner).Neighbors(), 8)
	})
	t.Run("Origin moves every cell", func(t *testing.T) {
		moved := construct(func(options *RectOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		for cell := range g.All() {
			geomtest.AssertPoint(t, moved.Cell(cell.Index()).Center(), cell.Center().Add(geom.Vec(100.0, 50.0)), cell.Index().String()+": ")
		}
	})
	t.Run("every callback is applied, in order", func(t *testing.T) {
		cardinal := construct(withDiagonalMovement, func(options *RectOptions) {
			options.Movement = MovementCardinal
		})

		assert.Length(t, cardinal.Cell(inner).Neighbors(), 4)
	})
	for _, movement := range []Movement{MovementNone, Movement(7)} {
		number := strconv.Itoa(int(movement))

		t.Run("panics for the movement "+number, func(t *testing.T) {
			assert.PanicsWith(t, func() {
				construct(func(options *RectOptions) {
					options.Movement = movement
				})
			}, "grid: unknown movement "+number)
		})
	}
}

func TestNewRect(t *testing.T) {
	construct := func(opts ...func(*RectOptions)) *Grid[int] {
		return NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0), opts...)
	}

	testNewRect(t, construct)
	t.Run("sits with its top-left corner at the world origin by default", func(t *testing.T) {
		geomtest.AssertPoint(t, construct().Bounds().Min, geom.Pt(0.0, 0.0))
	})
	t.Run("Origin places the top-left corner", func(t *testing.T) {
		moved := construct(func(options *RectOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		geomtest.AssertPoint(t, moved.Bounds().Min, geom.Pt(100.0, 50.0))
		geomtest.AssertPoint(t, moved.Cell(geom.Pt(0, 0)).Center(), geom.Pt(116.0, 66.0))
		assert.Equal(t, moved.IndexAt(geom.Pt(116.0, 66.0)), geom.Pt(0, 0))
	})
}

func TestNewIsometric(t *testing.T) {
	construct := func(opts ...func(*RectOptions)) *Grid[int] {
		return NewIsometric[int](geom.Sz(10, 5), geom.Sz(64.0, 32.0), opts...)
	}

	testNewRect(t, construct)
	t.Run("draws a row down and to the left of the one before", func(t *testing.T) {
		g := construct()

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 1)).Center(), g.Cell(geom.Pt(0, 0)).Center().Add(geom.Vec(-32.0, 16.0)))
	})
	t.Run("the first cell is as far right as the rows reach left", func(t *testing.T) {
		g := construct()

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(160.0, 16.0))
		geomtest.AssertNumber(t, g.Cell(geom.Pt(0, 4)).Bounds().Min.X, 0.0)
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
