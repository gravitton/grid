# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/grid/compare/v1.2.0...main)
### Changed
- The searches share the prefix of what they return: `AStar`, `UniformCostSearch`, `BreadthFirstSearch` and `GreedyBestFirstSearch` are `PathAStar`, `PathUniformCost`, `PathBreadthFirst` and `PathGreedyBestFirst`, and `UniformCostSearchField` and `BreadthFirstSearchField` are `FieldUniformCost` and `FieldBreadthFirst` (**breaking**)
- Every sight method of `Square` takes `blocked func(Index) bool`, as those of `hexagon` do: `HasLineOfSightFunc` is `HasLineOfSight` and the form taking a slice of blockers is removed (**breaking**)
- `Square.Line` and `LineSeq` are `LineTo` and `LineToSeq`, as `hexagon` names them (**breaking**)
- `Point.FieldOfView` and its `Func`, `Append` and `Seq` forms over given candidates are removed, as they are from `hexagon`: filter them by `HasLineOfSight`. `FieldOfView` is now the cells in sight within a radius (**breaking**)
- `Point` is `Square` and `Pt` is `Sq`, which reads as the counterpart of `hex.Hex`; `Square.Point()` still gives the `geom.Point` (**breaking**)
- Every panic of the package is an error, where it was a string, with the same message: the methods with no answer for a movement outside the constants, `MoveTo`, `Rescale` and the `HexLayout` constructors (**breaking** for a caller that recovers a string)
- Require Go 1.27, `geometry` v1.18.0, `hexagon` v1.7.0 and `github.com/gravitton/container` v1.0.0, the storage of a grid and the set under `Region`, and `assert` v1.7.0 for the tests
- `System` is renamed `Movement`, the name pathfinding libraries give the choice between four and eight neighbors, and its constants `Cardinal` and `Diagonal` are `MovementCardinal` and `MovementDiagonal`; `String` still prints `Cardinal` and `Diagonal` (**breaking**)
- `Movement` gained the direction methods that were package-level functions: `Movement.Directions()`, `Movement.Offsets()`, `Movement.Offset(direction)`, `Movement.Contains(direction)` and `Movement.Distance(from, to)`, matching the method style `geometry` v1.10.0 and `hexagon` v1.3.0 moved to (**breaking**)
- Neighbors are now ordered by increasing angle — `MovementCardinal` yields right, down, left, up (was east, north, west, south) — following the `Direction` reorder in `geometry` v1.11.0. The neighbor set is unchanged; only the visit order, and therefore pathfinding tie-breaks, differ (**breaking**)
- `Movement.Offsets` returns a fresh slice, allocated once, instead of one aliasing the package's own table, so callers can no longer mutate it (**breaking**)
- `Movement.Offset` returns the zero vector for a direction that is not one of the movement instead of panicking (**breaking**)
- A movement outside the constants answers as one with no neighbors where it panicked with `unsupported system`: `Movement.Directions` and `Offsets` return nil, `Contains` false and `Offset` the zero vector. `Movement.Distance` has no answer and panics with `grid: unknown movement <n>`, and `Movement.String` prints `None` for such a value instead of `Unknown` (**breaking**)
- The constructors no longer repeat the package name: `NewRectGrid`, `NewIsometricRectGrid`, `NewHexagonFlatTopGrid` and `NewHexagonPointyTopGrid` are `NewRect`, `NewIsometric`, `NewHexFlatTop` and `NewHexPointyTop`; the option types `RectGridOption` and `HexGridOption` are removed (**breaking**)
- Options are a callback that sets the fields of an options struct, as the `ui` package of the engine takes them: `NewRect` and `NewIsometric` take `func(*SquareOptions)`, with `Movement` and `Origin`, and `NewHexFlatTop` and `NewHexPointyTop` take `func(*HexOptions)`, with `EvenOffset` and `Origin`. `RectGridOpts.DiagonalMovement()` is `options.Movement = MovementDiagonal` and `HexGridOpts.EvenSystem()` is `options.EvenOffset = true`; the `RectGridOpts` and `HexGridOpts` variables are removed (**breaking**)
- `Grid.Set(index, value)` is removed: the value of a cell is read and written through the cell alone, `g.Cell(index).Set(value)` beside `g.Cell(index).Value()`, as its size, center, bounds and polygon already were. `Grid` keeps what relates cells or covers them all (**breaking**)
- `Grid.Get(index)` is `Grid.Cell(index)`, `Grid.At(point)` is `Grid.CellAt(point)` and `Cell.Get()` is `Cell.Value()`, so `Get` no longer names three different things (**breaking**)
- `Cell.Valid` is `Cell.IsContained`, the `Is` prefix every state check without an argument takes, as `IsEmpty` and `IsNone` do (**breaking**)
- `Grid.Has` and `Movement.Has` are `Contains`, as `Region.Contains` and the `geom` shapes say it (**breaking**)
- `Grid.Iter(options)` is `Grid.DrawOrder(bounds)`, taking the box itself, a zero one for the whole grid; `IterOptions` is removed (**breaking**)
- The callback types `ValidFunc`, `ValidIndexFunc` and `CostFunc` are removed; every method spells the function type in its signature, and the parameter that was `valid` is `passable` (**breaking**)
- Signatures spell the `geom` types in full, `geom.Size[float64]`, and an index as `Index`, instead of the aliases of `geometry/types/ints` and `floats`; the types are the same, so no caller changes
- `Point.Range(n)` is `Square.Range(n, movement)` and returns the indices within `n` steps of the movement, sorted by x and then by y and allocated once, instead of a Euclidean disc. It is what a rectangular `Grid.Range` is built on, as the hexagonal one is built on `hex.Hex.Range` (**breaking**)
- `Grid.Bounds` and `Cell.Bounds` return a `geom.Box[float64]`, and `Grid.DrawOrder` takes one, where they used `geom.Rectangle[float64]`: a bounds is the axis-aligned extent `geom` v1.15.0 gives every shape, with `Min` and `Max` as fields; `box.Rectangle()` converts back (**breaking**)
- `Grid.CellBounds` is `CellSize`: it returns a size, the width and height of a cell as drawn, as `Cell.Size`, `CellWidth` and `CellHeight` already say, while every `Bounds` is now a box (**breaking**)
- A `Cell` is passed by value: `Grid.Cell`, `Grid.CellAt`, `Grid.All` and `Grid.DrawOrder` give a `Cell[T]` and the `passable` and `cost` callbacks take one, where they used `*Cell[T]`, which was allocated for every call. `Cell.Value` still returns the pointer into the grid (**breaking**)
- `Grid.Path` and `Cell.PathTo` return the indices of the path, `[]geom.Point[int]`, as every search by name does, instead of `[]*Cell[T]`; `Grid.Cell(index)` gives the cell of one (**breaking**)
- `FieldBreadthFirst` and `FieldUniformCost` return a `Field` instead of a `map[geom.Point[int]]geom.Point[int]`: `field.Next(index)` is the step the map held (**breaking**)
- `Grid.Range(index, radius, passable)` is split in two, since one callback could not say that water is seen across and a tower hides what stands on it. `Grid.Range(index, radius)` is the plain range `Square.Range` and `hex.Hex.Range` are, a `Region` of the cells within the radius, with `RangeSeq`. Sight is `Grid.FieldOfView(index, radius, blocked)`, as `hex.Hex.FieldOfView` names it, with `FieldOfViewSeq`: a blocked cell is in sight and hides what lies behind it, the cell of the index is never asked about, and it asks about each cell at most once. `Cell.Range` follows, and every radius past the farthest cell costs what that one does (**breaking**)
- Range, `Neighbors` and `Cell.Neighbors` are empty from an index outside the grid, where they held the cells of the grid beside it (**breaking**)
- `Grid.DrawOrder` yields only cells of the grid, so a caller no longer filters on `Cell.IsContained`, and panics once ranged for bounds with a NaN or infinite coordinate on every layout (**breaking**)
- A search holds no per-node allocation and a grid no function values: `PathAStar` corner to corner on a 64×64 rectangular grid went from 20 651 allocations to 14, and the mapping between an index and the world is about twice as fast
- The package no longer depends on `github.com/gravitton/x`: its heap and its queue are replaced by a frontier and a queue of the search's own, and the set is `container.Set`
- The origin of a grid is the top-left corner of its bounds on every kind of grid, where it was the center of cell (0, 0): `SquareOptions.Origin`, `HexOptions.Origin` and `Grid.MoveTo` place `Bounds().Min`, and `Grid.Rescale` keeps it. An isometric grid moved right by the reach of its rows and a hexagonal one with `EvenOffset` half a cell down or right, so their bounds start at the origin as those of the others did (**breaking**)
- A grid maps through the lattice of its kind, every center and edge a whole number of steps from the origin, where it multiplied by a matrix each way: a cell edge that is a whole number is found exactly, where about one in seventy fell in the cell before it on whole cell sizes
- The cell-size helpers are named for the cell scale they return and for their constructor: `RectCellSize`, `IsometricRectCellSize`, `IsometricPixelPerfectRectCellSize`, `HexFlatTopCellSize`, `HexPointyTopCellSize`, `HexFlatTopIsometricPixelPerfectCellSize` and `HexPointyTopIsometricPixelPerfectCellSize` are `RectCellScale`, `IsometricCellScale`, `IsometricPixelCellScale`, `HexFlatTopCellScale`, `HexPointyTopCellScale`, `HexFlatTopPixelCellScale` and `HexPointyTopPixelCellScale`. A hexagonal one gives a circumradius, which "cell size", the box of a cell as drawn, no longer names (**breaking**)
- The constructors of a grid and of a `HexLayout` name what they take `cellScale`, where they were `grid` and `size`, `cellSize` or `hexSize`; no caller changes
- Rewrote the doc comments to one style, with the shared rules, every panic and the order of directions stated once in the package doc
- CI runs the shared workflow of `gravitton/.github`, adding the fused multiply-add check, and can be started by hand

