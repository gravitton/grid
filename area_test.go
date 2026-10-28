package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

// regionGrids returns one grid of each neighbor rule, 9 by 9.
func regionGrids() map[string]*Grid[int] {
	return map[string]*Grid[int]{
		"cardinal":   NewRect[int](geom.Sz(9, 9), geom.Sz(32.0, 32.0)),
		"diagonal":   NewRect[int](geom.Sz(9, 9), geom.Sz(32.0, 32.0), withDiagonalMovement),
		"flat-top":   NewHexFlatTop[int](geom.Sz(9, 9), geom.Sz(32.0, 32.0)),
		"pointy-top": NewHexPointyTop[int](geom.Sz(9, 9), geom.Sz(32.0, 32.0)),
	}
}

func TestGrid_Border(t *testing.T) {
	center := geom.Pt(4, 4)

	t.Run("a single cell is its own border", func(t *testing.T) {
		for name, g := range regionGrids() {
			assert.True(t, g.Border(Reg([]geom.Point[int]{center})).Equal(Reg([]geom.Point[int]{center})), name+": ")
		}
	})
	t.Run("the border of a range is its outermost ring", func(t *testing.T) {
		for name, g := range regionGrids() {
			var ring []geom.Point[int]
			for index := range g.Range(center, 3, nil).All() {
				if g.Distance(center, index) == 3 {
					ring = append(ring, index)
				}
			}

			assert.True(t, g.Border(g.Range(center, 3, nil)).Equal(Reg(ring)), name+": ")
		}
	})
	t.Run("a hole has a border of its own", func(t *testing.T) {
		g := regionGrids()["cardinal"]
		ringed := g.Range(center, 2, nil).Difference(Reg([]geom.Point[int]{center}))

		assert.True(t, g.Border(ringed).Contains(geom.Pt(4, 3)))
		assert.False(t, g.Border(g.Range(center, 2, nil)).Contains(geom.Pt(4, 3)))
	})
	t.Run("the edge of the grid is a border", func(t *testing.T) {
		for name, g := range regionGrids() {
			var whole, edge []geom.Point[int]
			for cell := range g.All() {
				whole = append(whole, cell.Index())
				if index := cell.Index(); index.X == 0 || index.Y == 0 || index.X == 8 || index.Y == 8 {
					edge = append(edge, index)
				}
			}

			assert.True(t, g.Border(Reg(whole)).Equal(Reg(edge)), name+": ")
		}
	})
	t.Run("the empty region has none", func(t *testing.T) {
		assert.True(t, regionGrids()["cardinal"].Border(Region{}).IsEmpty())
	})
	t.Run("is held by the region", func(t *testing.T) {
		for name, g := range regionGrids() {
			region := g.Range(center, 3, nil)

			assert.True(t, region.Encloses(g.Border(region)), name+": ")
		}
	})
}

