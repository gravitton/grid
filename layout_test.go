package grid

import (
	"testing"

	geom "github.com/gravitton/geometry"
)

func TestLayoutSquareFlatToPoint(t *testing.T) {
	l := NewLayout(SquareFlat, LayoutOpts.CellSize(geom.Sz(64.0, 32.0)))
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](0, 0)), 0.0, 0.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](1, 0)), 64.0, 0.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](0, 1)), 0.0, 32.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](2, 3)), 128.0, 96.0)
}

func TestLayoutSquareFlatFromPoint(t *testing.T) {
	l := NewLayout(SquareFlat, LayoutOpts.CellSize(geom.Sz(64.0, 32.0)))
	geom.AssertPoint(t, l.FromPoint(geom.Pt(0.0, 0.0)), 0, 0)
	geom.AssertPoint(t, l.FromPoint(geom.Pt(64.0, 0.0)), 1, 0)
	geom.AssertPoint(t, l.FromPoint(geom.Pt(0.0, 32.0)), 0, 1)
	geom.AssertPoint(t, l.FromPoint(geom.Pt(128.0, 96.0)), 2, 3)
}

func TestLayoutSquareFlatWithOrigin(t *testing.T) {
	l := NewLayout(SquareFlat, LayoutOpts.CellSize(geom.Sz(64.0, 32.0))).MoveTo(geom.Pt(100.0, 50.0))
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](0, 0)), 100.0, 50.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](1, 1)), 164.0, 82.0)
	geom.AssertPoint(t, l.FromPoint(geom.Pt(100.0, 50.0)), 0, 0)
}

func TestLayoutSquareFlatBoundsSpacing(t *testing.T) {
	l := NewLayout(SquareFlat, LayoutOpts.CellSize(geom.Sz(64.0, 32.0)))
	geom.AssertSize(t, l.CellBounds(), 64.0, 32.0)
	geom.AssertSize(t, l.CellSpacing(), 64.0, 32.0)
}

func TestLayoutHexPointyTopToPoint(t *testing.T) {
	l := NewLayout(HexPointyTop, LayoutOpts.CellSize(geom.Sz(10.0, 10.0)))
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](0, 0)), 0.0, 0.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](1, 0)), geom.Sqrt3*10, 0.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](0, 1)), geom.Sqrt3/2*10, 15.0)
}

func TestLayoutHexFlatTopBounds(t *testing.T) {
	l := NewLayout(HexFlatTop, LayoutOpts.CellSize(geom.Sz(10.0, 10.0)))
	geom.AssertSize(t, l.CellBounds(), 20.0, geom.Sqrt3*10)
	geom.AssertSize(t, l.CellSpacing(), 15.0, geom.Sqrt3*10)
}

func TestLayoutResize(t *testing.T) {
	l := NewLayout(SquareFlat, LayoutOpts.CellSize(geom.Sz(64.0, 32.0))).MoveTo(geom.Pt(10.0, 20.0)).Resize(geom.Sz(128.0, 64.0))
	geom.AssertSize(t, l.CellBounds(), 128.0, 64.0)
	geom.AssertSize(t, l.CellSpacing(), 128.0, 64.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](0, 0)), 10.0, 20.0)
	geom.AssertPoint(t, l.ToPoint(geom.Pt[int](1, 1)), 138.0, 84.0)
	geom.AssertPoint(t, l.FromPoint(geom.Pt(10.0, 20.0)), 0, 0)
	geom.AssertPoint(t, l.FromPoint(geom.Pt(138.0, 84.0)), 1, 1)
}

func TestLayoutRoundTrip(t *testing.T) {
	l := NewLayout(HexPointyTop, LayoutOpts.CellSize(geom.Sz(20.0, 20.0))).MoveTo(geom.Pt(100.0, 100.0))
	idx := geom.Pt[int](3, -2)
	geom.AssertPoint(t, l.FromPoint(l.ToPoint(idx)), idx.X, idx.Y)
}
