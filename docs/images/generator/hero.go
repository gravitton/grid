package main

import (
	"fmt"
	"iter"

	"github.com/gravitton/grid"
)

var heroMap = []string{
	"...............",
	"..#.....#......",
	"..#.....#....G.",
	"..#..~~.#......",
	"....S~~........",
	"..#..~~...##...",
	"..#.......#....",
	"..........#....",
	"...............",
}

func hero(t theme) []*panel {
	panels := make([]*panel, 0, 3)
	for _, l := range []layout{
		{"NewRect", cardinal, 18},
		{"NewIsometric", isometric(grid.MovementCardinal), 23},
		{"NewHexPointyTop", hexPointyTop, 18},
	} {
		g, m := paint(l.build, l.width, heroMap...)
		p := newPanel(t, g, "grid."+l.name)
		visible, path := p.explore(g.All(), m.at('S'), m.at('G'), 6)
		p.note = fmt.Sprintf("sees %d, path of %d at %g", visible.Len(), len(path), total(g, path))
		panels = append(panels, p)
	}

	return panels
}

func (p *panel) explore(cells iter.Seq[grid.Cell[Tile]], scout, goal grid.Index, radius int) (grid.Region, []grid.Index) {
	visible := p.grid.FieldOfView(scout, radius, opaque)
	path := p.grid.Path(scout, goal, walkable, cost)

	for cell := range cells {
		index := cell.Index()
		switch {
		case *cell.Value() != ground:
			p.cell(index, p.terrain(index))
		case visible.Contains(index):
			p.cell(index, p.theme.accentSoft)
		case p.grid.Distance(scout, index) <= radius:
			p.cell(index, p.theme.shade)
		default:
			p.cell(index, p.theme.cell)
		}
	}
	p.path(path, p.theme.accent)
	p.cell(scout, p.theme.accent)
	p.cell(goal, p.theme.cool)

	return visible, path
}
