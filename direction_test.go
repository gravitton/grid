package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

func TestDirection_String(t *testing.T) {
	cases := []struct {
		name string
		dir  Direction
		want string
	}{
		{"E", E, "E"},
		{"NE", NE, "NE"},
		{"N", N, "N"},
		{"NW", NW, "NW"},
		{"W", W, "W"},
		{"SW", SW, "SW"},
		{"S", S, "S"},
		{"SE", SE, "SE"},
		{"wrap_8_to_E", Direction(8), "E"},
		{"wrap_10_to_N", Direction(10), "N"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.dir.String(), tc.want)
		})
	}
}

func TestDirection_Opposite(t *testing.T) {
	t.Run("pairs", func(t *testing.T) {
		cases := []struct{ dir, want Direction }{
			{E, W},
			{NE, SW},
			{N, S},
			{NW, SE},
			{W, E},
			{SW, NE},
			{S, N},
			{SE, NW},
		}
		for _, tc := range cases {
			t.Run(tc.dir.String(), func(t *testing.T) {
				assert.Equal(t, tc.dir.Opposite(), tc.want)
			})
		}
	})
	t.Run("double opposite", func(t *testing.T) {
		for _, d := range []Direction{E, NE, N, NW, W, SW, S, SE} {
			assert.Equal(t, d.Opposite().Opposite(), d)
		}
	})
	t.Run("vectors sum to zero", func(t *testing.T) {
		for _, d := range []Direction{E, NE, N, NW} {
			v := Directions[d]
			opp := Directions[d.Opposite()]
			assert.Equal(t, v.X+opp.X, 0)
			assert.Equal(t, v.Y+opp.Y, 0)
		}
	})
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
	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, NeighborOffset(Cardinal, E), geom.Vec(1, 0))
		assert.Equal(t, NeighborOffset(Cardinal, N), geom.Vec(0, -1))
		assert.Equal(t, NeighborOffset(Cardinal, W), geom.Vec(-1, 0))
		assert.Equal(t, NeighborOffset(Cardinal, S), geom.Vec(0, 1))
	})
	t.Run("diagonal", func(t *testing.T) {
		for _, d := range []Direction{E, NE, N, NW, W, SW, S, SE} {
			t.Run(d.String(), func(t *testing.T) {
				assert.Equal(t, NeighborOffset(Diagonal, d), Directions[d])
			})
		}
	})
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

	t.Run("cardinal", func(t *testing.T) {
		assert.Equal(t, DistanceTo(origin, target, Cardinal), 7)
		assert.Equal(t, DistanceTo(origin, geom.Pt(-2, 3), Cardinal), 5)
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.Equal(t, DistanceTo(origin, target, Diagonal), 4)
		assert.Equal(t, DistanceTo(origin, geom.Pt(2, 2), Diagonal), 2)
	})
	t.Run("symmetric", func(t *testing.T) {
		assert.Equal(t, DistanceTo(origin, target, Cardinal), DistanceTo(target, origin, Cardinal))
		assert.Equal(t, DistanceTo(origin, target, Diagonal), DistanceTo(target, origin, Diagonal))
	})
}

func TestDistanceTo_Panic(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, r, "unsupported system")
	}()
	DistanceTo(geom.Pt(0, 0), geom.Pt(1, 1), System(99))
}
