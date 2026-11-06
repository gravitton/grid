# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/grid/compare/v1.3.0...main)
### Changed
- Require `container` v1.0.1

## [v1.3.0](https://github.com/gravitton/grid/compare/v1.2.0...v1.3.0) (2026-11-06)

One `Grid[T]` on rectangular, isometric and hexagonal layouts, and the queries a game asks of a map on all of them. `Range`, `Ring`, `Spiral`, `Line`, sight, the area queries and the searches answer in the indices of every layout, a set as a `Region` and a search as a `Field`, and `HexLayout` places hexes in the world with no grid around them.
A search allocates no node, the origin of a grid is the top-left corner of its bounds, a point on an edge goes to the same side in every cell, and every panic wraps a sentinel.

Every entry, with breaking changes marked, is in [docs/releases/v1.3.0.md](docs/releases/v1.3.0.md).

### Breaking
- The constructors are `NewRect`, `NewIsometric`, `NewHexFlatTop` and `NewHexPointyTop`, with options set in a callback on `SquareOptions` or `HexOptions`
- The origin of a grid is the top-left corner of its bounds, where it was the center of cell (0, 0)
- `Layout`, `Transform`, `NewGrid` and the custom layouts are removed; `Grid.MoveTo` and `Rescale` move and rescale a grid
- `System` is `Movement`, with the direction methods that were package functions, and neighbors come by increasing angle; `Direction` and the direction functions give way to those of `geometry`
- A `Cell` is passed by value and holds the value of its cell: `Grid.Get`, `At`, `Set`, `Has` and `Iter` are `Cell`, `CellAt`, `Cell.Set`, `Contains` and `DrawOrder`
- `Grid.Range` is the `Region` of the cells within a radius, and sight is `FieldOfView` with a `blocked` callback
- The searches are `Path<Algorithm>` and `Field<Algorithm>`, a path is its indices and a field a `Field`, and the callback types are removed
- `Bounds` is a `geom.Box`, `CellBounds` is `CellSize`, and the cell-size helpers are the `CellScale` helpers
- `Point` is `Square`, its sight takes `blocked` as in `hexagon`, it sees the same from either end and a diagonal wall blocks it
- `Array` and `Arr` moved to `github.com/gravitton/container`
- Every panic is an error wrapping a sentinel that `errors.Is` matches

### Added
- `Region` and `Reg`, with `Grid.Select`, `Border`, `Grow`, `Shrink`, `IsConnected` and `Components`
- `Field`, with `Grid.Field`, `Reachable` and the path, cost and step read off it
- `Grid.Neighbors`, `Ring`, `Spiral` and `Line`, each with its `Seq` form, and their short forms on `Cell`
- `Grid.HasLineOfSight` and `FieldOfView` on every layout
- `HexLayout` – hexes placed in the world with no grid, in any coordinate system
- `Grid.MoveTo`, `Rescale`, `Origin`, `CellScale`, `Layer` and `All`, and JSON for a grid
- The lattice methods of `Square`: neighbors, ring, spiral, sector, line, turns, reflections and sight, with `RangeLen`, `RingLen` and `SectorLen`
- `MovementNone`, `Movements`, `ParseMovement` and text marshalling for `Movement`
- `Index`, `Cell.String` and `Square.String`

### Changed
- A search allocates its field for the part of the grid it reaches and nothing per node: `PathAStar` corner to corner on a 64×64 grid went from 20 651 allocations to 5 and from about 117 µs to 12 µs
- A grid maps through the lattice of its kind, so an edge that is a whole number is found exactly and the boxes of the cells tile the bounds
- Require Go 1.27, `geometry` v1.18.0, `hexagon` v1.8.0 and `container` v1.0.0, and no longer `x`
- Doc comments are in one style, and CI runs the shared workflow of `gravitton/.github`

### Fixed
- A negative or NaN cost panics where a weighted search ran forever, and a cost of +Inf is a step never taken
- `IndexAt` puts a point on an edge on the same side in every cell, and the top-left corner of a grid is in it
- `DrawOrder` draws a flat-top grid with `EvenOffset` back to front, and picks the same cells on every architecture
- A cell scale, an origin or a size no grid can have panics, where it built a broken grid
- `Grid.Range` and `Spiral` near an edge allocate at the size of their result


## [v1.2.0](https://github.com/gravitton/grid/compare/v1.1.0...v1.2.0) (2026-05-12)
### Added
- `Array.Values` — iterator over all values (`iter.Seq[*T]`), completing the `Keys` / `Values` / `All` trio

### Changed
- `Array.Iter` renamed to `Array.Keys` — aligns with Go 1.23 iterator conventions (`iter.Seq[K]` iterators are named `Keys`)
- `Array.Iter2` renamed to `Array.All` — aligns with Go 1.23 iterator conventions (`iter.Seq2[K, V]` iterators are named `All`)
- `IterConfig` renamed to `IterOptions`


