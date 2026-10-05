// Package grid is a generic 2D grid for games: a [Grid] stores a value per cell, maps cells to
// the world for rectangular, isometric and hexagonal layouts, and answers range, sight and
// pathfinding queries. It builds on github.com/gravitton/geometry for the shapes and points it
// speaks and on github.com/gravitton/hexagon for the hexagonal coordinate math.
//
// # Types
//
// [Grid] is built by one of four constructors, [NewRect], [NewIsometric], [NewHexFlatTop] and
// [NewHexPointyTop], and is the same type whatever the layout. A [Cell] is a handle on one index
// of a grid, passed by value. [Layout] is the mapping between indices and the world, [Array] the
// storage under a grid, [Region] an immutable set of indices, [Point] an index of a square grid
// with its lattice queries, and [Movement] the choice between four and eight neighbors.
//
// # Indices
//
// An index is a geom.Point[int] of column and row on every layout. On a hexagonal grid it is
// the offset coordinate; axial coordinates never appear in the API.
//
// # Rules
//
// Outside the grid answers, it does not panic: [Grid.Contains] is false, a [Cell] has a nil
// value, [Grid.Set] does nothing, and a search from or to such an index finds nothing.
//
// A nil callback means no constraint. A nil passable admits every cell and a nil cost charges
// one per step, on every method that takes them.
//
// Range, distance and neighbors agree: [Grid.Range] holds only cells within that many steps of
// [Grid.Distance], and a neighbor is a cell at distance one.
//
// Sight is reciprocal: two cells see each other or neither does, and neither end blocks. A cell
// that is not passable is never in a range or a search field, and hides what lies behind it.
//
// A path holds both its ends and steps between neighbors; a search that finds none returns nil.
//
// Every Bounds is a geom.Box, the axis-aligned extent, and the size of a cell is its size as
// drawn.
//
// # Panics
//
// A degenerate value gets an answer; an argument with no meaning gets a panic:
//
//   - [Array.Get] and [Array.Set] for an index outside the array, as a slice does.
//   - Every method of [Movement] but IsNone and String, every method of [Point] that takes a
//     movement, and the constructors of a square grid, for a movement outside the constants.
//   - [Grid.IndexAt], [Grid.CellAt] and [Layout.FromPoint] for a point with a NaN or infinite
//     coordinate, through geom.Cast.
package grid
