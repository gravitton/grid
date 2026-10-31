package grid

import (
	"fmt"

	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"
)

// HexLayout places hexagonal cells on an unbounded plane with no grid around them: the
// center, the box and the polygon of any index, and the index under any point. An index is in
// the coordinate system the layout is built with, any of the seven of [hex.CoordinateSystem],
// so an axial layout takes [hex.Hex.Point] and gives back what [hex.Axial] turns into a hex. A
// cell is placed exactly where a hexagonal [Grid] of the same cell scale places it. It is built
// by [NewHexFlatTopLayout] or [NewHexPointyTopLayout]; the zero value is no layout, and its
// methods are not to be called.
type HexLayout struct {
	layout layout
	system hex.CoordinateSystem
}

// NewHexFlatTopLayout returns a layout of flat-top hexagons of the given cell scale, indexed in
// the system, with the center of the cell at index zero at the world origin. It panics for a
// cell scale that is not positive and finite on both axes, and for a system outside the seven.
func NewHexFlatTopLayout(cellScale geom.Size[float64], system hex.CoordinateSystem) *HexLayout {
	return newHexLayout(kindHexFlatTop, cellScale, system)
}

// NewHexPointyTopLayout returns a layout of pointy-top hexagons of the given cell scale,
// indexed in the system, with the center of the cell at index zero at the world origin. It
// panics for a cell scale that is not positive and finite on both axes, and for a system
// outside the seven.
func NewHexPointyTopLayout(cellScale geom.Size[float64], system hex.CoordinateSystem) *HexLayout {
	return newHexLayout(kindHexPointyTop, cellScale, system)
}

// newHexLayout returns a layout of hexagons of the kind, indexed in the system.
func newHexLayout(kind kind, cellScale geom.Size[float64], system hex.CoordinateSystem) *HexLayout {
	if system.IsNone() {
		panic(fmt.Errorf("grid: unknown coordinate system %d", system))
	}

	return &HexLayout{newLayout(kind, geom.Size[int]{}, cellScale, geom.Point[float64]{}), system}
}

// CellSize returns the size of a cell as drawn, the width and height of the box around it.
func (l *HexLayout) CellSize() geom.Size[float64] {
	return l.layout.cellSize
}

// CellSpacing returns the distance between the centers of adjacent cells along each axis.
func (l *HexLayout) CellSpacing() geom.Size[float64] {
	return l.layout.cellSpacing
}

// Center returns the center of the cell at the index in the world.
func (l *HexLayout) Center(index geom.Point[int]) geom.Point[float64] {
	return l.layout.latticeToWorld(l.position(index))
}

// Bounds returns the box around the cell at the index as drawn, in the world.
func (l *HexLayout) Bounds(index geom.Point[int]) geom.Box[float64] {
	return l.layout.cellBoundsAt(l.position(index))
}

// Polygon returns the hexagon of the cell at the index in the world.
func (l *HexLayout) Polygon(index geom.Point[int]) geom.RegularPolygon[float64] {
	polygon := l.layout.cellPolygon
	polygon.Center = l.Center(index)

	return polygon
}

// position returns where the center of the cell at the index is on the lattice.
func (l *HexLayout) position(index geom.Point[int]) geom.Point[int] {
	return l.layout.kind.hexToLattice(l.system.From(index))
}

// IndexAt returns the index of the cell the point of the world lies in. A point on an edge
// goes to the same side in every cell. It panics for a point with a NaN or infinite
// coordinate.
func (l *HexLayout) IndexAt(point geom.Point[float64]) geom.Point[int] {
	return l.system.To(l.layout.kind.hexFromLattice(l.layout.latticeFromWorld(point)).Round())
}

// MoveTo places the center of the cell at index zero at origin. It panics for an origin with
// a NaN or infinite coordinate.
func (l *HexLayout) MoveTo(origin geom.Point[float64]) {
	l.layout.moveTo(origin)
}

// Resize replaces the cell scale the constructor took, the center of the cell at index zero
// staying where it is. It panics for a cell scale that is not positive and finite on both
// axes.
func (l *HexLayout) Resize(cellScale geom.Size[float64]) {
	l.layout.resize(cellScale)
}
