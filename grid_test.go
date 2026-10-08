package grid

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
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

func TestGrid_Layout(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))

	t.Run("maps as the grid does", func(t *testing.T) {
		layout := g.Layout()

		assert.Equal(t, layout.Size(), g.Size())
		geomtest.AssertPoint(t, layout.ToPoint(geom.Pt(2, 3)), g.Cell(geom.Pt(2, 3)).Center())
		assert.Equal(t, layout.FromPoint(geom.Pt(70.0, 40.0)), g.IndexAt(geom.Pt(70.0, 40.0)))
	})
	t.Run("moving the copy leaves the grid where it is", func(t *testing.T) {
		moved := g.Layout().MoveTo(geom.Pt(500.0, 500.0))

		geomtest.AssertPoint(t, moved.Bounds().Min, geom.Pt(500.0, 500.0))
		geomtest.AssertPoint(t, moved.ToPoint(geom.Pt(0, 0)), geom.Pt(532.0, 516.0))
		geomtest.AssertPoint(t, g.Cell(geom.Pt(0, 0)).Center(), geom.Pt(32.0, 16.0))
	})
}

func TestGrid_Bounds(t *testing.T) {
	t.Run("a rectangular grid spans its columns and rows of cells", func(t *testing.T) {
		g := NewRect[int](geom.Sz(6, 5), geom.Sz(64.0, 32.0))

		geomtest.AssertBox(t, g.Bounds(), geom.Box[float64]{Min: geom.Pt(0.0, 0.0), Max: geom.Pt(384.0, 160.0)})
	})
	t.Run("is the box around every cell", func(t *testing.T) {
		for name, g := range layoutGrids() {
			var around geom.Box[float64]
			for cell := range g.All() {
				if around.IsZero() {
					around = cell.Bounds()
				}

				around = around.Union(cell.Bounds())
			}

			geomtest.AssertBox(t, g.Bounds(), around, name+": ")
		}
	})
}

func TestGrid_CellSize(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))

	geomtest.AssertSize(t, g.CellSize(), geom.Sz(64.0, 32.0))
}

func TestGrid_CellSpacing(t *testing.T) {
	t.Run("a rectangular grid steps by a whole cell", func(t *testing.T) {
		g := NewRect[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))

		geomtest.AssertSize(t, g.CellSpacing(), geom.Sz(64.0, 32.0))
	})
	t.Run("an isometric grid steps by half a cell", func(t *testing.T) {
		g := NewIsometric[int](geom.Sz(5, 5), geom.Sz(64.0, 32.0))

		geomtest.AssertSize(t, g.CellSpacing(), geom.Sz(32.0, 16.0))
	})
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

func TestGrid_IndexAt(t *testing.T) {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))

	t.Run("is the cell the point lies in", func(t *testing.T) {
		assert.Equal(t, g.IndexAt(geom.Pt(64.0, 96.0)), geom.Pt(2, 3))
		assert.Equal(t, g.IndexAt(geom.Pt(95.0, 127.0)), geom.Pt(2, 3))
	})
	t.Run("a point outside the grid gives an index outside it", func(t *testing.T) {
		assert.False(t, g.Contains(g.IndexAt(geom.Pt(-10.0, 40.0))))
		assert.False(t, g.Contains(g.IndexAt(geom.Pt(40.0, 500.0))))
	})
	t.Run("the center of a cell gives its index", func(t *testing.T) {
		for name, g := range layoutGrids() {
			for cell := range g.All() {
				assert.Equal(t, g.IndexAt(cell.Center()), cell.Index(), name+" at "+cell.Index().String()+": ")
			}
		}
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
	t.Run("allocates three times more where a cell is refused, for the cells refused and the sight past them", func(t *testing.T) {
		open := func(cell Cell[int]) bool {
			return cell.Index() != geom.Pt(5, 6)
		}

		for name, g := range grids {
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkRegion = g.Range(center, 4, open)
			}), 4.0, name+": ")
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
