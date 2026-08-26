package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

func TestPoint_Constructor(t *testing.T) {
	p := geom.Pt(3, -2)
	s := Pt(p.XY())
	assert.Equal(t, s.Point(), p)
}

func TestPoint_Range(t *testing.T) {
	origin := Pt(0, 0)

	assert.Equal(t, origin.Range(-1), nil)

	r0 := origin.Range(0)
	assert.Equal(t, len(r0), 1)
	assert.Equal(t, r0[0], geom.Pt(0, 0))

	// n=1: circle of radius 1 — center + 4 cardinal neighbors (diagonals have dist √2 > 1)
	r1 := origin.Range(1)
	assert.Equal(t, len(r1), 5)

	// n=2: 13 points with dx²+dy² ≤ 4
	r2 := origin.Range(2)
	assert.Equal(t, len(r2), 13)

	// all returned points are within radius
	for _, pt := range r2 {
		dx, dy := pt.X-origin.X, pt.Y-origin.Y
		assert.True(t, dx*dx+dy*dy <= 4)
	}

	// non-origin center shifts all points
	center := Pt(3, -2)
	for _, pt := range center.Range(1) {
		dx, dy := pt.X-center.X, pt.Y-center.Y
		assert.True(t, dx*dx+dy*dy <= 1)
	}
}

func TestPoint_HasLineOfSight(t *testing.T) {
	origin := Pt(0, 0)

	t.Run("clear", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), nil))
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{}))
		assert.True(t, origin.HasLineOfSight(origin.Point(), nil))
	})
	t.Run("blocked mid", func(t *testing.T) {
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{geom.Pt(1, 0)}))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{geom.Pt(2, 0)}))
	})
	t.Run("blocker beyond target", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 0), []ints.Point{geom.Pt(3, 0)}))
	})
	t.Run("target in blocking list", func(t *testing.T) {
		// can see into, not through
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{geom.Pt(3, 0)}))
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 3), nil))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 3), []ints.Point{geom.Pt(1, 1)}))
	})
	t.Run("corner between two blockers is opaque", func(t *testing.T) {
		corner := []ints.Point{geom.Pt(1, 0), geom.Pt(0, 1)}

		assert.False(t, origin.HasLineOfSight(geom.Pt(1, 1), corner))
		assert.False(t, origin.HasLineOfSight(geom.Pt(2, 2), corner))
	})
	t.Run("corner beside a single blocker stays open", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), []ints.Point{geom.Pt(1, 0)}))
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), []ints.Point{geom.Pt(0, 1)}))
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 2), []ints.Point{geom.Pt(2, 1)}))
	})
	t.Run("reciprocal", func(t *testing.T) {
		// a Bresenham walk traces a different chain of cells when reversed, so these
		// pairs disagreed until the endpoints were canonicalized
		blocking := []ints.Point{geom.Pt(1, 1)}
		for _, target := range []ints.Point{geom.Pt(1, 2), geom.Pt(2, 1), geom.Pt(2, 4), geom.Pt(4, 2)} {
			t.Run(target.String(), func(t *testing.T) {
				forward := origin.HasLineOfSight(target, blocking)
				backward := Pt(target.XY()).HasLineOfSight(origin.Point(), blocking)

				assert.Equal(t, forward, backward)
			})
		}
	})
	t.Run("reciprocal across a box", func(t *testing.T) {
		blocking := []ints.Point{geom.Pt(1, 1), geom.Pt(-2, 1), geom.Pt(0, -2)}
		for x := -4; x <= 4; x++ {
			for y := -4; y <= 4; y++ {
				a, b := Pt(0, 0), geom.Pt(x, y)

				assert.Equal(t, a.HasLineOfSight(b, blocking), Pt(x, y).HasLineOfSight(a.Point(), blocking))
			}
		}
	})
	t.Run("diagonal wall hides everything behind it", func(t *testing.T) {
		var wall []ints.Point
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
	candidates := center.Range(3)

	t.Run("no blocking", func(t *testing.T) {
		assert.Equal(t, len(center.FieldOfView(candidates, nil)), len(candidates))
		assert.Equal(t, len(center.FieldOfView(candidates, []ints.Point{})), len(candidates))
	})
	t.Run("empty candidates", func(t *testing.T) {
		assert.Equal(t, len(center.FieldOfView(nil, nil)), 0)
		assert.Equal(t, len(center.FieldOfView([]ints.Point{}, nil)), 0)
	})
	t.Run("adjacent always visible", func(t *testing.T) {
		allNeighbors := []ints.Point{
			geom.Pt(1, 0), geom.Pt(0, -1), geom.Pt(-1, 0), geom.Pt(0, 1),
			geom.Pt(1, -1), geom.Pt(-1, -1), geom.Pt(-1, 1), geom.Pt(1, 1),
		}
		visible := center.FieldOfView(allNeighbors, allNeighbors)
		assert.Equal(t, len(visible), len(allNeighbors))
	})
	t.Run("gap in blocking ring", func(t *testing.T) {
		// only (1,0) is unblocked, so (2,0) and (3,0) should be visible; (0,2) is not
		blocking := []ints.Point{
			geom.Pt(0, -1), geom.Pt(-1, -1), geom.Pt(-1, 0), geom.Pt(-1, 1),
			geom.Pt(0, 1), geom.Pt(1, -1), geom.Pt(1, 1),
		}
		visible := center.FieldOfView([]ints.Point{geom.Pt(2, 0), geom.Pt(3, 0), geom.Pt(0, 2)}, blocking)
		assert.Equal(t, len(visible), 2)
		assert.Equal(t, visible[0], geom.Pt(2, 0))
		assert.Equal(t, visible[1], geom.Pt(3, 0))
	})
}
