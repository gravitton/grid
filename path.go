package grid

import (
	"cmp"
	"slices"

	"github.com/gravitton/geometry/types/ints"
	"github.com/gravitton/x/container/heap"
	"github.com/gravitton/x/container/queue"
	xslices "github.com/gravitton/x/slices"
)

// Distance returns the distance between two points.
func (g *Grid[T]) Distance(from, to ints.Point) int {
	return g.distance(from, to)
}

// Range returns the indices within n steps of index, under the movement of the grid, that are
// valid and in sight of it. A nil valid admits every cell.
func (g *Grid[T]) Range(index ints.Point, n int, valid ValidFunc[T]) []ints.Point {
	valid = validOrAny(valid)

	return g.toRange(index, n, func(index ints.Point) bool {
		return g.valid(index) && valid(g.cell(index))
	})
}

// Path returns a path from start to goal.
// A nil valid admits every cell and a nil cost charges one per step, as in every search.
func (g *Grid[T]) Path(from, to ints.Point, valid ValidFunc[T], cost CostFunc[T]) []*Cell[T] {
	return xslices.Map(g.AStar(from, to, valid, cost), func(point ints.Point) *Cell[T] {
		return g.cell(point)
	})
}

// BreadthFirstSearchField returns a map of points to the previous point in the path.
func (g *Grid[T]) BreadthFirstSearchField(start ints.Point, valid ValidFunc[T]) map[ints.Point]ints.Point {
	return g.breadthFirstSearch(start, nil, valid)
}

// BreadthFirstSearch returns a path from start to goal.
func (g *Grid[T]) BreadthFirstSearch(start, goal ints.Point, valid ValidFunc[T]) []ints.Point {
	return g.reconstructPath(g.breadthFirstSearch(start, &goal, valid), start, goal)
}

// breadthFirstSearch implements algorithm for Breadth First Search
func (g *Grid[T]) breadthFirstSearch(start ints.Point, goal *ints.Point, valid ValidFunc[T]) map[ints.Point]ints.Point {
	valid = validOrAny(valid)

	if !g.checkPath(start, goal, valid) {
		return nil
	}

	frontier := queue.New[ints.Point]()
	frontier.Push(start)

	cameFrom := map[ints.Point]ints.Point{}
	cameFrom[start] = start

	for frontier.Len() > 0 {
		current := frontier.Pop()

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, next := range g.cellNeighbors(current) {
			if _, visited := cameFrom[next]; !visited && valid(g.cell(next)) {
				cameFrom[next] = current
				frontier.Push(next)
			}
		}
	}

	return cameFrom
}

// UniformCostSearchField returns a map of points to the previous point in the path, found by
// uniform cost search, also known as Dijkstra's algorithm.
func (g *Grid[T]) UniformCostSearchField(start ints.Point, valid ValidFunc[T], cost CostFunc[T]) map[ints.Point]ints.Point {
	return g.uniformCostSearch(start, nil, valid, cost)
}

// UniformCostSearch returns the cheapest path from start to goal. It is also known as
// Dijkstra's algorithm.
func (g *Grid[T]) UniformCostSearch(start, goal ints.Point, valid ValidFunc[T], cost CostFunc[T]) []ints.Point {
	return g.reconstructPath(g.uniformCostSearch(start, &goal, valid, cost), start, goal)
}

