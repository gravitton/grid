# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/grid/compare/v1.2.0...main)
### Changed
- Require `assert` v1.7.0 for the tests
- `Array[T]` is `Array[T, S]`, generic over the integer type of its sides and indices, as `geom` is generic over its number: `Size`, `Width`, `Height`, `Get`, `Set`, `Contains`, `Keys` and `All` speak `S`, while `Len` stays an `int`. `Arr` takes the size as `geom.Size[S]` and infers `S` from it, so `Arr[float64](geom.Sz(64, 64))` is still an `int` array and `Arr[float64](geom.Sz[int32](64, 64))` an `int32` one. A narrower `S` makes the array a smaller value, and `Get` and `Set` on an `int32` array take a half to two thirds of the time they take on an `int` one. `Grid` stores its cells with `int` and is unchanged (**breaking**)
- Every panic of the package is an error, where it was a string, with the same message: `Arr`, `Array.Get`, `Array.Set`, the methods with no answer for a movement outside the constants, `MoveTo`, `Resize` and the `HexLayout` constructors. A panic of `geom` passed through, as `IndexAt` lets `geom.Cast` through, is still a string (**breaking** for a caller that recovers a string)
- Require Go 1.27
- Require `github.com/gravitton/container` v1.0.0, the storage of a grid and the set under `Region`
- `System` is renamed `Movement`, the name pathfinding libraries give the choice between four and eight neighbors, and its constants `Cardinal` and `Diagonal` are `MovementCardinal` and `MovementDiagonal`; `String` still prints `Cardinal` and `Diagonal` (**breaking**)
- `Movement` gained the direction methods that were package-level functions: `Movement.Directions()`, `Movement.Offsets()`, `Movement.Offset(direction)`, `Movement.Contains(direction)` and `Movement.DistanceTo(from, to)`, matching the method style `geometry` v1.10.0 and `hexagon` v1.3.0 moved to (**breaking**)
- Neighbors are now ordered by increasing angle — `MovementCardinal` yields right, down, left, up (was east, north, west, south) — following the `Direction` reorder in `geometry` v1.11.0. The neighbor set is unchanged; only the visit order, and therefore pathfinding tie-breaks, differ (**breaking**)
- `Movement.Offsets` returns a fresh slice instead of one aliasing the package's own table, so callers can no longer mutate it (**breaking**)
- `Movement.Offsets` allocates once, where it built its directions and then its steps
- `Movement.Offset` returns the zero vector for a direction outside the system instead of panicking (**breaking**)
- A movement outside the constants answers as one with no neighbors where it panicked with `unsupported system`: `Movement.Directions` and `Offsets` return nil, `Contains` false and `Offset` the zero vector. `Movement.DistanceTo` has no answer and panics with `grid: unknown movement <n>`, and `Movement.String` prints `None` for such a value instead of `Unknown` (**breaking**)
- The constructors no longer repeat the package name: `NewRectGrid`, `NewIsometricRectGrid`, `NewHexagonFlatTopGrid` and `NewHexagonPointyTopGrid` are `NewRect`, `NewIsometric`, `NewHexFlatTop` and `NewHexPointyTop`; the option types `RectGridOption` and `HexGridOption` are removed (**breaking**)
- Options are a callback that sets the fields of an options struct, as the `ui` package of the engine takes them: `NewRect` and `NewIsometric` take `func(*RectOptions)`, with `Movement` and `Origin`, and `NewHexFlatTop` and `NewHexPointyTop` take `func(*HexOptions)`, with `EvenOffset` and `Origin`. `RectGridOpts.DiagonalMovement()` is `options.Movement = MovementDiagonal` and `HexGridOpts.EvenSystem()` is `options.EvenOffset = true`; the `RectGridOpts` and `HexGridOpts` variables are removed (**breaking**)
- `Grid.Set(index, value)` is removed: the value of a cell is read and written through the cell alone, `g.Cell(index).Set(value)` beside `g.Cell(index).Value()`, as its size, center, bounds and polygon already were. `Grid` keeps what relates cells or covers them all (**breaking**)
- `Grid.Get(index)` is `Grid.Cell(index)`, `Grid.At(point)` is `Grid.CellAt(point)` and `Cell.Get()` is `Cell.Value()`, so `Get` no longer names three different things; `Array.Get` is unchanged (**breaking**)
- `Cell.Valid` is `Cell.IsValid`, the `Is` prefix every state check without an argument takes, as `IsEmpty` and `IsNone` do (**breaking**)
- `Grid.Has`, `Array.Has` and `Movement.Has` are `Contains`, as `Region.Contains` and the `geom` shapes say it (**breaking**)
- `Grid.Iter(options)` is `Grid.DrawOrder(bounds)`, taking the rectangle itself, a zero one for the whole grid; `IterOptions` is removed (**breaking**)
- The callback types `ValidFunc`, `ValidIndexFunc` and `CostFunc` are removed; every method spells the function type in its signature, and the parameter that was `valid` is `passable` (**breaking**)
- Signatures spell the `geom` types in full, `geom.Point[int]` and `geom.Size[float64]`, instead of the aliases of `geometry/types/ints` and `floats`; the types are the same, so no caller changes
- `Point.Range(n)` is `Point.Range(n, movement)` and returns the indices within `n` steps of the movement, sorted by x and then by y and allocated once, instead of a Euclidean disc. It is what a rectangular `Grid.Range` is built on, as the hexagonal one is built on `hex.Hex.Range` (**breaking**)
- `Grid.Bounds` and `Cell.Bounds` return a `geom.Box[float64]`, and `Grid.DrawOrder` takes one, where they used `geom.Rectangle[float64]`: a bounds is the axis-aligned extent `geom` v1.15.0 gives every shape, with `Min` and `Max` as fields; `box.Rectangle()` converts back (**breaking**)
- `Grid.CellBounds` is `CellSize`: it returns a size, the width and height of a cell as drawn, as `Cell.Size`, `CellWidth` and `CellHeight` already say, while every `Bounds` is now a box (**breaking**)
- A `Cell` is passed by value: `Grid.Cell`, `Grid.CellAt`, `Grid.All` and `Grid.DrawOrder` give a `Cell[T]` and the `passable` and `cost` callbacks take one, where they used `*Cell[T]`, which was allocated for every call. `Cell.Value` still returns the pointer into the grid (**breaking**)
- `Grid.Path` and `Cell.PathTo` return the indices of the path, `[]geom.Point[int]`, as every search by name does, instead of `[]*Cell[T]`; `Grid.Cell(index)` gives the cell of one (**breaking**)
- A search no longer builds a slice of neighbors for every cell it expands, and a grid no longer calls through function values for its distance, range and world mapping. With the cells passed by value, `AStar` corner to corner on a 64×64 rectangular grid went from 20 651 allocations to 2 561, on a hexagonal one from 10 939 to 1 049, `Grid.Range` at 8 steps from 200 to 10, and `Grid.DrawOrder` over 64×64 cells from 4 228 to 3
- `BreadthFirstSearchField` and `UniformCostSearchField` return a `Field` instead of a `map[geom.Point[int]]geom.Point[int]`: `field.Next(index)` is the step the map held (**breaking**)
- `Grid.Range` and `Cell.Range` return a `Region` instead of a slice of indices; `Grid.AppendRange` and `Grid.RangeSeq` give the indices without one (**breaking**)
- `Grid.Range` and `Cell.Neighbors` are empty from an index outside the grid, where they held the cells of the grid beside it, and `Range` is empty from an index that is not passable itself, where it held the cells around it: nothing is seen from where nothing can stand, as a search from there finds nothing (**breaking**)
- A search keeps what it has reached in slices the size of the grid instead of two maps, and its frontier by value in a heap of its own. `AStar` corner to corner on a 64×64 rectangular grid went from 2 561 allocations to 14 and a hexagonal one from 1 049 to 13, whatever the length of the path; `Grid.Range` at 8 steps went from 10 to 4, and at 40 steps on a 101×101 grid from about 420 µs to about 315 µs, since it asks the grid about each cell instead of collecting the blockers first
- `Grid.Range` asks `passable` about each cell within the radius once, where it asked again at every step of every sightline, walks a sightline only where a cell was refused, since nothing can be hidden otherwise, and gathers its indices in the order of the region, so it sorts nothing. On a 101×101 grid a range of 40 steps that is refused nothing, with a nil `passable` or one that admits every cell, went from about 420 µs to 13 µs and a range of 8 steps from 5.1 µs to 0.7 µs, on a flat-top hexagonal grid from 1.45 ms to 18 µs and from 18 µs to 1.0 µs; behind walls a rectangular range of 40 steps went from about 265 µs to 165 µs. It allocates once, four times where a cell is refused, where it always allocated four, and `Grid.AppendRange` nothing with room where nothing is refused. `Grid.RangeSeq` gathers the range before it yields any of it
- The package no longer depends on `github.com/gravitton/x`: its heap and its queue are replaced by a frontier and a queue of the search's own, and the set is `container.Set`
- The origin of a grid is the top-left corner of its bounds on every kind of grid, where it was the center of cell (0, 0): `RectOptions.Origin`, `HexOptions.Origin` and `Grid.MoveTo` place `Bounds().Min`, and `Grid.Resize` keeps it. An isometric grid moved right by the reach of its rows and a hexagonal one with `EvenOffset` half a cell down or right, so their bounds start at the origin as those of the others did (**breaking**)
- A grid maps through the lattice of its kind, every center and edge a whole number of steps from the origin, where it multiplied by a matrix each way: a cell edge that is a whole number is found exactly, where about one in seventy fell in the cell before it on whole cell sizes
- The mapping between an index and the world is about twice as fast on every layout. On a 64×64 grid `Cell.Center` went from about 6 ns to 2–3 ns, `Cell.Bounds` from 10 ns to 3–4 ns, `Cell.Polygon` from about 19 ns to 7–8 ns, and `Grid.IndexAt` from 10–14 ns to 4 ns on a rectangular grid, 3 ns on an isometric one and about 7 ns on a hexagonal one. None of them allocates
- `Grid.Bounds` reads a box computed with the grid and allocates nothing, where it allocated one to three times
- The parameters of `NewRect` and `NewIsometric` are named `size` and `cellSize`, where they were `grid` and `size`; no caller changes
- The circumradius a kind of grid gives the polygon of its cells is one number, the private `kind.cellRadius`, where `kind.radius` returned it twice as a size, and a kind counts the whole box of a cell in steps, `kind.sizeSteps`, where `kind.halfSteps` counted half of it; no polygon and no bounds change
- `Grid.Range` on a hexagonal grid finds what is in sight past a refused cell in one sweep outward with `hex.Hex.FieldOfViewRangeSeq`, where it walked a sightline to every cell kept: behind walls on a 101×101 grid a range of 8 steps went from about 10 µs to 4 µs and one of 40 steps from about 400 µs to 65 µs. A refused rectangular range allocates three times where it allocated four
- `Point.FieldOfView` and `FieldOfViewFunc` return nil for no candidates, where they returned an empty slice, as `hex.Hex.FieldOfView` does since hexagon v1.6.0
- Bumped `geometry` to v1.17.0, `hexagon` to v1.6.0 and `assert` to v1.6.0
- Rewrote the doc comments to one style, with the shared rules, every panic and the order of directions stated once in the package doc

