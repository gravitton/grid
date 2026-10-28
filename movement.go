package grid

import (
	"fmt"

	geom "github.com/gravitton/geometry"
)

// Movement is the neighbor connectivity of a square grid.
type Movement int

const (
	// MovementCardinal steps to the four neighbors across an edge.
	MovementCardinal Movement = iota

	// MovementDiagonal steps to all eight neighbors, across a corner too.
	MovementDiagonal

	// MovementNone is the absence of a movement, with no neighbors. DistanceTo alone has no
	// answer for it and panics.
	MovementNone Movement = -1
)

// Movements returns both movements in order.
func Movements() [2]Movement {
	return [2]Movement{MovementCardinal, MovementDiagonal}
}

// ParseMovement returns the movement with the name String prints, and an error for any other
// string.
func ParseMovement(name string) (Movement, error) {
	if name == MovementNone.String() {
		return MovementNone, nil
	}

	for _, movement := range Movements() {
		if movement.String() == name {
			return movement, nil
		}
	}

	return MovementNone, fmt.Errorf("grid: unknown movement %q", name)
}

// Directions returns the directions of the neighbors of the movement, by increasing angle from
// right, and nil for a movement outside the constants.
func (m Movement) Directions() []geom.Direction {
	switch m {
	case MovementCardinal:
		directions := geom.CardinalDirections()

		return directions[:]
	case MovementDiagonal:
		directions := geom.Directions()

		return directions[:]
	default:
		return nil
	}
}

// Offsets returns the step to each neighbor of the movement, in the order of Directions, and
// nil for a movement outside the constants.
func (m Movement) Offsets() []geom.Vector[int] {
	switch m {
	case MovementCardinal:
		offsets := geom.CardinalOffsets[int]()

		return offsets[:]
	case MovementDiagonal:
		offsets := geom.DirectionOffsets[int]()

		return offsets[:]
	default:
		return nil
	}
}

// Offset returns the step in the direction, and the zero vector for a direction that is not one
// of the movement.
func (m Movement) Offset(direction geom.Direction) geom.Vector[int] {
	if !m.Contains(direction) {
		return geom.Vector[int]{}
	}

	return direction.Offset[int]()
}

// neighborsLen returns the number of neighbors of an index under the movement, the length of
// Directions, Offsets and Point.Neighbors, and zero for a movement outside the constants.
func (m Movement) neighborsLen() int {
	switch m {
	case MovementCardinal:
		return 4
	case MovementDiagonal:
		return 8
	default:
		return 0
	}
}

// DistanceTo returns the number of steps between two indices under the movement: the
// Manhattan distance for MovementCardinal, the Chebyshev distance for MovementDiagonal. It
// panics for a movement outside the constants.
func (m Movement) DistanceTo(from, to geom.Point[int]) int {
	switch m {
	case MovementCardinal:
		return from.ManhattanDistanceTo(to)
	case MovementDiagonal:
		return from.ChebyshevDistanceTo(to)
	default:
		panic(m.unknown())
	}
}

// Contains reports whether the direction is one of Directions.
func (m Movement) Contains(direction geom.Direction) bool {
	switch m {
	case MovementCardinal:
		return direction.IsCardinal()
	case MovementDiagonal:
		return !direction.IsNone()
	default:
		return false
	}
}

// IsNone reports whether the movement is outside the constants, MovementNone or any other
// value.
func (m Movement) IsNone() bool {
	return m != MovementCardinal && m != MovementDiagonal
}

// unknown returns the error a method with no answer for a movement outside the constants panics
// with.
func (m Movement) unknown() error {
	return fmt.Errorf("grid: unknown movement %d", m)
}

// String returns the name of the movement, "None" for one outside the constants.
func (m Movement) String() string {
	switch m {
	case MovementCardinal:
		return "Cardinal"
	case MovementDiagonal:
		return "Diagonal"
	default:
		return "None"
	}
}

// MarshalText implements encoding.TextMarshaler with the name String prints.
func (m Movement) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler through ParseMovement.
func (m *Movement) UnmarshalText(text []byte) error {
	movement, err := ParseMovement(string(text))
	if err != nil {
		return err
	}

	*m = movement

	return nil
}
