package grid

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

var (
	sinkDirections []geom.Direction
	sinkVectors    []geom.Vector[int]
)

func TestMovements(t *testing.T) {
	t.Run("lists both in order", func(t *testing.T) {
		assert.Equal(t, Movements(), [2]Movement{MovementCardinal, MovementDiagonal})
	})
	t.Run("returns a fresh array", func(t *testing.T) {
		movements := Movements()
		movements[0] = MovementNone

		assert.Equal(t, Movements()[0], MovementCardinal)
	})
}

func TestParseMovement(t *testing.T) {
	t.Run("every name round-trips", func(t *testing.T) {
		for _, movement := range Movements() {
			parsed, err := ParseMovement(movement.String())

			assert.NoError(t, err, movement.String()+": ")
			assert.Equal(t, parsed, movement, movement.String()+": ")
		}
	})
	t.Run("None parses to MovementNone", func(t *testing.T) {
		parsed, err := ParseMovement("None")

		assert.NoError(t, err)
		assert.Equal(t, parsed, MovementNone)
	})
	t.Run("unknown is an error", func(t *testing.T) {
		parsed, err := ParseMovement("Knight")

		assert.ErrorContains(t, err, `grid: unknown movement "Knight"`)
		assert.Equal(t, parsed, MovementNone)
	})
}

func TestMovement_Directions(t *testing.T) {
	t.Run("cardinal is the four edge directions by increasing angle", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Directions(), []geom.Direction{
			geom.DirectionRight, geom.DirectionDown, geom.DirectionLeft, geom.DirectionUp,
		})
	})
	t.Run("diagonal is all eight by increasing angle", func(t *testing.T) {
		assert.Equal(t, MovementDiagonal.Directions(), []geom.Direction{
			geom.DirectionRight, geom.DirectionDownRight, geom.DirectionDown, geom.DirectionDownLeft,
			geom.DirectionLeft, geom.DirectionUpLeft, geom.DirectionUp, geom.DirectionUpRight,
		})
	})
	t.Run("returns a fresh slice", func(t *testing.T) {
		MovementDiagonal.Directions()[0] = geom.DirectionNone

		assert.Equal(t, MovementDiagonal.Directions()[0], geom.DirectionRight)
		assert.Equal(t, geom.Directions()[0], geom.DirectionRight)
	})
	t.Run("allocates once at its length", func(t *testing.T) {
		for _, movement := range Movements() {
			directions := movement.Directions()

			assert.Equal(t, cap(directions), len(directions), movement.String()+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkDirections = movement.Directions()
			}), 1.0, movement.String()+": ")
		}
	})
	t.Run("is nil outside the constants", func(t *testing.T) {
		assert.Nil(t, MovementNone.Directions())
		assert.Nil(t, Movement(99).Directions())
	})
}

func TestMovement_Offsets(t *testing.T) {
	t.Run("cardinal is the four edge steps by increasing angle", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Offsets(), []geom.Vector[int]{
			geom.Vec(1, 0), geom.Vec(0, 1), geom.Vec(-1, 0), geom.Vec(0, -1),
		})
	})
	t.Run("diagonal is all eight by increasing angle", func(t *testing.T) {
		assert.Equal(t, MovementDiagonal.Offsets(), []geom.Vector[int]{
			geom.Vec(1, 0), geom.Vec(1, 1), geom.Vec(0, 1), geom.Vec(-1, 1),
			geom.Vec(-1, 0), geom.Vec(-1, -1), geom.Vec(0, -1), geom.Vec(1, -1),
		})
	})
	t.Run("is the offset of every direction, in order", func(t *testing.T) {
		for _, movement := range Movements() {
			directions := movement.Directions()
			offsets := movement.Offsets()

			assert.Equal(t, len(offsets), len(directions), movement.String()+": ")
			for i, direction := range directions {
				assert.Equal(t, offsets[i], direction.Offset[int](), movement.String()+" "+direction.String()+": ")
			}
		}
	})
	t.Run("returns a fresh slice", func(t *testing.T) {
		MovementDiagonal.Offsets()[0] = geom.Vec(9, 9)

		assert.Equal(t, MovementDiagonal.Offsets()[0], geom.Vec(1, 0))
	})
	t.Run("allocates once at its length", func(t *testing.T) {
		for _, movement := range Movements() {
			offsets := movement.Offsets()

			assert.Equal(t, cap(offsets), len(offsets), movement.String()+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkVectors = movement.Offsets()
			}), 1.0, movement.String()+": ")
		}
	})
	t.Run("is nil outside the constants", func(t *testing.T) {
		assert.Nil(t, MovementNone.Offsets())
		assert.Nil(t, Movement(99).Offsets())
	})
}

