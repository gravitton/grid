package grid

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"math"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/gravitton/assert"
	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

var (
	sinkRegion    Region
	sinkRegions   []Region
	sinkPoints    []Index
	sinkPoint     Index
	sinkCenter    geom.Point[float64]
	sinkBox       geom.Box[float64]
	sinkSize      geom.Size[float64]
	sinkBool      bool
	sinkInt       int
	sinkString    string
	sinkHexLayout *HexLayout
	sinkPolygon   geom.RegularPolygon[float64]
	sinkFloat     float64
	sinkField     Field
)

var (
	testStart = geom.Pt(0, 0)
	testGoal  = geom.Pt(4, 4)
	testWall  = geom.Pt(2, 2)
)

// layoutGrids returns one grid of each layout, 6 by 5.
func layoutGrids() map[string]*Grid[int] {
	return map[string]*Grid[int]{
		"rectangular": NewRect[int](geom.Sz(6, 5), geom.Sz(64.0, 32.0)),
		"isometric":   NewIsometric[int](geom.Sz(6, 5), geom.Sz(64.0, 32.0)),
		"flat-top":    NewHexFlatTop[int](geom.Sz(6, 5), geom.Sz(32.0, 20.0)),
		"pointy-top":  NewHexPointyTop[int](geom.Sz(6, 5), geom.Sz(32.0, 20.0)),
	}
}

// kindGrids returns one grid of each kind and offset, of the given size in cells and cell
// scale, at the given origin.
func kindGrids(size geom.Size[int], cellScale geom.Size[float64], origin geom.Point[float64]) map[string]*Grid[int] {
	rect := func(options *SquareOptions) {
		options.Origin = origin
	}
	odd := func(options *HexOptions) {
		options.Origin = origin
	}
	even := func(options *HexOptions) {
		options.Origin = origin
		options.EvenOffset = true
	}

	return map[string]*Grid[int]{
		"rectangular":     NewRect[int](size, cellScale, rect),
		"isometric":       NewIsometric[int](size, cellScale, rect),
		"flat-top":        NewHexFlatTop[int](size, cellScale, odd),
		"flat-top even":   NewHexFlatTop[int](size, cellScale, even),
		"pointy-top":      NewHexPointyTop[int](size, cellScale, odd),
		"pointy-top even": NewHexPointyTop[int](size, cellScale, even),
	}
}

// benchmarkGrids runs the benchmark on each of the kindGrids at 64 by 64 cells, in the order of
// their names.
func benchmarkGrids(b *testing.B, run func(b *testing.B, g *Grid[int])) {
	grids := kindGrids(geom.Sz(64, 64), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	for _, name := range slices.Sorted(maps.Keys(grids)) {
		b.Run(name, func(b *testing.B) {
			run(b, grids[name])
		})
	}
}

func withDiagonalMovement(options *SquareOptions) {
	options.Movement = MovementDiagonal
}

// hexGrids returns one hexagonal grid of each offset coordinate system, 9 by 9.
func hexGrids() map[string]*Grid[int] {
	even := func(options *HexOptions) {
		options.EvenOffset = true
	}

	return map[string]*Grid[int]{
		"flat-top":        NewHexFlatTop[int](geom.Sz(9, 9), geom.SzU(16.0)),
		"flat-top even":   NewHexFlatTop[int](geom.Sz(9, 9), geom.SzU(16.0), even),
		"pointy-top":      NewHexPointyTop[int](geom.Sz(9, 9), geom.SzU(16.0)),
		"pointy-top even": NewHexPointyTop[int](geom.Sz(9, 9), geom.SzU(16.0), even),
	}
}

// ruleGrids returns one grid of each neighbor rule, of the given size in cells.
func ruleGrids(size geom.Size[int]) map[string]*Grid[int] {
	return map[string]*Grid[int]{
		"cardinal":   NewRect[int](size, geom.Sz(32.0, 32.0)),
		"diagonal":   NewRect[int](size, geom.Sz(32.0, 32.0), withDiagonalMovement),
		"flat-top":   NewHexFlatTop[int](size, geom.Sz(32.0, 32.0)),
		"pointy-top": NewHexPointyTop[int](size, geom.Sz(32.0, 32.0)),
	}
}

// fuzzGrid returns a grid of the kind, the width and the height a fuzz target was given: one
// of the two movements of a rectangular grid, an isometric grid, or a hexagonal one of either
// orientation and offset, of one to ten cells a side.
func fuzzGrid(kind, width, height uint8) *Grid[int] {
	size, cell := geom.Sz(1+int(width%10), 1+int(height%10)), geom.Sz(32.0, 32.0)
	even := func(options *HexOptions) {
		options.EvenOffset = true
	}

	switch kind % 7 {
	case 0:
		return NewRect[int](size, cell)
	case 1:
		return NewRect[int](size, cell, withDiagonalMovement)
	case 2:
		return NewIsometric[int](size, cell)
	case 3:
		return NewHexFlatTop[int](size, cell)
	case 4:
		return NewHexFlatTop[int](size, cell, even)
	case 5:
		return NewHexPointyTop[int](size, cell)
	default:
		return NewHexPointyTop[int](size, cell, even)
	}
}

// fuzzIndex returns an index in or just around the grid from the two bytes a fuzz target was
// given.
func fuzzIndex(g *Grid[int], x, y uint8) Index {
	return geom.Pt(int(x)%(g.Width()+4)-2, int(y)%(g.Height()+4)-2)
}

// walledGrid returns a 5x5 grid with one wall in the middle.
func walledGrid() *Grid[int] {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	g.Cell(testWall).Set(1)

	return g
}

// sealedGrid returns a 5x5 grid with a wall across it, parting the start from the goal.
func sealedGrid() *Grid[int] {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	for y := range 5 {
		g.Cell(geom.Pt(2, y)).Set(1)
	}

	return g
}

func walkable(cell Cell[int]) bool {
	return *cell.Value() == 0
}

// weightedGrid returns a 5x5 grid whose middle column costs twenty-one to enter, with a free
// cell at its foot, so the cheapest way across is the twelve steps around.
func weightedGrid() *Grid[int] {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	for y := range 4 {
		g.Cell(geom.Pt(2, y)).Set(20)
	}

	return g
}

func stepCost(_, next Cell[int]) float64 {
	return float64(1 + *next.Value())
}

// climbCost charges one per step and one more per unit of height between the cells, so the
// cost of a step depends on both of them.
func climbCost(current, next Cell[int]) float64 {
	return float64(1 + max(*current.Value()-*next.Value(), *next.Value()-*current.Value()))
}

// scatterHeights sets every cell of the grid to a height from zero to three that varies with
// its index, so that climbCost makes a search find cheaper ways to cells it has reached.
func scatterHeights(g *Grid[int]) {
	for cell := range g.All() {
		cell.Set((cell.Index().X*3 + cell.Index().Y*7) % 4)
	}
}

// testNewRect runs the sections the two square constructors share.
func testNewRect(t *testing.T, construct func(opts ...func(*SquareOptions)) *Grid[int]) {
	g := construct()
	inner := geom.Pt(2, 2)

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, g.Size(), geom.Sz(10, 5))
	})
	t.Run("connects four neighbors by default", func(t *testing.T) {
		assert.Length(t, g.Cell(inner).Neighbors(), 4)
	})
	t.Run("every center maps back to its index", func(t *testing.T) {
		for cell := range g.All() {
			assert.Equal(t, g.IndexAt(cell.Center()), cell.Index(), cell.Index().String()+": ")
		}
	})
	t.Run("Movement chooses the neighbors", func(t *testing.T) {
		diagonal := construct(withDiagonalMovement)

		assert.Length(t, diagonal.Cell(inner).Neighbors(), 8)
	})
	t.Run("Origin moves every cell", func(t *testing.T) {
		moved := construct(func(options *SquareOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		for cell := range g.All() {
			geomtest.AssertPoint(t, moved.Cell(cell.Index()).Center(), cell.Center().Add(geom.Vec(100.0, 50.0)), cell.Index().String()+": ")
		}
	})
	t.Run("every callback is applied, in order", func(t *testing.T) {
		cardinal := construct(withDiagonalMovement, func(options *SquareOptions) {
			options.Movement = MovementCardinal
		})

		assert.Length(t, cardinal.Cell(inner).Neighbors(), 4)
	})
	for _, movement := range []Movement{MovementNone, Movement(7)} {
		number := strconv.Itoa(int(movement))

		t.Run("panics for the movement "+number, func(t *testing.T) {
			assert.PanicsWith(t, func() {
				construct(func(options *SquareOptions) {
					options.Movement = movement
				})
			}, ErrUnknownMovement)
		})
	}
}

func TestNewRect(t *testing.T) {
	construct := func(opts ...func(*SquareOptions)) *Grid[int] {
		return NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0), opts...)
	}

	testNewRect(t, construct)
	t.Run("sits with its top-left corner at the world origin by default", func(t *testing.T) {
		geomtest.AssertPoint(t, construct().Bounds().Min, geom.Pt(0.0, 0.0))
	})
	t.Run("Origin places the top-left corner", func(t *testing.T) {
		moved := construct(func(options *SquareOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		geomtest.AssertPoint(t, moved.Bounds().Min, geom.Pt(100.0, 50.0))
		geomtest.AssertPoint(t, moved.Cell(geom.Pt(0, 0)).Center(), geom.Pt(116.0, 66.0))
		assert.Equal(t, moved.IndexAt(geom.Pt(116.0, 66.0)), geom.Pt(0, 0))
	})
}

func TestNewIsometric(t *testing.T) {
	construct := func(opts ...func(*SquareOptions)) *Grid[int] {
		return NewIsometric[int](geom.Sz(10, 5), geom.Sz(64.0, 32.0), opts...)
	}

	testNewRect(t, construct)
	t.Run("draws a row down and to the left of the one before", func(t *testing.T) {
		g := construct()

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 1)).Center(), g.Cell(geom.Pt(0, 0)).Center().Add(geom.Vec(-32.0, 16.0)))
	})
	t.Run("the first cell is as far right as the rows reach left", func(t *testing.T) {
		g := construct()

		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(160.0, 16.0))
		geomtest.AssertNumber(t, g.Cell(geom.Pt(0, 4)).Bounds().Min.X, 0.0)
	})
}

// testNewHex runs the sections the two hexagonal constructors share.
func testNewHex(t *testing.T, construct func(opts ...func(*HexOptions)) *Grid[int]) {
	g := construct()
	second := geom.Pt(1, 1)

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, g.Size(), geom.Sz(4, 3))
	})
	t.Run("connects six neighbors", func(t *testing.T) {
		assert.Length(t, g.Cell(second).Neighbors(), 6)
	})
	t.Run("sits with its top-left corner at the world origin by default", func(t *testing.T) {
		geomtest.AssertPoint(t, g.Bounds().Min, geom.Pt(0.0, 0.0))
	})
	t.Run("every center maps back to its index", func(t *testing.T) {
		for cell := range g.All() {
			assert.Equal(t, g.IndexAt(cell.Center()), cell.Index(), cell.Index().String()+": ")
		}
	})
	t.Run("EvenOffset shifts the other rows or columns", func(t *testing.T) {
		even := construct(func(options *HexOptions) {
			options.EvenOffset = true
		})

		assert.NotEqual(t, even.Cell(second).Center(), g.Cell(second).Center())
		for cell := range even.All() {
			assert.Equal(t, even.IndexAt(cell.Center()), cell.Index(), cell.Index().String()+": ")
		}
	})
	t.Run("Origin moves every cell", func(t *testing.T) {
		moved := construct(func(options *HexOptions) {
			options.Origin = geom.Pt(100.0, 50.0)
		})

		geomtest.AssertPoint(t, moved.Bounds().Min, geom.Pt(100.0, 50.0))
		geomtest.AssertPoint(t, moved.Cell(second).Center(), g.Cell(second).Center().Add(geom.Vec(100.0, 50.0)))
	})
}

func TestNewHexFlatTop(t *testing.T) {
	testNewHex(t, func(opts ...func(*HexOptions)) *Grid[int] {
		return NewHexFlatTop[int](geom.Sz(4, 3), geom.SzU(16.0), opts...)
	})
}

func TestNewHexPointyTop(t *testing.T) {
	testNewHex(t, func(opts ...func(*HexOptions)) *Grid[int] {
		return NewHexPointyTop[int](geom.Sz(4, 3), geom.SzU(16.0), opts...)
	})
}

func TestRectCellScale(t *testing.T) {
	size := RectCellScale(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(64.0, 64.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewRect[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 64.0))
	})
}

func TestIsometricCellScale(t *testing.T) {
	size := IsometricCellScale(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(64.0, 36.9504172))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewIsometric[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 36.9504172))
	})
}

func TestIsometricPixelCellScale(t *testing.T) {
	size := IsometricPixelCellScale(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(64.0, 32.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewIsometric[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 32.0))
	})
}

func TestHexFlatTopCellScale(t *testing.T) {
	size := HexFlatTopCellScale(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(32.0, 32.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexFlatTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 55.4256258))
	})
}

func TestHexPointyTopCellScale(t *testing.T) {
	size := HexPointyTopCellScale(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(36.9504172, 36.9504172))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexPointyTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 73.9008346))
	})
}

func TestHexFlatTopPixelCellScale(t *testing.T) {
	size := HexFlatTopPixelCellScale(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(32.0, 27.7128129))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexFlatTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 48.0))
	})
}

func TestHexPointyTopPixelCellScale(t *testing.T) {
	size := HexPointyTopPixelCellScale(64)

	t.Run("is the size the constructor takes", func(t *testing.T) {
		geomtest.AssertSize(t, size, geom.Sz(36.9504172, 32.0))
	})
	t.Run("gives a tile of the width", func(t *testing.T) {
		geomtest.AssertSize(t, NewHexPointyTop[int](geom.SzU(1), size).CellSize(), geom.Sz(64.0, 64.0))
	})
}

func TestGrid_Size(t *testing.T) {
	g := NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.Size(), geom.Sz(10, 5))
}

func TestGrid_Width(t *testing.T) {
	g := NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.Width(), 10)
}

func TestGrid_Height(t *testing.T) {
	g := NewRect[int](geom.Sz(10, 5), geom.Sz(32.0, 32.0))

	assert.Equal(t, g.Height(), 5)
}

func TestGrid_Origin(t *testing.T) {
	for name, g := range kindGrids(geom.Sz(3, 3), geom.Sz(32.0, 20.0), geom.Pt(7.0, -3.0)) {
		assert.Equal(t, g.Origin(), geom.Pt(7.0, -3.0), name+": ")
		assert.Equal(t, g.Origin(), g.Bounds().Min, name+": ")

		g.MoveTo(geom.Pt(1.0, 2.0))

		assert.Equal(t, g.Origin(), geom.Pt(1.0, 2.0), name+": ")
	}
}

