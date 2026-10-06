package grid

import (
	"slices"
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

		for index, value := range a.All() {
			assert.Equal(t, *value, 0, index.String()+": ")
		}
	})
	t.Run("a size without columns or rows is empty", func(t *testing.T) {
		for _, size := range []geom.Size[int]{geom.Sz(0, 0), geom.Sz(4, 0), geom.Sz(0, 3)} {
			empty := Arr[int](size)

			assert.Equal(t, empty.Size(), size, size.String()+": ")
			assert.Equal(t, empty.Len(), 0, size.String()+": ")
			assert.False(t, empty.Contains(geom.Pt(0, 0)), size.String()+": ")
		}
	})
	t.Run("panics for a negative width or height", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Arr[int](geom.Sz(-1, 5))
		}, "grid: array of negative size -1x5")
		assert.PanicsWith(t, func() {
			Arr[int](geom.Sz(4, -3))
		}, "grid: array of negative size 4x-3")
	})
	t.Run("panics for two negatives, whose area is positive", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Arr[int](geom.Sz(-2, -3))
		}, "grid: array of negative size -2x-3")
	})
	t.Run("the zero value is the empty array", func(t *testing.T) {
		var zero Array[int]

		assert.Equal(t, zero.Len(), 0)
		assert.False(t, zero.Contains(geom.Pt(0, 0)))

		for range zero.All() {
			assert.Fail(t, "the zero array yielded an index")
		}
	})
	t.Run("allocates its values once", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkArray = Arr[int](geom.Sz(4, 3))
		}), 1.0)
	})
}

func TestArray_Size(t *testing.T) {
	assert.Equal(t, Arr[int](geom.Sz(4, 3)).Size(), geom.Sz(4, 3))
}

func TestArray_Width(t *testing.T) {
	assert.Equal(t, Arr[int](geom.Sz(4, 3)).Width(), 4)
}

func TestArray_Height(t *testing.T) {
	assert.Equal(t, Arr[int](geom.Sz(4, 3)).Height(), 3)
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

	for index, value := range a.All() {
		assert.Equal(t, *value, 9, index.String()+": ")
	}
}

func TestArray_Clear(t *testing.T) {
	a := Arr[int](geom.Sz(3, 3))
	a.Fill(9)
	a.Clear()

	for index, value := range a.All() {
		assert.Equal(t, *value, 0, index.String()+": ")
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
	t.Run("a plain copy shares the values, as a slice does", func(t *testing.T) {
		shared := original
		shared.Set(geom.Pt(1, 1), 7)

		assert.Equal(t, *original.Get(geom.Pt(1, 1)), 7)
	})
	t.Run("allocates its values once", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkArray = original.Clone()
		}), 1.0)
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
	t.Run("allocates nothing", func(t *testing.T) {
		a := Arr[int](geom.Sz(8, 8))

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			for index := range a.Keys() {
				sinkPoint = index
			}
		}), 0.0)
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
	t.Run("follows the order of Keys", func(t *testing.T) {
		var values, expected []*int
		for value := range a.Values() {
			values = append(values, value)
		}
		for index := range a.Keys() {
			expected = append(expected, a.Get(index))
		}

		assert.Equal(t, values, expected)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range a.Values() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			for value := range a.Values() {
				sinkBool = *value == 0
			}
		}), 0.0)
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
	t.Run("follows the order of Keys", func(t *testing.T) {
		var indices []geom.Point[int]
		for index := range a.All() {
			indices = append(indices, index)
		}

		assert.Equal(t, indices, slices.Collect(a.Keys()))
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			for index, value := range a.All() {
				sinkPoint, sinkBool = index, *value == 0
			}
		}), 0.0)
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
