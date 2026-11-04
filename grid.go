package grid

import (
	"encoding/json"
	"fmt"
	"iter"
	"math"
	"slices"

	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

// SquareOptions configures a grid of square cells. The zero value is the default: the four
// neighbors of [MovementCardinal], and the top-left corner of the bounds at the world origin.
type SquareOptions struct {
	// Movement chooses the neighbors of a cell.
	Movement Movement

	// Origin is where the top-left corner of the bounds of the grid is in the world.
	Origin geom.Point[float64]
}

// HexOptions configures a grid of hexagonal cells. The zero value is the default: the odd
// columns or rows shifted, and the top-left corner of the bounds at the world origin.
type HexOptions struct {
	// EvenOffset shifts the even columns of a flat-top grid, or the even rows of a pointy-top
	// one, in place of the odd ones.
	EvenOffset bool

	// Origin is where the top-left corner of the bounds of the grid is in the world.
	Origin geom.Point[float64]
}

// Grid holds a value of type T for every cell of a rectangular, isometric or hexagonal layout,
// places its cells in the world, and answers the neighbor, range, area and path queries between
// them. It is built by [NewRect], [NewIsometric], [NewHexFlatTop] or [NewHexPointyTop], or read
// back by [Grid.UnmarshalJSON]; the zero value is no grid, and no other method of it is to be
// called.
type Grid[T any] struct {
	cells    container.Array[T, int]
	layout   layout
	topology topology
}

// NewRect returns a grid of square cells of the given size in cells, each cell of the given
// cell scale in the world. Each opts function sets the fields of the [SquareOptions] it is given.
// It panics for a movement outside the constants, for a cell scale that is not positive and
// finite on both axes, and for an origin with a NaN or infinite coordinate.
func NewRect[T any](size geom.Size[int], cellScale geom.Size[float64], opts ...func(*SquareOptions)) *Grid[T] {
	return newSquare[T](size, cellScale, kindRect, opts)
}

// NewIsometric returns a grid of square cells drawn as diamonds, of the given size in cells,
// each diamond of the given cell scale in the world, the first cell at the top corner. Each
// opts function sets the fields of the [SquareOptions] it is given. It panics for a movement
// outside the constants, for a cell scale that is not positive and finite on both axes, and for
// an origin with a NaN or infinite coordinate.
func NewIsometric[T any](size geom.Size[int], cellScale geom.Size[float64], opts ...func(*SquareOptions)) *Grid[T] {
	return newSquare[T](size, cellScale, kindIsometric, opts)
}

// newSquare returns a grid of square cells of the kind, with the options the callbacks set.
func newSquare[T any](size geom.Size[int], cellScale geom.Size[float64], kind kind, opts []func(*SquareOptions)) *Grid[T] {
	var options SquareOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.Movement.IsNone() {
		panic(options.Movement.unknown())
	}

	return newGrid[T](size, newLayout(kind, size, cellScale, options.Origin), newTopology(kind, options.Movement))
}

// NewHexFlatTop returns a grid of flat-top hexagons of the given size in cells, each of the
// given cell scale, its circumradius per axis: a regular hexagon where the two are equal, one
// stretched along an axis where they are not. The box around a cell is twice the radius wide
// and √3 times it tall. The odd columns are shifted half a cell down, or the even ones with
// EvenOffset. Each opts function sets the fields of the [HexOptions] it is given. It panics for
// a cell scale that is not positive and finite on both axes, and for an origin with a NaN or
// infinite coordinate.
func NewHexFlatTop[T any](size geom.Size[int], cellScale geom.Size[float64], opts ...func(*HexOptions)) *Grid[T] {
	return newHex[T](size, cellScale, kindHexFlatTop, kindHexFlatTopEven, opts)
}

// NewHexPointyTop returns a grid of pointy-top hexagons of the given size in cells, each of the
// given cell scale, its circumradius per axis: a regular hexagon where the two are equal, one
// stretched along an axis where they are not. The box around a cell is √3 times the radius wide
// and twice it tall. The odd rows are shifted half a cell right, or the even ones with
// EvenOffset. Each opts function sets the fields of the [HexOptions] it is given. It panics for
// a cell scale that is not positive and finite on both axes, and for an origin with a NaN or
// infinite coordinate.
func NewHexPointyTop[T any](size geom.Size[int], cellScale geom.Size[float64], opts ...func(*HexOptions)) *Grid[T] {
	return newHex[T](size, cellScale, kindHexPointyTop, kindHexPointyTopEven, opts)
}

// newHex returns a grid of hexagonal cells of the odd kind, or of the even one where the
// options the callbacks set ask for it.
func newHex[T any](size geom.Size[int], cellScale geom.Size[float64], odd, even kind, opts []func(*HexOptions)) *Grid[T] {
	var options HexOptions
	for _, opt := range opts {
		opt(&options)
	}

	kind := odd
	if options.EvenOffset {
		kind = even
	}

	return newGrid[T](size, newLayout(kind, size, cellScale, options.Origin), newTopology(kind, MovementNone))
}

// newGrid returns a grid of the given size in cells, placed by the layout, with the neighbor
// rule of its kind of cell.
func newGrid[T any](size geom.Size[int], layout layout, topology topology) *Grid[T] {
	return &Grid[T]{
		cells:    container.Arr[T](size),
		layout:   layout,
		topology: topology,
	}
}

// RectCellScale returns the cell scale of NewRect for a square of the given width.
func RectCellScale(width float64) geom.Size[float64] {
	return geom.SzU(width)
}

// IsometricCellScale returns the cell scale of NewIsometric for a diamond of the given width
// whose edges rise at 30°, the angle of a true isometric projection.
func IsometricCellScale(width float64) geom.Size[float64] {
	return geom.Sz(width, width*geom.Sqrt3/3)
}

// IsometricPixelCellScale returns the cell scale of NewIsometric for a diamond of the given
// width and half of it tall, whose edges step two pixels across for each one up.
func IsometricPixelCellScale(width float64) geom.Size[float64] {
	return geom.Sz(width, width*0.5)
}

// HexFlatTopCellScale returns the cell scale of NewHexFlatTop for a hexagon of the given width
// and √3/2 of it tall.
func HexFlatTopCellScale(width float64) geom.Size[float64] {
	return geom.SzU(width).Scale(0.5)
}

// HexPointyTopCellScale returns the cell scale of NewHexPointyTop for a hexagon of the given
// width and 2/√3 of it tall.
func HexPointyTopCellScale(width float64) geom.Size[float64] {
	return geom.SzU(width * 2 / geom.Sqrt3).Scale(0.5)
}

// HexFlatTopPixelCellScale returns the cell scale of NewHexFlatTop for a hexagon of the given
// width and three quarters of it tall, the hexagon of pixel-art isometry.
func HexFlatTopPixelCellScale(width float64) geom.Size[float64] {
	return geom.Sz(width, width*geom.Sqrt3/2).Scale(0.5)
}

// HexPointyTopPixelCellScale returns the cell scale of NewHexPointyTop for a hexagon as tall as
// the given width, the hexagon of pixel-art isometry.
func HexPointyTopPixelCellScale(width float64) geom.Size[float64] {
	return geom.Sz(width*2/geom.Sqrt3, width).Scale(0.5)
}

// Size returns the number of columns and rows of the grid.
func (g *Grid[T]) Size() geom.Size[int] {
	return g.cells.Size()
}

// Width returns the number of columns of the grid.
func (g *Grid[T]) Width() int {
	return g.cells.Width()
}

// Height returns the number of rows of the grid.
func (g *Grid[T]) Height() int {
	return g.cells.Height()
}

// Origin returns where the top-left corner of the bounds of the grid is in the world.
func (g *Grid[T]) Origin() geom.Point[float64] {
	return g.layout.origin
}

// Bounds returns the box around every cell of the grid in the world, its top-left corner the
// origin of the grid, and an empty box at the origin for a grid of no cells.
func (g *Grid[T]) Bounds() geom.Box[float64] {
	return g.layout.bounds
}

// CellScale returns the cell scale the constructor or Rescale took.
func (g *Grid[T]) CellScale() geom.Size[float64] {
	return g.layout.cellScale
}

// CellSize returns the size of a cell as drawn, the width and height of the box around it.
func (g *Grid[T]) CellSize() geom.Size[float64] {
	return g.layout.cellSize
}

// CellSpacing returns the distance between the centers of adjacent cells along each axis.
func (g *Grid[T]) CellSpacing() geom.Size[float64] {
	return g.layout.cellSpacing
}

// CellWidth returns the width of CellSize.
func (g *Grid[T]) CellWidth() float64 {
	return g.layout.cellSize.Width
}

// CellHeight returns the height of CellSize.
func (g *Grid[T]) CellHeight() float64 {
	return g.layout.cellSize.Height
}

// Cell returns the cell at the index, a handle on its value and its place in the world.
func (g *Grid[T]) Cell(index Index) Cell[T] {
	return Cell[T]{g, index}
}

// CellAt returns the cell at the index of IndexAt. It panics for a point with a NaN or infinite
// coordinate.
func (g *Grid[T]) CellAt(point geom.Point[float64]) Cell[T] {
	return g.Cell(g.IndexAt(point))
}

// IndexAt returns the index of the cell the point of the world lies in, an index outside the
// grid for a point outside it. A point on an edge goes to the same side in every cell. It
// panics for a point with a NaN or infinite coordinate.
func (g *Grid[T]) IndexAt(point geom.Point[float64]) Index {
	return g.layout.fromWorld(point)
}

// Fill sets every cell to value.
func (g *Grid[T]) Fill(value T) {
	g.cells.Fill(value)
}

// Clear resets every cell to the zero value.
func (g *Grid[T]) Clear() {
	g.cells.Clear()
}

// Clone returns a copy of the grid with storage of its own, in the same place, its values
// copied as an assignment copies them.
func (g *Grid[T]) Clone() *Grid[T] {
	clone := *g
	clone.cells = g.cells.Clone()

	return &clone
}

// Layer returns a grid of zero values of type U over the same cells: of the size, the layout and
// the neighbor rule of the grid, in the same place, an index of one an index of the other. It
// moves and rescales apart from the grid.
func (g *Grid[T]) Layer[U any]() *Grid[U] {
	return newGrid[U](g.Size(), g.layout, g.topology)
}

// All returns an iterator over every cell of the grid, row by row.
func (g *Grid[T]) All() iter.Seq[Cell[T]] {
	return func(yield func(Cell[T]) bool) {
		for index := range g.cells.Keys() {
			if !yield(g.Cell(index)) {
				return
			}
		}
	}
}

// Select returns the cells of the grid selected reports.
func (g *Grid[T]) Select(selected func(Cell[T]) bool) Region {
	selected = passableOrAny(selected)

	points := make([]Index, 0, g.cells.Len())
	for x := range g.Width() {
		for y := range g.Height() {
			if cell := g.Cell(geom.Pt(x, y)); selected(cell) {
				points = append(points, cell.index)
			}
		}
	}

	return container.SetFrom(points)
}

// DrawOrder returns an iterator over the cells of the grid in view of bounds, back to front, so
// a cell is never drawn over one in front of it: row by row on a rectangular or pointy-top grid,
// by the sum of column and row on an isometric one, and on a flat-top one row by row with the
// columns drawn higher of each row first. A zero bounds is the bounds of the grid. Cells just
// beyond the edge of bounds are included where one may be partly in view. It panics, once
// ranged, for bounds with a NaN or infinite coordinate.
func (g *Grid[T]) DrawOrder(bounds geom.Box[float64]) iter.Seq[Cell[T]] {
	if bounds.IsZero() {
		bounds = g.Bounds()
	}

	return func(yield func(Cell[T]) bool) {
		if !bounds.Min.IsFinite() || !bounds.Max.IsFinite() {
			panic(fmt.Errorf("grid: bounds %s is not finite", bounds))
		}

		switch g.layout.kind {
		case kindIsometric:
			g.drawIsometric(bounds, yield)
		case kindHexFlatTop, kindHexFlatTopEven, kindHexPointyTop, kindHexPointyTopEven:
			g.drawHex(bounds, yield)
		default:
			g.drawRect(bounds, yield)
		}
	}
}

// drawRect yields the cells of a rectangular grid row by row, from the cell at one corner of
// bounds to the cell at the other.
func (g *Grid[T]) drawRect(bounds geom.Box[float64], yield func(Cell[T]) bool) {
	first, last := g.drawSpan(bounds, 0)

	for row := first.Y; row <= last.Y; row++ {
		for column := first.X; column <= last.X; column++ {
			if !yield(g.Cell(geom.Pt(column, row))) {
				return
			}
		}
	}
}

// drawIsometric yields the cells of an isometric grid in view of bounds by depth, the sum of
// column and row, which every cell of one screen row shares. A cell of depth d and column c is
// centered at (2c-d) spacings across and d spacings down from the center of the first cell.
func (g *Grid[T]) drawIsometric(bounds geom.Box[float64], yield func(Cell[T]) bool) {
	origin := g.Cell(geom.Pt(0, 0)).Center()
	spacing := g.CellSpacing()
	halfHeight := float64(g.CellSize().Height * 0.5)

	first := max(int(math.Floor((bounds.Min.Y-origin.Y)/spacing.Height))-1, 0)
	last := min(int(math.Ceil((bounds.Max.Y-origin.Y)/spacing.Height))+1, g.Width()+g.Height()-2)

	for depth := first; depth <= last; depth++ {
		y := float64(float64(depth)*spacing.Height) + origin.Y
		if y+halfHeight <= bounds.Min.Y || y-halfHeight >= bounds.Max.Y {
			continue
		}

		// The column centered at x is (x-origin.X)/(2·spacing) plus half the depth.
		left := int(math.Floor((bounds.Min.X-origin.X)/(2*spacing.Width)+float64(float64(depth)*0.5))) - 1
		right := int(math.Ceil((bounds.Max.X-origin.X)/(2*spacing.Width)+float64(float64(depth)*0.5))) + 1

		for column := max(left, 0, depth-g.Height()+1); column <= min(right, depth, g.Width()-1); column++ {
			cell := g.Cell(geom.Pt(column, depth-column))
			if !cell.Bounds().IntersectsBox(bounds) {
				continue
			}
			if !yield(cell) {
				return
			}
		}
	}
}

// drawHex yields the cells of a hexagonal grid in view of bounds row by row, one cell beyond
// the corners of bounds, since a cell in view can start in the row or column beside the one a
// corner is in. Where the columns are staggered, two cells of one screen row differ in parity,
// so a row is walked in a pass per parity, the one its kind draws higher first.
func (g *Grid[T]) drawHex(bounds geom.Box[float64], yield func(Cell[T]) bool) {
	first, last := g.drawSpan(bounds, 1)
	step, parity := g.layout.kind.drawColumns()

	for row := first.Y; row <= last.Y; row++ {
		for pass := range step {
			for column := first.X + (first.X+parity+pass)%step; column <= last.X; column += step {
				cell := g.Cell(geom.Pt(column, row))
				if !cell.Bounds().IntersectsBox(bounds) {
					continue
				}
				if !yield(cell) {
					return
				}
			}
		}
	}
}

// drawSpan returns the indices of the cells at the corners of bounds, widened by margin and cut
// to the grid.
func (g *Grid[T]) drawSpan(bounds geom.Box[float64], margin int) (first, last Index) {
	first, last = g.IndexAt(bounds.Min), g.IndexAt(bounds.Max)

	return geom.Pt(max(first.X-margin, 0), max(first.Y-margin, 0)),
		geom.Pt(min(last.X+margin, g.Width()-1), min(last.Y+margin, g.Height()-1))
}

// MoveTo places the top-left corner of the bounds of the grid at origin, every cell keeping its
// value and its size. It panics for an origin with a NaN or infinite coordinate.
func (g *Grid[T]) MoveTo(origin geom.Point[float64]) {
	g.layout.moveTo(origin)
}

// Rescale replaces the cell scale the constructor of the grid took, the origin staying where it
// is and every cell keeping its value. It panics for a cell scale that is not positive and
// finite on both axes.
func (g *Grid[T]) Rescale(cellScale geom.Size[float64]) {
	g.layout.rescale(cellScale)
}

// Neighbors returns the indices adjacent to the index that lie in the grid, by increasing
// angle.
func (g *Grid[T]) Neighbors(index Index) []Index {
	if !g.Contains(index) {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, len(g.topology.offsets(index))), g.NeighborsSeq(index))
}

