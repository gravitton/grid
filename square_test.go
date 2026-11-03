package grid

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

var testSquare = Sq(3, -2)

func TestSquare_Constructor(t *testing.T) {
	assert.Equal(t, Sq(3, -2).Point(), geom.Pt(3, -2))
}

func TestRangeLen(t *testing.T) {
	t.Run("is the length of Range", func(t *testing.T) {
		for _, movement := range Movements() {
			for n := -1; n < 6; n++ {
				assert.Equal(t, RangeLen(n, movement), len(testSquare.Range(n, movement)), movement.String()+" at "+strconv.Itoa(n)+": ")
			}
		}
	})
	t.Run("counts the index alone outside the constants", func(t *testing.T) {
		assert.Equal(t, RangeLen(3, MovementNone), 1)
		assert.Equal(t, RangeLen(-1, MovementNone), 0)
	})
}

func TestRingLen(t *testing.T) {
	t.Run("is the length of Ring", func(t *testing.T) {
		for _, movement := range Movements() {
			for n := -1; n < 6; n++ {
				assert.Equal(t, RingLen(n, movement), len(testSquare.Ring(n, movement)), movement.String()+" at "+strconv.Itoa(n)+": ")
			}
		}
	})
	t.Run("counts nothing past the index itself outside the constants", func(t *testing.T) {
		assert.Equal(t, RingLen(0, MovementNone), 1)
		assert.Equal(t, RingLen(3, MovementNone), 0)
	})
}

func TestSectorLen(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("is the length of Sector", func(t *testing.T) {
				for turns := range 10 {
					for radius := range 9 {
						assert.Equal(t, SectorLen(radius, turns, movement), len(testSquare.Sector(radius, geom.DirectionUp, turns, movement)), strconv.Itoa(turns)+": ")
					}
				}
			})
			t.Run("is zero for a negative radius or negative turns", func(t *testing.T) {
				assert.Equal(t, SectorLen(-1, 1, movement), 0)
				assert.Equal(t, SectorLen(3, -1, movement), 0)
			})
		})
	}
	t.Run("is r+t·r(r+1)/2, less ⌈r/2⌉ for odd t", func(t *testing.T) {
		assert.Equal(t, SectorLen(4, 1, MovementCardinal), 12)
		assert.Equal(t, SectorLen(4, 2, MovementDiagonal), 24)
	})
	t.Run("is every ring from as many turns as directions on", func(t *testing.T) {
		assert.Equal(t, SectorLen(3, 4, MovementCardinal), RangeLen(3, MovementCardinal)-1)
		assert.Equal(t, SectorLen(3, 99, MovementDiagonal), RangeLen(3, MovementDiagonal)-1)
	})
	t.Run("is zero for a movement outside the constants", func(t *testing.T) {
		assert.Equal(t, SectorLen(3, 1, MovementNone), 0)
	})
}

func TestSquare_Turn(t *testing.T) {
	t.Run("a quarter turn is clockwise on a screen", func(t *testing.T) {
		assert.Equal(t, Sq(1, 0).Turn(1), Sq(0, 1))
		assert.Equal(t, Sq(3, -2).Turn(1), Sq(2, 3))
	})
	t.Run("negative steps turn the other way", func(t *testing.T) {
		assert.Equal(t, Sq(1, 0).Turn(-1), Sq(0, -1))
		assert.Equal(t, testSquare.Turn(-1), testSquare.Turn(3))
	})
	t.Run("four steps are a full turn", func(t *testing.T) {
		for steps := -8; steps <= 8; steps++ {
			assert.Equal(t, testSquare.Turn(steps).Turn(4-steps), testSquare, strconv.Itoa(steps)+": ")
		}
	})
	t.Run("turns a neighbor as geom turns its direction", func(t *testing.T) {
		for _, direction := range geom.Directions() {
			assert.Equal(t, Sq(0, 0).Neighbor(direction).Turn(1), Sq(0, 0).Neighbor(direction.Turn(2)), direction.String()+": ")
		}
	})
	t.Run("keeps the distance from the origin", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.Equal(t, testSquare.Turn(1).DistanceTo(geom.Pt(0, 0), movement), testSquare.DistanceTo(geom.Pt(0, 0), movement), movement.String()+": ")
		}
	})
}

