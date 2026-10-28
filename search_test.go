package grid

import (
	"slices"
	"strconv"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

var (
	testStart = geom.Pt(0, 0)
	testGoal  = geom.Pt(4, 4)
	testWall  = geom.Pt(2, 2)
)

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

// testSearch runs the sections every search shares, from testStart to testGoal.
func testSearch(t *testing.T, search func(g *Grid[int], passable func(Cell[int]) bool) []geom.Point[int]) {
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
		for name, g := range regionGrids() {
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
			shortest := g.BreadthFirstSearch(testStart, index, walkable)

			assert.Length(t, path, len(shortest), index.String()+": ")
			assert.Equal(t, path[0], testStart, index.String()+": ")
			assert.Equal(t, path[len(path)-1], index, index.String()+": ")
		}
	})
}

func TestGrid_BreadthFirstSearchField(t *testing.T) {
	testSearchField(t, func(g *Grid[int], passable func(Cell[int]) bool) Field {
		return g.BreadthFirstSearchField(testStart, passable)
	})
}

func TestGrid_BreadthFirstSearch(t *testing.T) {
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []geom.Point[int] {
		return g.BreadthFirstSearch(testStart, testGoal, passable)
	})
}

func TestGrid_Reachable(t *testing.T) {
	t.Run("is the cells within the steps along passable cells", func(t *testing.T) {
		for name, g := range regionGrids() {
			for cell := range g.All() {
				if index := cell.Index(); (index.X*5+index.Y*3)%4 == 0 {
					cell.Set(1)
				}
			}
			start := geom.Pt(4, 5)
			field := g.BreadthFirstSearchField(start, walkable)

			for steps := range 6 {
				var expected []geom.Point[int]
				for index := range field.Region().All() {
					if len(field.PathTo(index))-1 <= steps {
						expected = append(expected, index)
					}
				}

				assert.True(t, g.Reachable(start, steps, walkable).Equal(Reg(expected)), name+" at "+strconv.Itoa(steps)+": ")
			}
		}
	})
	t.Run("goes around what a range sees past", func(t *testing.T) {
		g := sealedGrid()

		assert.False(t, g.Reachable(testStart, 20, walkable).Contains(testGoal))
		assert.Equal(t, g.Reachable(testStart, 20, walkable).Len(), 10)
	})
	t.Run("no steps is the start alone", func(t *testing.T) {
		assert.Equal(t, slices.Collect(walledGrid().Reachable(testStart, 0, nil).All()), []geom.Point[int]{testStart})
	})
	t.Run("negative steps reach nothing", func(t *testing.T) {
		assert.True(t, walledGrid().Reachable(testStart, -1, nil).IsEmpty())
	})
	t.Run("a start outside the grid reaches nothing", func(t *testing.T) {
		assert.True(t, walledGrid().Reachable(geom.Pt(-1, 0), 3, nil).IsEmpty())
	})
}

func TestGrid_UniformCostSearchField(t *testing.T) {
	testSearchField(t, func(g *Grid[int], passable func(Cell[int]) bool) Field {
		return g.UniformCostSearchField(testStart, passable, nil)
	})
}

func TestGrid_UniformCostSearch(t *testing.T) {
	t.Run("pays the least, around what is dear", func(t *testing.T) {
		path := weightedGrid().UniformCostSearch(testStart, geom.Pt(4, 0), nil, stepCost)

		assert.Length(t, path, 13)
		assert.NotContains(t, path, geom.Pt(2, 0))
	})
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []geom.Point[int] {
		return g.UniformCostSearch(testStart, testGoal, passable, nil)
	})
	t.Run("cost is never asked about a cell that is not passable", func(t *testing.T) {
		walledGrid().UniformCostSearch(testStart, testGoal, walkable, func(current, next Cell[int]) float64 {
			assert.NotEqual(t, next.Index(), testWall)

			return 1
		})
	})
}

func TestGrid_GreedyBestFirstSearch(t *testing.T) {
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []geom.Point[int] {
		return g.GreedyBestFirstSearch(testStart, testGoal, passable)
	})
}

func TestGrid_AStar(t *testing.T) {
	t.Run("pays what uniform cost search pays", func(t *testing.T) {
		for name, g := range regionGrids() {
			for cell := range g.All() {
				cell.Set((cell.Index().X*3 + cell.Index().Y*7) % 4)
			}

			for cell := range g.All() {
				expected, _ := g.UniformCostSearchField(testStart, nil, stepCost).CostTo(cell.Index())
				path := g.AStar(testStart, cell.Index(), nil, stepCost)

				paid := 0.0
				for i := 1; i < len(path); i++ {
					paid += stepCost(g.Cell(path[i-1]), g.Cell(path[i]))
				}

				assert.Equal(t, paid, expected, name+" at "+cell.Index().String()+": ")
			}
		}
	})
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []geom.Point[int] {
		return g.AStar(testStart, testGoal, passable, nil)
	})
	t.Run("cost is never asked about a cell that is not passable", func(t *testing.T) {
		walledGrid().AStar(testStart, testGoal, walkable, func(current, next Cell[int]) float64 {
			assert.NotEqual(t, next.Index(), testWall)

			return 1
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
}