// testCellBox returns the box around the cell at the index, measured from its center.
func testCellBox(g *Grid[int], index Index) geom.Box[float64] {
	center := g.Cell(index).Center()
	half := g.CellSize().Scale(0.5).Vector()

	return geom.BoxFromMinMax(center.Add(half.Negate()), center.Add(half))
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
			for name, g := range kindGrids(size, geom.Sz(32.0, 20.0), origin) {
				around := testCellBox(g, geom.Pt(0, 0))
				for index := range container.Arr[int](size).Keys() {
					around = around.Union(testCellBox(g, index))
				}

				geomtest.AssertBox(t, g.Bounds(), around, name+" "+size.String()+": ")
			}
		}
	})
	t.Run("starts at the origin", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			assert.Equal(t, g.Bounds().Min, origin, name+": ")
		}
	})
	t.Run("a grid of no cells has an empty box at the origin", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(0, 0), geom.Sz(32.0, 20.0), origin) {
			assert.Equal(t, g.Bounds(), geom.Box[float64]{Min: origin, Max: origin}, name+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
			allocs := testing.AllocsPerRun(100, func() {
				sinkBox = g.Bounds()
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func TestGrid_CellScale(t *testing.T) {
	for name, g := range kindGrids(geom.Sz(3, 3), geom.Sz(32.0, 20.0), geom.Pt(0.0, 0.0)) {
		assert.Equal(t, g.CellScale(), geom.Sz(32.0, 20.0), name+": ")

		g.Rescale(geom.Sz(8.0, 4.0))

		assert.Equal(t, g.CellScale(), geom.Sz(8.0, 4.0), name+": ")
	}
}

func TestGrid_CellSize(t *testing.T) {
	grids := kindGrids(geom.Sz(6, 5), geom.Sz(10.0, 20.0), geom.Pt(0.0, 0.0))

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
		for name, g := range kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0)) {
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
	grids := kindGrids(geom.Sz(6, 5), geom.Sz(10.0, 20.0), geom.Pt(0.0, 0.0))

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
		for name, g := range kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0)) {
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

func TestGrid_Cell(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	t.Run("is a handle on the index", func(t *testing.T) {
		assert.Equal(t, g.Cell(geom.Pt(1, 2)).Index(), geom.Pt(1, 2))
		assert.True(t, g.Cell(geom.Pt(1, 2)).IsContained())
	})
	t.Run("two cells of one index share the value", func(t *testing.T) {
		g.Cell(geom.Pt(1, 2)).Set(42)

		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
	t.Run("outside the grid keeps the index and is not valid", func(t *testing.T) {
		assert.Equal(t, g.Cell(geom.Pt(99, -1)).Index(), geom.Pt(99, -1))
		assert.False(t, g.Cell(geom.Pt(99, -1)).IsContained())
	})
}

func ExampleGrid_Cell() {
	type tile struct {
		Walkable bool
		Cost     float64
	}

	g := NewRect[tile](geom.Sz(100, 100), RectCellScale(32))
	index := geom.Pt(3, 4)

	cell := g.Cell(index)
	cell.Set(tile{Cost: 3})
	cell.Value().Cost = 5

	outside := g.Cell(geom.Pt(-1, 0))
	outside.Set(tile{Cost: 9})

	fmt.Println(g.Contains(index), g.Size())
	fmt.Println(cell, cell.IsContained(), cell.Value().Cost)
	fmt.Println(cell.Neighbors())
	fmt.Println(cell.DistanceTo(geom.Pt(0, 0)))
	fmt.Println(outside, outside.IsContained(), outside.Value() == nil)
	// Output:
	// true 100x100
	// Cell(3,4) true 5
	// [(4,4) (3,5) (2,4) (3,3)]
	// 7
	// Cell(-1,0) false true
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

// testNearest returns the index whose center is nearest the point among the cells around the
// given one, in the measure that makes a cell of the named layout the points nearest its
// center, and whether it is nearer than every other by more than rounding.
func testNearest(name string, g *Grid[int], around Index, point geom.Point[float64]) (Index, bool) {
	half := g.CellSize().Scale(0.5)
	nearest, best, second := around, math.Inf(1), math.Inf(1)

	for index := range container.Arr[int](geom.Sz(7, 7)).Keys() {
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

func TestGrid_IndexAt(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)
	grids := kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin)

	t.Run("a point inside a square is its cell", func(t *testing.T) {
		g := NewRect[int](geom.Sz(4, 3), geom.Sz(64.0, 32.0))

		assert.Equal(t, g.IndexAt(geom.Pt(1.0, 1.0)), geom.Pt(0, 0))
		assert.Equal(t, g.IndexAt(geom.Pt(63.0, 31.0)), geom.Pt(0, 0))
		assert.Equal(t, g.IndexAt(geom.Pt(130.0, 70.0)), geom.Pt(2, 2))
	})
	t.Run("the corner of the bounds is the first square wherever the grid is", func(t *testing.T) {
		for _, corner := range []geom.Point[float64]{origin, geom.Pt(0.1, -7.3), geom.Pt(-816.19, 732.64), geom.Pt(1e6/3, -1e-3)} {
			for _, cellScale := range []geom.Size[float64]{geom.Sz(32.0, 20.0), geom.Sz(49.0, 82.0), geom.Sz(8.48, 89.09)} {
				g := kindGrids(geom.Sz(6, 5), cellScale, corner)["rectangular"]

				assert.Equal(t, g.IndexAt(corner), geom.Pt(0, 0), corner.String()+" "+cellScale.String()+": ")
			}
		}
	})
	t.Run("an edge shared by two squares belongs to the one after it", func(t *testing.T) {
		corner := geom.Pt(7.0, -448.0)
		g := kindGrids(geom.Sz(6, 5), geom.Sz(49.0, 82.0), corner)["rectangular"]

		for index := range container.Arr[int](geom.Sz(9, 9)).Keys() {
			index = index.AddXY(-4, -4)
			edge := corner.AddXY(float64(index.X)*49, float64(index.Y)*82)

			assert.Equal(t, g.IndexAt(edge), index, index.String()+": ")
		}
	})
	t.Run("a vertex of a diamond goes to the same side on every diamond", func(t *testing.T) {
		g := kindGrids(geom.Sz(6, 5), geom.Sz(82.0, 49.0), geom.Pt(7.0, -448.0))["isometric"]
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
			for index := range container.Arr[int](geom.Sz(9, 9)).Keys() {
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
			for index := range container.Arr[int](geom.Sz(12, 12)).Keys() {
				index = index.AddXY(-4, -4)

				assert.Equal(t, g.IndexAt(g.Cell(index).Center()), index, name+" "+index.String()+": ")
			}
		}
	})
	t.Run("round-trips on cells too small to tell from nothing", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(6, 5), geom.Sz(1e-7, 3e-9), geom.Pt(0.0, 0.0)) {
			for index := range container.Arr[int](g.Size()).Keys() {
				assert.Equal(t, g.IndexAt(g.Cell(index).Center()), index, name+" "+index.String()+": ")
			}
		}
	})
	t.Run("a point is in the cell whose center is nearest", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(6, 5), geom.Sz(8.48, 89.09), geom.Pt(-816.19, 732.64)) {
			around := geom.Pt(2, 3)
			center := g.Cell(around).Center()

			for step := range container.Arr[int](geom.Sz(41, 41)).Keys() {
				point := center.Add(g.CellSize().Vector().MultiplyXY(float64(step.X-20)/13, float64(step.Y-20)/13))

				if nearest, clear := testNearest(name, g, around, point); clear {
					assert.Equal(t, g.IndexAt(point), nearest, name+" "+point.String()+": ")
				}
			}
		}
	})
	t.Run("a point just inside a vertex is the cell of the polygon", func(t *testing.T) {
		for name, g := range grids {
			for index := range container.Arr[int](g.Size()).Keys() {
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
				}, geom.ErrNonFinite)
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
		for name, g := range kindGrids(geom.Sz(6, 5), geom.Sz(width, height), geom.Pt(x, y)) {
			assert.Equal(t, g.IndexAt(g.Cell(index).Center()), index, name+": ")

			point := g.Cell(index).Center().Add(g.CellSize().Vector().MultiplyXY(across, down))
			if nearest, clear := testNearest(name, g, index, point); clear {
				assert.Equal(t, g.IndexAt(point), nearest, name+" nearest: ")
			}
		}

		assert.Equal(t, NewRect[int](geom.Sz(6, 5), geom.Sz(width, height), func(options *SquareOptions) {
			options.Origin = geom.Pt(x, y)
		}).IndexAt(geom.Pt(x, y)), geom.Pt(0, 0), "the corner of the bounds: ")

		corner := geom.Pt(math.Trunc(x), math.Trunc(y))
		whole := geom.Sz(math.Ceil(width), math.Ceil(height))
		grids := kindGrids(geom.Sz(6, 5), whole, corner)
		edge := corner.AddXY(float64(column)*whole.Width, float64(row)*whole.Height)
		top := grids["isometric"].Cell(index).Center().AddXY(0, -whole.Height/2)

		assert.Equal(t, grids["rectangular"].IndexAt(edge), index, "an edge of a square: ")
		assert.Equal(t, grids["isometric"].IndexAt(top), index, "the top of a diamond: ")
	})
}

func ExampleGrid_IndexAt() {
	g := NewRect[int](geom.Sz(100, 100), RectCellScale(32))

	cell := g.CellAt(geom.Pt(499.0, 123.4))

	fmt.Println(g.IndexAt(geom.Pt(499.0, 123.4)))
	fmt.Println(cell, cell.Center(), cell.Bounds())
	fmt.Println(g.Contains(g.IndexAt(geom.Pt(-5.0, 20.0))))
	// Output:
	// (15,3)
	// Cell(15,3) (496.00,112.00) (480.00,96.00)-(512.00,128.00)
	// false
}

func TestGrid_Fill(t *testing.T) {
	g := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	g.Fill(5)

	for cell := range g.All() {
		assert.Equal(t, *cell.Value(), 5, cell.Index().String()+": ")
	}
}

func TestGrid_Clear(t *testing.T) {
	g := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	g.Fill(5)
	g.Clear()

	for cell := range g.All() {
		assert.Equal(t, *cell.Value(), 0, cell.Index().String()+": ")
	}
}

func TestGrid_Clone(t *testing.T) {
	original := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))
	original.Fill(1)

	clone := original.Clone()
	clone.Cell(geom.Pt(0, 0)).Set(99)

	t.Run("holds the same values", func(t *testing.T) {
		assert.Equal(t, clone.Size(), original.Size())
		assert.Equal(t, *clone.Cell(geom.Pt(2, 2)).Value(), 1)
	})
	t.Run("keeps the layout and the neighbor rule", func(t *testing.T) {
		geomtest.AssertBox(t, clone.Bounds(), original.Bounds())
		assert.Equal(t, clone.Neighbors(geom.Pt(1, 1)), original.Neighbors(geom.Pt(1, 1)))
	})
	t.Run("a write to the clone leaves the original alone", func(t *testing.T) {
		assert.Equal(t, *original.Cell(geom.Pt(0, 0)).Value(), 1)
		assert.Equal(t, *clone.Cell(geom.Pt(0, 0)).Value(), 99)
	})
}

func TestGrid_Layer(t *testing.T) {
	for name, g := range layoutGrids() {
		g.Cell(geom.Pt(1, 1)).Set(7)
		layer := g.Layer[string]()

		t.Run(name, func(t *testing.T) {
			t.Run("keeps the size, the layout and the neighbor rule", func(t *testing.T) {
				assert.Equal(t, layer.Size(), g.Size())
				assert.Equal(t, layer.Bounds(), g.Bounds())
				assert.Equal(t, layer.Cell(geom.Pt(1, 1)).Center(), g.Cell(geom.Pt(1, 1)).Center())
				assert.Equal(t, layer.Neighbors(geom.Pt(1, 1)), g.Neighbors(geom.Pt(1, 1)))
			})
			t.Run("holds zero values", func(t *testing.T) {
				assert.Equal(t, *layer.Cell(geom.Pt(1, 1)).Value(), "")
			})
			t.Run("moves apart from the grid", func(t *testing.T) {
				before := g.Bounds()
				layer.MoveTo(geom.Pt(500.0, 500.0))

				assert.Equal(t, g.Bounds(), before)
			})
		})
	}
}

func ExampleGrid_Layer() {
	g := NewHexFlatTop[string](geom.Sz(4, 4), HexFlatTopCellScale(64))
	visited := g.Layer[bool]()

	visited.Cell(geom.Pt(1, 2)).Set(true)

	fmt.Println(visited.Size(), visited.Bounds() == g.Bounds())
	// Output:
	// 4x4 true
}

func TestGrid_All(t *testing.T) {
	g := NewHexFlatTop[int](geom.Sz(3, 2), geom.SzU(16.0))

	t.Run("yields every cell once, row by row", func(t *testing.T) {
		var indices []Index
		for cell := range g.All() {
			indices = append(indices, cell.Index())
		}

		assert.Equal(t, indices, []Index{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(2, 0), geom.Pt(0, 1), geom.Pt(1, 1), geom.Pt(2, 1)})
	})
	t.Run("a cell writes through to the grid", func(t *testing.T) {
		for cell := range g.All() {
			cell.Set(cell.Index().X)
		}

		assert.Equal(t, *g.Cell(geom.Pt(2, 1)).Value(), 2)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range g.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
}

func BenchmarkGrid_All(b *testing.B) {
	g := NewRect[int](geom.Sz(64, 64), geom.Sz(32.0, 32.0))

	for b.Loop() {
		for cell := range g.All() {
			sinkPoint = cell.Index()
		}
	}
}

func TestGrid_Select(t *testing.T) {
	g := NewRect[int](geom.Sz(3, 2), RectCellScale(32))
	g.Cell(geom.Pt(2, 0)).Set(1)
	g.Cell(geom.Pt(0, 1)).Set(1)

	t.Run("returns the cells selected reports", func(t *testing.T) {
		selected := g.Select(func(cell Cell[int]) bool {
			return *cell.Value() == 1
		})

		assert.Equal(t, selected.Elements(), []Index{geom.Pt(0, 1), geom.Pt(2, 0)})
	})
	t.Run("a nil selected is every cell", func(t *testing.T) {
		assert.Equal(t, g.Select(nil).Len(), 6)
	})
	t.Run("holds exactly what the callback picks, on every layout", func(t *testing.T) {
		for name, g := range layoutGrids() {
			corner := func(cell Cell[int]) bool {
				return g.Distance(cell.Index(), geom.Pt(0, 0)) <= 2
			}

			assert.True(t, g.Select(corner).Equal(g.Range(geom.Pt(0, 0), 2)), name+": ")
		}
	})
	t.Run("allocates once, sized for the grid", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkRegion = g.Select(nil)
		}), 1.0)
	})
}

