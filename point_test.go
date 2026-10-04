package grid

import (
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

func TestPoint_Constructor(t *testing.T) {
	assert.Equal(t, Pt(3, -2).Point(), geom.Pt(3, -2))
}

func TestPoint_Neighbors(t *testing.T) {
	t.Run("cardinal is the four edge neighbors by increasing angle", func(t *testing.T) {
		assert.Equal(t, Pt(3, -2).Neighbors(MovementCardinal), []geom.Point[int]{geom.Pt(4, -2), geom.Pt(3, -1), geom.Pt(2, -2), geom.Pt(3, -3)})
	})
	t.Run("diagonal is all eight", func(t *testing.T) {
		assert.Length(t, Pt(3, -2).Neighbors(MovementDiagonal), 8)
	})
}

func TestPoint_Range(t *testing.T) {
	center := Pt(3, -2)

	t.Run("a negative radius is nil", func(t *testing.T) {
		assert.Equal(t, center.Range(-1, MovementCardinal), nil)
	})
	t.Run("a zero radius is the point alone", func(t *testing.T) {
		assert.Equal(t, center.Range(0, MovementDiagonal), []geom.Point[int]{geom.Pt(3, -2)})
	})
	t.Run("cardinal is the diamond of Manhattan distance", func(t *testing.T) {
		assert.Equal(t, center.Range(1, MovementCardinal), []geom.Point[int]{geom.Pt(2, -2), geom.Pt(3, -3), geom.Pt(3, -2), geom.Pt(3, -1), geom.Pt(4, -2)})
		assert.Length(t, center.Range(3, MovementCardinal), 25)
	})
	t.Run("diagonal is the square of Chebyshev distance", func(t *testing.T) {
		assert.Length(t, center.Range(1, MovementDiagonal), 9)
		assert.Length(t, center.Range(3, MovementDiagonal), 49)
	})
	t.Run("holds exactly the indices within the distance", func(t *testing.T) {
		for _, movement := range Movements() {
			for n := range 5 {
				indices := center.Range(n, movement)

				for _, index := range indices {
					assert.LessOrEqual(t, center.DistanceTo(index, movement), n, movement.String()+" at "+index.String()+": ")
				}
				for _, index := range center.Range(n+1, movement) {
					assert.Equal(t, Reg(indices).Contains(index), center.DistanceTo(index, movement) <= n, movement.String()+" at "+index.String()+": ")
				}
			}
		}
	})
	t.Run("is sorted and allocated once at its length", func(t *testing.T) {
		for _, movement := range Movements() {
			indices := center.Range(4, movement)

			assert.True(t, slices.IsSortedFunc(indices, geom.Point[int].Compare), movement.String()+": ")
			assert.Equal(t, cap(indices), len(indices), movement.String()+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = center.Range(4, movement)
			}), 1.0, movement.String()+": ")
		}
	})
	t.Run("panics for a movement outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			center.Range(1, MovementNone)
		}, "grid: unknown movement -1")
	})
}

func TestPoint_DistanceTo(t *testing.T) {
	t.Run("cardinal is the Manhattan distance", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).DistanceTo(geom.Pt(3, 4), MovementCardinal), 7)
	})
	t.Run("diagonal is the Chebyshev distance", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).DistanceTo(geom.Pt(3, 4), MovementDiagonal), 4)
	})
}

