package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

// Point is a grid cell coordinate with spatial query methods (Range, FieldOfView, HasLineOfSight).
type Point ints.Point

// Pt constructs a Point at column x, row y.
func Pt(x, y int) Point {
	return Point(ints.Pt(x, y))
}

// DistanceTo returns the grid distance between two points for the given System.
// Cardinal uses Manhattan distance; Diagonal uses Chebyshev distance.
func (p Point) DistanceTo(to ints.Point, system System) int {
	return system.DistanceTo(p.Point(), to)
}

// Point returns the underlying coordinate as an ints.Point.
func (p Point) Point() ints.Point { return ints.Point(p) }

// Neighbors returns the neighbor coordinates for every direction in system.
func (p Point) Neighbors(system System) []ints.Point {
	offsets := system.Offsets()

	neighbors := make([]ints.Point, len(offsets))
	for i, offset := range offsets {
		neighbors[i] = p.Point().Add(offset)
	}

	return neighbors
}

// Range returns all cells within Euclidean distance n from p, inclusive.
func (p Point) Range(n int) []ints.Point {
	if n < 0 {
		return nil
	}

	results := make([]ints.Point, 0, (2*n+1)*(2*n+1))
	r2 := n * n

	for dx := -n; dx <= n; dx++ {
		for dy := -n; dy <= n; dy++ {
			if dx*dx+dy*dy <= r2 {
				results = append(results, geom.Pt(p.X+dx, p.Y+dy))
			}
		}
	}

	return results
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
func (p Point) HasLineOfSight(target ints.Point, blocking []ints.Point) bool {
	return p.hasLineOfSight(target, newPointSet(blocking))
}

func (p Point) hasLineOfSight(target ints.Point, blocking pointSet) bool {
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

// FieldOfView returns the subset of candidates visible from p,
// given a set of blocking points. Adjacent cells (Chebyshev distance ≤ 1)
// are always visible.
func (p Point) FieldOfView(candidates []ints.Point, blocking []ints.Point) []ints.Point {
	blocked := newPointSet(blocking)

	results := make([]ints.Point, 0, len(candidates))
	for _, candidate := range candidates {
		if len(blocking) == 0 || p.Point().ChebyshevDistanceTo(candidate) <= 1 || p.hasLineOfSight(candidate, blocked) {
			results = append(results, candidate)
		}
	}

	return results
}

// pointSet indexes grid coordinates for constant-time membership tests.
type pointSet map[ints.Point]struct{}

// newPointSet indexes the given points. It returns nil for an empty input;
// reads from a nil set are valid and never match.
func newPointSet(points []ints.Point) pointSet {
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
func (s pointSet) has(point ints.Point) bool {
	_, found := s[point]

	return found
}
