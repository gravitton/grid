package main

import (
	"fmt"
	"iter"
	"strconv"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/grid"
	hex "github.com/gravitton/hexagon"
)

var figures = []figure{
	{"hero", hero},
	{"explorations", explorations},
	{"grids", grids},
	{"world", world},
	{"draw-order", drawOrder},
	{"movement", movement},
	{"range", ranges},
	{"traversal", traversal},
	{"corners", corners},
	{"sight", sight},
	{"paths", paths},
	{"fields", fields},
	{"regions", regions},
	{"squares", squares},
	{"hex-layout", hexLayout},
}

var (
	cardinal = rect(grid.MovementCardinal)
	diagonal = rect(grid.MovementDiagonal)
)

type layout struct {
	name  string
	build builder
	width float64
}

func layouts(square, diamond, flat, pointy float64) []layout {
	return []layout{
		{"NewRect", cardinal, square},
		{"NewIsometric", isometric(grid.MovementCardinal), diamond},
		{"NewHexFlatTop", hexFlatTop, flat},
		{"NewHexPointyTop", hexPointyTop, pointy},
	}
}

func explorations(t theme) []*panel {
	panels := make([]*panel, 0, 3)
	for _, l := range []layout{
		{"NewRect", cardinal, 26},
		{"NewIsometric", isometric(grid.MovementCardinal), 43},
		{"NewHexPointyTop", hexPointyTop, 28},
	} {
		g, m := paint(l.build, l.width,
			"..#......",
			"..#...~..",
			"S.#..~~~.",
			".....~~~.",
			"..#...~#G",
			"..#....#.",
			"..#......",
			".........",
		)
		p := newPanel(t, g, "grid."+l.name)
		visible, path := p.explore(g.All(), m.at('S'), m.at('G'), 5)
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

func grids(t theme) []*panel {
	panels := make([]*panel, 0, 4)
	for _, l := range layouts(28, 48, 32, 30) {
		g := l.build(geom.Sz(5, 4), l.width)
		p := backdrop(t, g, "grid."+l.name)
		for cell := range g.All() {
			p.coordinate(cell.Index(), 9.5)
		}
		p.frame(g.Bounds(), t.muted)
		p.circle(g.Origin(), 3, style{fill: t.ink})
		panels = append(panels, p)
	}

	return panels
}

func world(t theme) []*panel {
	colors := []string{t.accent, t.cool, t.leaf, t.muted}

	panels := make([]*panel, 0, 4)
	for _, l := range layouts(32, 52, 36, 34) {
		g := l.build(geom.Sz(4, 3), l.width)
		p := newPanel(t, g, "g.IndexAt(point)")
		p.note = l.name
		bounds := g.Bounds()
		for y := bounds.Min.Y + 2; y < bounds.Max.Y; y += 5 {
			for x := bounds.Min.X + 2; x < bounds.Max.X; x += 5 {
				at := geom.Pt(x, y)
				index := g.IndexAt(at)
				if g.Contains(index) {
					p.circle(at, 1.3, style{fill: colors[geom.Mod(index.X, 2)+2*geom.Mod(index.Y, 2)]})
				}
			}
		}
		for cell := range g.All() {
			p.outline(cell.Index(), t.edge, 1)
		}
		panels = append(panels, p)
	}

	return panels
}

func movement(t theme) []*panel {
	center := geom.Pt(3, 3)
	movements := []struct {
		name   string
		metric string
		build  builder
	}{
		{"cardinal", "Manhattan", cardinal},
		{"diagonal", "Chebyshev", diagonal},
	}

	neighbors := make([]*panel, 0, len(movements))
	distances := make([]*panel, 0, len(movements))
	for _, m := range movements {
		g := m.build(geom.Sz(7, 7), 20)

		around := backdrop(t, g, "g.Neighbors(c)")
		around.note = m.name
		for i, index := range g.Neighbors(center) {
			around.cell(index, t.accentSoft)
			around.label(index, strconv.Itoa(i), t.ink)
		}
		around.cell(center, t.accent)
		around.label(center, "c", t.onAccent)
		neighbors = append(neighbors, around)

		far := backdrop(t, g, "g.Distance(c, index)")
		far.note = m.name + ", " + m.metric
		for cell := range g.All() {
			index := cell.Index()
			distance := g.Distance(center, index)
			if distance == 1 {
				far.cell(index, t.accentSoft)
			}
			far.label(index, strconv.Itoa(distance), t.ink)
		}
		far.cell(center, t.accent)
		far.label(center, "0", t.onAccent)
		distances = append(distances, far)
	}

	return append(neighbors, distances...)
}

func ranges(t theme) []*panel {
	center := geom.Pt(3, 3)
	kinds := []struct {
		note  string
		build builder
		width float64
	}{
		{"cardinal", cardinal, 20},
		{"diagonal", diagonal, 20},
		{"flat-top", hexFlatTop, 24},
		{"pointy-top", hexPointyTop, 22},
	}

	panels := make([]*panel, 0, len(kinds))
	for _, k := range kinds {
		g := k.build(geom.Sz(7, 7), k.width)
		p := backdrop(t, g, "g.Range(c, 2)")
		p.note = k.note
		p.region(g.Range(center, 2), t.accentSoft)
		p.cell(center, t.accent)
		p.label(center, "c", t.onAccent)
		panels = append(panels, p)
	}

	return panels
}

func traversal(t theme) []*panel {
	center := geom.Pt(3, 3)
	g := diagonal(geom.Sz(7, 7), 22)

	ring := backdrop(t, g, "g.Ring(c, 2)")
	ring.cells(g.Ring(center, 2), t.accentSoft)
	ring.path(append(g.Ring(center, 2), g.Ring(center, 2)[0]), t.accent)
	for i, index := range g.Ring(center, 2) {
		ring.label(index, strconv.Itoa(i), t.ink)
	}
	ring.cell(center, t.accent)
	ring.label(center, "c", t.onAccent)

	spiral := backdrop(t, g, "g.Spiral(c, 2)")
	spiral.cells(g.Spiral(center, 2), t.accentSoft)
	spiral.path(g.Spiral(center, 2), t.accent)
	for i, index := range g.Spiral(center, 2) {
		spiral.label(index, strconv.Itoa(i), t.ink)
	}

	from, to := geom.Pt(0, 5), geom.Pt(6, 1)
	panels := []*panel{ring, spiral}
	for _, k := range []struct {
		note  string
		build builder
	}{
		{"diagonal", diagonal},
		{"cardinal", cardinal},
	} {
		lined := k.build(geom.Sz(7, 7), 22)
		line := backdrop(t, lined, "g.Line(a, b)")
		line.note = k.note
		line.cells(lined.Line(from, to), t.accentSoft)
		line.segment(from, to, t.ink, false)
		line.cell(from, t.accent)
		line.cell(to, t.accent)
		line.label(from, "a", t.onAccent)
		line.label(to, "b", t.onAccent)
		panels = append(panels, line)
	}

	return panels
}

func sight(t theme) []*panel {
	rows := []string{
		".............",
		".............",
		"...#.........",
		"........##...",
		".............",
		"..#..........",
		"......c...#..",
		".........#...",
		"....#........",
		"....#....#...",
		".............",
		"......##.....",
		".............",
	}
	kinds := []struct {
		note  string
		build builder
	}{
		{"cardinal", cardinal},
		{"diagonal", diagonal},
		{"pointy-top", hexPointyTop},
	}

	panels := make([]*panel, 0, len(kinds))
	for _, k := range kinds {
		g, m := paint(k.build, 15, rows...)
		center := m.at('c')
		visible := g.FieldOfView(center, 6, opaque)

		p := newPanel(t, g, "g.FieldOfView(c, 6, opaque)")
		p.note = k.note
		for cell := range g.All() {
			index := cell.Index()
			switch {
			case opaque(cell):
				p.cell(index, t.wall)
			case visible.Contains(index):
				p.cell(index, t.accentSoft)
			case g.Distance(center, index) <= 6:
				p.cell(index, t.shade)
			default:
				p.cell(index, t.cell)
			}
		}
		p.cell(center, t.accent)
		panels = append(panels, p)
	}

	return panels
}

func corners(t theme) []*panel {
	cases := []struct {
		name string
		rows []string
	}{
		{"in the way", []string{"a...", ".#..", "....", "...b"}},
		{"a pillar", []string{"a#..", "....", "....", "...b"}},
		{"a diagonal wall", []string{"a#..", "#...", "....", "...b"}},
	}

	panels := make([]*panel, 0, len(cases))
	for _, c := range cases {
		g, m := paint(cardinal, 26, c.rows...)
		a, b := m.at('a'), m.at('b')
		seen := g.HasLineOfSight(a, b, opaque)

		p := backdrop(t, g, c.name+": "+verdict(seen))
		if seen {
			p.segment(a, b, t.accent, false)
		} else {
			p.segment(a, b, t.muted, true)
		}
		p.cell(a, t.accent)
		p.cell(b, t.accentSoft)
		p.label(a, "a", t.onAccent)
		p.label(b, "b", t.ink)
		panels = append(panels, p)
	}

	return panels
}

func paths(t theme) []*panel {
	g, m := paint(cardinal, 14,
		"...............",
		"........~~~....",
		"....#...~~~....",
		"....#...~~~....",
		".a..#...~~~..b.",
		"....#...~~~....",
		"....#...~~~....",
		"........~~~....",
		"...............",
	)
	start, goal := m.at('a'), m.at('b')
	searches := []struct {
		name string
		path []grid.Index
	}{
		{"g.PathBreadthFirst(a, b, walkable)", g.PathBreadthFirst(start, goal, walkable)},
		{"g.Path(a, b, walkable, cost)", g.Path(start, goal, walkable, cost)},
	}

	panels := make([]*panel, 0, len(searches))
	for _, s := range searches {
		p := backdrop(t, g, s.name)
		p.note = fmt.Sprintf("%d steps, cost %g", len(s.path)-1, total(g, s.path))
		p.path(s.path, t.accent)
		p.cell(start, t.accent)
		p.cell(goal, t.cool)
		p.label(start, "a", t.onAccent)
		p.label(goal, "b", t.onAccent)
		panels = append(panels, p)
	}

	return panels
}

func fields(t theme) []*panel {
	g, m := paint(cardinal, 22,
		"...........",
		"...........",
		"....##.....",
		"......#....",
		"..~~.u.#...",
		"..~~.......",
		"....###....",
		".......p...",
		"...........",
	)
	unit, picked := m.at('u'), m.at('p')
	reach := g.Reachable(unit, 6, walkable, cost)

	reachable := backdrop(t, g, "g.Reachable(u, 6, walkable, cost)")
	reachable.note = "field.CostTo(index) and field.PathTo(p)"
	for index := range reach.All() {
		if *g.Cell(index).Value() == ground {
			reachable.cell(index, t.accentSoft)
		}
	}
	reachable.path(reach.PathTo(picked), t.accent)
	for index := range reach.All() {
		spent, _ := reach.CostTo(index)
		reachable.label(index, strconv.FormatFloat(spent, 'f', -1, 64), t.ink)
	}
	reachable.cell(unit, t.accent)
	reachable.label(unit, "u", t.onAccent)
	reachable.outline(picked, t.accent, 2)

	h, n := paint(cardinal, 22,
		"...........",
		"..####.....",
		".....#.....",
		".....#..g..",
		".....#.....",
		"..####.....",
		"...........",
		"...#####...",
		"...........",
	)
	goal := n.at('g')
	toward := h.Field(goal, walkable, nil)

	flow := backdrop(t, h, "g.Field(g, walkable, nil)")
	flow.note = "field.Next(index), one step toward g"
	for index := range toward.All() {
		if next, ok := toward.Next(index); ok {
			flow.pointer(index, next, t.accent)
		}
	}
	flow.cell(goal, t.cool)
	flow.label(goal, "g", t.onAccent)

	return []*panel{reachable, flow}
}

func regions(t theme) []*panel {
	rows := []string{
		"............",
		".ooooooo....",
		".oo.oooo.oo.",
		".ooooooo.oo.",
		".ooooooo....",
		"............",
		"............",
	}
	g, m := paint(cardinal, 17, rows...)
	region := m.region('o')

	components := backdrop(t, g, "g.Components(r)")
	colors := []string{t.accent, t.cool, t.leaf}
	for i, component := range g.Components(region) {
		components.region(component, colors[i%len(colors)])
	}

	grow := backdrop(t, g, "g.Grow(r, 1)")
	grow.region(g.Grow(region, 1), t.coolSoft)
	grow.region(region, t.accentSoft)

	shrink := backdrop(t, g, "g.Shrink(r, 1)")
	shrink.region(region, t.accentSoft)
	shrink.region(g.Shrink(region, 1), t.accent)

	border := backdrop(t, g, "g.Border(r)")
	border.region(region, t.accentSoft)
	border.region(g.Border(region), t.accent)

	return []*panel{components, grow, shrink, border}
}

func drawOrder(t theme) []*panel {
	panels := make([]*panel, 0, 4)
	for _, l := range layouts(24, 44, 28, 26) {
		g := l.build(geom.Sz(4, 4), l.width)
		p := backdrop(t, g, "g.DrawOrder(box)")
		p.note = l.name
		order := make([]grid.Index, 0, 16)
		for cell := range g.DrawOrder(geom.Box[float64]{}) {
			order = append(order, cell.Index())
		}
		p.path(order, t.accentSoft)
		for i, index := range order {
			p.label(index, strconv.Itoa(i), t.ink)
		}
		panels = append(panels, p)
	}

	return panels
}

func verdict(seen bool) string {
	if seen {
		return "seen"
	}

	return "hidden"
}

func squares(t theme) []*panel {
	s := grid.Sq(3, 3)
	g := cardinal(geom.Sz(7, 7), 22)
	movements := []struct {
		note     string
		movement grid.Movement
	}{
		{"MovementCardinal", grid.MovementCardinal},
		{"MovementDiagonal", grid.MovementDiagonal},
	}

	panels := make([]*panel, 0, len(movements)+1)
	for _, m := range movements {
		p := backdrop(t, g, "s.Sector(3, geom.DirectionUp, 1, movement)")
		p.note = m.note
		for i, index := range s.Sector(3, geom.DirectionUp, 1, m.movement) {
			p.cell(index, t.accentSoft)
			p.label(index, strconv.Itoa(i), t.ink)
		}
		p.cell(s.Point(), t.accent)
		p.label(s.Point(), "s", t.onAccent)
		panels = append(panels, p)
	}

	a := grid.Sq(5, 2)
	turn := backdrop(t, g, "a.TurnAround(s, steps)")
	for steps := range 4 {
		turned := a.TurnAround(s.Point(), steps).Point()
		turn.cell(turned, t.accentSoft)
		turn.label(turned, strconv.Itoa(steps), t.ink)
	}
	turn.cell(s.Point(), t.cool)
	turn.label(s.Point(), "s", t.onAccent)
	turn.cell(a.Point(), t.accent)
	turn.label(a.Point(), "a", t.onAccent)

	return append(panels, turn)
}

func hexLayout(t theme) []*panel {
	views := []struct {
		build  func(cellScale geom.Size[float64], system hex.CoordinateSystem) *grid.HexLayout
		scale  geom.Size[float64]
		system hex.CoordinateSystem
		name   string
	}{
		{grid.NewHexPointyTopLayout, grid.HexPointyTopCellScale(34), hex.Axial, "NewHexPointyTopLayout"},
		{grid.NewHexFlatTopLayout, grid.HexFlatTopCellScale(38), hex.Axial, "NewHexFlatTopLayout"},
		{grid.NewHexPointyTopLayout, grid.HexPointyTopCellScale(34), hex.OffsetOddR, "NewHexPointyTopLayout"},
	}

	panels := make([]*panel, 0, len(views))
	for _, v := range views {
		layout := v.build(v.scale, v.system)
		p := place(t, layout, "grid."+v.name)
		p.note = "hex." + v.system.String()
		for _, h := range hex.Pt(0, 0).Range(2) {
			index := v.system.To(h)
			if h.IsZero() {
				p.cell(index, t.accentSoft)
			} else {
				p.cell(index, t.cell)
			}
			p.coordinate(index, 10)
		}
		panels = append(panels, p)
	}

	cursor := geom.Pt(60.0, -40.0)
	first := panels[0]
	first.outline(first.place.(*grid.HexLayout).IndexAt(cursor), t.accent, 2)
	first.circle(cursor, 3.5, style{fill: t.accent})

	return panels
}
