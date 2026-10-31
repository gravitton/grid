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

g := grid.NewRect[Tile](geom.Sz(100, 100), grid.RectCellScale(32))
g.Fill(Tile{Walkable: true, Cost: 1})
g.CellAt(cursor).Set(Tile{}) // a wall under the cursor

walkable := func(cell grid.Cell[Tile]) bool {
	return cell.Value().Walkable
}

path := g.Path(geom.Pt(0, 0), geom.Pt(10, 10), walkable, nil) // the indices from start to goal
```

### Grids

```go
size := geom.Sz(100, 100)

grid.NewRect[Tile](size, grid.RectCellScale(32))                 // squares, 4 neighbors
grid.NewIsometric[Tile](size, grid.IsometricPixelCellScale(64))  // 2:1 diamonds
grid.NewHexFlatTop[Tile](size, grid.HexFlatTopCellScale(64))     // odd columns shifted down
grid.NewHexPointyTop[Tile](size, grid.HexPointyTopCellScale(64)) // odd rows shifted right
```

Options are set in a callback, on [`RectOptions`][link-go-dev-reference] for the square grids and
`HexOptions` for the hexagonal ones:

```go
grid.NewRect[Tile](size, grid.RectCellScale(32), func(options *grid.RectOptions) {
	options.Movement = grid.MovementDiagonal // 8 neighbors
	options.Origin = geom.Pt(100.0, 50.0)    // the top-left corner of the grid in the world
})

grid.NewHexFlatTop[Tile](size, grid.HexFlatTopCellScale(64), func(options *grid.HexOptions) {
	options.EvenOffset = true // even columns shifted instead
})
```

A constructor takes the cell scale: the size of a square or a diamond, the circumradius per axis
of a hexagon. The helpers give it from the width a tile is drawn at:

| Helper                       | Tile, `w` wide                 |
|------------------------------|--------------------------------|
| `RectCellScale`              | Square, `w` tall               |
| `IsometricCellScale`         | Diamond at 30°, `w·0.577` tall |
| `IsometricPixelCellScale`    | Diamond, `w/2` tall            |
| `HexFlatTopCellScale`        | Regular hex, `w·√3/2` tall     |
| `HexPointyTopCellScale`      | Regular hex, `w·2/√3` tall     |
| `HexFlatTopPixelCellScale`   | Squashed hex, `w·3/4` tall     |
| `HexPointyTopPixelCellScale` | Squashed hex, `w` tall         |

Every constructor returns the same `*Grid[T]`, so everything below holds for all four layouts.

### Cells

```go
index := geom.Pt(3, 4)

g.Contains(index) // true, inside the grid
g.Size()          // Size{100, 100}; also Width and Height

cell := g.Cell(index)          // a Cell[Tile], a handle on the index
cell.Set(Tile{Cost: 3})        // a write outside the grid does nothing
cell.Value().Cost = 5          // a pointer into the grid, nil outside it
cell.IsValid()                 // true, the same answer as g.Contains
cell.Neighbors()               // the adjacent indices inside the grid; also g.Neighbors(index)
cell.DistanceTo(geom.Pt(0, 0)) // 7, in steps of the grid's movement
fmt.Println(cell)              // Cell(3,4)

g.Clone() // a copy with storage of its own; also Fill and Clear

for cell := range g.All() { // every cell, row by row
	cell.Value().Cost = 1
}
```

### World space

```go
g.IndexAt(geom.Pt(499.0, 123.4)) // the index of the cell under a pixel
g.CellAt(geom.Pt(499.0, 123.4))  // the same, as a cell

cell.Center()  // the pixel at the middle of the tile
cell.Bounds()  // the geom.Box around the tile
cell.Polygon() // the square, diamond or hexagon to draw or hit-test

g.Bounds()      // the geom.Box around the whole grid
g.CellSize()    // the size of one tile as drawn
g.CellSpacing() // the step between two tiles, smaller than the tile where they interlock

g.MoveTo(geom.Pt(100.0, 50.0))   // the top-left corner of the bounds, to scroll the grid
g.Resize(grid.RectCellScale(48)) // the scale of a tile, to zoom it
```

A `HexLayout` places hexagonal cells with no grid around them, for a map whose data lives
elsewhere, on an unbounded plane. Its indices are in the coordinate system it is built with,
axial or any other of `hex`:

```go
layout := grid.NewHexFlatTopLayout(grid.HexFlatTopCellScale(32), hex.Axial)
layout.MoveTo(geom.Pt(400.0, 300.0)) // the center of index zero

layout.Polygon(h.Point())              // where to draw a hex
hex.Axial.From(layout.IndexAt(cursor)) // the hex under the cursor
```

It places a cell exactly where a hexagonal grid of the same scale and system places it, and
puts a point on an edge on the same side.

### Pathfinding

```go
cost := func(current, next grid.Cell[Tile]) float64 {
	return next.Value().Cost
}

