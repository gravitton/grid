package grid

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/gravitton/assert"
)

// tile is an element ordered by x and then by y.
type tile struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c tile) Compare(other tile) int {
	if order := cmp.Compare(c.X, other.X); order != 0 {
		return order
	}

	return cmp.Compare(c.Y, other.Y)
}

func (c tile) String() string {
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

var (
	testSet      = setOfTiles(square(tile{0, 0}, 2)...)
	testSetOther = setOfTiles(square(tile{3, -1}, 2)...)
	testSetEmpty = Set[tile]{}

	sinkSet   Set[tile]
	sinkTiles []tile
)

// setOfTiles returns the set of the given tiles.
func setOfTiles(tiles ...tile) Set[tile] {
	return Reg(tiles)
}

// square returns the tiles within n columns and rows of the center, ordered by x and then by y.
func square(center tile, n int) []tile {
	tiles := make([]tile, 0, (2*n+1)*(2*n+1))
	for dx := -n; dx <= n; dx++ {
		for dy := -n; dy <= n; dy++ {
			tiles = append(tiles, tile{center.X + dx, center.Y + dy})
		}
	}

	return tiles
}

// referenceSet returns the elements a map-based set keeps of the two sets, sorted.
func referenceSet(a, b Set[tile], keep func(inA, inB bool) bool) []tile {
	seen := map[tile]bool{}
	for _, s := range []Set[tile]{a, b} {
		for elem := range s.All() {
			seen[elem] = true
		}
	}

	var kept []tile
	for elem := range seen {
		if keep(a.Contains(elem), b.Contains(elem)) {
			kept = append(kept, elem)
		}
	}
	slices.SortFunc(kept, tile.Compare)

	return kept
}

// testSetPairs returns the pairs of sets every operation is checked on.
func testSetPairs() [][2]Set[tile] {
	return [][2]Set[tile]{
		{testSet, testSetOther},
		{testSetOther, testSet},
		{testSet, testSet},
		{testSet, testSetEmpty},
		{testSetEmpty, testSet},
		{testSetEmpty, testSetEmpty},
		{testSet, setOfTiles(tile{9, 9})},
		{testSet, setOfTiles(square(tile{0, 0}, 1)...)},
	}
}

// testSetOperation runs the sections every set operation shares, checking it against the
// elements a map keeps of the two sets.
func testSetOperation(t *testing.T, operation func(a, b Set[tile]) Set[tile], keep func(inA, inB bool) bool) {
	t.Run("keeps what a map keeps", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			assert.Equal(t, slices.Collect(operation(pair[0], pair[1]).All()), referenceSet(pair[0], pair[1], keep), fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
	t.Run("leaves both sets alone", func(t *testing.T) {
		before, beforeOther := testSet.Elements(), testSetOther.Elements()
		sinkSet = operation(testSet, testSetOther)

		assert.Equal(t, testSet.Elements(), before)
		assert.Equal(t, testSetOther.Elements(), beforeOther)
	})
	t.Run("allocates once, and nothing where the result is empty or equals a side", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			allocs := 1.0
			if result := operation(pair[0], pair[1]); result.IsEmpty() || result.Equal(pair[0]) || result.Equal(pair[1]) {
				allocs = 0
			}

			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkSet = operation(pair[0], pair[1])
			}), allocs, fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
}

// fuzzSetOperation checks a set operation on two sets of fuzzed tiles against the elements a
// map keeps of them.
func fuzzSetOperation(f *testing.F, operation func(a, b Set[tile]) Set[tile], keep func(inA, inB bool) bool) {
	f.Add([]byte{1, 2, 3, 4, 5, 6}, []byte{3, 4, 5, 6, 7, 0})
	f.Add([]byte{}, []byte{0, 0})
	f.Add([]byte{7, 7, 7, 7}, []byte{})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		first, second := fuzzSet(a), fuzzSet(b)

		assert.Equal(t, slices.Collect(operation(first, second).All()), referenceSet(first, second, keep), fmt.Sprint(first.Elements(), " and ", second.Elements(), ": "))
	})
}

// fuzzSet returns the set of the tiles the bytes give in pairs, within four of the origin, so
// two fuzzed sets often share some.
func fuzzSet(data []byte) Set[tile] {
	tiles := make([]tile, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		tiles = append(tiles, tile{int(data[i]%8) - 4, int(data[i+1]%8) - 4})
	}

	return Reg(tiles)
}

func TestSet_Elements(t *testing.T) {
	t.Run("is the elements All yields, in order", func(t *testing.T) {
		assert.Equal(t, testSet.Elements(), slices.Collect(testSet.All()))
	})
	t.Run("is a fresh slice", func(t *testing.T) {
		elems := testSet.Elements()
		elems[0] = tile{9, 9}

		assert.False(t, testSet.Contains(tile{9, 9}))
	})
	t.Run("the empty set has none", func(t *testing.T) {
		assert.Equal(t, testSetEmpty.Elements(), nil)
	})
	t.Run("allocates once at its length", func(t *testing.T) {
		elems := testSet.Elements()

		assert.Equal(t, cap(elems), len(elems))
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkTiles = testSet.Elements()
		}), 1.0)
	})
}

