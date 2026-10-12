package view

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// An association end's name — a composite part's on the composition to its
// type — is the name of the edge's end at the type, not the edge's own text,
// and every form sets it at that end where its grammar has such a place: DOT
// as `headlabel`, PlantUML in quotes before the end's alias, D2 as the target
// arrowhead's label. A flowchart and the text form have none and fold it into
// the edge's text.
func TestEndNamesAreSetAtTheirEnd(t *testing.T) {
	rendering := render(t, "tree-edges.sysml", "FleetViews::structure")
	var wheels *Edge
	for i, edge := range rendering.Edges {
		if edge.Kind != EdgeComposition && edge.Kind != EdgeReference {
			continue
		}
		if edge.Label != "" || edge.FromLabel != "" {
			t.Errorf("edge %s -> %s carries %q as its own text and %q at its tail; a usage's name belongs at the type's end", edge.From, edge.To, edge.Label, edge.FromLabel)
		}
		if edge.ToLabel == "wheels[4]" {
			wheels = &rendering.Edges[i]
		}
	}
	if wheels == nil {
		t.Fatalf("no composition ends in wheels[4]; edges: %+v", rendering.Edges)
	}
	for form, want := range map[Form][]string{
		FormDot:      {`"n0" -> "n16" [headlabel="wheels[4]", dir=back, arrowtail=diamond];`},
		FormPlantUML: {`n0 *-- "wheels[4]" n16`},
		FormD2:       {`n0 <- n16: { class: composition; target-arrowhead: { label: "wheels[4]" } }`},
		FormMermaid:  {`◆ wheels[4]`},
		FormText:     {`Fleet::Vehicle *-- Fleet::Wheel: wheels[4]`},
	} {
		out, err := rendering.Write(form)
		if err != nil {
			t.Fatalf("write %s: %v", form, err)
		}
		for _, line := range want {
			if !strings.Contains(out, line) {
				t.Errorf("%s lacks %q:\n%s", form, line, out)
			}
		}
		if form == FormDot && (strings.Contains(out, ` label="wheels[4]"`) || strings.Contains(out, `[label="wheels[4]"`)) {
			t.Errorf("DOT sets the end's name along the edge:\n%s", out)
		}
	}
	data := rendering.Data()
	for _, edge := range data.Edges {
		if edge.ToLabel == "wheels[4]" && edge.Text() != "wheels[4]" {
			t.Errorf("Text() = %q, want the end's name", edge.Text())
		}
	}
	if got := (EdgeData{FromLabel: "a", Label: "own", ToLabel: "b"}).Text(); got != "a own b" {
		t.Errorf("Text() = %q, want the ends about the edge's own text", got)
	}
}

var (
	dotHeadLabel = regexp.MustCompile(`headlabel="([^"]*)"`)
	dotHeadLP    = regexp.MustCompile(`head_lp="([-0-9.]+),([-0-9.]+)"`)
)

// A routed composition's end name sits beside the last segment of its route
// before the type, nearer that end than the whole's, clear of the line; where
// several routes share a trunk — leave the whole down one segment, run along
// one bus and drop to their parts — and two drop to the same box, no two
// labels overlap, and none sits on a box.
func TestRoutedEndNamesSitAtTheirEndClearOfEachOther(t *testing.T) {
	rendering := render(t, "tree-trunk.sysml", "MicroscopeViews::structure")
	dot, err := rendering.DOT()
	if err != nil {
		t.Fatalf("DOT: %v", err)
	}
	if !strings.Contains(dot, "// layout: neato -n2\n") {
		t.Fatalf("the routed view is not drawn by neato -n2:\n%s", dot)
	}
	edges := map[string]Edge{}
	for _, edge := range rendering.Edges {
		if edge.Kind == EdgeComposition {
			edges[edge.ToLabel] = edge
		}
	}
	if len(edges) != 6 {
		t.Fatalf("compositions = %d, want the six parts' edges", len(edges))
	}
	skin := skinOf(StylePilot)
	var boxes []nodeBox
	for _, line := range strings.Split(dot, "\n") {
		name := dotHeadLabel.FindStringSubmatch(line)
		if name == nil {
			continue
		}
		edge, ok := edges[name[1]]
		if !ok {
			t.Errorf("headlabel %q names no composition end", name[1])
			continue
		}
		lp := dotHeadLP.FindStringSubmatch(line)
		if lp == nil {
			t.Errorf("%s has no head_lp:\n%s", name[1], line)
			continue
		}
		x, _ := strconv.ParseFloat(lp[1], 64)
		y, _ := strconv.ParseFloat(lp[2], 64)
		at := Point{X: x, Y: rendering.Canvas.Height - y}
		width, height := dotTextExtent([]string{name[1]}, skin.edgePts)
		boxes = append(boxes, labelBox(at, width, height))
		n := len(edge.Route)
		from, to := edge.Route[n-2], edge.Route[n-1]
		if across, along := segmentDistance(at, from, to); across > width/2+height/2+labelGap+1 || along > 0 {
			t.Errorf("%s is centred at %v, %.1f off the end segment %v-%v and %.1f past it", name[1], at, across, from, to, along)
		}
		if math.Hypot(at.X-to.X, at.Y-to.Y) >= math.Hypot(at.X-edge.Route[0].X, at.Y-edge.Route[0].Y) {
			t.Errorf("%s is centred at %v, nearer the whole's end %v than the type's %v", name[1], at, edge.Route[0], to)
		}
	}
	if len(boxes) != 6 {
		t.Fatalf("positioned end names = %d, want 6:\n%s", len(boxes), dot)
	}
	for i, a := range boxes {
		for _, b := range boxes[:i] {
			if a.overlap(b) > 0 {
				t.Errorf("labels %v and %v overlap", a, b)
			}
		}
	}
	var nodes []nodeBox
	var walk func(nodes []*Node)
	walk = func(list []*Node) {
		for _, node := range list {
			if g := node.Geometry; g != nil && g.HasSize {
				nodes = append(nodes, nodeBox{low: Point{X: g.X, Y: g.Y}, high: Point{X: g.X + g.Width, Y: g.Y + g.Height}})
			}
			walk(node.Children)
		}
	}
	walk(rendering.Roots)
	for _, label := range boxes {
		for _, node := range nodes {
			if label.overlap(node) > 0 {
				t.Errorf("label %v sits on the box %v", label, node)
			}
		}
	}
}

