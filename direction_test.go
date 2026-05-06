package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
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

func TestNeighborOffset(t *testing.T) {
	// Cardinal: all four cardinal directions return the correct vector.
	assert.Equal(t, NeighborOffset(Cardinal, E), geom.Vec(1, 0))
	assert.Equal(t, NeighborOffset(Cardinal, N), geom.Vec(0, -1))
	assert.Equal(t, NeighborOffset(Cardinal, W), geom.Vec(-1, 0))
	assert.Equal(t, NeighborOffset(Cardinal, S), geom.Vec(0, 1))

	// Diagonal: all eight directions return the correct vector.
	for _, d := range []Direction{E, NE, N, NW, W, SW, S, SE} {
		assert.Equal(t, NeighborOffset(Diagonal, d), Directions[d])
	}
}

func TestNeighborOffset_Panic(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, r, "diagonal direction not available in Cardinal movement system")
	}()
	NeighborOffset(Cardinal, NE)
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
