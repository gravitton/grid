package grid

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"testing"

	"github.com/gravitton/assert"
	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

var (
	sinkRegion    Region
	sinkPoints    []geom.Point[int]
	sinkPoint     geom.Point[int]
	sinkCenter    geom.Point[float64]
	sinkBox       geom.Box[float64]
	sinkSize      geom.Size[float64]
	sinkBool      bool
	sinkString    string
	sinkHexLayout *HexLayout
	sinkPolygon   geom.RegularPolygon[float64]
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

// testGrids returns one grid of each kind and offset, of the given size in cells, at the given
// origin.
func testGrids(size geom.Size[int], cellSize geom.Size[float64], origin geom.Point[float64]) map[string]*Grid[int] {
	rect := func(options *RectOptions) {
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
		"rectangular":     NewRect[int](size, cellSize, rect),
		"isometric":       NewIsometric[int](size, cellSize, rect),
		"flat-top":        NewHexFlatTop[int](size, cellSize, odd),
		"flat-top even":   NewHexFlatTop[int](size, cellSize, even),
		"pointy-top":      NewHexPointyTop[int](size, cellSize, odd),
		"pointy-top even": NewHexPointyTop[int](size, cellSize, even),
	}
}

// benchmarkGrids runs the benchmark on each of the testGrids at 64 by 64 cells, in the order of
// their names.
func benchmarkGrids(b *testing.B, run func(b *testing.B, g *Grid[int])) {
	grids := testGrids(geom.Sz(64, 64), geom.Sz(32.0, 20.0), geom.Pt(100.0, 50.0))

	for _, name := range slices.Sorted(maps.Keys(grids)) {
		b.Run(name, func(b *testing.B) {
			run(b, grids[name])
		})
	}
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

func TestGrid_Cell(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	t.Run("is a handle on the index", func(t *testing.T) {
		assert.Equal(t, g.Cell(geom.Pt(1, 2)).Index(), geom.Pt(1, 2))
		assert.True(t, g.Cell(geom.Pt(1, 2)).IsValid())
	})
	t.Run("two cells of one index share the value", func(t *testing.T) {
		g.Cell(geom.Pt(1, 2)).Set(42)

		assert.Equal(t, *g.Cell(geom.Pt(1, 2)).Value(), 42)
	})
	t.Run("outside the grid keeps the index and is not valid", func(t *testing.T) {
		assert.Equal(t, g.Cell(geom.Pt(99, -1)).Index(), geom.Pt(99, -1))
		assert.False(t, g.Cell(geom.Pt(99, -1)).IsValid())
	})
}

func ExampleGrid_Cell() {
	type tile struct {
		Walkable bool
		Cost     float64
	}

	g := NewRect[tile](geom.Sz(100, 100), RectCellSize(32))
	index := geom.Pt(3, 4)

	cell := g.Cell(index)
	cell.Set(tile{Cost: 3})
	cell.Value().Cost = 5

	outside := g.Cell(geom.Pt(-1, 0))
	outside.Set(tile{Cost: 9})

	fmt.Println(g.Contains(index), g.Size())
	fmt.Println(cell, cell.IsValid(), cell.Value().Cost)
	fmt.Println(cell.Neighbors())
	fmt.Println(cell.DistanceTo(geom.Pt(0, 0)))
	fmt.Println(outside, outside.IsValid(), outside.Value() == nil)
	// Output:
	// true 100x100
	// Cell(3,4) true 5
	// [(4,4) (3,5) (2,4) (3,3)]
	// 7
	// Cell(-1,0) false true
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

func TestGrid_All(t *testing.T) {
	g := NewHexFlatTop[int](geom.Sz(3, 2), geom.SzU(16.0))

	t.Run("yields every cell once, row by row", func(t *testing.T) {
		var indices []geom.Point[int]
		for cell := range g.All() {
			indices = append(indices, cell.Index())
		}

		assert.Equal(t, indices, []geom.Point[int]{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(2, 0), geom.Pt(0, 1), geom.Pt(1, 1), geom.Pt(2, 1)})
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
func fuzzIndex(g *Grid[int], x, y uint8) geom.Point[int] {
	return geom.Pt(int(x)%(g.Width()+4)-2, int(y)%(g.Height()+4)-2)
}

// rangeGrids returns one grid of each neighbor rule, 11 by 11, with a center well inside.
func rangeGrids() (map[string]*Grid[int], geom.Point[int]) {
	return map[string]*Grid[int]{
		"cardinal":   NewRect[int](geom.Sz(11, 11), geom.Sz(32.0, 32.0)),
		"diagonal":   NewRect[int](geom.Sz(11, 11), geom.Sz(32.0, 32.0), withDiagonalMovement),
		"flat-top":   NewHexFlatTop[int](geom.Sz(11, 11), geom.Sz(32.0, 32.0)),
		"pointy-top": NewHexPointyTop[int](geom.Sz(11, 11), geom.Sz(32.0, 32.0)),
	}, geom.Pt(5, 5)
}

func TestGrid_Neighbors(t *testing.T) {
	grids, center := rangeGrids()

	t.Run("cardinal has four, by increasing angle", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Neighbors(center), []geom.Point[int]{geom.Pt(6, 5), geom.Pt(5, 6), geom.Pt(4, 5), geom.Pt(5, 4)})
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
				var expected []geom.Point[int]
				for other := range g.All() {
					if g.Distance(cell.Index(), other.Index()) == 1 {
						expected = append(expected, other.Index())
					}
				}

				assert.True(t, Reg(g.Neighbors(cell.Index())).Equal(Reg(expected)), name+" at "+cell.Index().String()+": ")
			}
		}
	})
	t.Run("is the range of one step without its center", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				assert.True(t, Reg(g.Neighbors(index)).Equal(g.Range(index, 1, nil).Difference(Reg([]geom.Point[int]{index}))), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("an index outside the grid has none, even beside it", func(t *testing.T) {
		for name, g := range grids {
			assert.Nil(t, g.Neighbors(geom.Pt(-1, 0)), name+": ")
			assert.Nil(t, g.Neighbors(geom.Pt(40, 40)), name+": ")
		}
	})
	t.Run("allocates once, at the count of the neighbor rule", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, cap(g.Neighbors(geom.Pt(0, 0))), len(g.Neighbors(center)), name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = g.Neighbors(center)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Neighbors(b *testing.B) {
	grids, center := rangeGrids()

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

		var expected []geom.Point[int]
		for cell := range g.All() {
			if g.Distance(index, cell.Index()) == 1 {
				expected = append(expected, cell.Index())
			}
		}

		assert.Equal(t, len(neighbors), len(expected))
		assert.True(t, Reg(neighbors).Equal(Reg(expected)))

		for _, neighbor := range neighbors {
			assert.True(t, slices.Contains(g.Neighbors(neighbor), index), neighbor.String()+": ")
		}
	})
}

func TestGrid_AppendNeighbors(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("appends the indices of Neighbors", func(t *testing.T) {
				assert.Equal(t, g.AppendNeighbors([]geom.Point[int]{geom.Pt(-9, -9)}, center), append([]geom.Point[int]{geom.Pt(-9, -9)}, g.Neighbors(center)...))
			})
			t.Run("appends nothing for an index outside the grid", func(t *testing.T) {
				assert.Equal(t, g.AppendNeighbors([]geom.Point[int]{geom.Pt(-9, -9)}, geom.Pt(-1, 0)), []geom.Point[int]{geom.Pt(-9, -9)})
			})
			t.Run("allocates nothing with room", func(t *testing.T) {
				buffer := make([]geom.Point[int], 0, 8)

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = g.AppendNeighbors(buffer[:0], center)
				}), 0.0)
			})
		})
	}
}