func TestGrid_Grow(t *testing.T) {
	center := geom.Pt(4, 4)
	single := Reg([]geom.Point[int]{center})

	t.Run("a cell grown is the range around it", func(t *testing.T) {
		for name, g := range regionGrids() {
			for n := range 4 {
				assert.True(t, g.Grow(single, n).Equal(g.Range(center, n, nil)), name+": ")
			}
		}
	})
	t.Run("zero or fewer steps is the region", func(t *testing.T) {
		g := regionGrids()["cardinal"]

		assert.True(t, g.Grow(single, 0).Equal(single))
		assert.True(t, g.Grow(single, -2).Equal(single))
	})
	t.Run("stops at the edge of the grid", func(t *testing.T) {
		for name, g := range regionGrids() {
			for index := range g.Grow(Reg([]geom.Point[int]{geom.Pt(0, 0)}), 3).All() {
				assert.True(t, g.Contains(index), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("fills the whole grid and no more", func(t *testing.T) {
		for name, g := range regionGrids() {
			assert.Equal(t, g.Grow(single, 100).Len(), 81, name+": ")
		}
	})
	t.Run("keeps an index of the region outside the grid", func(t *testing.T) {
		outside := Reg([]geom.Point[int]{geom.Pt(-5, -5), center})

		assert.True(t, regionGrids()["cardinal"].Grow(outside, 1).Contains(geom.Pt(-5, -5)))
	})
	t.Run("the empty region stays empty", func(t *testing.T) {
		assert.True(t, regionGrids()["cardinal"].Grow(Region{}, 3).IsEmpty())
	})
}

func TestGrid_Shrink(t *testing.T) {
	center := geom.Pt(4, 4)

	t.Run("undoes a grow away from the edge", func(t *testing.T) {
		for name, g := range regionGrids() {
			region := g.Range(center, 1, nil)

			assert.True(t, g.Shrink(g.Grow(region, 2), 2).Equal(region), name+": ")
		}
	})
	t.Run("a range shrunk is the smaller range", func(t *testing.T) {
		for name, g := range regionGrids() {
			assert.True(t, g.Shrink(g.Range(center, 3, nil), 2).Equal(g.Range(center, 1, nil)), name+": ")
		}
	})
	t.Run("zero or fewer steps is the region", func(t *testing.T) {
		g := regionGrids()["cardinal"]
		region := g.Range(center, 2, nil)

		assert.True(t, g.Shrink(region, 0).Equal(region))
		assert.True(t, g.Shrink(region, -1).Equal(region))
	})
	t.Run("more steps than the region is deep leaves nothing", func(t *testing.T) {
		g := regionGrids()["cardinal"]

		assert.True(t, g.Shrink(g.Range(center, 2, nil), 10).IsEmpty())
	})
}

func TestGrid_IsConnected(t *testing.T) {
	t.Run("the empty region and a single cell are", func(t *testing.T) {
		g := regionGrids()["cardinal"]

		assert.True(t, g.IsConnected(Region{}))
		assert.True(t, g.IsConnected(Reg([]geom.Point[int]{geom.Pt(4, 4)})))
	})
	t.Run("a range is", func(t *testing.T) {
		for name, g := range regionGrids() {
			assert.True(t, g.IsConnected(g.Range(geom.Pt(4, 4), 3, nil)), name+": ")
		}
	})
	t.Run("two cells apart are not", func(t *testing.T) {
		for name, g := range regionGrids() {
			assert.False(t, g.IsConnected(Reg([]geom.Point[int]{geom.Pt(1, 1), geom.Pt(6, 6)})), name+": ")
		}
	})
	t.Run("cells touching at a corner are connected under diagonal movement alone", func(t *testing.T) {
		corner := Reg([]geom.Point[int]{geom.Pt(1, 1), geom.Pt(2, 2)})

		assert.False(t, regionGrids()["cardinal"].IsConnected(corner))
		assert.True(t, regionGrids()["diagonal"].IsConnected(corner))
	})
}

func TestGrid_Components(t *testing.T) {
	t.Run("the empty region has none", func(t *testing.T) {
		assert.Length(t, regionGrids()["cardinal"].Components(Region{}), 0)
	})
	t.Run("a connected region is its one component", func(t *testing.T) {
		for name, g := range regionGrids() {
			region := g.Range(geom.Pt(4, 4), 2, nil)
			components := g.Components(region)

			assert.Length(t, components, 1, name+": ")
			assert.True(t, components[0].Equal(region), name+": ")
		}
	})
	t.Run("parts apart come back in the order of their first index", func(t *testing.T) {
		for name, g := range regionGrids() {
			left, right := g.Range(geom.Pt(1, 1), 1, nil), g.Range(geom.Pt(6, 6), 1, nil)
			components := g.Components(left.Union(right))

			assert.Length(t, components, 2, name+": ")
			assert.True(t, components[0].Equal(left), name+": ")
			assert.True(t, components[1].Equal(right), name+": ")
		}
	})
	t.Run("every component is connected and together they are the region", func(t *testing.T) {
		for name, g := range regionGrids() {
			var scattered []geom.Point[int]
			for cell := range g.All() {
				if index := cell.Index(); (index.X*5+index.Y*3)%4 != 0 {
					scattered = append(scattered, index)
				}
			}
			region := Reg(scattered)

			union := Region{}
			for _, component := range g.Components(region) {
				assert.True(t, g.IsConnected(component), name+": ")
				assert.False(t, union.Intersects(component), name+": ")
				union = union.Union(component)
			}

			assert.True(t, union.Equal(region), name+": ")
		}
	})
}