func ExampleGrid_Select() {
	g := NewRect[int](geom.Sz(3, 2), RectCellScale(32))
	g.Cell(geom.Pt(2, 0)).Set(1)

	fmt.Println(g.Select(func(cell Cell[int]) bool {
		return *cell.Value() == 1
	}).Elements())
	// Output:
	// [(2,0)]
}

func TestGrid_DrawOrder(t *testing.T) {
	whole := geom.Box[float64]{}

	t.Run("a zero bounds walks every cell of the grid once", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(3, 4), geom.Sz(32.0, 20.0), geom.Pt(0.0, 0.0)) {
			seen := map[Index]int{}
			for cell := range g.DrawOrder(whole) {
				seen[cell.Index()]++
			}

			assert.Equal(t, len(seen), 12, name+": ")

			for index, times := range seen {
				assert.True(t, g.Contains(index), name+" at "+index.String()+": ")
				assert.Equal(t, times, 1, name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("a grid of no cells walks nothing", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(0, 0), geom.Sz(32.0, 20.0), geom.Pt(10.0, 10.0)) {
			count := 0
			for range g.DrawOrder(whole) {
				count++
			}

			assert.Equal(t, count, 0, name+": ")
		}
	})
	t.Run("stops at the break wherever it comes, on every layout", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(3, 3), geom.Sz(32.0, 20.0), geom.Pt(0.0, 0.0)) {
			for stop := 1; stop <= 9; stop++ {
				count := 0
				for range g.DrawOrder(whole) {
					count++
					if count == stop {
						break
					}
				}

				assert.Equal(t, count, stop, name+": ")
			}
		}
	})
	t.Run("every cell partly in view is yielded, on every layout and viewport", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(9, 7), geom.Sz(32.0, 20.0), geom.Pt(7.0, -3.0)) {
			corner := g.Bounds().Min

			for _, offset := range []geom.Vector[float64]{geom.Vec(1.0, 1.0), geom.Vec(30.0, 19.0), geom.Vec(61.5, 37.5), geom.Vec(-50.0, -50.0)} {
				viewport := geom.BoxFromMinMax(corner.Add(offset), corner.Add(offset).Add(geom.Vec(75.0, 45.0)))

				seen := map[Index]bool{}
				for cell := range g.DrawOrder(viewport) {
					seen[cell.Index()] = true
				}

				for cell := range g.All() {
					if cell.Bounds().IntersectsBox(viewport) {
						assert.True(t, seen[cell.Index()], name+" at "+cell.Index().String()+": ")
					}
				}
			}
		}
	})
	t.Run("a bounds keeps to the cells in view", func(t *testing.T) {
		g := NewRect[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
		viewport := geom.BoxFromMinMax(geom.Pt(-32.0, -32.0), geom.Pt(31.0, 31.0))

		count := 0
		for range g.DrawOrder(viewport) {
			count++
		}

		assert.Equal(t, count, 1)
	})
	t.Run("a viewport starting late in a cell reaches the last cell it touches", func(t *testing.T) {
		g := NewRect[int](geom.Sz(10, 10), geom.Sz(32.0, 32.0))
		viewport := geom.BoxFromMinMax(geom.Pt(31.0, 31.0), geom.Pt(64.5, 64.5))

		var last Index
		for cell := range g.DrawOrder(viewport) {
			last = cell.Index()
		}

		assert.Equal(t, last, geom.Pt(2, 2))
	})
	t.Run("isometric depth never decreases", func(t *testing.T) {
		g := NewIsometric[int](geom.Sz(4, 4), geom.Sz(64.0, 32.0))

		previous := -1000
		for cell := range g.DrawOrder(whole) {
			depth := cell.Index().X + cell.Index().Y
			assert.GreaterOrEqual(t, depth, previous)
			previous = depth
		}
	})
	t.Run("a cell is never drawn after one in front of it, on every layout", func(t *testing.T) {
		grids := kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), geom.Pt(0.0, 0.0))
		grids["flat-top even"] = NewHexFlatTop[int](geom.Sz(6, 5), geom.SzU(16.0), func(options *HexOptions) {
			options.EvenOffset = true
		})
		grids["pointy-top even"] = NewHexPointyTop[int](geom.Sz(6, 5), geom.SzU(16.0), func(options *HexOptions) {
			options.EvenOffset = true
		})

		for name, g := range grids {
			previous := math.Inf(-1)
			for cell := range g.DrawOrder(whole) {
				assert.GreaterOrEqual(t, cell.Center().Y, previous, name+" at "+cell.Index().String()+": ")
				previous = max(previous, cell.Center().Y-g.CellSpacing().Height)
			}
		}
	})
	t.Run("panics once ranged for bounds that are not finite, on every layout", func(t *testing.T) {
		for name, g := range kindGrids(geom.Sz(3, 3), geom.Sz(32.0, 20.0), geom.Pt(0.0, 0.0)) {
			bounds := geom.BoxFromMinMax(geom.Pt(0.0, 0.0), geom.Pt(math.Inf(1), 10.0))
			order := g.DrawOrder(bounds)

			assert.PanicsWith(t, func() {
				for range order {
					break
				}
			}, ErrNonFinite, name+": ")
		}
	})
}

func BenchmarkGrid_DrawOrder(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		view := geom.BoxFromMinMax(g.Bounds().Center(), g.Bounds().Center().Add(geom.Vec(640.0, 480.0)))

		for b.Loop() {
			for cell := range g.DrawOrder(view) {
				sinkPoint = cell.Index()
			}
		}
	})
}

func ExampleGrid_DrawOrder() {
	g := NewIsometric[int](geom.Sz(2, 2), IsometricPixelCellScale(64))

	for cell := range g.DrawOrder(geom.Box[float64]{}) {
		fmt.Print(cell.Index(), " ")
	}
	// Output:
	// (0,0) (0,1) (1,0) (1,1)
}

func TestGrid_MoveTo(t *testing.T) {
	origin := geom.Pt(100.0, 50.0)
	start := geom.Pt(7.0, 9.0)
	moved := func() (map[string]*Grid[int], map[string]*Grid[int]) {
		before := kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), start)
		after := kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), start)
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

		for name, built := range kindGrids(geom.Sz(6, 5), geom.Sz(32.0, 20.0), origin) {
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
		for name, g := range kindGrids(geom.Sz(0, 0), geom.Sz(32.0, 20.0), start) {
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
				}, ErrNonFinite)
			}
		})
	}
	t.Run("a failed move leaves the grid where it was", func(t *testing.T) {
		g := NewRect[int](geom.Sz(6, 5), geom.Sz(32.0, 20.0))

		assert.Panics(t, func() {
			g.MoveTo(geom.Pt(math.NaN(), 50.0))
		})
		assert.Equal(t, g.Bounds().Min, geom.Pt(0.0, 0.0))
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
	g := NewRect[int](geom.Sz(10, 10), RectCellScale(32))

	g.MoveTo(geom.Pt(100.0, 50.0))
	fmt.Println(g.Bounds(), g.Cell(geom.Pt(0, 0)).Center())

	g.Rescale(RectCellScale(48))
	fmt.Println(g.Bounds(), g.Cell(geom.Pt(0, 0)).Center())
	// Output:
	// (100.00,50.00)-(420.00,370.00) (116.00,66.00)
	// (100.00,50.00)-(580.00,530.00) (124.00,74.00)
}

func TestGrid_Rescale(t *testing.T) {
	origin := geom.Pt(10.0, 20.0)
	scale := geom.Sz(20.0, 12.0)
	grids := func() map[string]*Grid[int] {
		return kindGrids(geom.Sz(6, 5), geom.Sz(64.0, 32.0), origin)
	}

	t.Run("replaces the size of a square", func(t *testing.T) {
		g := grids()["rectangular"]
		g.Rescale(geom.Sz(128.0, 64.0))

		geomtest.AssertSize(t, g.CellSize(), geom.Sz(128.0, 64.0))
		geomtest.AssertSize(t, g.CellSpacing(), geom.Sz(128.0, 64.0))
	})
	t.Run("replaces the circumradius of a hexagon", func(t *testing.T) {
		g := grids()["flat-top"]
		g.Rescale(geom.Sz(10.0, 10.0))

		geomtest.AssertSize(t, g.CellSize(), geom.Sz(20.0, geom.Sqrt3*10))
	})
	t.Run("is the grid built at that size", func(t *testing.T) {
		built := kindGrids(geom.Sz(6, 5), scale, origin)

		for name, g := range grids() {
			g.Rescale(scale)

			assert.Equal(t, g.Bounds(), built[name].Bounds(), name+": ")
			assert.Equal(t, g.CellSize(), built[name].CellSize(), name+": ")
			assert.Equal(t, g.CellSpacing(), built[name].CellSpacing(), name+": ")
			assert.Equal(t, g.Cell(geom.Pt(4, 3)).Center(), built[name].Cell(geom.Pt(4, 3)).Center(), name+": ")
			assert.Equal(t, g.Cell(geom.Pt(4, 3)).Polygon(), built[name].Cell(geom.Pt(4, 3)).Polygon(), name+": ")
		}
	})
	t.Run("keeps the origin", func(t *testing.T) {
		for name, g := range grids() {
			g.Rescale(scale)

			assert.Equal(t, g.Bounds().Min, origin, name+": ")
		}
	})
	t.Run("keeps the values", func(t *testing.T) {
		g := grids()["rectangular"]
		g.Fill(5)
		g.Rescale(scale)

		assert.Equal(t, *g.Cell(geom.Pt(2, 3)).Value(), 5)
	})
	t.Run("leaves a clone at its size", func(t *testing.T) {
		g := grids()["rectangular"]
		clone := g.Clone()
		g.Rescale(scale)

		geomtest.AssertSize(t, clone.CellSize(), geom.Sz(64.0, 32.0))
	})
	t.Run("a failed resize leaves the grid as it was", func(t *testing.T) {
		g := grids()["rectangular"]

		assert.Panics(t, func() {
			g.Rescale(geom.Sz(-1.0, 2.0))
		})
		geomtest.AssertSize(t, g.CellSize(), geom.Sz(64.0, 32.0))
	})
	for name, bad := range map[string]geom.Size[float64]{
		"zero":     geom.Sz(0.0, 32.0),
		"negative": geom.Sz(32.0, -32.0),
		"NaN":      geom.Sz(math.NaN(), 32.0),
		"infinite": geom.Sz(32.0, math.Inf(1)),
	} {
		t.Run("panics for a "+name+" size", func(t *testing.T) {
			for _, g := range grids() {
				assert.PanicsWith(t, func() {
					g.Rescale(bad)
				}, ErrInvalidCellScale)
			}
		})
	}
	t.Run("allocates nothing", func(t *testing.T) {
		for name, g := range grids() {
			allocs := testing.AllocsPerRun(100, func() {
				g.Rescale(scale)
			})

			assert.Equal(t, allocs, 0.0, name+": ")
		}
	})
}

func BenchmarkGrid_Rescale(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for i := range b.N {
			g.Rescale(geom.Sz(float64(i&7+8), 20.0))
		}
	})
}

func testSeq(t *testing.T, seq func(g *Grid[int]) iter.Seq[Index], slice func(g *Grid[int]) []Index) {
	t.Helper()

	for name, g := range ruleGrids(geom.Sz(11, 11)) {
		t.Run(name, func(t *testing.T) {
			t.Run("yields what the slice form holds, in its order", func(t *testing.T) {
				assert.Equal(t, slices.Collect(seq(g)), slice(g))
			})
			t.Run("stops at the break", func(t *testing.T) {
				count := 0
				for range seq(g) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
		})
	}
}

func TestGrid_Neighbors(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	t.Run("cardinal has four, by increasing angle", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Neighbors(center), []Index{geom.Pt(6, 5), geom.Pt(5, 6), geom.Pt(4, 5), geom.Pt(5, 4)})
	})
	t.Run("diagonal has eight", func(t *testing.T) {
		assert.Length(t, grids["diagonal"].Neighbors(center), 8)
	})
	t.Run("a hexagon has six", func(t *testing.T) {
		assert.Length(t, grids["flat-top"].Neighbors(center), 6)
		assert.Length(t, grids["pointy-top"].Neighbors(center), 6)
	})
	t.Run("a corner keeps the neighbors inside the grid", func(t *testing.T) {
		assert.Length(t, grids["cardinal"].Neighbors(geom.Pt(0, 0)), 2)
		assert.Length(t, grids["diagonal"].Neighbors(geom.Pt(0, 0)), 3)
	})
	t.Run("is exactly the cells at distance one", func(t *testing.T) {
		for name, g := range grids {
			for cell := range g.All() {
				var expected []Index
				for other := range g.All() {
					if g.Distance(cell.Index(), other.Index()) == 1 {
						expected = append(expected, other.Index())
					}
				}

				assert.True(t, Reg(g.Neighbors(cell.Index())...).Equal(Reg(expected...)), name+" at "+cell.Index().String()+": ")
			}
		}
	})
	t.Run("is the range of one step without its center", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				assert.True(t, Reg(g.Neighbors(index)...).Equal(g.Range(index, 1).Difference(Reg(index))), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("an index outside the grid has none, even beside it", func(t *testing.T) {
		for name, g := range grids {
			assert.Nil(t, g.Neighbors(geom.Pt(-1, 0)), name+": ")
			assert.Nil(t, g.Neighbors(geom.Pt(40, 40)), name+": ")
		}
	})
	t.Run("allocates once, at its length", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				assert.Equal(t, cap(g.Neighbors(index)), len(g.Neighbors(index)), name+" at "+index.String()+": ")
			}
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = g.Neighbors(center)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Neighbors(b *testing.B) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	for _, name := range []string{"cardinal", "diagonal", "flat-top", "pointy-top"} {
		b.Run(name, func(b *testing.B) {
			g := grids[name]

			for b.Loop() {
				sinkPoints = g.Neighbors(center)
			}
		})
	}
}

func FuzzGrid_Neighbors(f *testing.F) {
	for kind := range uint8(7) {
		f.Add(kind, uint8(5), uint8(4), uint8(3), uint8(3))
		f.Add(kind, uint8(0), uint8(0), uint8(2), uint8(2))
		f.Add(kind, uint8(9), uint8(1), uint8(0), uint8(1))
	}

	f.Fuzz(func(t *testing.T, kind, width, height, x, y uint8) {
		g := fuzzGrid(kind, width, height)
		index := fuzzIndex(g, x, y)
		neighbors := g.Neighbors(index)

		if !g.Contains(index) {
			assert.Nil(t, neighbors)

			return
		}

		var expected []Index
		for cell := range g.All() {
			if g.Distance(index, cell.Index()) == 1 {
				expected = append(expected, cell.Index())
			}
		}

		assert.Equal(t, len(neighbors), len(expected))
		assert.True(t, Reg(neighbors...).Equal(Reg(expected...)))
		assert.Equal(t, cap(neighbors), len(neighbors))

		for _, neighbor := range neighbors {
			assert.True(t, slices.Contains(g.Neighbors(neighbor), index), neighbor.String()+": ")
		}
	})
}

func TestGrid_NeighborsSeq(t *testing.T) {
	center := geom.Pt(5, 5)

	testSeq(t, func(g *Grid[int]) iter.Seq[Index] {
		return g.NeighborsSeq(center)
	}, func(g *Grid[int]) []Index {
		return g.Neighbors(center)
	})
	t.Run("allocates nothing, into a buffer with room too", func(t *testing.T) {
		buffer := make([]Index, 0, 128)

		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				for index := range g.NeighborsSeq(center) {
					sinkPoint = index
				}
			}), 0.0, name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = slices.AppendSeq(buffer[:0], g.NeighborsSeq(center))
			}), 0.0, name+": ")
		}
	})
	t.Run("yields nothing from outside the grid", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Length(t, slices.Collect(g.NeighborsSeq(geom.Pt(-1, 5))), 0, name+": ")
		}
	})
}

