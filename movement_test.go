package grid

import (
	"encoding/json"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
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
	t.Run("none", func(t *testing.T) {
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
	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Directions(), []geom.Direction{geom.DirectionRight, geom.DirectionDown, geom.DirectionLeft, geom.DirectionUp})
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.Equal(t, MovementDiagonal.Directions(), []geom.Direction{
			geom.DirectionRight, geom.DirectionDownRight, geom.DirectionDown, geom.DirectionDownLeft,
			geom.DirectionLeft, geom.DirectionUpLeft, geom.DirectionUp, geom.DirectionUpRight,
		})
	})
	t.Run("copy", func(t *testing.T) {
		MovementDiagonal.Directions()[0] = geom.DirectionNone

		assert.Equal(t, MovementDiagonal.Directions()[0], geom.DirectionRight)
		assert.Equal(t, geom.Directions()[0], geom.DirectionRight)
	})
	t.Run("panics outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Movement(99).Directions()
		}, "grid: unknown movement 99")
		assert.PanicsWith(t, func() {
			MovementNone.Directions()
		}, "grid: unknown movement -1")
	})
}

func TestMovement_Offsets(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Offsets(), []geom.Vector[int]{
			geom.Vec(1, 0), geom.Vec(0, 1), geom.Vec(-1, 0), geom.Vec(0, -1),
		})
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.Equal(t, MovementDiagonal.Offsets(), []geom.Vector[int]{
			geom.Vec(1, 0), geom.Vec(1, 1), geom.Vec(0, 1), geom.Vec(-1, 1),
			geom.Vec(-1, 0), geom.Vec(-1, -1), geom.Vec(0, -1), geom.Vec(1, -1),
		})
	})
	t.Run("matches directions", func(t *testing.T) {
		for _, movement := range Movements() {
			t.Run(movement.String(), func(t *testing.T) {
				directions := movement.Directions()
				offsets := movement.Offsets()

				assert.Equal(t, len(offsets), len(directions))
				for i, direction := range directions {
					assert.Equal(t, offsets[i], direction.Offset[int]())
				}
			})
		}
	})
}

func TestMovement_Offset(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Offset(geom.East), geom.Vec(1, 0))
		assert.Equal(t, MovementCardinal.Offset(geom.North), geom.Vec(0, -1))
		assert.Equal(t, MovementCardinal.Offset(geom.West), geom.Vec(-1, 0))
		assert.Equal(t, MovementCardinal.Offset(geom.South), geom.Vec(0, 1))
	})
	t.Run("diagonal", func(t *testing.T) {
		for _, direction := range MovementDiagonal.Directions() {
			t.Run(direction.String(), func(t *testing.T) {
				assert.Equal(t, MovementDiagonal.Offset(direction), direction.Offset[int]())
			})
		}
	})
	t.Run("outside the movement", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Offset(geom.NorthEast), geom.Vec(0, 0))
		assert.Equal(t, MovementCardinal.Offset(geom.DirectionNone), geom.Vec(0, 0))
		assert.Equal(t, MovementDiagonal.Offset(geom.DirectionNone), geom.Vec(0, 0))
	})
	t.Run("wraps out of range directions", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.Offset(geom.DirectionRight+8), geom.Vec(1, 0))
		assert.Equal(t, MovementDiagonal.Offset(geom.DirectionDownRight-8), geom.Vec(1, 1))
	})
}

func TestMovement_Has(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		assert.True(t, MovementCardinal.Has(geom.East))
		assert.False(t, MovementCardinal.Has(geom.NorthEast))
		assert.False(t, MovementCardinal.Has(geom.DirectionNone))
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.True(t, MovementDiagonal.Has(geom.East))
		assert.True(t, MovementDiagonal.Has(geom.NorthEast))
		assert.False(t, MovementDiagonal.Has(geom.DirectionNone))
	})
	t.Run("panics outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Movement(99).Has(geom.East)
		}, "grid: unknown movement 99")
		assert.PanicsWith(t, func() {
			MovementNone.Has(geom.East)
		}, "grid: unknown movement -1")
	})
}

func TestMovement_DistanceTo(t *testing.T) {
	origin := geom.Pt(0, 0)
	target := geom.Pt(3, 4)

	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, MovementCardinal.DistanceTo(origin, target), 7)
		assert.Equal(t, MovementCardinal.DistanceTo(origin, geom.Pt(-2, 3)), 5)
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.Equal(t, MovementDiagonal.DistanceTo(origin, target), 4)
		assert.Equal(t, MovementDiagonal.DistanceTo(origin, geom.Pt(2, 2)), 2)
	})
	t.Run("symmetric", func(t *testing.T) {
		for _, movement := range Movements() {
			t.Run(movement.String(), func(t *testing.T) {
				assert.Equal(t, movement.DistanceTo(origin, target), movement.DistanceTo(target, origin))
			})
		}
	})
	t.Run("panics outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Movement(99).DistanceTo(geom.Pt(0, 0), geom.Pt(1, 1))
		}, "grid: unknown movement 99")
		assert.PanicsWith(t, func() {
			MovementNone.DistanceTo(geom.Pt(0, 0), geom.Pt(1, 1))
		}, "grid: unknown movement -1")
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
			assert.NoError(t, err)

			var decoded Movement
			assert.NoError(t, json.Unmarshal(data, &decoded))
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
