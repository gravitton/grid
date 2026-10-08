package grid

import (
	"maps"
	"math"
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

// testLayouts returns the layout of each of the testGrids.
func testLayouts(size geom.Size[int], cellSize geom.Size[float64], origin geom.Point[float64]) map[string]Layout {
	layouts := map[string]Layout{}
	for name, g := range testGrids(size, cellSize, origin) {
		layouts[name] = g.Layout()
	}

	return layouts
}

// testCellBox returns the box around the cell at the index, measured from its center.
func testCellBox(l Layout, index geom.Point[int]) geom.Box[float64] {
	center := l.ToPoint(index)
	half := l.CellSize().Scale(0.5).Vector()

	return geom.BoxFromMinMax(center.Add(half.Negate()), center.Add(half))
}

// testNearest returns the index whose center is nearest the point among the cells around the
// given one, in the measure that makes a cell of the named layout the points nearest its
// center, and whether it is nearer than every other by more than rounding.
func testNearest(name string, l Layout, around geom.Point[int], point geom.Point[float64]) (geom.Point[int], bool) {
	half := l.CellSize().Scale(0.5)
	nearest, best, second := around, math.Inf(1), math.Inf(1)

	for index := range Arr[int](geom.Sz(7, 7)).Keys() {
		index = around.AddXY(index.X-3, index.Y-3)

		offset := point.Subtract(l.ToPoint(index)).DivideXY(half.XY())
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

func TestLayout_Size(t *testing.T) {
	l := NewRect[int](geom.Sz(6, 5), geom.Sz(64.0, 32.0)).Layout()

	assert.Equal(t, l.Size(), geom.Sz(6, 5))
}

func TestLayout_Bounds(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)

	t.Run("a rectangular grid spans its columns and rows of cells", func(t *testing.T) {
		l := NewRect[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0)).Layout()

		geomtest.AssertBox(t, l.Bounds(), geom.Box[float64]{Min: geom.Pt(0.0, 0.0), Max: geom.Pt(256.0, 96.0)})
	})
	t.Run("an isometric grid spans its diamonds", func(t *testing.T) {
		l := NewIsometric[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0)).Layout()

		geomtest.AssertBox(t, l.Bounds(), geom.Box[float64]{Min: geom.Pt(0.0, 0.0), Max: geom.Pt(224.0, 112.0)})
	})
	t.Run("is the box around every cell", func(t *testing.T) {
		sizes := []geom.Size[int]{geom.Sz(1, 1), geom.Sz(1, 4), geom.Sz(4, 1), geom.Sz(2, 2), geom.Sz(2, 3), geom.Sz(3, 2), geom.Sz(6, 5)}

		for _, size := range sizes {
			for name, l := range testLayouts(size, geom.Sz(32.0, 20.0), origin) {
				around := testCellBox(l, geom.Pt(0, 0))
				for index := range Arr[int](size).Keys() {
					around = around.Union(testCellBox(l, index))
				}

				geomtest.AssertBox(t, l.Bounds(), around, name+" "+size.String()+": ")
			}
		}
	})
	t.Run("starts at the origin", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			assert.Equal(t, l.Bounds().Min, origin, name+": ")
		}
	})
	t.Run("a grid of no cells has an empty box at the origin", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(0, 0), geom.Sz(32.0, 20.0), origin) {
			assert.Equal(t, l.Bounds(), geom.Box[float64]{Min: origin, Max: origin}, name+": ")
		}
	})
	t.Run("the zero layout has none", func(t *testing.T) {
		var l Layout

		assert.True(t, l.Bounds().IsZero())
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkBox = l.Bounds()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func TestLayout_CellSize(t *testing.T) {
	layouts := testLayouts(geom.Sz(6, 5), geom.Sz(10.0, 20.0), geom.Pt(0.0, 0.0))

	t.Run("a square is the size it was given", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["rectangular"].CellSize(), geom.Sz(10.0, 20.0))
	})
	t.Run("a diamond is the size it was given", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["isometric"].CellSize(), geom.Sz(10.0, 20.0))
	})
	t.Run("a flat-top hexagon is two radii wide", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["flat-top"].CellSize(), geom.Sz(20.0, geom.Sqrt3*20))
	})
	t.Run("a pointy-top hexagon is two radii tall", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["pointy-top"].CellSize(), geom.Sz(geom.Sqrt3*10, 40.0))
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0)) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkSize = l.CellSize()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkLayout_CellSize(b *testing.B) {
	l := NewHexFlatTop[int](geom.Sz(64, 64), geom.Sz(32.0, 20.0)).Layout()

	for range b.N {
		sinkSize = l.CellSize()
	}
}

