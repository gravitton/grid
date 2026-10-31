package grid

import (
	geom "github.com/gravitton/geometry"
)

// sightMarks is what a range knows of the square of indices around its center: which cells it
// was refused and which it found in sight past them. The marks are made at the first refusal,
// so a range refused nothing makes none, and the cells beyond the edge of the grid block as the
// refused ones do.
type sightMarks struct {
	size   geom.Size[int]
	corner geom.Point[int]
	side   int
	marks  []uint8
}

const (
	markRefused uint8 = 1 << iota
	markSeen
)

// refuse marks the cell at the index as refused.
func (m *sightMarks) refuse(index geom.Point[int]) {
	if m.marks == nil {
		m.marks = make([]uint8, m.side*m.side)
	}

	m.marks[m.slot(index)] |= markRefused
}

// see marks the cell at the index as in sight.
func (m *sightMarks) see(index geom.Point[int]) {
	m.marks[m.slot(index)] |= markSeen
}

// blocks reports whether the index is beyond the edge of the grid or that of a cell refused.
func (m *sightMarks) blocks(index geom.Point[int]) bool {
	if index.X < 0 || index.X >= m.size.Width || index.Y < 0 || index.Y >= m.size.Height {
		return true
	}

	x, y := index.X-m.corner.X, index.Y-m.corner.Y

	return x >= 0 && x < m.side && y >= 0 && y < m.side && m.marks[x+y*m.side]&markRefused != 0
}

// seen reports whether the cell at the index, inside the square, was marked as in sight.
func (m *sightMarks) seen(index geom.Point[int]) bool {
	return m.marks[m.slot(index)]&markSeen != 0
}

// isEmpty reports whether no cell was refused.
func (m *sightMarks) isEmpty() bool {
	return m.marks == nil
}

// slot returns the place of the mark of an index inside the square.
func (m *sightMarks) slot(index geom.Point[int]) int {
	return index.X - m.corner.X + (index.Y-m.corner.Y)*m.side
}

// blockers answers whether an index blocks sight, from a set of indices or from a function, so
// the slice and the Func forms of the sight methods share one rule and neither pays for the
// other: the set is read without a call.
type blockers struct {
	set     pointSet
	blocked func(geom.Point[int]) bool
}

// has reports whether the index blocks.
func (b blockers) has(index geom.Point[int]) bool {
	if b.blocked != nil {
		return b.blocked(index)
	}

	return b.set.has(index)
}

// isEmpty reports whether nothing can block.
func (b blockers) isEmpty() bool {
	return b.blocked == nil && len(b.set) == 0
}

// pointSet is a set of indices read without a call.
type pointSet map[geom.Point[int]]struct{}

// newPointSet returns the set of the points, and nil, which holds nothing, for none.
func newPointSet(points []geom.Point[int]) pointSet {
	if len(points) == 0 {
		return nil
	}

	set := make(pointSet, len(points))
	for _, point := range points {
		set[point] = struct{}{}
	}

	return set
}

// has reports whether the set contains the point.
func (s pointSet) has(point geom.Point[int]) bool {
	_, found := s[point]

	return found
}
