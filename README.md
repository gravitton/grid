# Grid

[![Latest Stable Version][ico-release]][link-release]
[![Build Status][ico-workflow]][link-workflow]
[![Coverage Status][ico-coverage]][link-coverage]
[![Go Report Card][ico-go-report-card]][link-go-report-card]
[![Go Dev Reference][ico-go-dev-reference]][link-go-dev-reference]
[![Software License][ico-license]][link-licence]


## Installation

```bash
go get github.com/gravitton/grid
```


## Usage

```go
package main

import (
	geom "github.com/gravitton/geometry"
	"github.com/gravitton/grid"
)

type Tile struct {
	Revealed       bool
	MovementCost   int
}

g := grid.NewHexagonFlatTopGrid[Tile](geom.Sz(100, 100), geom.SzU(32), true)

cell := g.At(geom.Pt(499.0, 123.4))
cell.Index()
cell.Valid()
cell.Set(&Tile{Revealed: true, MovementCost: 1})

path := g.Path(geom.Pt(0, 0), geom.Pt(10, 10), func(current *grid.Cell[resource.Tile]) bool {
	return current.Get().Revealed
}, func(current, next *grid.Cell[resource.Tile]) float64 {
	return next.Get().MovementCost
})

```


## Credits

- [Tomáš Novotný](https://github.com/tomas-novotny)
- [All Contributors][link-contributors]


## License

The MIT License (MIT). Please see [License File][link-licence] for more information.


[ico-license]:              https://img.shields.io/github/license/gravitton/grid.svg?style=flat-square&colorB=blue
[ico-workflow]:             https://img.shields.io/github/actions/workflow/status/gravitton/grid/main.yml?branch=main&style=flat-square
[ico-release]:              https://img.shields.io/github/v/release/gravitton/grid?style=flat-square&colorB=blue
[ico-go-dev-reference]:     https://img.shields.io/badge/go.dev-reference-blue?style=flat-square
[ico-go-report-card]:       https://goreportcard.com/badge/github.com/gravitton/grid?style=flat-square
[ico-coverage]:             https://img.shields.io/coverallsCoverage/github/gravitton/grid?style=flat-square

[link-author]:              https://github.com/gravitton
[link-release]:             https://github.com/gravitton/grid/releases
[link-contributors]:        https://github.com/gravitton/grid/contributors
[link-licence]:             ./LICENSE.md
[link-changelog]:           ./CHANGELOG.md
[link-workflow]:            https://github.com/gravitton/grid/actions
[link-go-dev-reference]:    https://pkg.go.dev/github.com/gravitton/grid
[link-go-report-card]:      https://goreportcard.com/report/github.com/gravitton/grid
[link-coverage]:            https://coveralls.io/github/gravitton/grid