func TestLayout_CellSpacing(t *testing.T) {
	layouts := testLayouts(geom.Sz(6, 5), geom.Sz(10.0, 20.0), geom.Pt(0.0, 0.0))

	t.Run("squares do not interlock", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["rectangular"].CellSpacing(), geom.Sz(10.0, 20.0))
	})
	t.Run("diamonds interlock along both axes", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["isometric"].CellSpacing(), geom.Sz(5.0, 10.0))
	})
	t.Run("flat-top hexagons interlock along x", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["flat-top"].CellSpacing(), geom.Sz(15.0, geom.Sqrt3*20))
	})
	t.Run("pointy-top hexagons interlock along y", func(t *testing.T) {
		geomtest.AssertSize(t, layouts["pointy-top"].CellSpacing(), geom.Sz(geom.Sqrt3*10, 30.0))
	})
	t.Run("neighbors along an axis are one spacing apart", func(t *testing.T) {
		for name, l := range layouts {
			across := l.ToPoint(geom.Pt(3, 2)).Subtract(l.ToPoint(geom.Pt(2, 2)))
			down := l.ToPoint(geom.Pt(2, 3)).Subtract(l.ToPoint(geom.Pt(2, 2)))

			geomtest.AssertNumber(t, math.Abs(across.X), l.CellSpacing().Width, name+": ")
			geomtest.AssertNumber(t, math.Abs(down.Y), l.CellSpacing().Height, name+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0)) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkSize = l.CellSpacing()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkLayout_CellSpacing(b *testing.B) {
	l := NewHexFlatTop[int](geom.Sz(64, 64), geom.Sz(32.0, 20.0)).Layout()

	for range b.N {
		sinkSize = l.CellSpacing()
	}
}

func TestLayout_Translate(t *testing.T) {
	change := geom.Vec(30.0, -12.0)
	layouts := testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	t.Run("shifts the bounds", func(t *testing.T) {
		for name, l := range layouts {
			geomtest.AssertBox(t, l.Translate(change).Bounds(), l.Bounds().Translate(change), name+": ")
		}
	})
	t.Run("shifts every cell", func(t *testing.T) {
		for name, l := range layouts {
			moved := l.Translate(change)
			for index := range Arr[int](l.Size()).Keys() {
				geomtest.AssertPoint(t, moved.ToPoint(index), l.ToPoint(index).Add(change), name+" "+index.String()+": ")
			}
		}
	})
	t.Run("panics for a shift that is not finite", func(t *testing.T) {
		for name, l := range layouts {
			assert.Panics(t, func() {
				l.Translate(geom.Vec(math.Inf(1), 0.0))
			}, name+": ")
		}
	})
	t.Run("shifts the zero layout", func(t *testing.T) {
		var l Layout

		assert.Equal(t, l.Translate(change).Bounds(), geom.Box[float64]{Min: change.Point(), Max: change.Point()})
		geomtest.AssertPoint(t, l.Translate(change).ToPoint(geom.Pt(3, 2)), change.Point())
	})
}

func TestLayout_MoveTo(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)
	layouts := testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(7.0, 9.0))

	t.Run("places the corner of the bounds at the origin", func(t *testing.T) {
		for name, l := range layouts {
			assert.Equal(t, l.MoveTo(origin).Bounds().Min, origin, name+": ")
		}
	})
	t.Run("keeps the extent", func(t *testing.T) {
		for name, l := range layouts {
			geomtest.AssertSize(t, l.MoveTo(origin).Bounds().Size(), l.Bounds().Size(), name+": ")
		}
	})
	t.Run("changes nothing applied twice", func(t *testing.T) {
		for name, l := range layouts {
			moved := l.MoveTo(origin)

			geomtest.AssertPoint(t, moved.MoveTo(origin).ToPoint(geom.Pt(2, 3)), moved.ToPoint(geom.Pt(2, 3)), name+": ")
		}
	})
	t.Run("leaves the layout it was called on", func(t *testing.T) {
		for name, l := range layouts {
			l.MoveTo(origin)

			assert.Equal(t, l.Bounds().Min, geom.Pt(7.0, 9.0), name+": ")
		}
	})
	for name, corner := range map[string]geom.Point[float64]{
		"NaN":      geom.Pt(math.NaN(), 50.0),
		"infinite": geom.Pt(100.0, math.Inf(-1)),
	} {
		t.Run("panics for a "+name+" origin", func(t *testing.T) {
			for _, l := range layouts {
				assert.PanicsWith(t, func() {
					l.MoveTo(corner)
				}, "grid: origin "+corner.String()+" is not finite")
			}
		})
	}
	t.Run("places the zero layout", func(t *testing.T) {
		var l Layout

		assert.Equal(t, l.MoveTo(origin).Bounds(), geom.Box[float64]{Min: origin, Max: origin})
		geomtest.AssertPoint(t, l.MoveTo(origin).ToPoint(geom.Pt(3, 2)), origin)
	})
}

