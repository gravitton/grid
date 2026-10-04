package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

func TestGrid_DrawOrder(t *testing.T) {
	whole := geom.Box[float64]{}

	t.Run("a zero bounds walks the whole grid and one cell beyond it", func(t *testing.T) {
		g := NewRect[int](geom.Sz(3, 4), geom.Sz(32.0, 32.0))

		total, inside := 0, 0
		for cell := range g.DrawOrder(whole) {
			total++
			if cell.Valid() {
				inside++
			}
		}

		assert.Equal(t, inside, 12)
		assert.Equal(t, total, 20)
	})
	t.Run("a cell beyond the grid has no value", func(t *testing.T) {
		g := NewRect[int](geom.Sz(3, 4), geom.Sz(32.0, 32.0))

		for cell := range g.DrawOrder(whole) {
			if !cell.Valid() {
				assert.Nil(t, cell.Value())
			}
		}
	})
	t.Run("stops at the break", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

		count := 0
		for range g.DrawOrder(whole) {
			count++
			if count == 3 {
				break
			}
		}

		assert.Equal(t, count, 3)
	})
	t.Run("a bounds keeps to the cells in view", func(t *testing.T) {
		g := NewRect[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
		viewport := geom.BoxFromMinMax(geom.Pt(0.0, 0.0), geom.Pt(64.0, 64.0))

		inside := 0
		for cell := range g.DrawOrder(viewport) {
			if cell.Valid() {
				inside++
			}
		}

		assert.Equal(t, inside, 4)
	})
	t.Run("isometric depth never decreases", func(t *testing.T) {
		g := NewIsometric[int](geom.Sz(4, 4), geom.Sz(64.0, 32.0))

		previous := -1000
		for cell := range g.DrawOrder(whole) {
			if !cell.Valid() {
				continue
			}

			depth := cell.Index().X + cell.Index().Y
			assert.GreaterOrEqual(t, depth, previous)
			previous = depth
		}
	})
	t.Run("a hexagonal grid yields every cell", func(t *testing.T) {
		grids := map[string]*Grid[int]{
			"flat-top":   NewHexFlatTop[int](geom.Sz(4, 4), geom.SzU(16.0)),
			"pointy-top": NewHexPointyTop[int](geom.Sz(4, 4), geom.SzU(16.0)),
		}

		for name, g := range grids {
			inside := 0
			for cell := range g.DrawOrder(whole) {
				if cell.Valid() {
					inside++
				}
			}

			assert.Equal(t, inside, 16, name+": ")
		}
	})
}
