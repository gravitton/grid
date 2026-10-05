package grid

import (
	"fmt"

	geom "github.com/gravitton/geometry"
)

// Movement is the neighbor connectivity of a square grid.
type Movement int

const (
	// MovementCardinal connects the four edge neighbors: right, down, left, up.
	MovementCardinal Movement = iota

	// MovementDiagonal connects all eight neighbors, adding the four corner diagonals.
	MovementDiagonal

	// MovementNone is the absence of a movement. It names no neighbors, so Directions, Has and
	// DistanceTo have nothing to answer for it and panic.
	MovementNone Movement = -1
)

// Movements lists both movements in order. It returns a fresh array, so a caller cannot alter
// the list.
func Movements() [2]Movement {
	return [2]Movement{MovementCardinal, MovementDiagonal}
}

// ParseMovement returns the movement with the given name, as String prints it, and an error
// for any other string. "None" parses to MovementNone.
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

// Directions returns the neighbor directions of the movement, ordered by increasing angle
// from [geom.DirectionRight]. It panics for a movement outside the constants, as Has and
// DistanceTo do.
func (m Movement) Directions() []geom.Direction {
	switch m {
	case MovementCardinal:
		directions := geom.CardinalDirections()

		return directions[:]
	case MovementDiagonal:
		directions := geom.Directions()

		return directions[:]
	default:
		panic(fmt.Sprintf("grid: unknown movement %d", m))
	}
}

// Offsets returns the lattice step of every neighbor direction of the movement,
// in the same order as [Movement.Directions].
func (m Movement) Offsets() []geom.Vector[int] {
	directions := m.Directions()

	offsets := make([]geom.Vector[int], len(directions))
	for i, direction := range directions {
		offsets[i] = direction.Offset[int]()
	}

	return offsets
}

// Offset returns the lattice step of the direction, or the zero vector when the direction
// is not a neighbor direction of the movement.
func (m Movement) Offset(direction geom.Direction) geom.Vector[int] {
	if !m.Contains(direction) {
		return geom.Vector[int]{}
	}

	return direction.Offset[int]()
}

// neighborCount returns the number of neighbors of an index under the movement, and panics for
// a movement outside the constants, as Directions does.
func (m Movement) neighborCount() int {
	switch m {
	case MovementCardinal:
		return 4
	case MovementDiagonal:
		return 8
	default:
		panic(fmt.Sprintf("grid: unknown movement %d", m))
	}
}

// DistanceTo returns the grid distance between two indices.
// MovementCardinal uses Manhattan distance; MovementDiagonal uses Chebyshev distance.
func (m Movement) DistanceTo(from, to geom.Point[int]) int {
	switch m {
	case MovementCardinal:
		return from.ManhattanDistanceTo(to)
	case MovementDiagonal:
		return from.ChebyshevDistanceTo(to)
	default:
		panic(fmt.Sprintf("grid: unknown movement %d", m))
	}
}

// Contains reports whether the direction is one of the neighbor directions of the movement.
func (m Movement) Contains(direction geom.Direction) bool {
	switch m {
	case MovementCardinal:
		return direction.IsCardinal()
	case MovementDiagonal:
		return !direction.IsNone()
	default:
		panic(fmt.Sprintf("grid: unknown movement %d", m))
	}
}

// IsNone reports whether the movement is neither MovementCardinal nor MovementDiagonal. A
// Movement outside the constants is not normalized, so every such value counts as
// MovementNone.
func (m Movement) IsNone() bool {
	return m != MovementCardinal && m != MovementDiagonal
}

// String returns the name of the movement constant.
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

// MarshalText implements encoding.TextMarshaler with the name String prints, so a movement is
// stored as "Cardinal" in JSON and as a map key rather than as its number.
func (m Movement) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, the inverse of MarshalText through
// ParseMovement.
func (m *Movement) UnmarshalText(text []byte) error {
	movement, err := ParseMovement(string(text))
	if err != nil {
		return err
	}

	*m = movement

	return nil
}
