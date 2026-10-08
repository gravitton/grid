package grid

import (
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

func TestCell_Index(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.Cell(geom.Pt(1, 3)).Index(), geom.Pt(1, 3))
}

func TestCell_Size(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 16.0))

	geomtest.AssertSize(t, g.Cell(geom.Pt(1, 3)).Size(), g.CellSize())
}

func TestCell_Center(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(16.0, 16.0))
}

func TestCell_Bounds(t *testing.T) {
	awkward := testGrids(geom.Sz(6, 5), geom.Sz(8.48, 89.09), geom.Pt(-816.19, 732.64))

	t.Run("a square spans its column and row", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

		geomtest.AssertBox(t, g.Cell(geom.Pt(1, 2)).Bounds(), geom.Box[float64]{Min: geom.Pt(32.0, 64.0), Max: geom.Pt(64.0, 96.0)})
	})
	t.Run("is centered on the cell at its size", func(t *testing.T) {
		for name, g := range awkward {
			for cell := range g.All() {
				geomtest.AssertPoint(t, cell.Bounds().Center(), cell.Center(), name+" "+cell.Index().String()+": ")
				geomtest.AssertSize(t, cell.Bounds().Size(), cell.Size(), name+" "+cell.Index().String()+": ")
			}
		}
	})
	t.Run("neighboring squares meet in the same numbers", func(t *testing.T) {
		g := awkward["rectangular"]

		for cell := range g.All() {
			right := g.Cell(cell.Index().AddXY(1, 0)).Bounds()
			below := g.Cell(cell.Index().AddXY(0, 1)).Bounds()

			assert.Equal(t, right.Min.X, cell.Bounds().Max.X, cell.Index().String()+": ")
			assert.Equal(t, below.Min.Y, cell.Bounds().Max.Y, cell.Index().String()+": ")
		}
	})
	t.Run("the box around every cell is the bounds of the grid to the bit", func(t *testing.T) {
		for name, g := range awkward {
			around := g.Cell(geom.Pt(0, 0)).Bounds()
			for cell := range g.All() {
				around = around.Union(cell.Bounds())
			}

			assert.Equal(t, around, g.Bounds(), name+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		cell := awkward["flat-top"].Cell(geom.Pt(2, 3))
		allocs := testing.AllocsPerRun(100, func() {
			sinkBox = cell.Bounds()
		})

		assert.Equal(t, allocs, 0.0)
	})
}

func TestCell_Polygon(t *testing.T) {
	for name, g := range layoutGrids() {
		t.Run(name, func(t *testing.T) {
			cell := g.Cell(geom.Pt(2, 3))

			t.Run("is centered on the cell", func(t *testing.T) {
				geomtest.AssertPoint(t, cell.Polygon().Center, cell.Center())
			})
			t.Run("has a vertex for every side", func(t *testing.T) {
				assert.Equal(t, cell.Polygon().N, len(slices.Collect(cell.Polygon().Vertices())))
			})
			t.Run("fills the bounds of the cell", func(t *testing.T) {
				geomtest.AssertBox(t, cell.Polygon().Bounds(), cell.Bounds())
			})
		})
	}
}

func TestCell_Value(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	t.Run("points into the grid", func(t *testing.T) {
		*g.Cell(geom.Pt(1, 2)).Value() = 42

		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
	t.Run("is nil outside the grid", func(t *testing.T) {
		assert.Nil(t, g.Cell(geom.Pt(99, 99)).Value())
	})
}

func TestCell_Set(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	t.Run("stores the value", func(t *testing.T) {
		g.Cell(geom.Pt(1, 2)).Set(42)

		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
	t.Run("a write outside the grid does nothing", func(t *testing.T) {
		g.Cell(geom.Pt(99, 99)).Set(7)

		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
}

func TestCell_Neighbors(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		for _, index := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3), geom.Pt(-1, 0)} {
			assert.Equal(t, g.Cell(index).Neighbors(), g.Neighbors(index), name+" at "+index.String()+": ")
		}
	}
}

func TestCell_AppendNeighbors(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		assert.Equal(t, g.Cell(center).AppendNeighbors([]geom.Point[int]{geom.Pt(-9, -9)}), g.AppendNeighbors([]geom.Point[int]{geom.Pt(-9, -9)}, center), name+": ")
	}
}

func TestCell_Range(t *testing.T) {
	grids, center := rangeGrids()
	open := func(cell Cell[int]) bool {
		return cell.Index() != geom.Pt(5, 6)
	}

	for name, g := range grids {
		assert.True(t, g.Cell(center).Range(3, open).Equal(g.Range(center, 3, open)), name+": ")
	}
}

func TestCell_DistanceTo(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		assert.Equal(t, g.Cell(center).DistanceTo(geom.Pt(1, 9)), g.Distance(center, geom.Pt(1, 9)), name+": ")
	}
}

func TestCell_PathTo(t *testing.T) {
	grids, center := rangeGrids()
	open := func(cell Cell[int]) bool {
		return cell.Index() != geom.Pt(5, 6)
	}

	for name, g := range grids {
		assert.Equal(t, g.Cell(center).PathTo(geom.Pt(5, 9), open, nil), g.Path(center, geom.Pt(5, 9), open, nil), name+": ")
	}
}

func TestCell_IsValid(t *testing.T) {
	g := NewRect[int](geom.Sz(4, 4), geom.Sz(32.0, 32.0))

	t.Run("a cell of the grid is valid", func(t *testing.T) {
		assert.True(t, g.Cell(geom.Pt(0, 0)).IsValid())
		assert.True(t, g.Cell(geom.Pt(3, 3)).IsValid())
	})
	t.Run("a cell outside it is not", func(t *testing.T) {
		assert.False(t, g.Cell(geom.Pt(4, 0)).IsValid())
		assert.False(t, g.Cell(geom.Pt(0, -1)).IsValid())
	})
}

func TestCell_String(t *testing.T) {
	g := NewRect[int](geom.Sz(4, 4), geom.Sz(32.0, 32.0))

	t.Run("is the call that gives it", func(t *testing.T) {
		assert.Equal(t, g.Cell(geom.Pt(1, 3)).String(), "Cell(1,3)")
	})
	t.Run("a cell outside the grid prints the same way", func(t *testing.T) {
		assert.Equal(t, g.Cell(geom.Pt(-2, 9)).String(), "Cell(-2,9)")
	})
	t.Run("is Cell and its index as geom prints it", func(t *testing.T) {
		for cell := range g.All() {
			assert.Equal(t, cell.String(), "Cell"+cell.Index().String(), cell.Index().String()+": ")
		}
	})
	t.Run("allocates once", func(t *testing.T) {
		cell := g.Cell(geom.Pt(1, 3))

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkString = cell.String()
		}), 1.0)
	})
}

func BenchmarkCell_String(b *testing.B) {
	cell := NewRect[int](geom.Sz(4, 4), geom.Sz(32.0, 32.0)).Cell(geom.Pt(1, 3))

	for b.Loop() {
		sinkString = cell.String()
	}
}
