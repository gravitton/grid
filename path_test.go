package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

// --- Path ---

func TestGrid_Path_SameCell(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	path := g.Path(geom.Pt(2, 2), geom.Pt(2, 2), nil, nil)
	assert.Equal(t, len(path), 1)
	assert.Equal(t, path[0].Index(), geom.Pt(2, 2))
}

func TestGrid_Path_Straight(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	path := g.Path(geom.Pt(0, 0), geom.Pt(4, 0), nil, nil)
	assert.Equal(t, len(path), 5)
	assert.Equal(t, path[0].Index(), geom.Pt(0, 0))
	assert.Equal(t, path[4].Index(), geom.Pt(4, 0))
}

func TestGrid_Path_AroundWall(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 3), geom.Sz(32.0, 32.0))
	// Block the entire column 2 except row 2.
	g.Set(geom.Pt(2, 0), 1)
	g.Set(geom.Pt(2, 1), 1)
	walkable := func(c *Cell[int]) bool { return *c.Get() == 0 }
	path := g.Path(geom.Pt(0, 0), geom.Pt(4, 0), walkable, nil)
	// A path must exist going around the wall via row 2.
	assert.True(t, len(path) > 0)
	assert.Equal(t, path[0].Index(), geom.Pt(0, 0))
	assert.Equal(t, path[len(path)-1].Index(), geom.Pt(4, 0))
}

func TestGrid_Path_Blocked(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(3, 1), geom.Sz(32.0, 32.0))
	// Block the only path.
	g.Set(geom.Pt(1, 0), 1)
	walkable := func(c *Cell[int]) bool { return *c.Get() == 0 }
	path := g.Path(geom.Pt(0, 0), geom.Pt(2, 0), walkable, nil)
	assert.Equal(t, len(path), 0)
}

func TestGrid_Path_InvalidStart(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	path := g.Path(geom.Pt(-1, 0), geom.Pt(4, 4), nil, nil)
	assert.Equal(t, len(path), 0)
}

// --- Pathfinding algorithms (BFS, UCS, Greedy) ---

func TestGrid_BreadthFirstSearch(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	walkable := func(c *Cell[int]) bool { return true }
	path := g.BreadthFirstSearch(geom.Pt(0, 0), geom.Pt(4, 4), walkable)
	assert.True(t, len(path) > 0)
	assert.Equal(t, path[0], geom.Pt(0, 0))
	assert.Equal(t, path[len(path)-1], geom.Pt(4, 4))
}

func TestGrid_UniformCostSearch(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	walkable := func(c *Cell[int]) bool { return true }
	cost := func(cur, next *Cell[int]) float64 { return 1.0 }
	path := g.UniformCostSearch(geom.Pt(0, 0), geom.Pt(4, 4), walkable, cost)
	assert.True(t, len(path) > 0)
	assert.Equal(t, path[0], geom.Pt(0, 0))
	assert.Equal(t, path[len(path)-1], geom.Pt(4, 4))
}

func TestGrid_GreedyBestFirstSearch(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	walkable := func(c *Cell[int]) bool { return true }
	path := g.GreedyBestFirstSearch(geom.Pt(0, 0), geom.Pt(4, 4), walkable)
	assert.True(t, len(path) > 0)
	assert.Equal(t, path[0], geom.Pt(0, 0))
	assert.Equal(t, path[len(path)-1], geom.Pt(4, 4))
}

func TestGrid_BreadthFirstSearchField(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	walkable := func(c *Cell[int]) bool { return true }
	field := g.BreadthFirstSearchField(geom.Pt(2, 2), walkable)
	// The field should contain all reachable cells.
	assert.True(t, len(field) > 0)
	_, hasOrigin := field[geom.Pt(2, 2)]
	assert.True(t, hasOrigin)
}

func TestGrid_UniformCostSearchField(t *testing.T) {
	g := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	walkable := func(c *Cell[int]) bool { return true }
	cost := func(cur, next *Cell[int]) float64 { return 1.0 }
	field := g.UniformCostSearchField(geom.Pt(2, 2), walkable, cost)
	assert.True(t, len(field) > 0)
}