func TestGrid_Range(t *testing.T) {
	grids, center := rangeGrids()

	t.Run("a negative radius is empty", func(t *testing.T) {
		assert.True(t, grids["cardinal"].Range(center, -1, nil).IsEmpty())
	})
	t.Run("a zero radius is the center alone", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Range(center, 0, nil).Elements(), []geom.Point[int]{center})
	})
	t.Run("a radius beyond the grid is every cell, however far beyond", func(t *testing.T) {
		for name, g := range grids {
			across := g.Width() + g.Height()

			for _, radius := range []int{across, 1 << 40, math.MaxInt} {
				assert.Equal(t, g.Range(geom.Pt(0, 0), radius, nil).Len(), g.Width()*g.Height(), name+" "+strconv.Itoa(radius)+": ")
			}
		}
	})
	t.Run("the farthest cell is in a range of any radius beyond it, from every cell of a small grid", func(t *testing.T) {
		for size := range container.Arr[int](geom.Sz(7, 7)).Keys() {
			size := geom.Sz(size.X+1, size.Y+1)
			small := testGrids(size, geom.Sz(32.0, 20.0), geom.Pt(0.0, 0.0))
			small["diagonal"] = NewRect[int](size, geom.Sz(32.0, 32.0), withDiagonalMovement)

			for name, g := range small {
				for cell := range g.All() {
					assert.Equal(t, cell.Range(math.MaxInt, nil).Len(), g.Width()*g.Height(), name+" "+size.String()+" from "+cell.Index().String()+": ")
				}
			}
		}
	})
	t.Run("a radius beyond the grid still hides what is behind a wall", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return cell.Index() != geom.Pt(5, 6)
		}

		for name, g := range grids {
			assert.True(t, g.Range(center, math.MaxInt, open).Equal(g.Range(center, g.Width()+g.Height(), open)), name+": ")
		}
	})
	t.Run("cardinal is the diamond of Manhattan distance", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Range(center, 1, nil).Len(), 5)
		assert.Equal(t, grids["cardinal"].Range(center, 3, nil).Len(), 25)
	})
	t.Run("diagonal is the square of Chebyshev distance", func(t *testing.T) {
		assert.Equal(t, grids["diagonal"].Range(center, 1, nil).Len(), 9)
		assert.Equal(t, grids["diagonal"].Range(center, 3, nil).Len(), 49)
	})
	t.Run("is exactly the cells within the distance", func(t *testing.T) {
		for name, g := range grids {
			for n := range 5 {
				var expected []geom.Point[int]
				for cell := range g.All() {
					if g.Distance(center, cell.Index()) <= n {
						expected = append(expected, cell.Index())
					}
				}

				assert.True(t, g.Range(center, n, nil).Equal(Reg(expected)), name+" at "+strconv.Itoa(n)+": ")
			}
		}
	})
	t.Run("a cell that is not passable is left out and hides what is behind it", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return cell.Index() != geom.Pt(5, 6)
		}

		for _, name := range []string{"cardinal", "diagonal", "flat-top"} {
			reach := grids[name].Range(center, 2, open)

			assert.False(t, reach.Contains(geom.Pt(5, 6)), name+": ")
			assert.False(t, reach.Contains(geom.Pt(5, 7)), name+": ")
			assert.True(t, reach.Contains(geom.Pt(5, 4)), name+": ")
		}
	})
	t.Run("a cell beside the center is in range whatever blocks", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return cell.Index() != geom.Pt(6, 5) && cell.Index() != geom.Pt(5, 6)
		}

		assert.True(t, grids["diagonal"].Range(center, 2, open).Contains(geom.Pt(6, 6)))
	})
	t.Run("sight is reciprocal", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return (cell.Index().X*7+cell.Index().Y*13)%5 != 2
		}

		for name, g := range grids {
			for seen := range g.Range(center, 4, open).All() {
				assert.True(t, g.Range(seen, 4, open).Contains(center), name+" at "+seen.String()+": ")
			}
		}
	})
	t.Run("walls that hide nothing change nothing", func(t *testing.T) {
		open := func(Cell[int]) bool {
			return true
		}

		for name, g := range grids {
			for _, index := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				assert.True(t, g.Range(index, 4, nil).Equal(g.Range(index, 4, open)), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("is empty from an index outside the grid, even beside it", func(t *testing.T) {
		for name, g := range grids {
			assert.True(t, g.Range(geom.Pt(-1, 5), 3, nil).IsEmpty(), name+": ")
			assert.True(t, g.Range(geom.Pt(40, 40), 3, nil).IsEmpty(), name+": ")
		}
	})
	t.Run("is empty from an index that is not passable", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return cell.Index() != center
		}

		for name, g := range grids {
			assert.True(t, g.Range(center, 3, open).IsEmpty(), name+": ")
		}
	})
	t.Run("an empty range allocates nothing", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Range(geom.Pt(-1, 5), 4, nil)
			}), 0.0, name+": ")
		}
	})
	t.Run("asks about the index once", func(t *testing.T) {
		for name, g := range grids {
			asked := 0
			g.Range(center, 3, func(cell Cell[int]) bool {
				if cell.Index() == center {
					asked++
				}

				return true
			})

			assert.Equal(t, asked, 1, name+": ")
		}
	})
	t.Run("holds no index outside the grid", func(t *testing.T) {
		for name, g := range grids {
			for index := range g.Range(geom.Pt(0, 0), 2, nil).All() {
				assert.True(t, g.Contains(index), name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("asks about every cell once, whatever is hidden", func(t *testing.T) {
		for name, g := range grids {
			asked := map[geom.Point[int]]int{}
			g.Range(center, 4, func(cell Cell[int]) bool {
				asked[cell.Index()]++

				return (cell.Index().X*7+cell.Index().Y*13)%5 != 2 || cell.Index() == center
			})

			assert.Equal(t, len(asked), g.Range(center, 4, nil).Len(), name+": ")

			for index, times := range asked {
				assert.Equal(t, times, 1, name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("allocates once, its indices, where every cell passes", func(t *testing.T) {
		every := func(Cell[int]) bool {
			return true
		}

		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Range(center, 4, nil)
			}), 1.0, name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Range(center, 4, every)
			}), 1.0, name+": ")
		}
	})
	t.Run("allocates twice more where a cell is refused, for the marks and the sight past them, and once more on a hexagonal grid for its sweep", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return cell.Index() != geom.Pt(5, 6)
		}
		allocs := map[string]float64{"cardinal": 3, "diagonal": 3, "flat-top": 4, "pointy-top": 4}

		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Range(center, 4, open)
			}), allocs[name], name+": ")
		}
	})
}

