package grid

import (
	"iter"

	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

// Grid holds a value of type T for every cell of a rectangular, isometric or hexagonal layout,
// places its cells in the world, and answers the neighbor, range, area and path queries between
// them. It is built by [NewRect], [NewIsometric], [NewHexFlatTop] or [NewHexPointyTop]; the
// zero value is no grid, and its methods are not to be called.
type Grid[T any] struct {
	cells    container.Array[T, int]
	layout   layout
	topology topology
}

// newGrid returns a grid of the given size in cells, placed by the layout, with the neighbor
// rule of its kind of cell.
func newGrid[T any](size geom.Size[int], layout layout, topology topology) *Grid[T] {
	return &Grid[T]{
		cells:    container.Arr[T](size),
		layout:   layout,
		topology: topology,
	}
}

// Size returns the number of columns and rows of the grid.
func (g *Grid[T]) Size() geom.Size[int] {
	return g.cells.Size()
}

// Width returns the number of columns of the grid.
func (g *Grid[T]) Width() int {
	return g.cells.Width()
}

// Height returns the number of rows of the grid.
func (g *Grid[T]) Height() int {
	return g.cells.Height()
}

// Cell returns the cell at the index, a handle on its value and its place in the world.
func (g *Grid[T]) Cell(index geom.Point[int]) Cell[T] {
	return Cell[T]{g, index}
}

// Fill sets every cell to value.
func (g *Grid[T]) Fill(value T) {
	g.cells.Fill(value)
}

// Clear resets every cell to the zero value.
func (g *Grid[T]) Clear() {
	g.cells.Clear()
}

// Clone returns a copy of the grid with storage of its own, in the same place, its values
// copied as an assignment copies them.
func (g *Grid[T]) Clone() *Grid[T] {
	clone := *g
	clone.cells = g.cells.Clone()

	return &clone
}

// All returns an iterator over every cell of the grid, row by row.
func (g *Grid[T]) All() iter.Seq[Cell[T]] {
	return func(yield func(Cell[T]) bool) {
		for index := range g.cells.Keys() {
			if !yield(g.Cell(index)) {
				return
			}
		}
	}
}

// Neighbors returns the indices adjacent to the index that lie in the grid, by increasing
// angle.
func (g *Grid[T]) Neighbors(index geom.Point[int]) []geom.Point[int] {
	if !g.Contains(index) {
		return nil
	}

	return g.appendNeighbors(make([]geom.Point[int], 0, len(g.topology.offsets(index))), index)
}

// AppendNeighbors appends the indices of Neighbors to dst and returns the extended slice.
func (g *Grid[T]) AppendNeighbors(dst []geom.Point[int], index geom.Point[int]) []geom.Point[int] {
	if !g.Contains(index) {
		return dst
	}

	return g.appendNeighbors(dst, index)
}

// appendNeighbors appends the neighbors of an index the caller knows to be inside the grid.
func (g *Grid[T]) appendNeighbors(dst []geom.Point[int], index geom.Point[int]) []geom.Point[int] {
	for _, offset := range g.topology.offsets(index) {
		next := index.Add(offset)
		if g.Contains(next) {
			dst = append(dst, next)
		}
	}

	return dst
}

// Range returns the cells within radius of the index that are passable and in sight of it,
// asking passable about each cell at most once. It is empty for a negative radius and for an
// index that is not passable itself: nothing is seen from where nothing can stand.
func (g *Grid[T]) Range(index geom.Point[int], radius int, passable func(Cell[T]) bool) Region {
	if radius < 0 || !g.passes(index, passable) {
		return Region{}
	}

	radius = g.clampRadius(index, radius)

	return container.SetFrom(g.appendRange(make([]geom.Point[int], 0, g.topology.rangeLen(radius)), index, radius, passable))
}

// AppendRange appends the indices of Range to dst, in no order to rely on, and returns the
// extended slice.
func (g *Grid[T]) AppendRange(dst []geom.Point[int], index geom.Point[int], radius int, passable func(Cell[T]) bool) []geom.Point[int] {
	if radius < 0 || !g.passes(index, passable) {
		return dst
	}

	return g.appendRange(dst, index, g.clampRadius(index, radius), passable)
}

// appendRange appends the indices in range of an index the caller knows to pass. Only where a
// cell was refused can anything be hidden, so only then does the topology look for sight.
func (g *Grid[T]) appendRange(dst []geom.Point[int], index geom.Point[int], radius int, passable func(Cell[T]) bool) []geom.Point[int] {
	start := len(dst)
	dst = g.topology.appendWithin(dst, index, radius)

	passed := dst[:start]
	marks := sightMarks{size: g.Size(), corner: geom.Pt(index.X-radius-1, index.Y-radius-1), side: 2*radius + 3}
	for _, candidate := range dst[start:] {
		switch {
		case !g.Contains(candidate):
		case candidate == index || passable == nil || passable(g.Cell(candidate)):
			passed = append(passed, candidate)
		default:
			marks.refuse(candidate)
		}
	}

	if marks.isEmpty() {
		return passed
	}

	return g.appendSeen(passed[:start], index, radius, passed[start:], marks)
}

// appendSeen appends the candidates in sight of center past the cells refused and the edge of
// the grid. The candidates may lie in dst past its length: each is read before its place in
// dst is written.
func (g *Grid[T]) appendSeen(dst []geom.Point[int], center geom.Point[int], radius int, candidates []geom.Point[int], marks sightMarks) []geom.Point[int] {
	g.topology.sight(center, radius, candidates, &marks)

	for _, candidate := range candidates {
		if marks.seen(candidate) {
			dst = append(dst, candidate)
		}
	}

	return dst
}

// RangeSeq returns an iterator over the indices of AppendRange.
func (g *Grid[T]) RangeSeq(index geom.Point[int], radius int, passable func(Cell[T]) bool) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		g.walkRange(index, radius, passable, yield)
	}
}