func TestSquare_TurnAround(t *testing.T) {
	center := geom.Pt(1, 1)

	t.Run("turns around the center", func(t *testing.T) {
		assert.Equal(t, Sq(3, 1).TurnAround(center, 1), Sq(1, 3))
		assert.Equal(t, Sq(3, 1).TurnAround(center, 2), Sq(-1, 1))
	})
	t.Run("the center stays", func(t *testing.T) {
		assert.Equal(t, Square(center).TurnAround(center, 3), Square(center))
	})
	t.Run("around the origin is Turn", func(t *testing.T) {
		assert.Equal(t, testSquare.TurnAround(geom.Pt(0, 0), 3), testSquare.Turn(3))
	})
}

func TestSquare_ReflectX(t *testing.T) {
	t.Run("negates the row", func(t *testing.T) {
		assert.Equal(t, Sq(3, -2).ReflectX(), Sq(3, 2))
	})
	t.Run("twice is the point", func(t *testing.T) {
		assert.Equal(t, testSquare.ReflectX().ReflectX(), testSquare)
	})
}

func TestSquare_ReflectY(t *testing.T) {
	t.Run("negates the column", func(t *testing.T) {
		assert.Equal(t, Sq(3, -2).ReflectY(), Sq(-3, -2))
	})
	t.Run("both reflections are a half turn", func(t *testing.T) {
		assert.Equal(t, testSquare.ReflectX().ReflectY(), testSquare.Turn(2))
	})
}

func TestSquare_Neighbor(t *testing.T) {
	t.Run("steps once in the direction", func(t *testing.T) {
		assert.Equal(t, testSquare.Neighbor(geom.DirectionRight), Sq(4, -2))
		assert.Equal(t, testSquare.Neighbor(geom.DirectionUpLeft), Sq(2, -3))
	})
	t.Run("none stays", func(t *testing.T) {
		assert.Equal(t, testSquare.Neighbor(geom.DirectionNone), testSquare)
	})
}

func TestSquare_Neighbors(t *testing.T) {
	t.Run("cardinal is the four edge neighbors by increasing angle", func(t *testing.T) {
		assert.Equal(t, Sq(3, -2).Neighbors(MovementCardinal), []Index{geom.Pt(4, -2), geom.Pt(3, -1), geom.Pt(2, -2), geom.Pt(3, -3)})
	})
	t.Run("diagonal is all eight", func(t *testing.T) {
		assert.Length(t, Sq(3, -2).Neighbors(MovementDiagonal), 8)
	})
	t.Run("has none outside the constants", func(t *testing.T) {
		assert.Nil(t, testSquare.Neighbors(MovementNone))
		assert.Length(t, slices.Collect(testSquare.NeighborsSeq(Movement(99))), 0)
	})
}

func TestSquare_NeighborsSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Neighbors returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testSquare.NeighborsSeq(movement)), testSquare.Neighbors(movement))
			})
			t.Run("allocates nothing into a buffer with room", func(t *testing.T) {
				buffer := make([]Index, 0, 8)

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = slices.AppendSeq(buffer[:0], testSquare.NeighborsSeq(movement))
				}), 0.0)
			})
		})
	}
}

func TestSquare_Range(t *testing.T) {
	center := Sq(3, -2)

	t.Run("a negative radius is nil", func(t *testing.T) {
		assert.Equal(t, center.Range(-1, MovementCardinal), nil)
	})
	t.Run("a zero radius is the point alone", func(t *testing.T) {
		assert.Equal(t, center.Range(0, MovementDiagonal), []Index{geom.Pt(3, -2)})
	})
	t.Run("cardinal is the diamond of Manhattan distance", func(t *testing.T) {
		assert.Equal(t, center.Range(1, MovementCardinal), []Index{geom.Pt(2, -2), geom.Pt(3, -3), geom.Pt(3, -2), geom.Pt(3, -1), geom.Pt(4, -2)})
		assert.Length(t, center.Range(3, MovementCardinal), 25)
	})
	t.Run("diagonal is the square of Chebyshev distance", func(t *testing.T) {
		assert.Length(t, center.Range(1, MovementDiagonal), 9)
		assert.Length(t, center.Range(3, MovementDiagonal), 49)
	})
	t.Run("holds exactly the indices within the distance", func(t *testing.T) {
		for _, movement := range Movements() {
			for n := range 5 {
				indices := center.Range(n, movement)

				for _, index := range indices {
					assert.LessOrEqual(t, center.DistanceTo(index, movement), n, movement.String()+" at "+index.String()+": ")
				}
				for _, index := range center.Range(n+1, movement) {
					assert.Equal(t, Reg(indices).Contains(index), center.DistanceTo(index, movement) <= n, movement.String()+" at "+index.String()+": ")
				}
			}
		}
	})
	t.Run("is sorted and allocated once at its length", func(t *testing.T) {
		for _, movement := range Movements() {
			indices := center.Range(4, movement)

			assert.True(t, slices.IsSortedFunc(indices, Index.Compare), movement.String()+": ")
			assert.Equal(t, cap(indices), len(indices), movement.String()+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = center.Range(4, movement)
			}), 1.0, movement.String()+": ")
		}
	})
	t.Run("is the index alone outside the constants", func(t *testing.T) {
		assert.Equal(t, center.Range(3, MovementNone), []Index{center.Point()})
		assert.Nil(t, center.Range(-1, MovementNone))
	})
}

