package view

import (
	"cmp"
	"math"
	"slices"
	"strings"
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

// through is the coordinate of p through side, from the border inward.
func (s boundarySide) through(p Point) float64 {
	if s.across() {
		return p.Y
	}
	return p.X
}

// length and depth are a box's extents along side and through it.
func (s boundarySide) length(b nodeBox) float64 {
	width, height := b.size()
	if s.across() {
		return width
	}
	return height
}
func (s boundarySide) depth(b nodeBox) float64 {
	width, height := b.size()
	if s.across() {
		return height
	}
	return width
}

// isBoundaryKind reports whether a kind is drawn as a boundary symbol of the
// node declaring it: a port usage, or a case's actor, subject or stakeholder
// parameter (SysML v2 §7.26.5, §8.2.3.21).
func isBoundaryKind(kind string) bool {
	if isPortKind(kind) {
		return true
	}
	for _, word := range strings.Fields(kind) {
		switch word {
		case "actor", "subject", "stakeholder":
			return !isDefinitionKind(kind)
		}
	}
	return false
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
// boundary symbols the routes place over one another: those a side's routes
// meet are spread along that side, touching its border, in the order the routes
// come from, and a side's symbols keep out of a corner the symbols of the sides
// before it hold. A symbol overlapping none keeps the place its routes give it.
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
	var placed []nodeBox
	for side := leftSide; side <= bottomSide; side++ {
		if len(sides[side]) == 0 {
			continue
		}
		w.spreadAlong(box, side, sides[side], placed)
		for _, symbol := range sides[side] {
			placed = append(placed, w.boxes[symbol.id])
		}
	}
}

// boundarySymbol is the member node as a boundary symbol of its cluster when
// the routes of its edges place it: a node of a boundary kind with no Layout, no
// members and no neighbour-derived place that a route reaches or stops short of.
func (w *dotWriter) boundarySymbol(node *Node, ends map[string][]routeEnd) (boundarySymbol, bool) {
	box, ok := w.boxes[node.ID]
	if !ok || !isBoundaryKind(node.Kind) || node.Geometry != nil || len(node.Children) > 0 || w.placement.derived[node.ID] {
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

// sideRange is the stretch of a side its symbols may take: the side, less the
// corners symbols of the sides spread before it reach into.
type sideRange struct {
	low, high       float64
	lowCut, highCut bool
}

// freeRange is the range along side left to symbols of the given depth by the
// obstacles reaching into the lane such symbols occupy inside the border: an
// obstacle nearer the side's start cuts the range there, else at its end.
func (b nodeBox) freeRange(side boundarySide, depth float64, obstacles []nodeBox) sideRange {
	r := sideRange{low: side.along(b.low), high: side.along(b.high)}
	laneLow, laneHigh := b.lane(side, depth)
	for _, o := range obstacles {
		if side.through(o.high) <= laneLow || side.through(o.low) >= laneHigh {
			continue
		}
		start, end := side.along(o.low), side.along(o.high)
		if (start+end)/2 < (r.low+r.high)/2 {
			r.low, r.lowCut = math.Max(r.low, end+dotBoundaryGap), true
		} else {
			r.high, r.highCut = math.Min(r.high, start-dotBoundaryGap), true
		}
	}
	return r
}

// lane is the band through side, from its border inward by depth, that symbols
// of that depth touching the border occupy.
func (b nodeBox) lane(side boundarySide, depth float64) (low, high float64) {
	switch side {
	case leftSide:
		return b.low.X, b.low.X + depth
	case rightSide:
		return b.high.X - depth, b.high.X
	case topSide:
		return b.low.Y, b.low.Y + depth
	default:
		return b.high.Y - depth, b.high.Y
	}
}

// spreadAlong spreads the symbols on side of box that overlap one another along
// it: symbols sorted along the side, by where their routes meet it and then by
// where the routes come from, are pooled into runs that merge while
// neighbouring runs overlap, each run centred on the mean place of its symbols
// and kept within the side's free range when it fits, at the range's middle when
// it does not. A symbol in a run of its own is left as the routes place it
// unless an obstacle cuts it off the range.
func (w *dotWriter) spreadAlong(box nodeBox, side boundarySide, symbols []boundarySymbol, obstacles []nodeBox) {
	slices.SortStableFunc(symbols, func(a, b boundarySymbol) int {
		if c := cmp.Compare(side.along(a.at), side.along(b.at)); c != 0 {
			return c
		}
		return cmp.Compare(side.along(a.from), side.along(b.from))
	})
	var depth float64
	for _, symbol := range symbols {
		depth = math.Max(depth, side.depth(symbol.box))
	}
	rng := box.freeRange(side, depth, obstacles)
	var runs []symbolRun
	for _, symbol := range symbols {
		runs = append(runs, symbolRun{symbols: []boundarySymbol{symbol}, length: side.length(symbol.box), ideal: side.along(symbol.at)})
		for n := len(runs); n > 1 && runs[n-2].end(side, rng)+dotBoundaryGap > runs[n-1].start(side, rng); n = len(runs) {
			runs = append(runs[:n-2], runs[n-2].merge(runs[n-1]))
		}
	}
	for _, run := range runs {
		if len(run.symbols) < 2 && run.centre(side, rng) == side.along(run.symbols[0].box.centre()) {
			continue
		}
		at := run.start(side, rng)
		for _, symbol := range run.symbols {
			length := side.length(symbol.box)
			w.boxes[symbol.id] = box.boundaryBox(side, symbol.box, at+length/2)
			w.spread[symbol.id] = true
			at += length + dotBoundaryGap
		}
	}
}

// centre is where the run is centred along side, within the range rng: on its
// ideal, moved within the range when it fits; when it does not, it starts at a
// cut-off corner and runs past the far end, clear of the corner's symbol, or
// sits at the middle of an uncut side. A single symbol stays where the routes
// place it, moved only off a cut-off corner.
func (r symbolRun) centre(side boundarySide, rng sideRange) float64 {
	if len(r.symbols) < 2 {
		c := side.along(r.symbols[0].box.centre())
		if rng.highCut {
			c = math.Min(c, rng.high-r.length/2)
		}
		if rng.lowCut {
			c = math.Max(c, rng.low+r.length/2)
		}
		return c
	}
	if r.length > rng.high-rng.low {
		switch {
		case rng.lowCut:
			return rng.low + r.length/2
		case rng.highCut:
			return rng.high - r.length/2
		}
		return (rng.low + rng.high) / 2
	}
	return math.Max(rng.low+r.length/2, math.Min(r.ideal, rng.high-r.length/2))
}

// start and end are where the run begins and ends along the side.
func (r symbolRun) start(side boundarySide, rng sideRange) float64 {
	return r.centre(side, rng) - r.length/2
}
func (r symbolRun) end(side boundarySide, rng sideRange) float64 {
	return r.centre(side, rng) + r.length/2
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
