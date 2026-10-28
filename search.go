package grid

import (
	geom "github.com/gravitton/geometry"
)

// BreadthFirstSearchField returns the field of every cell reached from start, with the fewest
// steps to each, whatever a step costs.
func (g *Grid[T]) BreadthFirstSearchField(start geom.Point[int], passable func(Cell[T]) bool) Field {
	return g.breadthFirstSearch(start, nil, -1, passable)
}

// BreadthFirstSearch returns the path of the fewest steps from start to goal, both included,
// or nil where none is found.
func (g *Grid[T]) BreadthFirstSearch(start, goal geom.Point[int], passable func(Cell[T]) bool) []geom.Point[int] {
	return g.breadthFirstSearch(start, &goal, -1, passable).PathTo(goal)
}

// Reachable returns the cells reached from start in at most the given number of steps through
// passable cells, start included. It is empty for a negative number of steps and for a start
// that is not passable.
func (g *Grid[T]) Reachable(start geom.Point[int], steps int, passable func(Cell[T]) bool) Region {
	if steps < 0 {
		return Region{}
	}

	return g.breadthFirstSearch(start, nil, steps, passable).Region()
}

// breadthFirstSearch walks outward from start a step at a time, to the goal where one is given
// and no further than limit steps where it is not negative.
func (g *Grid[T]) breadthFirstSearch(start geom.Point[int], goal *geom.Point[int], limit int, passable func(Cell[T]) bool) Field {
	passable = passableOrAny(passable)

	if !g.checkPath(start, goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, false)

	// Every cell enters the queue once at most, so it never grows past the grid.
	queue := make([]int32, 1, len(field.from))
	queue[0] = field.position(start)

	for head, steps := 0, 0; head < len(queue) && steps != limit; steps++ {
		for ring := len(queue); head < ring; head++ {
			current := field.index(queue[head])
			if goal != nil && current == *goal {
				return field
			}

			for _, offset := range g.topology.offsets(current) {
				next := current.Add(offset)
				if !g.Contains(next) || field.from[field.position(next)] != 0 || !passable(g.Cell(next)) {
					continue
				}

				field.from[field.position(next)] = queue[head] + 1
				field.reached++
				queue = append(queue, field.position(next))
			}
		}
	}

	return field
}

// UniformCostSearchField returns the field of every cell reached from start, with the cheapest
// way to each. Uniform cost search is also known as Dijkstra's algorithm.
func (g *Grid[T]) UniformCostSearchField(start geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	return g.uniformCostSearch(start, nil, passable, cost)
}

// UniformCostSearch returns the cheapest path from start to goal, both included, or nil where
// none is found. Uniform cost search is also known as Dijkstra's algorithm.
func (g *Grid[T]) UniformCostSearch(start, goal geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return g.uniformCostSearch(start, &goal, passable, cost).PathTo(goal)
}

// uniformCostSearch expands the cheapest cell reached so far, to the goal where one is given.
func (g *Grid[T]) uniformCostSearch(start geom.Point[int], goal *geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	passable = passableOrAny(passable)
	cost = costOrUnit(cost)

	if !g.checkPath(start, goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, true)

	var open frontier
	open.push(field.position(start), 0)

	for open.len() > 0 {
		position := open.pop()
		current := field.index(position)
		if goal != nil && current == *goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.Contains(next) || !passable(g.Cell(next)) {
				continue
			}

			reached := field.costs[position] + cost(g.Cell(current), g.Cell(next))
			if field.improve(next, position, reached) {
				open.push(field.position(next), reached)
			}
		}
	}

	return field
}

// GreedyBestFirstSearch returns a path from start to goal, both included, or nil where none is
// found, heading for the cell nearest to the goal at every step. The path is not always the
// shortest.
func (g *Grid[T]) GreedyBestFirstSearch(start, goal geom.Point[int], passable func(Cell[T]) bool) []geom.Point[int] {
	return g.greedyBestFirstSearch(start, goal, passable).PathTo(goal)
}

