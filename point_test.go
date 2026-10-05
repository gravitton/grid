package grid

import (
	"slices"
	"strconv"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"
)

var testPoint = Pt(3, -2)

func TestPoint_Constructor(t *testing.T) {
	assert.Equal(t, Pt(3, -2).Point(), geom.Pt(3, -2))
}

func TestRangeLen(t *testing.T) {
	t.Run("is the length of Range", func(t *testing.T) {
		for _, movement := range Movements() {
			for n := -1; n < 6; n++ {
				assert.Equal(t, RangeLen(n, movement), len(testPoint.Range(n, movement)), movement.String()+" at "+strconv.Itoa(n)+": ")
			}
		}
	})
	t.Run("panics for a movement outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			RangeLen(1, MovementNone)
		}, "grid: unknown movement -1")
	})
}

func TestRingLen(t *testing.T) {
	t.Run("is the length of Ring", func(t *testing.T) {
		for _, movement := range Movements() {
			for n := -1; n < 6; n++ {
				assert.Equal(t, RingLen(n, movement), len(testPoint.Ring(n, movement)), movement.String()+" at "+strconv.Itoa(n)+": ")
			}
		}
	})
	t.Run("panics for a movement outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			RingLen(1, MovementNone)
		}, "grid: unknown movement -1")
	})
}

func TestPoint_Turn(t *testing.T) {
	t.Run("a quarter turn is clockwise on a screen", func(t *testing.T) {
		assert.Equal(t, Pt(1, 0).Turn(1), Pt(0, 1))
		assert.Equal(t, Pt(3, -2).Turn(1), Pt(2, 3))
	})
	t.Run("negative steps turn the other way", func(t *testing.T) {
		assert.Equal(t, Pt(1, 0).Turn(-1), Pt(0, -1))
		assert.Equal(t, testPoint.Turn(-1), testPoint.Turn(3))
	})
	t.Run("four steps are a full turn", func(t *testing.T) {
		for steps := -8; steps <= 8; steps++ {
			assert.Equal(t, testPoint.Turn(steps).Turn(4-steps), testPoint, strconv.Itoa(steps)+": ")
		}
	})
	t.Run("turns a neighbor as geom turns its direction", func(t *testing.T) {
		for _, direction := range geom.Directions() {
			assert.Equal(t, Point(Pt(0, 0).Neighbor(direction)).Turn(1).Point(), Pt(0, 0).Neighbor(direction.Turn(2)), direction.String()+": ")
		}
	})
	t.Run("keeps the distance from the origin", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.Equal(t, testPoint.Turn(1).DistanceTo(geom.Pt(0, 0), movement), testPoint.DistanceTo(geom.Pt(0, 0), movement), movement.String()+": ")
		}
	})
}

func TestPoint_TurnAround(t *testing.T) {
	center := geom.Pt(1, 1)

	t.Run("turns around the center", func(t *testing.T) {
		assert.Equal(t, Pt(3, 1).TurnAround(center, 1), Pt(1, 3))
		assert.Equal(t, Pt(3, 1).TurnAround(center, 2), Pt(-1, 1))
	})
	t.Run("the center stays", func(t *testing.T) {
		assert.Equal(t, Point(center).TurnAround(center, 3), Point(center))
	})
	t.Run("around the origin is Turn", func(t *testing.T) {
		assert.Equal(t, testPoint.TurnAround(geom.Pt(0, 0), 3), testPoint.Turn(3))
	})
}

func TestPoint_ReflectX(t *testing.T) {
	t.Run("negates the row", func(t *testing.T) {
		assert.Equal(t, Pt(3, -2).ReflectX(), Pt(3, 2))
	})
	t.Run("twice is the point", func(t *testing.T) {
		assert.Equal(t, testPoint.ReflectX().ReflectX(), testPoint)
	})
}

func TestPoint_ReflectY(t *testing.T) {
	t.Run("negates the column", func(t *testing.T) {
		assert.Equal(t, Pt(3, -2).ReflectY(), Pt(-3, -2))
	})
	t.Run("both reflections are a half turn", func(t *testing.T) {
		assert.Equal(t, testPoint.ReflectX().ReflectY(), testPoint.Turn(2))
	})
}