func TestSet_AppendElements(t *testing.T) {
	t.Run("appends what Elements returns", func(t *testing.T) {
		assert.Equal(t, testSet.AppendElements([]tile{{9, 9}}), append([]tile{{9, 9}}, testSet.Elements()...))
	})
	t.Run("the empty set appends nothing", func(t *testing.T) {
		assert.Equal(t, testSetEmpty.AppendElements([]tile{{9, 9}}), []tile{{9, 9}})
	})
	t.Run("allocates nothing with room", func(t *testing.T) {
		buffer := make([]tile, 0, testSet.Len())

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkTiles = testSet.AppendElements(buffer[:0])
		}), 0.0)
	})
}

func TestSet_All(t *testing.T) {
	t.Run("yields the elements in order", func(t *testing.T) {
		assert.True(t, slices.IsSortedFunc(slices.Collect(testSet.All()), tile.Compare))
		assert.Length(t, slices.Collect(testSet.All()), testSet.Len())
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range testSet.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("the empty set yields nothing", func(t *testing.T) {
		assert.Length(t, slices.Collect(testSetEmpty.All()), 0)
	})
}

func TestSet_Backward(t *testing.T) {
	t.Run("yields the elements in reverse order", func(t *testing.T) {
		elems := slices.Collect(testSet.All())
		slices.Reverse(elems)

		assert.Equal(t, slices.Collect(testSet.Backward()), elems)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range testSet.Backward() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("the empty set yields nothing", func(t *testing.T) {
		assert.Length(t, slices.Collect(testSetEmpty.Backward()), 0)
	})
}

func TestSet_Len(t *testing.T) {
	t.Run("counts each element once", func(t *testing.T) {
		assert.Equal(t, testSet.Len(), 25)
		assert.Equal(t, setOfTiles(tile{1, 1}, tile{1, 1}).Len(), 1)
	})
	t.Run("the empty set has none", func(t *testing.T) {
		assert.Equal(t, testSetEmpty.Len(), 0)
	})
}

func TestSet_At(t *testing.T) {
	t.Run("is the element at the place All yields it", func(t *testing.T) {
		i := 0
		for elem := range testSet.All() {
			assert.Equal(t, testSet.At(i), elem)
			i++
		}
	})
	t.Run("panics outside the set", func(t *testing.T) {
		assert.Panics(t, func() {
			testSet.At(testSet.Len())
		})
	})
}

func TestSet_Position(t *testing.T) {
	t.Run("finds every element at its place", func(t *testing.T) {
		for i := range testSet.Len() {
			position, found := testSet.Position(testSet.At(i))

			assert.True(t, found)
			assert.Equal(t, position, i)
		}
	})
	t.Run("an element that is not there has none", func(t *testing.T) {
		for _, elem := range []tile{{9, 9}, {-9, -9}} {
			_, found := testSet.Position(elem)

			assert.False(t, found, elem.String()+": ")
		}
	})
	t.Run("the empty set holds no place", func(t *testing.T) {
		_, found := testSetEmpty.Position(tile{0, 0})

		assert.False(t, found)
	})
}

func TestSet_First(t *testing.T) {
	t.Run("is the first element", func(t *testing.T) {
		elem, found := testSet.First()

		assert.True(t, found)
		assert.Equal(t, elem, tile{-2, -2})
	})
	t.Run("the empty set has none", func(t *testing.T) {
		_, found := testSetEmpty.First()

		assert.False(t, found)
	})
}

func TestSet_Last(t *testing.T) {
	t.Run("is the last element", func(t *testing.T) {
		elem, found := testSet.Last()

		assert.True(t, found)
		assert.Equal(t, elem, tile{2, 2})
	})
	t.Run("the empty set has none", func(t *testing.T) {
		_, found := testSetEmpty.Last()

		assert.False(t, found)
	})
}

func TestSet_Union(t *testing.T) {
	testSetOperation(t, Set[tile].Union, func(inA, inB bool) bool {
		return inA || inB
	})
}

func BenchmarkSet_Union(b *testing.B) {
	for b.Loop() {
		sinkSet = testSet.Union(testSetOther)
	}
}

func FuzzSet_Union(f *testing.F) {
	fuzzSetOperation(f, Set[tile].Union, func(inA, inB bool) bool {
		return inA || inB
	})
}

func TestSet_Intersection(t *testing.T) {
	testSetOperation(t, Set[tile].Intersection, func(inA, inB bool) bool {
		return inA && inB
	})
}

func BenchmarkSet_Intersection(b *testing.B) {
	for b.Loop() {
		sinkSet = testSet.Intersection(testSetOther)
	}
}

func FuzzSet_Intersection(f *testing.F) {
	fuzzSetOperation(f, Set[tile].Intersection, func(inA, inB bool) bool {
		return inA && inB
	})
}

func TestSet_Difference(t *testing.T) {
	testSetOperation(t, Set[tile].Difference, func(inA, inB bool) bool {
		return inA && !inB
	})
}

func FuzzSet_Difference(f *testing.F) {
	fuzzSetOperation(f, Set[tile].Difference, func(inA, inB bool) bool {
		return inA && !inB
	})
}

func TestSet_SymmetricDifference(t *testing.T) {
	testSetOperation(t, Set[tile].SymmetricDifference, func(inA, inB bool) bool {
		return inA != inB
	})
}

func FuzzSet_SymmetricDifference(f *testing.F) {
	fuzzSetOperation(f, Set[tile].SymmetricDifference, func(inA, inB bool) bool {
		return inA != inB
	})
}

func TestSet_Contains(t *testing.T) {
	t.Run("holds its elements", func(t *testing.T) {
		for elem := range testSet.All() {
			assert.True(t, testSet.Contains(elem), elem.String()+": ")
		}
	})
	t.Run("holds nothing else", func(t *testing.T) {
		assert.False(t, testSet.Contains(tile{3, 0}))
		assert.False(t, testSetEmpty.Contains(tile{0, 0}))
	})
}

func BenchmarkSet_Contains(b *testing.B) {
	for b.Loop() {
		sinkBool = testSet.Contains(tile{1, 1})
	}
}

func TestSet_Encloses(t *testing.T) {
	t.Run("agrees with the difference being empty", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			assert.Equal(t, pair[0].Encloses(pair[1]), pair[1].Difference(pair[0]).IsEmpty(), fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkBool = testSet.Encloses(testSetOther)
		}), 0.0)
	})
}

func TestSet_Intersects(t *testing.T) {
	t.Run("agrees with the intersection being empty", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			assert.Equal(t, pair[0].Intersects(pair[1]), !pair[0].Intersection(pair[1]).IsEmpty(), fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkBool = testSet.Intersects(testSetOther)
		}), 0.0)
	})
}