func BenchmarkSquare_Range(b *testing.B) {
	for _, movement := range Movements() {
		b.Run(movement.String(), func(b *testing.B) {
			for b.Loop() {
				sinkPoints = testSquare.Range(8, movement)
			}
		})
	}
}

func TestSquare_RangeSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Range returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testSquare.RangeSeq(3, movement)), testSquare.Range(3, movement))
			})
			t.Run("stops at the break", func(t *testing.T) {
				count := 0
				for range testSquare.RangeSeq(3, movement) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
			t.Run("allocates nothing", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					for index := range testSquare.RangeSeq(3, movement) {
						sinkPoint = index
					}
				}), 0.0)
			})
		})
	}
}

func TestSquare_Ring(t *testing.T) {
	t.Run("a negative radius is nil", func(t *testing.T) {
		assert.Equal(t, testSquare.Ring(-1, MovementCardinal), nil)
	})
	t.Run("a zero radius is the point alone", func(t *testing.T) {
		assert.Equal(t, testSquare.Ring(0, MovementDiagonal), []Index{testSquare.Point()})
	})
	t.Run("cardinal walks the diamond by increasing angle", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).Ring(2, MovementCardinal), []Index{
			geom.Pt(2, 0), geom.Pt(1, 1), geom.Pt(0, 2), geom.Pt(-1, 1), geom.Pt(-2, 0), geom.Pt(-1, -1), geom.Pt(0, -2), geom.Pt(1, -1),
		})
	})
	t.Run("diagonal walks the square by increasing angle", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).Ring(1, MovementDiagonal), []Index{
			geom.Pt(1, 0), geom.Pt(1, 1), geom.Pt(0, 1), geom.Pt(-1, 1), geom.Pt(-1, 0), geom.Pt(-1, -1), geom.Pt(0, -1), geom.Pt(1, -1),
		})
	})
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("a radius of one is the neighbors", func(t *testing.T) {
				assert.Equal(t, testSquare.Ring(1, movement), testSquare.Neighbors(movement))
			})
			t.Run("holds exactly the indices at the distance", func(t *testing.T) {
				for n := range 6 {
					ring := testSquare.Ring(n, movement)

					assert.Equal(t, Reg(ring).Len(), len(ring), strconv.Itoa(n)+": ")
					for _, index := range testSquare.Range(n+1, movement) {
						assert.Equal(t, slices.Contains(ring, index), testSquare.DistanceTo(index, movement) == n, strconv.Itoa(n)+" at "+index.String()+": ")
					}
				}
			})
			t.Run("allocates once at its length", func(t *testing.T) {
				ring := testSquare.Ring(4, movement)

				assert.Equal(t, cap(ring), len(ring))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testSquare.Ring(4, movement)
				}), 1.0)
			})
		})
	}
	t.Run("is the index alone at radius zero outside the constants, and empty beyond", func(t *testing.T) {
		assert.Equal(t, testSquare.Ring(0, MovementNone), []Index{testSquare.Point()})
		assert.Length(t, testSquare.Ring(3, MovementNone), 0)
	})
}