### Added
- An example for `Movement`
- `Array` marshals to JSON as `{"w":…,"h":…,"d":[…]}`, its values row by row, and reads back with an error for a negative size or for values that do not fill it
- `Cell.String` prints a cell as `Cell(x,y)`
- Examples for `Arr`, `Grid.Cell`, `Grid.IndexAt` and `Grid.Range`, benchmarks for `Array.Get`, `Array.Set`, `Grid.All`, `Grid.Neighbors` and `Grid.Range`, and the fuzz targets `FuzzArray_Set`, `FuzzGrid_Neighbors` and `FuzzGrid_Range`, which check a range against the distances and sight both ways on every kind of grid
- `Grid.Neighbors(index)` and `Grid.AppendNeighbors(dst, index)` — the indices adjacent to an index that lie inside the grid, which only `Cell.Neighbors` gave before, and `Cell.AppendNeighbors(dst)`. They are exactly the cells at distance one, in the order of the neighbor rule
- `HexLayout`, built by `NewHexFlatTopLayout(hexSize, system)` and `NewHexPointyTopLayout(hexSize, system)`: hexagonal cells placed in the world with no grid around them, on an unbounded plane, indexed in any `hex.CoordinateSystem`, with `Center`, `Bounds`, `Polygon`, `IndexAt`, `CellSize`, `CellSpacing`, `MoveTo` and `Resize`. A cell lands exactly where a hexagonal grid of the same scale and system puts it, and a point on an edge goes to the same side
- `Grid.Ring`, `Grid.Spiral` and `Grid.Line`, each with its `Append` and `Seq` forms: the ring at a number of steps from an index, the cells within it nearest first, and the straight line between two indices, in the indices of the grid on every layout. A ring and a spiral keep the cells inside the grid; a line is measured between any two indices, as `Grid.Distance` is, and along a zigzag edge of a hexagonal grid it can pass through an index outside it
- `Field` — what a search from one index found: `Next(index)`, the step back toward the start, `CostTo(index)`, `PathTo(index)`, `Contains(index)`, `Start`, `Len`, `IsEmpty` and `Region()`. It is a dense snapshot, so one field serves every unit heading to the same place
- `Grid.Reachable(start, steps, passable)` — the cells within a number of steps along passable cells, as a `Region`: where a unit can move, around what stands in its way
- `Grid.Border(region)`, `Grid.Grow(region, n)`, `Grid.Shrink(region, n)`, `Grid.IsConnected(region)` and `Grid.Components(region)` — the questions about a region that depend on which cells are adjacent, answered by the grid, for square and hexagonal grids alike
- `Region` — a set of grid indices for area queries, the `container.Set` of `github.com/gravitton/container` over `geom.Point[int]`, immutable and sorted, with the set operations, the relations, the iterators and JSON as an array
- `Reg(indices)` — the region of a slice of indices, in any order and with any repeats; a set of `hex.Hex` is `container.SetOf`
- `Point` gained the lattice methods `hex.Hex` has and `geom.Point` does not, each taking the `Movement` where the answer depends on the neighbors: `Neighbor(direction)`, `AppendNeighbors`, `Ring`, `Spiral` and `Line`, each with its `Append` and `Seq` form, `RangeSeq` and `AppendRange`, `DirectionTo`, `Turn(steps)` and `TurnAround(center, steps)` in quarter turns, `ReflectX` and `ReflectY`, and the package functions `RangeLen(n, movement)` and `RingLen(n, movement)` to size a buffer. A slice form allocates once at its length and a `Seq` form allocates nothing
- `Point.HasLineOfSightFunc` and `Point.FieldOfViewFunc`, with `AppendFieldOfViewFunc` and `FieldOfViewFuncSeq`, take `blocked func(geom.Point[int]) bool`, for blockers that are a lookup in a map or a grid; `AppendFieldOfView` and `FieldOfViewSeq` are the forms of the slice method. All of them decide by the one rule `HasLineOfSight` uses
- The package documentation states the rules the package keeps and lists every panic
- `RectOptions.Origin` and `HexOptions.Origin` — the world position of the top-left corner of the grid, which was fixed at the world origin
- `Cell.AppendRange` and `Cell.RangeSeq`, the `Append` and `Seq` forms of `Cell.Range`, as `Cell.AppendNeighbors` is of `Cell.Neighbors`
- `Grid.MoveTo(origin)` and `Grid.Resize(cellScale)` — place the top-left corner of the bounds of a grid and replace the size of its cells, in place, the values kept
- `Grid.All()` — an iterator over every cell of the grid, row by row, without the cells beyond the edge that `DrawOrder` includes
- `MovementNone`, `Movement.IsNone()`, `Movements()` and `ParseMovement(name)`, and text marshalling, so a `Movement` is stored in JSON and as a map key by the name `String` prints, as the `geom` and `hex` enums are
- `Movement.String()` — human-readable name for each system constant

