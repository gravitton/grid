package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

// testNewHex runs the sections the two hexagonal constructors share.
func testNewHex(t *testing.T, construct func(opts ...func(*HexOptions)) *Grid[int]) {
	g := construct()
	second := geom.Pt(1, 1)

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, g.Size(), geom.Sz(4, 3))
	})
	t.Run("connects six neighbors", func(t *testing.T) {
		assert.Length(t, g.Cell(second).Neighbors(), 6)
	})
	t.Run("sits with its top-left corner at the world origin by default", func(t *testing.T) {
		geomtest.AssertPoint(t, g.Bounds().Min, geom.Pt(0.0, 0.0))
	})
	t.Run("every center maps back to its index", func(t *testing.T) {
		for cell := range g.All() {
			assert.Equal(t, g.IndexAt(cell.Center()), cell.Index(), cell.Index().String()+": ")
		}
	})
	t.Run("EvenOffset shifts the other rows or columns", func(t *testing.T) {
		even := construct(func(options *HexOptions) {
			options.EvenOffset = true
		})

		assert.NotEqual(t, even.Cell(second).Center(), g.Cell(second).Center())
		for cell := range even.All() {
			assert.Equal(t, even.IndexAt(cell.Center()), cell.Index(), cell.Index().String()+": ")
		}
	})
	t.Run("Origin moves every cell", func(t *testing.T) {
		moved := construct(func(options *HexOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		geomtest.AssertPoint(t, moved.Bounds().Min, geom.Pt(100.0, 50.0))
		geomtest.AssertPoint(t, moved.Cell(second).Center(), g.Cell(second).Center().Add(geom.Vec(100.0, 50.0)))
	})
}

func TestNewHexFlatTop(t *testing.T) {
	testNewHex(t, func(opts ...func(*HexOptions)) *Grid[int] {
		return NewHexFlatTop[int](geom.Sz(4, 3), geom.SzU(16.0), opts...)
	})
}

func TestNewHexPointyTop(t *testing.T) {
	testNewHex(t, func(opts ...func(*HexOptions)) *Grid[int] {
		return NewHexPointyTop[int](geom.Sz(4, 3), geom.SzU(16.0), opts...)
	})
}

func TestHexFlatTopCellSize(t *testing.T) {
	size := HexFlatTopCellSize(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(32.0, 32.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexFlatTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 55.4256258))
	})
}

func TestHexPointyTopCellSize(t *testing.T) {
	size := HexPointyTopCellSize(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(36.9504172, 36.9504172))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexPointyTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 73.9008346))
	})
}

func TestHexFlatTopIsometricPixelPerfectCellSize(t *testing.T) {
	size := HexFlatTopIsometricPixelPerfectCellSize(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(32.0, 27.7128129))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexFlatTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 48.0))
	})
}

func TestHexPointyTopIsometricPixelPerfectCellSize(t *testing.T) {
	size := HexPointyTopIsometricPixelPerfectCellSize(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(36.9504172, 32.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexPointyTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 64.0))
	})
}
