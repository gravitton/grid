package grid

import (
	geom "github.com/gravitton/geometry"
)

// RectOptions configures a grid of square cells. The zero value is the default: the four
// neighbors of [MovementCardinal], and the top-left corner of the bounds at the world origin.
type RectOptions struct {
	// Movement chooses the neighbors of a cell.
	Movement Movement

	// Origin is where the top-left corner of the bounds of the grid is in the world.
	Origin geom.Point[float64]
}

// NewRect returns a grid of square cells of the given size in cells, each cell of the given
// size in the world. Each opts function sets the fields of the [RectOptions] it is given. It
// panics for a movement outside the constants, for a cell size that is not positive and finite
// on both axes, and for an origin with a NaN or infinite coordinate.
func NewRect[T any](size geom.Size[int], cellSize geom.Size[float64], opts ...func(*RectOptions)) *Grid[T] {
	return newRect[T](size, cellSize, kindRect, opts)
}

// NewIsometric returns a grid of square cells drawn as diamonds, of the given size in cells,
// each diamond of the given size in the world, the first cell at the top corner. Each
// opts function sets the fields of the [RectOptions] it is given. It panics for a movement
// outside the constants, for a cell size that is not positive and finite on both axes, and for
// an origin with a NaN or infinite coordinate.
func NewIsometric[T any](size geom.Size[int], cellSize geom.Size[float64], opts ...func(*RectOptions)) *Grid[T] {
	return newRect[T](size, cellSize, kindIsometric, opts)
}

func newRect[T any](size geom.Size[int], cellSize geom.Size[float64], kind kind, opts []func(*RectOptions)) *Grid[T] {
	var options RectOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.Movement.IsNone() {
		panic(options.Movement.unknown())
	}

	return newGrid[T](
		size,
		newLayout(kind, size, cellSize, options.Origin),
		newSquareTopology(options.Movement),
	)
}

// RectCellSize returns the cell size of NewRect for a square of the given width.
func RectCellSize(width float64) geom.Size[float64] {
	return geom.SzU(width)
}

// IsometricRectCellSize returns the cell size of NewIsometric for a diamond of the given width
// whose edges rise at 30°, the angle of a true isometric projection.
func IsometricRectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width, width*geom.Sqrt3/3)
}

// IsometricPixelPerfectRectCellSize returns the cell size of NewIsometric for a diamond of the
// given width and half of it tall, whose edges step two pixels across for each one up.
func IsometricPixelPerfectRectCellSize(width float64) geom.Size[float64] {
	return geom.Sz(width, width*0.5)
}

// squareTopology is the neighbor rule of square cells under a movement, the methods of Point
// with that movement.
type squareTopology struct {
	movement Movement

	// steps holds the neighbor offsets of the movement, the same from every index.
	steps []geom.Vector[int]
}

func newSquareTopology(movement Movement) *squareTopology {
	return &squareTopology{
		movement: movement,
		steps:    movement.Offsets(),
	}
}

func (t *squareTopology) offsets(geom.Point[int]) []geom.Vector[int] {
	return t.steps
}

func (t *squareTopology) distance(from, to geom.Point[int]) int {
	return Point(from).DistanceTo(to, t.movement)
}

func (t *squareTopology) rangeLen(radius int) int {
	return RangeLen(radius, t.movement)
}

func (t *squareTopology) appendWithin(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return Point(index).AppendRange(dst, radius, t.movement)
}

func (t *squareTopology) ringLen(radius int) int {
	return RingLen(radius, t.movement)
}

func (t *squareTopology) appendRing(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return Point(index).AppendRing(dst, radius, t.movement)
}

func (t *squareTopology) appendSpiral(dst []geom.Point[int], index geom.Point[int], radius int) []geom.Point[int] {
	return Point(index).AppendSpiral(dst, radius, t.movement)
}

func (t *squareTopology) appendLine(dst []geom.Point[int], from, to geom.Point[int]) []geom.Point[int] {
	return Point(from).AppendLine(dst, to, t.movement)
}

func (t *squareTopology) sight(center geom.Point[int], _ int, candidates []geom.Point[int], marks *sightMarks) {
	blocks := blockers{blocked: marks.blocks}
	for _, candidate := range candidates {
		if Point(center).sees(candidate, blocks) {
			marks.see(candidate)
		}
	}
}
