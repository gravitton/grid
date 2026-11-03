package grid

import (
	"iter"
	"math"
	"slices"

	geom "github.com/gravitton/geometry"
)

// Square is an index of a square grid, with the lattice queries [geom.Point] does not have: the
// neighbors, ranges, rings and lines around it and the sight from it. A method whose answer
// depends on which cells are adjacent takes the [Movement].
type Square geom.Point[int]

// Sq is shorthand for Square{x, y}.
func Sq(x, y int) Square {
	return Square{X: x, Y: y}
}

// RangeLen returns the length of Range and Spiral for the radius under the movement
// (1+2r(r+1) cardinal, (2r+1)² diagonal), one for a movement outside the constants and zero
// for a negative radius.
func RangeLen(radius int, movement Movement) int {
	count := movement.neighborsLen()
	if radius < 0 {
		return 0
	}

	return 1 + count*radius*(radius+1)/2
}

// RingLen returns the length of Ring for the radius under the movement (4r cardinal, 8r
// diagonal), one for a zero radius and zero for a negative one.
func RingLen(radius int, movement Movement) int {
	count := movement.neighborsLen()

	switch {
	case radius < 0:
		return 0
	case radius == 0:
		return 1
	default:
		return count * radius
	}
}

// SectorLen returns the length of Sector for the radius, the turns and the movement, of n
// directions (r+t·r(r+1)/2, less ⌈r/2⌉ for odd t, and n·r(r+1)/2 from n turns on), and zero for a
// negative radius or negative turns and for a movement outside the constants.
func SectorLen(radius, turns int, movement Movement) int {
	count := movement.neighborsLen()
	if radius < 0 || turns < 0 || count == 0 {
		return 0
	}
	if turns >= count {
		return count * radius * (radius + 1) / 2
	}

	length := radius + turns*radius*(radius+1)/2
	if turns%2 == 1 {
		return length - (radius+1)/2
	}

	return length
}

// Turn returns s turned about the origin by steps quarter-turns of increasing angle; negative
// steps go the other way.
func (s Square) Turn(steps int) Square {
	switch ((steps % 4) + 4) % 4 {
	case 1:
		return Square{X: -s.Y, Y: s.X}
	case 2:
		return Square{X: -s.X, Y: -s.Y}
	case 3:
		return Square{X: s.Y, Y: -s.X}
	default:
		return s
	}
}

// TurnAround returns s turned about pivot by steps quarter-turns of increasing angle; negative
// steps go the other way.
func (s Square) TurnAround(pivot Index, steps int) Square {
	turned := Square{X: s.X - pivot.X, Y: s.Y - pivot.Y}.Turn(steps)

	return Square{X: turned.X + pivot.X, Y: turned.Y + pivot.Y}
}

// ReflectX returns s reflected across the x-axis, its row negated.
func (s Square) ReflectX() Square {
	return Square{X: s.X, Y: -s.Y}
}

// ReflectY returns s reflected across the y-axis, its column negated.
func (s Square) ReflectY() Square {
	return Square{X: -s.X, Y: s.Y}
}

// Neighbor returns the index one step from s in the direction, and s itself for
// geom.DirectionNone.
func (s Square) Neighbor(direction geom.Direction) Square {
	return Square(s.Point().Add(direction.Offset[int]()))
}

// Neighbors returns the indices adjacent to s under the movement, by increasing angle from the
// one to the right, and nil for a movement outside the constants.
func (s Square) Neighbors(movement Movement) []Index {
	count := movement.neighborsLen()
	if count == 0 {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, count), s.NeighborsSeq(movement))
}

// NeighborsSeq returns an iterator over the indices of Neighbors, in the same order.
func (s Square) NeighborsSeq(movement Movement) iter.Seq[Index] {
	return s.RingSeq(1, movement)
}

