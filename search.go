package grid

import (
	"cmp"
	"slices"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/x/container/heap"
	"github.com/gravitton/x/container/queue"
)

// BreadthFirstSearchField returns a map of points to the previous point in the path.
func (g *Grid[T]) BreadthFirstSearchField(start geom.Point[int], passable func(Cell[T]) bool) map[geom.Point[int]]geom.Point[int] {
	return g.breadthFirstSearch(start, nil, passable)
}

// BreadthFirstSearch returns a path from start to goal.
func (g *Grid[T]) BreadthFirstSearch(start, goal geom.Point[int], passable func(Cell[T]) bool) []geom.Point[int] {
	return g.reconstructPath(g.breadthFirstSearch(start, &goal, passable), start, goal)
}

// breadthFirstSearch implements algorithm for Breadth First Search
func (g *Grid[T]) breadthFirstSearch(start geom.Point[int], goal *geom.Point[int], passable func(Cell[T]) bool) map[geom.Point[int]]geom.Point[int] {
	passable = passableOrAny(passable)

	if !g.checkPath(start, goal, passable) {
		return nil
	}

	frontier := queue.New[geom.Point[int]]()
	frontier.Push(start)

	cameFrom := map[geom.Point[int]]geom.Point[int]{}
	cameFrom[start] = start

	for frontier.Len() > 0 {
		current := frontier.Pop()

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.valid(next) {
				continue
			}

			if _, visited := cameFrom[next]; !visited && passable(g.cell(next)) {
				cameFrom[next] = current
				frontier.Push(next)
			}
		}
	}

	return cameFrom
}

// UniformCostSearchField returns a map of points to the previous point in the path, found by
// uniform cost search, also known as Dijkstra's algorithm.
func (g *Grid[T]) UniformCostSearchField(start geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) map[geom.Point[int]]geom.Point[int] {
	return g.uniformCostSearch(start, nil, passable, cost)
}

// UniformCostSearch returns the cheapest path from start to goal. It is also known as
// Dijkstra's algorithm.
func (g *Grid[T]) UniformCostSearch(start, goal geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return g.reconstructPath(g.uniformCostSearch(start, &goal, passable, cost), start, goal)
}

// uniformCostSearch implements algorithm for Uniform Cost Search (or Dijkstra’s Algorithm)
func (g *Grid[T]) uniformCostSearch(start geom.Point[int], goal *geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) map[geom.Point[int]]geom.Point[int] {
	passable = passableOrAny(passable)

	if !g.checkPath(start, goal, passable) {
		return nil
	}

	cost = costOrUnit(cost)

	frontier := heap.NewComparable[priorityItem]()
	frontier.Push(&priorityItem{start, 0})

	cameFrom := map[geom.Point[int]]geom.Point[int]{}
	cameFrom[start] = start

	costSoFar := map[geom.Point[int]]float64{}
	costSoFar[start] = 0

	for frontier.Len() > 0 {
		current := frontier.Pop().next

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.valid(next) {
				continue
			}

			if !passable(g.cell(next)) {
				continue
			}

			newCost := costSoFar[current] + cost(g.cell(current), g.cell(next))

			if currentCost, visited := costSoFar[next]; !visited || newCost < currentCost {
				costSoFar[next] = newCost
				cameFrom[next] = current
				frontier.Push(&priorityItem{next, newCost})
			}
		}
	}

	return cameFrom
}

// GreedyBestFirstSearch returns a path from start to goal.
func (g *Grid[T]) GreedyBestFirstSearch(start, goal geom.Point[int], passable func(Cell[T]) bool) []geom.Point[int] {
	return g.reconstructPath(g.greedyBestFirstSearch(start, &goal, passable), start, goal)
}

// greedyBestFirstSearch implements algorithm for Greedy Best First Search
func (g *Grid[T]) greedyBestFirstSearch(start geom.Point[int], goal *geom.Point[int], passable func(Cell[T]) bool) map[geom.Point[int]]geom.Point[int] {
	passable = passableOrAny(passable)

	if !g.checkPath(start, goal, passable) {
		return nil
	}

	frontier := heap.NewComparable[priorityItem]()
	frontier.Push(&priorityItem{start, 0})

	cameFrom := map[geom.Point[int]]geom.Point[int]{}
	cameFrom[start] = start

	for frontier.Len() > 0 {
		current := frontier.Pop().next

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.valid(next) {
				continue
			}

			if _, visited := cameFrom[next]; !visited && passable(g.cell(next)) {
				cameFrom[next] = current
				frontier.Push(&priorityItem{next, g.heuristic(next, *goal)})
			}
		}
	}

	return cameFrom
}

// AStar returns a path from start to goal.
func (g *Grid[T]) AStar(start, goal geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return g.reconstructPath(g.astar(start, &goal, passable, cost), start, goal)
}

// astar implements algorithm for A* Search
func (g *Grid[T]) astar(start geom.Point[int], goal *geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) map[geom.Point[int]]geom.Point[int] {
	passable = passableOrAny(passable)

	if !g.checkPath(start, goal, passable) {
		return nil
	}

	cost = costOrUnit(cost)

	frontier := heap.NewComparable[priorityItem]()
	frontier.Push(&priorityItem{start, 0})

	cameFrom := map[geom.Point[int]]geom.Point[int]{}
	cameFrom[start] = start

	costSoFar := map[geom.Point[int]]float64{}
	costSoFar[start] = 0

	for frontier.Len() > 0 {
		current := frontier.Pop().next

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, offset := range g.topology.offsets(current) {
			next := current.Add(offset)
			if !g.valid(next) {
				continue
			}

			if !passable(g.cell(next)) {
				continue
			}

			newCost := costSoFar[current] + cost(g.cell(current), g.cell(next))

			if currentCost, visited := costSoFar[next]; !visited || newCost < currentCost {
				costSoFar[next] = newCost
				cameFrom[next] = current
				frontier.Push(&priorityItem{next, newCost + g.heuristic(next, *goal)})
			}
		}
	}

	return cameFrom
}

// Path returns the cheapest path from start to goal, both included, or nil where there is
// none. It is [Grid.AStar], the search to reach for first. A nil passable admits every cell and
// a nil cost charges one per step, as in every search.
func (g *Grid[T]) Path(from, to geom.Point[int], passable func(Cell[T]) bool, cost func(current, next Cell[T]) float64) []geom.Point[int] {
	return g.AStar(from, to, passable, cost)
}

func (g *Grid[T]) heuristic(current, goal geom.Point[int]) float64 {
	return float64(g.Distance(current, goal))
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

func (g *Grid[T]) checkPath(start geom.Point[int], goal *geom.Point[int], passable func(Cell[T]) bool) bool {
	// check if start and goal are valid grid indexes
	if !g.valid(start) || (goal != nil && !g.valid(*goal)) {
		return false
	}

	// check if start and goal are passable
	if !passable(g.cell(start)) || (goal != nil && !passable(g.cell(*goal))) {
		return false
	}

	return true
}

func (g *Grid[T]) reconstructPath(fields map[geom.Point[int]]geom.Point[int], start, goal geom.Point[int]) []geom.Point[int] {
	if fields == nil {
		return nil
	}

	path := []geom.Point[int]{}
	current := goal

	for current != start {
		path = append(path, current)

		if next, ok := fields[current]; ok {
			current = next
		} else {
			return nil
		}
	}

	path = append(path, start)
	slices.Reverse(path)

	return path
}

type priorityItem struct {
	next     geom.Point[int]
	priority float64
}

func (i *priorityItem) Compare(item *priorityItem) int {
	return cmp.Compare(i.priority, item.priority)
}
