# Grid

[![Latest Stable Version][ico-release]][link-release]
[![Build Status][ico-workflow]][link-workflow]
[![Coverage Status][ico-coverage]][link-coverage]
[![Go Report Card][ico-go-report-card]][link-go-report-card]
[![Go Dev Reference][ico-go-dev-reference]][link-go-dev-reference]
[![Software License][ico-license]][link-licence]

Generic 2D grid library for game development with rectangular and hexagonal layouts, spatial queries, and pathfinding.


## Installation

```bash
go get github.com/gravitton/grid
```


## Usage

Rectangular grid:

```go
import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/grid"
)

type Tile struct {
	Walkable bool
	Cost     float64
}

g := grid.NewRectGrid[Tile](geom.Sz(100, 100), geom.Sz(32.0, 32.0))

g.Fill(Tile{Walkable: true, Cost: 1.0})

cell := g.At(geom.Pt(499.0, 123.4))
cell.Set(Tile{Walkable: false})
```

Hexagonal grid:

```go
g := grid.NewHexagonFlatTopGrid[Tile](geom.Sz(20, 20), geom.SzU(32.0))
```

Pathfinding:

```go
path := g.Path(
	geom.Pt(0, 0),
	geom.Pt(10, 10),
	func(c *grid.Cell[Tile]) bool {
		return c.Get().Walkable
	},
	func(current, next *grid.Cell[Tile]) float64 {
		return next.Get().Cost
	},
)
```

Iterating a viewport:

```go
for cell := range g.Iter(&grid.IterConfig{Bounds: viewport}) {
	draw(cell)
}
```


## API

Full documentation is available at [pkg.go.dev/github.com/gravitton/grid][link-go-dev-reference].

### Types

| Type | Description |
|---|---|
| `Grid[T]` | 2D grid with spatial mapping and graph operations |
| `Cell[T]` | Single cell — provides value access, spatial info, and pathfinding |
| `Array[T]` | Low-level flat 2D array |
| `Point` | Grid coordinate with spatial query methods (`Range`, `FieldOfView`, `HasLineOfSight`) |
| `Direction` | One of eight neighbor directions: `E`, `NE`, `N`, `NW`, `W`, `SW`, `S`, `SE` |
| `System` | Movement connectivity: `Cardinal` (4-dir) or `Diagonal` (8-dir) |

### Constructors

| Constructor | Description |
|---|---|
| `NewRectGrid[T](size, cellSize, opts...)` | Rectangular grid, 4-directional movement |
| `NewRectGrid[T](..., RectGridOpts.DiagonalMovement())` | Rectangular grid, 8-directional movement |
| `NewIsometricRectGrid[T](size, cellSize, opts...)` | Isometric (diamond) grid |
| `NewHexagonPointyTopGrid[T](size, hexSize, opts...)` | Hexagonal grid, pointy-top orientation |
| `NewHexagonFlatTopGrid[T](size, hexSize, opts...)` | Hexagonal grid, flat-top orientation |
| `Arr[T](size ints.Size)` | Standalone 2D array |

### Grid

```go
// Size
g.Size() ints.Size
g.Width() int
g.Height() int

// Spatial
g.Bounds() floats.Rectangle
g.CellBounds() floats.Size
g.CellSpacing() floats.Size

// Access
g.Get(index ints.Point) *Cell[T]
g.Has(index ints.Point) bool
g.Set(index ints.Point, value T)
g.At(point floats.Point) *Cell[T] // world-space → cell
g.IndexAt(point floats.Point) ints.Point

// Mutation
g.Fill(value T)
g.Clear()
g.Clone() *Grid[T]

// Iteration
g.Iter(config *IterConfig) iter.Seq[*Cell[T]] // config nil = all cells
```

### Cell

```go
c.Index() ints.Point
c.Valid() bool
c.Get() *T
c.Set(value T)
c.Center() floats.Point
c.Bounds() floats.Rectangle
c.Polygon() floats.RegularPolygon
c.Neighbors() []ints.Point
c.DistanceTo(to ints.Point) int
c.Range(n int, valid ValidFunc[T]) []ints.Point
c.PathTo(to ints.Point, valid ValidFunc[T], cost CostFunc[T]) []*Cell[T]
```

