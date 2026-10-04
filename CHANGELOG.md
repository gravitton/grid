# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/grid/compare/v1.2.0...main)
### Changed
- Require Go 1.27
- `System` is renamed `Movement`, the name pathfinding libraries give the choice between four and eight neighbors, and its constants `Cardinal` and `Diagonal` are `MovementCardinal` and `MovementDiagonal`; `String` still prints `Cardinal` and `Diagonal` (**breaking**)
- `Movement` gained the direction methods that were package-level functions: `Movement.Directions()`, `Movement.Offsets()`, `Movement.Offset(direction)`, `Movement.Has(direction)` and `Movement.DistanceTo(from, to)`, matching the method style `geometry` v1.10.0 and `hexagon` v1.3.0 moved to (**breaking**)
- Neighbors are now ordered by increasing angle — `MovementCardinal` yields right, down, left, up (was east, north, west, south) — following the `Direction` reorder in `geometry` v1.11.0. The neighbor set is unchanged; only the visit order, and therefore pathfinding tie-breaks, differ (**breaking**)
- `Movement.Offsets` returns a fresh slice instead of one aliasing the package's own table, so callers can no longer mutate it (**breaking**)
- `Movement.Offset` returns the zero vector for a direction outside the system instead of panicking (**breaking**)
- `Movement.Directions`, `Has` and `DistanceTo` panic with `grid: unknown movement <n>` for a value outside the constants, where the message was `unsupported system`, and `Movement.String` prints `None` for such a value instead of `Unknown` (**breaking**)
- Bumped `geometry` to v1.15.0, `hexagon` to v1.5.0 and `assert` to v1.6.0

### Added
- `MovementNone`, `Movement.IsNone()`, `Movements()` and `ParseMovement(name)`, and text marshalling, so a `Movement` is stored in JSON and as a map key by the name `String` prints, as the `geom` and `hex` enums are
- `Movement.String()` — human-readable name for each system constant

### Removed
- `Direction` and its `E`, `NE`, `N`, `NW`, `W`, `SW`, `S`, `SE` constants — superseded by `geom.Direction` from `geometry` v1.10.0 (**breaking**)
- `Directions`, `CardinalDirections`, `DiagonalDirections` — superseded by `geom.Directions`, `geom.CardinalDirections`, `geom.DiagonalDirections` and `Movement.Directions` (**breaking**)
- `NeighborOffsets(system)`, `NeighborOffset(system, direction)` and `DistanceTo(from, to, system)` — replaced by the `Movement` methods of the same names (**breaking**)

### Fixed
- `Grid.Iter` no longer writes the resolved bounds back into the caller's `IterOptions`, which made a reused options value stick to the first grid's bounds
- `Point.HasLineOfSight` is now reciprocal — it walks from a canonical endpoint, so `a` sees `b` exactly when `b` sees `a`. A Bresenham walk traces a different chain of cells when reversed, which let one point see another that could not see it back (**breaking**)
- `Point.HasLineOfSight` no longer sees through a diagonal wall — a line crossing a cell corner is now opaque when both cells flanking that corner block, so a diagonal run of blockers hides what is behind it. Sight still passes the corner of a single isolated blocker, and `Grid.Range` inherits the fix on rectangular grids (**breaking**)
- `Point.FieldOfView` indexes the blocking points once per call instead of scanning the slice at every step of every line walk, cutting `Grid.Range` on a rectangular grid from ~6.5 ms to ~220 µs at radius 40
- `Grid.cellNeighbors` allocates the neighbor slice at its final capacity instead of growing it, saving one allocation per node expansion in every search algorithm (~14% off `Grid.AStar`)
- `Grid.Range` on a rectangular grid is now the cells within `n` steps of the grid's movement — a Manhattan diamond for cardinal movement, a Chebyshev square for diagonal — instead of a Euclidean disc, which held cells further than `n` steps from `n = 3` on and left out the corners of a diagonal grid. `Point.Range` is unchanged (**breaking**)
- `Grid.Range` no longer returns the cells that fail `valid`, nor indices outside the grid; both still hide what lies behind them (**breaking**)
- `BreadthFirstSearchField` and `UniformCostSearchField` no longer hold the cells that fail `valid`, and no search asks `cost` about such a cell (**breaking**)
- A nil `valid` or `cost` is accepted by every search, both fields and `Grid.Range`, admitting every cell at a cost of one per step, as `Grid.Path` already did; they panicked before
- `Array.Get` and `Array.Set` panic on an index outside either axis, as documented. An `x` past the width used to read or write a cell of the next row


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
