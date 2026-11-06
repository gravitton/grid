package main

import (
	"fmt"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/grid"
)

type Tile struct {
	Wall bool
	Cost float64
}

var (
	ground = Tile{Cost: 1}
	rough  = Tile{Cost: 5}
	wall   = Tile{Wall: true}
)

type builder func(size geom.Size[int], width float64) *grid.Grid[Tile]

func rect(movement grid.Movement) builder {
	return func(size geom.Size[int], width float64) *grid.Grid[Tile] {
		return grid.NewRect[Tile](size, grid.RectCellScale(width), func(options *grid.SquareOptions) {
			options.Movement = movement
		})
	}
}

func isometric(movement grid.Movement) builder {
	return func(size geom.Size[int], width float64) *grid.Grid[Tile] {
		return grid.NewIsometric[Tile](size, grid.IsometricPixelCellScale(width), func(options *grid.SquareOptions) {
			options.Movement = movement
		})
	}
}

func hexFlatTop(size geom.Size[int], width float64) *grid.Grid[Tile] {
	return grid.NewHexFlatTop[Tile](size, grid.HexFlatTopCellScale(width))
}

func hexPointyTop(size geom.Size[int], width float64) *grid.Grid[Tile] {
	return grid.NewHexPointyTop[Tile](size, grid.HexPointyTopCellScale(width))
}

type marks map[rune][]grid.Index

func (m marks) at(r rune) grid.Index {
	if len(m[r]) != 1 {
		panic(fmt.Sprintf("diagrams: %d marks %q on the map", len(m[r]), r))
	}

	return m[r][0]
}

func (m marks) region(r rune) grid.Region {
	return grid.Reg(m[r]...)
}

func paint(build builder, width float64, rows ...string) (*grid.Grid[Tile], marks) {
	g := build(geom.Sz(len(rows[0]), len(rows)), width)
	g.Fill(ground)
	found := marks{}
	for y, row := range rows {
		if len(row) != len(rows[0]) {
			panic(fmt.Sprintf("diagrams: row %d of the map is %d long, not %d", y, len(row), len(rows[0])))
		}
		for x, r := range row {
			index := geom.Pt(x, y)
			switch r {
			case '#':
				g.Cell(index).Set(wall)
			case '~':
				g.Cell(index).Set(rough)
			case '.':
			default:
				found[r] = append(found[r], index)
			}
		}
	}

	return g, found
}

func walkable(cell grid.Cell[Tile]) bool {
	return !cell.Value().Wall
}

func opaque(cell grid.Cell[Tile]) bool {
	return cell.Value().Wall
}

func cost(_, next grid.Cell[Tile]) float64 {
	return next.Value().Cost
}

func total(g *grid.Grid[Tile], path []grid.Index) float64 {
	sum := 0.0
	for i := 1; i < len(path); i++ {
		sum += cost(g.Cell(path[i-1]), g.Cell(path[i]))
	}

	return sum
}
