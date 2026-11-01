package grid

import (
	"fmt"
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

func TestField_Start(t *testing.T) {
	t.Run("is where the search started", func(t *testing.T) {
		assert.Equal(t, walledGrid().BreadthFirstSearchField(testGoal, walkable).Start(), testGoal)
	})
	t.Run("the empty field starts at the origin", func(t *testing.T) {
		assert.Equal(t, Field{}.Start(), geom.Pt(0, 0))
	})
}

func TestField_Len(t *testing.T) {
	t.Run("counts the cells reached, the start included", func(t *testing.T) {
		assert.Equal(t, walledGrid().BreadthFirstSearchField(testStart, walkable).Len(), 24)
	})
	t.Run("the empty field reaches nothing", func(t *testing.T) {
		assert.Equal(t, Field{}.Len(), 0)
	})
}

func TestField_Next(t *testing.T) {
	g := walledGrid()
	field := g.BreadthFirstSearchField(testStart, walkable)

	t.Run("is a neighbor one step closer to the start", func(t *testing.T) {
		for index := range field.Region().All() {
			next, found := field.Next(index)
			if index == testStart {
				continue
			}

			assert.True(t, found, index.String()+": ")
			assert.Equal(t, g.Distance(index, next), 1, index.String()+": ")
			assert.Length(t, field.PathTo(next), len(field.PathTo(index))-1, index.String()+": ")
		}
	})
	t.Run("the start has none", func(t *testing.T) {
		_, found := field.Next(testStart)

		assert.False(t, found)
	})
	t.Run("a cell not reached has none", func(t *testing.T) {
		for _, index := range []Index{testWall, geom.Pt(-1, 0), geom.Pt(5, 5)} {
			_, found := field.Next(index)

			assert.False(t, found, index.String()+": ")
		}
	})
	t.Run("following it leads every cell to the start", func(t *testing.T) {
		for index := range field.Region().All() {
			steps := 0
			for current := index; current != testStart; steps++ {
				current, _ = field.Next(current)
			}

			assert.Equal(t, steps, len(field.PathTo(index))-1, index.String()+": ")
		}
	})
}

func TestField_CostTo(t *testing.T) {
	t.Run("a search that counts steps costs the steps", func(t *testing.T) {
		field := walledGrid().BreadthFirstSearchField(testStart, walkable)

		for index := range field.Region().All() {
			cost, found := field.CostTo(index)

			assert.True(t, found, index.String()+": ")
			assert.Equal(t, cost, float64(len(field.PathTo(index))-1), index.String()+": ")
		}
	})
	t.Run("a weighted search costs what its path costs", func(t *testing.T) {
		g := weightedGrid()
		field := g.UniformCostSearchField(testStart, nil, stepCost)

		for index := range field.Region().All() {
			path := field.PathTo(index)
			paid := 0.0
			for i := 1; i < len(path); i++ {
				paid += stepCost(g.Cell(path[i-1]), g.Cell(path[i]))
			}
			cost, _ := field.CostTo(index)

			assert.Equal(t, cost, paid, index.String()+": ")
		}
	})
	t.Run("goes around what is dear", func(t *testing.T) {
		cost, _ := weightedGrid().UniformCostSearchField(testStart, nil, stepCost).CostTo(geom.Pt(4, 0))

		assert.Equal(t, cost, 12.0)
	})
	t.Run("the start costs nothing", func(t *testing.T) {
		cost, found := weightedGrid().UniformCostSearchField(testStart, nil, stepCost).CostTo(testStart)

		assert.True(t, found)
		assert.Equal(t, cost, 0.0)
	})
	t.Run("a cell not reached has no cost", func(t *testing.T) {
		_, found := walledGrid().BreadthFirstSearchField(testStart, walkable).CostTo(testWall)

		assert.False(t, found)
	})
}

func TestField_PathTo(t *testing.T) {
	g := walledGrid()
	field := g.BreadthFirstSearchField(testStart, walkable)

	t.Run("runs from the start to the index by neighbors", func(t *testing.T) {
		path := field.PathTo(testGoal)

		assert.Equal(t, path[0], testStart)
		assert.Equal(t, path[len(path)-1], testGoal)
		assert.Length(t, path, g.Distance(testStart, testGoal)+1)
		for i := 1; i < len(path); i++ {
			assert.Equal(t, g.Distance(path[i-1], path[i]), 1)
		}
	})
	t.Run("to the start is the start alone", func(t *testing.T) {
		assert.Equal(t, field.PathTo(testStart), []Index{testStart})
	})
	t.Run("to a cell not reached is nil", func(t *testing.T) {
		assert.Equal(t, field.PathTo(testWall), nil)
		assert.Equal(t, Field{}.PathTo(testStart), nil)
	})
	t.Run("allocates once at its length", func(t *testing.T) {
		path := field.PathTo(testGoal)

		assert.Equal(t, cap(path), len(path))
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkPoints = field.PathTo(testGoal)
		}), 1.0)
	})
}