func TestMovement_Offset(t *testing.T) {
	t.Run("is the lattice step of a direction of the movement", func(t *testing.T) {
		for _, movement := range Movements() {
			for _, direction := range movement.Directions() {
				assert.Equal(t, movement.Offset(direction), direction.Offset[int](), movement.String()+" "+direction.String()+": ")
			}
		}
	})
	t.Run("is the zero vector for a direction outside the movement", func(t *testing.T) {
		for _, direction := range geom.DiagonalDirections() {
			assert.Equal(t, MovementCardinal.Offset(direction), geom.Vec(0, 0), direction.String()+": ")
		}
		for _, movement := range Movements() {
			assert.Equal(t, movement.Offset(geom.DirectionNone), geom.Vec(0, 0), movement.String()+": ")
		}
	})
	t.Run("wraps a direction out of range, as geom does", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Offset(geom.DirectionRight+8), geom.Vec(1, 0))
		assert.Equal(t, MovementDiagonal.Offset(geom.DirectionDownRight-8), geom.Vec(1, 1))
	})
	t.Run("is the zero vector outside the constants", func(t *testing.T) {
		for _, direction := range geom.Directions() {
			assert.Equal(t, MovementNone.Offset(direction), geom.Vec(0, 0), direction.String()+": ")
			assert.Equal(t, Movement(99).Offset(direction), geom.Vec(0, 0), direction.String()+": ")
		}
	})
}

func TestMovement_DistanceTo(t *testing.T) {
	origin := geom.Pt(0, 0)
	targets := []geom.Point[int]{geom.Pt(3, 4), geom.Pt(-2, 3), geom.Pt(2, 2), geom.Pt(0, -5), geom.Pt(0, 0)}

	t.Run("cardinal is the Manhattan distance", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.DistanceTo(origin, geom.Pt(3, 4)), 7)
		assert.Equal(t, MovementCardinal.DistanceTo(origin, geom.Pt(-2, 3)), 5)
	})
	t.Run("diagonal is the Chebyshev distance", func(t *testing.T) {
		assert.Equal(t, MovementDiagonal.DistanceTo(origin, geom.Pt(3, 4)), 4)
		assert.Equal(t, MovementDiagonal.DistanceTo(origin, geom.Pt(2, 2)), 2)
	})
	t.Run("is symmetric", func(t *testing.T) {
		for _, movement := range Movements() {
			for _, target := range targets {
				assert.Equal(t, movement.DistanceTo(target, testPoint.Point()), movement.DistanceTo(testPoint.Point(), target), fmt.Sprintf("%s %v: ", movement, target))
			}
		}
	})
	t.Run("an index is at distance zero from itself alone", func(t *testing.T) {
		for _, movement := range Movements() {
			for _, target := range targets {
				assert.Equal(t, movement.DistanceTo(origin, target) == 0, target == origin, fmt.Sprintf("%s %v: ", movement, target))
			}
		}
	})
	t.Run("every offset of the movement is one step away", func(t *testing.T) {
		for _, movement := range Movements() {
			for _, offset := range movement.Offsets() {
				assert.Equal(t, movement.DistanceTo(testPoint.Point(), testPoint.Point().Add(offset)), 1, fmt.Sprintf("%s %v: ", movement, offset))
			}
		}
	})
	t.Run("a diagonal step is two cardinal ones", func(t *testing.T) {
		for _, direction := range geom.DiagonalDirections() {
			assert.Equal(t, MovementCardinal.DistanceTo(origin, origin.Add(direction.Offset[int]())), 2, direction.String()+": ")
		}
	})
	t.Run("panics outside the constants, where no index is reached from another", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Movement(99).DistanceTo(geom.Pt(0, 0), geom.Pt(1, 1))
		}, "grid: unknown movement 99")
		assert.PanicsWith(t, func() {
			MovementNone.DistanceTo(geom.Pt(0, 0), geom.Pt(1, 1))
		}, "grid: unknown movement -1")
	})
}