func TestSet_Equal(t *testing.T) {
	t.Run("the same elements in any order given", func(t *testing.T) {
		elems := square(tile{0, 0}, 2)
		slices.Reverse(elems)

		assert.True(t, testSet.Equal(setOfTiles(elems...)))
	})
	t.Run("different elements", func(t *testing.T) {
		assert.False(t, testSet.Equal(testSetOther))
		assert.False(t, testSet.Equal(testSet.Difference(setOfTiles(tile{0, 0}))))
	})
	t.Run("the empty set equals the zero value", func(t *testing.T) {
		assert.True(t, testSet.Difference(testSet).Equal(testSetEmpty))
	})
}

func TestSet_IsEmpty(t *testing.T) {
	t.Run("the zero value and a set of no elements are", func(t *testing.T) {
		assert.True(t, testSetEmpty.IsEmpty())
		assert.True(t, testSet.Difference(testSet).IsEmpty())
	})
	t.Run("a set with an element is not", func(t *testing.T) {
		assert.False(t, setOfTiles(tile{0, 0}).IsEmpty())
	})
}

func TestSet_JSON(t *testing.T) {
	t.Run("round-trips as an array of elements", func(t *testing.T) {
		data, err := json.Marshal(setOfTiles(tile{1, 0}, tile{0, 0}))
		assert.NoError(t, err)
		assert.Equal(t, string(data), `[{"x":0,"y":0},{"x":1,"y":0}]`)

		var decoded Set[tile]
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.True(t, decoded.Equal(setOfTiles(tile{0, 0}, tile{1, 0})))
	})
	t.Run("empty is an empty array", func(t *testing.T) {
		data, err := json.Marshal(testSetEmpty)
		assert.NoError(t, err)
		assert.Equal(t, string(data), "[]")

		decoded := testSet
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.True(t, decoded.IsEmpty())
	})
	t.Run("null is the empty set", func(t *testing.T) {
		decoded := testSet
		assert.NoError(t, json.Unmarshal([]byte("null"), &decoded))
		assert.True(t, decoded.IsEmpty())
	})
	t.Run("orders and drops repeats on the way in", func(t *testing.T) {
		var decoded Set[tile]
		assert.NoError(t, json.Unmarshal([]byte(`[{"x":1,"y":0},{"x":0,"y":0},{"x":1,"y":0}]`), &decoded))
		assert.Equal(t, slices.Collect(decoded.All()), []tile{{0, 0}, {1, 0}})
	})
	t.Run("anything but an array of elements fails", func(t *testing.T) {
		var decoded Set[tile]
		assert.Error(t, json.Unmarshal([]byte(`{"x":1}`), &decoded))
	})
}
