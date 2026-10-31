package grid

import (
	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

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

// squareTopology is the neighbor rule of square cells under a movement, the methods of Point
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

func (t *squareTopology) offsets(geom.Point[int]) []geom.Vector[int] {
	return t.steps
}

func (t *squareTopology) distance(from, to geom.Point[int]) int {
	return Point(from).DistanceTo(to, t.movement)
}

func (t *squareTopology) rangeLen(radius int) int {
	return RangeLen(radius, t.movement)
}

func (t *squareTopology) appendWithin(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return Point(index).AppendRange(dst, radius, t.movement)
}

func (t *squareTopology) ringLen(radius int) int {
	return RingLen(radius, t.movement)
}

func (t *squareTopology) appendRing(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return Point(index).AppendRing(dst, radius, t.movement)
}

func (t *squareTopology) appendSpiral(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return Point(index).AppendSpiral(dst, radius, t.movement)
}

func (t *squareTopology) appendLine(dst []geom.Point[int], from, to geom.Point[int]) []geom.Point[int] {
	return Point(from).AppendLine(dst, to, t.movement)
}

func (t *squareTopology) sight(center geom.Point[int], _ int, candidates []geom.Point[int], marks *sightMarks) {
	blocks := blockers{blocked: marks.blocks}
	for _, candidate := range candidates {
		if Point(center).sees(candidate, blocks) {
			marks.see(candidate)
		}
	}
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

func (t *hexTopology) offsets(index geom.Point[int]) []geom.Vector[int] {
	if t.byColumn {
		return t.steps[index.X&1][:]
	}

	return t.steps[index.Y&1][:]
}

func (t *hexTopology) distance(from, to geom.Point[int]) int {
	return t.system.From(from).DistanceTo(t.system.From(to))
}

func (t *hexTopology) rangeLen(radius int) int {
	return hex.RangeLen(radius)
}

func (t *hexTopology) appendWithin(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
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

func (t *hexTopology) appendRing(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	for candidate := range t.system.From(index).RingSeq(radius) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) appendSpiral(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	for candidate := range t.system.From(index).SpiralSeq(radius) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) appendLine(dst []geom.Point[int], from, to geom.Point[int]) []geom.Point[int] {
	for candidate := range t.system.From(from).LineSeq(t.system.From(to)) {
		dst = append(dst, t.system.To(candidate))
	}

	return dst
}

func (t *hexTopology) sight(center geom.Point[int], radius int, _ []geom.Point[int], marks *sightMarks) {
	sight := t.system.From(center).FieldOfViewRangeSeq(radius, func(candidate hex.Hex) bool {
		return marks.blocks(t.system.To(candidate))
	})
	for visible := range sight {
		marks.see(t.system.To(visible))
	}
}