// NeighborsSeq returns an iterator over the indices of Neighbors, in the same order.
func (g *Grid[T]) NeighborsSeq(index Index) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		g.walkNeighbors(index, yield)
	}
}

// walkNeighbors yields the indices of Neighbors and reports whether it yielded them all. It checks
// the index itself, so NeighborsSeq only builds its closure and stays small enough to inline.
func (g *Grid[T]) walkNeighbors(index Index, yield func(Index) bool) bool {
	if !g.Contains(index) {
		return true
	}

	for _, offset := range g.topology.offsets(index) {
		if next := index.Add(offset); g.Contains(next) && !yield(next) {
			return false
		}
	}

	return true
}

// Range returns the cells within radius of the index. It is empty for a negative radius.
func (g *Grid[T]) Range(index Index, radius int) Region {
	if radius < 0 || !g.Contains(index) {
		return Region{}
	}

	return container.SetFrom(g.gatherRange(index, g.clampRadius(index, radius)))
}

// RangeSeq returns an iterator over the indices of Range, in the order of a region.
func (g *Grid[T]) RangeSeq(index Index, radius int) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		g.walkRange(index, radius, yield)
	}
}

// gatherRange returns the indices in range of an index the caller knows to be inside the grid,
// at a radius it knows to be cut to the grid.
func (g *Grid[T]) gatherRange(index Index, radius int) []Index {
	return slices.AppendSeq(make([]Index, 0, g.topology.rangeLen(radius)), g.RangeSeq(index, radius))
}

