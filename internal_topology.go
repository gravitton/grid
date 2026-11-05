package grid

import (
	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// topology is the neighbor rule of one kind of cell, the primitives every neighbor query of a
// grid is built on: the methods of Square under a movement, or those of hex.Hex reached through
// an offset coordinate system. It is one struct asking which it is in each method, never an
// interface, so the compiler sees through a walk and its closures stay off the heap.
type topology struct {
	// movement chooses the neighbors of square cells, and is MovementNone for hexagonal ones.
	movement Movement

	// system is the coordinate system of hexagonal cells, and is none for square ones.
	system hex.CoordinateSystem

	// steps holds the neighbor offsets of an index of even and of odd parity, since the shifted
	// rows or columns of an offset system step differently from the others. Square cells hold
	// the same offsets twice.
	steps [2][]geom.Vector[int]

	// byColumn is true where the parity is that of the column, as in a flat-top grid.
	byColumn bool
}

// newTopology returns the neighbor rule of the kind, the movement that of a kind of square cells
// alone.
func newTopology(kind kind, movement Movement) topology {
	system := kind.system()

	if kind.sides() == 6 {
		even, odd := system.Offsets(geom.Pt(0, 0)), system.Offsets(geom.Pt(1, 1))

		return topology{
			movement: MovementNone,
			system:   system,
			steps:    [2][]geom.Vector[int]{even[:], odd[:]},
			byColumn: system == hex.OffsetOddQ || system == hex.OffsetEvenQ,
		}
	}

	steps := movement.Offsets()

	return topology{movement: movement, system: system, steps: [2][]geom.Vector[int]{steps, steps}}
}

// isHex reports whether the cells are hexagonal.
func (t *topology) isHex() bool {
	return !t.system.IsNone()
}

// offsets returns the steps from the index to its neighbors, in the order of the directions of
// the kind. The slice is the topology's own and is never written to.
func (t *topology) offsets(index Index) []geom.Vector[int] {
	if t.byColumn {
		return t.steps[index.X&1]
	}

	return t.steps[index.Y&1]
}

// distance returns the number of steps between two indices, whether or not they lie inside a
// grid.
func (t *topology) distance(from, to Index) int {
	if t.isHex() {
		return t.system.From(from).DistanceTo(t.system.From(to))
	}

	return Square(from).DistanceTo(to, t.movement)
}

// rangeLen returns the number of indices within radius of an index, what walkWithin yields.
func (t *topology) rangeLen(radius int) int {
	if t.isHex() {
		return hex.RangeLen(radius)
	}

	return RangeLen(radius, t.movement)
}

// rangeLenIn returns the number of indices within radius of index that lie in a grid of the
// given size, what walkWithin yields inside it. A range short of every edge is all inside, and
// any other is counted a line of the lattice at a time.
func (t *topology) rangeLenIn(index Index, radius int, size geom.Size[int]) int {
	if radius <= min(index.X, index.Y, size.Width-1-index.X, size.Height-1-index.Y) {
		return t.rangeLen(radius)
	}
	if t.isHex() {
		return t.hexRangeLenIn(index, radius, size)
	}

	count := 0
	for x := max(index.X-radius, 0); x <= min(index.X+radius, size.Width-1); x++ {
		reach := radius
		if t.movement == MovementCardinal {
			reach -= geom.Abs(x - index.X)
		}

		count += spanLen(index.Y-reach, index.Y+reach, size.Height)
	}

	return count
}

// hexRangeLenIn is rangeLenIn for hexagonal cells. The hexes within radius of one axial step
// along the axis of the parity form a run of one column, or of one row, of the offset system.
func (t *topology) hexRangeLenIn(index Index, radius int, size geom.Size[int]) int {
	center := t.system.From(index)

	count := 0
	for step := -radius; step <= radius; step++ {
		low, high := max(-radius, -step-radius), min(radius, -step+radius)

		if t.byColumn {
			first, last := t.system.To(center.Add(hex.Pt(step, low))), t.system.To(center.Add(hex.Pt(step, high)))
			if first.X >= 0 && first.X < size.Width {
				count += spanLen(first.Y, last.Y, size.Height)
			}

			continue
		}

		first, last := t.system.To(center.Add(hex.Pt(low, step))), t.system.To(center.Add(hex.Pt(high, step)))
		if first.Y >= 0 && first.Y < size.Height {
			count += spanLen(first.X, last.X, size.Width)
		}
	}

	return count
}

// spanLen returns the number of whole numbers from low to high that lie in [0, length).
func spanLen(low, high, length int) int {
	return max(min(high, length-1)-max(low, 0)+1, 0)
}

// walkWithin yields every index within radius of index, inside a grid or not, in the order of a
// region, so a range kept in that order is a region with nothing to sort, and reports whether it
// yielded them all.
func (t *topology) walkWithin(index Index, radius int, yield func(Index) bool) bool {
	if t.isHex() {
		return t.walkHexWithin(index, radius, yield)
	}

	return Square(index).within(radius, t.movement, yield)
}

// walkHexWithin is walkWithin for hexagonal cells. The range of a hex comes in the order of a
// region only where the parity is that of the column, so otherwise it walks the box around it.
func (t *topology) walkHexWithin(index Index, radius int, yield func(Index) bool) bool {
	center := t.system.From(index)

	if t.byColumn {
		for candidate := range center.RangeSeq(radius) {
			if !yield(t.system.To(candidate)) {
				return false
			}
		}

		return true
	}

	for x := index.X - radius; x <= index.X+radius; x++ {
		for y := index.Y - radius; y <= index.Y+radius; y++ {
			if candidate := geom.Pt(x, y); center.DistanceTo(t.system.From(candidate)) <= radius && !yield(candidate) {
				return false
			}
		}
	}

	return true
}

// ringLenIn returns the number of indices at exactly radius steps from index that lie in a grid
// of the given size, what walkRing yields inside it.
func (t *topology) ringLenIn(index Index, radius int, size geom.Size[int]) int {
	return t.rangeLenIn(index, radius, size) - t.rangeLenIn(index, radius-1, size)
}

// walkRing yields every index at exactly radius steps from index, inside a grid or not, in the
// order the kind walks around it, and reports whether it yielded them all.
func (t *topology) walkRing(index Index, radius int, yield func(Index) bool) bool {
	if t.isHex() {
		for candidate := range t.system.From(index).RingSeq(radius) {
			if !yield(t.system.To(candidate)) {
				return false
			}
		}

		return true
	}

	return Square(index).ring(radius, t.movement, yield)
}

// walkSpiral yields every index within radius of index, inside a grid or not, ring by ring
// outward from it, and reports whether it yielded them all.
func (t *topology) walkSpiral(index Index, radius int, yield func(Index) bool) bool {
	if t.isHex() {
		for candidate := range t.system.From(index).SpiralSeq(radius) {
			if !yield(t.system.To(candidate)) {
				return false
			}
		}

		return true
	}

	return Square(index).spiral(radius, t.movement, yield)
}

// walkLine yields the indices of the straight line from one index to another, both included,
// inside a grid or not, and reports whether it yielded them all.
func (t *topology) walkLine(from, to Index, yield func(Index) bool) bool {
	if t.isHex() {
		for candidate := range t.system.From(from).LineToSeq(t.system.From(to)) {
			if !yield(t.system.To(candidate)) {
				return false
			}
		}

		return true
	}

	return Square(from).line(to, t.movement, yield)
}

// sees reports whether from sees to past the indices blocked reports, neither end blocking. A nil
// blocked blocks nothing.
func (t *topology) sees(from, to Index, blocked func(Index) bool) bool {
	if blocked == nil {
		return true
	}

	if t.isHex() {
		return t.system.From(from).HasLineOfSight(t.system.From(to), func(candidate hex.Hex) bool {
			return blocked(t.system.To(candidate))
		})
	}

	return Square(from).hasLineOfSight(to, blocked)
}

// walkFieldOfView yields the indices within radius of center it sees past the indices blocked
// reports, inside a grid or not, as it walks them, and reports whether it yielded them all. A nil
// blocked blocks nothing.
func (t *topology) walkFieldOfView(center Index, radius int, blocked func(Index) bool, yield func(Index) bool) bool {
	if t.isHex() {
		return t.walkHexFieldOfView(center, radius, blocked, yield)
	}

	return Square(center).fieldOfView(radius, t.movement, blocked, yield)
}

// walkHexFieldOfView is walkFieldOfView for hexagonal cells.
func (t *topology) walkHexFieldOfView(center Index, radius int, blocked func(Index) bool, yield func(Index) bool) bool {
	if blocked == nil {
		return t.walkSpiral(center, radius, yield)
	}

	sight := t.system.From(center).FieldOfViewSeq(radius, func(candidate hex.Hex) bool {
		return blocked(t.system.To(candidate))
	})
	for visible := range sight {
		if !yield(t.system.To(visible)) {
			return false
		}
	}

	return true
}

// sight marks as seen the indices within radius of center that are in sight of it past the
// indices the marks block, neither end blocking: every one of the candidates that is, and any
// others the kind finds on the way.
func (t *topology) sight(center Index, radius int, candidates []Index, marks *sightMarks) {
	if t.isHex() {
		sight := t.system.From(center).FieldOfViewSeq(radius, func(candidate hex.Hex) bool {
			return marks.blocks(t.system.To(candidate))
		})
		for visible := range sight {
			marks.see(t.system.To(visible))
		}

		return
	}

	Square(center).markSeen(candidates, marks)
}
