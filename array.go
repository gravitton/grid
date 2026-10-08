package grid

import (
	"encoding/json"
	"fmt"
	"iter"
	"math"

	geom "github.com/gravitton/geometry"
)

// Array is a table of values of type T, one for every index of a size in columns and rows,
// stored row by row. It is the storage under a [Grid] and serves alone where a table needs no
// layout. Like a slice it is a handle: a copy of an Array shares its values, and Clone is what
// gives storage of its own. The zero value is the empty array.
type Array[T any] struct {
	size geom.Size[int]
	data []T
}

// Arr returns an array of the given size, every value zero. It panics for a negative width or
// height, and for a size of more values than an int counts.
func Arr[T any](size geom.Size[int]) Array[T] {
	if err := validateSize(size); err != nil {
		panic(err.Error())
	}

	return Array[T]{size, make([]T, size.Width*size.Height)}
}

// validateSize returns the error of a size no array has: a negative width or height, or a
// width and a height whose product is past the largest int and would wrap to a length that is
// not theirs.
func validateSize(size geom.Size[int]) error {
	if size.Width < 0 || size.Height < 0 {
		return fmt.Errorf("grid: array of negative size %s", size)
	}
	if size.Height != 0 && size.Width > math.MaxInt/size.Height {
		return fmt.Errorf("grid: array of size %s is too large", size)
	}

	return nil
}

// Size returns the number of columns and rows of the array.
func (a Array[T]) Size() geom.Size[int] {
	return a.size
}

// Width returns the number of columns of the array.
func (a Array[T]) Width() int {
	return a.size.Width
}

// Height returns the number of rows of the array.
func (a Array[T]) Height() int {
	return a.size.Height
}

// Len returns the number of values of the array, its width times its height.
func (a Array[T]) Len() int {
	return len(a.data)
}

// Get returns a pointer to the value at index, into the array. It panics for an index outside
// the array.
func (a Array[T]) Get(index geom.Point[int]) *T {
	if !a.Contains(index) {
		panic(a.outside(index))
	}

	return a.at(index)
}

// Set stores value at index. It panics for an index outside the array.
func (a Array[T]) Set(index geom.Point[int], value T) {
	if !a.Contains(index) {
		panic(a.outside(index))
	}

	*a.at(index) = value
}

// Fill sets every value of the array to value.
func (a Array[T]) Fill(value T) {
	for i := range a.data {
		a.data[i] = value
	}
}

// Clear resets every value of the array to the zero value.
func (a Array[T]) Clear() {
	clear(a.data)
}

// Clone returns a copy of the array with storage of its own: the same values, copied as an
// assignment copies them, so a write to one array leaves the other alone.
func (a Array[T]) Clone() Array[T] {
	data := make([]T, len(a.data))
	copy(data, a.data)

	return Array[T]{a.size, data}
}

// Keys returns an iterator over every index of the array, row by row.
func (a Array[T]) Keys() iter.Seq[geom.Point[int]] {
	return func(yield func(geom.Point[int]) bool) {
		for y := range a.Height() {
			for x := range a.Width() {
				if !yield(geom.Pt(x, y)) {
					return
				}
			}
		}
	}
}

// Values returns an iterator over a pointer to every value, into the array, in the order of
// Keys.
func (a Array[T]) Values() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for i := range a.data {
			if !yield(&a.data[i]) {
				return
			}
		}
	}
}

// All returns an iterator over every index with a pointer to its value, into the array, in the
// order of Keys.
func (a Array[T]) All() iter.Seq2[geom.Point[int], *T] {
	return func(yield func(geom.Point[int], *T) bool) {
		for index := range a.Keys() {
			if !yield(index, a.at(index)) {
				return
			}
		}
	}
}

// at returns a pointer to the value at an index the caller knows to be inside the array.
func (a Array[T]) at(index geom.Point[int]) *T {
	return &a.data[index.X+index.Y*a.size.Width]
}

// outside returns the panic message Get and Set share for an index outside the array.
func (a Array[T]) outside(index geom.Point[int]) string {
	return fmt.Sprintf("grid: index %s outside array of size %s", index, a.size)
}

// Contains reports whether the index is inside the array: a column and a row it has.
func (a Array[T]) Contains(index geom.Point[int]) bool {
	return index.X >= 0 && index.X < a.size.Width && index.Y >= 0 && index.Y < a.size.Height
}

// MarshalJSON implements json.Marshaler with the size of the array, flat as a [geom.Size]
// writes it, and its values row by row. The zero value writes null for its values, an array
// made by Arr an empty array.
func (a Array[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Size   geom.Size[int] `json:",embed"`
		Values []T            `json:"d"`
	}{a.size, a.data})
}

// UnmarshalJSON implements json.Unmarshaler. It returns an error for a negative width or height
// and for a number of values other than the width times the height.
func (a *Array[T]) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Size   geom.Size[int] `json:",embed"`
		Values []T            `json:"d"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	size := decoded.Size
	if err := validateSize(size); err != nil {
		return err
	}

	if len(decoded.Values) != size.Width*size.Height {
		return fmt.Errorf("grid: array of size %s with %d values", size, len(decoded.Values))
	}

	*a = Array[T]{size, decoded.Values}

	return nil
}
