<div align="center" width="100%">

<a href="https://github.com/gravitton">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gravitton/grid/refs/heads/main/docs/images/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gravitton/grid/refs/heads/main/docs/images/logo-light.svg">
  <img alt="Gravitton grid" src="https://raw.githubusercontent.com/gravitton/grid/refs/heads/main/docs/images/logo-light.svg" width="300">
</picture>
</a>

[![Latest Stable Version][ico-release]][link-release]
[![Build Status][ico-workflow]][link-workflow]
[![Coverage Status][ico-coverage]][link-coverage]
[![Go Dev Reference][ico-go-dev-reference]][link-go-dev-reference]
[![Software License][ico-license]][link-licence]

Generic 2D grid library for game development

<hr>

</div>


## Features

- **Generic** – `Grid[T]` holds any cell type, in one flat array.
- **Four layouts** – rectangular, isometric, hexagonal flat-top and pointy-top, behind one API.
- **World space** – a pixel to its cell and back, with the center, bounds and polygon of every cell.
- **Pathfinding** – A*, Dijkstra, greedy best-first and breadth-first search, with flow fields.
- **Visibility** – range with field of view, and a line of sight that reads the same from either end.
- **Draw order** – a viewport iterated back to front, whatever the layout.
- **Regions** – immutable sets of cells with union, intersection and difference.
- **Movement** – four or eight neighbors on a square grid, six on a hexagonal one.

The rules behind them are in the [package documentation][link-go-dev-reference].

## Installation

```shell
go get github.com/gravitton/grid
```

## Usage

```go
import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/grid"
)
```

```go
type Tile struct {
	Walkable bool
	Cost     float64
}

g := grid.NewRectGrid[Tile](geom.Sz(100, 100), grid.RectCellSize(32))
g.Fill(Tile{Walkable: true, Cost: 1})
g.At(cursor).Set(Tile{}) // a wall under the cursor

walkable := func(cell *grid.Cell[Tile]) bool {
	return cell.Get().Walkable
}

path := g.Path(geom.Pt(0, 0), geom.Pt(10, 10), walkable, nil) // []*Cell[Tile], start to goal
```

### Grids

```go
size := geom.Sz(100, 100)

grid.NewRectGrid[Tile](size, grid.RectCellSize(32))                               // squares, 4 neighbors
grid.NewIsometricRectGrid[Tile](size, grid.IsometricPixelPerfectRectCellSize(64)) // 2:1 diamonds
grid.NewHexagonFlatTopGrid[Tile](size, grid.HexFlatTopCellSize(64))               // odd columns shifted down
grid.NewHexagonPointyTopGrid[Tile](size, grid.HexPointyTopCellSize(64))           // odd rows shifted right

grid.NewRectGrid[Tile](size, grid.RectCellSize(32), grid.RectGridOpts.DiagonalMovement()) // 8 neighbors
grid.NewHexagonFlatTopGrid[Tile](size, geom.SzU(32.0), grid.HexGridOpts.EvenSystem())     // even columns shifted instead
```

A rectangular grid takes the size of a cell, a hexagonal one its circumradius per axis. The
helpers give either from the width a tile is drawn at:

| Helper                                      | Tile, `w` wide                 |
|---------------------------------------------|--------------------------------|
| `RectCellSize`                              | Square, `w` tall               |
| `IsometricRectCellSize`                     | Diamond at 30°, `w·0.577` tall |
| `IsometricPixelPerfectRectCellSize`         | Diamond, `w/2` tall            |
| `HexFlatTopCellSize`                        | Regular hex, `w·√3/2` tall     |
| `HexPointyTopCellSize`                      | Regular hex, `w·2/√3` tall     |
| `HexFlatTopIsometricPixelPerfectCellSize`   | Squashed hex, `w·3/4` tall     |
| `HexPointyTopIsometricPixelPerfectCellSize` | Squashed hex, `w` tall         |

Every constructor returns the same `*Grid[T]`, so everything below holds for all four layouts.

### Cells

```go
index := geom.Pt(3, 4)

g.Has(index)                // true, inside the grid
g.Set(index, Tile{Cost: 3}) // a write outside the grid does nothing
g.Size()                    // Size{100, 100}; also Width and Height

cell := g.Get(index)           // a *Cell[Tile], a handle on the index
cell.Get().Cost = 5            // a pointer into the grid, nil outside it
cell.Valid()                   // true, the same answer as g.Has
cell.Neighbors()               // the adjacent indices, those outside the grid included
cell.DistanceTo(geom.Pt(0, 0)) // 7, in steps of the grid's movement

g.Clone() // a deep copy; also Fill and Clear
```

### World space

```go
g.IndexAt(geom.Pt(499.0, 123.4)) // the index of the cell under a pixel
g.At(geom.Pt(499.0, 123.4))      // the same, as a cell

cell.Center()  // the pixel at the middle of the tile
cell.Bounds()  // the rectangle around the tile
cell.Polygon() // the square, diamond or hexagon to draw or hit-test

g.Bounds()      // the rectangle around the whole grid
g.CellBounds()  // the size of one tile
g.CellSpacing() // the step between two tiles, smaller than the tile where they interlock
```

### Pathfinding