// Range returns the indices within radius of s under the movement, s included, in the order of
// a region: a diamond for MovementCardinal, a square for MovementDiagonal and s alone for a
// movement outside the constants. It returns nil for a negative radius.
func (s Square) Range(radius int, movement Movement) []Index {
	size := RangeLen(radius, movement)
	if radius < 0 {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, size), s.RangeSeq(radius, movement))
}

// RangeSeq returns an iterator over the indices of Range, in the same order.
func (s Square) RangeSeq(radius int, movement Movement) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		s.within(radius, movement, yield)
	}
}

// within yields the indices within radius of s column by column, each as far up and down as
// the movement reaches, and reports whether it yielded them all.
func (s Square) within(radius int, movement Movement, yield func(Index) bool) bool {
	if movement.IsNone() {
		return radius < 0 || yield(s.Point())
	}

	for dx := -radius; dx <= radius; dx++ {
		reach := radius
		if movement == MovementCardinal {
			reach -= geom.Abs(dx)
		}

		for dy := -reach; dy <= reach; dy++ {
			if !yield(geom.Pt(s.X+dx, s.Y+dy)) {
				return false
			}
		}
	}

	return true
}

// Ring returns the indices at exactly radius steps from s under the movement, by increasing
// angle from the one to the right, so the ring of radius one is Neighbors. It returns nil for a
// negative radius and s alone for a zero one.
func (s Square) Ring(radius int, movement Movement) []Index {
	size := RingLen(radius, movement)
	if radius < 0 {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, size), s.RingSeq(radius, movement))
}

// RingSeq returns an iterator over the indices of Ring, in the same order.
func (s Square) RingSeq(radius int, movement Movement) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		s.ring(radius, movement, yield)
	}
}

// ring yields the indices at radius steps from s, walking from each neighbor step scaled by
// radius toward the next, and reports whether it yielded them all.
func (s Square) ring(radius int, movement Movement, yield func(Index) bool) bool {
	if radius < 0 {
		return true
	}
	if radius == 0 {
		return yield(s.Point())
	}

	corners, count := movement.corners(), movement.neighborsLen()
	offsets := corners[:count]

	for i, corner := range offsets {
		next := offsets[(i+1)%len(offsets)]

		for j := range radius {
			if !yield(geom.Pt(s.X+corner.X*radius+(next.X-corner.X)*j, s.Y+corner.Y*radius+(next.Y-corner.Y)*j)) {
				return false
			}
		}
	}

	return true
}

// Spiral returns the indices of Range ring by ring outward from s, each in the order of Ring.
// It returns nil for a negative radius.
func (s Square) Spiral(radius int, movement Movement) []Index {
	size := RangeLen(radius, movement)
	if radius < 0 {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, size), s.SpiralSeq(radius, movement))
}

// SpiralSeq returns an iterator over the indices of Spiral, in the same order.
func (s Square) SpiralSeq(radius int, movement Movement) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		s.spiral(radius, movement, yield)
	}
}

// spiral yields the rings from zero to radius steps in turn and reports whether it yielded them
// all.
func (s Square) spiral(radius int, movement Movement, yield func(Index) bool) bool {
	for distance := 0; distance <= radius; distance++ {
		if !s.ring(distance, movement, yield) {
			return false
		}
	}

	return true
}

// Sector returns the indices within radius of s and within turns steps between the directions of
// the movement about the direction, half of them either side, both borders included, ring by ring
// outward, by increasing angle in each ring. Zero turns is the ray along the direction, and as
// many turns as the movement has directions the whole range, each ring from the index opposite
// the direction. Two sectors of one turn about neighboring directions share the indices on their
// border. s itself is left out, so it returns nil for a radius under one, for negative turns and
// for a direction that is not one of the movement.
func (s Square) Sector(radius int, direction geom.Direction, turns int, movement Movement) []Index {
	if radius < 1 || turns < 0 || !movement.Contains(direction) {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, SectorLen(radius, turns, movement)), s.SectorSeq(radius, direction, turns, movement))
}