// walkRange yields the indices of Range and reports whether it yielded them all. It checks the
// index itself, so RangeSeq only builds its closure and stays small enough to inline.
func (g *Grid[T]) walkRange(index geom.Point[int], radius int, passable func(Cell[T]) bool, yield func(geom.Point[int]) bool) bool {
	if radius < 0 || !g.passes(index, passable) {
		return true
	}

	radius = g.clampRadius(index, radius)

	for _, found := range g.appendRange(make([]geom.Point[int], 0, g.topology.rangeLen(radius)), index, radius, passable) {
		if !yield(found) {
			return false
		}
	}

	return true
}

// clampRadius returns the radius cut to the distance from the index to the farthest cell of
// the grid, one of its corners, since a range gathers every index within its radius before it
// asks which lie in the grid. A radius short of the nearest edge needs no cut.
func (g *Grid[T]) clampRadius(index geom.Point[int], radius int) int {
	x, y := index.XY()
	right, bottom := g.Width()-1, g.Height()-1
	if radius <= min(x, y, right-x, bottom-y) {
		return radius
	}

	return min(radius, max(
		g.topology.distance(index, geom.Pt(0, 0)),
		g.topology.distance(index, geom.Pt(right, 0)),
		g.topology.distance(index, geom.Pt(0, bottom)),
		g.topology.distance(index, geom.Pt(right, bottom)),
	))
}

// passes reports whether the index is inside the grid and its cell passable.
func (g *Grid[T]) passes(index geom.Point[int], passable func(Cell[T]) bool) bool {
	return g.Contains(index) && (passable == nil || passable(g.Cell(index)))
}

// Ring returns the indices at exactly radius steps from the index that lie in the grid, by
// increasing angle around it, and the index alone for a zero radius. It returns nil for a
// negative radius and for one past every cell of the grid.
func (g *Grid[T]) Ring(index geom.Point[int], radius int) []geom.Point[int] {
	if !g.hasRing(index, radius) {
		return nil
	}

	return g.appendRing(make([]geom.Point[int], 0, g.topology.ringLen(radius)), index, radius)
}

