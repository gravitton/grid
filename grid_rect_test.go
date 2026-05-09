package grid

import (
	"testing"

	geom "github.com/gravitton/geometry"
)

func TestRectCellSize(t *testing.T) {
	size := RectCellSize(64)
	geom.AssertSize(t, size, 64.0, 64.0)
	g := NewRectGrid[int](geom.SzU(1), size)
	geom.AssertSize(t, g.CellBounds(), 64.0, 64.0)
}

func TestIsometricRectCellSize(t *testing.T) {
	// height = width·tan30° = width/√3 ≈ 36.9504172
	size := IsometricRectCellSize(64)
	geom.AssertSize(t, size, 64.0, 36.9504172)
	g := NewIsometricRectGrid[int](geom.SzU(1), size)
	geom.AssertSize(t, g.CellBounds(), 64.0, 36.9504172)
}

func TestIsometricPixelPerfectRectCellSize(t *testing.T) {
	// 2:1 ratio
	size := IsometricPixelPerfectRectCellSize(64)
	geom.AssertSize(t, size, 64.0, 32.0)
	g := NewIsometricRectGrid[int](geom.SzU(1), size)
	geom.AssertSize(t, g.CellBounds(), 64.0, 32.0)
}