func TestSquare_RingSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Ring returns", func(t *testing.T) {
				for n := -1; n < 4; n++ {
					assert.Equal(t, slices.Collect(testSquare.RingSeq(n, movement)), testSquare.Ring(n, movement), strconv.Itoa(n)+": ")
				}
			})
			t.Run("stops at the break", func(t *testing.T) {
				for n := range 3 {
					count := 0
					for range testSquare.RingSeq(n, movement) {
						count++

						break
					}

					assert.Equal(t, count, 1, strconv.Itoa(n)+": ")
				}
			})
			t.Run("allocates nothing", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					for index := range testSquare.RingSeq(3, movement) {
						sinkPoint = index
					}
				}), 0.0)
			})
		})
	}
}

func TestSquare_Spiral(t *testing.T) {
	t.Run("a negative radius is nil", func(t *testing.T) {
		assert.Equal(t, testSquare.Spiral(-1, MovementCardinal), nil)
	})
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("starts at the point", func(t *testing.T) {
				assert.Equal(t, testSquare.Spiral(3, movement)[0], testSquare.Point())
			})
			t.Run("holds what Range holds", func(t *testing.T) {
				assert.True(t, Reg(testSquare.Spiral(4, movement)).Equal(Reg(testSquare.Range(4, movement))))
			})
			t.Run("never steps back toward the point", func(t *testing.T) {
				previous := 0
				for _, index := range testSquare.Spiral(4, movement) {
					distance := testSquare.DistanceTo(index, movement)
					assert.GreaterOrEqual(t, distance, previous, index.String()+": ")
					previous = distance
				}
			})
			t.Run("allocates once at its length", func(t *testing.T) {
				spiral := testSquare.Spiral(4, movement)

				assert.Equal(t, cap(spiral), len(spiral))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testSquare.Spiral(4, movement)
				}), 1.0)
			})
		})
	}
	t.Run("is the index alone outside the constants", func(t *testing.T) {
		assert.Equal(t, testSquare.Spiral(3, MovementNone), []Index{testSquare.Point()})
	})
}

func BenchmarkSquare_Spiral(b *testing.B) {
	for _, movement := range Movements() {
		b.Run(movement.String(), func(b *testing.B) {
			for b.Loop() {
				sinkPoints = testSquare.Spiral(8, movement)
			}
		})
	}
}

func TestSquare_SpiralSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Spiral returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testSquare.SpiralSeq(3, movement)), testSquare.Spiral(3, movement))
			})
			t.Run("stops at the break, inside a ring too", func(t *testing.T) {
				count := 0
				for range testSquare.SpiralSeq(3, movement) {
					count++
					if count == 3 {
						break
					}
				}

				assert.Equal(t, count, 3)
			})
		})
	}
}

