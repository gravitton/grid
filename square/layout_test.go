package square

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
)

var (
	layoutFlat               = LayoutFlat(geom.Sz(10.0, 10.0), geom.Pt(0.0, 0.0))
	layoutFlatWithOrigin     = LayoutFlat(geom.Sz(10.0, 10.0), geom.Pt(-100.0, 50.0))
	layoutFlatWithOriginSkew = LayoutFlat(geom.Sz(10.0, 8.0), geom.Pt(-100.0, 50.0))

	layoutIsometric               = LayoutIsometric(geom.Sz(10.0, 10.0), geom.Pt(0.0, 0.0))
	layoutIsometricWithOrigin     = LayoutIsometric(geom.Sz(10.0, 10.0), geom.Pt(-100.0, 50.0))
	layoutIsometricWithOriginSkew = LayoutIsometric(geom.Sz(10.0, 8.0), geom.Pt(-100.0, 50.0))

	// Classic ~30° isometric: size.W = 2 * size.H produces the 2:1 diamond.
	layoutIsometricClassic = LayoutIsometric(geom.Sz(20.0, 10.0), geom.Pt(0.0, 0.0))
)

func TestLayout_Constructors(t *testing.T) {
	size := geom.Sz(10.0, 8.0)
	origin := geom.Pt(5.0, -3.0)

	flat := LayoutFlat(size, origin)
	assert.Equal(t, flat.Size, size)
	assert.Equal(t, flat.Origin, origin)

	iso := LayoutIsometric(size, origin)
	assert.Equal(t, iso.Size, size)
	assert.Equal(t, iso.Origin, origin)
}

func TestLayout_Bounds(t *testing.T) {
	geom.AssertSize(t, layoutFlat.Bounds(), 10.0, 10.0)
	geom.AssertSize(t, layoutFlatWithOrigin.Bounds(), 10.0, 10.0)
	geom.AssertSize(t, layoutFlatWithOriginSkew.Bounds(), 10.0, 8.0)

	geom.AssertSize(t, layoutIsometric.Bounds(), 10.0, 10.0)
	geom.AssertSize(t, layoutIsometricWithOrigin.Bounds(), 10.0, 10.0)
	geom.AssertSize(t, layoutIsometricWithOriginSkew.Bounds(), 10.0, 8.0)
}

func TestLayout_Spacing(t *testing.T) {
	geom.AssertSize(t, layoutFlat.Spacing(), 10.0, 10.0)
	geom.AssertSize(t, layoutFlatWithOrigin.Spacing(), 10.0, 10.0)
	geom.AssertSize(t, layoutFlatWithOriginSkew.Spacing(), 10.0, 8.0)

	geom.AssertSize(t, layoutIsometric.Spacing(), 5.0, 5.0)
	geom.AssertSize(t, layoutIsometricWithOrigin.Spacing(), 5.0, 5.0)
	geom.AssertSize(t, layoutIsometricWithOriginSkew.Spacing(), 5.0, 4.0)
}

func TestLayout_ToPoint(t *testing.T) {
	tests := []struct {
		layout   Layout
		index    [2]int
		expected floats.Point
	}{
		// flat
		{layoutFlat, [2]int{0, 0}, geom.Pt(0.0, 0.0)},
		{layoutFlat, [2]int{1, 0}, geom.Pt(10.0, 0.0)},
		{layoutFlat, [2]int{0, 1}, geom.Pt(0.0, 10.0)},
		{layoutFlat, [2]int{2, 3}, geom.Pt(20.0, 30.0)},
		{layoutFlat, [2]int{-1, -2}, geom.Pt(-10.0, -20.0)},

		{layoutFlatWithOrigin, [2]int{0, 0}, geom.Pt(-100.0, 50.0)},
		{layoutFlatWithOrigin, [2]int{1, 1}, geom.Pt(-90.0, 60.0)},

		{layoutFlatWithOriginSkew, [2]int{0, 0}, geom.Pt(-100.0, 50.0)},
		{layoutFlatWithOriginSkew, [2]int{1, 1}, geom.Pt(-90.0, 58.0)},

		// isometric: screen_x = (col-row)*0.5*W, screen_y = (col+row)*0.5*H
		{layoutIsometric, [2]int{0, 0}, geom.Pt(0.0, 0.0)},
		{layoutIsometric, [2]int{1, 0}, geom.Pt(5.0, 5.0)},
		{layoutIsometric, [2]int{0, 1}, geom.Pt(-5.0, 5.0)},
		{layoutIsometric, [2]int{1, 1}, geom.Pt(0.0, 10.0)},
		{layoutIsometric, [2]int{2, 1}, geom.Pt(5.0, 15.0)},

		{layoutIsometricWithOrigin, [2]int{0, 0}, geom.Pt(-100.0, 50.0)},
		{layoutIsometricWithOrigin, [2]int{1, 0}, geom.Pt(-95.0, 55.0)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%d_%d", layoutName(test.layout), test.index[0], test.index[1]), func(t *testing.T) {
			idx := geom.Pt(test.index[0], test.index[1])
			actual := test.layout.ToPoint(idx)

			geom.AssertPoint(t, actual, test.expected.X, test.expected.Y)

			// round-trip: FromPoint → ToPoint
			roundTrip := test.layout.FromPoint(actual).Round().Int()
			assert.Equal(t, roundTrip, idx)
		})
	}
}

