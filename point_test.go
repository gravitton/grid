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

	// clear line with no blockers
	assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), nil))
	assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{}))

	// same point always has line of sight to itself
	assert.True(t, origin.HasLineOfSight(origin.Point(), nil))

	// blocked in the middle
	assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{geom.Pt(1, 0)}))
	assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{geom.Pt(2, 0)}))

	// blocker beyond the target does not affect visibility
	assert.True(t, origin.HasLineOfSight(geom.Pt(2, 0), []ints.Point{geom.Pt(3, 0)}))

	// target in the blocking list is still visible (can see into, not through)
	assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []ints.Point{geom.Pt(3, 0)}))

	// diagonal line of sight
	assert.True(t, origin.HasLineOfSight(geom.Pt(3, 3), nil))
	assert.False(t, origin.HasLineOfSight(geom.Pt(3, 3), []ints.Point{geom.Pt(1, 1)}))
}

func TestPoint_FieldOfView(t *testing.T) {
	center := Pt(0, 0)
	candidates := center.Range(3)

	// no blocking: all candidates are visible
	assert.Equal(t, len(center.FieldOfView(candidates, nil)), len(candidates))
	assert.Equal(t, len(center.FieldOfView(candidates, []ints.Point{})), len(candidates))

	// empty candidates
	assert.Equal(t, len(center.FieldOfView(nil, nil)), 0)
	assert.Equal(t, len(center.FieldOfView([]ints.Point{}, nil)), 0)

	// adjacent cells (Chebyshev distance ≤ 1) are always visible even when blocking
	allNeighbors := []ints.Point{
		geom.Pt(1, 0), geom.Pt(0, -1), geom.Pt(-1, 0), geom.Pt(0, 1),
		geom.Pt(1, -1), geom.Pt(-1, -1), geom.Pt(-1, 1), geom.Pt(1, 1),
	}
	visible := center.FieldOfView(allNeighbors, allNeighbors)
	assert.Equal(t, len(visible), len(allNeighbors))

	// a gap in the blocking ring allows sight through it
	// only (1,0) is unblocked, so (2,0) and (3,0) should be visible
	blocking := []ints.Point{
		geom.Pt(0, -1), geom.Pt(-1, -1), geom.Pt(-1, 0), geom.Pt(-1, 1),
		geom.Pt(0, 1), geom.Pt(1, -1), geom.Pt(1, 1),
	}
	visible = center.FieldOfView([]ints.Point{geom.Pt(2, 0), geom.Pt(3, 0), geom.Pt(0, 2)}, blocking)
	assert.Equal(t, len(visible), 2)
	assert.Equal(t, visible[0], geom.Pt(2, 0))
	assert.Equal(t, visible[1], geom.Pt(3, 0))
}