func TestSquare_Sector(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("sectors of one turn cover the range but s, sharing only their borders", func(t *testing.T) {
				covered := map[Index]int{}
				for _, direction := range movement.Directions() {
					for _, index := range testSquare.Sector(6, direction, 1, movement) {
						covered[index]++
					}
				}

				for _, index := range testSquare.Range(6, movement) {
					times := covered[index]
					if index == testSquare.Point() {
						assert.Equal(t, times, 0)

						continue
					}

					assert.True(t, times == 1 || times == 2 && testSquare.DistanceTo(index, movement)%2 == 0, index.String()+": ")
				}
			})
			t.Run("zero turns is the ray along the direction", func(t *testing.T) {
				for _, direction := range movement.Directions() {
					var ray []Index
					for ring := 1; ring <= 4; ring++ {
						ray = append(ray, geom.Pt(testSquare.X+direction.Offset[int]().X*ring, testSquare.Y+direction.Offset[int]().Y*ring))
					}

					assert.Equal(t, testSquare.Sector(4, direction, 0, movement), ray, direction.String()+": ")
				}
			})
			t.Run("as many turns as directions is the range but s, each ring from the index opposite", func(t *testing.T) {
				full := testSquare.Sector(4, geom.DirectionRight, movement.neighborsLen(), movement)

				assert.True(t, Reg(full).Equal(Reg(testSquare.Range(4, movement)).Difference(Reg([]Index{testSquare.Point()}))))
				assert.Equal(t, full[0], testSquare.Neighbor(geom.DirectionLeft).Point())
				assert.Equal(t, testSquare.Sector(4, geom.DirectionRight, 99, movement), full)
			})
			t.Run("widens with its turns", func(t *testing.T) {
				for turns := range movement.neighborsLen() {
					narrow, wide := Reg(testSquare.Sector(5, geom.DirectionDown, turns, movement)), Reg(testSquare.Sector(5, geom.DirectionDown, turns+1, movement))

					assert.True(t, wide.Encloses(narrow), strconv.Itoa(turns)+": ")
				}
			})
			t.Run("is ring by ring outward", func(t *testing.T) {
				previous := 0
				for _, index := range testSquare.Sector(6, geom.DirectionUp, 2, movement) {
					distance := testSquare.DistanceTo(index, movement)
					assert.GreaterOrEqual(t, distance, previous)
					previous = distance
				}
			})
			t.Run("turns with its direction", func(t *testing.T) {
				for _, direction := range movement.Directions() {
					var turned []Index
					for _, index := range Sq(0, 0).Sector(5, direction, 3, movement) {
						turned = append(turned, Square(index).Turn(1).Point())
					}

					assert.Equal(t, turned, Sq(0, 0).Sector(5, direction.Turn(2), 3, movement), direction.String()+": ")
				}
			})
			t.Run("is nil for a radius under one, negative turns and a direction not of the movement", func(t *testing.T) {
				assert.Nil(t, testSquare.Sector(0, geom.DirectionRight, 1, movement))
				assert.Nil(t, testSquare.Sector(3, geom.DirectionRight, -1, movement))
				assert.Nil(t, testSquare.Sector(3, geom.DirectionNone, 1, movement))
			})
			t.Run("allocates once, at its length", func(t *testing.T) {
				assert.Equal(t, cap(testSquare.Sector(7, geom.DirectionUp, 2, movement)), SectorLen(7, 2, movement))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testSquare.Sector(7, geom.DirectionUp, 2, movement)
				}), 1.0)
			})
		})
	}
	t.Run("a cardinal sector of one turn is a cone by increasing angle", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).Sector(2, geom.DirectionRight, 1, MovementCardinal), []Index{geom.Pt(1, 0), geom.Pt(1, -1), geom.Pt(2, 0), geom.Pt(1, 1)})
	})
	t.Run("a diagonal sector of one turn walks around the corner of its direction", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).Sector(2, geom.DirectionDownRight, 1, MovementDiagonal), []Index{geom.Pt(1, 1), geom.Pt(2, 1), geom.Pt(2, 2), geom.Pt(1, 2)})
	})
	t.Run("is nil for a diagonal direction under MovementCardinal", func(t *testing.T) {
		assert.Nil(t, testSquare.Sector(3, geom.DirectionDownRight, 1, MovementCardinal))
	})
}

func TestSquare_SectorSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Sector returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testSquare.SectorSeq(4, geom.DirectionLeft, 2, movement)), testSquare.Sector(4, geom.DirectionLeft, 2, movement))
			})
			t.Run("yields nothing for negative turns or a direction not of the movement", func(t *testing.T) {
				assert.Length(t, slices.Collect(testSquare.SectorSeq(4, geom.DirectionLeft, -1, movement)), 0)
				assert.Length(t, slices.Collect(testSquare.SectorSeq(4, geom.DirectionNone, 1, movement)), 0)
			})
			t.Run("stops at the break wherever it comes", func(t *testing.T) {
				for stop := 1; stop <= SectorLen(4, 3, movement); stop++ {
					count := 0
					for range testSquare.SectorSeq(4, geom.DirectionLeft, 3, movement) {
						count++
						if count == stop {
							break
						}
					}

					assert.Equal(t, count, stop)
				}
			})
			t.Run("allocates nothing into a buffer with room", func(t *testing.T) {
				buffer := make([]Index, 0, 128)

				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = slices.AppendSeq(buffer[:0], testSquare.SectorSeq(7, geom.DirectionUp, 2, movement))
				}), 0.0)
			})
		})
	}
}

func TestSquare_DistanceTo(t *testing.T) {
	t.Run("cardinal is the Manhattan distance", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).DistanceTo(geom.Pt(3, 4), MovementCardinal), 7)
	})
	t.Run("diagonal is the Chebyshev distance", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).DistanceTo(geom.Pt(3, 4), MovementDiagonal), 4)
	})
}

