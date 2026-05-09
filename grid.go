package grid

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/geometry/types/floats"
	"github.com/gravitton/geometry/types/ints"
)

// Grid is a 2D grid.
type Grid[T any] struct {
	cells       Array[T]
	cellSize    floats.Size
	cellSpacing floats.Size
	bounds      floats.Rectangle
	kind        kind

	toPoint   func(index ints.Point) floats.Point
	fromPoint func(point floats.Point) ints.Point
	polygon   func(index ints.Point) floats.RegularPolygon
	distance  func(from, to ints.Point) int
	toRange   func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point
	neighbors func(index ints.Point) []ints.Vector
}

// ValidIndexFunc is a function that returns true if an index is valid.
type ValidIndexFunc func(index ints.Point) bool

// ValidFunc is a function that returns true if a cell is valid.
type ValidFunc[T any] func(current *Cell[T]) bool

// CostFunc is a function that returns the cost of moving from one cell to another.
type CostFunc[T any] func(current, next *Cell[T]) float64

type kind int

const (
	kindDefault kind = iota
	kindIsometric
	kindHexagonalFlatTop
	kindHexagonalPointyTop
)

// NewGrid creates a new grid.
func NewGrid[T any](
	layout Layout,
	distance func(from, to ints.Point) int,
	toRange func(index ints.Point, n int, valid ValidIndexFunc) []ints.Point,
	neighbors func(index ints.Point) []ints.Vector,
) *Grid[T] {
	return &Grid[T]{
		cells:       Arr[T](layout.Size()),
		cellSize:    layout.CellBounds(),
		cellSpacing: layout.CellSpacing(),
		bounds:      layout.Bounds(),
		kind:        layout.kind(),
		toPoint:     layout.ToPoint,
		fromPoint:   layout.FromPoint,
		polygon:     layout.CellPolygon,
		distance:    distance,
		toRange:     toRange,
		neighbors:   neighbors,
	}
}

// Size returns the size of the grid.
func (g *Grid[T]) Size() ints.Size {
	return g.cells.Size()
}

// Width returns the width of the grid.
func (g *Grid[T]) Width() int {
	return g.cells.Width()
}

// Height returns the height of the grid.
func (g *Grid[T]) Height() int {
	return g.cells.Height()
}

// Bounds returns the bounds of the grid.
func (g *Grid[T]) Bounds() floats.Rectangle {
	return g.bounds
}

// CellBounds returns the size of a cell.
func (g *Grid[T]) CellBounds() floats.Size {
	return g.cellSize
}

// CellSpacing returns the spacing between cell centers.
func (g *Grid[T]) CellSpacing() floats.Size {
	return g.cellSpacing
}

// CellWidth returns the width of a cell.
func (g *Grid[T]) CellWidth() float64 {
	return g.cellSize.Width
}

// CellHeight returns the height of a cell.
func (g *Grid[T]) CellHeight() float64 {
	return g.cellSize.Height
}

// Get returns a cell.
func (g *Grid[T]) Get(index ints.Point) *Cell[T] {
	return g.cell(index)
}

// Has reports whether the index is within the grid bounds.
func (g *Grid[T]) Has(index ints.Point) bool {
	return g.valid(index)
}

// Set stores value at the given index. Out-of-bounds writes are no-ops.
func (g *Grid[T]) Set(index ints.Point, value T) {
	g.setData(index, value)
}

// Fill sets every cell to value.
func (g *Grid[T]) Fill(value T) {
	g.cells.Fill(value)
}

// Clear resets every cell to the zero value.
func (g *Grid[T]) Clear() {
	g.cells.Clear()
}

// Clone returns a deep copy of the grid with the same layout and cell values.
func (g *Grid[T]) Clone() *Grid[T] {
	clone := *g
	clone.cells = g.cells.Clone()

	return &clone
}

// IndexAt returns the index of the cell at a given point.
func (g *Grid[T]) IndexAt(point floats.Point) ints.Point {
	return g.fromPoint(point)
}

// At returns the cell at a given point.
func (g *Grid[T]) At(point floats.Point) *Cell[T] {
	return g.cell(g.fromPoint(point))
}

func (g *Grid[T]) valid(index ints.Point) bool {
	return g.cells.valid(index)
}

func (g *Grid[T]) getData(index ints.Point) *T {
	if !g.cells.valid(index) {
		return nil
	}
	return g.cells.Get(index)
}

func (g *Grid[T]) setData(index ints.Point, value T) {
	if !g.cells.valid(index) {
		return
	}
	g.cells.Set(index, value)
}

func (g *Grid[T]) cell(index ints.Point) *Cell[T] {
	return &Cell[T]{g, index}
}

func (g *Grid[T]) cellCenter(index ints.Point) floats.Point {
	return g.toPoint(index)
}

func (g *Grid[T]) cellBounds(index ints.Point) floats.Rectangle {
	return geom.Rect(g.cellCenter(index), g.cellSize)
}

func (g *Grid[T]) cellPolygon(index ints.Point) floats.RegularPolygon {
	return g.polygon(index)
}

func (g *Grid[T]) cellNeighbors(index ints.Point) []ints.Point {
	var neighbors []ints.Point
	for _, dir := range g.neighbors(index) {
		next := index.Add(dir)
		if g.valid(next) {
			neighbors = append(neighbors, next)
		}
	}
	return neighbors
}
