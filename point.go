package grid

import (
	"iter"
	"math"

	geom "github.com/gravitton/geometry"
)

// Point is an index of a square grid, with the lattice queries [geom.Point] does not have: the
// neighbors, ranges, rings and lines around it and the sight from it. A method whose answer
// depends on which cells are adjacent takes the [Movement].
type Point geom.Point[int]

// Pt is shorthand for Point{x, y}.
func Pt(x, y int) Point {
	return Point{X: x, Y: y}
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

// Turn returns the point turned about the origin by steps quarter-turns of increasing angle;
// negative steps go the other way.
func (p Point) Turn(steps int) Point {
	switch ((steps % 4) + 4) % 4 {
	case 1:
		return Point{X: -p.Y, Y: p.X}
	case 2:
		return Point{X: -p.X, Y: -p.Y}
	case 3:
		return Point{X: p.Y, Y: -p.X}
	default:
		return p
	}
}

// TurnAround returns the point turned about center by steps quarter-turns of increasing angle;
// negative steps go the other way.
func (p Point) TurnAround(center geom.Point[int], steps int) Point {
	turned := Point{X: p.X - center.X, Y: p.Y - center.Y}.Turn(steps)

	return Point{X: turned.X + center.X, Y: turned.Y + center.Y}
}

// ReflectX returns the point reflected across the x-axis, its row negated.
func (p Point) ReflectX() Point {
	return Point{X: p.X, Y: -p.Y}
}

// ReflectY returns the point reflected across the y-axis, its column negated.
func (p Point) ReflectY() Point {
	return Point{X: -p.X, Y: p.Y}
}

// Neighbor returns the index one step from p in the direction, and p itself for
// geom.DirectionNone.
func (p Point) Neighbor(direction geom.Direction) geom.Point[int] {
	return p.Point().Add(direction.Offset[int]())
}

// Neighbors returns the indices adjacent to p under the movement, by increasing angle from the
// one to the right, and nil for a movement outside the constants.
func (p Point) Neighbors(movement Movement) []geom.Point[int] {
	count := movement.neighborsLen()
	if count == 0 {
		return nil
	}

	return p.AppendNeighbors(make([]geom.Point[int], 0, count), movement)
}

// AppendNeighbors appends the indices of Neighbors to dst and returns the extended slice.
func (p Point) AppendNeighbors(dst []geom.Point[int], movement Movement) []geom.Point[int] {
	switch movement {
	case MovementCardinal:
		for _, offset := range geom.CardinalOffsets[int]() {
			dst = append(dst, p.Point().Add(offset))
		}
	case MovementDiagonal:
		for _, offset := range geom.DirectionOffsets[int]() {
			dst = append(dst, p.Point().Add(offset))
		}
	}

	return dst
}

// Range returns the indices within radius of p under the movement, p included, in the order of
// a region: a diamond for MovementCardinal, a square for MovementDiagonal and p alone for a
// movement outside the constants. It returns nil for a negative radius.
func (p Point) Range(radius int, movement Movement) []geom.Point[int] {
	size := RangeLen(radius, movement)
	if radius < 0 {
		return nil
	}

	return p.AppendRange(make([]geom.Point[int], 0, size), radius, movement)
}

// AppendRange appends the indices of Range to dst and returns the extended slice.
func (p Point) AppendRange(dst []geom.Point[int], radius int, movement Movement) []geom.Point[int] {
	p.within(radius, movement, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// RangeSeq returns an iterator over the indices of Range, in the same order.
func (p Point) RangeSeq(radius int, movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.within(radius, movement, yield)
	}
}

// within yields the indices within radius of p column by column, each as far up and down as
// the movement reaches, and reports whether it yielded them all.
func (p Point) within(radius int, movement Movement, yield func(geom.Point[int]) bool) bool {
	if movement.IsNone() {
		return radius < 0 || yield(p.Point())
	}

	for dx := -radius; dx <= radius; dx++ {
		reach := radius
		if movement == MovementCardinal {
			reach -= geom.Abs(dx)
		}

		for dy := -reach; dy <= reach; dy++ {
			if !yield(geom.Pt(p.X+dx, p.Y+dy)) {
				return false
			}
		}
	}

	return true
}

// Ring returns the indices at exactly radius steps from p under the movement, by increasing
// angle from the one to the right, so the ring of radius one is Neighbors. It returns nil for a
// negative radius and p alone for a zero one.
func (p Point) Ring(radius int, movement Movement) []geom.Point[int] {
	size := RingLen(radius, movement)
	if radius < 0 {
		return nil
	}

	return p.AppendRing(make([]geom.Point[int], 0, size), radius, movement)
}

// AppendRing appends the indices of Ring to dst and returns the extended slice.
func (p Point) AppendRing(dst []geom.Point[int], radius int, movement Movement) []geom.Point[int] {
	p.ring(radius, movement, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// RingSeq returns an iterator over the indices of Ring, in the same order.
func (p Point) RingSeq(radius int, movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.ring(radius, movement, yield)
	}
}

// ring yields the indices at radius steps from p, walking from each neighbor step scaled by
// radius toward the next, and reports whether it yielded them all.
func (p Point) ring(radius int, movement Movement, yield func(geom.Point[int]) bool) bool {
	if radius < 0 {
		return true
	}
	if radius == 0 {
		return yield(p.Point())
	}

	cardinals, all := geom.CardinalOffsets[int](), geom.DirectionOffsets[int]()

	var offsets []geom.Vector[int]
	switch movement {
	case MovementCardinal:
		offsets = cardinals[:]
	case MovementDiagonal:
		offsets = all[:]
	}

	for i, corner := range offsets {
		next := offsets[(i+1)%len(offsets)]

		for j := range radius {
			if !yield(geom.Pt(p.X+corner.X*radius+(next.X-corner.X)*j, p.Y+corner.Y*radius+(next.Y-corner.Y)*j)) {
				return false
			}
		}
	}

	return true
}

// Spiral returns the indices of Range ring by ring outward from p, each in the order of Ring.
// It returns nil for a negative radius.
func (p Point) Spiral(radius int, movement Movement) []geom.Point[int] {
	size := RangeLen(radius, movement)
	if radius < 0 {
		return nil
	}

	return p.AppendSpiral(make([]geom.Point[int], 0, size), radius, movement)
}

// AppendSpiral appends the indices of Spiral to dst and returns the extended slice.
func (p Point) AppendSpiral(dst []geom.Point[int], radius int, movement Movement) []geom.Point[int] {
	p.spiral(radius, movement, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// SpiralSeq returns an iterator over the indices of Spiral, in the same order.
func (p Point) SpiralSeq(radius int, movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.spiral(radius, movement, yield)
	}
}

// spiral yields the rings from zero to radius steps in turn and reports whether it yielded them
// all.
func (p Point) spiral(radius int, movement Movement, yield func(geom.Point[int]) bool) bool {
	for distance := 0; distance <= radius; distance++ {
		if !p.ring(distance, movement, yield) {
			return false
		}
	}

	return true
}

// DistanceTo returns the number of steps from p to the index under the movement: the Manhattan
// distance for MovementCardinal, the Chebyshev distance for MovementDiagonal. It panics for a
// movement outside the constants.
func (p Point) DistanceTo(to geom.Point[int], movement Movement) int {
	return movement.DistanceTo(p.Point(), to)
}

// DirectionTo returns the direction of the movement nearest to the one from p to target, and
// geom.DirectionNone for p itself and for a movement outside the constants. Under
// MovementCardinal a target on a diagonal answers the horizontal direction.
func (p Point) DirectionTo(target geom.Point[int], movement Movement) geom.Direction {
	dx, dy := target.X-p.X, target.Y-p.Y

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

// Line returns the indices of the straight line from p to target, both included, each a
// neighbor of the one before under the movement. It is the traversal, not the sight rule: a
// blocker off the line can still block sight. It panics for a movement outside the constants.
func (p Point) Line(target geom.Point[int], movement Movement) []geom.Point[int] {
	return p.AppendLine(make([]geom.Point[int], 0, p.DistanceTo(target, movement)+1), target, movement)
}

// AppendLine appends the indices of Line to dst and returns the extended slice. It panics for a
// movement outside the constants.
func (p Point) AppendLine(dst []geom.Point[int], target geom.Point[int], movement Movement) []geom.Point[int] {
	p.line(target, movement, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// LineSeq returns an iterator over the indices of Line, in the same order. It panics, once
// ranged, for a movement outside the constants.
func (p Point) LineSeq(target geom.Point[int], movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.line(target, movement, yield)
	}
}

// line yields the indices from p to target and reports whether it yielded them all. Under a
// cardinal movement each step goes along the axis further behind the line, i and j counting the
// steps along each; under a diagonal one it is the Bresenham walk.
func (p Point) line(target geom.Point[int], movement Movement, yield func(geom.Point[int]) bool) bool {
	if movement.IsNone() {
		panic(movement.unknown())
	}

	x, y := p.X, p.Y
	dx, dy := geom.Abs(target.X-x), geom.Abs(target.Y-y)
	sx, sy := geom.Sign(target.X-x), geom.Sign(target.Y-y)

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

	for err := dx - dy; x != target.X || y != target.Y; {
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

// HasLineOfSight reports whether p sees target past the blocking indices. Where the line passes
// the corner between two cells it is stopped only when both block, so a diagonal run of
// blockers is opaque and a single one is not.
func (p Point) HasLineOfSight(target geom.Point[int], blocking []geom.Point[int]) bool {
	if len(blocking) == 0 {
		return true
	}

	return p.hasLineOfSight(target, blockers{set: newPointSet(blocking)})
}

// HasLineOfSightFunc reports whether p sees target past the indices blocked reports, asking it
// only about the indices the line passes.
func (p Point) HasLineOfSightFunc(target geom.Point[int], blocked func(geom.Point[int]) bool) bool {
	if blocked == nil {
		return true
	}

	return p.hasLineOfSight(target, blockers{blocked: blocked})
}

// FieldOfView returns the candidates HasLineOfSight sees from p past the blocking indices, in
// the order given, and nil for no candidates. A candidate adjacent to p, across a corner too, is
// always in sight.
func (p Point) FieldOfView(candidates []geom.Point[int], blocking []geom.Point[int]) []geom.Point[int] {
	if len(candidates) == 0 {
		return nil
	}

	return p.AppendFieldOfView(make([]geom.Point[int], 0, len(candidates)), candidates, blocking)
}

// AppendFieldOfView appends the indices of FieldOfView to dst and returns the extended slice.
func (p Point) AppendFieldOfView(dst []geom.Point[int], candidates []geom.Point[int], blocking []geom.Point[int]) []geom.Point[int] {
	p.fieldOfView(candidates, blockers{set: newPointSet(blocking)}, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// FieldOfViewSeq returns an iterator over the indices of FieldOfView, in the same order.
func (p Point) FieldOfViewSeq(candidates []geom.Point[int], blocking []geom.Point[int]) iter.Seq[geom.Point[int]] {
	blocks := blockers{set: newPointSet(blocking)}

	return func(yield func(geom.Point[int]) bool) {
		p.fieldOfView(candidates, blocks, yield)
	}
}

// FieldOfViewFunc returns the candidates HasLineOfSightFunc sees from p past the indices
// blocked reports, in the order given, and nil for no candidates. A candidate adjacent to p,
// across a corner too, is always in sight.
func (p Point) FieldOfViewFunc(candidates []geom.Point[int], blocked func(geom.Point[int]) bool) []geom.Point[int] {
	if len(candidates) == 0 {
		return nil
	}

	return p.AppendFieldOfViewFunc(make([]geom.Point[int], 0, len(candidates)), candidates, blocked)
}

// AppendFieldOfViewFunc appends the indices of FieldOfViewFunc to dst and returns the extended
// slice.
func (p Point) AppendFieldOfViewFunc(dst []geom.Point[int], candidates []geom.Point[int], blocked func(geom.Point[int]) bool) []geom.Point[int] {
	p.fieldOfView(candidates, blockers{blocked: blocked}, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// FieldOfViewFuncSeq returns an iterator over the indices of FieldOfViewFunc, in the same
// order.
func (p Point) FieldOfViewFuncSeq(candidates []geom.Point[int], blocked func(geom.Point[int]) bool) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.fieldOfView(candidates, blockers{blocked: blocked}, yield)
	}
}

// fieldOfView yields the candidates p sees past the blockers and reports whether it yielded
// them all. A candidate adjacent to p is in sight whatever blocks.
func (p Point) fieldOfView(candidates []geom.Point[int], blocks blockers, yield func(geom.Point[int]) bool) bool {
	none := blocks.isEmpty()

	for _, candidate := range candidates {
		if !none && !p.sees(candidate, blocks) {
			continue
		}

		if !yield(candidate) {
			return false
		}
	}

	return true
}

// sees reports whether the candidate is in the field of view of p: adjacent to it, corners
// included, or in line of sight past the blockers.
func (p Point) sees(candidate geom.Point[int], blocks blockers) bool {
	return p.Point().ChebyshevDistanceTo(candidate) <= 1 || p.hasLineOfSight(candidate, blocks)
}

// hasLineOfSight is the sight rule every form shares, asking the blockers about the cells the
// line from p to target passes.
func (p Point) hasLineOfSight(target geom.Point[int], blocks blockers) bool {
	// A Bresenham walk is not mirror-symmetric — reversing the endpoints can trace a
	// different chain of cells — so always walk from the lower endpoint. Neither
	// endpoint is ever tested against the blockers, which makes the swap free of meaning.
	from, to := p.Point(), target
	if to.Compare(from) < 0 {
		from, to = to, from
	}

	x0, y0 := from.XY()
	x1, y1 := to.XY()

	// the walk starts at the lower endpoint, so it never steps back along x
	dx := x1 - x0
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}

	sy := 1
	if y0 > y1 {
		sy = -1
	}

	x, y, err := x0, y0, dx-dy

	for {
		if x == x1 && y == y1 {
			return true
		}

		e2 := 2 * err
		stepX := e2 > -dy
		stepY := e2 < dx

		// crossing a corner: opaque only when both flanking cells block
		if stepX && stepY && blocks.has(geom.Pt(x+1, y)) && blocks.has(geom.Pt(x, y+sy)) {
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
		if blocks.has(geom.Pt(x, y)) {
			return false
		}
	}
}

// Point returns p as the geom.Point it is.
func (p Point) Point() geom.Point[int] {
	return geom.Point[int](p)
}

// blockers answers whether an index blocks sight, from a set of indices or from a function, so
// the slice and the Func forms of the sight methods share one rule and neither pays for the
// other: the set is read without a call.
type blockers struct {
	set     pointSet
	blocked func(geom.Point[int]) bool
}

// has reports whether the index blocks.
func (b blockers) has(index geom.Point[int]) bool {
	if b.blocked != nil {
		return b.blocked(index)
	}

	return b.set.has(index)
}

// isEmpty reports whether nothing can block.
func (b blockers) isEmpty() bool {
	return b.blocked == nil && len(b.set) == 0
}

// pointSet is a set of indices read without a call.
type pointSet map[geom.Point[int]]struct{}

// newPointSet returns the set of the points, and nil, which holds nothing, for none.
func newPointSet(points []geom.Point[int]) pointSet {
	if len(points) == 0 {
		return nil
	}

	set := make(pointSet, len(points))
	for _, point := range points {
		set[point] = struct{}{}
	}

	return set
}

// has reports whether the set contains the point.
func (s pointSet) has(point geom.Point[int]) bool {
	_, found := s[point]

	return found
}