func TestSquare_DirectionTo(t *testing.T) {
	t.Run("the point itself has no direction", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.Equal(t, testSquare.DirectionTo(testSquare.Point(), movement), geom.DirectionNone, movement.String()+": ")
		}
	})
	t.Run("a neighbor is in its own direction", func(t *testing.T) {
		for _, movement := range Movements() {
			for _, direction := range movement.Directions() {
				assert.Equal(t, testSquare.DirectionTo(testSquare.Neighbor(direction).Point(), movement), direction, movement.String()+" "+direction.String()+": ")
			}
		}
	})
	t.Run("diagonal is the nearest of the eight", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(5, 1), MovementDiagonal), geom.DirectionRight)
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(4, 3), MovementDiagonal), geom.DirectionDownRight)
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(-1, -6), MovementDiagonal), geom.DirectionUp)
	})
	t.Run("cardinal is the nearest of the four", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(4, 3), MovementCardinal), geom.DirectionRight)
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(-4, 3), MovementCardinal), geom.DirectionLeft)
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(2, 3), MovementCardinal), geom.DirectionDown)
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(2, -3), MovementCardinal), geom.DirectionUp)
	})
	t.Run("cardinal answers the horizontal on a diagonal", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(3, 3), MovementCardinal), geom.DirectionRight)
		assert.Equal(t, Sq(0, 0).DirectionTo(geom.Pt(-3, 3), MovementCardinal), geom.DirectionLeft)
	})
	t.Run("has no direction outside the constants", func(t *testing.T) {
		assert.Equal(t, testSquare.DirectionTo(geom.Pt(0, 0), MovementNone), geom.DirectionNone)
	})
}

func TestSquare_LineTo(t *testing.T) {
	t.Run("to the point itself is the point alone", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.Equal(t, testSquare.LineTo(testSquare.Point(), movement), []Index{testSquare.Point()}, movement.String()+": ")
		}
	})
	t.Run("diagonal steps across corners", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).LineTo(geom.Pt(3, 3), MovementDiagonal), []Index{geom.Pt(0, 0), geom.Pt(1, 1), geom.Pt(2, 2), geom.Pt(3, 3)})
	})
	t.Run("cardinal steps along one axis at a time", func(t *testing.T) {
		assert.Equal(t, Sq(0, 0).LineTo(geom.Pt(2, 1), MovementCardinal), []Index{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(1, 1), geom.Pt(2, 1)})
	})
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("runs from the point to the target by neighbors", func(t *testing.T) {
				for _, target := range testSquare.Range(5, MovementDiagonal) {
					line := testSquare.LineTo(target, movement)

					assert.Equal(t, line[0], testSquare.Point(), target.String()+": ")
					assert.Equal(t, line[len(line)-1], target, target.String()+": ")
					assert.Length(t, line, testSquare.DistanceTo(target, movement)+1, target.String()+": ")
					for i := 1; i < len(line); i++ {
						assert.Equal(t, Square(line[i-1]).DistanceTo(line[i], movement), 1, target.String()+": ")
					}
				}
			})
			t.Run("allocates once at its length", func(t *testing.T) {
				line := testSquare.LineTo(geom.Pt(-4, 7), movement)

				assert.Equal(t, cap(line), len(line))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testSquare.LineTo(geom.Pt(-4, 7), movement)
				}), 1.0)
			})
		})
	}
	t.Run("panics for a movement outside the constants, which draws no line", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			testSquare.LineTo(geom.Pt(0, 0), MovementNone)
		}, "grid: unknown movement -1")
		assert.PanicsWith(t, func() {
			sinkPoints = slices.Collect(testSquare.LineToSeq(geom.Pt(0, 0), Movement(99)))
		}, "grid: unknown movement 99")
	})
}

func BenchmarkSquare_LineTo(b *testing.B) {
	for _, movement := range Movements() {
		b.Run(movement.String(), func(b *testing.B) {
			for b.Loop() {
				sinkPoints = testSquare.LineTo(geom.Pt(40, 17), movement)
			}
		})
	}
}

func TestSquare_LineToSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Line returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testSquare.LineToSeq(geom.Pt(-4, 7), movement)), testSquare.LineTo(geom.Pt(-4, 7), movement))
			})
			t.Run("stops at the break, at the first index and after it", func(t *testing.T) {
				for stop := 1; stop <= 2; stop++ {
					count := 0
					for range testSquare.LineToSeq(geom.Pt(-4, 7), movement) {
						count++
						if count == stop {
							break
						}
					}

					assert.Equal(t, count, stop)
				}
			})
			t.Run("allocates nothing", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					for index := range testSquare.LineToSeq(geom.Pt(-4, 7), movement) {
						sinkPoint = index
					}
				}), 0.0)
			})
		})
	}
}