// walkRange yields the indices of Range and reports whether it yielded them all.
func (g *Grid[T]) walkRange(index Index, radius int, yield func(Index) bool) bool {
	if radius < 0 || !g.Contains(index) {
		return true
	}

	return g.topology.walkWithin(index, g.clampRadius(index, radius), func(found Index) bool {
		return !g.Contains(found) || yield(found)
	})
}

// clampRadius returns the radius cut to the distance from the index to the farthest cell of
// the grid, one of its corners, since a range gathers every index within its radius before it
// asks which lie in the grid. A radius short of the nearest edge needs no cut.
func (g *Grid[T]) clampRadius(index Index, radius int) int {
	x, y := index.XY()
	right, bottom := g.Width()-1, g.Height()-1
	if radius <= min(x, y, right-x, bottom-y) {
		return radius
	}

	return min(radius, max(
		g.topology.distance(index, geom.Pt(0, 0)),
		g.topology.distance(index, geom.Pt(right, 0)),
		g.topology.distance(index, geom.Pt(0, bottom)),
		g.topology.distance(index, geom.Pt(right, bottom)),
	))
}

// Ring returns the indices at exactly radius steps from the index that lie in the grid, by
// increasing angle around it, and the index alone for a zero radius. It returns nil for a
// negative radius and for one past every cell of the grid.
func (g *Grid[T]) Ring(index Index, radius int) []Index {
	if !g.hasRing(index, radius) {
		return nil
	}

	return slices.AppendSeq(make([]Index, 0, g.topology.ringLen(radius)), g.RingSeq(index, radius))
}