func TestGrid_Range(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	t.Run("a negative radius is empty", func(t *testing.T) {
		assert.True(t, grids["cardinal"].Range(center, -1).IsEmpty())
	})
	t.Run("a zero radius is the center alone", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Range(center, 0).Elements(), []Index{center})
	})
	t.Run("a radius beyond the grid is every cell, however far beyond", func(t *testing.T) {
		for name, g := range grids {
			across := g.Width() + g.Height()

			for _, radius := range []int{across, 1 << 40, math.MaxInt} {
				assert.Equal(t, g.Range(geom.Pt(0, 0), radius).Len(), g.Width()*g.Height(), name+" "+strconv.Itoa(radius)+": ")
			}
		}
	})
	t.Run("the farthest cell is in a range of any radius beyond it, from every cell of a small grid", func(t *testing.T) {
		for size := range container.Arr[int](geom.Sz(7, 7)).Keys() {
			size := geom.Sz(size.X+1, size.Y+1)
			small := kindGrids(size, geom.Sz(32.0, 20.0), geom.Pt(0.0, 0.0))
			small["diagonal"] = NewRect[int](size, geom.Sz(32.0, 32.0), withDiagonalMovement)

			for name, g := range small {
				for cell := range g.All() {
					assert.Equal(t, cell.Range(math.MaxInt).Len(), g.Width()*g.Height(), name+" "+size.String()+" from "+cell.Index().String()+": ")
				}
			}
		}
	})
	t.Run("cardinal is the diamond of Manhattan distance", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Range(center, 1).Len(), 5)
		assert.Equal(t, grids["cardinal"].Range(center, 3).Len(), 25)
	})
	t.Run("diagonal is the square of Chebyshev distance", func(t *testing.T) {
		assert.Equal(t, grids["diagonal"].Range(center, 1).Len(), 9)
		assert.Equal(t, grids["diagonal"].Range(center, 3).Len(), 49)
	})
	t.Run("is exactly the cells within the distance", func(t *testing.T) {
		for name, g := range grids {
			for n := range 5 {
				var expected []Index
				for cell := range g.All() {
					if g.Distance(center, cell.Index()) <= n {
						expected = append(expected, cell.Index())
					}
				}

				assert.True(t, g.Range(center, n).Equal(Reg(expected...)), name+" at "+strconv.Itoa(n)+": ")
			}
		}
	})
	t.Run("is empty from an index outside the grid, even beside it", func(t *testing.T) {
		for name, g := range grids {
			assert.True(t, g.Range(geom.Pt(-1, 5), 3).IsEmpty(), name+": ")
			assert.True(t, g.Range(geom.Pt(40, 40), 3).IsEmpty(), name+": ")
		}
	})
	t.Run("an empty range allocates nothing", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Range(geom.Pt(-1, 5), 4)
			}), 0.0, name+": ")
		}
	})
	t.Run("holds no index outside the grid", func(t *testing.T) {
		for name, g := range grids {
			for index := range g.Range(geom.Pt(0, 0), 2).All() {
				assert.True(t, g.Contains(index), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("allocates once, its indices", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Range(center, 4)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Range(b *testing.B) {
	for _, kind := range []struct {
		name string
		grid *Grid[int]
	}{
		{"rectangular", NewRect[int](geom.Sz(101, 101), geom.Sz(32.0, 32.0))},
		{"hexagonal", NewHexFlatTop[int](geom.Sz(101, 101), geom.Sz(32.0, 32.0))},
	} {
		for _, radius := range []int{8, 40} {
			b.Run(kind.name+" at "+strconv.Itoa(radius), func(b *testing.B) {
				for b.Loop() {
					sinkRegion = kind.grid.Range(geom.Pt(50, 50), radius)
				}
			})
		}
	}
}

func ExampleGrid_Range() {
	g := NewRect[bool](geom.Sz(5, 5), RectCellScale(32))
	index := geom.Pt(2, 2)

	fmt.Println(g.Distance(index, geom.Pt(4, 3)))
	fmt.Println(g.Range(index, 1).Elements())
	fmt.Println(g.Cell(index).Range(2).Equal(g.Range(index, 2)))
	// Output:
	// 3
	// [(1,2) (2,1) (2,2) (2,3) (3,2)]
	// true
}

func FuzzGrid_Range(f *testing.F) {
	for kind := range uint8(7) {
		f.Add(kind, uint8(8), uint8(8), uint8(6), uint8(6), uint8(3))
		f.Add(kind, uint8(0), uint8(0), uint8(2), uint8(2), uint8(2))
		f.Add(kind, uint8(9), uint8(3), uint8(0), uint8(1), uint8(4))
	}

	f.Fuzz(func(t *testing.T, kind, width, height, x, y, reach uint8) {
		g := fuzzGrid(kind, width, height)
		index, radius := fuzzIndex(g, x, y), int(reach%6)
		everything := g.Range(index, radius)

		assert.True(t, Reg(slices.Collect(g.RangeSeq(index, radius))...).Equal(everything))
		assert.True(t, slices.IsSortedFunc(slices.Collect(g.RangeSeq(index, radius)), Index.Compare))

		if !g.Contains(index) {
			assert.True(t, everything.IsEmpty())

			return
		}

		var within []Index
		for cell := range g.All() {
			if g.Distance(index, cell.Index()) <= radius {
				within = append(within, cell.Index())
			}
		}

		assert.True(t, everything.Equal(Reg(within...)))
		assert.Equal(t, cap(g.Spiral(index, radius)), everything.Len())
		assert.Equal(t, cap(g.Ring(index, radius)), len(g.Ring(index, radius)))
	})
}

func TestGrid_RangeSeq(t *testing.T) {
	center := geom.Pt(5, 5)

	testSeq(t, func(g *Grid[int]) iter.Seq[Index] {
		return g.RangeSeq(center, 3)
	}, func(g *Grid[int]) []Index {
		return g.Range(center, 3).Elements()
	})
	t.Run("allocates nothing, into a buffer with room too", func(t *testing.T) {
		buffer := make([]Index, 0, 128)

		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				for index := range g.RangeSeq(center, 3) {
					sinkPoint = index
				}
			}), 0.0, name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = slices.AppendSeq(buffer[:0], g.RangeSeq(center, 3))
			}), 0.0, name+": ")
		}
	})
	t.Run("a radius beyond the grid yields what one across it does", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Equal(t, slices.Collect(g.RangeSeq(center, math.MaxInt)), slices.Collect(g.RangeSeq(center, g.Width()+g.Height())), name+": ")
		}
	})
	t.Run("yields nothing for a negative radius or from outside the grid", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Length(t, slices.Collect(g.RangeSeq(center, -1)), 0, name+": ")
			assert.Length(t, slices.Collect(g.RangeSeq(geom.Pt(-1, 5), 3)), 0, name+": ")
		}
	})
}

func TestGrid_Ring(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	t.Run("cardinal walks the diamond, by increasing angle", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Ring(center, 1), grids["cardinal"].Neighbors(center))
	})
	t.Run("a radius of zero is the index alone", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, g.Ring(center, 0), []Index{center}, name+": ")
		}
	})
	t.Run("is exactly the cells at the distance of the radius", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				for radius := range 6 {
					var expected []Index
					for other := range g.All() {
						if g.Distance(index, other.Index()) == radius {
							expected = append(expected, other.Index())
						}
					}

					ring := g.Ring(index, radius)

					assert.Length(t, ring, len(expected), name+" at "+index.String()+": ")
					assert.True(t, Reg(ring...).Equal(Reg(expected...)), name+" at "+index.String()+": ")
				}
			}
		}
	})
	t.Run("a radius no cell is at has none", func(t *testing.T) {
		for name, g := range grids {
			assert.Nil(t, g.Ring(center, 40), name+": ")
			assert.Nil(t, g.Ring(center, math.MaxInt), name+": ")
		}
	})
	t.Run("a negative radius has none", func(t *testing.T) {
		for name, g := range grids {
			assert.Nil(t, g.Ring(center, -1), name+": ")
		}
	})
	t.Run("an index outside the grid has none, even beside it", func(t *testing.T) {
		for name, g := range grids {
			assert.Nil(t, g.Ring(geom.Pt(-1, 5), 2), name+": ")
		}
	})
	t.Run("allocates once, at its length", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				for radius := range 8 {
					assert.Equal(t, cap(g.Ring(index, radius)), len(g.Ring(index, radius)), name+" at "+index.String()+": ")
				}
			}
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = g.Ring(center, 3)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Ring(b *testing.B) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	for _, name := range []string{"cardinal", "diagonal", "flat-top", "pointy-top"} {
		b.Run(name, func(b *testing.B) {
			g := grids[name]

			for b.Loop() {
				sinkPoints = g.Ring(center, 4)
			}
		})
	}
}

func ExampleGrid_Ring() {
	g := NewHexFlatTop[bool](geom.Sz(5, 5), HexFlatTopCellScale(32))

	index := geom.Pt(1, 1)

	fmt.Println(g.Ring(index, 1))
	fmt.Println(g.Ring(index, 2))
	fmt.Println(g.Spiral(index, 1))
	fmt.Println(g.Line(geom.Pt(0, 0), geom.Pt(4, 2)))
	// Output:
	// [(2,2) (1,2) (0,2) (0,1) (1,0) (2,1)]
	// [(3,2) (2,3) (1,3) (0,3) (0,0) (2,0) (3,0) (3,1)]
	// [(1,1) (2,2) (1,2) (0,2) (0,1) (1,0) (2,1)]
	// [(0,0) (1,0) (2,1) (3,1) (4,2)]
}

func TestGrid_RingSeq(t *testing.T) {
	center := geom.Pt(5, 5)

	testSeq(t, func(g *Grid[int]) iter.Seq[Index] {
		return g.RingSeq(center, 3)
	}, func(g *Grid[int]) []Index {
		return g.Ring(center, 3)
	})
	t.Run("allocates nothing, into a buffer with room too", func(t *testing.T) {
		buffer := make([]Index, 0, 128)

		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				for index := range g.RingSeq(center, 3) {
					sinkPoint = index
				}
			}), 0.0, name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = slices.AppendSeq(buffer[:0], g.RingSeq(center, 3))
			}), 0.0, name+": ")
		}
	})
	t.Run("yields nothing where Ring has none", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Length(t, slices.Collect(g.RingSeq(center, -1)), 0, name+": ")
			assert.Length(t, slices.Collect(g.RingSeq(center, 40)), 0, name+": ")
		}
	})
}

func TestGrid_Spiral(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	t.Run("starts at the index", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, g.Spiral(center, 3)[0], center, name+": ")
		}
	})
	t.Run("is the rings in turn, outward", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				var expected []Index
				for radius := range 5 {
					expected = slices.AppendSeq(expected, g.RingSeq(index, radius))
				}

				assert.Equal(t, g.Spiral(index, 4), expected, name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("holds the cells of the range", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				assert.True(t, Reg(g.Spiral(index, 4)...).Equal(g.Range(index, 4)), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("a radius beyond the grid is the whole grid", func(t *testing.T) {
		for name, g := range grids {
			assert.Length(t, g.Spiral(center, math.MaxInt), g.Width()*g.Height(), name+": ")
		}
	})
	t.Run("a negative radius has none", func(t *testing.T) {
		for name, g := range grids {
			assert.Nil(t, g.Spiral(center, -1), name+": ")
		}
	})
	t.Run("an index outside the grid has none, even beside it", func(t *testing.T) {
		for name, g := range grids {
			assert.Nil(t, g.Spiral(geom.Pt(-1, 5), 2), name+": ")
		}
	})
	t.Run("allocates once, at its length", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				for radius := range 8 {
					assert.Equal(t, cap(g.Spiral(index, radius)), len(g.Spiral(index, radius)), name+" at "+index.String()+": ")
				}
			}
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = g.Spiral(center, 3)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Spiral(b *testing.B) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	for _, name := range []string{"cardinal", "diagonal", "flat-top", "pointy-top"} {
		b.Run(name, func(b *testing.B) {
			g := grids[name]

			for b.Loop() {
				sinkPoints = g.Spiral(center, 4)
			}
		})
	}
}

func spiralBreakAllocs(g *Grid[int], index Index, radius int) float64 {
	return testing.AllocsPerRun(10, func() {
		for found := range g.SpiralSeq(index, radius) {
			sinkPoint = found

			break
		}
	})
}

func TestGrid_SpiralSeq(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	testSeq(t, func(g *Grid[int]) iter.Seq[Index] {
		return g.SpiralSeq(center, 3)
	}, func(g *Grid[int]) []Index {
		return g.Spiral(center, 3)
	})
	t.Run("allocates nothing, into a buffer with room too", func(t *testing.T) {
		buffer := make([]Index, 0, 128)

		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				for index := range g.SpiralSeq(center, 3) {
					sinkPoint = index
				}
			}), 0.0, name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = slices.AppendSeq(buffer[:0], g.SpiralSeq(center, 3))
			}), 0.0, name+": ")
		}
	})

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("yields what Spiral returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(g.SpiralSeq(geom.Pt(1, 1), 3)), g.Spiral(geom.Pt(1, 1), 3))
			})
			t.Run("yields nothing where Spiral has none", func(t *testing.T) {
				assert.Length(t, slices.Collect(g.SpiralSeq(center, -1)), 0)
			})
			t.Run("stops at the break", func(t *testing.T) {
				count := 0
				for range g.SpiralSeq(center, 3) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
			t.Run("walks no further than the break, whatever the radius", func(t *testing.T) {
				near, far := spiralBreakAllocs(g, center, 1), spiralBreakAllocs(g, center, g.Width()+g.Height())

				assert.Equal(t, far, near)
			})
			t.Run("yields no index outside the grid and every one inside it, from a corner", func(t *testing.T) {
				assert.Equal(t, slices.Collect(g.SpiralSeq(geom.Pt(0, 0), math.MaxInt)), g.Spiral(geom.Pt(0, 0), math.MaxInt))
			})
		})
	}
}

