package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

// Movement is the neighbor connectivity of a square grid.
type Movement int

const (
	// Cardinal connects the four edge neighbors: right, down, left, up.
	Cardinal Movement = iota

	// Diagonal connects all eight neighbors, adding the four corner diagonals.
	Diagonal
)

// Directions returns the neighbor directions of the movement, ordered by increasing angle
// from [geom.DirectionRight].
func (s Movement) Directions() []geom.Direction {
	switch s {
	case Cardinal:
		directions := geom.CardinalDirections()

		return directions[:]
	case Diagonal:
		directions := geom.Directions()

		return directions[:]
	default:
		panic("unsupported movement")
	}
}

// Offsets returns the lattice step of every neighbor direction of the movement,
// in the same order as [Movement.Directions].
func (s Movement) Offsets() []ints.Vector {
	directions := s.Directions()

	offsets := make([]ints.Vector, len(directions))
	for i, direction := range directions {
		offsets[i] = direction.Offset[int]()
	}

	return offsets
}

// within returns every index at most n steps from center under the movement, the center included,
// and nil for a negative n.
func (s Movement) within(center ints.Point, n int) []ints.Point {
	if n < 0 {
		return nil
	}

	indices := make([]ints.Point, 0, s.withinLen(n))
	for dx := -n; dx <= n; dx++ {
		for dy := -n; dy <= n; dy++ {
			index := center.AddXY(dx, dy)
			if s.DistanceTo(center, index) <= n {
				indices = append(indices, index)
			}
		}
	}

	return indices
}

// withinLen returns the number of indices within returns for a radius of n.
func (s Movement) withinLen(n int) int {
	if s == Cardinal {
		return 2*n*(n+1) + 1
	}

	return (2*n + 1) * (2*n + 1)
}

// Offset returns the lattice step of the direction, or the zero vector when the direction
// is not a neighbor direction of the movement.
func (s Movement) Offset(direction geom.Direction) ints.Vector {
	if !s.Has(direction) {
		return ints.Vector{}
	}

	return direction.Offset[int]()
}

// Has reports whether the direction is one of the neighbor directions of the movement.
func (s Movement) Has(direction geom.Direction) bool {
	switch s {
	case Cardinal:
		return direction.IsCardinal()
	case Diagonal:
		return !direction.IsNone()
	default:
		panic("unsupported movement")
	}
}

// DistanceTo returns the grid distance between two indices.
// Cardinal uses Manhattan distance; Diagonal uses Chebyshev distance.
func (s Movement) DistanceTo(from, to ints.Point) int {
	switch s {
	case Cardinal:
		return from.ManhattanDistanceTo(to)
	case Diagonal:
		return from.ChebyshevDistanceTo(to)
	default:
		panic("unsupported movement")
	}
}

// String returns the name of the movement constant.
func (s Movement) String() string {
	switch s {
	case Cardinal:
		return "Cardinal"
	case Diagonal:
		return "Diagonal"
	default:
		return "Unknown"
	}
}
