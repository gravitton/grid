package grid

import (
	geom "github.com/gravitton/geometry"
)

// sightMarks is what a range knows of the square of indices around its center: which cells were
// refused and which it found in sight past them. The marks are made at the first refusal, so a
// range refused nothing makes none. A line stays inside the box of its ends, so every cell it
// passes is one the range asked about, or outside the grid, where nothing stands.
type sightMarks struct {
	corner Index
	side   int
	marks  []uint8
}

const (
	markRefused uint8 = 1 << iota
	markSeen
)

// askSight returns the marks of a range around center over the candidates, asking blocked about
// each one but center, and none for a nil blocked.
func askSight(center Index, radius int, candidates []Index, blocked func(Index) bool) sightMarks {
	marks := sightMarks{corner: geom.Pt(center.X-radius-1, center.Y-radius-1), side: 2*radius + 3}
	if blocked == nil {
		return marks
	}

	for _, candidate := range candidates {
		if candidate != center && blocked(candidate) {
			marks.refuse(candidate)
		}
	}

	return marks
}

// refuse marks the cell at the index as refused.
func (m *sightMarks) refuse(index Index) {
	if m.marks == nil {
		m.marks = make([]uint8, m.side*m.side)
	}

	m.marks[m.slot(index)] |= markRefused
}

// see marks the cell at the index as in sight.
func (m *sightMarks) see(index Index) {
	m.marks[m.slot(index)] |= markSeen
}

// blocks reports whether the index is that of a cell refused.
func (m *sightMarks) blocks(index Index) bool {
	x, y := index.X-m.corner.X, index.Y-m.corner.Y

	return x >= 0 && x < m.side && y >= 0 && y < m.side && m.marks[x+y*m.side]&markRefused != 0
}

// keepSeen returns dst with the indices from start on that were not marked as in sight taken
// out, in place and in their order.
func (m *sightMarks) keepSeen(dst []Index, start int) []Index {
	kept := dst[:start]
	for _, candidate := range dst[start:] {
		if m.marks[m.slot(candidate)]&markSeen != 0 {
			kept = append(kept, candidate)
		}
	}

	return kept
}

// isEmpty reports whether no cell was refused.
func (m *sightMarks) isEmpty() bool {
	return m.marks == nil
}

// slot returns the place of the mark of an index inside the square.
func (m *sightMarks) slot(index Index) int {
	return index.X - m.corner.X + (index.Y-m.corner.Y)*m.side
}

// sightMemo remembers what blocked said of the square of indices around a center, so a walk that
// cannot ask about every index first asks about each once, as the lines reach it.
type sightMemo struct {
	blocked func(Index) bool
	corner  Index
	side    int
	marks   []uint8
}

const (
	memoAsked uint8 = 1 << iota
	memoBlocked
)

// newSightMemo returns the memo of a range around center.
func newSightMemo(center Index, radius int, blocked func(Index) bool) *sightMemo {
	side := 2*radius + 3

	return &sightMemo{blocked, geom.Pt(center.X-radius-1, center.Y-radius-1), side, make([]uint8, side*side)}
}

// has reports whether blocked said the index blocks, asking it the first time.
func (m *sightMemo) has(index Index) bool {
	mark := &m.marks[index.X-m.corner.X+(index.Y-m.corner.Y)*m.side]
	if *mark&memoAsked == 0 {
		*mark = memoAsked
		if m.blocked(index) {
			*mark |= memoBlocked
		}
	}

	return *mark&memoBlocked != 0
}
