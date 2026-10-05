package grid

import (
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
	g.Set(testWall, 1)

	return g
}

// sealedGrid returns a 5x5 grid with a wall across it, parting the start from the goal.
func sealedGrid() *Grid[int] {
	g := NewRect[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	for y := range 5 {
		g.Set(geom.Pt(2, y), 1)
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
}

// testSearchField runs the sections every search field shares, from testStart.
func testSearchField(t *testing.T, field func(g *Grid[int], passable func(Cell[int]) bool) map[geom.Point[int]]geom.Point[int]) {
	t.Run("nil callbacks reach every cell", func(t *testing.T) {
		assert.Length(t, field(walledGrid(), nil), 25)
	})
	t.Run("holds the start", func(t *testing.T) {
		assert.Equal(t, field(walledGrid(), walkable)[testStart], testStart)
	})
	t.Run("a wall is never in the field", func(t *testing.T) {
		reached := field(walledGrid(), walkable)
		_, found := reached[testWall]

		assert.False(t, found)
		assert.Length(t, reached, 24)
	})
	t.Run("holds only what was reached", func(t *testing.T) {
		assert.Length(t, field(sealedGrid(), walkable), 10)
	})
}

func TestGrid_BreadthFirstSearchField(t *testing.T) {
	testSearchField(t, func(g *Grid[int], passable func(Cell[int]) bool) map[geom.Point[int]]geom.Point[int] {
		return g.BreadthFirstSearchField(testStart, passable)
	})
}

func TestGrid_BreadthFirstSearch(t *testing.T) {
	testSearch(t, func(g *Grid[int], passable func(Cell[int]) bool) []geom.Point[int] {
		return g.BreadthFirstSearch(testStart, testGoal, passable)
	})
}

func TestGrid_UniformCostSearchField(t *testing.T) {
	testSearchField(t, func(g *Grid[int], passable func(Cell[int]) bool) map[geom.Point[int]]geom.Point[int] {
		return g.UniformCostSearchField(testStart, passable, nil)
	})
}

func TestGrid_UniformCostSearch(t *testing.T) {
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
		g.Set(geom.Pt(2, 0), 1)
		g.Set(geom.Pt(2, 1), 1)

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
