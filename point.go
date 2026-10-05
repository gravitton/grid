package grid

import (
	"iter"
	"math"

	geom "github.com/gravitton/geometry"
)

// Point is an index of a square grid, with the lattice queries [geom.Point] does not have:
// the neighbors, ranges, rings and lines around it and the sight from it. It is to a square
// grid what hex.Hex is to a hexagonal one. A method whose answer depends on which cells are
// adjacent takes the [Movement], and panics for one outside the constants.
type Point geom.Point[int]

// Pt returns the Point at column x and row y.
func Pt(x, y int) Point {
	return Point{X: x, Y: y}
}

// RangeLen returns the number of indices Point.Range and Point.Spiral return for n steps under
// the movement, to size a buffer with, and zero for a negative n.
func RangeLen(n int, movement Movement) int {
	count := movement.neighborCount()
	if n < 0 {
		return 0
	}

	return 1 + count*n*(n+1)/2
}

// RingLen returns the number of indices Point.Ring returns for n steps under the movement, to
// size a buffer with: zero for a negative n and one for a zero n.
func RingLen(n int, movement Movement) int {
	count := movement.neighborCount()

	switch {
	case n < 0:
		return 0
	case n == 0:
		return 1
	default:
		return count * n
	}
}

// Turn returns the point turned around the origin by steps quarter turns of increasing angle,
// the sense of geom.Direction.Turn: clockwise as drawn on a screen with Y pointing down.
// Negative steps turn the other way.
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

// TurnAround returns the point turned around center by steps quarter turns, as Turn turns it
// around the origin.
func (p Point) TurnAround(center geom.Point[int], steps int) Point {
	turned := Point{X: p.X - center.X, Y: p.Y - center.Y}.Turn(steps)

	return Point{X: turned.X + center.X, Y: turned.Y + center.Y}
}

// ReflectX returns the point mirrored across the x-axis, its row negated.
func (p Point) ReflectX() Point {
	return Point{X: p.X, Y: -p.Y}
}

// ReflectY returns the point mirrored across the y-axis, its column negated.
func (p Point) ReflectY() Point {
	return Point{X: -p.X, Y: p.Y}
}

// Neighbor returns the index one step from p in the direction, and p itself for
// geom.DirectionNone.
func (p Point) Neighbor(direction geom.Direction) geom.Point[int] {
	return p.Point().Add(direction.Offset[int]())
}

// Neighbors returns the indices adjacent to p under the movement, by increasing angle from
// geom.DirectionRight.
func (p Point) Neighbors(movement Movement) []geom.Point[int] {
	return p.AppendNeighbors(make([]geom.Point[int], 0, movement.neighborCount()), movement)
}

// AppendNeighbors appends the indices Neighbors returns to dst and returns the extended slice.
func (p Point) AppendNeighbors(dst []geom.Point[int], movement Movement) []geom.Point[int] {
	count := movement.neighborCount()
	directions := geom.Directions()

	for i := 0; i < len(directions); i += len(directions) / count {
		dst = append(dst, p.Neighbor(directions[i]))
	}

	return dst
}

// Range returns every index at most n steps from p under the movement, p included, ordered by
// x and then by y: a diamond for MovementCardinal, a square for MovementDiagonal. It returns
// nil for a negative n.
func (p Point) Range(n int, movement Movement) []geom.Point[int] {
	size := RangeLen(n, movement)
	if n < 0 {
		return nil
	}

	return p.AppendRange(make([]geom.Point[int], 0, size), n, movement)
}

