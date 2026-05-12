package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

// --- Iter ---

func TestGrid_Iter_All(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(3, 4), geom.Sz(32.0, 32.0))
	// Iter includes one extra row and column on each edge for partial visibility,
	// so a 3×4 grid yields (3+1)×(4+1) = 20 cells total.
	total, valid := 0, 0
	for cell := range g.Iter(nil) {
		total++
		if cell.Valid() {
			valid++
		}
	}
	assert.Equal(t, valid, 12)
	assert.Equal(t, total, 20)
}

func TestGrid_Iter_BorderCells_NotInGrid(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(3, 4), geom.Sz(32.0, 32.0))
	for cell := range g.Iter(nil) {
		if !cell.Valid() {
			assert.Equal(t, cell.Get(), (*int)(nil))
		}
	}
}

func TestGrid_Iter_EarlyStop(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	count := 0
	for range g.Iter(nil) {
		count++
		if count == 3 {
			break
		}
	}
	assert.Equal(t, count, 3)
}

func TestGrid_Iter_Bounds(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
	// Viewport covering only cells (0,0)–(1,1) centers plus half-cell padding.
	viewport := geom.RectFromMinMax(geom.Pt(0.0, 0.0), geom.Pt(64.0, 64.0))
	valid := 0
	for cell := range g.Iter(&IterOptions{Bounds: viewport}) {
		if cell.Valid() {
			valid++
		}
	}
	// Exactly 4 valid cells fall within a 2×2 viewport.
	assert.Equal(t, valid, 4)
}

func TestGrid_Iter_Isometric_DrawOrder(t *testing.T) {
	g := NewIsometricRectGrid[int](geom.Sz(4, 4), geom.Sz(64.0, 32.0))
	// Painter's algorithm: col+row (depth) must be non-decreasing across the sequence.
	prevDepth := -1000
	for cell := range g.Iter(nil) {
		if !cell.Valid() {
			continue
		}
		x, y := cell.Index().XY()
		depth := x + y
		assert.True(t, depth >= prevDepth)
		prevDepth = depth
	}
}

func TestGrid_Iter_HexFlatTop_All(t *testing.T) {
	g := NewHexagonFlatTopGrid[int](geom.Sz(4, 4), geom.SzU(16.0))
	valid := 0
	for cell := range g.Iter(nil) {
		if cell.Valid() {
			valid++
		}
	}
	assert.Equal(t, valid, 16)
}

func TestGrid_Iter_HexPointyTop_All(t *testing.T) {
	g := NewHexagonPointyTopGrid[int](geom.Sz(4, 4), geom.SzU(16.0))
	valid := 0
	for cell := range g.Iter(nil) {
		if cell.Valid() {
			valid++
		}
	}
	assert.Equal(t, valid, 16)
}