func TestLayout_Resize(t *testing.T) {
	origin := geom.Pt(10.0, 20.0)
	layouts := testLayouts(geom.Sz(6, 5), geom.Sz(64.0, 32.0), origin)

	t.Run("replaces the size of a square", func(t *testing.T) {
		l := layouts["rectangular"].Resize(geom.Sz(128.0, 64.0))

		geomtest.AssertSize(t, l.CellSize(), geom.Sz(128.0, 64.0))
		geomtest.AssertSize(t, l.CellSpacing(), geom.Sz(128.0, 64.0))
	})
	t.Run("replaces the circumradius of a hexagon", func(t *testing.T) {
		l := layouts["flat-top"].Resize(geom.Sz(10.0, 10.0))

		geomtest.AssertSize(t, l.CellSize(), geom.Sz(20.0, geom.Sqrt3*10))
	})
	t.Run("is the layout built at that size", func(t *testing.T) {
		built := testLayouts(geom.Sz(6, 5), geom.Sz(20.0, 12.0), origin)

		for name, l := range layouts {
			resized := l.Resize(geom.Sz(20.0, 12.0))

			geomtest.AssertBox(t, resized.Bounds(), built[name].Bounds(), name+": ")
			geomtest.AssertPoint(t, resized.ToPoint(geom.Pt(4, 3)), built[name].ToPoint(geom.Pt(4, 3)), name+": ")
		}
	})
	t.Run("keeps the origin", func(t *testing.T) {
		for name, l := range layouts {
			assert.Equal(t, l.Resize(geom.Sz(20.0, 12.0)).Bounds().Min, origin, name+": ")
		}
	})
	t.Run("panics for a size too small to halve", func(t *testing.T) {
		smallest := geom.Sz(math.SmallestNonzeroFloat64, 32.0)

		assert.PanicsWith(t, func() {
			layouts["rectangular"].Resize(smallest)
		}, "grid: cell of size "+smallest.String()+" cannot be laid out")
	})
	t.Run("sizes the zero layout", func(t *testing.T) {
		var l Layout

		geomtest.AssertSize(t, l.Resize(geom.Sz(20.0, 12.0)).CellSize(), geom.Sz(20.0, 12.0))
	})
	for name, scale := range map[string]geom.Size[float64]{
		"zero":     geom.Sz(0.0, 32.0),
		"negative": geom.Sz(32.0, -32.0),
		"NaN":      geom.Sz(math.NaN(), 32.0),
		"infinite": geom.Sz(32.0, math.Inf(1)),
	} {
		t.Run("panics for a "+name+" size", func(t *testing.T) {
			for _, l := range layouts {
				assert.PanicsWith(t, func() {
					l.Resize(scale)
				}, "grid: cell of size "+scale.String()+" cannot be laid out")
			}
		})
	}
}