// RingSeq returns an iterator over the indices of Ring, in the same order.
func (g *Grid[T]) RingSeq(index Index, radius int) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		g.walkRing(index, radius, yield)
	}
}

// walkRing yields the indices of Ring and reports whether it yielded them all.
func (g *Grid[T]) walkRing(index Index, radius int, yield func(Index) bool) bool {
	if !g.hasRing(index, radius) {
		return true
	}

	return g.topology.walkRing(index, radius, func(found Index) bool {
		return !g.Contains(found) || yield(found)
	})
}

// hasRing reports whether the index is inside the grid and a cell of it can be at radius steps
// from the index: none is past the farthest cell.
func (g *Grid[T]) hasRing(index Index, radius int) bool {
	return radius >= 0 && g.Contains(index) && g.clampRadius(index, radius) == radius
}

// Spiral returns the indices within radius of the index that lie in the grid, ring by ring
// outward in the order of Ring. It returns nil for a negative radius.
func (g *Grid[T]) Spiral(index Index, radius int) []Index {
	if radius < 0 || !g.Contains(index) {
		return nil
	}

	size := g.topology.rangeLen(g.clampRadius(index, radius))

	return slices.AppendSeq(make([]Index, 0, size), g.SpiralSeq(index, radius))
}

// SpiralSeq returns an iterator over the indices of Spiral, in the same order.
func (g *Grid[T]) SpiralSeq(index Index, radius int) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		g.walkSpiral(index, radius, yield)
	}
}