// AppendRange appends the indices Range returns to dst and returns the extended slice.
func (p Point) AppendRange(dst []geom.Point[int], n int, movement Movement) []geom.Point[int] {
	p.within(n, movement.neighborCount() == 4, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// RangeSeq returns an iterator over the indices Range returns, in its order.
func (p Point) RangeSeq(n int, movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.within(n, movement.neighborCount() == 4, yield)
	}
}

// within yields the indices at most n steps from p, column by column, each column as far up
// and down as the movement reaches, and reports whether it yielded them all.
func (p Point) within(n int, cardinal bool, yield func(geom.Point[int]) bool) bool {
	for dx := -n; dx <= n; dx++ {
		reach := n
		if cardinal {
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

// Ring returns the indices exactly n steps from p under the movement, by increasing angle from
// the one n steps to the right, so Ring(1) is Neighbors. It returns nil for a negative n and p
// alone for a zero n.
func (p Point) Ring(n int, movement Movement) []geom.Point[int] {
	size := RingLen(n, movement)
	if n < 0 {
		return nil
	}

	return p.AppendRing(make([]geom.Point[int], 0, size), n, movement)
}

// AppendRing appends the indices Ring returns to dst and returns the extended slice.
func (p Point) AppendRing(dst []geom.Point[int], n int, movement Movement) []geom.Point[int] {
	p.ring(n, movement.neighborCount(), func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// RingSeq returns an iterator over the indices Ring returns, in its order.
func (p Point) RingSeq(n int, movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.ring(n, movement.neighborCount(), yield)
	}
}

// ring yields the indices n steps from p for a movement of count neighbors, walking from each
// neighbor direction scaled by n toward the next, and reports whether it yielded them all.
func (p Point) ring(n int, count int, yield func(geom.Point[int]) bool) bool {
	if n < 0 {
		return true
	}
	if n == 0 {
		return yield(p.Point())
	}

	directions := geom.Directions()
	stride := len(directions) / count

	for i := 0; i < len(directions); i += stride {
		corner := directions[i].Offset[int]()
		next := directions[(i+stride)%len(directions)].Offset[int]()

		for j := range n {
			if !yield(geom.Pt(p.X+corner.X*n+(next.X-corner.X)*j, p.Y+corner.Y*n+(next.Y-corner.Y)*j)) {
				return false
			}
		}
	}

	return true
}

// Spiral returns the indices Range returns, ordered outward ring by ring from p, each ring as
// Ring orders it. It returns nil for a negative n.
func (p Point) Spiral(n int, movement Movement) []geom.Point[int] {
	size := RangeLen(n, movement)
	if n < 0 {
		return nil
	}

	return p.AppendSpiral(make([]geom.Point[int], 0, size), n, movement)
}

// AppendSpiral appends the indices Spiral returns to dst and returns the extended slice.
func (p Point) AppendSpiral(dst []geom.Point[int], n int, movement Movement) []geom.Point[int] {
	p.spiral(n, movement.neighborCount(), func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// SpiralSeq returns an iterator over the indices Spiral returns, in its order, so a caller
// looking for the nearest match stops at it.
func (p Point) SpiralSeq(n int, movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.spiral(n, movement.neighborCount(), yield)
	}
}

// spiral yields the rings of radius zero to n in turn and reports whether it yielded them all.
func (p Point) spiral(n int, count int, yield func(geom.Point[int]) bool) bool {
	for radius := 0; radius <= n; radius++ {
		if !p.ring(radius, count, yield) {
			return false
		}
	}

	return true
}

// DistanceTo returns the number of steps from p to the index under the movement: the Manhattan
// distance for MovementCardinal, the Chebyshev distance for MovementDiagonal.
func (p Point) DistanceTo(to geom.Point[int], movement Movement) int {
	return movement.DistanceTo(p.Point(), to)
}

// DirectionTo returns the direction of the movement nearest to the one from p to target, and
// geom.DirectionNone for p itself. Under MovementCardinal a target on a diagonal answers the
// horizontal direction.
func (p Point) DirectionTo(target geom.Point[int], movement Movement) geom.Direction {
	cardinal := movement.neighborCount() == 4
	dx, dy := target.X-p.X, target.Y-p.Y

	switch {
	case dx == 0 && dy == 0:
		return geom.DirectionNone
	case !cardinal:
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

// Line returns the indices from p to target, both included, each a neighbor of the one before
// it under the movement, so its length is the distance plus one. It is the traversal, not the
// sight rule: HasLineOfSight decides what a blocker hides.
func (p Point) Line(target geom.Point[int], movement Movement) []geom.Point[int] {
	return p.AppendLine(make([]geom.Point[int], 0, p.DistanceTo(target, movement)+1), target, movement)
}

// AppendLine appends the indices Line returns to dst and returns the extended slice.
func (p Point) AppendLine(dst []geom.Point[int], target geom.Point[int], movement Movement) []geom.Point[int] {
	p.line(target, movement.neighborCount() == 4, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// LineSeq returns an iterator over the indices Line returns, in its order.
func (p Point) LineSeq(target geom.Point[int], movement Movement) iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		p.line(target, movement.neighborCount() == 4, yield)
	}
}

// line yields the indices from p to target and reports whether it yielded them all. Under a
// cardinal movement every step goes along one axis, the one further behind the line, i and j
// counting the steps taken along each; under a diagonal one it is the Bresenham walk.
func (p Point) line(target geom.Point[int], cardinal bool, yield func(geom.Point[int]) bool) bool {
	x, y := p.X, p.Y
	dx, dy := geom.Abs(target.X-x), geom.Abs(target.Y-y)
	sx, sy := geom.Sign(target.X-x), geom.Sign(target.Y-y)

	if !yield(geom.Pt(x, y)) {
		return false
	}

	if cardinal {
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

// HasLineOfSight reports whether p sees target past the blocking indices. The line between the
// two is walked from the lower of them, so the answer is the same from either end, and neither
// end blocks. Where the line passes the corner between two cells it is stopped only when both
// block, so a diagonal run of blockers is opaque and a single one is not.
func (p Point) HasLineOfSight(target geom.Point[int], blocking []geom.Point[int]) bool {
	if len(blocking) == 0 {
		return true
	}

	return p.hasLineOfSight(target, blockers{set: newPointSet(blocking)})
}

// HasLineOfSightFunc is HasLineOfSight asking blocked about each index the line passes, for a
// caller whose blockers are a lookup in a map or a grid. A nil blocked blocks nothing.
func (p Point) HasLineOfSightFunc(target geom.Point[int], blocked func(geom.Point[int]) bool) bool {
	if blocked == nil {
		return true
	}

	return p.hasLineOfSight(target, blockers{blocked: blocked})
}

// FieldOfView returns the candidates p has line of sight to past the blocking indices, in the
// order given. The indices adjacent to p, corners included, are always in sight.
func (p Point) FieldOfView(candidates []geom.Point[int], blocking []geom.Point[int]) []geom.Point[int] {
	return p.AppendFieldOfView(make([]geom.Point[int], 0, len(candidates)), candidates, blocking)
}

// AppendFieldOfView appends the indices FieldOfView returns to dst and returns the extended
// slice.
func (p Point) AppendFieldOfView(dst []geom.Point[int], candidates []geom.Point[int], blocking []geom.Point[int]) []geom.Point[int] {
	p.fieldOfView(candidates, blockers{set: newPointSet(blocking)}, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// FieldOfViewSeq returns an iterator over the indices FieldOfView returns, in its order.
func (p Point) FieldOfViewSeq(candidates []geom.Point[int], blocking []geom.Point[int]) iter.Seq[geom.Point[int]] {
	blocks := blockers{set: newPointSet(blocking)}

	return func(yield func(geom.Point[int]) bool) {
		p.fieldOfView(candidates, blocks, yield)
	}
}

// FieldOfViewFunc is FieldOfView asking blocked about each index a line passes. A nil blocked
// blocks nothing.
func (p Point) FieldOfViewFunc(candidates []geom.Point[int], blocked func(geom.Point[int]) bool) []geom.Point[int] {
	return p.AppendFieldOfViewFunc(make([]geom.Point[int], 0, len(candidates)), candidates, blocked)
}

// AppendFieldOfViewFunc appends the indices FieldOfViewFunc returns to dst and returns the
// extended slice.
func (p Point) AppendFieldOfViewFunc(dst []geom.Point[int], candidates []geom.Point[int], blocked func(geom.Point[int]) bool) []geom.Point[int] {
	p.fieldOfView(candidates, blockers{blocked: blocked}, func(index geom.Point[int]) bool {
		dst = append(dst, index)

		return true
	})

	return dst
}

// FieldOfViewFuncSeq returns an iterator over the indices FieldOfViewFunc returns, in its
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
		if !none && p.Point().ChebyshevDistanceTo(candidate) > 1 && !p.hasLineOfSight(candidate, blocks) {
			continue
		}

		if !yield(candidate) {
			return false
		}
	}

	return true
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

	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}

	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
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
		if stepX && stepY && blocks.has(geom.Pt(x+sx, y)) && blocks.has(geom.Pt(x, y+sy)) {
			return false
		}

		if stepX {
			err -= dy
			x += sx
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

// Point returns the index as the geom.Point it is.
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

// pointSet indexes grid coordinates for constant-time membership tests.
type pointSet map[geom.Point[int]]struct{}

// newPointSet indexes the given points. It returns nil for an empty input;
// reads from a nil set are valid and never match.
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