g.Path(from, to, walkable, cost) // the cheapest path as indices, both ends included; nil where there is none
cell.PathTo(to, walkable, cost)  // the same, from a cell
```

A nil callback means no constraint: every cell is passable, every step costs one. `Path` is A*;
the searches are also there by name:

| Method                                           | Algorithm         | Weighted | Shortest |
|--------------------------------------------------|-------------------|----------|----------|
| `AStar(start, goal, passable, cost)`             | A*                | yes      | yes      |
| `UniformCostSearch(start, goal, passable, cost)` | Dijkstra          | yes      | yes      |
| `BreadthFirstSearch(start, goal, passable)`      | Breadth-first     | no       | in steps |
| `GreedyBestFirstSearch(start, goal, passable)`   | Greedy best-first | no       | no       |

The two exhaustive ones also come as a `Field`: every cell the search reached, with the step
back toward the start and the cost of getting there, for many units heading to one place:

```go
field := g.UniformCostSearchField(goal, walkable, cost) // also BreadthFirstSearchField

next, ok := field.Next(unit) // one step closer to the goal; false at the goal and where it is not reached
field.CostTo(unit)           // what the way costs, and whether there is one
field.PathTo(unit)           // the whole path, from the goal to the unit; also AppendPathTo
field.Region()               // every cell reached
```

Where a unit can move to is the cells within its steps along passable ones:

```go
g.Reachable(unit, 4, walkable) // a Region, going around what stands in the way
```

### Range and visibility

```go
g.Distance(from, to)        // in steps of the grid's movement
g.Range(index, 3, walkable) // a Region: the cells within 3 steps that are in sight; nil admits every cell
cell.Range(3, walkable)     // the same, from a cell
```

A cell that is not passable is left out and hides what lies behind it, and a range from a cell
that is not passable itself, or from an index outside the grid, is empty. Two cells always see
each other or neither does. On a square grid a sightline across a corner is blocked only when
both cells beside the corner block: a diagonal wall is opaque, a single pillar is not. A
hexagonal grid uses the rule of [`gravitton/hexagon`][link-hexagon].

### Points

`grid.Point` is to a square grid what `hex.Hex` is to a hexagonal one: an index with the
lattice queries `geom.Point` does not have. A method whose answer depends on which cells are
adjacent takes the `Movement`:

```go
p := grid.Pt(5, 5)

p.Neighbors(grid.MovementDiagonal)           // all 8, by increasing angle
p.Neighbor(geom.DirectionUp)                 // Point{5, 4}
p.DistanceTo(target, grid.MovementCardinal)  // Manhattan; MovementDiagonal gives Chebyshev
p.DirectionTo(target, grid.MovementCardinal) // the nearest of the four directions

p.Range(2, grid.MovementCardinal)     // the indices within 2 steps: a diamond, or a square for MovementDiagonal
p.Ring(2, grid.MovementCardinal)      // the indices exactly 2 steps away, by increasing angle
p.Spiral(2, grid.MovementCardinal)    // the same set as Range, ordered outward
p.Line(target, grid.MovementDiagonal) // the indices to target, each a neighbor of the last
```

Each traversal has an `Append` form taking a buffer and a `Seq` form that walks without one:

```go
buffer := make([]geom.Point[int], 0, grid.RangeLen(2, grid.MovementDiagonal)) // 25; RingLen sizes a ring
buffer = p.AppendRange(buffer[:0], 2, grid.MovementDiagonal)

for near := range p.SpiralSeq(5, grid.MovementCardinal) { // nearest first, stops at the break
	if occupied[near] {
		break
	}
}
```

Sight takes the blockers as a slice, or as a function for a map or a grid to answer:

```go
p.HasLineOfSight(target, blocking)     // blocking is a slice of indices
p.FieldOfView(candidates, blocking)    // the candidates in sight; also Append and Seq
p.HasLineOfSightFunc(target, blocked)  // blocked is a func(geom.Point[int]) bool
p.FieldOfViewFunc(candidates, blocked) // also AppendFieldOfViewFunc and FieldOfViewFuncSeq
```

A quarter turn and a mirror keep a shape on the lattice:

```go
p.Turn(1)                       // Point{-5, 5}, a quarter turn around the origin, clockwise on a screen
p.TurnAround(geom.Pt(4, 4), -1) // the other way, around another index
p.ReflectX()                    // Point{5, -5}, mirrored across the x-axis; also ReflectY
```

### Drawing

```go
for cell := range g.DrawOrder(viewport) { // a zero box walks the whole grid
	if !cell.IsValid() {
		continue
	}

	draw(cell.Polygon(), cell.Value())
}
```

`DrawOrder` yields the tiles back to front, in the order the layout needs:

| Layout         | Order                                      |
|----------------|--------------------------------------------|
| Rectangular    | Row by row, top to bottom                  |
| Isometric      | By depth, `column + row` ascending         |
| Hex flat-top   | Row by row, the columns drawn higher first |
| Hex pointy-top | Row by row                                 |

A tile partly in view is included, so a cell may lie outside the grid.

### Regions

`Region` is an immutable set of indices, held once each and sorted. It knows nothing of the
grid they came from, so it serves square and hexagonal grids alike:

```go
reach := g.Reachable(unit, 3, walkable) // where the unit can move
blast := g.Range(target, 1, walkable)   // what a blast at the target hits