func TestPoint_Neighbor(t *testing.T) {
	t.Run("steps once in the direction", func(t *testing.T) {
		assert.Equal(t, testPoint.Neighbor(geom.DirectionRight), geom.Pt(4, -2))
		assert.Equal(t, testPoint.Neighbor(geom.DirectionUpLeft), geom.Pt(2, -3))
	})
	t.Run("none stays", func(t *testing.T) {
		assert.Equal(t, testPoint.Neighbor(geom.DirectionNone), testPoint.Point())
	})
}

func TestPoint_Neighbors(t *testing.T) {
	t.Run("cardinal is the four edge neighbors by increasing angle", func(t *testing.T) {
		assert.Equal(t, Pt(3, -2).Neighbors(MovementCardinal), []geom.Point[int]{geom.Pt(4, -2), geom.Pt(3, -1), geom.Pt(2, -2), geom.Pt(3, -3)})
	})
	t.Run("diagonal is all eight", func(t *testing.T) {
		assert.Length(t, Pt(3, -2).Neighbors(MovementDiagonal), 8)
	})
}

func TestPoint_AppendNeighbors(t *testing.T) {
	for _, movement := range Movements() {
		assert.Equal(t, testPoint.AppendNeighbors([]geom.Point[int]{testPoint.Point()}, movement), append([]geom.Point[int]{testPoint.Point()}, testPoint.Neighbors(movement)...), movement.String()+": ")
	}
}

func TestPoint_Range(t *testing.T) {
	center := Pt(3, -2)

	t.Run("a negative radius is nil", func(t *testing.T) {
		assert.Equal(t, center.Range(-1, MovementCardinal), nil)
	})
	t.Run("a zero radius is the point alone", func(t *testing.T) {
		assert.Equal(t, center.Range(0, MovementDiagonal), []geom.Point[int]{geom.Pt(3, -2)})
	})
	t.Run("cardinal is the diamond of Manhattan distance", func(t *testing.T) {
		assert.Equal(t, center.Range(1, MovementCardinal), []geom.Point[int]{geom.Pt(2, -2), geom.Pt(3, -3), geom.Pt(3, -2), geom.Pt(3, -1), geom.Pt(4, -2)})
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

			assert.True(t, slices.IsSortedFunc(indices, geom.Point[int].Compare), movement.String()+": ")
			assert.Equal(t, cap(indices), len(indices), movement.String()+": ")
			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkPoints = center.Range(4, movement)
			}), 1.0, movement.String()+": ")
		}
	})
	t.Run("panics for a movement outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			center.Range(1, MovementNone)
		}, "grid: unknown movement -1")
	})
}

func TestPoint_AppendRange(t *testing.T) {
	buffer := make([]geom.Point[int], 0, 64)

	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("appends what Range returns", func(t *testing.T) {
				assert.Equal(t, testPoint.AppendRange([]geom.Point[int]{testPoint.Point()}, 2, movement), append([]geom.Point[int]{testPoint.Point()}, testPoint.Range(2, movement)...))
			})
			t.Run("a negative radius appends nothing", func(t *testing.T) {
				assert.Length(t, testPoint.AppendRange(buffer[:0], -1, movement), 0)
			})
			t.Run("allocates nothing with room", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testPoint.AppendRange(buffer[:0], 3, movement)
				}), 0.0)
			})
		})
	}
}

func TestPoint_RangeSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Range returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testPoint.RangeSeq(3, movement)), testPoint.Range(3, movement))
			})
			t.Run("stops at the break", func(t *testing.T) {
				count := 0
				for range testPoint.RangeSeq(3, movement) {
					count++

					break
				}

				assert.Equal(t, count, 1)
			})
			t.Run("allocates nothing", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					for index := range testPoint.RangeSeq(3, movement) {
						sinkPoint = index
					}
				}), 0.0)
			})
		})
	}
}