func BenchmarkGrid_Range(b *testing.B) {
	open := func(cell Cell[int]) bool {
		return (cell.Index().X*7+cell.Index().Y*13)%11 != 3
	}
	every := func(Cell[int]) bool {
		return true
	}

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
					sinkRegion = kind.grid.Range(geom.Pt(50, 50), radius, nil)
				}
			})
			b.Run(kind.name+" at "+strconv.Itoa(radius)+" where every cell passes", func(b *testing.B) {
				for b.Loop() {
					sinkRegion = kind.grid.Range(geom.Pt(50, 50), radius, every)
				}
			})
			b.Run(kind.name+" at "+strconv.Itoa(radius)+" with walls", func(b *testing.B) {
				for b.Loop() {
					sinkRegion = kind.grid.Range(geom.Pt(50, 50), radius, open)
				}
			})
		}
	}
}

func ExampleGrid_Range() {
	g := NewRect[bool](geom.Sz(5, 5), RectCellSize(32))
	g.Fill(true)
	g.Cell(geom.Pt(2, 1)).Set(false)

	walkable := func(cell Cell[bool]) bool {
		return *cell.Value()
	}

	index := geom.Pt(2, 2)

	fmt.Println(g.Distance(index, geom.Pt(4, 3)))
	fmt.Println(g.Range(index, 1, nil).Elements())
	fmt.Println(g.Range(index, 2, walkable).Elements())
	fmt.Println(g.Cell(index).Range(2, walkable).Equal(g.Range(index, 2, walkable)))
	fmt.Println(g.Range(geom.Pt(2, 1), 2, walkable).IsEmpty())
	// Output:
	// 3
	// [(1,2) (2,1) (2,2) (2,3) (3,2)]
	// [(0,2) (1,1) (1,2) (1,3) (2,2) (2,3) (2,4) (3,1) (3,2) (3,3) (4,2)]
	// true
	// true
}