// walkSpiral yields the indices of Spiral and reports whether it yielded them all.
func (g *Grid[T]) walkSpiral(index Index, radius int, yield func(Index) bool) bool {
	if radius < 0 || !g.Contains(index) {
		return true
	}

	return g.topology.walkSpiral(index, g.clampRadius(index, radius), func(found Index) bool {
		return !g.Contains(found) || yield(found)
	})
}

// Border returns the indices of the region with a neighbor outside it: its outline, around a
// hole as well. Beyond the edge of the grid is outside every region, so a region reaching the
// edge has its border there.
func (g *Grid[T]) Border(region Region) Region {
	size := 0
	for index := range region.All() {
		if g.onBorder(region, index) {
			size++
		}
	}
	if size == 0 {
		return Region{}
	}

	points := make([]Index, 0, size)
	for index := range region.All() {
		if g.onBorder(region, index) {
			points = append(points, index)
		}
	}

	return container.SetFrom(points)
}

// Grow returns the region widened by n steps: the region with every cell of the grid within n
// steps of one of its indices. It returns the region itself for an n of zero or less.
func (g *Grid[T]) Grow(region Region, n int) Region {
	if n <= 0 || region.IsEmpty() {
		return region
	}

	// held marks the cells of the grid already in the result, so each is added once.
	held := make([]bool, g.cells.Len())
	for index := range region.All() {
		if g.Contains(index) {
			held[g.slot(index)] = true
		}
	}

	points := region.Elements()

	// Each step grows from the indices the step before it added, the region itself at first.
	for ring, step := points, 0; step < n && len(ring) > 0; step++ {
		added := len(points)

		for _, index := range ring {
			for _, offset := range g.topology.offsets(index) {
				next := index.Add(offset)
				if !g.Contains(next) || held[g.slot(next)] {
					continue
				}

				held[g.slot(next)] = true
				points = append(points, next)
			}
		}

		ring = points[added:]
	}

	return container.SetFrom(points)
}

// Shrink returns the region worn away by n steps, its border removed n times over. It returns
// the region itself for an n of zero or less.
func (g *Grid[T]) Shrink(region Region, n int) Region {
	for step := 0; step < n && !region.IsEmpty(); step++ {
		region = region.Difference(g.Border(region))
	}

	return region
}