func TestPoint_Ring(t *testing.T) {
	t.Run("a negative radius is nil", func(t *testing.T) {
		assert.Equal(t, testPoint.Ring(-1, MovementCardinal), nil)
	})
	t.Run("a zero radius is the point alone", func(t *testing.T) {
		assert.Equal(t, testPoint.Ring(0, MovementDiagonal), []geom.Point[int]{testPoint.Point()})
	})
	t.Run("cardinal walks the diamond by increasing angle", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).Ring(2, MovementCardinal), []geom.Point[int]{
			geom.Pt(2, 0), geom.Pt(1, 1), geom.Pt(0, 2), geom.Pt(-1, 1), geom.Pt(-2, 0), geom.Pt(-1, -1), geom.Pt(0, -2), geom.Pt(1, -1),
		})
	})
	t.Run("diagonal walks the square by increasing angle", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).Ring(1, MovementDiagonal), []geom.Point[int]{
			geom.Pt(1, 0), geom.Pt(1, 1), geom.Pt(0, 1), geom.Pt(-1, 1), geom.Pt(-1, 0), geom.Pt(-1, -1), geom.Pt(0, -1), geom.Pt(1, -1),
		})
	})
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("a radius of one is the neighbors", func(t *testing.T) {
				assert.Equal(t, testPoint.Ring(1, movement), testPoint.Neighbors(movement))
			})
			t.Run("holds exactly the indices at the distance", func(t *testing.T) {
				for n := range 6 {
					ring := testPoint.Ring(n, movement)

					assert.Equal(t, Reg(ring).Len(), len(ring), strconv.Itoa(n)+": ")
					for _, index := range testPoint.Range(n+1, movement) {
						assert.Equal(t, slices.Contains(ring, index), testPoint.DistanceTo(index, movement) == n, strconv.Itoa(n)+" at "+index.String()+": ")
					}
				}
			})
			t.Run("allocates once at its length", func(t *testing.T) {
				ring := testPoint.Ring(4, movement)

				assert.Equal(t, cap(ring), len(ring))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testPoint.Ring(4, movement)
				}), 1.0)
			})
		})
	}
	t.Run("panics for a movement outside the constants", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			testPoint.Ring(-1, MovementNone)
		}, "grid: unknown movement -1")
	})
}

func TestPoint_AppendRing(t *testing.T) {
	for _, movement := range Movements() {
		assert.Equal(t, testPoint.AppendRing([]geom.Point[int]{testPoint.Point()}, 3, movement), append([]geom.Point[int]{testPoint.Point()}, testPoint.Ring(3, movement)...), movement.String()+": ")
	}
}

func TestPoint_RingSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Ring returns", func(t *testing.T) {
				for n := -1; n < 4; n++ {
					assert.Equal(t, slices.Collect(testPoint.RingSeq(n, movement)), testPoint.Ring(n, movement), strconv.Itoa(n)+": ")
				}
			})
			t.Run("stops at the break", func(t *testing.T) {
				for n := range 3 {
					count := 0
					for range testPoint.RingSeq(n, movement) {
						count++

						break
					}

					assert.Equal(t, count, 1, strconv.Itoa(n)+": ")
				}
			})
			t.Run("allocates nothing", func(t *testing.T) {
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					for index := range testPoint.RingSeq(3, movement) {
						sinkPoint = index
					}
				}), 0.0)
			})
		})
	}
}

func TestPoint_Spiral(t *testing.T) {
	t.Run("a negative radius is nil", func(t *testing.T) {
		assert.Equal(t, testPoint.Spiral(-1, MovementCardinal), nil)
	})
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("starts at the point", func(t *testing.T) {
				assert.Equal(t, testPoint.Spiral(3, movement)[0], testPoint.Point())
			})
			t.Run("holds what Range holds", func(t *testing.T) {
				assert.True(t, Reg(testPoint.Spiral(4, movement)).Equal(Reg(testPoint.Range(4, movement))))
			})
			t.Run("never steps back toward the point", func(t *testing.T) {
				previous := 0
				for _, index := range testPoint.Spiral(4, movement) {
					distance := testPoint.DistanceTo(index, movement)
					assert.GreaterOrEqual(t, distance, previous, index.String()+": ")
					previous = distance
				}
			})
			t.Run("allocates once at its length", func(t *testing.T) {
				spiral := testPoint.Spiral(4, movement)

				assert.Equal(t, cap(spiral), len(spiral))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testPoint.Spiral(4, movement)
				}), 1.0)
			})
		})
	}
}