func FuzzGrid_Range(f *testing.F) {
	for kind := range uint8(7) {
		f.Add(kind, uint8(8), uint8(8), uint8(6), uint8(6), uint8(3), uint8(5))
		f.Add(kind, uint8(0), uint8(0), uint8(2), uint8(2), uint8(2), uint8(0))
		f.Add(kind, uint8(9), uint8(3), uint8(0), uint8(1), uint8(4), uint8(17))
	}

	f.Fuzz(func(t *testing.T, kind, width, height, x, y, reach, seed uint8) {
		g := fuzzGrid(kind, width, height)
		index, radius := fuzzIndex(g, x, y), int(reach%6)
		open := func(cell Cell[int]) bool {
			return (cell.Index().X*7+cell.Index().Y*13+int(seed))%5 != 2
		}

		everything, seen := g.Range(index, radius, nil), g.Range(index, radius, open)

		assert.True(t, Reg(g.AppendRange(nil, index, radius, open)).Equal(seen))
		assert.True(t, slices.IsSortedFunc(g.AppendRange(nil, index, radius, open), geom.Point[int].Compare))
		assert.Equal(t, slices.Collect(g.RangeSeq(index, radius, open)), g.AppendRange(nil, index, radius, open))

		if !g.Contains(index) {
			assert.True(t, everything.IsEmpty())
			assert.True(t, seen.IsEmpty())

			return
		}

		var within []geom.Point[int]
		for cell := range g.All() {
			if g.Distance(index, cell.Index()) <= radius {
				within = append(within, cell.Index())
			}
		}

		assert.True(t, everything.Equal(Reg(within)))

		if !open(g.Cell(index)) {
			assert.True(t, seen.IsEmpty())

			return
		}

		assert.True(t, seen.Contains(index))
		assert.True(t, everything.Encloses(seen))

		for other := range seen.All() {
			assert.True(t, open(g.Cell(other)), other.String()+": ")
			assert.True(t, g.Range(other, radius, open).Contains(index), other.String()+": ")
		}
	})
}

