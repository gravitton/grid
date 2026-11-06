package main

import (
	"fmt"
	"math"

	geom "github.com/gravitton/geometry"
	"github.com/gravitton/grid"
)

type placement interface {
	Center(index grid.Index) geom.Point[float64]
	Polygon(index grid.Index) geom.RegularPolygon[float64]
	CellSize() geom.Size[float64]
}

type cells struct {
	*grid.Grid[Tile]
}

func (c cells) Center(index grid.Index) geom.Point[float64] {
	return c.Cell(index).Center()
}

func (c cells) Polygon(index grid.Index) geom.RegularPolygon[float64] {
	return c.Cell(index).Polygon()
}

type panel struct {
	canvas
	place placement
	grid  *grid.Grid[Tile]
}

func newPanel(t theme, g *grid.Grid[Tile], caption string) *panel {
	p := place(t, cells{g}, caption)
	p.grid = g

	return p
}

func place(t theme, at placement, caption string) *panel {
	return &panel{canvas: newCanvas(t, caption), place: at}
}

func window(t theme, g *grid.Grid[Tile], view geom.Box[float64]) *panel {
	p := newPanel(t, g, "")
	p.clip(view)

	return p
}

func backdrop(t theme, g *grid.Grid[Tile], caption string) *panel {
	p := newPanel(t, g, caption)
	for cell := range g.All() {
		p.cell(cell.Index(), p.terrain(cell.Index()))
	}

	return p
}

func (p *panel) size() float64 {
	size := p.place.CellSize()

	return min(size.Width, size.Height) / 2
}

func (p *panel) center(index grid.Index) point {
	return p.place.Center(index)
}

func (p *panel) terrain(index grid.Index) string {
	switch *p.grid.Cell(index).Value() {
	case wall:
		return p.theme.wall
	case rough:
		return p.theme.leafSoft
	default:
		return p.theme.cell
	}
}

func (p *panel) cell(index grid.Index, fill string) {
	p.tile(index, style{fill: fill, stroke: p.theme.edge, width: 1})
}

func (p *panel) cells(indices []grid.Index, fill string) {
	for _, index := range indices {
		p.cell(index, fill)
	}
}

func (p *panel) region(region grid.Region, fill string) {
	for index := range region.All() {
		p.cell(index, fill)
	}
}

func (p *panel) outline(index grid.Index, color string, width float64) {
	p.tile(index, style{stroke: color, width: width})
}

func (p *panel) tile(index grid.Index, s style) {
	p.polygon(p.place.Polygon(index).Polygon().Points, s)
}

func (p *panel) text(index grid.Index, dy float64, f font, spans ...span) {
	p.canvas.text(p.center(index).AddXY(0, dy), f, spans...)
}

func (p *panel) label(index grid.Index, text, color string) {
	p.text(index, 0, font{size: math.Round(p.size() * 0.9)}, span{text, color})
}

func (p *panel) coordinate(index grid.Index, size float64) {
	p.text(index, 0, font{size: size}, span{fmt.Sprint(index.X), p.theme.accent}, span{",", p.theme.muted}, span{fmt.Sprint(index.Y), p.theme.cool})
}

func (p *panel) segment(a, b grid.Index, color string, dashed bool) {
	p.line(p.center(a), p.center(b), style{stroke: color, width: 2, dashed: dashed})
}

func (p *panel) path(indices []grid.Index, color string) {
	points := make([]point, len(indices))
	for i, index := range indices {
		points[i] = p.center(index)
	}

	p.polyline(points, style{stroke: color, width: 2})
}

func (p *panel) pointer(from, to grid.Index, color string) {
	a, b := p.center(from), p.center(to)
	along := b.Subtract(a)

	p.arrow(a.Add(along.Multiply(-0.3)), a.Add(along.Multiply(0.4)), style{stroke: color, width: 1.5})
}

func (p *panel) frame(box geom.Box[float64], color string) {
	p.polygon(box.Polygon().Points, style{stroke: color, width: 1.2, dashed: true})
}
