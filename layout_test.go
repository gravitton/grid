package grid

import (
	"fmt"
	"math"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

// testCellBox returns the box around the cell at the index, measured from its center.
func testCellBox(g *Grid[int], index geom.Point[int]) geom.Box[float64] {
	center := g.Cell(index).Center()
	half := g.CellSize().Scale(0.5).Vector()

	return geom.BoxFromMinMax(center.Add(half.Negate()), center.Add(half))
}

// testNearest returns the index whose center is nearest the point among the cells around the
// given one, in the measure that makes a cell of the named layout the points nearest its
// center, and whether it is nearer than every other by more than rounding.
func testNearest(name string, g *Grid[int], around geom.Point[int], point geom.Point[float64]) (geom.Point[int], bool) {
	half := g.CellSize().Scale(0.5)
	nearest, best, second := around, math.Inf(1), math.Inf(1)

	for index := range Arr[int](geom.Sz(7, 7)).Keys() {
		index = around.AddXY(index.X-3, index.Y-3)

		offset := point.Subtract(g.Cell(index).Center()).DivideXY(half.XY())
		x, y := math.Abs(offset.X), math.Abs(offset.Y)

		var distance float64
		switch name {
		case "rectangular":
			distance = max(x, y)
		case "isometric":
			distance = x + y
		case "flat-top", "flat-top even":
			distance = math.Hypot(x, y*geom.Sqrt3/2)
		default:
			distance = math.Hypot(x*geom.Sqrt3/2, y)
		}

		if distance < best {
			nearest, best, second = index, distance, best
		} else if distance < second {
			second = distance
		}
	}

	return nearest, second-best > 1e-6
}

func TestGrid_Bounds(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)

	t.Run("a rectangular grid spans its columns and rows of cells", func(t *testing.T) {
		g := NewRect[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0))

		geomtest.AssertBox(t, g.Bounds(), geom.Box[float64]{Min: geom.Pt(0.0, 0.0), Max: geom.Pt(256.0, 96.0)})
	})
	t.Run("an isometric grid spans its diamonds", func(t *testing.T) {
		g := NewIsometric[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0))

		geomtest.AssertBox(t, g.Bounds(), geom.Box[float64]{Min: geom.Pt(0.0, 0.0), Max: geom.Pt(224.0, 112.0)})
	})
	t.Run("is the box around every cell", func(t *testing.T) {
		sizes := []geom.Size[int]{geom.Sz(1, 1), geom.Sz(1, 4), geom.Sz(4, 1), geom.Sz(2, 2), geom.Sz(2, 3), geom.Sz(3, 2), geom.Sz(6, 5)}

		for _, size := range sizes {
			for name, g := range testGrids(size, geom.Sz(32.0, 20.0), origin) {
				around := testCellBox(g, geom.Pt(0, 0))
				for index := range Arr[int](size).Keys() {
					around = around.Union(testCellBox(g, index))
				}

				geomtest.AssertBox(t, g.Bounds(), around, name+" "+size.String()+": ")
			}
		}
	})
	t.Run("starts at the origin", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			assert.Equal(t, g.Bounds().Min, origin, name+": ")
		}
	})
	t.Run("a grid of no cells has an empty box at the origin", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(0, 0), geom.Sz(32.0, 20.0), origin) {
			assert.Equal(t, g.Bounds(), geom.Box[float64]{Min: origin, Max: origin}, name+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkBox = g.Bounds()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func TestGrid_CellSize(t *testing.T) {
	grids := testGrids(geom.Sz(6, 5), geom.Sz(10.0, 20.0), geom.Pt(0.0, 0.0))

	t.Run("a square is the size it was given", func(t *testing.T) {
		geomtest.AssertSize(t, grids["rectangular"].CellSize(), geom.Sz(10.0, 20.0))
	})
	t.Run("a diamond is the size it was given", func(t *testing.T) {
		geomtest.AssertSize(t, grids["isometric"].CellSize(), geom.Sz(10.0, 20.0))
	})
	t.Run("a flat-top hexagon is two radii wide", func(t *testing.T) {
		geomtest.AssertSize(t, grids["flat-top"].CellSize(), geom.Sz(20.0, geom.Sqrt3*20))
	})
	t.Run("a pointy-top hexagon is two radii tall", func(t *testing.T) {
		geomtest.AssertSize(t, grids["pointy-top"].CellSize(), geom.Sz(geom.Sqrt3*10, 40.0))
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0)) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkSize = g.CellSize()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkGrid_CellSize(b *testing.B) {
	g := NewHexFlatTop[int](geom.Sz(64, 64), geom.Sz(32.0, 20.0))

	for range b.N {
		sinkSize = g.CellSize()
	}
}

func TestGrid_CellSpacing(t *testing.T) {
	grids := testGrids(geom.Sz(6, 5), geom.Sz(10.0, 20.0), geom.Pt(0.0, 0.0))

	t.Run("squares do not interlock", func(t *testing.T) {
		geomtest.AssertSize(t, grids["rectangular"].CellSpacing(), geom.Sz(10.0, 20.0))
	})
	t.Run("diamonds interlock along both axes", func(t *testing.T) {
		geomtest.AssertSize(t, grids["isometric"].CellSpacing(), geom.Sz(5.0, 10.0))
	})
	t.Run("flat-top hexagons interlock along x", func(t *testing.T) {
		geomtest.AssertSize(t, grids["flat-top"].CellSpacing(), geom.Sz(15.0, geom.Sqrt3*20))
	})
	t.Run("pointy-top hexagons interlock along y", func(t *testing.T) {
		geomtest.AssertSize(t, grids["pointy-top"].CellSpacing(), geom.Sz(geom.Sqrt3*10, 30.0))
	})
	t.Run("neighbors along an axis are one spacing apart", func(t *testing.T) {
		for name, g := range grids {
			across := g.Cell(geom.Pt(3, 2)).Center().Subtract(g.Cell(geom.Pt(2, 2)).Center())
			down := g.Cell(geom.Pt(2, 3)).Center().Subtract(g.Cell(geom.Pt(2, 2)).Center())

			geomtest.AssertNumber(t, math.Abs(across.X), g.CellSpacing().Width, name+": ")
			geomtest.AssertNumber(t, math.Abs(down.Y), g.CellSpacing().Height, name+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0)) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkSize = g.CellSpacing()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkGrid_CellSpacing(b *testing.B) {
	g := NewHexFlatTop[int](geom.Sz(64, 64), geom.Sz(32.0, 20.0))

	for range b.N {
		sinkSize = g.CellSpacing()
	}
}

func TestGrid_CellWidth(t *testing.T) {
	for name, g := range layoutGrids() {
		assert.Equal(t, g.CellWidth(), g.CellSize().Width, name+": ")
	}
}

func TestGrid_CellHeight(t *testing.T) {
	for name, g := range layoutGrids() {
		assert.Equal(t, g.CellHeight(), g.CellSize().Height, name+": ")
	}
}

func TestGrid_IndexAt(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)
	grids := testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin)

	t.Run("a point inside a square is its cell", func(t *testing.T) {
		g := NewRect[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0))

		assert.Equal(t, g.IndexAt(geom.Pt(1.0, 1.0)), geom.Pt(0, 0))
		assert.Equal(t, g.IndexAt(geom.Pt(63.0, 31.0)), geom.Pt(0, 0))
		assert.Equal(t, g.IndexAt(geom.Pt(130.0, 70.0)), geom.Pt(2, 2))
	})
	t.Run("the corner of the bounds is the first square wherever the grid is", func(t *testing.T) {
		for _, corner := range []geom.Point[float64]{origin, geom.Pt(0.1, -7.3), geom.Pt(-816.19, 732.64), geom.Pt(1e6/3, -1e-3)} {
			for _, cellSize := range []geom.Size[float64]{geom.Sz(32.0, 20.0), geom.Sz(49.0, 82.0), geom.Sz(8.48, 89.09)} {
				g := testGrids(geom.Sz(6, 5), cellSize, corner)["rectangular"]

				assert.Equal(t, g.IndexAt(corner), geom.Pt(0, 0), corner.String()+" "+cellSize.String()+": ")
			}
		}
	})
	t.Run("an edge shared by two squares belongs to the one after it", func(t *testing.T) {
		corner := geom.Pt(7.0, -448.0)
		g := testGrids(geom.Sz(6, 5), geom.Sz(49.0, 82.0), corner)["rectangular"]

		for index := range Arr[int](geom.Sz(9, 9)).Keys() {
			index = index.AddXY(-4, -4)
			edge := corner.AddXY(float64(index.X)*49, float64(index.Y)*82)

			assert.Equal(t, g.IndexAt(edge), index, index.String()+": ")
		}
	})
	t.Run("a vertex of a diamond goes to the same side on every diamond", func(t *testing.T) {
		g := testGrids(geom.Sz(6, 5), geom.Sz(82.0, 49.0), geom.Pt(7.0, -448.0))["isometric"]
		vertices := map[string]geom.Vector[float64]{
			"top":    geom.Vec(0.0, -24.5),
			"right":  geom.Vec(41.0, 0.0),
			"bottom": geom.Vec(0.0, 24.5),
			"left":   geom.Vec(-41.0, 0.0),
		}
		owners := map[string]geom.Vector[int]{
			"top":    geom.Vec(0, 0),
			"right":  geom.Vec(1, 0),
			"bottom": geom.Vec(1, 1),
			"left":   geom.Vec(0, 1),
		}

		for name, vertex := range vertices {
			for index := range Arr[int](geom.Sz(9, 9)).Keys() {
				index = index.AddXY(-4, -4)

				assert.Equal(t, g.IndexAt(g.Cell(index).Center().Add(vertex)), index.Add(owners[name]), name+" of "+index.String()+": ")
			}
		}
	})
	t.Run("a point outside the grid gives an index outside it", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

		assert.False(t, g.Contains(g.IndexAt(geom.Pt(-10.0, 40.0))))
		assert.False(t, g.Contains(g.IndexAt(geom.Pt(40.0, 500.0))))
	})
	t.Run("the center of a cell gives its index, beyond the grid too", func(t *testing.T) {
		for name, g := range grids {
			for index := range Arr[int](geom.Sz(12, 12)).Keys() {
				index = index.AddXY(-4, -4)

				assert.Equal(t, g.IndexAt(g.Cell(index).Center()), index, name+" "+index.String()+": ")
			}
		}
	})
	t.Run("round-trips on cells too small to tell from nothing", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(6, 5), geom.Sz(1e-7, 3e-9), geom.Pt(0.0, 0.0)) {
			for index := range Arr[int](g.Size()).Keys() {
				assert.Equal(t, g.IndexAt(g.Cell(index).Center()), index, name+" "+index.String()+": ")
			}
		}
	})
	t.Run("a point is in the cell whose center is nearest", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(6, 5), geom.Sz(8.48, 89.09), geom.Pt(-816.19, 732.64)) {
			around := geom.Pt(2, 3)
			center := g.Cell(around).Center()

			for step := range Arr[int](geom.Sz(41, 41)).Keys() {
				point := center.Add(g.CellSize().Vector().MultiplyXY(float64(step.X-20)/13, float64(step.Y-20)/13))

				if nearest, clear := testNearest(name, g, around, point); clear {
					assert.Equal(t, g.IndexAt(point), nearest, name+" "+point.String()+": ")
				}
			}
		}
	})
	t.Run("a point just inside a vertex is the cell of the polygon", func(t *testing.T) {
		for name, g := range grids {
			for index := range Arr[int](g.Size()).Keys() {
				center := g.Cell(index).Center()
				for vertex := range g.Cell(index).Polygon().Vertices() {
					assert.Equal(t, g.IndexAt(center.Lerp(vertex, 0.99)), index, name+" "+index.String()+": ")
				}
			}
		}
	})
	for name, point := range map[string]geom.Point[float64]{
		"NaN":      geom.Pt(math.NaN(), 50.0),
		"infinite": geom.Pt(100.0, math.Inf(1)),
	} {
		t.Run("panics for a "+name+" point on every layout, as geom.Cast does", func(t *testing.T) {
			for _, g := range grids {
				assert.PanicsWith(t, func() {
					g.IndexAt(point)
				}, "geom: cast of a non-finite value to an integer")
			}
		})
	}
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range grids {
			allocs := testing.AllocsPerRun(100, func() {
				sinkPoint = g.IndexAt(geom.Pt(170.0, 95.0))
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkGrid_IndexAt(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for i := range b.N {
			sinkPoint = g.IndexAt(geom.Pt(float64(i&1023), 77.0))
		}
	})
}

func FuzzGrid_IndexAt(f *testing.F) {
	f.Add(0, 0, 64.0, 32.0, 0.0, 0.0, 0.0, 0.0)
	f.Add(3, -2, 20.0, 20.0, 100.0, 100.0, 0.3, -0.8)
	f.Add(-7, 11, 0.5, 300.0, -1000.0, 250.5, -1.0, 0.49)

	f.Fuzz(func(t *testing.T, column, row int, width, height, x, y, across, down float64) {
		if max(column, -column, row, -row) > 1000 {
			t.Skip()
		}
		if !(width >= 0.01 && width <= 1000 && height >= 0.01 && height <= 1000) {
			t.Skip()
		}
		if !(math.Abs(x) <= 1e6 && math.Abs(y) <= 1e6) {
			t.Skip()
		}

		if !(math.Abs(across) <= 1 && math.Abs(down) <= 1) {
			t.Skip()
		}

		index := geom.Pt(column, row)
		for name, g := range testGrids(geom.Sz(6, 5), geom.Sz(width, height), geom.Pt(x, y)) {
			assert.Equal(t, g.IndexAt(g.Cell(index).Center()), index, name+": ")

			point := g.Cell(index).Center().Add(g.CellSize().Vector().MultiplyXY(across, down))
			if nearest, clear := testNearest(name, g, index, point); clear {
				assert.Equal(t, g.IndexAt(point), nearest, name+" nearest: ")
			}
		}

		assert.Equal(t, NewRect[int](geom.Sz(6, 5), geom.Sz(width, height), func(options *RectOptions) {
			options.Origin = geom.Pt(x, y)
		}).IndexAt(geom.Pt(x, y)), geom.Pt(0, 0), "the corner of the bounds: ")

		corner := geom.Pt(math.Trunc(x), math.Trunc(y))
		whole := geom.Sz(math.Ceil(width), math.Ceil(height))
		grids := testGrids(geom.Sz(6, 5), whole, corner)
		edge := corner.AddXY(float64(column)*whole.Width, float64(row)*whole.Height)
		top := grids["isometric"].Cell(index).Center().AddXY(0, -whole.Height/2)

		assert.Equal(t, grids["rectangular"].IndexAt(edge), index, "an edge of a square: ")
		assert.Equal(t, grids["isometric"].IndexAt(top), index, "the top of a diamond: ")
	})
}

func ExampleGrid_IndexAt() {
	g := NewRect[int](geom.Sz(100, 100), RectCellSize(32))

	cell := g.CellAt(geom.Pt(499.0, 123.4))

	fmt.Println(g.IndexAt(geom.Pt(499.0, 123.4)))
	fmt.Println(cell, cell.Center(), cell.Bounds())
	fmt.Println(g.Contains(g.IndexAt(geom.Pt(-5.0, 20.0))))
	// Output:
	// (15,3)
	// Cell(15,3) (496.00,112.00) (480.00,96.00)-(512.00,128.00)
	// false
}

func TestGrid_CellAt(t *testing.T) {
	t.Run("is the cell of IndexAt", func(t *testing.T) {
		for name, g := range layoutGrids() {
			for _, point := range []geom.Point[float64]{geom.Pt(64.0, 96.0), geom.Pt(10.5, 3.25), geom.Pt(-40.0, 900.0)} {
				assert.Equal(t, g.CellAt(point).Index(), g.IndexAt(point), name+" at "+point.String()+": ")
			}
		}
	})
	t.Run("writes through to the grid", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
		g.CellAt(geom.Pt(64.0, 96.0)).Set(7)

		assert.Equal(t, *g.Cell(geom.Pt(2, 3)).Value(), 7)
	})
}