func TestPoint_AppendSpiral(t *testing.T) {
	for _, movement := range Movements() {
		assert.Equal(t, testPoint.AppendSpiral([]geom.Point[int]{testPoint.Point()}, 3, movement), append([]geom.Point[int]{testPoint.Point()}, testPoint.Spiral(3, movement)...), movement.String()+": ")
	}
}

func TestPoint_SpiralSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Spiral returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testPoint.SpiralSeq(3, movement)), testPoint.Spiral(3, movement))
			})
			t.Run("stops at the break, inside a ring too", func(t *testing.T) {
				count := 0
				for range testPoint.SpiralSeq(3, movement) {
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

func TestPoint_DistanceTo(t *testing.T) {
	t.Run("cardinal is the Manhattan distance", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).DistanceTo(geom.Pt(3, 4), MovementCardinal), 7)
	})
	t.Run("diagonal is the Chebyshev distance", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).DistanceTo(geom.Pt(3, 4), MovementDiagonal), 4)
	})
}

func TestPoint_DirectionTo(t *testing.T) {
	t.Run("the point itself has no direction", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.Equal(t, testPoint.DirectionTo(testPoint.Point(), movement), geom.DirectionNone, movement.String()+": ")
		}
	})
	t.Run("a neighbor is in its own direction", func(t *testing.T) {
		for _, movement := range Movements() {
			for _, direction := range movement.Directions() {
				assert.Equal(t, testPoint.DirectionTo(testPoint.Neighbor(direction), movement), direction, movement.String()+" "+direction.String()+": ")
			}
		}
	})
	t.Run("diagonal is the nearest of the eight", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(5, 1), MovementDiagonal), geom.DirectionRight)
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(4, 3), MovementDiagonal), geom.DirectionDownRight)
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(-1, -6), MovementDiagonal), geom.DirectionUp)
	})
	t.Run("cardinal is the nearest of the four", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(4, 3), MovementCardinal), geom.DirectionRight)
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(-4, 3), MovementCardinal), geom.DirectionLeft)
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(2, 3), MovementCardinal), geom.DirectionDown)
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(2, -3), MovementCardinal), geom.DirectionUp)
	})
	t.Run("cardinal answers the horizontal on a diagonal", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(3, 3), MovementCardinal), geom.DirectionRight)
		assert.Equal(t, Pt(0, 0).DirectionTo(geom.Pt(-3, 3), MovementCardinal), geom.DirectionLeft)
	})
}

func TestPoint_Line(t *testing.T) {
	t.Run("to the point itself is the point alone", func(t *testing.T) {
		for _, movement := range Movements() {
			assert.Equal(t, testPoint.Line(testPoint.Point(), movement), []geom.Point[int]{testPoint.Point()}, movement.String()+": ")
		}
	})
	t.Run("diagonal steps across corners", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).Line(geom.Pt(3, 3), MovementDiagonal), []geom.Point[int]{geom.Pt(0, 0), geom.Pt(1, 1), geom.Pt(2, 2), geom.Pt(3, 3)})
	})
	t.Run("cardinal steps along one axis at a time", func(t *testing.T) {
		assert.Equal(t, Pt(0, 0).Line(geom.Pt(2, 1), MovementCardinal), []geom.Point[int]{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(1, 1), geom.Pt(2, 1)})
	})
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("runs from the point to the target by neighbors", func(t *testing.T) {
				for _, target := range testPoint.Range(5, MovementDiagonal) {
					line := testPoint.Line(target, movement)

					assert.Equal(t, line[0], testPoint.Point(), target.String()+": ")
					assert.Equal(t, line[len(line)-1], target, target.String()+": ")
					assert.Length(t, line, testPoint.DistanceTo(target, movement)+1, target.String()+": ")
					for i := 1; i < len(line); i++ {
						assert.Equal(t, Point(line[i-1]).DistanceTo(line[i], movement), 1, target.String()+": ")
					}
				}
			})
			t.Run("allocates once at its length", func(t *testing.T) {
				line := testPoint.Line(geom.Pt(-4, 7), movement)

				assert.Equal(t, cap(line), len(line))
				assert.Equal(t, testing.AllocsPerRun(10, func() {
					sinkPoints = testPoint.Line(geom.Pt(-4, 7), movement)
				}), 1.0)
			})
		})
	}
}