```go
cost := func(current, next *grid.Cell[Tile]) float64 {
	return next.Get().Cost
}

g.Path(from, to, walkable, cost) // A*, as cells; nil for either callback means every cell, at cost 1
cell.PathTo(to, walkable, cost)  // the same, from a cell
```

The algorithms by name return indices:

| Method                                      | Algorithm         | Weighted | Shortest |
|---------------------------------------------|-------------------|----------|----------|
| `AStar(from, to, valid, cost)`              | A*                | yes      | yes      |
| `UniformCostSearch(from, to, valid, cost)`  | Dijkstra          | yes      | yes      |
| `BreadthFirstSearch(from, to, valid)`       | Breadth-first     | no       | in steps |
| `GreedyBestFirstSearch(from, to, valid)`    | Greedy best-first | no       | no       |

The two exhaustive ones also come as a field, the step back toward the start from every cell
reached, for many units heading to one place:

```go
field := g.UniformCostSearchField(goal, walkable, cost) // also BreadthFirstSearchField
next := field[unit]                                     // one step closer to the goal
```

### Range and visibility

```go
g.Distance(from, to)        // in steps of the grid's movement
g.Range(index, 3, walkable) // the indices within 3 that are in sight; valid is required
cell.Range(3, walkable)     // the same, from a cell
```

A cell that fails `valid` is left out and hides what lies behind it. On a square grid the
sightline is a Bresenham walk from a canonical end, so two cells always see each other or
neither does. Where it crosses a corner it is blocked only when both cells flanking the corner
block: a diagonal wall is opaque, a single pillar is not. A hexagonal grid uses the rule of
[`gravitton/hexagon`][link-hexagon].

The same queries exist without a grid, on `grid.Point`:

```go
p := grid.Pt(5, 5)

p.Range(2)                          // the indices within a Euclidean distance of 2
p.HasLineOfSight(target, blocking)  // blocking is a slice of indices
p.FieldOfView(candidates, blocking) // the candidates in sight
p.Neighbors(grid.Diagonal)          // all 8, by increasing angle
p.DistanceTo(target, grid.Cardinal) // Manhattan; Diagonal gives Chebyshev
```

### Drawing

```go
for cell := range g.Iter(&grid.IterOptions{Bounds: viewport}) { // nil walks the whole grid
	if !cell.Valid() {
		continue
	}

	draw(cell.Polygon(), cell.Get())
}
```

`Iter` yields the tiles back to front, in the order the layout needs:

| Layout         | Order                                |
|----------------|--------------------------------------|
| Rectangular    | Row by row, top to bottom            |
| Isometric      | By depth, `column + row` ascending   |
| Hex flat-top   | Two passes, even columns before odd  |
| Hex pointy-top | Row by row                           |

One row and column beyond the bounds are included on every side so a tile half in view is not
culled, which is why a cell may lie outside the grid.

### Movement systems

A square grid moves by one of two systems, over the directions of
[`gravitton/geometry`][link-geometry]:

```go
grid.Cardinal.Directions()                             // Right, Down, Left, Up
grid.Diagonal.Offsets()                                // the 8 steps as vectors, by increasing angle
grid.Cardinal.Offset(geom.DirectionUp)                 // Vector{0, -1}, the zero vector outside the system
grid.Cardinal.Has(geom.DirectionUpRight)               // false
grid.Diagonal.DistanceTo(geom.Pt(0, 0), geom.Pt(3, 5)) // 5, Chebyshev; Cardinal gives 8
```

### Arrays

`Array[T]` is the flat storage under a grid, for a 2D table without a layout:

```go
heat := grid.Arr[float64](geom.Sz(64, 64))
heat.Set(geom.Pt(1, 2), 0.5)
*heat.Get(geom.Pt(1, 2)) += 0.25

for index, value := range heat.All() { // also Keys and Values
	*value *= 0.9
}
```

Full reference: [pkg.go.dev][link-go-dev-reference].

## Credits

- [Tomáš Novotný](https://github.com/tomas-novotny)
- [All Contributors][link-contributors]

## License

The MIT License (MIT). Please see [License File][link-licence] for more information.


[ico-license]:              https://img.shields.io/github/license/gravitton/grid.svg?style=flat-square&colorB=blue
[ico-workflow]:             https://img.shields.io/github/actions/workflow/status/gravitton/grid/main.yml?branch=main&style=flat-square
[ico-release]:              https://img.shields.io/github/v/release/gravitton/grid?style=flat-square&colorB=blue
[ico-go-dev-reference]:     https://img.shields.io/badge/go.dev-reference-blue?style=flat-square
[ico-coverage]:             https://img.shields.io/coverallsCoverage/github/gravitton/grid?style=flat-square

[link-author]:              https://github.com/gravitton
[link-release]:             https://github.com/gravitton/grid/releases
[link-contributors]:        https://github.com/gravitton/grid/contributors
[link-licence]:             ./LICENSE.md
[link-changelog]:           ./CHANGELOG.md
[link-workflow]:            https://github.com/gravitton/grid/actions
[link-go-dev-reference]:    https://pkg.go.dev/github.com/gravitton/grid
[link-geometry]:            https://github.com/gravitton/geometry
[link-hexagon]:             https://github.com/gravitton/hexagon
[link-coverage]:            https://coveralls.io/github/gravitton/grid
