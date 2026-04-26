package square

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

func TestDirection_String(t *testing.T) {
	assert.Equal(t, E.String(), "E")
	assert.Equal(t, NE.String(), "NE")
	assert.Equal(t, N.String(), "N")
	assert.Equal(t, NW.String(), "NW")
	assert.Equal(t, W.String(), "W")
	assert.Equal(t, SW.String(), "SW")
	assert.Equal(t, S.String(), "S")
	assert.Equal(t, SE.String(), "SE")
	// values beyond 7 wrap via %8
	assert.Equal(t, Direction(8).String(), "E")
	assert.Equal(t, Direction(10).String(), "N")
}

func TestDirection_Opposite(t *testing.T) {
	assert.Equal(t, E.Opposite(), W)
	assert.Equal(t, NE.Opposite(), SW)
	assert.Equal(t, N.Opposite(), S)
	assert.Equal(t, NW.Opposite(), SE)
	assert.Equal(t, W.Opposite(), E)
	assert.Equal(t, SW.Opposite(), NE)
	assert.Equal(t, S.Opposite(), N)
	assert.Equal(t, SE.Opposite(), NW)

	// double opposite returns the original direction
	for _, d := range []Direction{E, NE, N, NW, W, SW, S, SE} {
		assert.Equal(t, d.Opposite().Opposite(), d)
	}

	// neighbor vectors of opposite directions sum to zero
	for _, d := range []Direction{E, NE, N, NW} {
		v := Directions[d]
		opp := Directions[d.Opposite()]
		assert.Equal(t, v.X+opp.X, 0)
		assert.Equal(t, v.Y+opp.Y, 0)
	}
}

func TestAllDirections_Order(t *testing.T) {
	assert.Equal(t, Directions[E], geom.Vec(1, 0))
	assert.Equal(t, Directions[NE], geom.Vec(1, -1))
	assert.Equal(t, Directions[N], geom.Vec(0, -1))
	assert.Equal(t, Directions[NW], geom.Vec(-1, -1))
	assert.Equal(t, Directions[W], geom.Vec(-1, 0))
	assert.Equal(t, Directions[SW], geom.Vec(-1, 1))
	assert.Equal(t, Directions[S], geom.Vec(0, 1))
	assert.Equal(t, Directions[SE], geom.Vec(1, 1))
}

func TestNeighborOffsets(t *testing.T) {
	cardinal := NeighborOffsets(Cardinal)
	assert.Equal(t, len(cardinal), 4)
	assert.Equal(t, cardinal, CardinalDirections[:])

	diagonal := NeighborOffsets(Diagonal)
	assert.Equal(t, len(diagonal), 8)
	assert.Equal(t, diagonal, Directions[:])
}

func TestNeighborOffsets_Panic(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, r, "unsupported system")
	}()
	NeighborOffsets(System(99))
}

func TestDistanceTo(t *testing.T) {
	origin := geom.Pt(0, 0)
	target := geom.Pt(3, 4)

	// Cardinal: Manhattan distance
	assert.Equal(t, DistanceTo(origin, target, Cardinal), 7)
	assert.Equal(t, DistanceTo(origin, geom.Pt(-2, 3), Cardinal), 5)

	// Diagonal: Chebyshev distance
	assert.Equal(t, DistanceTo(origin, target, Diagonal), 4)
	assert.Equal(t, DistanceTo(origin, geom.Pt(2, 2), Diagonal), 2)

	// symmetric
	assert.Equal(t, DistanceTo(origin, target, Cardinal), DistanceTo(target, origin, Cardinal))
	assert.Equal(t, DistanceTo(origin, target, Diagonal), DistanceTo(target, origin, Diagonal))
}

func TestDistanceTo_Panic(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, r, "unsupported system")
	}()
	DistanceTo(geom.Pt(0, 0), geom.Pt(1, 1), System(99))
}

