package grid

import (
	"testing"

	geom "github.com/gravitton/geometry"
)

func TestHexFlatTopCellSize(t *testing.T) {
	// circumradius W = H = width/2; pixel bounding box: width × width·√3/2 ≈ 55.4256258
	size := HexFlatTopCellSize(64)
	geom.AssertSize(t, size, 32.0, 32.0)
	g := NewHexagonFlatTopGrid[int](geom.SzU(1), size)
	geom.AssertSize(t, g.CellBounds(), 64.0, 55.4256258)
}

func TestHexPointyTopCellSize(t *testing.T) {
	// circumradius W = H = width/√3 ≈ 36.9504172; pixel bounding box: width × width·2/√3 ≈ 73.9008346
	size := HexPointyTopCellSize(64)
	geom.AssertSize(t, size, 36.9504172, 36.9504172)
	g := NewHexagonPointyTopGrid[int](geom.SzU(1), size)
	geom.AssertSize(t, g.CellBounds(), 64.0, 73.9008346)
}

func TestHexFlatTopIsometricPixelPerfectCellSize(t *testing.T) {
	// 4:3 ratio — circumradius W = width/2, H = width·√3/4 ≈ 27.7128129; pixel bounding box: width × width·3/4
	size := HexFlatTopIsometricPixelPerfectCellSize(64)
	geom.AssertSize(t, size, 32.0, 27.7128129)
	g := NewHexagonFlatTopGrid[int](geom.SzU(1), size)
	geom.AssertSize(t, g.CellBounds(), 64.0, 48.0)
}

func TestHexPointyTopIsometricPixelPerfectCellSize(t *testing.T) {
	// 1:1 ratio — circumradius W = width/√3 ≈ 36.9504172, H = width/2; pixel bounding box: width × width
	size := HexPointyTopIsometricPixelPerfectCellSize(64)
	geom.AssertSize(t, size, 36.9504172, 32.0)
	g := NewHexagonPointyTopGrid[int](geom.SzU(1), size)
	geom.AssertSize(t, g.CellBounds(), 64.0, 64.0)
}