func BenchmarkField_PathTo(b *testing.B) {
	benchmarkGrids(b, func(b *testing.B, g *Grid[int]) {
		field := g.BreadthFirstSearchField(geom.Pt(0, 0), nil)

		for b.Loop() {
			sinkPoints = field.PathTo(geom.Pt(63, 63))
		}
	})
}

func TestField_AppendPathTo(t *testing.T) {
	field := walledGrid().BreadthFirstSearchField(testStart, walkable)

	t.Run("appends the indices of PathTo", func(t *testing.T) {
		prefix := []Index{geom.Pt(-9, -9)}

		assert.Equal(t, field.AppendPathTo(prefix, testGoal), append(prefix, field.PathTo(testGoal)...))
	})
	t.Run("to a cell not reached appends nothing", func(t *testing.T) {
		assert.Length(t, field.AppendPathTo(nil, testWall), 0)
	})
	t.Run("allocates nothing with room", func(t *testing.T) {
		buffer := make([]Index, 0, 16)

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkPoints = field.AppendPathTo(buffer[:0], testGoal)
		}), 0.0)
	})
}

func TestField_Contains(t *testing.T) {
	field := walledGrid().BreadthFirstSearchField(testStart, walkable)

	t.Run("a cell reached", func(t *testing.T) {
		assert.True(t, field.Contains(testStart))
		assert.True(t, field.Contains(testGoal))
	})
	t.Run("a wall", func(t *testing.T) {
		assert.False(t, field.Contains(testWall))
	})
	t.Run("outside the grid", func(t *testing.T) {
		for _, index := range []Index{geom.Pt(-1, 0), geom.Pt(0, -1), geom.Pt(5, 0), geom.Pt(0, 5)} {
			assert.False(t, field.Contains(index), index.String()+": ")
		}
	})
	t.Run("the empty field holds nothing", func(t *testing.T) {
		assert.False(t, Field{}.Contains(testStart))
	})
}

func TestField_IsEmpty(t *testing.T) {
	t.Run("a search that started is not", func(t *testing.T) {
		assert.False(t, walledGrid().BreadthFirstSearchField(testStart, walkable).IsEmpty())
	})
	t.Run("a search from outside the grid is", func(t *testing.T) {
		assert.True(t, walledGrid().BreadthFirstSearchField(geom.Pt(-1, 0), walkable).IsEmpty())
	})
	t.Run("the zero value is", func(t *testing.T) {
		assert.True(t, Field{}.IsEmpty())
	})
}

func TestField_All(t *testing.T) {
	field := NewRect[int](geom.Sz(3, 3), geom.SzU(32.0)).BreadthFirstSearchField(geom.Pt(1, 1), nil)

	t.Run("yields what Region holds, in its order", func(t *testing.T) {
		assert.Equal(t, slices.Collect(field.All()), field.Region().Elements())
	})
	t.Run("the empty field yields nothing", func(t *testing.T) {
		assert.Length(t, slices.Collect(Field{}.All()), 0)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range field.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
}

func TestField_Region(t *testing.T) {
	field := sealedGrid().BreadthFirstSearchField(testStart, walkable)

	t.Run("is the cells reached", func(t *testing.T) {
		region := field.Region()

		assert.Equal(t, region.Len(), field.Len())
		for index := range region.All() {
			assert.True(t, field.Contains(index), index.String()+": ")
		}
	})
	t.Run("is sorted as a region is", func(t *testing.T) {
		assert.True(t, slices.IsSortedFunc(field.Region().Elements(), Index.Compare))
	})
	t.Run("the empty field is the empty region", func(t *testing.T) {
		assert.True(t, Field{}.Region().IsEmpty())
	})
	t.Run("allocates once", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkRegion = field.Region()
		}), 1.0)
	})
}

func ExampleField() {
	g := NewRect[bool](geom.Sz(5, 3), RectCellScale(32))
	g.Fill(true)
	g.Cell(geom.Pt(2, 0)).Set(false)
	g.Cell(geom.Pt(2, 1)).Set(false)

	walkable := func(cell Cell[bool]) bool {
		return *cell.Value()
	}

	field := g.BreadthFirstSearchField(geom.Pt(4, 0), walkable)

	fmt.Println(field.Next(geom.Pt(0, 0)))
	fmt.Println(field.CostTo(geom.Pt(0, 0)))
	fmt.Println(field.PathTo(geom.Pt(0, 0)))
	fmt.Println(field.Contains(geom.Pt(2, 0)), field.Region().Len())
	// Output:
	// (0,1) true
	// 8 true
	// [(4,0) (4,1) (4,2) (3,2) (2,2) (1,2) (0,2) (0,1) (0,0)]
	// false 13
}