func blockedBy(indices []Index) func(Index) bool {
	return func(index Index) bool {
		return slices.Contains(indices, index)
	}
}

func TestSquare_HasLineOfSight(t *testing.T) {
	origin := Sq(0, 0)

	t.Run("clear", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), nil))
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), blockedBy([]Index{})))
		assert.True(t, origin.HasLineOfSight(origin.Point(), nil))
	})
	t.Run("allocates nothing", func(t *testing.T) {
		blocked := blockedBy([]Index{geom.Pt(1, 1)})

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkBool = origin.HasLineOfSight(geom.Pt(6, 4), blocked)
		}), 0.0)
	})
	t.Run("blocked mid", func(t *testing.T) {
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), blockedBy([]Index{geom.Pt(1, 0)})))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), blockedBy([]Index{geom.Pt(2, 0)})))
	})
	t.Run("blocker beyond target", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 0), blockedBy([]Index{geom.Pt(3, 0)})))
	})
	t.Run("target in blocking list", func(t *testing.T) {
		// can see into, not through
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), blockedBy([]Index{geom.Pt(3, 0)})))
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 3), nil))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 3), blockedBy([]Index{geom.Pt(1, 1)})))
	})
	t.Run("corner between two blockers is opaque", func(t *testing.T) {
		corner := []Index{geom.Pt(1, 0), geom.Pt(0, 1)}

		assert.False(t, origin.HasLineOfSight(geom.Pt(1, 1), blockedBy(corner)))
		assert.False(t, origin.HasLineOfSight(geom.Pt(2, 2), blockedBy(corner)))
	})
	t.Run("corner beside a single blocker stays open", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), blockedBy([]Index{geom.Pt(1, 0)})))
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), blockedBy([]Index{geom.Pt(0, 1)})))
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 2), blockedBy([]Index{geom.Pt(2, 1)})))
	})
	t.Run("reciprocal", func(t *testing.T) {
		// a Bresenham walk traces a different chain of cells when reversed, so these
		// pairs disagreed until the endpoints were canonicalized
		blocking := []Index{geom.Pt(1, 1)}
		for _, target := range []Index{geom.Pt(1, 2), geom.Pt(2, 1), geom.Pt(2, 4), geom.Pt(4, 2)} {
			t.Run(target.String(), func(t *testing.T) {
				forward := origin.HasLineOfSight(target, blockedBy(blocking))
				backward := Sq(target.XY()).HasLineOfSight(origin.Point(), blockedBy(blocking))

				assert.Equal(t, forward, backward)
			})
		}
	})
	t.Run("reciprocal across a box", func(t *testing.T) {
		blocking := []Index{geom.Pt(1, 1), geom.Pt(-2, 1), geom.Pt(0, -2)}
		for x := -4; x <= 4; x++ {
			for y := -4; y <= 4; y++ {
				a, b := Sq(0, 0), geom.Pt(x, y)

				assert.Equal(t, a.HasLineOfSight(b, blockedBy(blocking)), Sq(x, y).HasLineOfSight(a.Point(), blockedBy(blocking)))
			}
		}
	})
	t.Run("diagonal wall hides everything behind it", func(t *testing.T) {
		var wall []Index
		for i := -3; i <= 3; i++ {
			wall = append(wall, geom.Pt(2+i, 2-i))
		}

		for x := 3; x <= 6; x++ {
			for y := 3; y <= 6; y++ {
				assert.False(t, origin.HasLineOfSight(geom.Pt(x, y), blockedBy(wall)))
			}
		}
	})
}

func BenchmarkSquare_HasLineOfSight(b *testing.B) {
	blocking := []Index{geom.Pt(10, 5), geom.Pt(20, 12), geom.Pt(-4, 3)}

	for b.Loop() {
		sinkBool = Sq(0, 0).HasLineOfSight(geom.Pt(30, 17), blockedBy(blocking))
	}
}

func FuzzSquare_HasLineOfSight(f *testing.F) {
	f.Add(int8(0), int8(0), int8(4), int8(2), uint8(3))
	f.Add(int8(-3), int8(2), int8(5), int8(-6), uint8(17))
	f.Add(int8(1), int8(1), int8(1), int8(1), uint8(0))

	f.Fuzz(func(t *testing.T, x0, y0, x1, y1 int8, seed uint8) {
		from, to := Sq(int(x0)%12, int(y0)%12), geom.Pt(int(x1)%12, int(y1)%12)
		blocked := func(index Index) bool {
			return (index.X*7+index.Y*13+int(seed))%5 == 0
		}
		seen := from.HasLineOfSight(to, blocked)

		assert.Equal(t, Square(to).HasLineOfSight(from.Point(), blocked), seen)
		assert.True(t, from.HasLineOfSight(to, nil))
	})
}

