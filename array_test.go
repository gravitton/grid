package grid

import (
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

func TestArray_Constructor(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, a.Size(), geom.Sz(4, 3))
	})
	t.Run("holds a zero value per index", func(t *testing.T) {
		assert.Equal(t, a.Len(), 12)
		assert.Equal(t, *a.Get(geom.Pt(3, 2)), 0)
	})
}

func TestArray_Size(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	assert.Equal(t, a.Size(), geom.Sz(4, 3))
	assert.Equal(t, a.Width(), 4)
	assert.Equal(t, a.Height(), 3)
}

func TestArray_Len(t *testing.T) {
	assert.Equal(t, Arr[int](geom.Sz(4, 3)).Len(), 12)
}

func TestArray_Get(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("returns a pointer to the stored value", func(t *testing.T) {
		a.Set(geom.Pt(1, 2), 42)
		*a.Get(geom.Pt(1, 2))++

		assert.Equal(t, *a.Get(geom.Pt(1, 2)), 43)
	})
	t.Run("panics past the width instead of reading the next row", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			a.Get(geom.Pt(4, 0))
		}, "grid: index (4,0) outside array of size 4x3")
	})
	t.Run("panics before the first column", func(t *testing.T) {
		assert.Panics(t, func() {
			a.Get(geom.Pt(-1, 1))
		})
	})
	t.Run("panics outside the rows", func(t *testing.T) {
		assert.Panics(t, func() {
			a.Get(geom.Pt(0, 3))
		})
	})
}

func TestArray_Set(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("stores the value at the index alone", func(t *testing.T) {
		a.Set(geom.Pt(1, 2), 42)
		a.Set(geom.Pt(0, 0), 7)

		assert.Equal(t, *a.Get(geom.Pt(1, 2)), 42)
		assert.Equal(t, *a.Get(geom.Pt(0, 0)), 7)
	})
	t.Run("panics past the width instead of writing the next row", func(t *testing.T) {
		assert.Panics(t, func() {
			a.Set(geom.Pt(4, 0), 7)
		})
		assert.Equal(t, *a.Get(geom.Pt(0, 1)), 0)
	})
}

func TestArray_Fill(t *testing.T) {
	a := Arr[int](geom.Sz(3, 3))
	a.Fill(9)

	for value := range a.Values() {
		assert.Equal(t, *value, 9)
	}
}

func TestArray_Clear(t *testing.T) {
	a := Arr[int](geom.Sz(3, 3))
	a.Fill(9)
	a.Clear()

	for value := range a.Values() {
		assert.Equal(t, *value, 0)
	}
}

func TestArray_Clone(t *testing.T) {
	original := Arr[int](geom.Sz(3, 3))
	original.Fill(1)

	clone := original.Clone()
	clone.Set(geom.Pt(0, 0), 99)

	t.Run("holds the same values", func(t *testing.T) {
		assert.Equal(t, clone.Size(), original.Size())
		assert.Equal(t, *clone.Get(geom.Pt(2, 2)), 1)
	})
	t.Run("a write to the clone leaves the original alone", func(t *testing.T) {
		assert.Equal(t, *original.Get(geom.Pt(0, 0)), 1)
		assert.Equal(t, *clone.Get(geom.Pt(0, 0)), 99)
	})
}

func TestArray_Keys(t *testing.T) {
	t.Run("yields every index row by row", func(t *testing.T) {
		var indices []geom.Point[int]
		for index := range Arr[int](geom.Sz(3, 2)).Keys() {
			indices = append(indices, index)
		}

		assert.Equal(t, indices, []geom.Point[int]{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(2, 0), geom.Pt(0, 1), geom.Pt(1, 1), geom.Pt(2, 1)})
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range Arr[int](geom.Sz(5, 5)).Keys() {
			count++
			if count == 3 {
				break
			}
		}

		assert.Equal(t, count, 3)
	})
}

func TestArray_Values(t *testing.T) {
	a := Arr[int](geom.Sz(2, 2))

	t.Run("yields a pointer into the array", func(t *testing.T) {
		for value := range a.Values() {
			*value = 5
		}

		assert.Equal(t, *a.Get(geom.Pt(1, 1)), 5)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range a.Values() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
}

func TestArray_All(t *testing.T) {
	a := Arr[int](geom.Sz(2, 2))
	a.Set(geom.Pt(1, 0), 42)

	t.Run("yields every index with its value", func(t *testing.T) {
		count := 0
		for index, value := range a.All() {
			count++
			assert.Equal(t, value, a.Get(index), index.String()+": ")
		}

		assert.Equal(t, count, 4)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range a.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
}

func TestArray_Contains(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("the corners are inside", func(t *testing.T) {
		assert.True(t, a.Contains(geom.Pt(0, 0)))
		assert.True(t, a.Contains(geom.Pt(3, 2)))
	})
	t.Run("a negative coordinate is outside", func(t *testing.T) {
		assert.False(t, a.Contains(geom.Pt(-1, 0)))
		assert.False(t, a.Contains(geom.Pt(0, -1)))
	})
	t.Run("the size itself is outside", func(t *testing.T) {
		assert.False(t, a.Contains(geom.Pt(4, 0)))
		assert.False(t, a.Contains(geom.Pt(0, 3)))
	})
}