func TestLayout_ToPoint(t *testing.T) {
	origin := geom.Pt(0.0, 0.0)

	t.Run("a square is a cell from the one before it", func(t *testing.T) {
		l := NewRect[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0)).Layout()

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(32.0, 16.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 0)), geom.Pt(96.0, 16.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 1)), geom.Pt(32.0, 48.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(2, 3)), geom.Pt(160.0, 112.0))
	})
	t.Run("a diamond steps down both ways", func(t *testing.T) {
		l := NewIsometric[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0)).Layout()

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(96.0, 16.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 0)), geom.Pt(128.0, 32.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 1)), geom.Pt(64.0, 32.0))
	})
	t.Run("the odd columns of flat-top hexagons are shifted down", func(t *testing.T) {
		l := NewHexFlatTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0)).Layout()

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(10.0, geom.Sqrt3*5))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 0)), geom.Pt(25.0, geom.Sqrt3*10))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 1)), geom.Pt(10.0, geom.Sqrt3*15))
	})
	t.Run("the even columns of flat-top hexagons are shifted down with EvenOffset", func(t *testing.T) {
		l := NewHexFlatTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0), func(options *HexOptions) {
			options.EvenOffset = true
		}).Layout()

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(10.0, geom.Sqrt3*10))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 0)), geom.Pt(25.0, geom.Sqrt3*5))
	})
	t.Run("the odd rows of pointy-top hexagons are shifted right", func(t *testing.T) {
		l := NewHexPointyTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0)).Layout()

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(geom.Sqrt3*5, 10.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 0)), geom.Pt(geom.Sqrt3*15, 10.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 1)), geom.Pt(geom.Sqrt3*10, 25.0))
	})
	t.Run("the even rows of pointy-top hexagons are shifted right with EvenOffset", func(t *testing.T) {
		l := NewHexPointyTop[int](geom.Sz(4, 3), geom.Sz(10.0, 10.0), func(options *HexOptions) {
			options.EvenOffset = true
		}).Layout()

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(geom.Sqrt3*10, 10.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 1)), geom.Pt(geom.Sqrt3*5, 25.0))
	})
	t.Run("the zero layout has every cell at the world origin", func(t *testing.T) {
		var l Layout

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(3, 2)), origin)
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkCenter = l.ToPoint(geom.Pt(4, 3))
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkLayout_ToPoint(b *testing.B) {
	layouts := testLayouts(geom.Sz(64, 64), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	for _, name := range slices.Sorted(maps.Keys(layouts)) {
		l := layouts[name]

		b.Run(name, func(b *testing.B) {
			for i := range b.N {
				sinkCenter = l.ToPoint(geom.Pt(i&63, 7))
			}
		})
	}
}

func TestLayout_FromPoint(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)
	layouts := testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin)

	t.Run("a point inside a square is its cell", func(t *testing.T) {
		l := NewRect[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0)).Layout()

		assert.Equal(t, l.FromPoint(geom.Pt(1.0, 1.0)), geom.Pt(0, 0))
		assert.Equal(t, l.FromPoint(geom.Pt(63.0, 31.0)), geom.Pt(0, 0))
		assert.Equal(t, l.FromPoint(geom.Pt(130.0, 70.0)), geom.Pt(2, 2))
	})
	t.Run("the corner of the bounds is the first square wherever the grid is", func(t *testing.T) {
		for _, corner := range []geom.Point[float64]{origin, geom.Pt(0.1, -7.3), geom.Pt(-816.19, 732.64), geom.Pt(1e6/3, -1e-3)} {
			for _, cellSize := range []geom.Size[float64]{geom.Sz(32.0, 20.0), geom.Sz(49.0, 82.0), geom.Sz(8.48, 89.09)} {
				l := testLayouts(geom.Sz(6, 5), cellSize, corner)["rectangular"]

				assert.Equal(t, l.FromPoint(corner), geom.Pt(0, 0), corner.String()+" "+cellSize.String()+": ")
			}
		}
	})
	t.Run("an edge shared by two squares belongs to the one after it", func(t *testing.T) {
		corner := geom.Pt(7.0, -448.0)
		l := testLayouts(geom.Sz(6, 5), geom.Sz(49.0, 82.0), corner)["rectangular"]

		for index := range Arr[int](geom.Sz(9, 9)).Keys() {
			index = index.AddXY(-4, -4)
			edge := corner.AddXY(float64(index.X)*49, float64(index.Y)*82)

			assert.Equal(t, l.FromPoint(edge), index, index.String()+": ")
		}
	})
	t.Run("a vertex of a diamond goes to the same side on every diamond", func(t *testing.T) {
		l := testLayouts(geom.Sz(6, 5), geom.Sz(82.0, 49.0), geom.Pt(7.0, -448.0))["isometric"]
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

				assert.Equal(t, l.FromPoint(l.ToPoint(index).Add(vertex)), index.Add(owners[name]), name+" of "+index.String()+": ")
			}
		}
	})
	t.Run("round-trips through ToPoint", func(t *testing.T) {
		for name, l := range layouts {
			for index := range Arr[int](geom.Sz(12, 12)).Keys() {
				index = index.AddXY(-4, -4)

				assert.Equal(t, l.FromPoint(l.ToPoint(index)), index, name+" "+index.String()+": ")
			}
		}
	})
	t.Run("round-trips on cells too small to tell from nothing", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(1e-7, 3e-9), geom.Pt(0.0, 0.0)) {
			for index := range Arr[int](l.Size()).Keys() {
				assert.Equal(t, l.FromPoint(l.ToPoint(index)), index, name+" "+index.String()+": ")
			}
		}
	})
	t.Run("a point is in the cell whose center is nearest", func(t *testing.T) {
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(8.48, 89.09), geom.Pt(-816.19, 732.64)) {
			around := geom.Pt(2, 3)
			center := l.ToPoint(around)

			for step := range Arr[int](geom.Sz(41, 41)).Keys() {
				point := center.Add(l.CellSize().Vector().MultiplyXY(float64(step.X-20)/13, float64(step.Y-20)/13))

				if nearest, clear := testNearest(name, l, around, point); clear {
					assert.Equal(t, l.FromPoint(point), nearest, name+" "+point.String()+": ")
				}
			}
		}
	})
	t.Run("a point just inside a vertex is the cell of the polygon", func(t *testing.T) {
		for name, l := range layouts {
			for index := range Arr[int](l.Size()).Keys() {
				center := l.ToPoint(index)
				for vertex := range l.CellPolygon(index).Vertices() {
					assert.Equal(t, l.FromPoint(center.Lerp(vertex, 0.99)), index, name+" "+index.String()+": ")
				}
			}
		}
	})
	t.Run("the zero layout has every point in its first cell", func(t *testing.T) {
		var l Layout

		assert.Equal(t, l.FromPoint(geom.Pt(300.0, -200.0)), geom.Pt(0, 0))
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, l := range layouts {
			allocs := testing.AllocsPerRun(100, func() {
				sinkPoint = l.FromPoint(geom.Pt(170.0, 95.0))
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkLayout_FromPoint(b *testing.B) {
	layouts := testLayouts(geom.Sz(64, 64), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	for _, name := range slices.Sorted(maps.Keys(layouts)) {
		l := layouts[name]

		b.Run(name, func(b *testing.B) {
			for i := range b.N {
				sinkPoint = l.FromPoint(geom.Pt(float64(i&1023), 77.0))
			}
		})
	}
}

func FuzzLayout_FromPoint(f *testing.F) {
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
		for name, l := range testLayouts(geom.Sz(6, 5), geom.Sz(width, height), geom.Pt(x, y)) {
			assert.Equal(t, l.FromPoint(l.ToPoint(index)), index, name+": ")

			point := l.ToPoint(index).Add(l.CellSize().Vector().MultiplyXY(across, down))
			if nearest, clear := testNearest(name, l, index, point); clear {
				assert.Equal(t, l.FromPoint(point), nearest, name+" nearest: ")
			}
		}

		assert.Equal(t, NewRect[int](geom.Sz(6, 5), geom.Sz(width, height), func(options *RectOptions) {
			options.Origin = geom.Pt(x, y)
		}).IndexAt(geom.Pt(x, y)), geom.Pt(0, 0), "the corner of the bounds: ")

		corner := geom.Pt(math.Trunc(x), math.Trunc(y))
		whole := geom.Sz(math.Ceil(width), math.Ceil(height))
		layouts := testLayouts(geom.Sz(6, 5), whole, corner)
		edge := corner.AddXY(float64(column)*whole.Width, float64(row)*whole.Height)
		top := layouts["isometric"].ToPoint(index).AddXY(0, -whole.Height/2)

		assert.Equal(t, layouts["rectangular"].FromPoint(edge), index, "an edge of a square: ")
		assert.Equal(t, layouts["isometric"].FromPoint(top), index, "the top of a diamond: ")
	})
}

func TestLayout_CellPolygon(t *testing.T) {
	layouts := testLayouts(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	t.Run("a square and a diamond have four sides", func(t *testing.T) {
		assert.Equal(t, layouts["rectangular"].CellPolygon(geom.Pt(2, 3)).N, 4)
		assert.Equal(t, layouts["isometric"].CellPolygon(geom.Pt(2, 3)).N, 4)
	})
	t.Run("a hexagon has six sides", func(t *testing.T) {
		assert.Equal(t, layouts["flat-top"].CellPolygon(geom.Pt(2, 3)).N, 6)
		assert.Equal(t, layouts["pointy-top"].CellPolygon(geom.Pt(2, 3)).N, 6)
	})
	t.Run("is centered on the cell", func(t *testing.T) {
		for name, l := range layouts {
			for index := range Arr[int](l.Size()).Keys() {
				geomtest.AssertPoint(t, l.CellPolygon(index).Center, l.ToPoint(index), name+" "+index.String()+": ")
			}
		}
	})
	t.Run("fills the cell size", func(t *testing.T) {
		for name, l := range layouts {
			for index := range Arr[int](l.Size()).Keys() {
				geomtest.AssertBox(t, l.CellPolygon(index).Bounds(), testCellBox(l, index), name+" "+index.String()+": ")
			}
		}
	})
	t.Run("the zero layout has a polygon of no size", func(t *testing.T) {
		var l Layout

		assert.True(t, l.CellPolygon(geom.Pt(3, 2)).Size.IsZero())
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, l := range layouts {
			allocs := testing.AllocsPerRun(100, func() {
				sinkPolygon = l.CellPolygon(geom.Pt(4, 3))
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}
