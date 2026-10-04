package grid

import (
	"fmt"
	"iter"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/ints"
)

// Array is a 2D array.
type Array[T any] struct {
	size ints.Size
	data []T
}

// Arr makes a new Array from Size.
func Arr[T any](size ints.Size) Array[T] {
	return Array[T]{size, make([]T, size.Area())}
}

// Size returns Array size.
func (a Array[T]) Size() ints.Size {
	return a.size
}

// Width returns Array width.
func (a Array[T]) Width() int {
	return a.size.Width
}

// Height returns Array height.
func (a Array[T]) Height() int {
	return a.size.Height
}

// Len returns Array length.
func (a Array[T]) Len() int {
	return len(a.data)
}

// Has reports whether the index is within bounds.
func (a Array[T]) Has(index ints.Point) bool {
	return a.valid(index)
}

// Get returns a pointer to the value at index. Panics if index is out of bounds.
func (a Array[T]) Get(index ints.Point) *T {
	if !a.valid(index) {
		panic(a.outside(index))
	}

	return a.at(index)
}

// Set stores value at index. Panics if index is out of bounds.
func (a Array[T]) Set(index ints.Point, value T) {
	if !a.valid(index) {
		panic(a.outside(index))
	}

	*a.at(index) = value
}

// Fill sets every element to value.
func (a Array[T]) Fill(value T) {
	for i := range a.data {
		a.data[i] = value
	}
}

// Clear resets every element to the zero value.
func (a Array[T]) Clear() {
	clear(a.data)
}

// Clone returns a deep copy of the array.
func (a Array[T]) Clone() Array[T] {
	data := make([]T, len(a.data))
	copy(data, a.data)

	return Array[T]{a.size, data}
}

// Keys returns iterator over all points.
func (a Array[T]) Keys() iter.Seq[ints.Point] {
	return func(yield func(ints.Point) bool) {
		for y := range a.Height() {
			for x := range a.Width() {
				if !yield(geom.Pt(x, y)) {
					return
				}
			}
		}
	}
}

// Values returns an iterator over all values.
func (a Array[T]) Values() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for i := range a.data {
			if !yield(&a.data[i]) {
				return
			}
		}
	}
}

// All returns an iterator over all (point, value) pairs.
func (a Array[T]) All() iter.Seq2[ints.Point, *T] {
	return func(yield func(ints.Point, *T) bool) {
		for pt := range a.Keys() {
			if !yield(pt, a.at(pt)) {
				return
			}
		}
	}
}

// at returns a pointer to the value at an index the caller knows to be valid.
func (a Array[T]) at(index ints.Point) *T {
	return &a.data[index.X+index.Y*a.size.Width]
}

// outside returns the panic message for an index outside the array. The formatting is kept out
// of Get and Set so they stay small enough to inline.
func (a Array[T]) outside(index ints.Point) string {
	return fmt.Sprintf("grid: index %s outside array of size %s", index, a.size)
}

func (a Array[T]) valid(index ints.Point) bool {
	return index.X >= 0 && index.X < a.Width() && index.Y >= 0 && index.Y < a.Height()
}
