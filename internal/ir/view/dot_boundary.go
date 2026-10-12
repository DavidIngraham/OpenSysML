package view

import (
	"cmp"
	"math"
	"slices"
)

// dotBoundaryGap is the space, in pixels, between the boundary symbols of a
// node spread along one of its sides.
const dotBoundaryGap = 6

// boundarySide is a side of a boxed node, where the members of it that routes
// place sit as its boundary symbols.
type boundarySide int

const (
	leftSide boundarySide = iota
	rightSide
	topSide
	bottomSide
)

// across reports whether the side runs across the page, so that the symbols on
// it spread along x rather than y.
func (s boundarySide) across() bool { return s == topSide || s == bottomSide }

// boundarySymbol is a member of a boxed node that its routes place: the box the
// routes gave it, where they meet it and the mean of where they come from.
type boundarySymbol struct {
	id       string
	box      nodeBox
	at, from Point
}

// along is the coordinate of p along side.
func (s boundarySide) along(p Point) float64 {
	if s.across() {
		return p.X
	}
	return p.Y
}

// length is a box's extent along side.
func (s boundarySide) length(b nodeBox) float64 {
	width, height := b.size()
	if s.across() {
		return width
	}
	return height
}

// sideToward is the side of b a route reaching it at `at` from the direction
// dir comes to: the side nearest the point it reaches, the one facing most
// squarely where the route comes from when the point is as near to two.
func (b nodeBox) sideToward(at, dir Point) boundarySide {
	nearness := [4]float64{leftSide: -math.Abs(at.X - b.low.X), rightSide: -math.Abs(b.high.X - at.X),
		topSide: -math.Abs(at.Y - b.low.Y), bottomSide: -math.Abs(b.high.Y - at.Y)}
	facing := [4]float64{leftSide: -dir.X, rightSide: dir.X, topSide: -dir.Y, bottomSide: dir.Y}
	best := leftSide
	for side := rightSide; side <= bottomSide; side++ {
		if nearness[side] > nearness[best] || nearness[side] == nearness[best] && facing[side] > facing[best] {
			best = side
		}
	}
	return best
}

// toward is the point of the box's border on the way from its centre toward p;
// the centre when p is the centre or lies within the box.
func (b nodeBox) toward(p Point) Point {
	c := b.centre()
	dx, dy := p.X-c.X, p.Y-c.Y
	if b.holds(p) || dx == 0 && dy == 0 {
		return c
	}
	t := math.Min((b.high.X-b.low.X)/2/math.Abs(dx), (b.high.Y-b.low.Y)/2/math.Abs(dy))
	return Point{X: halfPixel(c.X + dx*t), Y: halfPixel(c.Y + dy*t)}
}

// spreadBoundaries moves apart, under every boxed cluster at or under node, the
// members the routes place over one another, the boundary symbols of the
// cluster: those a side's routes meet are spread along that side, touching its
// border, in the order the routes come from. A symbol overlapping none of its
// side keeps the place its routes give it.
func (w *dotWriter) spreadBoundaries(node *Node, ends map[string][]routeEnd) {
	for _, child := range node.Children {
		w.spreadBoundaries(child, ends)
	}
	box, ok := w.boxes[node.ID]
	if !ok || len(node.Children) == 0 || w.tree {
		return
	}
	var sides [4][]boundarySymbol
	for _, child := range node.Children {
		symbol, ok := w.boundarySymbol(child, ends)
		if !ok {
			continue
		}
		side := box.sideToward(symbol.at, Point{X: symbol.from.X - symbol.at.X, Y: symbol.from.Y - symbol.at.Y})
		sides[side] = append(sides[side], symbol)
	}
	for side, symbols := range sides {
		if len(symbols) > 1 {
			w.spreadAlong(box, boundarySide(side), symbols)
		}
	}
}

// boundarySymbol is the member node as a boundary symbol of its cluster when
// the routes of its edges place it: a node with no Layout, no members and no
// neighbour-derived place that a route reaches or stops short of.
func (w *dotWriter) boundarySymbol(node *Node, ends map[string][]routeEnd) (boundarySymbol, bool) {
	box, ok := w.boxes[node.ID]
	if !ok || node.Geometry != nil || len(node.Children) > 0 || w.placement.derived[node.ID] {
		return boundarySymbol{}, false
	}
	reaching := reachingEnds(ends[node.ID])
	if len(reaching) == 0 {
		return boundarySymbol{}, false
	}
	symbol := boundarySymbol{id: node.ID, box: box}
	for _, end := range reaching {
		symbol.at.X += end.at.X
		symbol.at.Y += end.at.Y
		symbol.from.X += end.next.X
		symbol.from.Y += end.next.Y
	}
	n := float64(len(reaching))
	symbol.at = Point{X: symbol.at.X / n, Y: symbol.at.Y / n}
	symbol.from = Point{X: symbol.from.X / n, Y: symbol.from.Y / n}
	return symbol, true
}