// IsConnected reports whether every index of the region is reached from every other by steps
// between neighbors inside it. The empty region is connected.
func (g *Grid[T]) IsConnected(region Region) bool {
	if region.Len() <= 1 {
		return true
	}

	return len(g.component(region, 0, make([]bool, region.Len()), make([]int, 0, region.Len()))) == region.Len()
}

// Components returns the connected pieces of the region, none adjacent to another, ordered by
// their first index, and nil for the empty region.
func (g *Grid[T]) Components(region Region) []Region {
	var components []Region

	visited := make([]bool, region.Len())
	members := make([]int, 0, region.Len())
	for i := range region.Len() {
		if visited[i] {
			continue
		}

		members = g.component(region, i, visited, members[:0])
		slices.Sort(members)

		points := make([]Index, len(members))
		for j, member := range members {
			points[j] = region.At(member)
		}

		components = append(components, container.SetFrom(points))
	}

	return components
}

// onBorder reports whether an index of the region has a neighbor outside it.
func (g *Grid[T]) onBorder(region Region, index Index) bool {
	for _, offset := range g.topology.offsets(index) {
		if next := index.Add(offset); !g.Contains(next) || !region.Contains(next) {
			return true
		}
	}

	return false
}

// component appends to members the places, among the sorted indices of the region, of the
// indices connected to the one at the given place, marking each in visited.
func (g *Grid[T]) component(region Region, from int, visited []bool, members []int) []int {
	members = append(members, from)
	visited[from] = true

	for next := 0; next < len(members); next++ {
		index := region.At(members[next])

		for _, offset := range g.topology.offsets(index) {
			if neighbor, found := region.Index(index.Add(offset)); found && !visited[neighbor] {
				visited[neighbor] = true
				members = append(members, neighbor)
			}
		}
	}

	return members
}

// slot returns the place of an index inside the grid among its cells, row by row.
func (g *Grid[T]) slot(index Index) int {
	return index.X + index.Y*g.Width()
}

// Distance returns the number of steps between two indices, whether or not they lie in the
// grid.
func (g *Grid[T]) Distance(from, to Index) int {
	return g.topology.distance(from, to)
}

// Line returns the indices of the straight line from one index to the other, both included,
// each a neighbor of the one before. Unlike Ring and Spiral it is measured between any two
// indices, whether or not they lie in the grid, and between two cells on a zigzag edge of a
// hexagonal grid it may pass through an index outside it.
func (g *Grid[T]) Line(from, to Index) []Index {
	return slices.AppendSeq(make([]Index, 0, g.topology.distance(from, to)+1), g.LineSeq(from, to))
}

// LineSeq returns an iterator over the indices of Line, in the same order.
func (g *Grid[T]) LineSeq(from, to Index) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		g.topology.walkLine(from, to, yield)
	}
}

// HasLineOfSight reports whether from sees to past the cells blocked reports, asking it only
// about the indices the line passes that lie in the grid.
func (g *Grid[T]) HasLineOfSight(from, to Index, blocked func(Cell[T]) bool) bool {
	return g.Contains(from) && g.Contains(to) && g.topology.sees(from, to, g.blockedIndex(blocked))
}

// FieldOfView returns the cells of Range the index sees past the cells blocked reports, asking it
// about each cell at most once. A blocked cell in sight is in the result, and the cell of the
// index is never asked about.
func (g *Grid[T]) FieldOfView(index Index, radius int, blocked func(Cell[T]) bool) Region {
	if radius < 0 || !g.Contains(index) {
		return Region{}
	}

	radius = g.clampRadius(index, radius)
	points := g.gatherRange(index, radius)

	// Only where a cell was refused can anything be hidden, so only then is sight looked for.
	marks := askSight(index, radius, points, g.blockedIndex(blocked))
	if marks.isEmpty() {
		return container.SetFrom(points)
	}

	return container.SetFrom(g.keepInSight(points, index, radius, marks))
}

// FieldOfViewSeq returns an iterator over the indices of FieldOfView, in no order to rely on.
func (g *Grid[T]) FieldOfViewSeq(index Index, radius int, blocked func(Cell[T]) bool) iter.Seq[Index] {
	return func(yield func(Index) bool) {
		g.walkFieldOfView(index, radius, blocked, yield)
	}
}

// keepInSight returns the indices center sees. It takes the marks by value, so only a range
// refused a cell puts them on the heap.
func (g *Grid[T]) keepInSight(indices []Index, center Index, radius int, marks sightMarks) []Index {
	g.topology.sight(center, radius, indices, &marks)

	return marks.keepSeen(indices)
}