func TestGrid_AppendRange(t *testing.T) {
	grids, center := rangeGrids()
	open := func(cell Cell[int]) bool {
		return cell.Index() != geom.Pt(5, 6)
	}

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("appends the indices of Range", func(t *testing.T) {
				appended := g.AppendRange([]geom.Point[int]{geom.Pt(-9, -9)}, center, 3, open)

				assert.Equal(t, appended[0], geom.Pt(-9, -9))
				assert.Equal(t, len(appended)-1, g.Range(center, 3, open).Len())
				assert.True(t, Reg(appended[1:]).Equal(g.Range(center, 3, open)))
			})
			t.Run("appends in the order of the region, so Range has nothing to sort", func(t *testing.T) {
				for _, index := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
					assert.Equal(t, g.AppendRange(nil, index, 4, nil), g.Range(index, 4, nil).Elements(), index.String()+": ")
					assert.Equal(t, g.AppendRange(nil, index, 4, open), g.Range(index, 4, open).Elements(), index.String()+": ")
					assert.True(t, g.Range(index, 4, open).Equal(Reg(g.AppendRange(nil, index, 4, open))), index.String()+": ")
				}
			})
			t.Run("a radius beyond the grid appends what one across it does", func(t *testing.T) {
				assert.Equal(t, g.AppendRange(nil, center, math.MaxInt, open), g.AppendRange(nil, center, g.Width()+g.Height(), open))
			})
			t.Run("a negative radius appends nothing and asks nothing", func(t *testing.T) {
				asked := 0

				assert.Length(t, g.AppendRange(nil, center, -1, func(Cell[int]) bool {
					asked++

					return true
				}), 0)
				assert.Equal(t, asked, 0)
			})
			t.Run("an index outside the grid appends nothing", func(t *testing.T) {
				assert.Length(t, g.AppendRange(nil, geom.Pt(-1, 5), 3, nil), 0)
			})
			t.Run("an index that is not passable appends nothing", func(t *testing.T) {
				assert.Length(t, g.AppendRange(nil, geom.Pt(5, 6), 3, open), 0)
			})
			t.Run("allocates nothing with room where every cell passes", func(t *testing.T) {
				buffer := make([]geom.Point[int], 0, 128)

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = g.AppendRange(buffer[:0], center, 4, nil)
				}), 0.0)
			})
		})
	}
}