func TestGrid_Search(t *testing.T) {
	wall := geom.Pt(2, 2)
	start, goal := geom.Pt(0, 0), geom.Pt(4, 4)

	walled := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	walled.Set(wall, 1)

	sealed := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	for y := range 5 {
		sealed.Set(geom.Pt(2, y), 1)
	}

	walkable := func(cell *Cell[int]) bool {
		return *cell.Get() == 0
	}

	searches := map[string]func(g *Grid[int], valid ValidFunc[int], cost CostFunc[int]) []ints.Point{
		"BreadthFirstSearch": func(g *Grid[int], valid ValidFunc[int], cost CostFunc[int]) []ints.Point {
			return g.BreadthFirstSearch(start, goal, valid)
		},
		"UniformCostSearch": func(g *Grid[int], valid ValidFunc[int], cost CostFunc[int]) []ints.Point {
			return g.UniformCostSearch(start, goal, valid, cost)
		},
		"GreedyBestFirstSearch": func(g *Grid[int], valid ValidFunc[int], cost CostFunc[int]) []ints.Point {
			return g.GreedyBestFirstSearch(start, goal, valid)
		},
		"AStar": func(g *Grid[int], valid ValidFunc[int], cost CostFunc[int]) []ints.Point {
			return g.AStar(start, goal, valid, cost)
		},
	}

	for name, search := range searches {
		t.Run(name, func(t *testing.T) {
			t.Run("nil callbacks admit every cell at unit cost", func(t *testing.T) {
				path := search(walled, nil, nil)

				assert.Equal(t, path[0], start)
				assert.Equal(t, path[len(path)-1], goal)
			})

			t.Run("the path steps between neighbors and around the wall", func(t *testing.T) {
				path := search(walled, walkable, nil)

				assert.NotContains(t, path, wall)
				for i := 1; i < len(path); i++ {
					assert.Equal(t, walled.Distance(path[i-1], path[i]), 1)
				}
			})

			t.Run("a goal walled off has no path", func(t *testing.T) {
				assert.Equal(t, search(sealed, walkable, nil), nil)
			})
		})
	}

	t.Run("cost is never asked about a cell that is not valid", func(t *testing.T) {
		cost := func(current, next *Cell[int]) float64 {
			assert.NotEqual(t, next.Index(), wall)

			return 1
		}

		walled.UniformCostSearch(start, goal, walkable, cost)
		walled.AStar(start, goal, walkable, cost)
	})
}

func TestGrid_SearchField(t *testing.T) {
	wall := geom.Pt(2, 2)
	start := geom.Pt(0, 0)

	walled := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	walled.Set(wall, 1)

	sealed := NewRectGrid[int](geom.Sz(5, 5), geom.Sz(32.0, 32.0))
	for y := range 5 {
		sealed.Set(geom.Pt(2, y), 1)
	}

	walkable := func(cell *Cell[int]) bool {
		return *cell.Get() == 0
	}

	fields := map[string]func(g *Grid[int], valid ValidFunc[int]) map[ints.Point]ints.Point{
		"BreadthFirstSearchField": func(g *Grid[int], valid ValidFunc[int]) map[ints.Point]ints.Point {
			return g.BreadthFirstSearchField(start, valid)
		},
		"UniformCostSearchField": func(g *Grid[int], valid ValidFunc[int]) map[ints.Point]ints.Point {
			return g.UniformCostSearchField(start, valid, nil)
		},
	}

	for name, field := range fields {
		t.Run(name, func(t *testing.T) {
			t.Run("nil callbacks reach every cell", func(t *testing.T) {
				assert.Length(t, field(walled, nil), 25)
			})

			t.Run("a wall is never in the field", func(t *testing.T) {
				reached := field(walled, walkable)
				_, found := reached[wall]

				assert.False(t, found)
				assert.Length(t, reached, 24)
			})

			t.Run("holds only what was reached", func(t *testing.T) {
				assert.Length(t, field(sealed, walkable), 10)
			})
		})
	}
}
