package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
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
