package view

import (
	"cmp"
	"math"
	"slices"
)

// dotBoundaryGap is the space, in pixels, between the pins of a node spread
// along one of its sides.
const dotBoundaryGap = 6

// boundarySide is a side of a boxed node, where the pins its routes place sit.
type boundarySide int

const (
	leftSide boundarySide = iota
	rightSide
	topSide
	bottomSide
)

// across reports whether the side runs across the page, so that the pins on it
// spread along x rather than y.
func (s boundarySide) across() bool { return s == topSide || s == bottomSide }

// boundarySymbol is a pin of a boxed node that its routes place: the box the
// routes gave it, where they meet the node and the mean of where they come from.
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

// through is the coordinate of p through side, from the border outward.
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

// pinSide is the side of b a pin sits on for a route ending at end on its way from next: the side
// end lies off — past a corner, the side it lies farther off — and for an end on or inside the
// border the nearest side, the one facing next when two are as near.
func (b nodeBox) pinSide(end, next Point) boundarySide {
	dx := end.X - math.Min(math.Max(end.X, b.low.X), b.high.X)
	dy := end.Y - math.Min(math.Max(end.Y, b.low.Y), b.high.Y)
	switch {
	case dx == 0 && dy == 0:
		return b.nearestSide(end, Point{X: next.X - end.X, Y: next.Y - end.Y})
	case math.Abs(dx) >= math.Abs(dy) && dx < 0:
		return leftSide
	case math.Abs(dx) >= math.Abs(dy):
		return rightSide
	case dy < 0:
		return topSide
	}
	return bottomSide
}

// nearestSide is the side of b nearest p; of two as near, the one facing dir most squarely.
func (b nodeBox) nearestSide(p, dir Point) boundarySide {
	distances := [4]float64{leftSide: p.X - b.low.X, rightSide: b.high.X - p.X, topSide: p.Y - b.low.Y, bottomSide: b.high.Y - p.Y}
	facing := [4]float64{leftSide: -dir.X, rightSide: dir.X, topSide: -dir.Y, bottomSide: dir.Y}
	nearest := leftSide
	for side := rightSide; side <= bottomSide; side++ {
		if distances[side] < distances[nearest] || distances[side] == distances[nearest] && facing[side] > facing[nearest] {
			nearest = side
		}
	}
	return nearest
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

// spreadPins moves apart, along the side they sit on, the pins of a node the routes place over
// one another, a side's pins keeping out of a corner the pins of the sides before it hold.
func (w *dotWriter) spreadPins(box nodeBox, sides [4][]boundarySymbol) {
	var placed []nodeBox
	for side := leftSide; side <= bottomSide; side++ {
		if len(sides[side]) == 0 {
			continue
		}
		for id, pin := range spreadAlong(box, side, sides[side], placed) {
			w.pins[id] = pin
		}
		for _, symbol := range sides[side] {
			placed = append(placed, w.pins[symbol.id])
		}
	}
}

// symbolRun is a run of pins laid a gap apart along a side, centred on the mean of where
// their routes meet it; a run of one pin lies where the routes place it.
type symbolRun struct {
	symbols []boundarySymbol
	length  float64
	ideal   float64
}

// sideRange is the stretch of a side its pins may take: the side, less the
// corners pins of the sides spread before it reach into.
type sideRange struct {
	low, high       float64
	lowCut, highCut bool
}

// freeRange is the range along side left to pins of the given depth by the obstacles in
// their lane: an obstacle nearer the side's start cuts the range there, else at its end.
func (b nodeBox) freeRange(side boundarySide, depth float64, obstacles []nodeBox) sideRange {
	r := sideRange{low: side.along(b.low), high: side.along(b.high)}
	middle := (r.low + r.high) / 2
	laneLow, laneHigh := b.lane(side, depth)
	for _, o := range obstacles {
		if side.through(o.high) <= laneLow || side.through(o.low) >= laneHigh {
			continue
		}
		start, end := side.along(o.low), side.along(o.high)
		if (start+end)/2 < middle {
			r.low, r.lowCut = math.Max(r.low, end+dotBoundaryGap), true
		} else {
			r.high, r.highCut = math.Min(r.high, start-dotBoundaryGap), true
		}
	}
	return r
}

// lane is the band outside side, from its border outward by depth, that pins
// of that depth touching the border occupy.
func (b nodeBox) lane(side boundarySide, depth float64) (low, high float64) {
	switch side {
	case leftSide:
		return b.low.X - depth, b.low.X
	case rightSide:
		return b.high.X, b.high.X + depth
	case topSide:
		return b.low.Y - depth, b.low.Y
	default:
		return b.high.Y, b.high.Y + depth
	}
}

// spreadAlong pools the pins on side of box into runs that merge while they overlap, each run
// centred on its pins' mean place and kept within the side's free range when it fits; it
// returns the box of every pin moved, by ID.
func spreadAlong(box nodeBox, side boundarySide, symbols []boundarySymbol, obstacles []nodeBox) map[string]nodeBox {
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
	moved := map[string]nodeBox{}
	for _, run := range runs {
		if len(run.symbols) < 2 && run.centre(side, rng) == side.along(run.symbols[0].box.centre()) {
			continue
		}
		at := run.start(side, rng)
		for _, symbol := range run.symbols {
			length := side.length(symbol.box)
			moved[symbol.id] = box.pinOn(side, symbol.box, at+length/2)
			at += length + dotBoundaryGap
		}
	}
	return moved
}

// centre is where the run is centred along side: on its ideal within rng when it fits; else from
// a cut-off corner past the far end, or at rng's middle when both corners or none are cut off.
func (r symbolRun) centre(side boundarySide, rng sideRange) float64 {
	if r.length > rng.high-rng.low && (len(r.symbols) > 1 || rng.lowCut && rng.highCut) {
		switch {
		case rng.lowCut && !rng.highCut:
			return rng.low + r.length/2
		case rng.highCut && !rng.lowCut:
			return rng.high - r.length/2
		}
		return (rng.low + rng.high) / 2
	}
	if len(r.symbols) < 2 {
		c := side.along(r.symbols[0].box.centre())
		if rng.lowCut {
			c = math.Max(c, rng.low+r.length/2)
		}
		if rng.highCut {
			c = math.Min(c, rng.high-r.length/2)
		}
		return c
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

// pinOn is the box of a pin of the given box's size on side of b, touching the
// border from outside, centred at `at` along it, to the half pixel.
func (b nodeBox) pinOn(side boundarySide, pin nodeBox, at float64) nodeBox {
	width, height := pin.size()
	var c Point
	switch side {
	case leftSide:
		c = Point{X: b.low.X - width/2, Y: at}
	case rightSide:
		c = Point{X: b.high.X + width/2, Y: at}
	case topSide:
		c = Point{X: at, Y: b.low.Y - height/2}
	case bottomSide:
		c = Point{X: at, Y: b.high.Y + height/2}
	}
	c = Point{X: halfPixel(c.X), Y: halfPixel(c.Y)}
	return nodeBox{low: Point{X: c.X - width/2, Y: c.Y - height/2}, high: Point{X: c.X + width/2, Y: c.Y + height/2}, stated: true}
}
