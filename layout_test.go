package grid

import (
	"testing"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/geomtest"
)

func TestLayout_CellSize(t *testing.T) {
	t.Run("a square is its cell size", func(t *testing.T) {
		l := newLayout(squareFlat, geom.Sz(1, 1), geom.Sz(64.0, 32.0), nil, nil)

		geomtest.AssertSize(t, l.CellSize(), geom.Sz(64.0, 32.0))
	})
	t.Run("a flat-top hexagon is two radii wide", func(t *testing.T) {
		l := newLayout(hexFlatTop, geom.Sz(1, 1), geom.Sz(10.0, 10.0), nil, nil)

		geomtest.AssertSize(t, l.CellSize(), geom.Sz(20.0, geom.Sqrt3*10))
	})
}

func TestLayout_CellSpacing(t *testing.T) {
	t.Run("squares do not interlock", func(t *testing.T) {
		l := newLayout(squareFlat, geom.Sz(1, 1), geom.Sz(64.0, 32.0), nil, nil)

		geomtest.AssertSize(t, l.CellSpacing(), geom.Sz(64.0, 32.0))
	})
	t.Run("flat-top hexagons interlock along x", func(t *testing.T) {
		l := newLayout(hexFlatTop, geom.Sz(1, 1), geom.Sz(10.0, 10.0), nil, nil)

		geomtest.AssertSize(t, l.CellSpacing(), geom.Sz(15.0, geom.Sqrt3*10))
	})
}

func TestLayout_MoveTo(t *testing.T) {
	l := newLayout(squareFlat, geom.Sz(1, 1), geom.Sz(64.0, 32.0), nil, nil).MoveTo(geom.Pt(100.0, 50.0))

	t.Run("places the first cell at the origin", func(t *testing.T) {
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(100.0, 50.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 1)), geom.Pt(164.0, 82.0))
	})
	t.Run("maps the origin back to the first cell", func(t *testing.T) {
		geomtest.AssertPoint(t, l.FromPoint(geom.Pt(100.0, 50.0)), geom.Pt(0, 0))
	})
}

func TestLayout_Resize(t *testing.T) {
	l := newLayout(squareFlat, geom.Sz(1, 1), geom.Sz(64.0, 32.0), nil, nil).MoveTo(geom.Pt(10.0, 20.0)).Resize(geom.Sz(128.0, 64.0))

	t.Run("replaces the cell size", func(t *testing.T) {
		geomtest.AssertSize(t, l.CellSize(), geom.Sz(128.0, 64.0))
		geomtest.AssertSize(t, l.CellSpacing(), geom.Sz(128.0, 64.0))
	})
	t.Run("keeps the origin", func(t *testing.T) {
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(10.0, 20.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 1)), geom.Pt(138.0, 84.0))
		geomtest.AssertPoint(t, l.FromPoint(geom.Pt(10.0, 20.0)), geom.Pt(0, 0))
		geomtest.AssertPoint(t, l.FromPoint(geom.Pt(138.0, 84.0)), geom.Pt(1, 1))
	})
}

func TestLayout_ToPoint(t *testing.T) {
	t.Run("square", func(t *testing.T) {
		l := newLayout(squareFlat, geom.Sz(1, 1), geom.Sz(64.0, 32.0), nil, nil)

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(0.0, 0.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 0)), geom.Pt(64.0, 0.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 1)), geom.Pt(0.0, 32.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(2, 3)), geom.Pt(128.0, 96.0))
	})
	t.Run("pointy-top hexagon", func(t *testing.T) {
		l := newLayout(hexPointyTop, geom.Sz(1, 1), geom.Sz(10.0, 10.0), nil, nil)

		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 0)), geom.Pt(0.0, 0.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(1, 0)), geom.Pt(geom.Sqrt3*10, 0.0))
		geomtest.AssertPoint(t, l.ToPoint(geom.Pt(0, 1)), geom.Pt(geom.Sqrt3/2*10, 15.0))
	})
}

func TestLayout_FromPoint(t *testing.T) {
	t.Run("square", func(t *testing.T) {
		l := newLayout(squareFlat, geom.Sz(1, 1), geom.Sz(64.0, 32.0), nil, nil)

		geomtest.AssertPoint(t, l.FromPoint(geom.Pt(0.0, 0.0)), geom.Pt(0, 0))
		geomtest.AssertPoint(t, l.FromPoint(geom.Pt(64.0, 0.0)), geom.Pt(1, 0))
		geomtest.AssertPoint(t, l.FromPoint(geom.Pt(0.0, 32.0)), geom.Pt(0, 1))
		geomtest.AssertPoint(t, l.FromPoint(geom.Pt(128.0, 96.0)), geom.Pt(2, 3))
	})
	t.Run("round-trips through ToPoint", func(t *testing.T) {
		l := newLayout(hexPointyTop, geom.Sz(1, 1), geom.Sz(20.0, 20.0), nil, nil).MoveTo(geom.Pt(100.0, 100.0))
		index := geom.Pt(3, -2)

		geomtest.AssertPoint(t, l.FromPoint(l.ToPoint(index)), index)
	})
}
