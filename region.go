package grid

import (
	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

// Region is the [container.Set] of grid indices every area query returns: the cells a unit can
// reach, a room, the overlap of two ranges. Its indices are ordered by x and then by y. It knows
// nothing of the grid they belong to, so what depends on which cells are adjacent is a method of
// [Grid]. The zero value is the empty region.
type Region = container.Set[geom.Point[int]]

// Reg returns the region of the given indices, in any order and with any repeats. The region
// does not change with the slice.
func Reg(indices []geom.Point[int]) Region {
	return container.SetOf(indices...)
}