// AppendRing appends the indices of Ring to dst and returns the extended slice.
func (g *Grid[T]) AppendRing(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	if !g.hasRing(index, radius) {
		return dst
	}

	return g.appendRing(dst, index, radius)
}

// RingSeq returns an iterator over the indices of Ring, in the same order.
func (g *Grid[T]) RingSeq(index geom.Point[int], radius int) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		g.walkRing(index, radius, yield)
	}
}

// appendRing appends the ring of an index the caller knows to have one.
func (g *Grid[T]) appendRing(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return g.keepInside(g.topology.appendRing(dst, index, radius), len(dst))
}

// walkRing yields the indices of Ring and reports whether it yielded them all.
func (g *Grid[T]) walkRing(index geom.Point[int], radius int, yield func(geom.Point[int]) bool) bool {
	for _, found := range g.Ring(index, radius) {
		if !yield(found) {
			return false
		}
	}

	return true
}

// hasRing reports whether the index is inside the grid and a cell of it can be at radius steps
// from the index: none is past the farthest cell.
func (g *Grid[T]) hasRing(index geom.Point[int], radius int) bool {
	return radius >= 0 && g.Contains(index) && g.clampRadius(index, radius) == radius
}

// Spiral returns the indices within radius of the index that lie in the grid, ring by ring
// outward in the order of Ring. It returns nil for a negative radius.
func (g *Grid[T]) Spiral(index geom.Point[int], radius int) []geom.Point[int] {
	if radius < 0 || !g.Contains(index) {
		return nil
	}

	radius = g.clampRadius(index, radius)

	return g.appendSpiral(make([]geom.Point[int], 0, g.topology.rangeLen(radius)), index, radius)
}

// AppendSpiral appends the indices of Spiral to dst and returns the extended slice.
func (g *Grid[T]) AppendSpiral(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	if radius < 0 || !g.Contains(index) {
		return dst
	}

	return g.appendSpiral(dst, index, g.clampRadius(index, radius))
}

// SpiralSeq returns an iterator over the indices of Spiral, in the same order.
func (g *Grid[T]) SpiralSeq(index geom.Point[int], radius int) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		g.walkSpiral(index, radius, yield)
	}
}

// appendSpiral appends the spiral of an index the caller knows to be inside the grid.
func (g *Grid[T]) appendSpiral(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return g.keepInside(g.topology.appendSpiral(dst, index, radius), len(dst))
}

// walkSpiral yields the indices of Spiral and reports whether it yielded them all.
func (g *Grid[T]) walkSpiral(index geom.Point[int], radius int, yield func(geom.Point[int]) bool) bool {
	for _, found := range g.Spiral(index, radius) {
		if !yield(found) {
			return false
		}
	}

	return true
}

// keepInside returns dst with the indices from start on that lie outside the grid taken out,
// in place and in their order.
func (g *Grid[T]) keepInside(dst []geom.Point[int], start int) []geom.Point[int] {
	kept := dst[:start]
	for _, candidate := range dst[start:] {
		if g.Contains(candidate) {
			kept = append(kept, candidate)
		}
	}

	return kept
}

// Distance returns the number of steps between two indices, whether or not they lie in the
// grid.
func (g *Grid[T]) Distance(from, to geom.Point[int]) int {
	return g.topology.distance(from, to)
}

// Line returns the indices of the straight line from one index to the other, both included,
// each a neighbor of the one before. Unlike Ring and Spiral it is measured between any two
// indices, whether or not they lie in the grid, and between two cells on a zigzag edge of a
// hexagonal grid it may pass through an index outside it.
func (g *Grid[T]) Line(from, to geom.Point[int]) []geom.Point[int] {
	return g.topology.appendLine(make([]geom.Point[int], 0, g.topology.distance(from, to)+1), from, to)
}

// AppendLine appends the indices of Line to dst and returns the extended slice.
func (g *Grid[T]) AppendLine(dst []geom.Point[int], from, to geom.Point[int]) []geom.Point[int] {
	return g.topology.appendLine(dst, from, to)
}

