package grid

import (
	"slices"

	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// topology is the neighbor rule of one kind of cell, the primitives every neighbor query of a
// grid is built on.
type topology interface {
	// rule returns the movement of the kind, MovementNone where it has none.
	rule() Movement

	// offsets returns the steps from the index to its neighbors, in the order of the directions
	// of the kind. The slice is the topology's own and is never written to.
	offsets(index Index) []geom.Vector[int]

	// distance returns the number of steps between two indices, whether or not they lie inside
	// a grid.
	distance(from, to Index) int

	// rangeLen returns the number of indices within radius of an index, what appendWithin
	// appends.
	rangeLen(radius int) int

	// appendWithin appends every index within radius of index, inside a grid or not, in the
	// order of a region, so a range kept in that order is a region with nothing to sort.
	appendWithin(dst []Index, index Index, radius int) []Index

	// ringLen returns the number of indices at exactly radius steps from an index, what
	// appendRing appends.
	ringLen(radius int) int

	// appendRing appends every index at exactly radius steps from index, inside a grid or not,
	// in the order the kind walks around it.
	appendRing(dst []Index, index Index, radius int) []Index

	// appendSpiral appends every index within radius of index, inside a grid or not, ring by
	// ring outward from it.
	appendSpiral(dst []Index, index Index, radius int) []Index

	// walkSpiral yields what appendSpiral appends as it walks and reports whether it yielded it
	// all.
	walkSpiral(index Index, radius int, yield func(Index) bool) bool

	// appendLine appends the indices of the straight line from one index to another, both
	// included, inside a grid or not.
	appendLine(dst []Index, from, to Index) []Index

	// sees reports whether from sees to past the indices blocked reports, neither end blocking.
	// A nil blocked blocks nothing.
	sees(from, to Index, blocked func(Index) bool) bool

	// sight marks as seen the indices within radius of center that are in sight of it past the
	// indices the marks block, neither end blocking: every one of the candidates that is, and
	// any others the kind finds on the way.
	sight(center Index, radius int, candidates []Index, marks *sightMarks)
}

// newTopology returns the neighbor rule of the kind, the movement that of a kind of square cells
// alone.
func newTopology(kind kind, movement Movement) topology {
	if kind.sides() == 4 {
		return newSquareTopology(movement)
	}

	return newHexTopology(kind.system())
}

// squareTopology is the neighbor rule of square cells under a movement, the methods of Square
// with that movement.
type squareTopology struct {
	movement Movement

	// steps holds the neighbor offsets of the movement, the same from every index.
	steps []geom.Vector[int]
}

// newSquareTopology returns the neighbor rule of square cells under the movement.
func newSquareTopology(movement Movement) *squareTopology {
	return &squareTopology{
		movement: movement,
		steps:    movement.Offsets(),
	}
}

func (t *squareTopology) rule() Movement {
	return t.movement
}

func (t *squareTopology) offsets(Index) []geom.Vector[int] {
	return t.steps
}

func (t *squareTopology) distance(from, to Index) int {
	return Square(from).DistanceTo(to, t.movement)
}

func (t *squareTopology) rangeLen(radius int) int {
	return RangeLen(radius, t.movement)
}

func (t *squareTopology) appendWithin(dst []Index, index Index, radius int) []Index {
	return slices.AppendSeq(dst, Square(index).RangeSeq(radius, t.movement))
}

func (t *squareTopology) ringLen(radius int) int {
	return RingLen(radius, t.movement)
}

func (t *squareTopology) appendRing(dst []Index, index Index, radius int) []Index {
	return slices.AppendSeq(dst, Square(index).RingSeq(radius, t.movement))
}

func (t *squareTopology) appendSpiral(dst []Index, index Index, radius int) []Index {
	return slices.AppendSeq(dst, Square(index).SpiralSeq(radius, t.movement))
}

func (t *squareTopology) walkSpiral(index Index, radius int, yield func(Index) bool) bool {
	return Square(index).spiral(radius, t.movement, yield)
}

func (t *squareTopology) appendLine(dst []Index, from, to Index) []Index {
	return slices.AppendSeq(dst, Square(from).LineToSeq(to, t.movement))
}

func (t *squareTopology) sees(from, to Index, blocked func(Index) bool) bool {
	return blocked == nil || Square(from).sees(to, blocked)
}

func (t *squareTopology) sight(center Index, _ int, candidates []Index, marks *sightMarks) {
	Square(center).markSeen(candidates, marks)
}

// hexTopology is the neighbor rule of hexagonal cells indexed in an offset coordinate system,
// the methods of hex.Hex reached through the system.
type hexTopology struct {
	system hex.CoordinateSystem

	// steps holds the neighbor offsets of an index of even and of odd parity, since the shifted
	// rows or columns of an offset system step differently from the others.
	steps [2][6]geom.Vector[int]

	// byColumn is true where the parity is that of the column, as in a flat-top grid.
	byColumn bool
}

// newHexTopology returns the neighbor rule of hexagonal cells indexed in the system.
func newHexTopology(system hex.CoordinateSystem) *hexTopology {
	return &hexTopology{
		system:   system,
		steps:    [2][6]geom.Vector[int]{system.Offsets(geom.Pt(0, 0)), system.Offsets(geom.Pt(1, 1))},
		byColumn: system == hex.OffsetOddQ || system == hex.OffsetEvenQ,
	}
}

func (t *hexTopology) rule() Movement {
	return MovementNone
}

func (t *hexTopology) offsets(index Index) []geom.Vector[int] {
	if t.byColumn {
		return t.steps[index.X&1][:]
	}

	return t.steps[index.Y&1][:]
}

func (t *hexTopology) distance(from, to Index) int {
	return t.system.From(from).DistanceTo(t.system.From(to))
}

func (t *hexTopology) rangeLen(radius int) int {
	return hex.RangeLen(radius)
}

func (t *hexTopology) appendWithin(dst []Index, index Index, radius int) []Index {
	center := t.system.From(index)
	if t.byColumn {
		for candidate := range center.RangeSeq(radius) {
			dst = append(dst, t.system.To(candidate))
		}

		return dst
	}

	for x := index.X - radius; x <= index.X+radius; x++ {
		for y := index.Y - radius; y <= index.Y+radius; y++ {
			if candidate := geom.Pt(x, y); center.DistanceTo(t.system.From(candidate)) <= radius {
				dst = append(dst, candidate)
			}
		}
	}

	return dst
}

func (t *hexTopology) ringLen(radius int) int {
	return hex.RingLen(radius)
}

func (t *hexTopology) appendRing(dst []Index, index Index, radius int) []Index {
	for candidate := range t.system.From(index).RingSeq(radius) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) appendSpiral(dst []Index, index Index, radius int) []Index {
	for candidate := range t.system.From(index).SpiralSeq(radius) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) walkSpiral(index Index, radius int, yield func(Index) bool) bool {
	for candidate := range t.system.From(index).SpiralSeq(radius) {
		if !yield(t.system.To(candidate)) {
			return false
		}
	}

	return true
}

func (t *hexTopology) appendLine(dst []Index, from, to Index) []Index {
	for candidate := range t.system.From(from).LineToSeq(t.system.From(to)) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) sees(from, to Index, blocked func(Index) bool) bool {
	if blocked == nil {
		return true
	}

	return t.system.From(from).HasLineOfSight(t.system.From(to), func(candidate hex.Hex) bool {
		return blocked(t.system.To(candidate))
	})
}

func (t *hexTopology) sight(center Index, radius int, _ []Index, marks *sightMarks) {
	sight := t.system.From(center).FieldOfViewSeq(radius, func(candidate hex.Hex) bool {
		return marks.blocks(t.system.To(candidate))
	})
	for visible := range sight {
		marks.see(t.system.To(visible))
	}
}