### Removed
- `Array` and `Arr`, moved to `github.com/gravitton/container` with the changes above; `Grid` stores its cells in a `container.Array[T, int]`. No alias is left (**breaking**)
- `Layout`, with every method of it: a grid could not take a moved or resized one back, and what it measured is on `Grid` and `Cell`. `Grid.MoveTo` and `Grid.Resize` move and resize the grid itself (**breaking**)
- `NewGrid`, `NewLayout`, `NewTransform`, `Transform`, `Layout.With`, `LayoutOption`, `LayoutOpts` and the `SquareFlat`, `SquareIsometric`, `HexFlatTop` and `HexPointyTop` variables. A custom transform could only describe a rotated or sheared square lattice, and such a grid was iterated and bounded as one of the four built in. (**breaking**)
- `Direction` and its `E`, `NE`, `N`, `NW`, `W`, `SW`, `S`, `SE` constants — superseded by `geom.Direction` from `geometry` v1.10.0 (**breaking**)
- `Directions`, `CardinalDirections`, `DiagonalDirections` — superseded by `geom.Directions`, `geom.CardinalDirections`, `geom.DiagonalDirections` and `Movement.Directions` (**breaking**)
- `NeighborOffsets(system)`, `NeighborOffset(system, direction)` and `DistanceTo(from, to, system)` — replaced by the `Movement` methods of the same names (**breaking**)