func TestPoint_AppendLine(t *testing.T) {
	for _, movement := range Movements() {
		assert.Equal(t, testPoint.AppendLine([]geom.Point[int]{testPoint.Point()}, geom.Pt(-4, 7), movement), append([]geom.Point[int]{testPoint.Point()}, testPoint.Line(geom.Pt(-4, 7), movement)...), movement.String()+": ")
	}
}

func TestPoint_LineSeq(t *testing.T) {
	for _, movement := range Movements() {
		t.Run(movement.String(), func(t *testing.T) {
			t.Run("yields what Line returns", func(t *testing.T) {
				assert.Equal(t, slices.Collect(testPoint.LineSeq(geom.Pt(-4, 7), movement)), testPoint.Line(geom.Pt(-4, 7), movement))
			})
			t.Run("stops at the break, at the first index and after it", func(t *testing.T) {
				for stop := 1; stop <= 2; stop++ {
					count := 0
					for range testPoint.LineSeq(geom.Pt(-4, 7), movement) {
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
					for index := range testPoint.LineSeq(geom.Pt(-4, 7), movement) {
						sinkPoint = index
					}
				}), 0.0)
			})
		})
	}
}

func TestPoint_HasLineOfSight(t *testing.T) {
	origin := Pt(0, 0)

	t.Run("clear", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), nil))
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{}))
		assert.True(t, origin.HasLineOfSight(origin.Point(), nil))
	})
	t.Run("blocked mid", func(t *testing.T) {
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{geom.Pt(1, 0)}))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{geom.Pt(2, 0)}))
	})
	t.Run("blocker beyond target", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 0), []geom.Point[int]{geom.Pt(3, 0)}))
	})
	t.Run("target in blocking list", func(t *testing.T) {
		// can see into, not through
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 0), []geom.Point[int]{geom.Pt(3, 0)}))
	})
	t.Run("diagonal", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(3, 3), nil))
		assert.False(t, origin.HasLineOfSight(geom.Pt(3, 3), []geom.Point[int]{geom.Pt(1, 1)}))
	})
	t.Run("corner between two blockers is opaque", func(t *testing.T) {
		corner := []geom.Point[int]{geom.Pt(1, 0), geom.Pt(0, 1)}

		assert.False(t, origin.HasLineOfSight(geom.Pt(1, 1), corner))
		assert.False(t, origin.HasLineOfSight(geom.Pt(2, 2), corner))
	})
	t.Run("corner beside a single blocker stays open", func(t *testing.T) {
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), []geom.Point[int]{geom.Pt(1, 0)}))
		assert.True(t, origin.HasLineOfSight(geom.Pt(1, 1), []geom.Point[int]{geom.Pt(0, 1)}))
		assert.True(t, origin.HasLineOfSight(geom.Pt(2, 2), []geom.Point[int]{geom.Pt(2, 1)}))
	})
	t.Run("reciprocal", func(t *testing.T) {
		// a Bresenham walk traces a different chain of cells when reversed, so these
		// pairs disagreed until the endpoints were canonicalized
		blocking := []geom.Point[int]{geom.Pt(1, 1)}
		for _, target := range []geom.Point[int]{geom.Pt(1, 2), geom.Pt(2, 1), geom.Pt(2, 4), geom.Pt(4, 2)} {
			t.Run(target.String(), func(t *testing.T) {
				forward := origin.HasLineOfSight(target, blocking)
				backward := Pt(target.XY()).HasLineOfSight(origin.Point(), blocking)

				assert.Equal(t, forward, backward)
			})
		}
	})
	t.Run("reciprocal across a box", func(t *testing.T) {
		blocking := []geom.Point[int]{geom.Pt(1, 1), geom.Pt(-2, 1), geom.Pt(0, -2)}
		for x := -4; x <= 4; x++ {
			for y := -4; y <= 4; y++ {
				a, b := Pt(0, 0), geom.Pt(x, y)

				assert.Equal(t, a.HasLineOfSight(b, blocking), Pt(x, y).HasLineOfSight(a.Point(), blocking))
			}
		}
	})
	t.Run("diagonal wall hides everything behind it", func(t *testing.T) {
		var wall []geom.Point[int]
		for i := -3; i <= 3; i++ {
			wall = append(wall, geom.Pt(2+i, 2-i))
		}

		for x := 3; x <= 6; x++ {
			for y := 3; y <= 6; y++ {
				assert.False(t, origin.HasLineOfSight(geom.Pt(x, y), wall))
			}
		}
	})
}

