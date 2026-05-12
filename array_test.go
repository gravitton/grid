package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

func TestArr_Size(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))
	assert.Equal(t, a.Size(), geom.Sz(4, 3))
	assert.Equal(t, a.Width(), 4)
	assert.Equal(t, a.Height(), 3)
	assert.Equal(t, a.Len(), 12)
}

func TestArray_Has(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))
	assert.True(t, a.Has(geom.Pt(0, 0)))
	assert.True(t, a.Has(geom.Pt(3, 2)))
	assert.False(t, a.Has(geom.Pt(-1, 0)))
	assert.False(t, a.Has(geom.Pt(0, -1)))
	assert.False(t, a.Has(geom.Pt(4, 0)))
	assert.False(t, a.Has(geom.Pt(0, 3)))
}

func TestArray_SetGet(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))
	a.Set(geom.Pt(1, 2), 42)
	assert.Equal(t, *a.Get(geom.Pt(1, 2)), 42)
	a.Set(geom.Pt(0, 0), 7)
	assert.Equal(t, *a.Get(geom.Pt(0, 0)), 7)
}

func TestArray_Fill(t *testing.T) {
	a := Arr[int](geom.Sz(3, 3))
	a.Fill(9)
	for y := range 3 {
		for x := range 3 {
			assert.Equal(t, *a.Get(geom.Pt(x, y)), 9)
		}
	}
}

func TestArray_Clear(t *testing.T) {
	a := Arr[int](geom.Sz(3, 3))
	a.Fill(9)
	a.Clear()
	for y := range 3 {
		for x := range 3 {
			assert.Equal(t, *a.Get(geom.Pt(x, y)), 0)
		}
	}
}

func TestArray_Clone_Independence(t *testing.T) {
	a1 := Arr[int](geom.Sz(3, 3))
	a1.Fill(1)
	a2 := a1.Clone()
	a2.Set(geom.Pt(0, 0), 99)
	// Original is unchanged.
	assert.Equal(t, *a1.Get(geom.Pt(0, 0)), 1)
	assert.Equal(t, *a2.Get(geom.Pt(0, 0)), 99)
}

func TestArray_Iter_Order(t *testing.T) {
	a := Arr[int](geom.Sz(3, 2))
	var points []ints.Point
	for p := range a.Keys() {
		points = append(points, p)
	}
	assert.Equal(t, len(points), 6)
	// Row-major: (0,0), (1,0), (2,0), (0,1), (1,1), (2,1)
	assert.Equal(t, points[0], geom.Pt(0, 0))
	assert.Equal(t, points[1], geom.Pt(1, 0))
	assert.Equal(t, points[2], geom.Pt(2, 0))
	assert.Equal(t, points[3], geom.Pt(0, 1))
}

func TestArray_Iter2(t *testing.T) {
	a := Arr[int](geom.Sz(2, 2))
	a.Set(geom.Pt(1, 0), 42)
	count := 0
	found := false
	for p, v := range a.All() {
		count++
		if p == geom.Pt(1, 0) {
			assert.Equal(t, *v, 42)
			found = true
		}
	}
	assert.Equal(t, count, 4)
	assert.True(t, found)
}

func TestArray_Iter_EarlyStop(t *testing.T) {
	a := Arr[int](geom.Sz(5, 5))
	count := 0
	for range a.Keys() {
		count++
		if count == 3 {
			break
		}
	}
	assert.Equal(t, count, 3)
}
