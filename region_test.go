package grid

import (
	"fmt"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

var testRegionEmpty = Region{}

func TestRegion_Constructor(t *testing.T) {
	t.Run("indices give a region, sorted and without repeats", func(t *testing.T) {
		var region Region = Reg(geom.Pt(1, 2), geom.Pt(-1, 0), geom.Pt(1, 2), geom.Pt(1, -3))

		assert.Equal(t, region.Elements(), []Index{geom.Pt(-1, 0), geom.Pt(1, -3), geom.Pt(1, 2)})
	})
	t.Run("copies a slice spread into it", func(t *testing.T) {
		points := []Index{geom.Pt(0, 0), geom.Pt(1, 0)}
		region := Reg(points...)
		points[0] = geom.Pt(9, 9)

		assert.True(t, region.Contains(geom.Pt(0, 0)))
	})
	t.Run("no indices and a nil slice are the empty region", func(t *testing.T) {
		var none []Index

		assert.True(t, Reg().Equal(testRegionEmpty))
		assert.True(t, Reg(none...).Equal(testRegionEmpty))
	})
	t.Run("allocates once, for the copy", func(t *testing.T) {
		points := []Index{geom.Pt(1, 0), geom.Pt(0, 0), geom.Pt(1, 0)}

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkRegion = Reg(points...)
		}), 1.0)
	})
}

func ExampleReg() {
	reach := Reg(Sq(0, 0).Range(1, MovementDiagonal)...)
	blast := Reg(Sq(1, 0).Range(1, MovementDiagonal)...)

	fmt.Println(reach.Intersection(blast).Elements())
	fmt.Println(reach.Difference(blast).Len(), reach.Union(blast).Len(), reach.Contains(geom.Pt(1, 0)))
	// Output:
	// [(0,-1) (0,0) (0,1) (1,-1) (1,0) (1,1)]
	// 3 12 true
}