func TestGrid_MoveTo(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)
	start := geom.Pt(7.0, 9.0)
	moved := func() (map[string]*Grid[int], map[string]*Grid[int]) {
		before := testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), start)
		after := testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), start)
		for _, g := range after {
			g.MoveTo(origin)
		}

		return before, after
	}

	t.Run("places the corner of the bounds at the origin", func(t *testing.T) {
		_, after := moved()

		for name, g := range after {
			assert.Equal(t, g.Bounds().Min, origin, name+": ")
		}
	})
	t.Run("keeps the extent", func(t *testing.T) {
		before, after := moved()

		for name, g := range after {
			geomtest.AssertSize(t, g.Bounds().Size(), before[name].Bounds().Size(), name+": ")
		}
	})
	t.Run("shifts every cell by what the corner moved", func(t *testing.T) {
		before, after := moved()

		for name, g := range after {
			for cell := range g.All() {
				geomtest.AssertPoint(t, cell.Center(), before[name].Cell(cell.Index()).Center().Add(origin.Subtract(start)), name+" "+cell.Index().String()+": ")
			}
		}
	})
	t.Run("is the grid built at that origin", func(t *testing.T) {
		_, after := moved()

		for name, built := range testGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			assert.Equal(t, after[name].Bounds(), built.Bounds(), name+": ")
			assert.Equal(t, after[name].Cell(geom.Pt(4, 3)).Center(), built.Cell(geom.Pt(4, 3)).Center(), name+": ")
		}
	})
	t.Run("changes nothing applied twice", func(t *testing.T) {
		_, after := moved()

		for name, g := range after {
			center := g.Cell(geom.Pt(2, 3)).Center()
			g.MoveTo(origin)

			assert.Equal(t, g.Cell(geom.Pt(2, 3)).Center(), center, name+": ")
		}
	})
	t.Run("keeps the values", func(t *testing.T) {
		g := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
		g.Fill(5)
		g.MoveTo(origin)

		assert.Equal(t, *g.CellAt(geom.Pt(116.0, 66.0)).Value(), 5)
	})
	t.Run("leaves a clone where it is", func(t *testing.T) {
		g := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
		clone := g.Clone()
		g.MoveTo(origin)

		assert.Equal(t, clone.Bounds().Min, geom.Pt(0.0, 0.0))
	})
	t.Run("places a grid of no cells", func(t *testing.T) {
		for name, g := range testGrids(geom.Sz(0, 0), geom.Sz(32.0, 20.0), start) {
			g.MoveTo(origin)

			assert.Equal(t, g.Bounds(), geom.Box[float64]{Min: origin, Max: origin}, name+": ")
		}
	})
	for name, corner := range map[string]geom.Point[float64]{
		"NaN":      geom.Pt(math.NaN(), 50.0),
		"infinite": geom.Pt(100.0, math.Inf(-1)),
	} {
		t.Run("panics for a "+name+" origin", func(t *testing.T) {
			before, _ := moved()

			for _, g := range before {
				assert.PanicsWith(t, func() {
					g.MoveTo(corner)
				}, "grid: grid cannot be laid out at origin "+corner.String())
			}
		})
	}
	t.Run("panics for an origin the far corner is past the largest number at", func(t *testing.T) {
		far := geom.Pt(math.MaxFloat64, 50.0)

		for _, g := range testGrids(geom.Sz(6, 5), geom.Sz(1e300, 1e300), start) {
			assert.PanicsWith(t, func() {
				g.MoveTo(far)
			}, "grid: grid cannot be laid out at origin "+far.String())
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		before, _ := moved()

		for name, g := range before {
			allocs := testing.AllocsPerRun(100, func() {
				g.MoveTo(origin)
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkGrid_MoveTo(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for i := range b.N {
			g.MoveTo(geom.Pt(float64(i&7), 1.0))
		}
	})
}

func ExampleGrid_MoveTo() {
	g := NewRect[int](geom.Sz(10, 10), RectCellSize(32))

	g.MoveTo(geom.Pt(100.0, 50.0))
	fmt.Println(g.Bounds(), g.Cell(geom.Pt(0, 0)).Center())

	g.Resize(RectCellSize(48))
	fmt.Println(g.Bounds(), g.Cell(geom.Pt(0, 0)).Center())
	// Output:
	// (100.00,50.00)-(420.00,370.00) (116.00,66.00)
	// (100.00,50.00)-(580.00,530.00) (124.00,74.00)
}

func TestGrid_Resize(t *testing.T) {
	origin := geom.Pt(10.0, 20.0)
	scale := geom.Sz(20.0, 12.0)
	grids := func() map[string]*Grid[int] {
		return testGrids(geom.Sz(6, 5), geom.Sz(64.0, 32.0), origin)
	}

	t.Run("replaces the size of a square", func(t *testing.T) {
		g := grids()["rectangular"]
		g.Resize(geom.Sz(128.0, 64.0))

		geomtest.AssertSize(t, g.CellSize(), geom.Sz(128.0, 64.0))
		geomtest.AssertSize(t, g.CellSpacing(), geom.Sz(128.0, 64.0))
	})
	t.Run("replaces the circumradius of a hexagon", func(t *testing.T) {
		g := grids()["flat-top"]
		g.Resize(geom.Sz(10.0, 10.0))

		geomtest.AssertSize(t, g.CellSize(), geom.Sz(20.0, geom.Sqrt3*10))
	})
	t.Run("is the grid built at that size", func(t *testing.T) {
		built := testGrids(geom.Sz(6, 5), scale, origin)

		for name, g := range grids() {
			g.Resize(scale)

			assert.Equal(t, g.Bounds(), built[name].Bounds(), name+": ")
			assert.Equal(t, g.CellSize(), built[name].CellSize(), name+": ")
			assert.Equal(t, g.CellSpacing(), built[name].CellSpacing(), name+": ")
			assert.Equal(t, g.Cell(geom.Pt(4, 3)).Center(), built[name].Cell(geom.Pt(4, 3)).Center(), name+": ")
			assert.Equal(t, g.Cell(geom.Pt(4, 3)).Polygon(), built[name].Cell(geom.Pt(4, 3)).Polygon(), name+": ")
		}
	})
	t.Run("keeps the origin", func(t *testing.T) {
		for name, g := range grids() {
			g.Resize(scale)

			assert.Equal(t, g.Bounds().Min, origin, name+": ")
		}
	})
	t.Run("keeps the values", func(t *testing.T) {
		g := grids()["rectangular"]
		g.Fill(5)
		g.Resize(scale)

		assert.Equal(t, *g.Cell(geom.Pt(2, 3)).Value(), 5)
	})
	t.Run("leaves a clone at its size", func(t *testing.T) {
		g := grids()["rectangular"]
		clone := g.Clone()
		g.Resize(scale)

		geomtest.AssertSize(t, clone.CellSize(), geom.Sz(64.0, 32.0))
	})
	t.Run("panics for a size too small to halve", func(t *testing.T) {
		smallest := geom.Sz(math.SmallestNonzeroFloat64, 32.0)

		assert.PanicsWith(t, func() {
			grids()["rectangular"].Resize(smallest)
		}, "grid: cell of size "+smallest.String()+" cannot be laid out")
	})
	for name, bad := range map[string]geom.Size[float64]{
		"zero":      geom.Sz(0.0, 32.0),
		"negative":  geom.Sz(32.0, -32.0),
		"NaN":       geom.Sz(math.NaN(), 32.0),
		"infinite":  geom.Sz(32.0, math.Inf(1)),
		"too large": geom.Sz(math.MaxFloat64, 32.0),
	} {
		t.Run("panics for a "+name+" size", func(t *testing.T) {
			for _, g := range grids() {
				assert.PanicsWith(t, func() {
					g.Resize(bad)
				}, "grid: cell of size "+bad.String()+" cannot be laid out")
			}
		})
	}
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range grids() {
			allocs := testing.AllocsPerRun(100, func() {
				g.Resize(scale)
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkGrid_Resize(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for i := range b.N {
			g.Resize(geom.Sz(float64(i&7+8), 20.0))
		}
	})
}