func TestMovement_Contains(t *testing.T) {
	t.Run("cardinal has an edge direction and no corner", func(t *testing.T) {
		assert.True(t, MovementCardinal.Contains(geom.DirectionRight))
		assert.False(t, MovementCardinal.Contains(geom.DirectionUpRight))
	})
	t.Run("diagonal has both", func(t *testing.T) {
		assert.True(t, MovementDiagonal.Contains(geom.DirectionRight))
		assert.True(t, MovementDiagonal.Contains(geom.DirectionUpRight))
	})
	t.Run("agrees with Directions", func(t *testing.T) {
		for _, movement := range Movements() {
			directions := movement.Directions()

			for _, direction := range geom.Directions() {
				assert.Equal(t, movement.Contains(direction), slices.Contains(directions, direction), movement.String()+" "+direction.String()+": ")
			}
		}
	})
	t.Run("no movement has DirectionNone", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.False(t, movement.Contains(geom.DirectionNone), movement.String()+": ")
		}
	})
	t.Run("is false outside the constants", func(t *testing.T) {
		for _, direction := range geom.Directions() {
			assert.False(t, MovementNone.Contains(direction), direction.String()+": ")
			assert.False(t, Movement(99).Contains(direction), direction.String()+": ")
		}
	})
}

func TestMovement_IsNone(t *testing.T) {
	t.Run("the two movements are not none", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.False(t, movement.IsNone(), movement.String()+": ")
		}
	})
	t.Run("none and every value outside the constants are", func(t *testing.T) {
		assert.True(t, MovementNone.IsNone())
		assert.True(t, Movement(99).IsNone())
	})
}

func TestMovement_String(t *testing.T) {
	t.Run("names the constant without its prefix", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.String(), "Cardinal")
		assert.Equal(t, MovementDiagonal.String(), "Diagonal")
	})
	t.Run("none and every value outside the constants print None", func(t *testing.T) {
		assert.Equal(t, MovementNone.String(), "None")
		assert.Equal(t, Movement(99).String(), "None")
	})
}

func TestMovement_JSON(t *testing.T) {
	t.Run("wire format is the name", func(t *testing.T) {
		assert.JSON(t, MovementDiagonal, `"Diagonal"`)
		assert.JSON(t, MovementNone, `"None"`)
	})
	t.Run("round-trip", func(t *testing.T) {
		movements := Movements()
		for _, movement := range append(movements[:], MovementNone) {
			data, err := json.Marshal(movement)
			assert.NoError(t, err, movement.String()+": ")

			var decoded Movement
			assert.NoError(t, json.Unmarshal(data, &decoded), movement.String()+": ")
			assert.Equal(t, decoded, movement, movement.String()+": ")
		}
	})
	t.Run("a map key is the name", func(t *testing.T) {
		assert.JSON(t, map[Movement]int{MovementCardinal: 4}, `{"Cardinal":4}`)
	})
	t.Run("an unknown name is an error", func(t *testing.T) {
		var decoded Movement

		assert.Error(t, json.Unmarshal([]byte(`"Knight"`), &decoded))
	})
}

func TestMovement_Properties(t *testing.T) {
	movements := Movements()
	for _, movement := range append(movements[:], MovementNone) {
		text, err := movement.MarshalText()

		assert.NoError(t, err, movement.String()+": ")
		assert.Equal(t, string(text), movement.String(), movement.String()+": ")
	}
}

func ExampleMovement() {
	fmt.Println(MovementCardinal.Directions())
	fmt.Println(MovementDiagonal.Offsets())
	fmt.Println(MovementCardinal.Offset(geom.DirectionUp), MovementCardinal.Offset(geom.DirectionUpRight))
	fmt.Println(MovementCardinal.Contains(geom.DirectionUpRight))
	fmt.Println(MovementDiagonal.DistanceTo(geom.Pt(0, 0), geom.Pt(3, 5)), MovementCardinal.DistanceTo(geom.Pt(0, 0), geom.Pt(3, 5)))

	fmt.Println(Movements())
	fmt.Println(ParseMovement("Diagonal"))

	data, _ := json.Marshal(MovementCardinal)
	fmt.Println(string(data))
	// Output:
	// [Right Down Left Up]
	// [⟨1,0⟩ ⟨1,1⟩ ⟨0,1⟩ ⟨-1,1⟩ ⟨-1,0⟩ ⟨-1,-1⟩ ⟨0,-1⟩ ⟨1,-1⟩]
	// ⟨0,-1⟩ ⟨0,0⟩
	// false
	// 5 8
	// [Cardinal Diagonal]
	// Diagonal <nil>
	// "Cardinal"
}
