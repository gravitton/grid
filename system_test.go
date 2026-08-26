package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

func TestSystem_String(t *testing.T) {
	assert.Equal(t, Cardinal.String(), "Cardinal")
	assert.Equal(t, Diagonal.String(), "Diagonal")
	assert.Equal(t, System(99).String(), "Unknown")
}

func TestSystem_Directions(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, Cardinal.Directions(), []geom.Direction{geom.DirectionRight, geom.DirectionDown, geom.DirectionLeft, geom.DirectionUp})
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.Equal(t, Diagonal.Directions(), []geom.Direction{
			geom.DirectionRight, geom.DirectionDownRight, geom.DirectionDown, geom.DirectionDownLeft,
			geom.DirectionLeft, geom.DirectionUpLeft, geom.DirectionUp, geom.DirectionUpRight,
		})
	})
	t.Run("copy", func(t *testing.T) {
		Diagonal.Directions()[0] = geom.DirectionNone

		assert.Equal(t, Diagonal.Directions()[0], geom.DirectionRight)
		assert.Equal(t, geom.Directions[0], geom.DirectionRight)
	})
}

func TestSystem_Directions_Panic(t *testing.T) {
	defer func() {
		assert.Equal(t, recover(), "unsupported system")
	}()
	System(99).Directions()
}

func TestSystem_Offsets(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, Cardinal.Offsets(), []geom.Vector[int]{
			geom.Vec(1, 0), geom.Vec(0, 1), geom.Vec(-1, 0), geom.Vec(0, -1),
		})
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.Equal(t, Diagonal.Offsets(), []geom.Vector[int]{
			geom.Vec(1, 0), geom.Vec(1, 1), geom.Vec(0, 1), geom.Vec(-1, 1),
			geom.Vec(-1, 0), geom.Vec(-1, -1), geom.Vec(0, -1), geom.Vec(1, -1),
		})
	})
	t.Run("matches directions", func(t *testing.T) {
		for _, system := range []System{Cardinal, Diagonal} {
			t.Run(system.String(), func(t *testing.T) {
				directions := system.Directions()
				offsets := system.Offsets()

				assert.Equal(t, len(offsets), len(directions))
				for i, direction := range directions {
					assert.Equal(t, offsets[i], direction.Offset[int]())
				}
			})
		}
	})
}

func TestSystem_Offset(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, Cardinal.Offset(geom.East), geom.Vec(1, 0))
		assert.Equal(t, Cardinal.Offset(geom.North), geom.Vec(0, -1))
		assert.Equal(t, Cardinal.Offset(geom.West), geom.Vec(-1, 0))
		assert.Equal(t, Cardinal.Offset(geom.South), geom.Vec(0, 1))
	})
	t.Run("diagonal", func(t *testing.T) {
		for _, direction := range Diagonal.Directions() {
			t.Run(direction.String(), func(t *testing.T) {
				assert.Equal(t, Diagonal.Offset(direction), direction.Offset[int]())
			})
		}
	})
	t.Run("outside the system", func(t *testing.T) {
		assert.Equal(t, Cardinal.Offset(geom.NorthEast), geom.Vec(0, 0))
		assert.Equal(t, Cardinal.Offset(geom.DirectionNone), geom.Vec(0, 0))
		assert.Equal(t, Diagonal.Offset(geom.DirectionNone), geom.Vec(0, 0))
	})
	t.Run("wraps out of range directions", func(t *testing.T) {
		assert.Equal(t, Cardinal.Offset(geom.DirectionRight+8), geom.Vec(1, 0))
		assert.Equal(t, Diagonal.Offset(geom.DirectionDownRight-8), geom.Vec(1, 1))
	})
}

func TestSystem_Has(t *testing.T) {
	t.Run("cardinal", func(t *testing.T) {
		assert.True(t, Cardinal.Has(geom.East))
		assert.False(t, Cardinal.Has(geom.NorthEast))
		assert.False(t, Cardinal.Has(geom.DirectionNone))
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.True(t, Diagonal.Has(geom.East))
		assert.True(t, Diagonal.Has(geom.NorthEast))
		assert.False(t, Diagonal.Has(geom.DirectionNone))
	})
}

func TestSystem_Has_Panic(t *testing.T) {
	defer func() {
		assert.Equal(t, recover(), "unsupported system")
	}()
	System(99).Has(geom.East)
}

func TestSystem_DistanceTo(t *testing.T) {
	origin := geom.Pt(0, 0)
	target := geom.Pt(3, 4)

	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, Cardinal.DistanceTo(origin, target), 7)
		assert.Equal(t, Cardinal.DistanceTo(origin, geom.Pt(-2, 3)), 5)
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.Equal(t, Diagonal.DistanceTo(origin, target), 4)
		assert.Equal(t, Diagonal.DistanceTo(origin, geom.Pt(2, 2)), 2)
	})
	t.Run("symmetric", func(t *testing.T) {
		for _, system := range []System{Cardinal, Diagonal} {
			t.Run(system.String(), func(t *testing.T) {
				assert.Equal(t, system.DistanceTo(origin, target), system.DistanceTo(target, origin))
			})
		}
	})
}

func TestSystem_DistanceTo_Panic(t *testing.T) {
	defer func() {
		assert.Equal(t, recover(), "unsupported system")
	}()
	System(99).DistanceTo(geom.Pt(0, 0), geom.Pt(1, 1))
}