func TestGrid_Border(t *testing.T) {
	center := geom.Pt(4, 4)

	t.Run("a single cell is its own border", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			assert.True(t, g.Border(Reg(center)).Equal(Reg(center)), name+": ")
		}
	})
	t.Run("the border of a range is its outermost ring", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			var ring []Index
			for index := range g.Range(center, 3).All() {
				if g.Distance(center, index) == 3 {
					ring = append(ring, index)
				}
			}

			assert.True(t, g.Border(g.Range(center, 3)).Equal(Reg(ring...)), name+": ")
		}
	})
	t.Run("a hole has a border of its own", func(t *testing.T) {
		g := ruleGrids(geom.Sz(9, 9))["cardinal"]
		ringed := g.Range(center, 2).Difference(Reg(center))

		assert.True(t, g.Border(ringed).Contains(geom.Pt(4, 3)))
		assert.False(t, g.Border(g.Range(center, 2)).Contains(geom.Pt(4, 3)))
	})
	t.Run("the edge of the grid is a border", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			var whole, edge []Index
			for cell := range g.All() {
				whole = append(whole, cell.Index())
				if index := cell.Index(); index.X == 0 || index.Y == 0 || index.X == 8 || index.Y == 8 {
					edge = append(edge, index)
				}
			}

			assert.True(t, g.Border(Reg(whole...)).Equal(Reg(edge...)), name+": ")
		}
	})
	t.Run("the empty region has none", func(t *testing.T) {
		assert.True(t, ruleGrids(geom.Sz(9, 9))["cardinal"].Border(Region{}).IsEmpty())
	})
	t.Run("is held by the region", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			region := g.Range(center, 3)

			assert.True(t, region.Encloses(g.Border(region)), name+": ")
		}
	})
	t.Run("allocates once", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			region := g.Range(center, 3)

			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Border(region)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Border(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		region := g.Range(geom.Pt(32, 32), 16)

		for b.Loop() {
			sinkRegion = g.Border(region)
		}
	})
}

func ExampleGrid_Border() {
	g := NewRect[int](geom.Sz(5, 5), RectCellScale(32))
	room := g.Range(geom.Pt(2, 2), 1)

	fmt.Println(g.Border(room).Len(), g.Grow(room, 1).Len(), g.Shrink(room, 1).Elements())
	fmt.Println(g.IsConnected(room), len(g.Components(Reg(geom.Pt(0, 0), geom.Pt(4, 4)))))
	// Output:
	// 4 13 [(2,2)]
	// true 2
}

func TestGrid_Grow(t *testing.T) {
	center := geom.Pt(4, 4)
	single := Reg(center)

	t.Run("a cell grown is the range around it", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			for n := range 4 {
				assert.True(t, g.Grow(single, n).Equal(g.Range(center, n)), name+": ")
			}
		}
	})
	t.Run("zero or fewer steps is the region", func(t *testing.T) {
		g := ruleGrids(geom.Sz(9, 9))["cardinal"]

		assert.True(t, g.Grow(single, 0).Equal(single))
		assert.True(t, g.Grow(single, -2).Equal(single))
	})
	t.Run("stops at the edge of the grid", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			for index := range g.Grow(Reg(geom.Pt(0, 0)), 3).All() {
				assert.True(t, g.Contains(index), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("fills the whole grid and no more", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			assert.Equal(t, g.Grow(single, 100).Len(), 81, name+": ")
		}
	})
	t.Run("keeps an index of the region outside the grid", func(t *testing.T) {
		outside := Reg(geom.Pt(-5, -5), center)

		assert.True(t, ruleGrids(geom.Sz(9, 9))["cardinal"].Grow(outside, 1).Contains(geom.Pt(-5, -5)))
	})
	t.Run("the empty region stays empty", func(t *testing.T) {
		assert.True(t, ruleGrids(geom.Sz(9, 9))["cardinal"].Grow(Region{}, 3).IsEmpty())
	})
	t.Run("allocates its marks once and its indices as they double", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			assert.LessOrEqual(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Grow(single, 3)
			}), 10.0, name+": ")
		}
	})
}

func BenchmarkGrid_Grow(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		region := g.Range(geom.Pt(32, 32), 8)

		for b.Loop() {
			sinkRegion = g.Grow(region, 4)
		}
	})
}

func TestGrid_Shrink(t *testing.T) {
	center := geom.Pt(4, 4)

	t.Run("undoes a grow away from the edge", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			region := g.Range(center, 1)

			assert.True(t, g.Shrink(g.Grow(region, 2), 2).Equal(region), name+": ")
		}
	})
	t.Run("a range shrunk is the smaller range", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			assert.True(t, g.Shrink(g.Range(center, 3), 2).Equal(g.Range(center, 1)), name+": ")
		}
	})
	t.Run("zero or fewer steps is the region", func(t *testing.T) {
		g := ruleGrids(geom.Sz(9, 9))["cardinal"]
		region := g.Range(center, 2)

		assert.True(t, g.Shrink(region, 0).Equal(region))
		assert.True(t, g.Shrink(region, -1).Equal(region))
	})
	t.Run("more steps than the region is deep leaves nothing", func(t *testing.T) {
		g := ruleGrids(geom.Sz(9, 9))["cardinal"]

		assert.True(t, g.Shrink(g.Range(center, 2), 10).IsEmpty())
	})
	t.Run("allocates twice a step", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			region := g.Range(center, 3)

			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Shrink(region, 2)
			}), 4.0, name+": ")
		}
	})
}

func BenchmarkGrid_Shrink(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		region := g.Range(geom.Pt(32, 32), 16)

		for b.Loop() {
			sinkRegion = g.Shrink(region, 4)
		}
	})
}

func TestGrid_IsConnected(t *testing.T) {
	t.Run("the empty region and a single cell are", func(t *testing.T) {
		g := ruleGrids(geom.Sz(9, 9))["cardinal"]

		assert.True(t, g.IsConnected(Region{}))
		assert.True(t, g.IsConnected(Reg(geom.Pt(4, 4))))
	})
	t.Run("a range is", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			assert.True(t, g.IsConnected(g.Range(geom.Pt(4, 4), 3)), name+": ")
		}
	})
	t.Run("two cells apart are not", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			assert.False(t, g.IsConnected(Reg(geom.Pt(1, 1), geom.Pt(6, 6))), name+": ")
		}
	})
	t.Run("cells touching at a corner are connected under diagonal movement alone", func(t *testing.T) {
		corner := Reg(geom.Pt(1, 1), geom.Pt(2, 2))

		assert.False(t, ruleGrids(geom.Sz(9, 9))["cardinal"].IsConnected(corner))
		assert.True(t, ruleGrids(geom.Sz(9, 9))["diagonal"].IsConnected(corner))
	})
	t.Run("allocates twice", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			region := g.Range(geom.Pt(4, 4), 4)

			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkBool = g.IsConnected(region)
			}), 2.0, name+": ")
		}
	})
}

func BenchmarkGrid_IsConnected(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		region := g.Range(geom.Pt(32, 32), 16)

		for b.Loop() {
			sinkBool = g.IsConnected(region)
		}
	})
}

func TestGrid_Components(t *testing.T) {
	t.Run("the empty region has none", func(t *testing.T) {
		assert.Length(t, ruleGrids(geom.Sz(9, 9))["cardinal"].Components(Region{}), 0)
	})
	t.Run("a connected region is its one component", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			region := g.Range(geom.Pt(4, 4), 2)
			components := g.Components(region)

			assert.Length(t, components, 1, name+": ")
			assert.True(t, components[0].Equal(region), name+": ")
		}
	})
	t.Run("parts apart come back in the order of their first index", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			left, right := g.Range(geom.Pt(1, 1), 1), g.Range(geom.Pt(6, 6), 1)
			components := g.Components(left.Union(right))

			assert.Length(t, components, 2, name+": ")
			assert.True(t, components[0].Equal(left), name+": ")
			assert.True(t, components[1].Equal(right), name+": ")
		}
	})
	t.Run("every component is connected and together they are the region", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			var scattered []Index
			for cell := range g.All() {
				if index := cell.Index(); (index.X*5+index.Y*3)%4 != 0 {
					scattered = append(scattered, index)
				}
			}
			region := Reg(scattered...)

			union := Region{}
			for _, component := range g.Components(region) {
				assert.True(t, g.IsConnected(component), name+": ")
				assert.False(t, union.Intersects(component), name+": ")
				union = union.Union(component)
			}

			assert.True(t, union.Equal(region), name+": ")
		}
	})
	t.Run("allocates twice, then once a component and once for the list of one", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			region := g.Range(geom.Pt(4, 4), 4)

			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegions = g.Components(region)
			}), 4.0, name+": ")
		}
	})
}

func BenchmarkGrid_Components(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		region := g.Range(geom.Pt(16, 16), 8).Union(g.Range(geom.Pt(48, 48), 8))

		for b.Loop() {
			sinkRegions = g.Components(region)
		}
	})
}

func TestGrid_Distance(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	t.Run("cardinal is the Manhattan distance", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Distance(geom.Pt(0, 0), geom.Pt(3, 4)), 7)
	})
	t.Run("diagonal is the Chebyshev distance", func(t *testing.T) {
		assert.Equal(t, grids["diagonal"].Distance(geom.Pt(0, 0), geom.Pt(3, 4)), 4)
	})
	t.Run("a hexagon counts the steps across its offset rows and columns", func(t *testing.T) {
		assert.Equal(t, grids["flat-top"].Distance(geom.Pt(0, 0), geom.Pt(4, 0)), 4)
		assert.Equal(t, grids["flat-top"].Distance(geom.Pt(0, 0), geom.Pt(2, 1)), 2)
		assert.Equal(t, grids["pointy-top"].Distance(geom.Pt(0, 0), geom.Pt(0, 4)), 4)
		assert.Equal(t, grids["pointy-top"].Distance(geom.Pt(0, 0), geom.Pt(1, 2)), 2)
	})
	t.Run("is zero to itself alone and the same both ways", func(t *testing.T) {
		for name, g := range grids {
			for cell := range g.All() {
				message := name + " at " + cell.Index().String() + ": "

				assert.Equal(t, g.Distance(center, cell.Index()) == 0, cell.Index() == center, message)
				assert.Equal(t, g.Distance(center, cell.Index()), g.Distance(cell.Index(), center), message)
			}
		}
	})
	t.Run("a step to a neighbor changes it by one at most", func(t *testing.T) {
		for name, g := range grids {
			for cell := range g.All() {
				for _, neighbor := range g.Neighbors(cell.Index()) {
					change := g.Distance(center, neighbor) - g.Distance(center, cell.Index())

					assert.True(t, change >= -1 && change <= 1, name+" at "+cell.Index().String()+": ")
				}
			}
		}
	})
	t.Run("measures between indices outside the grid too", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Distance(geom.Pt(-2, 0), geom.Pt(20, 0)), 22)
	})
}

func TestGrid_Line(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	t.Run("a line to itself is the index alone", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, g.Line(center, center), []Index{center}, name+": ")
		}
	})
	t.Run("steps from one end to the other by neighbors", func(t *testing.T) {
		for name, g := range grids {
			for from := range g.All() {
				for _, to := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
					line := g.Line(from.Index(), to)

					assert.Length(t, line, g.Distance(from.Index(), to)+1, name+" from "+from.Index().String()+": ")
					assert.Equal(t, line[0], from.Index(), name+" from "+from.Index().String()+": ")
					assert.Equal(t, line[len(line)-1], to, name+" from "+from.Index().String()+": ")
					for i := 1; i < len(line); i++ {
						assert.Equal(t, g.Distance(line[i-1], line[i]), 1, name+" from "+from.Index().String()+": ")
					}
				}
			}
		}
	})
	t.Run("stays inside a square grid", func(t *testing.T) {
		for _, name := range []string{"cardinal", "diagonal"} {
			g := grids[name]
			for from := range g.All() {
				for _, index := range g.Line(from.Index(), geom.Pt(10, 3)) {
					assert.True(t, g.Contains(index), name+" from "+from.Index().String()+": ")
				}
			}
		}
	})
	t.Run("is measured outside the grid too", func(t *testing.T) {
		for name, g := range grids {
			assert.Length(t, g.Line(geom.Pt(-3, -3), geom.Pt(-3, 0)), g.Distance(geom.Pt(-3, -3), geom.Pt(-3, 0))+1, name+": ")
		}
	})
	t.Run("keeps its steps along a zigzag edge of a hexagonal grid", func(t *testing.T) {
		for name, g := range hexGrids() {
			for from := range g.All() {
				for to := range g.All() {
					assert.Length(t, g.Line(from.Index(), to.Index()), g.Distance(from.Index(), to.Index())+1, name+": ")
				}
			}
		}
	})
	t.Run("allocates once, at the count of its steps", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, cap(g.Line(geom.Pt(0, 0), geom.Pt(10, 3))), len(g.Line(geom.Pt(0, 0), geom.Pt(10, 3))), name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = g.Line(geom.Pt(0, 0), geom.Pt(10, 3))
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Line(b *testing.B) {
	grids := ruleGrids(geom.Sz(11, 11))

	for _, name := range []string{"cardinal", "diagonal", "flat-top", "pointy-top"} {
		b.Run(name, func(b *testing.B) {
			g := grids[name]

			for b.Loop() {
				sinkPoints = g.Line(geom.Pt(0, 0), geom.Pt(10, 3))
			}
		})
	}
}

func TestGrid_LineSeq(t *testing.T) {
	center := geom.Pt(5, 5)

	testSeq(t, func(g *Grid[int]) iter.Seq[Index] {
		return g.LineSeq(center, geom.Pt(1, 9))
	}, func(g *Grid[int]) []Index {
		return g.Line(center, geom.Pt(1, 9))
	})
	t.Run("allocates nothing, into a buffer with room too", func(t *testing.T) {
		buffer := make([]Index, 0, 128)

		for name, g := range ruleGrids(geom.Sz(11, 11)) {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				for index := range g.LineSeq(center, geom.Pt(1, 9)) {
					sinkPoint = index
				}
			}), 0.0, name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = slices.AppendSeq(buffer[:0], g.LineSeq(center, geom.Pt(1, 9)))
			}), 0.0, name+": ")
		}
	})
}