## [v1.1.0](https://github.com/gravitton/grid/compare/v1.0.0...v1.1.0) (2026-05-09)
### Added
- `IsometricRectCellSize(width)` — computes the cell size for `NewIsometricRectGrid` at a true 30° isometric angle
- `IsometricPixelPerfectRectCellSize(width)` — computes the cell size for `NewIsometricRectGrid` at a pixel-art-friendly 2:1 (width × width/2) ratio
- `HexFlatTopCellSize(width)` — computes the hex circumradius for `NewHexagonFlatTopGrid` where each tile is exactly `width` pixels wide
- `HexPointyTopCellSize(width)` — computes the hex circumradius for `NewHexagonPointyTopGrid` where each tile is exactly `width` pixels wide
- `RectCellSize(width)` — computes the cell size for `NewRectGrid` for a square `width × width` tile
- `HexFlatTopIsometricPixelPerfectCellSize(width)` — computes the hex circumradius for `NewHexagonFlatTopGrid` at a pixel-perfect isometric width
- `HexPointyTopIsometricPixelPerfectCellSize(width)` — computes the hex circumradius for `NewHexagonPointyTopGrid` at a pixel-perfect isometric width

### Changed
- `Grid.Iter` now automatically selects the correct draw order (painter's algorithm) based on the grid layout — no manual configuration required
- Isometric grids use diagonal depth (`col+row`) ordering; hex flat-top grids use a two-pass even/odd column iteration; hex pointy-top and rectangular grids use row-major order
- `Iter` includes one extra row and column at every edge of the bounds so partially visible tiles are not culled; border cells may be out-of-bounds (`cell.Valid() == false`)


## v1.0.0 (2026-05-06)
### Added
- `Grid[T]` — generic 2D grid with configurable layout, spatial mapping, and graph operations
- `Cell[T]` — single grid cell with value access, spatial info (`Center`, `Bounds`, `Polygon`), and graph methods (`Neighbors`, `DistanceTo`, `Range`, `PathTo`)
- `Array[T]` — low-level flat 2D array with `Get`, `Set`, `Fill`, `Clear`, `Clone`, `Iter`, `All`
- `NewRectGrid[T]` — rectangular grid with 4-directional (cardinal) movement
- `NewIsometricRectGrid[T]` — isometric (diamond-projection) rectangular grid
- `RectGridOpts.DiagonalMovement()` — option to enable 8-directional movement on rectangular grids
- `NewHexagonPointyTopGrid[T]` — hexagonal grid with pointy-top orientation and odd-row offset
- `NewHexagonFlatTopGrid[T]` — hexagonal grid with flat-top orientation and odd-column offset
- `HexGridOpts.EvenSystem()` — option to switch hex offset convention from odd to even
- `Direction` type and eight named constants (`E`, `NE`, `N`, `NW`, `W`, `SW`, `S`, `SE`) with `Opposite()` and `String()` methods
- `System` type — `Cardinal` (4-directional) and `Diagonal` (8-directional) movement modes
- `CardinalDirections`, `DiagonalDirections`, `Directions` — pre-built direction vector arrays; `Directions` is indexed by `Direction` constant
- `NeighborOffsets(system)` — returns direction vectors for the given system
- `NeighborOffset(system, direction)` — returns the unit vector for a single direction; panics if a diagonal direction is passed with `Cardinal`
- `DistanceTo(from, to, system)` — Manhattan distance for `Cardinal`, Chebyshev for `Diagonal`
- `Point` — grid coordinate type with spatial query methods: `Range`, `FieldOfView`, `HasLineOfSight`, `Neighbors`, `DistanceTo`
- `Pt(x, y)` — shorthand constructor for `Point`
- `Layout` — mapping between grid coordinates and pixel space, with `AlignTopLeft`, `MoveTo`, `Resize`, `Add`, `With`
- `Transform` — predefined coordinate mapping for a layout type; built-in: `SquareFlat`, `SquareIsometric`, `HexPointyTop`, `HexFlatTop`
- `NewLayout` / `NewTransform` — constructors for custom grid layouts
- Pathfinding on `Grid[T]`: `Path` (A*), `AStar`, `GreedyBestFirstSearch`, `UniformCostSearch`, `BreadthFirstSearch`, and `*Field` variants returning the full came-from map
- `Grid.Range` — returns all cell indices within Euclidean distance n, filtered through a caller-supplied validity function and a field-of-view algorithm; blocked cells occlude cells behind them (Bresenham on rect grids, hex FoV on hex grids)
- `Grid.Distance` — grid distance between two indices using the configured distance function
- `Grid.Iter` — iterator over all cells or a bounded viewport subset