func TestGrid_RangeSeq(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("yields what AppendRange appends", func(t *testing.T) {
				assert.Equal(t, slices.Collect(g.RangeSeq(center, 3, nil)), g.AppendRange(nil, center, 3, nil))
			})
			t.Run("a radius beyond the grid yields what one across it does", func(t *testing.T) {
				assert.Equal(t, slices.Collect(g.RangeSeq(center, math.MaxInt, nil)), g.AppendRange(nil, center, g.Width()+g.Height(), nil))
			})
			t.Run("yields what AppendRange appends behind walls", func(t *testing.T) {
				open := func(cell Cell[int]) bool {
					return cell.Index() != geom.Pt(5, 6)
				}

				assert.Equal(t, slices.Collect(g.RangeSeq(center, 3, open)), g.AppendRange(nil, center, 3, open))
			})
			t.Run("stops at the break behind walls", func(t *testing.T) {
				open := func(cell Cell[int]) bool {
					return cell.Index() != geom.Pt(5, 6)
				}

				count := 0
				for range g.RangeSeq(center, 3, open) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
			t.Run("yields nothing for a negative radius", func(t *testing.T) {
				assert.Length(t, slices.Collect(g.RangeSeq(center, -1, nil)), 0)
				assert.Length(t, slices.Collect(g.RangeSeq(center, -1, func(Cell[int]) bool {
					return true
				})), 0)
			})
			t.Run("yields nothing from an index outside the grid", func(t *testing.T) {
				assert.Length(t, slices.Collect(g.RangeSeq(geom.Pt(-1, 5), 3, nil)), 0)
			})
			t.Run("stops at the break", func(t *testing.T) {
				count := 0
				for range g.RangeSeq(center, 3, nil) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
		})
	}
}

func TestGrid_Ring(t *testing.T) {
	grids, center := rangeGrids()

	t.Run("cardinal walks the diamond, by increasing angle", func(t *testing.T) {
		assert.Equal(t, grids["cardinal"].Ring(center, 1), grids["cardinal"].Neighbors(center))
	})
	t.Run("a radius of zero is the index alone", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, g.Ring(center, 0), []geom.Point[int]{center}, name+": ")
		}
	})
	t.Run("is exactly the cells at the distance of the radius", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				for radius := range 6 {
					var expected []geom.Point[int]
					for other := range g.All() {
						if g.Distance(index, other.Index()) == radius {
							expected = append(expected, other.Index())
						}
					}

					ring := g.Ring(index, radius)

					assert.Length(t, ring, len(expected), name+" at "+index.String()+": ")
					assert.True(t, Reg(ring).Equal(Reg(expected)), name+" at "+index.String()+": ")
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
	t.Run("allocates once, at the count of the ring", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, cap(g.Ring(geom.Pt(0, 0), 3)), len(g.Ring(center, 3)), name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = g.Ring(center, 3)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Ring(b *testing.B) {
	grids, center := rangeGrids()

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
	g := NewHexFlatTop[bool](geom.Sz(5, 5), HexFlatTopCellSize(32))

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

func TestGrid_AppendRing(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("appends what Ring returns", func(t *testing.T) {
				assert.Equal(t, g.AppendRing([]geom.Point[int]{geom.Pt(9, 9)}, geom.Pt(1, 1), 3), append([]geom.Point[int]{geom.Pt(9, 9)}, g.Ring(geom.Pt(1, 1), 3)...))
			})
			t.Run("appends nothing where Ring has none", func(t *testing.T) {
				assert.Length(t, g.AppendRing(nil, center, -1), 0)
				assert.Length(t, g.AppendRing(nil, center, 40), 0)
				assert.Length(t, g.AppendRing(nil, geom.Pt(-1, 5), 2), 0)
			})
			t.Run("allocates nothing with room", func(t *testing.T) {
				buffer := make([]geom.Point[int], 0, 64)

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = g.AppendRing(buffer, center, 3)
				}), 0.0)
			})
		})
	}
}

func TestGrid_RingSeq(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("yields what Ring returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(g.RingSeq(geom.Pt(1, 1), 3)), g.Ring(geom.Pt(1, 1), 3))
			})
			t.Run("yields nothing where Ring has none", func(t *testing.T) {
				assert.Length(t, slices.Collect(g.RingSeq(center, -1)), 0)
			})
			t.Run("stops at the break", func(t *testing.T) {
				count := 0
				for range g.RingSeq(center, 3) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
		})
	}
}