// symbolRun is a run of boundary symbols laid end to end along a side, a gap
// apart, centred on the mean of where their routes meet the side; a run of one
// symbol lies where the routes place it.
type symbolRun struct {
	symbols []boundarySymbol
	length  float64
	ideal   float64
}

// spreadAlong spreads the symbols on side of box that overlap one another along
// it: symbols sorted along the side, by where their routes meet it and then by
// where the routes come from, are pooled into runs that merge while
// neighbouring runs overlap, each run centred on the mean place of its symbols
// and kept within the side when it fits, at the side's middle when it does not.
// A symbol in a run of its own is left as the routes place it.
func (w *dotWriter) spreadAlong(box nodeBox, side boundarySide, symbols []boundarySymbol) {
	slices.SortStableFunc(symbols, func(a, b boundarySymbol) int {
		if c := cmp.Compare(side.along(a.at), side.along(b.at)); c != 0 {
			return c
		}
		return cmp.Compare(side.along(a.from), side.along(b.from))
	})
	low, high := side.along(box.low), side.along(box.high)
	var runs []symbolRun
	for _, symbol := range symbols {
		runs = append(runs, symbolRun{symbols: []boundarySymbol{symbol}, length: side.length(symbol.box), ideal: side.along(symbol.at)})
		for n := len(runs); n > 1 && runs[n-2].end(side, low, high)+dotBoundaryGap > runs[n-1].start(side, low, high); n = len(runs) {
			runs = append(runs[:n-2], runs[n-2].merge(runs[n-1]))
		}
	}
	for _, run := range runs {
		if len(run.symbols) < 2 {
			continue
		}
		at := run.start(side, low, high)
		for _, symbol := range run.symbols {
			length := side.length(symbol.box)
			w.boxes[symbol.id] = box.boundaryBox(side, symbol.box, at+length/2)
			w.spread[symbol.id] = true
			at += length + dotBoundaryGap
		}
	}
}

// centre is where the run is centred along side, between its ends low and
// high: on its ideal, moved within the side when it fits, else the side's
// middle. A single symbol stays where the routes place it.
func (r symbolRun) centre(side boundarySide, low, high float64) float64 {
	if len(r.symbols) < 2 {
		return side.along(r.symbols[0].box.centre())
	}
	if r.length > high-low {
		return (low + high) / 2
	}
	return math.Max(low+r.length/2, math.Min(r.ideal, high-r.length/2))
}

// start and end are where the run begins and ends along the side.
func (r symbolRun) start(side boundarySide, low, high float64) float64 {
	return r.centre(side, low, high) - r.length/2
}
func (r symbolRun) end(side boundarySide, low, high float64) float64 {
	return r.centre(side, low, high) + r.length/2
}

// merge is the run of r's symbols followed by next's, centred on the mean of
// both runs' ideals weighted by their symbols.
func (r symbolRun) merge(next symbolRun) symbolRun {
	n, m := float64(len(r.symbols)), float64(len(next.symbols))
	return symbolRun{symbols: append(r.symbols, next.symbols...), length: r.length + dotBoundaryGap + next.length,
		ideal: (r.ideal*n + next.ideal*m) / (n + m)}
}

// boundaryBox is the box of a symbol of the given box's size on side of the
// cluster's box, touching the border from inside, centred at `at` along it.
func (b nodeBox) boundaryBox(side boundarySide, symbol nodeBox, at float64) nodeBox {
	width, height := symbol.size()
	var c Point
	switch side {
	case leftSide:
		c = Point{X: b.low.X + width/2, Y: at}
	case rightSide:
		c = Point{X: b.high.X - width/2, Y: at}
	case topSide:
		c = Point{X: at, Y: b.low.Y + height/2}
	case bottomSide:
		c = Point{X: at, Y: b.high.Y - height/2}
	}
	return nodeBox{low: Point{X: c.X - width/2, Y: c.Y - height/2}, high: Point{X: c.X + width/2, Y: c.Y + height/2}}
}
