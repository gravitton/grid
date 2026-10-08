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
	grids := testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	t.Run("a square is a cell from the one before it", func(t *testing.T) {
		g := NewRect[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0))

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(32.0, 16.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(1, 0)).Center(), geom.Pt(96.0, 16.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 1)).Center(), geom.Pt(32.0, 48.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(2, 3)).Center(), geom.Pt(160.0, 112.0))
	})
	t.Run("a diamond steps down both ways", func(t *testing.T) {
		g := NewIsometric[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0))

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(96.0, 16.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(1, 0)).Center(), geom.Pt(128.0, 32.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 1)).Center(), geom.Pt(64.0, 32.0))
	})
	t.Run("the odd columns of flat-top hexagons are shifted down", func(t *testing.T) {
		g := NewHexFlatTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0))

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(10.0, geom.Sqrt3*5))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(1, 0)).Center(), geom.Pt(25.0, geom.Sqrt3*10))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 1)).Center(), geom.Pt(10.0, geom.Sqrt3*15))
	})
	t.Run("the even columns of flat-top hexagons are shifted down with EvenOffset", func(t *testing.T) {
		g := NewHexFlatTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0), func(options *HexOptions) {
			options.EvenOffset = true
		})

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(10.0, geom.Sqrt3*10))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(1, 0)).Center(), geom.Pt(25.0, geom.Sqrt3*5))
	})
	t.Run("the odd rows of pointy-top hexagons are shifted right", func(t *testing.T) {
		g := NewHexPointyTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0))

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(geom.Sqrt3*5, 10.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(1, 0)).Center(), geom.Pt(geom.Sqrt3*15, 10.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 1)).Center(), geom.Pt(geom.Sqrt3*10, 25.0))
	})
	t.Run("the even rows of pointy-top hexagons are shifted right with EvenOffset", func(t *testing.T) {
		g := NewHexPointyTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0), func(options *HexOptions) {
			options.EvenOffset = true
		})

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(geom.Sqrt3*10, 10.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 1)).Center(), geom.Pt(geom.Sqrt3*5, 25.0))
	})
	t.Run("rounds each product before it is added", func(t *testing.T) {
		corner, cellSize := geom.Pt(-816.19, 732.64), geom.Sz(8.48, 89.09)
		g := testGrids(geom.Sz(64, 64), cellSize, corner)["rectangular"]

		for index := range Arr[int](g.Size()).Keys() {
			center := geom.Pt(
				corner.X+float64(float64(2*index.X+1)*float64(cellSize.Width*0.5)),
				corner.Y+float64(float64(2*index.Y+1)*float64(cellSize.Height*0.5)),
			)

			assert.Equal(t, g.Cell(index).Center(), center, index.String()+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range grids {
			cell := g.Cell(geom.Pt(2, 3))
			allocs := testing.AllocsPerRun(100, func() {
				sinkCenter = cell.Center()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkCell_Center(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for i := range b.N {
			sinkCenter = g.Cell(geom.Pt(i&63, 7)).Center()
		}
	})
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

func BenchmarkCell_Bounds(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for i := range b.N {
			sinkBox = g.Cell(geom.Pt(i&63, 7)).Bounds()
		}
	})
}

func TestCell_Polygon(t *testing.T) {
	grids := testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	t.Run("a square and a diamond have four sides", func(t *testing.T) {
		assert.Equal(t, grids["rectangular"].Cell(geom.Pt(2, 3)).Polygon().N, 4)
		assert.Equal(t, grids["isometric"].Cell(geom.Pt(2, 3)).Polygon().N, 4)
	})
	t.Run("a hexagon has six sides", func(t *testing.T) {
		assert.Equal(t, grids["flat-top"].Cell(geom.Pt(2, 3)).Polygon().N, 6)
		assert.Equal(t, grids["pointy-top"].Cell(geom.Pt(2, 3)).Polygon().N, 6)
	})
	t.Run("is centered on the cell", func(t *testing.T) {
		for name, g := range grids {
			for index := range Arr[int](g.Size()).Keys() {
				geomtest.AssertPoint(t, g.Cell(index).Polygon().Center, g.Cell(index).Center(), name+" "+index.String()+": ")
			}
		}
	})
	t.Run("fills the bounds of the cell", func(t *testing.T) {
		for name, g := range grids {
			for index := range Arr[int](g.Size()).Keys() {
				geomtest.AssertBox(t, g.Cell(index).Polygon().Bounds(), g.Cell(index).Bounds(), name+" "+index.String()+": ")
			}
		}
	})
	t.Run("has a vertex for every side", func(t *testing.T) {
		for name, g := range grids {
			polygon := g.Cell(geom.Pt(2, 3)).Polygon()

			assert.Equal(t, polygon.N, len(slices.Collect(polygon.Vertices())), name+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range grids {
			cell := g.Cell(geom.Pt(4, 3))
			allocs := testing.AllocsPerRun(100, func() {
				sinkPolygon = cell.Polygon()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkCell_Polygon(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for i := range b.N {
			sinkPolygon = g.Cell(geom.Pt(i&63, 7)).Polygon()
		}
	})
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

func TestCell_AppendRange(t *testing.T) {
	grids, center := rangeGrids()
	open := func(cell Cell[int]) bool {
		return cell.Index() != geom.Pt(5, 6)
	}

	for name, g := range grids {
		assert.Equal(t, g.Cell(center).AppendRange([]geom.Point[int]{geom.Pt(-9, -9)}, 3, open), g.AppendRange([]geom.Point[int]{geom.Pt(-9, -9)}, center, 3, open), name+": ")
	}
}

func TestCell_RangeSeq(t *testing.T) {
	grids, center := rangeGrids()
	open := func(cell Cell[int]) bool {
		return cell.Index() != geom.Pt(5, 6)
	}

	for name, g := range grids {
		assert.Equal(t, slices.Collect(g.Cell(center).RangeSeq(3, open)), slices.Collect(g.RangeSeq(center, 3, open)), name+": ")
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