// uniformCostSearch implements algorithm for Uniform Cost Search (or Dijkstra’s Algorithm)
func (g *Grid[T]) uniformCostSearch(start ints.Point, goal *ints.Point, valid ValidFunc[T], cost CostFunc[T]) map[ints.Point]ints.Point {
	valid = validOrAny(valid)

	if !g.checkPath(start, goal, valid) {
		return nil
	}

	cost = costOrUnit(cost)

	frontier := heap.NewComparable[priorityItem]()
	frontier.Push(&priorityItem{start, 0})

	cameFrom := map[ints.Point]ints.Point{}
	cameFrom[start] = start

	costSoFar := map[ints.Point]float64{}
	costSoFar[start] = 0

	for frontier.Len() > 0 {
		current := frontier.Pop().next

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, next := range g.cellNeighbors(current) {
			if !valid(g.cell(next)) {
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
func (g *Grid[T]) GreedyBestFirstSearch(start, goal ints.Point, valid ValidFunc[T]) []ints.Point {
	return g.reconstructPath(g.greedyBestFirstSearch(start, &goal, valid), start, goal)
}

// greedyBestFirstSearch implements algorithm for Greedy Best First Search
func (g *Grid[T]) greedyBestFirstSearch(start ints.Point, goal *ints.Point, valid ValidFunc[T]) map[ints.Point]ints.Point {
	valid = validOrAny(valid)

	if !g.checkPath(start, goal, valid) {
		return nil
	}

	frontier := heap.NewComparable[priorityItem]()
	frontier.Push(&priorityItem{start, 0})

	cameFrom := map[ints.Point]ints.Point{}
	cameFrom[start] = start

	for frontier.Len() > 0 {
		current := frontier.Pop().next

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, next := range g.cellNeighbors(current) {
			if _, visited := cameFrom[next]; !visited && valid(g.cell(next)) {
				cameFrom[next] = current
				frontier.Push(&priorityItem{next, g.heuristic(next, *goal)})
			}
		}
	}

	return cameFrom
}

// AStar returns a path from start to goal.
func (g *Grid[T]) AStar(start, goal ints.Point, valid ValidFunc[T], cost CostFunc[T]) []ints.Point {
	return g.reconstructPath(g.astar(start, &goal, valid, cost), start, goal)
}

// astar implements algorithm for A* Search
func (g *Grid[T]) astar(start ints.Point, goal *ints.Point, valid ValidFunc[T], cost CostFunc[T]) map[ints.Point]ints.Point {
	valid = validOrAny(valid)

	if !g.checkPath(start, goal, valid) {
		return nil
	}

	cost = costOrUnit(cost)

	frontier := heap.NewComparable[priorityItem]()
	frontier.Push(&priorityItem{start, 0})

	cameFrom := map[ints.Point]ints.Point{}
	cameFrom[start] = start

	costSoFar := map[ints.Point]float64{}
	costSoFar[start] = 0

	for frontier.Len() > 0 {
		current := frontier.Pop().next

		// early exit on reaching goal
		if goal != nil && current == *goal {
			break
		}

		for _, next := range g.cellNeighbors(current) {
			if !valid(g.cell(next)) {
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

type priorityItem struct {
	next     ints.Point
	priority float64
}

func (i *priorityItem) Compare(item *priorityItem) int {
	return cmp.Compare(i.priority, item.priority)
}

func (g *Grid[T]) heuristic(current, goal ints.Point) float64 {
	return float64(g.Distance(current, goal))
}

// validOrAny returns valid, or a function admitting every cell where it is nil.
func validOrAny[T any](valid ValidFunc[T]) ValidFunc[T] {
	if valid != nil {
		return valid
	}

	return func(*Cell[T]) bool {
		return true
	}
}

// costOrUnit returns cost, or a function charging one per step where it is nil.
func costOrUnit[T any](cost CostFunc[T]) CostFunc[T] {
	if cost != nil {
		return cost
	}

	return func(_, _ *Cell[T]) float64 {
		return 1
	}
}

func (g *Grid[T]) checkPath(start ints.Point, goal *ints.Point, valid ValidFunc[T]) bool {
	// check if start and goal are valid grid indexes
	if !g.valid(start) || (goal != nil && !g.valid(*goal)) {
		return false
	}

	// check if start and goal are valid with user defined function
	if !valid(g.cell(start)) || (goal != nil && !valid(g.cell(*goal))) {
		return false
	}

	return true
}

func (g *Grid[T]) reconstructPath(fields map[ints.Point]ints.Point, start, goal ints.Point) []ints.Point {
	if fields == nil {
		return nil
	}

	path := []ints.Point{}
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