reach.Intersection(blast)     // also Union, Difference and SymmetricDifference, each one pass and one allocation
reach.Intersects(blast)       // whether they share a cell, without building it; also Encloses
reach.Contains(geom.Pt(4, 4)) // a binary search
reach.Len()                   // also IsEmpty and Equal

for index := range reach.All() { // by x, then by y
	highlight(g.Cell(index))
}

reach.Elements()    // the indices as a fresh slice; also AppendElements
json.Marshal(reach) // an array of points
```

It is `container.Set` over `geom.Point[int]`, from [`gravitton/container`][link-container],
and `grid.Reg` builds one from any slice of indices, in any order, with repeats. The same set
over hexes is `container.SetOf`, with no grid and no conversion to points:

```go
seen := container.SetOf(hex.Pt(0, 0).Range(3)...) // a set of hex.Hex
hit := container.SetOf(hex.Pt(3, -1).Range(1)...)

seen.Intersection(hit) // the hexes in both
seen.Encloses(hit)     // false: the blast reaches past what is seen
```

A set of hexes and a `Region` are different types, since a hexagonal `Grid` numbers its cells
by offset coordinates. What a hex walks, the grid walks in its own indices, on every layout:

```go
g.Ring(index, 2)   // the cells exactly two steps away, around the index
g.Spiral(index, 2) // the cells within two steps, nearest first, ring by ring
g.Line(from, to)   // the straight line between two indices, both included
```

Each has its `Append` and `Seq` forms. A ring and a spiral keep the cells inside the grid; a
line is measured between any two indices, as `Distance` is.

What depends on which cells are adjacent is asked of the grid, which knows:

```go
g.Border(reach)      // the cells of the region with a neighbor outside it, around a hole as well
g.Grow(reach, 1)     // the region and every cell within one step of it; Shrink wears it away
g.IsConnected(reach) // whether it is one piece
g.Components(reach)  // its pieces, as regions
```

`Range` also has `AppendRange` and `RangeSeq`, for a caller who wants the indices without a
region.

### Movement

A square grid moves in one of two ways, over the directions of
[`gravitton/geometry`][link-geometry]:

```go
grid.MovementCardinal.Directions()                             // Right, Down, Left, Up
grid.MovementDiagonal.Offsets()                                // the 8 steps as vectors, by increasing angle
grid.MovementCardinal.Offset(geom.DirectionUp)                 // Vector{0, -1}, the zero vector for a direction it lacks
grid.MovementCardinal.Contains(geom.DirectionUpRight)          // false
grid.MovementDiagonal.Distance(geom.Pt(0, 0), geom.Pt(3, 5))   // 5, Chebyshev; MovementCardinal gives 8

grid.Movements()                    // both, as a fresh array
grid.ParseMovement("Diagonal")      // the name back to the constant, "None" to MovementNone
json.Marshal(grid.MovementCardinal) // "Cardinal"
```

Full reference: [pkg.go.dev][link-go-dev-reference].

## Credits

- [Tomáš Novotný](https://github.com/tomas-novotny)
- [All Contributors][link-contributors]

## License

The MIT License (MIT). Please see [License File][link-licence] for more information.


[ico-license]:              https://img.shields.io/github/license/gravitton/grid.svg?style=flat-square&color=blue
[ico-workflow]:             https://img.shields.io/github/actions/workflow/status/gravitton/grid/main.yml?branch=main&style=flat-square
[ico-release]:              https://img.shields.io/github/v/release/gravitton/grid?style=flat-square&color=blue
[ico-go-dev-reference]:     https://img.shields.io/badge/go.dev-reference-blue?style=flat-square
[ico-coverage]:             https://img.shields.io/coverallsCoverage/github/gravitton/grid?style=flat-square

[link-release]:             https://github.com/gravitton/grid/releases
[link-contributors]:        https://github.com/gravitton/grid/contributors
[link-licence]:             ./LICENSE.md
[link-workflow]:            https://github.com/gravitton/grid/actions
[link-go-dev-reference]:    https://pkg.go.dev/github.com/gravitton/grid
[link-geometry]:            https://github.com/gravitton/geometry
[link-container]:           https://github.com/gravitton/container
[link-hexagon]:             https://github.com/gravitton/hexagon
[link-coverage]:            https://coveralls.io/github/gravitton/grid