// walkFieldOfView yields the indices of FieldOfView as it finds them and reports whether it
// yielded them all.
func (g *Grid[T]) walkFieldOfView(index Index, radius int, blocked func(Cell[T]) bool, yield func(Index) bool) bool {
	if radius < 0 || !g.Contains(index) {
		return true
	}

	return g.topology.walkFieldOfView(index, g.clampRadius(index, radius), g.blockedIndex(blocked), func(found Index) bool {
		return !g.Contains(found) || yield(found)
	})
}

// blockedIndex returns blocked over indices, nil where it is nil. What lies outside the grid
// blocks nothing, since nothing stands there.
func (g *Grid[T]) blockedIndex(blocked func(Cell[T]) bool) func(Index) bool {
	if blocked == nil {
		return nil
	}

	return func(index Index) bool {
		return g.Contains(index) && blocked(g.Cell(index))
	}
}

// Path returns the path of PathAStar from start to goal, both included, or nil where none is found.
func (g *Grid[T]) Path(start, goal Index, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []Index {
	return g.astar(start, goal, passable, cost).PathTo(goal)
}

// PathBreadthFirst returns the path of the fewest steps from start to goal, both included,
// or nil where none is found.
func (g *Grid[T]) PathBreadthFirst(start, goal Index, passable func(Cell[T]) bool) []Index {
	return g.breadthFirstSearch(start, &goal, math.Inf(1), passable).PathTo(goal)
}

// PathUniformCost returns the cheapest path from start to goal, both included, or nil where
// none is found. Uniform cost search is also known as Dijkstra's algorithm.
func (g *Grid[T]) PathUniformCost(start, goal Index, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []Index {
	return g.uniformCostSearch(start, &goal, math.Inf(1), passable, cost).PathTo(goal)
}

// PathGreedyBestFirst returns a path from start to goal, both included, or nil where none is
// found, heading for the cell nearest to the goal at every step. The path is not always the
// shortest.
func (g *Grid[T]) PathGreedyBestFirst(start, goal Index, passable func(Cell[T]) bool) []Index {
	return g.greedyBestFirstSearch(start, goal, passable).PathTo(goal)
}

// PathAStar returns the cheapest path from start to goal, both included, or nil where none is
// found. A cost below one per step can make the path dearer than the cheapest.
func (g *Grid[T]) PathAStar(start, goal Index, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []Index {
	return g.astar(start, goal, passable, cost).PathTo(goal)
}

// Field returns the field of every cell reached from start, with the cheapest way to each.
func (g *Grid[T]) Field(start Index, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	if cost == nil {
		return g.breadthFirstSearch(start, nil, math.Inf(1), passable)
	}

	return g.uniformCostSearch(start, nil, math.Inf(1), passable, cost)
}

// FieldBreadthFirst returns the field of every cell reached from start, with the fewest
// steps to each, whatever a step costs.
func (g *Grid[T]) FieldBreadthFirst(start Index, passable func(Cell[T]) bool) Field {
	return g.breadthFirstSearch(start, nil, math.Inf(1), passable)
}

// FieldUniformCost returns the field of every cell reached from start, with the cheapest
// way to each. Uniform cost search is also known as Dijkstra's algorithm.
func (g *Grid[T]) FieldUniformCost(start Index, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	return g.uniformCostSearch(start, nil, math.Inf(1), passable, cost)
}

// Reachable returns the field of the cells reached from start for at most budget, the search
// walking no further, with the cheapest way to each. A nil cost makes the budget a number of
// steps. It is empty for a negative budget or a NaN.
func (g *Grid[T]) Reachable(start Index, budget float64, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	if budget < 0 || math.IsNaN(budget) {
		return Field{}
	}

	if cost == nil {
		return g.breadthFirstSearch(start, nil, budget, passable)
	}

	return g.uniformCostSearch(start, nil, budget, passable, cost)
}

// breadthFirstSearch walks outward from start a step at a time, to the goal where one is given
// and no more steps than the budget.
func (g *Grid[T]) breadthFirstSearch(start Index, goal *Index, budget float64, passable func(Cell[T]) bool) Field {
	passable = passableOrAny(passable)

	if !g.checkPath(start, goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, false)

	// Every cell enters the queue once at most, so it never grows past the grid.
	queue := make([]int32, 1, len(field.from))
	queue[0] = field.slot(start)

	for head, steps := 0, 1; head < len(queue) && float64(steps) <= budget; steps++ {
		for ring := len(queue); head < ring; head++ {
			current := field.index(queue[head])
			if goal != nil && current == *goal {
				return field
			}

			for _, offset := range g.topology.offsets(current) {
				next := current.Add(offset)
				if !g.Contains(next) || field.from[field.slot(next)] != 0 || !passable(g.Cell(next)) {
					continue
				}

				field.reach(next, queue[head])
				queue = append(queue, field.slot(next))
			}
		}
	}

	return field
}

// uniformCostSearch expands the cheapest cell reached so far, to the goal where one is given and
// no further than the budget.
func (g *Grid[T]) uniformCostSearch(start Index, goal *Index, budget float64, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	passable = passableOrAny(passable)
	cost = costOrUnit(cost)

	if !g.checkPath(start, goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, true)

	var open frontier
	open.push(field.slot(start), 0)

	for open.len() > 0 {
		slot := open.pop()
		current := field.index(slot)
		if goal != nil && current == *goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.Contains(next) || !passable(g.Cell(next)) {
				continue
			}

			reached := field.costs[slot] + checkedCost(cost, g.Cell(current), g.Cell(next))
			if reached <= budget && field.improve(next, slot, reached) {
				open.push(field.slot(next), reached)
			}
		}
	}

	return field
}

// greedyBestFirstSearch expands the cell reached so far that is nearest to the goal.
func (g *Grid[T]) greedyBestFirstSearch(start, goal Index, passable func(Cell[T]) bool) Field {
	passable = passableOrAny(passable)

	if !g.checkPath(start, &goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, false)

	var open frontier
	open.push(field.slot(start), 0)

	for open.len() > 0 {
		slot := open.pop()
		current := field.index(slot)
		if current == goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.Contains(next) || field.from[field.slot(next)] != 0 || !passable(g.Cell(next)) {
				continue
			}

			field.reach(next, slot)
			open.push(field.slot(next), g.heuristic(next, goal))
		}
	}

	return field
}

// astar expands the cell whose cost so far and distance to the goal sum to the least.
func (g *Grid[T]) astar(start, goal Index, passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	passable = passableOrAny(passable)
	cost = costOrUnit(cost)

	if !g.checkPath(start, &goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, true)

	var open frontier
	open.push(field.slot(start), 0)

	for open.len() > 0 {
		slot := open.pop()
		current := field.index(slot)
		if current == goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.Contains(next) || !passable(g.Cell(next)) {
				continue
			}

			reached := field.costs[slot] + checkedCost(cost, g.Cell(current), g.Cell(next))
			if field.improve(next, slot, reached) {
				open.push(field.slot(next), reached+g.heuristic(next, goal))
			}
		}
	}

	return field
}