// testSearch runs the sections every search shares, from testStart to testGoal.
func wallAt(index Index) func(Cell[int]) bool {
	return func(cell Cell[int]) bool {
		return cell.Index() == index
	}
}

func scatteredWalls(seed int) func(Cell[int]) bool {
	return func(cell Cell[int]) bool {
		return (cell.Index().X*7+cell.Index().Y*13+seed)%5 == 2
	}
}

func TestGrid_HasLineOfSight(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	t.Run("a nil blocked sees between any two cells", func(t *testing.T) {
		for name, g := range grids {
			assert.True(t, g.HasLineOfSight(center, geom.Pt(0, 9), nil), name+": ")
		}
	})
	t.Run("a wall on the line hides what is behind it", func(t *testing.T) {
		for _, name := range []string{"cardinal", "diagonal", "flat-top"} {
			assert.False(t, grids[name].HasLineOfSight(center, geom.Pt(5, 7), wallAt(geom.Pt(5, 6))), name+": ")
		}
	})
	t.Run("neither end blocks", func(t *testing.T) {
		ends := func(cell Cell[int]) bool {
			return cell.Index() == center || cell.Index() == geom.Pt(5, 9)
		}

		for name, g := range grids {
			assert.True(t, g.HasLineOfSight(center, geom.Pt(5, 9), ends), name+": ")
		}
	})
	t.Run("is false from or to an index outside the grid", func(t *testing.T) {
		for name, g := range grids {
			assert.False(t, g.HasLineOfSight(geom.Pt(-1, 5), center, nil), name+": ")
			assert.False(t, g.HasLineOfSight(center, geom.Pt(40, 5), nil), name+": ")
		}
	})
	t.Run("what lies outside the grid blocks nothing and is never asked about", func(t *testing.T) {
		for name, g := range grids {
			for from := range g.All() {
				for to := range g.All() {
					assert.True(t, g.HasLineOfSight(from.Index(), to.Index(), func(cell Cell[int]) bool {
						assert.True(t, cell.IsContained(), name+": ")

						return false
					}), name+": ")
				}
			}
		}
	})
}

func TestGrid_FieldOfView(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)
	wall := wallAt(geom.Pt(5, 6))

	t.Run("a negative radius is empty", func(t *testing.T) {
		assert.True(t, grids["cardinal"].FieldOfView(center, -1, nil).IsEmpty())
	})
	t.Run("a nil blocked is Range", func(t *testing.T) {
		for name, g := range grids {
			assert.True(t, g.FieldOfView(center, 4, nil).Equal(g.Range(center, 4)), name+": ")
		}
	})
	t.Run("a radius beyond the grid still hides what is behind a wall", func(t *testing.T) {
		for name, g := range grids {
			assert.True(t, g.FieldOfView(center, math.MaxInt, wall).Equal(g.FieldOfView(center, g.Width()+g.Height(), wall)), name+": ")
		}
	})
	t.Run("a wall is in sight and hides what is behind it", func(t *testing.T) {
		for _, name := range []string{"cardinal", "diagonal", "flat-top"} {
			seen := grids[name].FieldOfView(center, 2, wall)

			assert.True(t, seen.Contains(geom.Pt(5, 6)), name+": ")
			assert.False(t, seen.Contains(geom.Pt(5, 7)), name+": ")
			assert.True(t, seen.Contains(geom.Pt(5, 4)), name+": ")
		}
	})
	t.Run("a diagonal neighbor behind two blockers is hidden, as HasLineOfSight says", func(t *testing.T) {
		walls := func(cell Cell[int]) bool {
			return cell.Index() == geom.Pt(6, 5) || cell.Index() == geom.Pt(5, 6)
		}

		assert.False(t, grids["diagonal"].FieldOfView(center, 2, walls).Contains(geom.Pt(6, 6)))
		assert.False(t, grids["diagonal"].HasLineOfSight(center, geom.Pt(6, 6), walls))
	})
	t.Run("the cell of the center is never asked about, so it sees from a wall", func(t *testing.T) {
		for name, g := range grids {
			seen := g.FieldOfView(center, 3, func(cell Cell[int]) bool {
				assert.NotEqual(t, cell.Index(), center, name+": ")

				return false
			})

			assert.True(t, seen.Equal(g.Range(center, 3)), name+": ")
		}
	})
	t.Run("sight is reciprocal", func(t *testing.T) {
		for name, g := range grids {
			for seen := range g.FieldOfView(center, 4, scatteredWalls(0)).All() {
				assert.True(t, g.FieldOfView(seen, 4, scatteredWalls(0)).Contains(center), name+" at "+seen.String()+": ")
			}
		}
	})
	t.Run("walls that hide nothing change nothing", func(t *testing.T) {
		never := func(Cell[int]) bool {
			return false
		}

		for name, g := range grids {
			for _, index := range []Index{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				assert.True(t, g.FieldOfView(index, 4, nil).Equal(g.FieldOfView(index, 4, never)), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("is empty from an index outside the grid", func(t *testing.T) {
		for name, g := range grids {
			assert.True(t, g.FieldOfView(geom.Pt(-1, 5), 3, nil).IsEmpty(), name+": ")
		}
	})
	t.Run("an empty range allocates nothing", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.FieldOfView(geom.Pt(-1, 5), 4, nil)
			}), 0.0, name+": ")
		}
	})
	t.Run("asks about every cell at most once, whatever is hidden", func(t *testing.T) {
		for name, g := range grids {
			asked := map[Index]int{}
			g.FieldOfView(center, 4, func(cell Cell[int]) bool {
				asked[cell.Index()]++

				return scatteredWalls(0)(cell)
			})

			for index, times := range asked {
				assert.Equal(t, times, 1, name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("allocates once, its indices, where a nil blocked hides nothing", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.FieldOfView(center, 4, nil)
			}), 1.0, name+": ")
		}
	})
	t.Run("allocates once more where a cell is refused, for the marks, and once more on a hexagonal grid for its sweep", func(t *testing.T) {
		allocs := map[string]float64{"cardinal": 2, "diagonal": 2, "flat-top": 3, "pointy-top": 3}

		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.FieldOfView(center, 4, wall)
			}), allocs[name], name+": ")
		}
	})
}

func BenchmarkGrid_FieldOfView(b *testing.B) {
	walls := func(cell Cell[int]) bool {
		return (cell.Index().X*7+cell.Index().Y*13)%11 == 3
	}
	never := func(Cell[int]) bool {
		return false
	}

	for _, kind := range []struct {
		name string
		grid *Grid[int]
	}{
		{"rectangular", NewRect[int](geom.Sz(101, 101), geom.Sz(32.0, 32.0))},
		{"hexagonal", NewHexFlatTop[int](geom.Sz(101, 101), geom.Sz(32.0, 32.0))},
	} {
		for _, radius := range []int{8, 40} {
			b.Run(kind.name+" at "+strconv.Itoa(radius)+" where nothing blocks", func(b *testing.B) {
				for b.Loop() {
					sinkRegion = kind.grid.FieldOfView(geom.Pt(50, 50), radius, never)
				}
			})
			b.Run(kind.name+" at "+strconv.Itoa(radius)+" with walls", func(b *testing.B) {
				for b.Loop() {
					sinkRegion = kind.grid.FieldOfView(geom.Pt(50, 50), radius, walls)
				}
			})
		}
	}
}

func ExampleGrid_FieldOfView() {
	g := NewRect[bool](geom.Sz(5, 5), RectCellScale(32))
	g.Cell(geom.Pt(2, 1)).Set(true)

	opaque := func(cell Cell[bool]) bool {
		return *cell.Value()
	}

	index := geom.Pt(2, 2)

	fmt.Println(g.FieldOfView(index, 2, opaque).Elements())
	fmt.Println(g.HasLineOfSight(index, geom.Pt(2, 0), opaque))
	// Output:
	// [(0,2) (1,1) (1,2) (1,3) (2,1) (2,2) (2,3) (2,4) (3,1) (3,2) (3,3) (4,2)]
	// false
}

func FuzzGrid_FieldOfView(f *testing.F) {
	for kind := range uint8(7) {
		f.Add(kind, uint8(8), uint8(8), uint8(6), uint8(6), uint8(3), uint8(5))
		f.Add(kind, uint8(0), uint8(0), uint8(2), uint8(2), uint8(2), uint8(0))
		f.Add(kind, uint8(9), uint8(3), uint8(0), uint8(1), uint8(4), uint8(17))
	}

	f.Fuzz(func(t *testing.T, kind, width, height, x, y, reach, seed uint8) {
		g := fuzzGrid(kind, width, height)
		index, radius := fuzzIndex(g, x, y), int(reach%6)
		blocked := scatteredWalls(int(seed))
		seen := g.FieldOfView(index, radius, blocked)

		assert.True(t, Reg(slices.Collect(g.FieldOfViewSeq(index, radius, blocked))...).Equal(seen))

		if !g.Contains(index) {
			assert.True(t, seen.IsEmpty())

			return
		}

		assert.True(t, seen.Contains(index))

		for other := range g.Range(index, radius).All() {
			in := seen.Contains(other)

			assert.Equal(t, in, g.HasLineOfSight(index, other, blocked), other.String()+": ")
			assert.Equal(t, in, g.FieldOfView(other, radius, blocked).Contains(index), other.String()+": ")
		}
	})
}

func TestGrid_FieldOfViewSeq(t *testing.T) {
	grids, center := ruleGrids(geom.Sz(11, 11)), geom.Pt(5, 5)

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("yields what FieldOfView holds", func(t *testing.T) {
				assert.True(t, Reg(slices.Collect(g.FieldOfViewSeq(center, 4, scatteredWalls(0)))...).Equal(g.FieldOfView(center, 4, scatteredWalls(0))))
			})
			t.Run("a nil blocked yields the range", func(t *testing.T) {
				assert.True(t, Reg(slices.Collect(g.FieldOfViewSeq(center, 3, nil))...).Equal(g.Range(center, 3)))
			})
			t.Run("yields no index outside the grid", func(t *testing.T) {
				for index := range g.FieldOfViewSeq(geom.Pt(0, 0), 3, nil) {
					assert.True(t, g.Contains(index), index.String()+": ")
				}
			})
			t.Run("a radius beyond the grid yields what one across it does", func(t *testing.T) {
				assert.True(t, Reg(slices.Collect(g.FieldOfViewSeq(center, math.MaxInt, wallAt(geom.Pt(5, 6))))...).Equal(g.FieldOfView(center, g.Width()+g.Height(), wallAt(geom.Pt(5, 6)))))
			})
			t.Run("yields nothing for a negative radius or from outside the grid", func(t *testing.T) {
				assert.Length(t, slices.Collect(g.FieldOfViewSeq(center, -1, nil)), 0)
				assert.Length(t, slices.Collect(g.FieldOfViewSeq(geom.Pt(-1, 5), 3, nil)), 0)
			})
			t.Run("stops at the break, with walls and without", func(t *testing.T) {
				for _, blocked := range []func(Cell[int]) bool{nil, scatteredWalls(0)} {
					count := 0
					for range g.FieldOfViewSeq(center, 4, blocked) {
						count++

						break
					}

					assert.Equal(t, count, 1)
				}
			})
			t.Run("asks about every cell at most once", func(t *testing.T) {
				asked := map[Index]int{}
				sinkPoints = slices.Collect(g.FieldOfViewSeq(center, 4, func(cell Cell[int]) bool {
					asked[cell.Index()]++

					return scatteredWalls(0)(cell)
				}))

				for index, times := range asked {
					assert.Equal(t, times, 1, index.String()+": ")
				}
			})
			t.Run("allocates nothing for a nil blocked, and once for what it remembers otherwise", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					for index := range g.FieldOfViewSeq(center, 4, nil) {
						sinkPoint = index
					}
				}), 0.0)

				wall := wallAt(geom.Pt(5, 6))

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					for index := range g.FieldOfViewSeq(center, 4, wall) {
						sinkPoint = index
					}
				}), 1.0)
			})
		})
	}
}

func testCost(t *testing.T, search func(g *Grid[int], cost func(current, next Cell[int]) float64) []Index) {
	t.Helper()

	g := NewRect[int](geom.Sz(4, 4), geom.SzU(32.0))

	t.Run("panics for a negative cost", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			search(g, func(_, _ Cell[int]) float64 {
				return -0.5
			})
		}, ErrInvalidCost)
	})
	t.Run("panics for a NaN cost", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			search(g, func(_, _ Cell[int]) float64 {
				return math.NaN()
			})
		}, ErrInvalidCost)
	})
	t.Run("an infinite cost is a step never taken", func(t *testing.T) {
		path := search(g, func(_, next Cell[int]) float64 {
			if next.Index().X == 1 && next.Index().Y < 3 {
				return math.Inf(1)
			}

			return 1
		})

		assert.True(t, slices.Contains(path, geom.Pt(1, 3)))
		assert.False(t, slices.Contains(path, geom.Pt(1, 0)))
	})
	t.Run("a wall of infinite cost is no path", func(t *testing.T) {
		assert.Nil(t, search(g, func(_, next Cell[int]) float64 {
			if next.Index().X == 1 {
				return math.Inf(1)
			}

			return 1
		}))
	})
}