func TestGrid_Spiral(t *testing.T) {
	grids, center := rangeGrids()

	t.Run("starts at the index", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, g.Spiral(center, 3)[0], center, name+": ")
		}
	})
	t.Run("is the rings in turn, outward", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				var expected []geom.Point[int]
				for radius := range 5 {
					expected = g.AppendRing(expected, index, radius)
				}

				assert.Equal(t, g.Spiral(index, 4), expected, name+" at "+index.String()+": ")
			}
		}
	})
	t.Run("holds the cells of the range", func(t *testing.T) {
		for name, g := range grids {
			for _, index := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
				assert.True(t, Reg(g.Spiral(index, 4)).Equal(g.Range(index, 4, nil)), name+" at "+index.String()+": ")
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
	t.Run("allocates once, at the count of the range", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, cap(g.Spiral(geom.Pt(0, 0), 3)), len(g.Spiral(center, 3)), name+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = g.Spiral(center, 3)
			}), 1.0, name+": ")
		}
	})
}

func BenchmarkGrid_Spiral(b *testing.B) {
	grids, center := rangeGrids()

	for _, name := range []string{"cardinal", "diagonal", "flat-top", "pointy-top"} {
		b.Run(name, func(b *testing.B) {
			g := grids[name]

			for b.Loop() {
				sinkPoints = g.Spiral(center, 4)
			}
		})
	}
}

func TestGrid_AppendSpiral(t *testing.T) {
	grids, center := rangeGrids()

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("appends what Spiral returns", func(t *testing.T) {
				assert.Equal(t, g.AppendSpiral([]geom.Point[int]{geom.Pt(9, 9)}, geom.Pt(1, 1), 3), append([]geom.Point[int]{geom.Pt(9, 9)}, g.Spiral(geom.Pt(1, 1), 3)...))
			})
			t.Run("appends nothing where Spiral has none", func(t *testing.T) {
				assert.Length(t, g.AppendSpiral(nil, center, -1), 0)
				assert.Length(t, g.AppendSpiral(nil, geom.Pt(-1, 5), 2), 0)
			})
			t.Run("allocates nothing with room", func(t *testing.T) {
				buffer := make([]geom.Point[int], 0, 64)

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = g.AppendSpiral(buffer, center, 3)
				}), 0.0)
			})
		})
	}
}

func TestGrid_SpiralSeq(t *testing.T) {
	grids, center := rangeGrids()

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
		})
	}
}

func TestGrid_Distance(t *testing.T) {
	grids, center := rangeGrids()

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
	grids, center := rangeGrids()

	t.Run("a line to itself is the index alone", func(t *testing.T) {
		for name, g := range grids {
			assert.Equal(t, g.Line(center, center), []geom.Point[int]{center}, name+": ")
		}
	})
	t.Run("steps from one end to the other by neighbors", func(t *testing.T) {
		for name, g := range grids {
			for from := range g.All() {
				for _, to := range []geom.Point[int]{center, geom.Pt(0, 0), geom.Pt(10, 3)} {
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
	grids, _ := rangeGrids()

	for _, name := range []string{"cardinal", "diagonal", "flat-top", "pointy-top"} {
		b.Run(name, func(b *testing.B) {
			g := grids[name]

			for b.Loop() {
				sinkPoints = g.Line(geom.Pt(0, 0), geom.Pt(10, 3))
			}
		})
	}
}

func TestGrid_AppendLine(t *testing.T) {
	grids, _ := rangeGrids()

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("appends what Line returns", func(t *testing.T) {
				assert.Equal(t, g.AppendLine([]geom.Point[int]{geom.Pt(9, 9)}, geom.Pt(1, 1), geom.Pt(10, 3)), append([]geom.Point[int]{geom.Pt(9, 9)}, g.Line(geom.Pt(1, 1), geom.Pt(10, 3))...))
			})
			t.Run("allocates nothing with room", func(t *testing.T) {
				buffer := make([]geom.Point[int], 0, 64)

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = g.AppendLine(buffer, geom.Pt(0, 0), geom.Pt(10, 3))
				}), 0.0)
			})
		})
	}
}

func TestGrid_LineSeq(t *testing.T) {
	grids, _ := rangeGrids()

	for name, g := range grids {
		t.Run(name, func(t *testing.T) {
			t.Run("yields what Line returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(g.LineSeq(geom.Pt(1, 1), geom.Pt(10, 3))), g.Line(geom.Pt(1, 1), geom.Pt(10, 3)))
			})
			t.Run("stops at the break", func(t *testing.T) {
				count := 0
				for range g.LineSeq(geom.Pt(1, 1), geom.Pt(10, 3)) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
		})
	}
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
