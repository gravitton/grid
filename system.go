package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

// System is the neighbor connectivity of a square grid.
type System int

const (
	// Cardinal connects the four edge neighbors: right, down, left, up.
	Cardinal System = iota

	// Diagonal connects all eight neighbors, adding the four corner diagonals.
	Diagonal
)

// Directions returns the neighbor directions of the system, ordered by increasing angle
// from [geom.DirectionRight].
func (s System) Directions() []geom.Direction {
	switch s {
	case Cardinal:
		directions := geom.CardinalDirections

		return directions[:]
	case Diagonal:
		directions := geom.Directions

		return directions[:]
	default:
		panic("unsupported system")
	}
}

// Offsets returns the lattice step of every neighbor direction of the system,
// in the same order as [System.Directions].
func (s System) Offsets() []ints.Vector {
	directions := s.Directions()

	offsets := make([]ints.Vector, len(directions))
	for i, direction := range directions {
		offsets[i] = direction.Offset[int]()
	}

	return offsets
}

// Offset returns the lattice step of the direction, or the zero vector when the direction
// is not a neighbor direction of the system.
func (s System) Offset(direction geom.Direction) ints.Vector {
	if !s.Has(direction) {
		return ints.Vector{}
	}

	return direction.Offset[int]()
}

// Has reports whether the direction is one of the neighbor directions of the system.
func (s System) Has(direction geom.Direction) bool {
	switch s {
	case Cardinal:
		return direction.IsCardinal()
	case Diagonal:
		return !direction.IsNone()
	default:
		panic("unsupported system")
	}
}

// DistanceTo returns the grid distance between two indices.
// Cardinal uses Manhattan distance; Diagonal uses Chebyshev distance.
func (s System) DistanceTo(from, to ints.Point) int {
	switch s {
	case Cardinal:
		return from.ManhattanDistanceTo(to)
	case Diagonal:
		return from.ChebyshevDistanceTo(to)
	default:
		panic("unsupported system")
	}
}

// String returns the name of the system constant.
func (s System) String() string {
	switch s {
	case Cardinal:
		return "Cardinal"
	case Diagonal:
		return "Diagonal"
	default:
		return "Unknown"
	}
}