// heuristic returns the least a path from current to goal can cost at one per step.
func (g *Grid[T]) heuristic(current, goal Index) float64 {
	return float64(g.Distance(current, goal))
}

// checkPath reports whether a search can start: the start, and the goal where one is given,
// are inside the grid and passable.
func (g *Grid[T]) checkPath(start Index, goal *Index, passable func(Cell[T]) bool) bool {
	if !g.Contains(start) || (goal != nil && !g.Contains(*goal)) {
		return false
	}

	return passable(g.Cell(start)) && (goal == nil || passable(g.Cell(*goal)))
}

// Contains reports whether the index is that of a cell of the grid.
func (g *Grid[T]) Contains(index Index) bool {
	return g.cells.Contains(index)
}

// MarshalJSON implements json.Marshaler with the kind of the grid, the movement of a square one,
// its cell scale and origin, and its size and values.
func (g Grid[T]) MarshalJSON() ([]byte, error) {
	var movement *Movement
	if !g.topology.isHex() {
		movement = &g.topology.movement
	}

	return json.Marshal(struct {
		Kind      kind                    `json:"k"`
		Movement  *Movement               `json:"m,omitempty"`
		CellScale geom.Size[float64]      `json:"c"`
		Origin    geom.Point[float64]     `json:"o"`
		Cells     container.Array[T, int] `json:"v"`
	}{g.layout.kind, movement, g.layout.cellScale, g.layout.origin, g.cells})
}

// UnmarshalJSON implements json.Unmarshaler, replacing the grid, a zero one too, with the one
// MarshalJSON wrote, and reading null as nothing. It returns an error for what no constructor
// builds, and leaves the grid as it was on an error.
func (g *Grid[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var decoded struct {
		Kind      kind                    `json:"k"`
		Movement  Movement                `json:"m"`
		CellScale geom.Size[float64]      `json:"c"`
		Origin    geom.Point[float64]     `json:"o"`
		Cells     container.Array[T, int] `json:"v"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	if decoded.Kind.sides() == 4 && decoded.Movement.IsNone() {
		return fmt.Errorf("grid: square cells with no movement")
	}

	if err := checkCellScale(decoded.CellScale); err != nil {
		return err
	}

	*g = Grid[T]{
		cells:    decoded.Cells,
		layout:   newLayout(decoded.Kind, decoded.Cells.Size(), decoded.CellScale, decoded.Origin),
		topology: newTopology(decoded.Kind, decoded.Movement),
	}

	return nil
}