func TestPoint_HasLineOfSightFunc(t *testing.T) {
	blocking := []geom.Point[int]{geom.Pt(1, 1), geom.Pt(-2, 1), geom.Pt(0, -2), geom.Pt(3, 0), geom.Pt(2, -1)}
	blocked := func(index geom.Point[int]) bool {
		return slices.Contains(blocking, index)
	}

	t.Run("a nil blocked blocks nothing", func(t *testing.T) {
		assert.True(t, Pt(0, 0).HasLineOfSightFunc(geom.Pt(5, 5), nil))
	})
	t.Run("agrees with the slice form", func(t *testing.T) {
		for _, target := range Pt(0, 0).Range(6, MovementDiagonal) {
			assert.Equal(t, Pt(0, 0).HasLineOfSightFunc(target, blocked), Pt(0, 0).HasLineOfSight(target, blocking), target.String()+": ")
		}
	})
	t.Run("reads the same from either end", func(t *testing.T) {
		for _, target := range Pt(0, 0).Range(6, MovementDiagonal) {
			assert.Equal(t, Pt(0, 0).HasLineOfSightFunc(target, blocked), Point(target).HasLineOfSightFunc(geom.Pt(0, 0), blocked), target.String()+": ")
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkBool = Pt(0, 0).HasLineOfSightFunc(geom.Pt(6, 4), blocked)
		}), 0.0)
	})
}

// testFieldOfView runs the sections the forms of the field of view share, each form given as
// the slice it produces from the candidates and the blocking indices.
func testFieldOfView(t *testing.T, field func(p Point, candidates, blocking []geom.Point[int]) []geom.Point[int]) {
	center := Pt(0, 0)
	candidates := center.Range(4, MovementDiagonal)
	blocking := []geom.Point[int]{geom.Pt(1, 0), geom.Pt(1, 1), geom.Pt(-2, 1), geom.Pt(0, -2)}

	t.Run("is the candidates with line of sight, in the order given", func(t *testing.T) {
		var expected []geom.Point[int]
		for _, candidate := range candidates {
			if center.Point().ChebyshevDistanceTo(candidate) <= 1 || center.HasLineOfSight(candidate, blocking) {
				expected = append(expected, candidate)
			}
		}

		assert.Equal(t, field(center, candidates, blocking), expected)
	})
	t.Run("nothing blocking sees every candidate", func(t *testing.T) {
		assert.Equal(t, field(center, candidates, nil), candidates)
	})
	t.Run("no candidates see nothing", func(t *testing.T) {
		assert.Length(t, field(center, nil, blocking), 0)
	})
}