func TestPoint_HasLineOfSight(t *testing.T) {
	origin := Pt(0, 0)

	t.Run("clear", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), nil))
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{}))
		assert.True(t, origin.HasLineOfSight(origin.Point(), nil))
	})
	t.Run("blocked mid", func(t *testing.T) {
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{geom.Pt(1, 0)}))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{geom.Pt(2, 0)}))
	})
	t.Run("blocker beyond target", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 0), []geom.Point[int]{geom.Pt(3, 0)}))
	})
	t.Run("target in blocking list", func(t *testing.T) {
		// can see into, not through
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{geom.Pt(3, 0)}))
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 3), nil))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 3), []geom.Point[int]{geom.Pt(1, 1)}))
	})
	t.Run("corner between two blockers is opaque", func(t *testing.T) {
		corner := []geom.Point[int]{geom.Pt(1, 0), geom.Pt(0, 1)}

		assert.False(t, origin.HasLineOfSight(geom.Pt(1, 1), corner))
		assert.False(t, origin.HasLineOfSight(geom.Pt(2, 2), corner))
	})
	t.Run("corner beside a single blocker stays open", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), []geom.Point[int]{geom.Pt(1, 0)}))
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), []geom.Point[int]{geom.Pt(0, 1)}))
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 2), []geom.Point[int]{geom.Pt(2, 1)}))
	})
	t.Run("reciprocal", func(t *testing.T) {
		// a Bresenham walk traces a different chain of cells when reversed, so these
		// pairs disagreed until the endpoints were canonicalized
		blocking := []geom.Point[int]{geom.Pt(1, 1)}
		for _, target := range []geom.Point[int]{geom.Pt(1, 2), geom.Pt(2, 1), geom.Pt(2, 4), geom.Pt(4, 2)} {
			t.Run(target.String(), func(t *testing.T) {
				forward := origin.HasLineOfSight(target, blocking)
				backward := Pt(target.XY()).HasLineOfSight(origin.Point(), blocking)

				assert.Equal(t, forward, backward)
			})
		}
	})
	t.Run("reciprocal across a box", func(t *testing.T) {
		blocking := []geom.Point[int]{geom.Pt(1, 1), geom.Pt(-2, 1), geom.Pt(0, -2)}
		for x := -4; x <= 4; x++ {
			for y := -4; y <= 4; y++ {
				a, b := Pt(0, 0), geom.Pt(x, y)

				assert.Equal(t, a.HasLineOfSight(b, blocking), Pt(x, y).HasLineOfSight(a.Point(), blocking))
			}
		}
	})
	t.Run("diagonal wall hides everything behind it", func(t *testing.T) {
		var wall []geom.Point[int]
		for i := -3; i <= 3; i++ {
			wall = append(wall, geom.Pt(2+i, 2-i))
		}

		for x := 3; x <= 6; x++ {
			for y := 3; y <= 6; y++ {
				assert.False(t, origin.HasLineOfSight(geom.Pt(x, y), wall))
			}
		}
	})
}

func TestPoint_FieldOfView(t *testing.T) {
	center := Pt(0, 0)
	candidates := center.Range(3, MovementDiagonal)

	t.Run("no blocking", func(t *testing.T) {
		assert.Equal(t, len(center.FieldOfView(candidates, nil)), len(candidates))
		assert.Equal(t, len(center.FieldOfView(candidates, []geom.Point[int]{})), len(candidates))
	})
	t.Run("empty candidates", func(t *testing.T) {
		assert.Equal(t, len(center.FieldOfView(nil, nil)), 0)
		assert.Equal(t, len(center.FieldOfView([]geom.Point[int]{}, nil)), 0)
	})
	t.Run("adjacent always visible", func(t *testing.T) {
		allNeighbors := []geom.Point[int]{
			geom.Pt(1, 0), geom.Pt(0, -1), geom.Pt(-1, 0), geom.Pt(0, 1),
			geom.Pt(1, -1), geom.Pt(-1, -1), geom.Pt(-1, 1), geom.Pt(1, 1),
		}
		visible := center.FieldOfView(allNeighbors, allNeighbors)
		assert.Equal(t, len(visible), len(allNeighbors))
	})
	t.Run("gap in blocking ring", func(t *testing.T) {
		// only (1,0) is unblocked, so (2,0) and (3,0) should be visible; (0,2) is not
		blocking := []geom.Point[int]{
			geom.Pt(0, -1), geom.Pt(-1, -1), geom.Pt(-1, 0), geom.Pt(-1, 1),
			geom.Pt(0, 1), geom.Pt(1, -1), geom.Pt(1, 1),
		}
		visible := center.FieldOfView([]geom.Point[int]{geom.Pt(2, 0), geom.Pt(3, 0), geom.Pt(0, 2)}, blocking)
		assert.Equal(t, len(visible), 2)
		assert.Equal(t, visible[0], geom.Pt(2, 0))
		assert.Equal(t, visible[1], geom.Pt(3, 0))
	})
}

func TestPoint_Point(t *testing.T) {
	assert.Equal(t, Pt(3, -2).Point(), geom.Pt(3, -2))
}
