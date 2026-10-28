package grid

import (
	"slices"

	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

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

	points := make([]geom.Point[int], 0, size)
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
			held[g.position(index)] = true
		}
	}

	points := region.Elements()

	// Each step grows from the indices the step before it added, the region itself at first.
	for ring, step := points, 0; step < n && len(ring) > 0; step++ {
		added := len(points)

		for _, index := range ring {
			for _, offset := range g.topology.offsets(index) {
				next := index.Add(offset)
				if !g.Contains(next) || held[g.position(next)] {
					continue
				}

				held[g.position(next)] = true
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

	return len(g.component(region, 0, make([]bool, region.Len()))) == region.Len()
}

// Components returns the connected pieces of the region, none adjacent to another, ordered by
// their first index, and nil for the empty region.
func (g *Grid[T]) Components(region Region) []Region {
	var components []Region

	visited := make([]bool, region.Len())
	for i := range region.Len() {
		if visited[i] {
			continue
		}

		members := g.component(region, i, visited)
		slices.Sort(members)

		points := make([]geom.Point[int], len(members))
		for j, member := range members {
			points[j] = region.At(member)
		}

		components = append(components, container.SetFrom(points))
	}

	return components
}

// onBorder reports whether an index of the region has a neighbor outside it.
func (g *Grid[T]) onBorder(region Region, index geom.Point[int]) bool {
	for _, offset := range g.topology.offsets(index) {
		if next := index.Add(offset); !g.Contains(next) || !region.Contains(next) {
			return true
		}
	}

	return false
}

// component returns the places, among the sorted indices of the region, of the indices
// connected to the one at the given place, marking each in visited.
func (g *Grid[T]) component(region Region, from int, visited []bool) []int {
	members := []int{from}
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

// position returns the place of an index inside the grid among its cells, row by row.
func (g *Grid[T]) position(index geom.Point[int]) int {
	return index.X + index.Y*g.Width()
}