func TestPoint_FieldOfView(t *testing.T) {
	center := Pt(0, 0)
	candidates := center.Range(3, MovementDiagonal)

	t.Run("no blocking", func(t *testing.T) {
		assert.Equal(t, len(center.FieldOfView(candidates, nil)), len(candidates))
		assert.Equal(t, len(center.FieldOfView(candidates, []geom.Point[int]{})), len(candidates))
	})
	t.Run("empty candidates", func(t *testing.T) {
		assert.Equal(t, len(center.FieldOfView(nil, nil)), 0)
		assert.Equal(t, len(center.FieldOfView([]geom.Point[int]{}, nil)), 0)
	})
	t.Run("adjacent always visible", func(t *testing.T) {
		allNeighbors := []geom.Point[int]{
			geom.Pt(1, 0), geom.Pt(0, -1), geom.Pt(-1, 0), geom.Pt(0, 1),
			geom.Pt(1, -1), geom.Pt(-1, -1), geom.Pt(-1, 1), geom.Pt(1, 1),
		}
		visible := center.FieldOfView(allNeighbors, allNeighbors)
		assert.Equal(t, len(visible), len(allNeighbors))
	})
	t.Run("gap in blocking ring", func(t *testing.T) {
		// only (1,0) is unblocked, so (2,0) and (3,0) should be visible; (0,2) is not
		blocking := []geom.Point[int]{
			geom.Pt(0, -1), geom.Pt(-1, -1), geom.Pt(-1, 0), geom.Pt(-1, 1),
			geom.Pt(0, 1), geom.Pt(1, -1), geom.Pt(1, 1),
		}
		visible := center.FieldOfView([]geom.Point[int]{geom.Pt(2, 0), geom.Pt(3, 0), geom.Pt(0, 2)}, blocking)
		assert.Equal(t, len(visible), 2)
		assert.Equal(t, visible[0], geom.Pt(2, 0))
		assert.Equal(t, visible[1], geom.Pt(3, 0))
	})
}

func TestPoint_AppendFieldOfView(t *testing.T) {
	testFieldOfView(t, func(p Point, candidates, blocking []geom.Point[int]) []geom.Point[int] {
		return p.AppendFieldOfView([]geom.Point[int]{p.Point()}, candidates, blocking)[1:]
	})
}

func TestPoint_FieldOfViewSeq(t *testing.T) {
	testFieldOfView(t, func(p Point, candidates, blocking []geom.Point[int]) []geom.Point[int] {
		return slices.Collect(p.FieldOfViewSeq(candidates, blocking))
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range Pt(0, 0).FieldOfViewSeq(Pt(0, 0).Range(2, MovementDiagonal), []geom.Point[int]{geom.Pt(5, 5)}) {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
}

func TestPoint_FieldOfViewFunc(t *testing.T) {
	testFieldOfView(t, func(p Point, candidates, blocking []geom.Point[int]) []geom.Point[int] {
		return p.FieldOfViewFunc(candidates, func(index geom.Point[int]) bool {
			return slices.Contains(blocking, index)
		})
	})
	t.Run("a nil blocked sees every candidate", func(t *testing.T) {
		candidates := Pt(0, 0).Range(3, MovementDiagonal)

		assert.Equal(t, Pt(0, 0).FieldOfViewFunc(candidates, nil), candidates)
	})
}

func TestPoint_AppendFieldOfViewFunc(t *testing.T) {
	testFieldOfView(t, func(p Point, candidates, blocking []geom.Point[int]) []geom.Point[int] {
		return p.AppendFieldOfViewFunc([]geom.Point[int]{p.Point()}, candidates, func(index geom.Point[int]) bool {
			return slices.Contains(blocking, index)
		})[1:]
	})
}

func TestPoint_FieldOfViewFuncSeq(t *testing.T) {
	testFieldOfView(t, func(p Point, candidates, blocking []geom.Point[int]) []geom.Point[int] {
		return slices.Collect(p.FieldOfViewFuncSeq(candidates, func(index geom.Point[int]) bool {
			return slices.Contains(blocking, index)
		}))
	})
}

func TestPoint_Point(t *testing.T) {
	assert.Equal(t, Pt(3, -2).Point(), geom.Pt(3, -2))
}