### Added
- `Cell.String` prints a cell as `Cell(x,y)`
- Examples for `Movement`, `Square`, `Field`, `Grid.Cell`, `Grid.IndexAt`, `Grid.Range`, `Grid.Reachable`, `Grid.Path`, `Grid.Border` and `Grid.DrawOrder`; benchmarks for every search, unweighted and weighted, `Grid.FieldUniformCost`, `Field.CostTo`, `Field.PathTo`, `Grid.All`, `Grid.Neighbors`, `Grid.Range`, `Grid.Reachable`, the region methods of `Grid`, `Grid.DrawOrder` and the traversals and sight of `Square`; and the fuzz targets `FuzzGrid_Neighbors` and `FuzzGrid_Range`, which check a range against the distances and sight both ways on every kind of grid, `FuzzGrid_PathAStar`, which checks a path against breadth-first and uniform cost search, and `FuzzPoint_HasLineOfSight`, which checks sight both ways and against its `Func` form
- `Grid.Neighbors(index)` and `Grid.NeighborsSeq(index)` — the indices adjacent to an index that lie inside the grid, which only `Cell.Neighbors` gave before. They are exactly the cells at distance one, in the order of the neighbor rule
- `HexLayout`, built by `NewHexFlatTopLayout(cellScale, system)` and `NewHexPointyTopLayout(cellScale, system)`: hexagonal cells placed in the world with no grid around them, on an unbounded plane, indexed in any `hex.CoordinateSystem`, with `Center`, `Bounds`, `Polygon`, `IndexAt`, `CellSize`, `CellSpacing`, `MoveTo` and `Rescale`. A cell lands exactly where a hexagonal grid of the same scale and system puts it, and a point on an edge goes to the same side
- `Grid.Ring`, `Grid.Spiral` and `Grid.Line`, each with its `Seq` form, which walks no further than where it is stopped and allocates nothing: the ring at a number of steps from an index, the cells within it nearest first, and the straight line between two indices, in the indices of the grid on every layout. A ring and a spiral keep the cells inside the grid; a line is measured between any two indices, as `Grid.Distance` is, and along a zigzag edge of a hexagonal grid it can pass through an index outside it
- `Field` — what a search from one index found: `Next(index)`, the step back toward the start, `CostTo(index)`, `PathTo(index)` and `AppendPathTo(dst, index)`, `Contains(index)`, `Start`, `Len`, `IsEmpty` and `Region()`. A field searched from a destination serves every unit heading there, where a step costs the same both ways. It keeps the number of steps to every index, so `CostTo` answers at once and `PathTo` walks the path once
- `Grid.Reachable(start, budget, passable, cost)` and `Cell.Reachable` — the `Field` of the cells reached for at most a budget, walking no further, with the path and cost to each; a nil cost makes the budget a number of steps
- `Grid.Border(region)`, `Grid.Grow(region, n)`, `Grid.Shrink(region, n)`, `Grid.IsConnected(region)` and `Grid.Components(region)` — the questions about a region that depend on which cells are adjacent, answered by the grid, for square and hexagonal grids alike
- `Region` — a set of grid indices for area queries, the `container.Set` of `github.com/gravitton/container` over `geom.Point[int]`, immutable and sorted, with the set operations, the relations, the iterators and JSON as an array
- `Reg(indices)` — the region of a slice of indices, in any order and with any repeats; a set of `hex.Hex` is `container.SetOf`
- `Square` gained the lattice methods `hex.Hex` has and `geom.Point` does not, each taking the `Movement` where the answer depends on the neighbors: `Neighbor(direction)`, `NeighborsSeq`, `Ring`, `Spiral` and `LineTo`, each with its `Seq` form, `RangeSeq`, `DirectionTo`, `Turn(steps)` and `TurnAround(center, steps)` in quarter turns, `ReflectX` and `ReflectY`, and the package functions `RangeLen(n, movement)` and `RingLen(n, movement)` to size a buffer. A slice form allocates once at its length and a `Seq` form allocates nothing, so `slices.AppendSeq` is the buffered form, as in `hexagon`
- The package documentation states the rules the package keeps and lists every panic
- `SquareOptions.Origin` and `HexOptions.Origin` — the world position of the top-left corner of the grid, which was fixed at the world origin
- `Grid.MoveTo(origin)` and `Grid.Rescale(cellScale)` — place the top-left corner of the bounds of a grid and replace its cell scale, in place, the values kept
- `Grid.CellScale()` and `Grid.Origin()`, and `HexLayout.CellScale()` and `HexLayout.Origin()`, to read back what a constructor, `Rescale` and `MoveTo` took
- `Grid.Layer[U]()` — a grid of another type over the same cells, of the size, the layout and the neighbor rule of the grid
- `Grid.MarshalJSON` and `Grid.UnmarshalJSON` — a grid written with its layout and values and read back into a zero `Grid`, an error for what no constructor builds, where `json.Marshal` wrote `{}` and lost every value
- `Index`, an alias of `geom.Point[int]` for the index of a cell, which every signature now speaks
- `Grid.Select(selected)` — the `Region` of the cells a callback picks
- `Grid.HasLineOfSight(from, to, blocked)`, and `Cell.HasLineOfSight` and `Cell.FieldOfView`: sight on every layout. They agree with `Grid.FieldOfView` cell for cell
- `Cell.Ring`, `Cell.Spiral` and `Cell.LineTo`
- `Grid.Field(start, passable, cost)` and `Cell.Field` — the field of every cell reached from start, counting steps for a nil cost; with `Path` and `Reachable` it is one of the defaults, and the searches named for their algorithm are there to ask for one
- `Field.All()` — the indices the search reached
- `Square.Sector(radius, direction, turns, movement)`, with `SectorSeq`, and `SectorLen(radius, turns, movement)` — the cone of indices within a radius and a number of steps between directions about a direction, ring by ring, the counterpart of `hex.Hex.Sector`
- `Square.String` prints a square as `(x,y)`, as `hex.Hex` does
- `Square.FieldOfView(radius, movement, blocked)`, with `FieldOfViewSeq`, the counterpart of `hex.Hex.FieldOfView`
- `Grid.All()` — an iterator over every cell of the grid, row by row
- `MovementNone`, `Movement.IsNone()`, `Movements()` and `ParseMovement(name)`, and text marshalling, so a `Movement` is stored in JSON and as a map key by the name `String` prints, as the `geom` and `hex` enums are
- `Movement.String()` — the name of each movement