func TestRange(t *testing.T) {
	origin := geom.Pt(0, 0)

	assert.Equal(t, Range(origin, -1), nil)

	r0 := Range(origin, 0)
	assert.Equal(t, len(r0), 1)
	assert.Equal(t, r0[0], geom.Pt(0, 0))

	// n=1: circle of radius 1 — center + 4 cardinal neighbors (diagonals have dist √2 > 1)
	r1 := Range(origin, 1)
	assert.Equal(t, len(r1), 5)

	// n=2: 13 points with dx²+dy² ≤ 4
	r2 := Range(origin, 2)
	assert.Equal(t, len(r2), 13)

	// all returned points are within radius
	for _, pt := range r2 {
		dx, dy := pt.X-origin.X, pt.Y-origin.Y
		assert.True(t, dx*dx+dy*dy <= 4)
	}

	// non-origin center shifts all points
	center := geom.Pt(3, -2)
	for _, pt := range Range(center, 1) {
		dx, dy := pt.X-center.X, pt.Y-center.Y
		assert.True(t, dx*dx+dy*dy <= 1)
	}
}

func TestHasLineOfSight(t *testing.T) {
	origin := geom.Pt(0, 0)

	// clear line with no blockers
	assert.True(t, HasLineOfSight(origin, geom.Pt(3, 0), nil))
	assert.True(t, HasLineOfSight(origin, geom.Pt(3, 0), []ints.Point{}))

	// same point always has line of sight to itself
	assert.True(t, HasLineOfSight(origin, origin, nil))

	// blocked in the middle
	assert.False(t, HasLineOfSight(origin, geom.Pt(3, 0), []ints.Point{geom.Pt(1, 0)}))
	assert.False(t, HasLineOfSight(origin, geom.Pt(3, 0), []ints.Point{geom.Pt(2, 0)}))

	// blocker beyond the target does not affect visibility
	assert.True(t, HasLineOfSight(origin, geom.Pt(2, 0), []ints.Point{geom.Pt(3, 0)}))

	// target in the blocking list is still visible (can see into, not through)
	assert.True(t, HasLineOfSight(origin, geom.Pt(3, 0), []ints.Point{geom.Pt(3, 0)}))

	// diagonal line of sight
	assert.True(t, HasLineOfSight(origin, geom.Pt(3, 3), nil))
	assert.False(t, HasLineOfSight(origin, geom.Pt(3, 3), []ints.Point{geom.Pt(1, 1)}))
}

func TestFieldOfView(t *testing.T) {
	center := geom.Pt(0, 0)
	candidates := Range(center, 3)

	// no blocking: all candidates are visible
	assert.Equal(t, len(FieldOfView(center, candidates, nil)), len(candidates))
	assert.Equal(t, len(FieldOfView(center, candidates, []ints.Point{})), len(candidates))

	// empty candidates
	assert.Equal(t, len(FieldOfView(center, nil, nil)), 0)
	assert.Equal(t, len(FieldOfView(center, []ints.Point{}, nil)), 0)

	// adjacent cells (Chebyshev distance ≤ 1) are always visible even when blocking
	allNeighbors := []ints.Point{
		geom.Pt(1, 0), geom.Pt(0, -1), geom.Pt(-1, 0), geom.Pt(0, 1),
		geom.Pt(1, -1), geom.Pt(-1, -1), geom.Pt(-1, 1), geom.Pt(1, 1),
	}
	visible := FieldOfView(center, allNeighbors, allNeighbors)
	assert.Equal(t, len(visible), len(allNeighbors))

	// a gap in the blocking ring allows sight through it
	// only (1,0) is unblocked, so (2,0) and (3,0) should be visible
	blocking := []ints.Point{
		geom.Pt(0, -1), geom.Pt(-1, -1), geom.Pt(-1, 0), geom.Pt(-1, 1),
		geom.Pt(0, 1), geom.Pt(1, -1), geom.Pt(1, 1),
	}
	visible = FieldOfView(center, []ints.Point{geom.Pt(2, 0), geom.Pt(3, 0), geom.Pt(0, 2)}, blocking)
	assert.Equal(t, len(visible), 2)
	assert.Equal(t, visible[0], geom.Pt(2, 0))
	assert.Equal(t, visible[1], geom.Pt(3, 0))
}