### Pathfinding

All algorithms accept an optional `ValidFunc` to mark cells as passable and an optional `CostFunc` for weighted traversal. `Path` uses A* internally.

| Method | Algorithm | Field variant |
|---|---|---|
| `Path(from, to, valid, cost)` | A* | — |
| `AStar(from, to, valid, cost)` | A* | — |
| `GreedyBestFirstSearch(from, to, valid)` | Greedy Best-First | — |
| `UniformCostSearch(from, to, valid, cost)` | Dijkstra | `UniformCostSearchField` |
| `BreadthFirstSearch(from, to, valid)` | BFS | `BreadthFirstSearchField` |

`*Field` variants return the full `map[ints.Point]ints.Point` came-from map instead of a single path.

### Array

```go
a.Size() ints.Size
a.Width() int
a.Height() int
a.Len() int
a.Has(index ints.Point) bool
a.Get(index ints.Point) *T
a.Set(index ints.Point, value T)
a.Fill(value T)
a.Clear()
a.Clone() Array[T]
a.Iter() iter.Seq[ints.Point]
a.Iter2() iter.Seq2[ints.Point, *T]
```

### Spatial range

```go
g.Distance(from, to ints.Point) int
g.Range(index ints.Point, n int, valid ValidFunc[T]) []ints.Point
```

Both rectangular and hexagonal grids apply a field-of-view algorithm to `Range` — blocked cells occlude cells behind them. On rectangular grids this uses Bresenham line-of-sight; on hexagonal grids it uses the hex FoV algorithm.

### Directions

```go
// Direction constants (ordered counterclockwise from East)
E, NE, N, NW, W, SW, S, SE Direction

// Direction vectors
CardinalDirections [4]ints.Vector  // E, N, W, S
DiagonalDirections [4]ints.Vector  // NE, NW, SW, SE
Directions         [8]ints.Vector  // all 8, indexed by Direction constant

d.Opposite() Direction
d.String() string

NeighborOffsets(system System) []ints.Vector
NeighborOffset(system System, direction Direction) ints.Vector
DistanceTo(from, to ints.Point, system System) int
```

### Point

`Point` is a grid coordinate with spatial query methods, mirroring the hex grid API:

```go
p := grid.Pt(x, y)
p.Range(n int) []ints.Point
p.FieldOfView(candidates, blocking []ints.Point) []ints.Point
p.HasLineOfSight(target ints.Point, blocking []ints.Point) bool
p.Neighbors(system System) []ints.Point
p.DistanceTo(to ints.Point, system System) int
p.Point() ints.Point
```

### Callbacks

```go
type ValidFunc[T any]  func(current *Cell[T]) bool
type CostFunc[T any]   func(current, next *Cell[T]) float64
```


## Credits

- [Tomáš Novotný](https://github.com/tomas-novotny)
- [All Contributors][link-contributors]


## License

The MIT License (MIT). Please see [License File][link-licence] for more information.


[ico-license]:              https://img.shields.io/github/license/gravitton/grid.svg?style=flat-square&colorB=blue
[ico-workflow]:             https://img.shields.io/github/actions/workflow/status/gravitton/grid/main.yml?branch=main&style=flat-square
[ico-release]:              https://img.shields.io/github/v/release/gravitton/grid?style=flat-square&colorB=blue
[ico-go-dev-reference]:     https://img.shields.io/badge/go.dev-reference-blue?style=flat-square
[ico-go-report-card]:       https://goreportcard.com/badge/github.com/gravitton/grid?style=flat-square
[ico-coverage]:             https://img.shields.io/coverallsCoverage/github/gravitton/grid?style=flat-square

[link-author]:              https://github.com/gravitton
[link-release]:             https://github.com/gravitton/grid/releases
[link-contributors]:        https://github.com/gravitton/grid/contributors
[link-licence]:             ./LICENSE.md
[link-changelog]:           ./CHANGELOG.md
[link-workflow]:            https://github.com/gravitton/grid/actions
[link-go-dev-reference]:    https://pkg.go.dev/github.com/gravitton/grid
[link-go-report-card]:      https://goreportcard.com/report/github.com/gravitton/grid
[link-coverage]:            https://coveralls.io/github/gravitton/grid