### Removed
- `Array` and `Arr`, moved to `github.com/gravitton/container`, where `Array` is generic over the integer type of its sides as `Array[T, S]`; `Grid` stores its cells in a `container.Array[T, int]`. No alias is left (**breaking**)
- `Layout`, with every method of it: a grid could not take a moved or resized one back, and what it measured is on `Grid` and `Cell`. `Grid.MoveTo` and `Grid.Resize` move and resize the grid itself (**breaking**)
- `NewGrid`, `NewLayout`, `NewTransform`, `Transform`, `Layout.With`, `LayoutOption`, `LayoutOpts` and the `SquareFlat`, `SquareIsometric`, `HexFlatTop` and `HexPointyTop` variables. A custom transform could only describe a rotated or sheared square lattice, and such a grid was iterated and bounded as one of the four built in. (**breaking**)
- `Direction` and its `E`, `NE`, `N`, `NW`, `W`, `SW`, `S`, `SE` constants — superseded by `geom.Direction` from `geometry` v1.10.0 (**breaking**)
- `Directions`, `CardinalDirections`, `DiagonalDirections` — superseded by `geom.Directions`, `geom.CardinalDirections`, `geom.DiagonalDirections` and `Movement.Directions` (**breaking**)
- `NeighborOffsets(system)`, `NeighborOffset(system, direction)` and `DistanceTo(from, to, system)` — replaced by `Movement.Offsets`, `Movement.Offset` and `Movement.Distance` (**breaking**)

