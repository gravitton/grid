package grid

import (
	"fmt"
	"math"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
	hex "github.com/gravitton/hexagon"
)

func testNewHexLayout(t *testing.T, construct func(cellScale geom.Size[float64], system hex.CoordinateSystem) *HexLayout, orientation geom.Orientation) {
	t.Run("places the center of index zero at the world origin", func(t *testing.T) {
		for _, system := range hex.CoordinateSystems() {
			assert.Equal(t, construct(geom.Sz(32.0, 20.0), system).Center(geom.Pt(0, 0)), geom.Pt(0.0, 0.0), system.String()+": ")
		}
	})
	t.Run("draws a hexagon of the given radius and orientation", func(t *testing.T) {
		polygon := construct(geom.Sz(32.0, 20.0), hex.Axial).Polygon(geom.Pt(0, 0))

		geomtest.AssertRegularPolygon(t, polygon, geom.RegularPolygonWithOrientation(geom.Pt(0.0, 0.0), geom.Sz(32.0, 20.0), 6, orientation))
	})
	t.Run("indexes in the given coordinate system", func(t *testing.T) {
		axial := construct(geom.Sz(32.0, 20.0), hex.Axial)

		for _, system := range hex.CoordinateSystems() {
			layout := construct(geom.Sz(32.0, 20.0), system)
			for _, h := range hex.Pt(2, -1).Spiral(2) {
				assert.Equal(t, layout.Center(system.To(h)), axial.Center(h.Point()), system.String()+" "+h.String()+": ")
			}
		}
	})
	for name, bad := range map[string]geom.Size[float64]{
		"zero":     geom.Sz(0.0, 32.0),
		"negative": geom.Sz(32.0, -32.0),
		"NaN":      geom.Sz(math.NaN(), 32.0),
		"infinite": geom.Sz(32.0, math.Inf(1)),
	} {
		t.Run("panics for a "+name+" size", func(t *testing.T) {
			assert.PanicsWith(t, func() {
				construct(bad, hex.Axial)
			}, "grid: cell scale "+bad.String()+" is not positive and finite")
		})
	}
	t.Run("panics for a system outside the seven", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			construct(geom.Sz(32.0, 20.0), hex.CoordinateSystem(9))
		}, "grid: unknown coordinate system 9")
		assert.PanicsWith(t, func() {
			construct(geom.Sz(32.0, 20.0), hex.CoordinateSystemNone)
		}, "grid: unknown coordinate system -1")
	})
	t.Run("allocates once", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(100, func() {
			sinkHexLayout = construct(geom.Sz(32.0, 20.0), hex.OffsetOddR)
		}), 1.0)
	})
}

func TestNewHexFlatTopLayout(t *testing.T) {
	testNewHexLayout(t, NewHexFlatTopLayout, geom.OrientationFlatTop)
}

func TestNewHexPointyTopLayout(t *testing.T) {
	testNewHexLayout(t, NewHexPointyTopLayout, geom.OrientationPointyTop)
}

// hexLayoutGrid is a hexagonal grid with the layout that places its cells and the coordinate
// system both index them in.
type hexLayoutGrid struct {
	layout *HexLayout
	grid   *Grid[int]
	system hex.CoordinateSystem
}

// hexLayoutGrids returns, for each hexagonal grid of kindGrids, the layout of the same scale and
// coordinate system with the center of index zero where the grid places it.
func hexLayoutGrids() map[string]hexLayoutGrid {
	scale := geom.Sz(32.0, 20.0)
	grids := kindGrids(geom.Sz(6, 5), scale, geom.Pt(100.0, 50.0))
	tests := map[string]hexLayoutGrid{}

	for name, system := range map[string]hex.CoordinateSystem{
		"flat-top":        hex.OffsetOddQ,
		"flat-top even":   hex.OffsetEvenQ,
		"pointy-top":      hex.OffsetOddR,
		"pointy-top even": hex.OffsetEvenR,
	} {
		layout := NewHexPointyTopLayout(scale, system)
		if system == hex.OffsetOddQ || system == hex.OffsetEvenQ {
			layout = NewHexFlatTopLayout(scale, system)
		}
		layout.MoveTo(grids[name].Cell(geom.Pt(0, 0)).Center())

		tests[name] = hexLayoutGrid{layout, grids[name], system}
	}

	return tests
}

func TestHexLayout_CellSize(t *testing.T) {
	for name, test := range hexLayoutGrids() {
		assert.Equal(t, test.layout.CellSize(), test.grid.CellSize(), name+": ")
	}
}