// segmentDistance is how far p lies from the line through the segment, and
// how far past the segment's extent along it, 0 when within it.
func segmentDistance(p, from, to Point) (across, along float64) {
	dx, dy := to.X-from.X, to.Y-from.Y
	length := math.Hypot(dx, dy)
	ux, uy := dx/length, dy/length
	px, py := p.X-from.X, p.Y-from.Y
	across = math.Abs(px*uy - py*ux)
	t := px*ux + py*uy
	switch {
	case t < 0:
		along = -t
	case t > length:
		along = t - length
	}
	return across, along
}

// Labels that would overlap are placed in order: the first takes the spot
// nearest its end, the next the other side of the line, then a step farther
// along; a label never moves one placed before it, so the placement is the
// same on every run.
func TestLabelPlacerKeepsLabelsApart(t *testing.T) {
	w := &dotWriter{skin: skinOf(StylePilot)}
	route := []Point{{X: 100, Y: 0}, {X: 100, Y: 300}}
	width, height := dotTextExtent([]string{"label"}, w.skin.edgePts)
	first := w.placeLabel(route, "label", labelTo)
	second := w.placeLabel(route, "label", labelTo)
	third := w.placeLabel(route, "label", labelTo)
	if first.X <= 100 || second.X >= 100 {
		t.Errorf("first at %v, second at %v; want them on either side of the line", first, second)
	}
	if first.Y != second.Y || third.Y >= first.Y {
		t.Errorf("first at %v, second at %v, third at %v; want the third a step farther from the end", first, second, third)
	}
	placed := []Point{first, second, third}
	for i, a := range placed {
		if a.Y > 300-labelEndClearance-height/2 {
			t.Errorf("label %d at %v sits on the arrowhead at the end", i, a)
		}
		for _, b := range placed[:i] {
			if labelBox(a, width, height).overlap(labelBox(b, width, height)) > 0 {
				t.Errorf("labels at %v and %v overlap", a, b)
			}
		}
	}
	w.placer = nil
	if again := w.placeLabel(route, "label", labelTo); again != first {
		t.Errorf("placed again at %v, want %v", again, first)
	}
}

// An end name whose last segment is hemmed in — a box on either side of it —
// steps back to the segment before rather than sit on a box; a line across
// the spot beside the last segment is borne, the name staying at its end.
func TestEndNamesLeaveAHemmedInLastSegment(t *testing.T) {
	w := &dotWriter{skin: skinOf(StylePilot), clusters: map[string]bool{}}
	route := []Point{{X: 0, Y: 0}, {X: 200, Y: 0}, {X: 200, Y: 100}}
	width, _ := dotTextExtent([]string{"hemmed"}, w.skin.edgePts)
	w.boxes = map[string]nodeBox{
		"left":  {low: Point{X: 200 - width - 20, Y: 10}, high: Point{X: 196, Y: 100}},
		"right": {low: Point{X: 204, Y: 10}, high: Point{X: 200 + width + 20, Y: 100}},
	}
	at := w.placeLabel(route, "hemmed", labelTo)
	if at.Y > 0 || at.X >= 200 {
		t.Errorf("hemmed-in end name at %v, want it beside the segment before the last", at)
	}
	w.placer, w.boxes = nil, map[string]nodeBox{}
	w.edges = []Edge{{Route: []Point{{X: 220, Y: 50}, {X: 300, Y: 50}}}}
	at = w.placeLabel(route, "crossed", labelTo)
	if at.X <= 200 || at.Y < 50 {
		t.Errorf("end name at %v, want it beside the last segment though a line crosses it", at)
	}
}