// LineSeq returns an iterator over the indices of Line, in the same order.
func (g *Grid[T]) LineSeq(from, to geom.Point[int]) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		g.walkLine(from, to, yield)
	}
}

// walkLine yields the indices of Line and reports whether it yielded them all.
func (g *Grid[T]) walkLine(from, to geom.Point[int], yield func(geom.Point[int]) bool) bool {
	for _, found := range g.Line(from, to) {
		if !yield(found) {
			return false
		}
	}

	return true
}

// Contains reports whether the index is that of a cell of the grid.
func (g *Grid[T]) Contains(index geom.Point[int]) bool {
	return g.cells.Contains(index)
}

// topology is the neighbor rule of one kind of cell, the primitives every neighbor query of a
// grid is built on.
type topology interface {
	// offsets returns the steps from the index to its neighbors, in the order of the directions
	// of the kind. The slice is the topology's own and is never written to.
	offsets(index geom.Point[int]) []geom.Vector[int]

	// distance returns the number of steps between two indices, whether or not they lie inside
	// a grid.
	distance(from, to geom.Point[int]) int

	// rangeLen returns the number of indices within radius of an index, what appendWithin
	// appends.
	rangeLen(radius int) int

	// appendWithin appends every index within radius of index, inside a grid or not, in the
	// order of a region, so a range kept in that order is a region with nothing to sort.
	appendWithin(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int]

	// ringLen returns the number of indices at exactly radius steps from an index, what
	// appendRing appends.
	ringLen(radius int) int

	// appendRing appends every index at exactly radius steps from index, inside a grid or not,
	// in the order the kind walks around it.
	appendRing(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int]

	// appendSpiral appends every index within radius of index, inside a grid or not, ring by
	// ring outward from it.
	appendSpiral(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int]

	// appendLine appends the indices of the straight line from one index to another, both
	// included, inside a grid or not.
	appendLine(dst []geom.Point[int], from, to geom.Point[int]) []geom.Point[int]

	// sight marks as seen the indices within radius of center that are in sight of it past the
	// indices the marks block, neither end blocking: every one of the candidates that is, and
	// any others the kind finds on the way.
	sight(center geom.Point[int], radius int, candidates []geom.Point[int], marks *sightMarks)
}

// sightMarks is what a range knows of the square of indices around its center: which cells it
// was refused and which it found in sight past them. The marks are made at the first refusal,
// so a range refused nothing makes none, and the cells beyond the edge of the grid block as the
// refused ones do.
type sightMarks struct {
	size   geom.Size[int]
	corner geom.Point[int]
	side   int
	marks  []uint8
}

const (
	markRefused uint8 = 1 << iota
	markSeen
)

// refuse marks the cell at the index as refused.
func (m *sightMarks) refuse(index geom.Point[int]) {
	if m.marks == nil {
		m.marks = make([]uint8, m.side*m.side)
	}

	m.marks[m.offset(index)] |= markRefused
}

// see marks the cell at the index as in sight.
func (m *sightMarks) see(index geom.Point[int]) {
	m.marks[m.offset(index)] |= markSeen
}

// blocks reports whether the index is beyond the edge of the grid or that of a cell refused.
func (m *sightMarks) blocks(index geom.Point[int]) bool {
	if index.X < 0 || index.X >= m.size.Width || index.Y < 0 || index.Y >= m.size.Height {
		return true
	}

	x, y := index.X-m.corner.X, index.Y-m.corner.Y

	return x >= 0 && x < m.side && y >= 0 && y < m.side && m.marks[x+y*m.side]&markRefused != 0
}

// seen reports whether the cell at the index, inside the square, was marked as in sight.
func (m *sightMarks) seen(index geom.Point[int]) bool {
	return m.marks[m.offset(index)]&markSeen != 0
}

// isEmpty reports whether no cell was refused.
func (m *sightMarks) isEmpty() bool {
	return m.marks == nil
}

// offset returns the place of the mark of an index inside the square.
func (m *sightMarks) offset(index geom.Point[int]) int {
	return index.X - m.corner.X + (index.Y-m.corner.Y)*m.side
}
