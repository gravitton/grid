// Package grid provides generic 2D grids for games: a [Grid] stores a value per cell, places
// its cells in the world for rectangular, isometric and hexagonal layouts, and answers range,
// sight, area and pathfinding queries. It builds on github.com/gravitton/geometry for the
// shapes and points it speaks, on github.com/gravitton/hexagon for the hexagonal coordinate
// math, and on github.com/gravitton/container for the storage of a grid and the set under a
// [Region].
//
// # Types
//
//   - [Grid] is built by [NewRect], [NewIsometric], [NewHexFlatTop] or [NewHexPointyTop] and is
//     the same type on every layout, and written to and read back from JSON. A [Cell] is a
//     handle on one index of a grid, passed by value, and [Grid.Layer] a grid of another type
//     over the same cells.
//   - [HexLayout] places hexagonal cells on an unbounded plane with no grid around them, for a
//     map whose data lives elsewhere.
//   - [Region] is the immutable sorted set of indices every area query returns. A search returns
//     a [Field] or a path: [Grid.Path], [Grid.Field] and [Grid.Reachable] ask for the answer, and
//     the searches named for their algorithm choose how it is found.
//   - [Square] is an index of a square grid with its lattice queries, and [Movement] the choice
//     between its four and eight neighbors.
//
// # Indices
//
// An [Index] is a geom.Point[int] of column and row on every layout, the offset coordinate on a
// hexagonal grid, and on a [HexLayout] the coordinate in the system it is built with. A hex.Hex
// never appears in the API. Directions, neighbors and rings are ordered by increasing angle:
// clockwise as drawn, with Y pointing down in the world, and negative turns go the other way.
//
// # Placement
//
// The origin of a grid is the top-left corner of its bounds on every layout. A point on an
// edge goes to the same side in every cell: on a square to the cell after it, to the right or
// below, so the corner of the bounds is in the first cell; on a diamond a vertex to the cell it
// is the top of; on a hexagon where [hex.FractionalHex.Round] puts it. The boxes of neighbors
// meet in the same numbers, and together they are the bounds of the grid.
//
// # Rules
//
// Outside the grid answers, it does not panic: [Grid.Contains] is false, a [Cell] has a nil
// value and stores nothing, Neighbors, Range, Ring and Spiral from it are empty, and a search
// from or to it finds nothing, and it sees nothing. [Grid.Distance] and [Grid.Line] alone measure
// between any two indices.
//
// A nil callback means no constraint: a nil passable admits every cell, a nil cost charges one
// per step, a nil blocked blocks nothing, and a nil selected selects every cell.
//
// Range, distance and neighbors agree: [Grid.Range] is exactly the cells within radius by
// [Grid.Distance], and a neighbor is a cell at distance one.
//
// Sight is reciprocal: two cells see each other or neither does, and neither end blocks. What
// blocks sight is apart from what can be walked on: a cell blocked hides what lies behind it and
// is itself in sight, and what lies outside the grid blocks nothing. A cell that is not passable
// is never in a search field.
//
// A path holds both its ends and steps between neighbors; a search that finds none returns nil.
// A cost is what a step adds to a path: +Inf is a step never taken, and a negative one or a NaN
// has no meaning and panics.
//
// A [Region] never changes: every operation returns a region, and one that would equal a side
// is that side.
//
// Every Bounds is a geom.Box, the axis-aligned extent. The cell size is the box of a cell as
// drawn; the cell scale a constructor takes is the size of a square or a diamond and the
// circumradius per axis of a hexagon.
//
// # Allocations
//
// A slice or region result is allocated once, at its final size, but for those that make room
// for the most they can hold: [Grid.Neighbors] and [Grid.Ring] for every index the neighbor rule
// gives, [Grid.Select] for every cell of the grid, and [Grid.FieldOfView] and [Square.FieldOfView]
// for every cell of the range where something is hidden. A traversal has a Seq form that
// allocates nothing and walks no further than where it is stopped, and slices.AppendSeq puts one
// in a buffer. A search allocates the field it returns, for the part of the grid it reaches, and
// reuses the rest of what it needs from one search to the next, safe for searches from several
// goroutines at once; its path has the Append form of [Field.PathTo]. What
// depends on which cells of a region are adjacent builds its result in steps: [Grid.Grow]
// marks the cells of the grid and gathers what it adds as it goes, [Grid.Shrink] builds a
// region at every step, and [Grid.IsConnected] and [Grid.Components] mark the indices of the
// region.
//
// # Supported range
//
// A cell is placed to the precision a float64 has where the grid is, and the package is tested
// for grids within a million units of the world origin, with cells from a hundredth of a unit
// to a thousand: there the center of a cell maps back to its index and an edge that is a
// number a float holds goes to the same side in every cell. Further out a step from one cell
// to the next is rounded more and more coarsely, and a point near an edge can land in the
// cell beside it; no method tests for that. A point too far for an index to count gives a
// platform-dependent one.
//
// # Panics
//
// A degenerate value gets an answer; an argument with no meaning gets a panic, with an error
// that errors.Is matches to its sentinel:
//
//   - [ErrUnknownMovement] for a movement outside the constants to [Movement.Distance],
//     [Square.DistanceTo], [Square.LineTo] with its Seq form, and [NewRect] and [NewIsometric]: no
//     index is reached from another under it. Every other method answers as for no neighbors.
//   - [geom.ErrNonFinite] for a point with a NaN or infinite coordinate to [Grid.IndexAt],
//     [Grid.CellAt] and [HexLayout.IndexAt]: no cell is at it.
//   - [ErrNonFinite] for an origin with a NaN or infinite coordinate to the constructors of a
//     grid, [Grid.MoveTo] and [HexLayout.MoveTo], and for such bounds to the iterator of
//     [Grid.DrawOrder]: no cell is drawn from it.
//   - [ErrInvalidCellScale] for a cell scale that is not positive and finite on both axes to every
//     constructor, [Grid.Rescale] and [HexLayout.Rescale]: no cell can be drawn at it.
//   - [ErrInvalidCost] for a cost that is negative or NaN to [Grid.Path], [Grid.Field],
//     [Grid.Reachable], [Grid.PathUniformCost], [Grid.FieldUniformCost], [Grid.PathAStar],
//     [Cell.PathTo], [Cell.Field] and [Cell.Reachable]: no search ends on it.
//   - [hex.ErrUnknownCoordinateSystem] for a coordinate system outside the seven to
//     [NewHexFlatTopLayout] and [NewHexPointyTopLayout]: no index can be read in it.
//   - The error of container.Arr for a negative size, or one of more cells than an int counts, to
//     the constructors of a grid: no storage holds it.
package grid