func TestHexLayout_CellSpacing(t *testing.T) {
	for name, test := range hexLayoutGrids() {
		assert.Equal(t, test.layout.CellSpacing(), test.grid.CellSpacing(), name+": ")
	}
}

func TestHexLayout_Center(t *testing.T) {
	t.Run("places every cell where a grid places it", func(t *testing.T) {
		for name, test := range hexLayoutGrids() {
			for cell := range test.grid.All() {
				geomtest.AssertPoint(t, test.layout.Center(cell.Index()), cell.Center(), name+" "+cell.Index().String()+": ")
			}
		}
	})
	t.Run("places an index beyond any grid", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.Axial)

		geomtest.AssertPoint(t, layout.Center(geom.Pt(-2, 3)), geom.Pt(-96.0, 4*geom.Sqrt3*10))
	})
	t.Run("allocates nothing", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.OffsetOddQ)

		assert.Equal(t, testing.AllocsPerRun(100, func() {
			sinkCenter = layout.Center(geom.Pt(3, -2))
		}), 0.0)
	})
}

func BenchmarkHexLayout_Center(b *testing.B) {
	layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.OffsetOddQ)

	for i := range b.N {
		sinkCenter = layout.Center(geom.Pt(i&63, -(i & 31)))
	}
}

func TestHexLayout_Bounds(t *testing.T) {
	t.Run("is the box a grid gives the cell", func(t *testing.T) {
		for name, test := range hexLayoutGrids() {
			for cell := range test.grid.All() {
				geomtest.AssertBox(t, test.layout.Bounds(cell.Index()), cell.Bounds(), name+" "+cell.Index().String()+": ")
			}
		}
	})
	t.Run("holds the polygon", func(t *testing.T) {
		for name, test := range hexLayoutGrids() {
			for _, h := range hex.Pt(1, -3).Spiral(2) {
				index := test.system.To(h)

				geomtest.AssertBox(t, test.layout.Bounds(index), test.layout.Polygon(index).Bounds(), name+" "+index.String()+": ")
			}
		}
	})
	t.Run("meets the box of a neighbor in the same number", func(t *testing.T) {
		flatTop := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.Axial)
		pointyTop := NewHexPointyTopLayout(geom.Sz(32.0, 20.0), hex.Axial)

		for _, h := range hex.Pt(-4, 7).Spiral(3) {
			assert.Equal(t, flatTop.Bounds(h.Point()).Max.Y, flatTop.Bounds(h.Add(hex.Pt(0, 1)).Point()).Min.Y, "flat-top "+h.String()+": ")
			assert.Equal(t, pointyTop.Bounds(h.Point()).Max.X, pointyTop.Bounds(h.Add(hex.Pt(1, 0)).Point()).Min.X, "pointy-top "+h.String()+": ")
		}
	})
}

func TestHexLayout_Polygon(t *testing.T) {
	t.Run("is the polygon a grid gives the cell", func(t *testing.T) {
		for name, test := range hexLayoutGrids() {
			for cell := range test.grid.All() {
				geomtest.AssertRegularPolygon(t, test.layout.Polygon(cell.Index()), cell.Polygon(), name+" "+cell.Index().String()+": ")
			}
		}
	})
	t.Run("is centered on the cell", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.DoubleHeight)

		assert.Equal(t, layout.Polygon(geom.Pt(3, -1)).Center, layout.Center(geom.Pt(3, -1)))
	})
}

func TestHexLayout_IndexAt(t *testing.T) {
	t.Run("finds every index at its center", func(t *testing.T) {
		for name, test := range hexLayoutGrids() {
			for _, h := range hex.Pt(-2, 5).Spiral(6) {
				index := test.system.To(h)

				assert.Equal(t, test.layout.IndexAt(test.layout.Center(index)), index, name+" "+index.String()+": ")
			}
		}
	})
	t.Run("finds the index a grid finds", func(t *testing.T) {
		for name, test := range hexLayoutGrids() {
			for cell := range test.grid.All() {
				polygon := cell.Polygon()
				for vertex := range polygon.Vertices() {
					point := polygon.Center.Lerp(vertex, 0.9)

					assert.Equal(t, test.layout.IndexAt(point), test.grid.IndexAt(point), name+" "+point.String()+": ")
				}
			}
		}
	})
	t.Run("puts a point on an edge where a grid puts it", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.OffsetOddQ)
		g := NewHexFlatTop[int](geom.Sz(6, 5), geom.Sz(32.0, 20.0))
		layout.MoveTo(g.Cell(geom.Pt(0, 0)).Center())

		for x := 32.0; x <= 160; x += 16 {
			point := geom.Pt(x, 52.0)

			assert.Equal(t, layout.IndexAt(point), g.IndexAt(point), point.String()+": ")
		}
	})
	t.Run("panics for a NaN point", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.Axial)

		assert.Panics(t, func() {
			sinkPoint = layout.IndexAt(geom.Pt(math.NaN(), 0.0))
		})
	})
	t.Run("allocates nothing", func(t *testing.T) {
		layout := NewHexPointyTopLayout(geom.Sz(32.0, 20.0), hex.OffsetOddR)

		assert.Equal(t, testing.AllocsPerRun(100, func() {
			sinkPoint = layout.IndexAt(geom.Pt(123.4, -56.7))
		}), 0.0)
	})
}

