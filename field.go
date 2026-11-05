package grid

import (
	"iter"
	"math"
	"slices"

	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

// Field is what a search from one index found: every index it reached, with the step back
// toward the start and the cost of getting there. A field searched from a destination
// serves every unit heading there, its path running from the destination to the unit, where a
// step costs the same both ways. It does not change with the grid. The zero value is the empty
// field.
type Field struct {
	size  geom.Size[int]
	start Index

	// columns is the number of chunks across the grid.
	columns int

	// chunks holds, for every chunk of the grid, one more than its place among the chunks the
	// search touched, so zero marks a chunk it never touched.
	chunks []int32

	// corners holds, for every chunk the search touched, the index at its top-left corner.
	corners []Index

	// links holds the step back and the steps of every slot of the chunks touched, a chunk
	// after another in the order they were touched.
	links []link

	// costs holds the cost of reaching every slot of links, and is nil for a search that counts
	// steps.
	costs []float64

	reached int
}

// newField returns a field over a grid of the given size, holding the start alone, with room
// for costs where the search weighs its steps, and at first for the chunks within the radius
// of the start, every chunk for a negative one.
func newField(size geom.Size[int], start Index, weighted bool, radius int) Field {
	columns := (size.Width + chunkMask) >> chunkShift
	rows := (size.Height + chunkMask) >> chunkShift

	room := columns * rows
	if radius >= 0 {
		across := (min(start.X+radius, size.Width-1) >> chunkShift) - (max(start.X-radius, 0) >> chunkShift) + 1
		down := (min(start.Y+radius, size.Height-1) >> chunkShift) - (max(start.Y-radius, 0) >> chunkShift) + 1
		room = across * down
	}

	f := Field{
		size:    size,
		start:   start,
		columns: columns,
		chunks:  make([]int32, columns*rows),
		corners: make([]Index, 0, room),
		links:   make([]link, 0, room*chunkCells),
		reached: 1,
	}
	if weighted {
		f.costs = make([]float64, 0, room*chunkCells)
	}

	slot := f.claim(start)
	f.links[slot].from = slot + 1

	return f
}

// Start returns the index the search started from, and the zero index for the empty field.
func (f Field) Start() Index {
	return f.start
}

// Len returns the number of indices the search reached, the start included.
func (f Field) Len() int {
	return f.reached
}

// Region returns the indices the search reached.
func (f Field) Region() Region {
	if f.reached == 0 {
		return Region{}
	}

	return container.SetFrom(slices.AppendSeq(make([]Index, 0, f.reached), f.All()))
}

// All returns an iterator over the indices the search reached, in the order of a region.
func (f Field) All() iter.Seq[Index] {
	return func(yield func(Index) bool) {
		if f.reached == 0 {
			return
		}

		rows := len(f.chunks) / f.columns
		for x := range f.size.Width {
			for row := range rows {
				place := f.chunks[row*f.columns+x>>chunkShift]
				if place == 0 {
					continue
				}

				for y := row << chunkShift; y < min((row+1)<<chunkShift, f.size.Height); y++ {
					if f.links[f.slotIn(place, geom.Pt(x, y))].from != 0 && !yield(geom.Pt(x, y)) {
						return
					}
				}
			}
		}
	}
}

// reaches reports whether the search reached the slot.
func (f *Field) reaches(slot int32) bool {
	return f.links[slot].from != 0
}

// reach records that the slot, which the search had not reached, is reached from another.
func (f *Field) reach(slot, from int32) {
	f.links[slot] = link{from: from + 1, steps: f.links[from].steps + 1}
	f.reached++
}

// improve records that index is reached from the slot at the given cost, where the cost is
// finite and the search had not reached it or had paid more, and returns its slot and whether
// it did.
func (f *Field) improve(index Index, from int32, cost float64) (int32, bool) {
	if math.IsInf(cost, 1) {
		return 0, false
	}

	slot := f.claim(index)

	switch {
	case f.links[slot].from == 0:
		f.reached++
	case cost >= f.costs[slot]:
		return 0, false
	}

	f.links[slot] = link{from: from + 1, steps: f.links[from].steps + 1}
	f.costs[slot] = cost

	return slot, true
}

// claim returns the slot of an index inside the grid, giving its chunk room in the slices of
// the field where the search had not touched it.
func (f *Field) claim(index Index) int32 {
	chunk := (index.Y>>chunkShift)*f.columns + index.X>>chunkShift
	if f.chunks[chunk] == 0 {
		f.corners = append(f.corners, geom.Pt(index.X&^chunkMask, index.Y&^chunkMask))
		f.chunks[chunk] = int32(len(f.corners))
		f.links = slices.Grow(f.links, chunkCells)[:len(f.links)+chunkCells]
		if f.costs != nil {
			f.costs = slices.Grow(f.costs, chunkCells)[:len(f.costs)+chunkCells]
		}
	}

	return f.slotIn(f.chunks[chunk], index)
}

// slotBeside returns the slot of an index beside a node, read off the slot of the node, and
// false where the two lie in different chunks.
func (f Field) slotBeside(index Index, beside node) (int32, bool) {
	if ((index.X^beside.index.X)|(index.Y^beside.index.Y))&^chunkMask != 0 {
		return 0, false
	}

	return beside.slot&^(chunkCells-1) | f.cellIn(index), true
}

// Next returns the index one step closer to the start, and false for the start itself and for
// an index the search did not reach.
func (f Field) Next(index Index) (Index, bool) {
	slot, found := f.lookup(index)
	if !found || index == f.start {
		return Index{}, false
	}

	return f.index(f.links[slot].from - 1), true
}

// CostTo returns what the search paid to reach the index, the number of steps where it did not
// weigh them, and false for an index it did not reach.
func (f Field) CostTo(index Index) (float64, bool) {
	slot, found := f.lookup(index)
	if !found {
		return 0, false
	}
	if f.costs != nil {
		return f.costs[slot], true
	}

	return float64(f.links[slot].steps), true
}

// PathTo returns the path the search found from the start to the index, both included, and nil
// for an index it did not reach.
func (f Field) PathTo(index Index) []Index {
	slot, found := f.lookup(index)
	if !found {
		return nil
	}

	return f.appendPath(make([]Index, 0, f.links[slot].steps+1), slot)
}

// AppendPathTo appends the indices of PathTo to dst and returns the extended slice.
func (f Field) AppendPathTo(dst []Index, index Index) []Index {
	slot, found := f.lookup(index)
	if !found {
		return dst
	}

	return f.appendPath(dst, slot)
}

// appendPath appends the path from the start to the slot of an index the search reached.
func (f Field) appendPath(dst []Index, slot int32) []Index {
	start, length := len(dst), int(f.links[slot].steps)+1
	dst = slices.Grow(dst, length)[:start+length]
	for i := len(dst) - 1; i >= start; i-- {
		dst[i] = f.index(slot)
		slot = f.links[slot].from - 1
	}

	return dst
}

// Contains reports whether the search reached the index.
func (f Field) Contains(index Index) bool {
	_, found := f.lookup(index)

	return found
}

// IsEmpty reports whether the search reached nothing, not even its start.
func (f Field) IsEmpty() bool {
	return f.reached == 0
}

// lookup returns the slot of an index the search reached, and false for any other index.
func (f Field) lookup(index Index) (int32, bool) {
	if index.X < 0 || index.X >= f.size.Width || index.Y < 0 || index.Y >= f.size.Height {
		return 0, false
	}

	place := f.chunks[(index.Y>>chunkShift)*f.columns+index.X>>chunkShift]
	if place == 0 {
		return 0, false
	}

	slot := f.slotIn(place, index)

	return slot, f.links[slot].from != 0
}

// startSlot returns the slot of the start, in the chunk the search touched first.
func (f Field) startSlot() int32 {
	return f.slotIn(1, f.start)
}

// slotIn returns the slot of an index in the chunk at the given place, one more than its place
// among the chunks touched.
func (f Field) slotIn(place int32, index Index) int32 {
	return (place-1)<<(2*chunkShift) | f.cellIn(index)
}

// cellIn returns the place of an index within its chunk, row by row.
func (f Field) cellIn(index Index) int32 {
	return int32((index.Y&chunkMask)<<chunkShift | index.X&chunkMask)
}

// index returns the index at a slot, the inverse of slot.
func (f Field) index(slot int32) Index {
	corner := f.corners[slot>>(2*chunkShift)]
	local := int(slot & (chunkCells - 1))

	return geom.Pt(corner.X|local&chunkMask, corner.Y|local>>chunkShift)
}