func testSearch(t *testing.T, search func(g *Grid[int], passable func(Cell[int]) bool) []Index) {
	t.Run("nil callbacks admit every cell at unit cost", func(t *testing.T) {
		path := search(walledGrid(), nil)

		assert.Equal(t, path[0], testStart)
		assert.Equal(t, path[len(path)-1], testGoal)
	})
	t.Run("the path steps between neighbors and around the wall", func(t *testing.T) {
		g := walledGrid()
		path := search(g, walkable)

		assert.NotContains(t, path, testWall)
		for i := 1; i < len(path); i++ {
			assert.Equal(t, g.Distance(path[i-1], path[i]), 1)
		}
	})
	t.Run("a goal walled off has no path", func(t *testing.T) {
		assert.Equal(t, search(sealedGrid(), walkable), nil)
	})
	t.Run("a start or a goal outside the grid has no path", func(t *testing.T) {
		g := NewRect[int](geom.Sz(3, 3), geom.Sz(32.0, 32.0))

		assert.Equal(t, search(g, nil), nil)
	})
	t.Run("finds its way on every kind of grid", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			path := search(g, nil)

			assert.Equal(t, path[0], testStart, name+": ")
			assert.Equal(t, path[len(path)-1], testGoal, name+": ")
			for i := 1; i < len(path); i++ {
				assert.Equal(t, g.Distance(path[i-1], path[i]), 1, name+": ")
			}
		}
	})
	t.Run("allocates no more on a long path than on a short one, but for its buffers", func(t *testing.T) {
		small := NewRect[int](geom.Sz(8, 8), geom.Sz(32.0, 32.0))
		large := NewRect[int](geom.Sz(64, 64), geom.Sz(32.0, 32.0))

		short := testing.AllocsPerRun(5, func() {
			sinkPoints = search(small, walkable)
		})
		long := testing.AllocsPerRun(5, func() {
			sinkPoints = search(large, walkable)
		})

		assert.LessOrEqual(t, short, 12.0)
		assert.LessOrEqual(t, long, 24.0)
	})
}

// testSearchField runs the sections every search field shares, from testStart.
func testSearchField(t *testing.T, field func(g *Grid[int], passable func(Cell[int]) bool) Field) {
	t.Run("nil callbacks reach every cell", func(t *testing.T) {
		assert.Equal(t, field(walledGrid(), nil).Len(), 25)
	})
	t.Run("holds the start", func(t *testing.T) {
		reached := field(walledGrid(), walkable)

		assert.Equal(t, reached.Start(), testStart)
		assert.True(t, reached.Contains(testStart))
	})
	t.Run("a wall is never in the field", func(t *testing.T) {
		reached := field(walledGrid(), walkable)

		assert.False(t, reached.Contains(testWall))
		assert.Equal(t, reached.Len(), 24)
	})
	t.Run("holds only what was reached", func(t *testing.T) {
		assert.Equal(t, field(sealedGrid(), walkable).Len(), 10)
	})
	t.Run("a start that is not passable has an empty field", func(t *testing.T) {
		blocked := func(Cell[int]) bool {
			return false
		}

		assert.True(t, field(walledGrid(), blocked).IsEmpty())
	})
	t.Run("the path to every cell is as short as any", func(t *testing.T) {
		g := walledGrid()
		reached := field(g, walkable)

		for index := range reached.Region().All() {
			path := reached.PathTo(index)
			shortest := g.PathBreadthFirst(testStart, index, walkable)

			assert.Length(t, path, len(shortest), index.String()+": ")
			assert.Equal(t, path[0], testStart, index.String()+": ")
			assert.Equal(t, path[len(path)-1], index, index.String()+": ")
		}
	})
}

// benchmarkSearch runs the search corner to corner on each of the benchmarkGrids.
func benchmarkSearch(b *testing.B, search func(g *Grid[int], start, goal Index) []Index) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for b.Loop() {
			sinkPoints = search(g, geom.Pt(0, 0), geom.Pt(63, 63))
		}
	})
}

// benchmarkCostSearch runs the search corner to corner on each of the benchmarkGrids, every
// step costing one and then climbCost over scatterHeights.
func benchmarkCostSearch(b *testing.B, search func(g *Grid[int], start, goal Index, cost func(current, next Cell[int]) float64) []Index) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		b.Run("unit", func(b *testing.B) {
			for b.Loop() {
				sinkPoints = search(g, geom.Pt(0, 0), geom.Pt(63, 63), nil)
			}
		})
		b.Run("weighted", func(b *testing.B) {
			scatterHeights(g)

			for b.Loop() {
				sinkPoints = search(g, geom.Pt(0, 0), geom.Pt(63, 63), climbCost)
			}
		})
	})
}

func TestGrid_Path(t *testing.T) {
	t.Run("to the cell itself is that cell alone", func(t *testing.T) {
		path := walledGrid().Path(testStart, testStart, nil, nil)

		assert.Length(t, path, 1)
		assert.Equal(t, path[0], testStart)
	})
	t.Run("a straight path holds both ends", func(t *testing.T) {
		path := walledGrid().Path(testStart, geom.Pt(4, 0), nil, nil)

		assert.Length(t, path, 5)
		assert.Equal(t, path[0], testStart)
		assert.Equal(t, path[4], geom.Pt(4, 0))
	})
	t.Run("goes around a wall", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 3), geom.Sz(32.0, 32.0))
		g.Cell(geom.Pt(2, 0)).Set(1)
		g.Cell(geom.Pt(2, 1)).Set(1)

		path := g.Path(testStart, geom.Pt(4, 0), walkable, nil)

		assert.Length(t, path, 9)
		assert.Equal(t, path[0], testStart)
		assert.Equal(t, path[len(path)-1], geom.Pt(4, 0))
	})
	t.Run("a goal walled off has no path", func(t *testing.T) {
		assert.Length(t, sealedGrid().Path(testStart, testGoal, walkable, nil), 0)
	})
	t.Run("a start outside the grid has no path", func(t *testing.T) {
		assert.Length(t, walledGrid().Path(geom.Pt(-1, 0), testGoal, nil, nil), 0)
	})
	t.Run("searches from several goroutines at once find what one alone finds", func(t *testing.T) {
		g := NewRect[int](geom.Sz(40, 40), RectCellScale(32))
		scatterHeights(g)
		expected := g.Path(testStart, geom.Pt(39, 39), nil, climbCost)

		paths := make([][]Index, 8)
		var group sync.WaitGroup
		for i := range paths {
			group.Go(func() {
				paths[i] = g.Path(testStart, geom.Pt(39, 39), nil, climbCost)
			})
		}
		group.Wait()

		for _, path := range paths {
			assert.Equal(t, path, expected)
		}
	})
}

func ExampleGrid_Path() {
	g := NewRect[float64](geom.Sz(5, 3), RectCellScale(32))
	g.Fill(1)
	g.Cell(geom.Pt(2, 0)).Set(0)
	g.Cell(geom.Pt(2, 1)).Set(0)

	walkable := func(cell Cell[float64]) bool {
		return *cell.Value() > 0
	}
	cost := func(_, next Cell[float64]) float64 {
		return *next.Value()
	}

	fmt.Println(g.Path(geom.Pt(0, 0), geom.Pt(4, 0), walkable, cost))
	fmt.Println(g.Cell(geom.Pt(0, 0)).PathTo(geom.Pt(4, 0), walkable, cost)[4])
	fmt.Println(g.Path(geom.Pt(0, 0), geom.Pt(2, 0), walkable, cost) == nil)
	// Output:
	// [(0,0) (1,0) (1,1) (1,2) (2,2) (3,2) (4,2) (4,1) (4,0)]
	// (2,2)
	// true
}

func TestGrid_PathBreadthFirst(t *testing.T) {
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []Index {
		return g.PathBreadthFirst(testStart, testGoal, passable)
	})
}

func BenchmarkGrid_PathBreadthFirst(b *testing.B) {
	benchmarkSearch(b, func(g *Grid[int], start, goal Index) []Index {
		return g.PathBreadthFirst(start, goal, nil)
	})
}

func TestGrid_PathUniformCost(t *testing.T) {
	testCost(t, func(g *Grid[int], cost func(current, next Cell[int]) float64) []Index {
		return g.PathUniformCost(geom.Pt(0, 0), geom.Pt(3, 0), nil, cost)
	})
	t.Run("pays the least, around what is dear", func(t *testing.T) {
		path := weightedGrid().PathUniformCost(testStart, geom.Pt(4, 0), nil, stepCost)

		assert.Length(t, path, 13)
		assert.NotContains(t, path, geom.Pt(2, 0))
	})
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []Index {
		return g.PathUniformCost(testStart, testGoal, passable, nil)
	})
	t.Run("cost is never asked about a cell that is not passable", func(t *testing.T) {
		walledGrid().PathUniformCost(testStart, testGoal, walkable, func(current, next Cell[int]) float64 {
			assert.NotEqual(t, next.Index(), testWall)

			return 1
		})
	})
}

func BenchmarkGrid_PathUniformCost(b *testing.B) {
	benchmarkCostSearch(b, func(g *Grid[int], start, goal Index, cost func(current, next Cell[int]) float64) []Index {
		return g.PathUniformCost(start, goal, nil, cost)
	})
}

func TestGrid_PathGreedyBestFirst(t *testing.T) {
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []Index {
		return g.PathGreedyBestFirst(testStart, testGoal, passable)
	})
	t.Run("to the cell itself is that cell alone", func(t *testing.T) {
		assert.Equal(t, walledGrid().PathGreedyBestFirst(testStart, testStart, nil), []Index{testStart})
	})
	t.Run("crosses a grid of many chunks, on every kind of grid", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(40, 40)) {
			path := g.PathGreedyBestFirst(testStart, geom.Pt(39, 39), nil)

			assert.Equal(t, path[0], testStart, name+": ")
			assert.Equal(t, path[len(path)-1], geom.Pt(39, 39), name+": ")
			for i := 1; i < len(path); i++ {
				assert.Equal(t, g.Distance(path[i-1], path[i]), 1, name+": ")
			}
		}
	})
}

func BenchmarkGrid_PathGreedyBestFirst(b *testing.B) {
	benchmarkSearch(b, func(g *Grid[int], start, goal Index) []Index {
		return g.PathGreedyBestFirst(start, goal, nil)
	})
}

func TestGrid_PathAStar(t *testing.T) {
	testCost(t, func(g *Grid[int], cost func(current, next Cell[int]) float64) []Index {
		return g.PathAStar(geom.Pt(0, 0), geom.Pt(3, 0), nil, cost)
	})
	t.Run("pays what uniform cost search pays", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			for cell := range g.All() {
				cell.Set((cell.Index().X*3 + cell.Index().Y*7) % 4)
			}

			for cell := range g.All() {
				expected, _ := g.FieldUniformCost(testStart, nil, stepCost).CostTo(cell.Index())
				path := g.PathAStar(testStart, cell.Index(), nil, stepCost)

				paid := 0.0
				for i := 1; i < len(path); i++ {
					paid += stepCost(g.Cell(path[i-1]), g.Cell(path[i]))
				}

				assert.Equal(t, paid, expected, name+" at "+cell.Index().String()+": ")
			}
		}
	})
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []Index {
		return g.PathAStar(testStart, testGoal, passable, nil)
	})
	t.Run("cost is never asked about a cell that is not passable", func(t *testing.T) {
		walledGrid().PathAStar(testStart, testGoal, walkable, func(current, next Cell[int]) float64 {
			assert.NotEqual(t, next.Index(), testWall)

			return 1
		})
	})
}

func BenchmarkGrid_PathAStar(b *testing.B) {
	benchmarkCostSearch(b, func(g *Grid[int], start, goal Index, cost func(current, next Cell[int]) float64) []Index {
		return g.PathAStar(start, goal, nil, cost)
	})
}

func FuzzGrid_PathAStar(f *testing.F) {
	for kind := range uint8(7) {
		f.Add(kind, uint8(8), uint8(8), uint8(2), uint8(2), uint8(9), uint8(9), uint8(3))
		f.Add(kind, uint8(4), uint8(9), uint8(0), uint8(0), uint8(4), uint8(9), uint8(11))
		f.Add(kind, uint8(0), uint8(0), uint8(2), uint8(2), uint8(2), uint8(2), uint8(0))
	}

	f.Fuzz(func(t *testing.T, kind, width, height, x0, y0, x1, y1, seed uint8) {
		g := fuzzGrid(kind, width, height)
		for cell := range g.All() {
			cell.Set((cell.Index().X*7 + cell.Index().Y*13 + int(seed)) % 5)
		}
		passable := func(cell Cell[int]) bool {
			return *cell.Value() != 4
		}
		cost := func(_, next Cell[int]) float64 {
			return float64(1 + *next.Value())
		}
		start, goal := fuzzIndex(g, x0, y0), fuzzIndex(g, x1, y1)

		path, shortest := g.PathAStar(start, goal, passable, cost), g.PathBreadthFirst(start, goal, passable)

		assert.Equal(t, path == nil, shortest == nil)
		if path == nil {
			return
		}

		assert.Equal(t, path[0], start)
		assert.Equal(t, path[len(path)-1], goal)

		paid := 0.0
		for i, index := range path {
			assert.True(t, passable(g.Cell(index)), index.String()+": ")
			if i > 0 {
				assert.Equal(t, g.Distance(path[i-1], index), 1, index.String()+": ")
				paid += cost(g.Cell(path[i-1]), g.Cell(index))
			}
		}

		cheapest, _ := g.FieldUniformCost(start, passable, cost).CostTo(goal)
		assert.Equal(t, paid, cheapest)
		assert.Length(t, g.PathAStar(start, goal, passable, nil), len(shortest))
		assert.Length(t, g.PathUniformCost(start, goal, passable, nil), len(shortest))
	})
}

func TestGrid_Field(t *testing.T) {
	t.Run("a nil cost counts the steps of FieldBreadthFirst, on every layout", func(t *testing.T) {
		for name, g := range layoutGrids() {
			field, expected := g.Field(geom.Pt(1, 1), nil, nil), g.FieldBreadthFirst(geom.Pt(1, 1), nil)

			assert.True(t, field.Region().Equal(expected.Region()), name+": ")
			assert.Equal(t, field.PathTo(geom.Pt(2, 2)), expected.PathTo(geom.Pt(2, 2)), name+": ")
		}
	})
	t.Run("a cost weighs the steps of FieldUniformCost, on every layout", func(t *testing.T) {
		for name, g := range layoutGrids() {
			field, expected := g.Field(geom.Pt(1, 1), nil, stepCost), g.FieldUniformCost(geom.Pt(1, 1), nil, stepCost)

			assert.Equal(t, field.PathTo(geom.Pt(2, 2)), expected.PathTo(geom.Pt(2, 2)), name+": ")
		}
	})
}

func TestGrid_FieldBreadthFirst(t *testing.T) {
	testSearchField(t, func(g *Grid[int], passable func(Cell[int]) bool) Field {
		return g.FieldBreadthFirst(testStart, passable)
	})
}

func BenchmarkGrid_FieldBreadthFirst(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for b.Loop() {
			sinkField = g.FieldBreadthFirst(geom.Pt(0, 0), nil)
		}
	})
}