func BenchmarkHexLayout_IndexAt(b *testing.B) {
	layout := NewHexPointyTopLayout(geom.Sz(32.0, 20.0), hex.OffsetOddR)

	for i := range b.N {
		sinkPoint = layout.IndexAt(geom.Pt(float64(i&1023), float64(i&511)))
	}
}

func TestHexLayout_MoveTo(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)

	t.Run("places the center of index zero at the origin", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.OffsetEvenQ)
		layout.MoveTo(origin)

		assert.Equal(t, layout.Center(geom.Pt(0, 0)), origin)
	})
	t.Run("shifts every cell by what the origin moved", func(t *testing.T) {
		before := NewHexPointyTopLayout(geom.Sz(32.0, 20.0), hex.Axial)
		after := NewHexPointyTopLayout(geom.Sz(32.0, 20.0), hex.Axial)
		after.MoveTo(origin)

		for _, h := range hex.Pt(3, -1).Spiral(3) {
			geomtest.AssertPoint(t, after.Center(h.Point()), before.Center(h.Point()).Add(origin.Vector()), h.String()+": ")
		}
	})
	t.Run("a failed move leaves the layout where it was", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.Axial)

		assert.PanicsWith(t, func() {
			layout.MoveTo(geom.Pt(math.Inf(1), 50.0))
		}, "grid: origin "+geom.Pt(math.Inf(1), 50.0).String()+" is not finite")
		assert.Equal(t, layout.Center(geom.Pt(0, 0)), geom.Pt(0.0, 0.0))
	})
}

func TestHexLayout_Resize(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)
	scale := geom.Sz(10.0, 12.0)

	t.Run("is the layout built at that scale and moved", func(t *testing.T) {
		resized := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.Axial)
		resized.MoveTo(origin)
		resized.Resize(scale)
		built := NewHexFlatTopLayout(scale, hex.Axial)
		built.MoveTo(origin)

		assert.Equal(t, resized.CellSize(), built.CellSize())
		for _, h := range hex.Pt(3, -1).Spiral(3) {
			assert.Equal(t, resized.Polygon(h.Point()), built.Polygon(h.Point()), h.String()+": ")
		}
	})
	t.Run("a failed resize leaves the layout as it was", func(t *testing.T) {
		layout := NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.Axial)

		assert.PanicsWith(t, func() {
			layout.Resize(geom.Sz(-1.0, 2.0))
		}, "grid: cell scale "+geom.Sz(-1.0, 2.0).String()+" is not positive and finite")
		assert.Equal(t, layout.CellSize(), NewHexFlatTopLayout(geom.Sz(32.0, 20.0), hex.Axial).CellSize())
	})
}

func ExampleHexLayout() {
	layout := NewHexFlatTopLayout(HexFlatTopCellScale(32), hex.Axial)
	layout.MoveTo(geom.Pt(400.0, 300.0))

	h := hex.Pt(2, -1)

	fmt.Println(layout.Center(h.Point()))
	fmt.Println(layout.Bounds(h.Point()))
	fmt.Println(hex.Axial.From(layout.IndexAt(geom.Pt(470.0, 320.0))))

	offset := NewHexFlatTopLayout(HexFlatTopCellScale(32), hex.OffsetOddQ)
	offset.MoveTo(geom.Pt(400.0, 300.0))
	fmt.Println(offset.IndexAt(geom.Pt(470.0, 320.0)))
	// Output:
	// (448.00,300.00)
	// (432.00,286.14)-(464.00,313.86)
	// (3,-1)
	// (3,0)
}
