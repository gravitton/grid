package grid

import (
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

// Has returns if index is valid.
func (a Array[T]) Has(index ints.Point) bool {
	return a.valid(index)
}

// Get value from given index
func (a Array[T]) Get(index ints.Point) *T {
	return &a.data[a.index(index)]
}

// Set value to given index
func (a Array[T]) Set(index ints.Point, value T) {
	a.data[a.index(index)] = value
}

// Fill all values with given value
func (a Array[T]) Fill(value T) {
	for i := 0; i < len(a.data); i++ {
		a.data[i] = value
	}
}

// Clear all values
func (a Array[T]) Clear() {
	clear(a.data)
}

// Clone returns a deep copy of the array.
func (a Array[T]) Clone() Array[T] {
	data := make([]T, len(a.data))
	copy(data, a.data)

	return Array[T]{a.size, data}
}

// Iter returns iterator over all points.
func (a Array[T]) Iter() iter.Seq[ints.Point] {
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

// Iter2 returns an iterator over all (point, value) pairs.
func (a Array[T]) Iter2() iter.Seq2[ints.Point, *T] {
	return func(yield func(ints.Point, *T) bool) {
		for pt := range a.Iter() {
			if !yield(pt, &a.data[a.index(pt)]) {
				return
			}
		}
	}
}

func (a Array[T]) index(index ints.Point) int {
	return index.X + index.Y*a.size.Width
}

func (a Array[T]) valid(index ints.Point) bool {
	return index.X >= 0 && index.X < a.Width() && index.Y >= 0 && index.Y < a.Height()
}