### Fixed
- A negative or NaN cost made every weighted search run forever, each step improving on what it had found. It panics with `grid: cost … is negative or NaN`, and a cost of +Inf is a step never taken, where the path through it was returned
- `Grid.DrawOrder` on a flat-top grid with `EvenOffset` draws back to front, where it drew the higher columns of a row last, and a viewport that starts late in a cell reaches the last cell it touches
- `Cell.Bounds` takes its corners from the lattice of the layout as `Cell.Center` does, where it measured half a cell from the center: the boxes of neighboring squares meet in the same numbers and the box around every cell is `Grid.Bounds` to the bit, where they differed in the last digit on most origins and cell sizes
- `Grid.IndexAt` and `Grid.CellAt` rounded a point on an edge away from zero, so a cell at or before index zero lost its top and left edges and the top-left corner of a grid was outside the grid. An edge goes to the same side in every cell: on a square to the cell after it, the top vertex of a diamond to its own, on a hexagon where `hex.FractionalHex.Round` puts it, and the corner of the bounds of a rectangular grid is in the first cell wherever the grid is. Requires the `geometry` and `hexagon` releases after v1.15.0 and v1.5.0
- `Bounds` of a grid of no cells is an empty box at the origin, where it was a box of two cells around it
- A cell scale that is not positive and finite on both axes panics with `grid: cell scale … is not positive and finite`, in the constructors and in `Grid.Rescale`, where zero panicked with `geom: division by zero`, a negative size built a mirrored grid with the wrong bounds and NaN one with no bounds. An origin with a NaN or infinite coordinate panics with `grid: origin … is not finite`, in the constructors and `Grid.MoveTo`, where it gave a grid of no bounds
- The constructors of a grid panic for a negative size or one of more cells than an int counts, through `container.Arr`. A size with one negative side panicked in the runtime with no word of the grid, one with two built a grid holding values that no index reached, and a product that wrapped gave a grid that contained cells it had no values for
- `Grid.DrawOrder` on an isometric grid rounds each product before adding it, so arm64 and amd64 built for v3 no longer fuse it into a multiply-add and a viewport edge picks the same depths and columns on every architecture
- `Square.HasLineOfSight` is now reciprocal — it walks from a canonical endpoint, so `a` sees `b` exactly when `b` sees `a`. A Bresenham walk traces a different chain of cells when reversed, which let one point see another that could not see it back (**breaking**)
- `Square.HasLineOfSight` no longer sees through a diagonal wall — a line crossing a cell corner is now opaque when both cells flanking that corner block, so a diagonal run of blockers hides what is behind it. Sight still passes the corner of a single isolated blocker, and `Grid.FieldOfView` has the rule on rectangular grids (**breaking**)
- `Grid.Range` on a rectangular grid is now the cells within `n` steps of the grid's movement — a Manhattan diamond for cardinal movement, a Chebyshev square for diagonal — instead of a Euclidean disc, which held cells further than `n` steps from `n = 3` on and left out the corners of a diagonal grid (**breaking**)
- `Grid.Range` no longer returns the cells that fail `passable`, nor indices outside the grid; both still hide what lies behind them (**breaking**)
- `FieldBreadthFirst` and `FieldUniformCost` no longer hold the cells that fail `passable`, and no search asks `cost` about such a cell (**breaking**)
- A nil `passable` or `cost` is accepted by every search, both fields and `Grid.Range`, admitting every cell at a cost of one per step, as `Grid.Path` already did; they panicked before


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
