package grid

import (
	"github.com/gravitton/container"
)

// Region is the [container.Set] of grid indices every area query returns: the cells a unit can
// reach, a room, the overlap of two ranges. Its indices are ordered by x and then by y. It knows
// nothing of the grid they belong to, so what depends on which cells are adjacent is a method of
// [Grid]. The zero value is the empty region.
type Region = container.Set[Index]

// Reg returns the region of the given indices, in any order and with any repeats. The region
// does not change with a slice spread into it.
func Reg(indices ...Index) Region {
	return container.SetOf(indices...)
}
