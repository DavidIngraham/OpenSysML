package view

import (
	"math"
	"strings"
	"testing"
)

// portSquare checks that the DOT text pins id as the notation's port square, dotPinSize on a
// side, centred where the Layout's rectangle puts it on its owner's border.
func portSquare(t *testing.T, boxes map[string]nodeBox, id string, centre Point) {
	t.Helper()
	box, ok := boxes[id]
	if !ok {
		t.Fatalf("no pinned box for %s; boxes: %v", id, boxes)
	}
	if width, height := box.size(); math.Abs(width-dotPinSize) > 0.5 || math.Abs(height-dotPinSize) > 0.5 {
		t.Errorf("%s drawn %vx%v, want the port square %dx%d: %v", id, width, height, dotPinSize, dotPinSize, box)
	}
	if got := box.centre(); math.Abs(got.X-centre.X) > 0.5 || math.Abs(got.Y-centre.Y) > 0.5 {
		t.Errorf("%s centred at %v, want %v", id, got, centre)
	}
}

// Ports a part has from its definition, exposed beside the part, whose Layouts are the source
// tool's rectangles sized to the names: each is drawn once, as the pin of the part, the
// notation's square on the side of the part the rectangle's centre lies off, level with it; the
// connector between them joins the two pins, and neither port is a root of its own.
func TestLayoutSizedTypedPortsAreThePartsPins(t *testing.T) {
	rendering := render(t, "layout-ports.sysml", "LayoutPorts::typed")
	if len(rendering.Roots) != 2 {
		t.Fatalf("roots = %v, want the two parts alone", rendering.Roots)
	}
	detector, amplifier := findNode(t, rendering.Roots, "LayoutPorts::Chain::detector"), findNode(t, rendering.Roots, "LayoutPorts::Chain::amplifier")
	if len(detector.Ports) != 1 || len(amplifier.Ports) != 1 {
		t.Fatalf("pins: detector %v, amplifier %v, want one each", detector.Ports, amplifier.Ports)
	}
	out, in := detector.Ports[0].ID, amplifier.Ports[0].ID
	if len(rendering.Edges) != 1 || rendering.Edges[0].FromPort != out || rendering.Edges[0].ToPort != in {
		t.Fatalf("edges = %+v, want one from %s to %s", rendering.Edges, out, in)
	}
	dot, err := rendering.DOT()
	if err != nil {
		t.Fatal(err)
	}
	boxes := pinnedBoxes(t, dot)
	portSquare(t, boxes, out, Point{X: 265, Y: 180 + dotPinSize/2})
	portSquare(t, boxes, in, Point{X: 265, Y: 240 - dotPinSize/2})
	for _, name := range []string{`"tx"`, `"rx"`} {
		if got := strings.Count(dot, "xlabel="+name); got != 1 {
			t.Errorf("DOT names the port %s %d times, want once beside its square:\n%s", name, got, dot)
		}
	}
	if strings.Contains(dot, "width=1.8") {
		t.Errorf("DOT draws a port at the source rectangle's width:\n%s", dot)
	}
}

// Ports a part declares, drawn as its members, whose Layouts are the source tool's rectangles:
// each is the square on the side of its part nearest the rectangle's centre — the right border
// it straddles, the left border, and the bottom border nearest a centre inside the nested part
// — and the connector's route is led on to the squares.
func TestLayoutSizedDeclaredPortsArePinsOnTheirOwnersBorder(t *testing.T) {
	rendering := render(t, "layout-ports.sysml", "LayoutPorts::declared")
	dot, err := rendering.DOT()
	if err != nil {
		t.Fatal(err)
	}
	boxes, ends := pinnedBoxes(t, dot), routeEndsAt(t, dot)
	out := child(t, rendering, "LayoutPorts::Bench::supply", "outlet").ID
	in := child(t, rendering, "LayoutPorts::Bench::instrument", "inlet").ID
	probe := findNode(t, rendering.Roots, "probe").ID
	portSquare(t, boxes, out, Point{X: 240 + dotPinSize/2, Y: 78})
	portSquare(t, boxes, in, Point{X: 400 - dotPinSize/2, Y: 60})
	portSquare(t, boxes, probe, Point{X: 565, Y: 180 + dotPinSize/2})
	clusters := clusterBoxes(t, dot)
	supply, instrument := clusters[findNode(t, rendering.Roots, "LayoutPorts::Bench::supply").ID], clusters[findNode(t, rendering.Roots, "LayoutPorts::Bench::instrument").ID]
	if supply.low.X != 40 || supply.high.X != 240 || instrument.low.X != 400 || instrument.high.X != 700 {
		t.Errorf("owners moved: supply %v, instrument %v", supply, instrument)
	}
	routesEndOnPins(t, boxes, ends, in)
	if routeEnds := ends[in]; len(routeEnds) != 1 {
		t.Fatalf("feed ends %v, want one at the instrument's pin", routeEnds)
	}
}

// A definition's port in a tree rendering follows the same rule: the square on the border the
// rectangle's centre lies on, not the rectangle, though a member nested in the port is left
// undrawn beneath it.
func TestLayoutSizedPortOfADefinitionIsAPinInATree(t *testing.T) {
	rendering := render(t, "layout-ports.sysml", "LayoutPorts::tree")
	dot, err := rendering.DOT()
	if err != nil {
		t.Fatal(err)
	}
	out := child(t, rendering, "LayoutPorts::Detector", "tx").ID
	portSquare(t, pinnedBoxes(t, dot), out, Point{X: 240 + dotPinSize/2, Y: 118})
}