func TestSquare_FieldOfView(t *testing.T) {
	center := Sq(5, 5)
	blocked := func(index Index) bool {
		return index == geom.Pt(5, 6)
	}

	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("a nil blocked is Spiral", func(t *testing.T) {
				assert.Equal(t, center.FieldOfView(3, movement, nil), center.Spiral(3, movement))
			})
			t.Run("a wall is in sight and hides what is behind it, in the order of Spiral", func(t *testing.T) {
				seen := center.FieldOfView(3, movement, blocked)

				assert.True(t, slices.Contains(seen, geom.Pt(5, 6)))
				assert.False(t, slices.Contains(seen, geom.Pt(5, 7)))
				assert.Equal(t, seen[0], center.Point())
			})
			t.Run("is exactly the indices of Spiral HasLineOfSight sees", func(t *testing.T) {
				every := func(index Index) bool {
					return (index.X*7+index.Y*13)%5 == 2
				}

				var expected []Index
				for _, index := range center.Spiral(4, movement) {
					if center.HasLineOfSight(index, every) {
						expected = append(expected, index)
					}
				}

				assert.Equal(t, center.FieldOfView(4, movement, every), expected)
			})
			t.Run("asks about every index at most once and never about s", func(t *testing.T) {
				asked := map[Index]int{}
				center.FieldOfView(4, movement, func(index Index) bool {
					asked[index]++

					return (index.X*7+index.Y*13)%5 == 2
				})

				assert.Equal(t, asked[center.Point()], 0)

				for index, times := range asked {
					assert.Equal(t, times, 1, index.String()+": ")
				}
			})
			t.Run("a negative radius is nil", func(t *testing.T) {
				assert.Nil(t, center.FieldOfView(-1, movement, blocked))
			})
			t.Run("allocates once more for what it asked where something can block", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = center.FieldOfView(3, movement, blocked)
				}), 2.0)
			})
			t.Run("allocates once, at its length, where nothing blocks", func(t *testing.T) {
				assert.Equal(t, cap(center.FieldOfView(3, movement, nil)), RangeLen(3, movement))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = center.FieldOfView(3, movement, nil)
				}), 1.0)
			})
		})
	}
}

func TestSquare_FieldOfViewSeq(t *testing.T) {
	blocked := func(index Index) bool {
		return (index.X*7+index.Y*13)%5 == 2
	}

	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what FieldOfView returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(Sq(5, 5).FieldOfViewSeq(4, movement, blocked)), Sq(5, 5).FieldOfView(4, movement, blocked))
			})
			t.Run("asks about nothing past where it is stopped", func(t *testing.T) {
				asked := 0
				for range Sq(5, 5).FieldOfViewSeq(40, movement, func(Index) bool {
					asked++

					return false
				}) {
					break
				}

				assert.Equal(t, asked, 0)
			})
		})
	}
}

func TestSquare_Point(t *testing.T) {
	assert.Equal(t, Sq(3, -2).Point(), geom.Pt(3, -2))
}

func TestSquare_String(t *testing.T) {
	assert.Equal(t, Sq(3, -2).String(), "(3,-2)")
	assert.Equal(t, fmt.Sprint(Sq(3, -2)), "(3,-2)")
}

func ExampleSquare() {
	s := Sq(5, 5)

	fmt.Println(s.Neighbor(geom.DirectionUp), s.DistanceTo(geom.Pt(8, 1), MovementCardinal))
	fmt.Println(s.Ring(1, MovementCardinal))
	fmt.Println(s.LineTo(geom.Pt(8, 6), MovementDiagonal))
	fmt.Println(s.HasLineOfSight(geom.Pt(8, 5), blockedBy([]Index{geom.Pt(6, 5)})))
	fmt.Println(s.Turn(1), s.ReflectX())
	// Output:
	// (5,4) 7
	// [(6,5) (5,6) (4,5) (5,4)]
	// [(5,5) (6,5) (7,6) (8,6)]
	// false
	// (-5,5) (5,-5)
}
