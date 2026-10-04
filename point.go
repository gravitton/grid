package grid

import (
	geom "github.com/gravitton/geometry"
)

// Point is a grid cell coordinate with spatial query methods (Range, FieldOfView, HasLineOfSight).
type Point geom.Point[int]

// Pt constructs a Point at column x, row y.
func Pt(x, y int) Point {
	return Point(geom.Pt(x, y))
}

// Neighbors returns the neighbor coordinates for every direction in movement.
func (p Point) Neighbors(movement Movement) []geom.Point[int] {
	offsets := movement.Offsets()

	neighbors := make([]geom.Point[int], len(offsets))
	for i, offset := range offsets {
		neighbors[i] = p.Point().Add(offset)
	}

	return neighbors
}

// Range returns every index at most n steps from p under the movement, p included, ordered by
// x and then by y: a diamond for MovementCardinal, a square for MovementDiagonal. It returns
// nil for a negative n.
func (p Point) Range(n int, movement Movement) []geom.Point[int] {
	if n < 0 {
		return nil
	}

	indices := make([]geom.Point[int], 0, movement.rangeLen(n))
	for dx := -n; dx <= n; dx++ {
		for dy := -n; dy <= n; dy++ {
			index := geom.Pt(p.X+dx, p.Y+dy)
			if movement.DistanceTo(p.Point(), index) <= n {
				indices = append(indices, index)
			}
		}
	}

	return indices
}

// DistanceTo returns the grid distance between two points for the given Movement.
// MovementCardinal uses Manhattan distance; MovementDiagonal uses Chebyshev distance.
func (p Point) DistanceTo(to geom.Point[int], movement Movement) int {
	return movement.DistanceTo(p.Point(), to)
}

// HasLineOfSight reports whether there is a clear line of sight from p to target,
// given a set of blocking points. All intermediate cells (excluding p and target)
// must not be in blocking. The target itself may be a blocker — it is
// visible but does not allow sight through it.
//
// Where the line crosses a cell corner it is blocked only when both cells flanking
// that corner block, so a diagonal run of blockers is opaque while sight still
// passes diagonally by a single isolated one.
//
// Visibility is reciprocal: p sees target exactly when target sees p.
func (p Point) HasLineOfSight(target geom.Point[int], blocking []geom.Point[int]) bool {
	return p.hasLineOfSight(target, newPointSet(blocking))
}

// FieldOfView returns the subset of candidates visible from p,
// given a set of blocking points. Adjacent cells (Chebyshev distance ≤ 1)
// are always visible.
func (p Point) FieldOfView(candidates []geom.Point[int], blocking []geom.Point[int]) []geom.Point[int] {
	blocked := newPointSet(blocking)

	results := make([]geom.Point[int], 0, len(candidates))
	for _, candidate := range candidates {
		if len(blocking) == 0 || p.Point().ChebyshevDistanceTo(candidate) <= 1 || p.hasLineOfSight(candidate, blocked) {
			results = append(results, candidate)
		}
	}

	return results
}

func (p Point) hasLineOfSight(target geom.Point[int], blocking pointSet) bool {
	// A Bresenham walk is not mirror-symmetric — reversing the endpoints can trace a
	// different chain of cells — so always walk from the lower endpoint. Neither
	// endpoint is ever tested against blocking, which makes the swap free of meaning.
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
		if stepX && stepY && blocking.has(geom.Pt(x+sx, y)) && blocking.has(geom.Pt(x, y+sy)) {
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
		if blocking.has(geom.Pt(x, y)) {
			return false
		}
	}
}

// Point returns the underlying coordinate as a geom.Point.
func (p Point) Point() geom.Point[int] {
	return geom.Point[int](p)
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