// SectorSeq returns an iterator over the indices of Sector, in the same order.
func (s Square) SectorSeq(radius int, direction geom.Direction, turns int, movement Movement) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		s.sector(radius, direction, turns, movement, yield)
	}
}

// sector yields the indices of Sector and reports whether it yielded them all. A ring holds ring
// indices between two corners, so it walks turns·ring/2 of them either side of the corner of the
// direction, as hex.Hex.SectorSeq does around a ring of hexes.
func (s Square) sector(radius int, direction geom.Direction, turns int, movement Movement, yield func(Index) bool) bool {
	if turns < 0 || !movement.Contains(direction) {
		return true
	}

	corners, count := movement.corners(), movement.neighborsLen()
	axis := 0
	for corners[axis] != direction.Offset[int]() {
		axis++
	}

	turns = min(turns, count)
	for ring := 1; ring <= radius; ring++ {
		reach, length := turns*ring/2, count*ring
		start := (axis*ring - reach + length) % length
		side, step := start/ring, start%ring

		for range min(2*reach+1, length) {
			corner, next := corners[side], corners[(side+1)%count]
			if !yield(geom.Pt(s.X+corner.X*ring+(next.X-corner.X)*step, s.Y+corner.Y*ring+(next.Y-corner.Y)*step)) {
				return false
			}

			step++
			if step == ring {
				side, step = (side+1)%count, 0
			}
		}
	}

	return true
}

// DistanceTo returns the number of steps from s to the index under the movement: the Manhattan
// distance for MovementCardinal, the Chebyshev distance for MovementDiagonal. It panics for a
// movement outside the constants.
func (s Square) DistanceTo(index Index, movement Movement) int {
	return movement.Distance(s.Point(), index)
}

// DirectionTo returns the direction of the movement nearest to the one from s to the index, and
// geom.DirectionNone for s itself and for a movement outside the constants. Under
// MovementCardinal an index on a diagonal answers the horizontal direction.
func (s Square) DirectionTo(index Index, movement Movement) geom.Direction {
	dx, dy := index.X-s.X, index.Y-s.Y

	switch {
	case dx == 0 && dy == 0, movement.IsNone():
		return geom.DirectionNone
	case movement == MovementDiagonal:
		return geom.DirectionFromAngle(math.Atan2(float64(dy), float64(dx)))
	case geom.Abs(dx) >= geom.Abs(dy) && dx > 0:
		return geom.DirectionRight
	case geom.Abs(dx) >= geom.Abs(dy):
		return geom.DirectionLeft
	case dy > 0:
		return geom.DirectionDown
	default:
		return geom.DirectionUp
	}
}

// LineTo returns the indices of the straight line from s to the index, both included, each a
// neighbor of the one before under the movement. It panics for a movement outside the constants.
func (s Square) LineTo(index Index, movement Movement) []Index {
	return slices.AppendSeq(make([]Index, 0, s.DistanceTo(index, movement)+1), s.LineToSeq(index, movement))
}

// LineToSeq returns an iterator over the indices of LineTo, in the same order. It panics, once
// ranged, for a movement outside the constants.
func (s Square) LineToSeq(index Index, movement Movement) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		s.line(index, movement, yield)
	}
}

// line yields the indices from s to the index and reports whether it yielded them all. Under a
// cardinal movement each step goes along the axis further behind the line, i and j counting the
// steps along each; under a diagonal one it is the Bresenham walk.
func (s Square) line(index Index, movement Movement, yield func(Index) bool) bool {
	if movement.IsNone() {
		panic(movement.unknown())
	}

	x, y := s.X, s.Y
	dx, dy := geom.Abs(index.X-x), geom.Abs(index.Y-y)
	sx, sy := geom.Sign(index.X-x), geom.Sign(index.Y-y)

	if !yield(geom.Pt(x, y)) {
		return false
	}

	if movement == MovementCardinal {
		for i, j := 0, 0; i < dx || j < dy; {
			if (1+2*i)*dy < (1+2*j)*dx {
				x += sx
				i++
			} else {
				y += sy
				j++
			}

			if !yield(geom.Pt(x, y)) {
				return false
			}
		}

		return true
	}

	for err := dx - dy; x != index.X || y != index.Y; {
		doubled := 2 * err
		if doubled > -dy {
			err -= dy
			x += sx
		}
		if doubled < dx {
			err += dx
			y += sy
		}

		if !yield(geom.Pt(x, y)) {
			return false
		}
	}

	return true
}

