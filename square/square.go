package square

import (
	"fmt"
	"slices"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

// System lists the supported neighbor connectivity modes for a square grid.
type System int

const (
	// Cardinal uses 4-directional movement: N, E, S, W.
	Cardinal System = iota

	// Diagonal uses 8-directional movement: N, NE, E, SE, S, SW, W, NW.
	Diagonal
)

// Direction represents one of the eight neighbor directions around a square cell.
// Directions are ordered counterclockwise starting from East (right), matching
// the convention used in the hexagon package.
type Direction int

const (
	E  Direction = iota // (1,  0)
	NE                  // (1, -1)
	N                   // (0, -1)
	NW                  // (-1,-1)
	W                   // (-1, 0)
	SW                  // (-1, 1)
	S                   // (0,  1)
	SE                  // (1,  1)
)

// Opposite returns the direction directly opposite to d (rotated 180°).
func (d Direction) Opposite() Direction {
	return (d + 4) % 8
}

// String returns the name of the direction constant.
func (d Direction) String() string {
	switch d % 8 {
	case E:
		return "E"
	case NE:
		return "NE"
	case N:
		return "N"
	case NW:
		return "NW"
	case W:
		return "W"
	case SW:
		return "SW"
	case S:
		return "S"
	case SE:
		return "SE"
	default:
		return fmt.Sprintf("Direction(%d)", int(d))
	}
}

// CardinalDirections lists the 4 cardinal direction vectors: E, N, W, S.
var CardinalDirections = [4]ints.Vector{
	geom.Vec(1, 0),  // E
	geom.Vec(0, -1), // N
	geom.Vec(-1, 0), // W
	geom.Vec(0, 1),  // S
}

// DiagonalDirections lists the 4 diagonal direction vectors: NE, NW, SW, SE.
var DiagonalDirections = [4]ints.Vector{
	geom.Vec(1, -1),  // NE
	geom.Vec(-1, -1), // NW
	geom.Vec(-1, 1),  // SW
	geom.Vec(1, 1),   // SE
}

// Directions lists all 8 direction vectors ordered counterclockwise from E.
var Directions = [8]ints.Vector{
	geom.Vec(1, 0),   // E
	geom.Vec(1, -1),  // NE
	geom.Vec(0, -1),  // N
	geom.Vec(-1, -1), // NW
	geom.Vec(-1, 0),  // W
	geom.Vec(-1, 1),  // SW
	geom.Vec(0, 1),   // S
	geom.Vec(1, 1),   // SE
}

// NeighborOffsets returns neighbor direction vectors for the given System.
// Cardinal returns 4 vectors (E, N, W, S); Diagonal returns all 8.
func NeighborOffsets(system System) []ints.Vector {
	switch system {
	case Cardinal:
		return CardinalDirections[:]
	case Diagonal:
		return Directions[:]
	default:
		panic("unsupported system")
	}
}

// DistanceTo returns the grid distance between two points for the given System.
// Cardinal uses Manhattan distance; Diagonal uses Chebyshev distance.
func DistanceTo(from, to ints.Point, system System) int {
	switch system {
	case Cardinal:
		return from.ManhattanDistanceTo(to)
	case Diagonal:
		return from.ChebyshevDistanceTo(to)
	default:
		panic("unsupported system")
	}
}

// Range returns all points within Euclidean distance n from center, inclusive.
// Returns nil when n < 0.
func Range(center ints.Point, n int) []ints.Point {
	if n < 0 {
		return nil
	}

	results := make([]ints.Point, 0, (2*n+1)*(2*n+1))
	r2 := n * n

	for dx := -n; dx <= n; dx++ {
		for dy := -n; dy <= n; dy++ {
			if dx*dx+dy*dy <= r2 {
				results = append(results, geom.Pt(center.X+dx, center.Y+dy))
			}
		}
	}

	return results
}

// HasLineOfSight reports whether there is a clear line of sight from src to dst,
// given a set of blocking points. All intermediate cells (excluding src and dst)
// must not be in blocking. The destination itself may be a blocker — it is
// visible but does not allow sight through it.
func HasLineOfSight(src, dst ints.Point, blocking []ints.Point) bool {
	x0, y0 := src.X, src.Y
	x1, y1 := dst.X, dst.Y

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
func FieldOfView(center ints.Point, candidates []ints.Point, blocking []ints.Point) []ints.Point {
	results := make([]ints.Point, 0, len(candidates))
	for _, candidate := range candidates {
		if len(blocking) == 0 || center.ChebyshevDistanceTo(candidate) <= 1 || HasLineOfSight(center, candidate, blocking) {
			results = append(results, candidate)
		}
	}
	return results
}
