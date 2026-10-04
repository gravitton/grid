package grid

import (
	"math"

	geom "github.com/gravitton/geometry"
)

// RectOptions configures a grid of square cells. The zero value is the default: the four
// neighbors of [MovementCardinal], and the top-left corner of the grid at the world origin.
type RectOptions struct {
	// Movement chooses the neighbors of a cell.
	Movement Movement

	// Origin is the world position of the top-left corner of the bounds of the grid.
	Origin geom.Point[float64]
}

// NewRect returns a rectangular grid of the given size in cells, each cell of the given size
// in the world. Each opts function sets the fields of the [RectOptions] it is given.
func NewRect[T any](grid geom.Size[int], size geom.Size[float64], opts ...func(*RectOptions)) *Grid[T] {
	return newRect[T](grid, size, squareFlat, opts)
}

// NewIsometric returns an isometric grid, its square cells drawn as diamonds, of the given
// size in cells, each diamond of the given size in the world. Each opts function sets the
// fields of the [RectOptions] it is given.
func NewIsometric[T any](grid geom.Size[int], size geom.Size[float64], opts ...func(*RectOptions)) *Grid[T] {
	return newRect[T](grid, size, squareIsometric, opts)
}

func newRect[T any](grid geom.Size[int], size geom.Size[float64], transform *transform, opts []func(*RectOptions)) *Grid[T] {
	var options RectOptions
	for _, opt := range opts {
		opt(&options)
	}

	movement := options.Movement
	offsets := movement.Offsets()

	return newGrid[T](
		newLayout(transform, grid, size, nil, nil).AlignTopLeft().Translate(options.Origin.Vector()),
		func(from, to geom.Point[int]) int {
			return Point(from).DistanceTo(to, movement)
		},
		func(index geom.Point[int], n int, passable func(index geom.Point[int]) bool) []geom.Point[int] {
			p := Point(index)

			var candidates, blocking []geom.Point[int]
			for _, candidate := range p.Range(n, movement) {
				if passable(candidate) {
					candidates = append(candidates, candidate)
				} else {
					blocking = append(blocking, candidate)
				}
			}

			return p.FieldOfView(candidates, blocking)
		},
		func(index geom.Point[int]) []geom.Vector[int] {
			return offsets
		},
	)
}

// RectCellSize returns the cell size for NewRect where each tile is exactly
// width pixels wide and width pixels tall (1:1 ratio).
func RectCellSize(width float64) geom.Size[float64] {
	return geom.SzU(width)
}

// IsometricRectCellSize returns the cell size for NewIsometric where each
// diamond tile is width pixels wide at a geometrically accurate 30° isometric angle
// (height = width·tan30° ≈ width·0.577).
// For a pixel-art-friendly 2:1 ratio use IsometricPixelPerfectRectCellSize.
func IsometricRectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width, width*math.Tan(geom.ToRadians(30)))
}

// IsometricPixelPerfectRectCellSize returns the cell size for NewIsometric
// where each diamond tile is width pixels wide and width/2 pixels tall (2:1 ratio).
// The 2:1 ratio is the classic pixel-art isometric convention — tile edges land on
// exact pixel boundaries for crisp rendering.
// For a geometrically accurate 30° angle use IsometricRectCellSize.
func IsometricPixelPerfectRectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width, width*0.5)
}