func TestGrid_FieldUniformCost(t *testing.T) {
	testSearchField(t, func(g *Grid[int], passable func(Cell[int]) bool) Field {
		return g.FieldUniformCost(testStart, passable, nil)
	})
	t.Run("expands a cell once, however often it is reached more cheaply", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			scatterHeights(g)

			asked := map[Index]int{}
			g.FieldUniformCost(testStart, nil, func(current, next Cell[int]) float64 {
				asked[current.Index()]++

				return climbCost(current, next)
			})

			for index, times := range asked {
				assert.LessOrEqual(t, times, len(g.Neighbors(index)), name+" at "+index.String()+": ")
			}
		}
	})
}

func BenchmarkGrid_FieldUniformCost(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		scatterHeights(g)

		for b.Loop() {
			sinkField = g.FieldUniformCost(geom.Pt(0, 0), nil, climbCost)
		}
	})
}

func TestGrid_Reachable(t *testing.T) {
	t.Run("is the cells within the steps along passable cells", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(9, 9)) {
			for cell := range g.All() {
				if index := cell.Index(); (index.X*5+index.Y*3)%4 == 0 {
					cell.Set(1)
				}
			}
			start := geom.Pt(4, 5)
			field := g.FieldBreadthFirst(start, walkable)

			for steps := range 6 {
				var expected []Index
				for index := range field.Region().All() {
					if len(field.PathTo(index))-1 <= steps {
						expected = append(expected, index)
					}
				}

				assert.True(t, g.Reachable(start, float64(steps), walkable, nil).Region().Equal(Reg(expected...)), name+" at "+strconv.Itoa(steps)+": ")
			}
		}
	})
	t.Run("goes around what a range sees past", func(t *testing.T) {
		g := sealedGrid()

		assert.False(t, g.Reachable(testStart, 20, walkable, nil).Region().Contains(testGoal))
		assert.Equal(t, g.Reachable(testStart, 20, walkable, nil).Region().Len(), 10)
	})
	t.Run("on a grid of many chunks is the range of the steps, on every kind of grid", func(t *testing.T) {
		for name, g := range ruleGrids(geom.Sz(40, 40)) {
			field := g.Reachable(geom.Pt(20, 20), 6, nil, nil)

			assert.True(t, field.Region().Equal(g.Range(geom.Pt(20, 20), 6)), name+": ")
			assert.False(t, field.Contains(geom.Pt(39, 0)), name+": ")
		}
	})
	t.Run("no steps is the start alone", func(t *testing.T) {
		assert.Equal(t, slices.Collect(walledGrid().Reachable(testStart, 0, nil, nil).Region().All()), []Index{testStart})
	})
	t.Run("negative steps reach nothing", func(t *testing.T) {
		assert.True(t, walledGrid().Reachable(testStart, -1, nil, nil).Region().IsEmpty())
	})
	t.Run("is the cells of the field within the budget, on every layout", func(t *testing.T) {
		for name, g := range layoutGrids() {
			field := g.Field(geom.Pt(1, 1), nil, stepCost)

			for _, budget := range []float64{0, 0.5, 1, 2.5, 3, 40} {
				var expected []Index
				for index := range field.All() {
					if cost, _ := field.CostTo(index); cost <= budget {
						expected = append(expected, index)
					}
				}

				assert.True(t, g.Reachable(geom.Pt(1, 1), budget, nil, stepCost).Region().Equal(Reg(expected...)), name+": ")
			}
		}
	})
	t.Run("a nil cost agrees with a cost of one per step, on every layout, a fraction of a step too", func(t *testing.T) {
		one := func(_, _ Cell[int]) float64 {
			return 1
		}

		for name, g := range layoutGrids() {
			for _, budget := range []float64{0, 0.5, 1, 2.5, 3, math.Inf(1)} {
				assert.True(t, g.Reachable(geom.Pt(1, 1), budget, nil, nil).Region().Equal(g.Reachable(geom.Pt(1, 1), budget, nil, one).Region()), name+": ")
			}
		}
	})
	t.Run("a budget holds what costs exactly it", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 1), RectCellScale(32))
		half := func(_, _ Cell[int]) float64 {
			return 0.5
		}

		assert.True(t, g.Reachable(geom.Pt(0, 0), 1.5, nil, half).Contains(geom.Pt(3, 0)))
		assert.False(t, g.Reachable(geom.Pt(0, 0), 1.25, nil, half).Contains(geom.Pt(3, 0)))
	})
	t.Run("gives the path and the cost to a cell it reached", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 1), RectCellScale(32))
		field := g.Reachable(geom.Pt(0, 0), 2, nil, nil)
		cost, found := field.CostTo(geom.Pt(2, 0))

		assert.True(t, found)
		assert.Equal(t, cost, 2.0)
		assert.Equal(t, field.PathTo(geom.Pt(2, 0)), []Index{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(2, 0)})
	})
	t.Run("walks no further than the budget", func(t *testing.T) {
		g := NewRect[int](geom.Sz(64, 64), RectCellScale(32))
		asked := 0

		g.Reachable(geom.Pt(32, 32), 2, nil, func(_, _ Cell[int]) float64 {
			asked++

			return 1
		})

		assert.LessOrEqual(t, asked, 13*4)
	})
	t.Run("a NaN budget reaches nothing", func(t *testing.T) {
		assert.True(t, walledGrid().Reachable(testStart, math.NaN(), nil, nil).IsEmpty())
	})
	t.Run("a start outside the grid reaches nothing", func(t *testing.T) {
		assert.True(t, walledGrid().Reachable(geom.Pt(-1, 0), 3, nil, nil).Region().IsEmpty())
	})
}

func BenchmarkGrid_Reachable(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		for b.Loop() {
			sinkRegion = g.Reachable(geom.Pt(32, 32), 8, nil, nil).Region()
		}
	})
	b.Run("large", func(b *testing.B) {
		g := NewRect[int](geom.Sz(512, 512), RectCellScale(32))

		for b.Loop() {
			sinkRegion = g.Reachable(geom.Pt(256, 256), 8, nil, nil).Region()
		}
	})
}

func ExampleGrid_Reachable() {
	g := NewRect[bool](geom.Sz(5, 5), RectCellScale(32))
	g.Fill(true)
	g.Cell(geom.Pt(1, 2)).Set(false)

	walkable := func(cell Cell[bool]) bool {
		return *cell.Value()
	}

	reach := g.Reachable(geom.Pt(2, 2), 1, walkable, nil).Region()
	blast := g.Range(geom.Pt(1, 1), 1)

	fmt.Println(reach.Elements())
	fmt.Println(reach.Intersection(blast).Elements())
	fmt.Println(reach.Intersects(blast), reach.Contains(geom.Pt(1, 2)))
	// Output:
	// [(2,1) (2,2) (2,3) (3,2)]
	// [(2,1)]
	// true false
}

func TestGrid_Contains(t *testing.T) {
	g := NewRect[int](geom.Sz(4, 3), geom.Sz(32.0, 32.0))

	t.Run("the corners are inside", func(t *testing.T) {
		assert.True(t, g.Contains(geom.Pt(0, 0)))
		assert.True(t, g.Contains(geom.Pt(3, 2)))
	})
	t.Run("a negative coordinate is outside", func(t *testing.T) {
		assert.False(t, g.Contains(geom.Pt(-1, 0)))
		assert.False(t, g.Contains(geom.Pt(0, -1)))
	})
	t.Run("the size itself is outside", func(t *testing.T) {
		assert.False(t, g.Contains(geom.Pt(4, 0)))
		assert.False(t, g.Contains(geom.Pt(0, 3)))
	})
}

func TestGrid_JSON(t *testing.T) {
	grids := kindGrids(geom.Sz(3, 2), geom.Sz(32.0, 20.0), geom.Pt(7.0, -3.0))
	grids["diagonal"] = NewRect[int](geom.Sz(3, 2), geom.Sz(32.0, 20.0), withDiagonalMovement)
	grids["flat-top even"] = NewHexFlatTop[int](geom.Sz(3, 2), geom.Sz(32.0, 20.0), func(options *HexOptions) {
		options.EvenOffset = true
	})
	grids["pointy-top even"] = NewHexPointyTop[int](geom.Sz(3, 2), geom.Sz(32.0, 20.0), func(options *HexOptions) {
		options.EvenOffset = true
	})
	square := NewRect[int](geom.Sz(2, 1), RectCellScale(32))
	square.Cell(geom.Pt(1, 0)).Set(7)

	t.Run("wire format is the constructor and the cells", func(t *testing.T) {
		assert.JSON(t, square, `{"k":"Rect","m":"Cardinal","c":{"w":32,"h":32},"o":{"x":0,"y":0},"v":{"w":2,"h":1,"d":[0,7]}}`)
	})
	t.Run("a hexagonal grid writes no movement", func(t *testing.T) {
		assert.JSON(t, NewHexPointyTop[int](geom.Sz(1, 1), geom.SzU(16.0)), `{"k":"HexPointyTop","c":{"w":16,"h":16},"o":{"x":0,"y":0},"v":{"w":1,"h":1,"d":[0]}}`)
	})
	t.Run("a grid held by value is written too", func(t *testing.T) {
		assert.JSON(t, struct {
			Map Grid[int] `json:"map"`
		}{*square}, `{"map":{"k":"Rect","m":"Cardinal","c":{"w":32,"h":32},"o":{"x":0,"y":0},"v":{"w":2,"h":1,"d":[0,7]}}}`)
	})
	t.Run("round-trip, on every layout", func(t *testing.T) {
		for name, g := range grids {
			for cell := range g.All() {
				cell.Set(cell.Index().X + 10*cell.Index().Y)
			}

			data, err := json.Marshal(g)
			assert.NoError(t, err, name+": ")

			var decoded Grid[int]
			assert.NoError(t, json.Unmarshal(data, &decoded), name+": ")

			assert.Equal(t, decoded.Size(), g.Size(), name+": ")
			assert.Equal(t, decoded.Bounds(), g.Bounds(), name+": ")
			assert.Equal(t, decoded.CellScale(), g.CellScale(), name+": ")

			for cell := range g.All() {
				index := cell.Index()

				assert.Equal(t, *decoded.Cell(index).Value(), *cell.Value(), name+" at "+index.String()+": ")
				assert.Equal(t, decoded.Cell(index).Center(), cell.Center(), name+" at "+index.String()+": ")
				assert.Equal(t, decoded.Neighbors(index), g.Neighbors(index), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("a grid nested in a struct round-trips", func(t *testing.T) {
		type save struct {
			Map *Grid[int] `json:"map"`
		}

		data, err := json.Marshal(save{square})
		assert.NoError(t, err)

		var decoded save
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, *decoded.Map.Cell(geom.Pt(1, 0)).Value(), 7)
	})
	t.Run("null reads as nothing", func(t *testing.T) {
		kept := square.Clone()

		assert.NoError(t, json.Unmarshal([]byte("null"), kept))
		assert.Equal(t, kept.Size(), square.Size())
	})
	for _, bad := range []struct {
		name string
		data string
	}{
		{"malformed", `{"k":`},
		{"no kind", `{}`},
		{"an unknown kind", `{"k":"Triangle","c":{"w":1,"h":1}}`},
		{"square cells with no movement", `{"k":"Rect","m":"None","c":{"w":1,"h":1}}`},
		{"an unknown movement", `{"k":"Rect","m":"Knight","c":{"w":1,"h":1}}`},
		{"no cell scale", `{"k":"Rect"}`},
		{"a negative cell scale", `{"k":"HexFlatTop","c":{"w":-1,"h":1}}`},
		{"a negative size", `{"k":"Rect","c":{"w":1,"h":1},"v":{"w":-1,"h":1,"d":[]}}`},
		{"too few values", `{"k":"Rect","c":{"w":1,"h":1},"v":{"w":2,"h":2,"d":[1]}}`},
	} {
		t.Run(bad.name+" is an error that leaves the grid as it was", func(t *testing.T) {
			kept := square.Clone()

			assert.Error(t, json.Unmarshal([]byte(bad.data), kept))
			assert.Equal(t, kept.Size(), square.Size())
			assert.Equal(t, *kept.Cell(geom.Pt(1, 0)).Value(), 7)
		})
	}
	t.Run("square cells with no movement are ErrUnknownMovement", func(t *testing.T) {
		assert.ErrorIs(t, json.Unmarshal([]byte(`{"k":"Rect","m":"None","c":{"w":1,"h":1}}`), new(Grid[int])), ErrUnknownMovement)
	})
	t.Run("a cell scale no cell can be drawn at is ErrInvalidCellScale", func(t *testing.T) {
		assert.ErrorIs(t, json.Unmarshal([]byte(`{"k":"HexFlatTop","c":{"w":-1,"h":1}}`), new(Grid[int])), ErrInvalidCellScale)
	})
}

func FuzzGrid_JSON(f *testing.F) {
	f.Add(`{"k":"Rect","m":"Cardinal","c":{"w":32,"h":32},"o":{"x":0,"y":0},"v":{"w":2,"h":1,"d":[0,7]}}`)
	f.Add(`{"k":"HexFlatTopEven","m":"None","c":{"w":1e308,"h":1e-308},"o":{"x":1e308,"y":-1e308},"v":{"w":1,"h":1,"d":[1]}}`)
	f.Add(`{"k":"Isometric","c":{"w":1,"h":1},"v":{"w":0,"h":9,"d":[]}}`)

	f.Fuzz(func(t *testing.T, data string) {
		var decoded Grid[int]
		if json.Unmarshal([]byte(data), &decoded) != nil || data == "null" {
			return
		}

		written, err := json.Marshal(decoded)
		assert.NoError(t, err)

		var again Grid[int]
		assert.NoError(t, json.Unmarshal(written, &again))
		assert.Equal(t, again.Size(), decoded.Size())
		assert.Equal(t, again.Bounds().IsZero(), decoded.Bounds().IsZero())
	})
}

func ExampleGrid_MarshalJSON() {
	g := NewRect[int](geom.Sz(2, 1), RectCellScale(32), func(options *SquareOptions) {
		options.Movement = MovementDiagonal
	})
	g.Cell(geom.Pt(1, 0)).Set(7)

	data, _ := json.Marshal(g)
	fmt.Println(string(data))

	var loaded Grid[int]
	fmt.Println(json.Unmarshal(data, &loaded), len(loaded.Neighbors(geom.Pt(0, 0))), *loaded.Cell(geom.Pt(1, 0)).Value())
	// Output:
	// {"k":"Rect","m":"Diagonal","c":{"w":32,"h":32},"o":{"x":0,"y":0},"v":{"w":2,"h":1,"d":[0,7]}}
	// <nil> 1 7
}