### Fixed
- `Cell.Bounds` takes its corners from the lattice of the layout as `Cell.Center` does, where it measured half a cell from the center: the boxes of neighboring squares meet in the same numbers and the box around every cell is `Grid.Bounds` to the bit, where they differed in the last digit on most origins and cell sizes
- `Grid.IndexAt` and `Grid.CellAt` rounded a point on an edge away from zero, so a cell at or before index zero lost its top and left edges and the top-left corner of a grid was outside the grid. An edge goes to the same side in every cell: on a square to the cell after it, the top vertex of a diamond to its own, on a hexagon where `hex.FractionalHex.Round` puts it, and the corner of the bounds of a rectangular grid is in the first cell wherever the grid is. Requires the `geometry` and `hexagon` releases after v1.15.0 and v1.5.0
- `Bounds` of a grid of no cells is an empty box at the origin, where it was a box of two cells around it
- A cell size or hexagon radius that is not positive and finite on both axes panics with `grid: cell scale … cannot be laid out`, in the constructors and in `Grid.Resize`, where zero panicked with `geom: division by zero`, a negative size built a mirrored grid with the wrong bounds and NaN one with no bounds. An origin with a NaN or infinite coordinate panics with `grid: origin … is not finite`, in the constructors and `Grid.MoveTo`, where it gave a grid of no bounds
- `Grid.Range`, `AppendRange` and `RangeSeq` cost what a radius to the farthest cell costs for any larger one: a radius of a thousand on a grid of five cells took milliseconds and thirty megabytes for its 25 cells, and `math.MaxInt` never returned
- `Arr`, the constructors of a grid and the JSON of an `Array` refuse a size of more values than an int counts with `grid: array of size <size> is too large`, where the product wrapped and gave an array, or a grid, that contained cells it had no values for
- `Arr` panics with `grid: array of negative size <size>` for a negative width or height. A size with one negative side panicked in the runtime with no word of the array, and one with two built an array holding values that no index reached; the grid constructors, which store through `Arr`, follow
- `Grid.DrawOrder` on an isometric grid rounds each product before adding it, so arm64 and amd64 built for v3 no longer fuse it into a multiply-add and a viewport edge picks the same depths and columns on every architecture
- `Grid.DrawOrder` no longer writes the resolved bounds back into the caller's `IterOptions`, which made a reused options value stick to the first grid's bounds
- `Point.HasLineOfSight` is now reciprocal — it walks from a canonical endpoint, so `a` sees `b` exactly when `b` sees `a`. A Bresenham walk traces a different chain of cells when reversed, which let one point see another that could not see it back (**breaking**)
- `Point.HasLineOfSight` no longer sees through a diagonal wall — a line crossing a cell corner is now opaque when both cells flanking that corner block, so a diagonal run of blockers hides what is behind it. Sight still passes the corner of a single isolated blocker, and `Grid.Range` inherits the fix on rectangular grids (**breaking**)
- `Point.FieldOfView` indexes the blocking points once per call instead of scanning the slice at every step of every line walk, cutting `Grid.Range` on a rectangular grid from ~6.5 ms to ~220 µs at radius 40
- `Grid.cellNeighbors` allocates the neighbor slice at its final capacity instead of growing it, saving one allocation per node expansion in every search algorithm (~14% off `Grid.AStar`)
- `Grid.Range` on a rectangular grid is now the cells within `n` steps of the grid's movement — a Manhattan diamond for cardinal movement, a Chebyshev square for diagonal — instead of a Euclidean disc, which held cells further than `n` steps from `n = 3` on and left out the corners of a diagonal grid (**breaking**)
- `Grid.Range` no longer returns the cells that fail `valid`, nor indices outside the grid; both still hide what lies behind them (**breaking**)
- `BreadthFirstSearchField` and `UniformCostSearchField` no longer hold the cells that fail `valid`, and no search asks `cost` about such a cell (**breaking**)
- A nil `valid` or `cost` is accepted by every search, both fields and `Grid.Range`, admitting every cell at a cost of one per step, as `Grid.Path` already did; they panicked before
- `Array.Get` and `Array.Set` panic on an index outside either axis, as documented. An `x` past the width used to read or write a cell of the next row. The panic is an error naming the index and the size, formatted only when printed, so both methods inline and run in about half the time
- `Array.Clone` of the zero value is the zero value, where it gave empty storage that marshalled its values as `[]` instead of `null`


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
