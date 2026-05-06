package grid

import (
	"slices"

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
	return DistanceTo(ints.Point(p), to, system)
}

// Point returns the underlying coordinate as an ints.Point.
func (p Point) Point() ints.Point { return ints.Point(p) }

// Neighbors returns the neighbor coordinates for every direction in system.
func (p Point) Neighbors(system System) []ints.Point {
	directions := NeighborOffsets(system)
	neighbors := make([]ints.Point, len(directions))
	for i, v := range directions {
		neighbors[i] = ints.Pt(p.X+v.X, p.Y+v.Y)
	}

	return neighbors
}

// Range returns all cells within Euclidean distance n from s, inclusive.
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

// HasLineOfSight reports whether there is a clear line of sight from src to dst,
// given a set of blocking points. All intermediate cells (excluding src and dst)
// must not be in blocking. The destination itself may be a blocker — it is
// visible but does not allow sight through it.
func (p Point) HasLineOfSight(target ints.Point, blocking []ints.Point) bool {
	x0, y0 := ints.Point(p).XY()
	x1, y1 := target.XY()

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
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}

		if x == x1 && y == y1 {
			return true
		}
		if slices.Contains(blocking, geom.Pt(x, y)) {
			return false
		}
	}
}

// FieldOfView returns the subset of candidates visible from center,
// given a set of blocking points. Adjacent cells (Chebyshev distance ≤ 1)
// are always visible.
func (p Point) FieldOfView(candidates []ints.Point, blocking []ints.Point) []ints.Point {
	results := make([]ints.Point, 0, len(candidates))
	for _, candidate := range candidates {
		if len(blocking) == 0 || ints.Point(p).ChebyshevDistanceTo(candidate) <= 1 || p.HasLineOfSight(candidate, blocking) {
			results = append(results, candidate)
		}
	}
	return results
}