// greedyBestFirstSearch expands the cell reached so far that is nearest to the goal.
func (g *Grid[T]) greedyBestFirstSearch(start, goal geom.Point[int], passable func(Cell[T]) bool) Field {
	passable = passableOrAny(passable)

	if !g.checkPath(start, &goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, false)

	var open frontier
	open.push(field.position(start), 0)

	for open.len() > 0 {
		position := open.pop()
		current := field.index(position)
		if current == goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.Contains(next) || field.from[field.position(next)] != 0 || !passable(g.Cell(next)) {
				continue
			}

			field.from[field.position(next)] = position + 1
			field.reached++
			open.push(field.position(next), g.heuristic(next, goal))
		}
	}

	return field
}

// AStar returns the cheapest path from start to goal, both included, or nil where none is
// found. A cost below one per step can make the path dearer than the cheapest.
func (g *Grid[T]) AStar(start, goal geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return g.astar(start, goal, passable, cost).PathTo(goal)
}

// astar expands the cell whose cost so far and distance to the goal sum to the least.
func (g *Grid[T]) astar(start, goal geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) Field {
	passable = passableOrAny(passable)
	cost = costOrUnit(cost)

	if !g.checkPath(start, &goal, passable) {
		return Field{}
	}

	field := newField(g.Size(), start, true)

	var open frontier
	open.push(field.position(start), 0)

	for open.len() > 0 {
		position := open.pop()
		current := field.index(position)
		if current == goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.Contains(next) || !passable(g.Cell(next)) {
				continue
			}

			reached := field.costs[position] + cost(g.Cell(current), g.Cell(next))
			if field.improve(next, position, reached) {
				open.push(field.position(next), reached+g.heuristic(next, goal))
			}
		}
	}

	return field
}

// Path returns the path of AStar from one index to the other, both included, or nil where none
// is found.
func (g *Grid[T]) Path(from, to geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return g.AStar(from, to, passable, cost)
}

// heuristic returns the least a path from current to goal can cost at one per step.
func (g *Grid[T]) heuristic(current, goal geom.Point[int]) float64 {
	return float64(g.Distance(current, goal))
}

// checkPath reports whether a search can start: the start, and the goal where one is given,
// are inside the grid and passable.
func (g *Grid[T]) checkPath(start geom.Point[int], goal *geom.Point[int], passable func(Cell[T]) bool) bool {
	if !g.Contains(start) || (goal != nil && !g.Contains(*goal)) {
		return false
	}

	return passable(g.Cell(start)) && (goal == nil || passable(g.Cell(*goal)))
}

// passableOrAny returns passable, or a function admitting every cell where it is nil.
func passableOrAny[T any](passable func(Cell[T]) bool) func(Cell[T]) bool {
	if passable != nil {
		return passable
	}

	return func(Cell[T]) bool {
		return true
	}
}

// costOrUnit returns cost, or a function charging one per step where it is nil.
func costOrUnit[T any](cost func(current, next Cell[T]) float64) func(current, next Cell[T]) float64 {
	if cost != nil {
		return cost
	}

	return func(_, _ Cell[T]) float64 {
		return 1
	}
}

// frontier is the cells a search has reached and not yet expanded, a binary heap that hands
// out the one of the least priority first. It holds its items by value, and a cell reached
// again more cheaply is pushed again rather than moved.
type frontier struct {
	items []frontierItem
}

type frontierItem struct {
	position int32
	priority float64
}

func (f *frontier) len() int {
	return len(f.items)
}

// push adds a position with its priority and sifts it up to its place.
func (f *frontier) push(position int32, priority float64) {
	f.items = append(f.items, frontierItem{position, priority})

	for i := len(f.items) - 1; i > 0; {
		parent := (i - 1) / 2
		if f.items[parent].priority <= f.items[i].priority {
			break
		}

		f.items[parent], f.items[i] = f.items[i], f.items[parent]
		i = parent
	}
}

// pop removes the position of the least priority and sifts the last item down from the top.
func (f *frontier) pop() int32 {
	top := f.items[0].position

	last := len(f.items) - 1
	f.items[0] = f.items[last]
	f.items = f.items[:last]

	for i := 0; ; {
		least := i
		for child := 2*i + 1; child <= 2*i+2 && child < last; child++ {
			if f.items[child].priority < f.items[least].priority {
				least = child
			}
		}
		if least == i {
			break
		}

		f.items[i], f.items[least] = f.items[least], f.items[i]
		i = least
	}

	return top
}