func TestLayout_FromPoint_OffCenter(t *testing.T) {
	// Origin=(0,0) means center of cell (0,0) is at (0,0).
	// Pixels anywhere within the cell should round back to the same index.
	l := layoutFlat // size 10×10, origin (0,0) = center of cell (0,0)

	for _, tc := range []struct {
		pixel    floats.Point
		expected [2]int
	}{
		{geom.Pt(0.0, 0.0), [2]int{0, 0}},   // exact center
		{geom.Pt(-4.9, -4.9), [2]int{0, 0}}, // near top-left corner
		{geom.Pt(4.9, 4.9), [2]int{0, 0}},   // near bottom-right corner
		{geom.Pt(5.1, 0.0), [2]int{1, 0}},   // just past right boundary
		{geom.Pt(0.0, 5.1), [2]int{0, 1}},   // just past bottom boundary
		{geom.Pt(10.0, 0.0), [2]int{1, 0}},  // center of cell (1,0)
	} {
		t.Run(fmt.Sprintf("%.1f_%.1f", tc.pixel.X, tc.pixel.Y), func(t *testing.T) {
			got := l.FromPoint(tc.pixel).Round().Int()
			assert.Equal(t, got, geom.Pt(tc.expected[0], tc.expected[1]))
		})
	}
}

func TestLayout_CellPolygon(t *testing.T) {
	// flat: FlatTop square (4 sides, 45° vertices), circumradius = size * sqrt2/2
	// so the vertex-to-vertex diagonal equals the cell's pixel width/height.
	geom.AssertRegularPolygon(t, layoutFlat.CellPolygon(geom.Pt(0, 0)), 0.0, 0.0, 7.071068, 7.071068, 4, 45*geom.DegToRad)
	geom.AssertRegularPolygon(t, layoutFlat.CellPolygon(geom.Pt(1, 0)), 10.0, 0.0, 7.071068, 7.071068, 4, 45*geom.DegToRad)

	// isometric: perfect diamond — rx = size.W * 0.5, ry = size.H * 0.5.
	// Aspect ratio is controlled by size, not the transform.
	geom.AssertRegularPolygon(t, layoutIsometric.CellPolygon(geom.Pt(0, 0)), 0.0, 0.0, 5.0, 5.0, 4, 90*geom.DegToRad)
	geom.AssertRegularPolygon(t, layoutIsometric.CellPolygon(geom.Pt(1, 0)), 5.0, 5.0, 5.0, 5.0, 4, 90*geom.DegToRad)

	// classic 2:1 isometric via size skew (20×10): rx=10, ry=5.
	geom.AssertRegularPolygon(t, layoutIsometricClassic.CellPolygon(geom.Pt(0, 0)), 0.0, 0.0, 10.0, 5.0, 4, 90*geom.DegToRad)
	geom.AssertRegularPolygon(t, layoutIsometricClassic.CellPolygon(geom.Pt(1, 0)), 10.0, 5.0, 10.0, 5.0, 4, 90*geom.DegToRad)
}

func layoutName(layout Layout) string {
	name := []string{}

	switch layout.transform.orientation {
	case geom.FlatTop:
		name = append(name, "flat")
	case geom.PointyTop:
		name = append(name, "isometric")
	}

	if !layout.Origin.IsZero() {
		name = append(name, "translated")
	}

	if layout.Size.Width != layout.Size.Height {
		name = append(name, "skewed")
	}

	return strings.Join(name, "-")
}
