package grid

import (
	"fmt"

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

// NeighborOffset returns the unit vector for direction in the given system.
// For Cardinal, only even Direction values (E, N, W, S) are valid — passing a
// diagonal direction returns a wrong vector or panics with index out of range.
func NeighborOffset(system System, direction Direction) ints.Vector {
	if system == Cardinal && direction%2 != 0 {
		panic("diagonal direction not available in Cardinal movement system")
	}
	return Directions[direction]
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