// HasLineOfSight reports whether s sees target past the indices blocked reports, asking it only
// about the indices the line passes and the two beside each corner it crosses. Where it crosses a
// corner it is stopped only when both cells beside it block, so a diagonal run of blockers is
// opaque and a single one is not. It differs from LineTo: a cell off the line can block it.
func (s Square) HasLineOfSight(target Index, blocked func(Index) bool) bool {
	return blocked == nil || s.hasLineOfSight(target, blocked)
}

// FieldOfView returns the indices of Spiral that HasLineOfSight sees past the indices blocked
// reports, in that order, asking it about each index at most once. It returns nil for a negative
// radius.
func (s Square) FieldOfView(radius int, movement Movement, blocked func(Index) bool) []Index {
	if radius < 0 {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, RangeLen(radius, movement)), s.FieldOfViewSeq(radius, movement, blocked))
}

// FieldOfViewSeq returns an iterator over the indices of FieldOfView, in the same order.
func (s Square) FieldOfViewSeq(radius int, movement Movement, blocked func(Index) bool) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		s.fieldOfView(radius, movement, blocked, yield)
	}
}

// fieldOfView yields the indices of Spiral s sees as it walks them, asking about an index the
// first time a line passes it, and reports whether it yielded them all.
func (s Square) fieldOfView(radius int, movement Movement, blocked func(Index) bool, yield func(Index) bool) bool {
	if blocked == nil {
		return s.spiral(radius, movement, yield)
	}

	memo := newSightMemo(s.Point(), radius, blocked)

	return s.spiral(radius, movement, func(index Index) bool {
		return !s.hasLineOfSight(index, memo.has) || yield(index)
	})
}

// markSeen marks the candidates s sees past the cells the marks block.
func (s Square) markSeen(candidates []Index, marks *sightMarks) {
	for _, candidate := range candidates {
		if s.hasLineOfSight(candidate, marks.blocks) {
			marks.see(candidate)
		}
	}
}

// hasLineOfSight is the sight rule every form shares, asking blocked about the cells the line
// from s to target passes.
func (s Square) hasLineOfSight(target Index, blocked func(Index) bool) bool {
	// A Bresenham walk is not mirror-symmetric, so it always starts at the lower end; neither end
	// is asked about, so the swap changes nothing else.
	from, to := s.Point(), target
	if to.Compare(from) < 0 {
		from, to = to, from
	}

	x0, y0 := from.XY()
	x1, y1 := to.XY()

	// Starting at the lower end, the walk never steps back along x.
	dx, dy := x1-x0, geom.Abs(y1-y0)
	sy := geom.Sign(y1 - y0)

	x, y, err := x0, y0, dx-dy

	for {
		if x == x1 && y == y1 {
			return true
		}

		doubled := 2 * err
		stepX := doubled > -dy
		stepY := doubled < dx

		// Crossing a corner is stopped only when both cells beside it block.
		if stepX && stepY && blocked(geom.Pt(x+1, y)) && blocked(geom.Pt(x, y+sy)) {
			return false
		}

		if stepX {
			err -= dy
			x++
		}
		if stepY {
			err += dx
			y += sy
		}

		if x == x1 && y == y1 {
			return true
		}
		if blocked(geom.Pt(x, y)) {
			return false
		}
	}
}

// Point returns s as the geom.Point it is.
func (s Square) Point() geom.Point[int] {
	return geom.Point[int](s)
}

// String returns s as its coordinates: (x,y).
func (s Square) String() string {
	return s.Point().String()
}
