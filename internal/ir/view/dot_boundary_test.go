package view

import (
	"maps"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var (
	dotPinnedNode = regexp.MustCompile(`(?m)^\s*"(n\d+)" \[.*pos="([-0-9.]+),([-0-9.]+)!".*width=([0-9.]+), height=([0-9.]+)`)
	dotRoutedEdge = regexp.MustCompile(`(?m)^\s*"(n\d+)" -> "(n\d+)" \[.*pos="([^"]+)"`)
)

// pinnedBoxes is the box of every node the DOT text pins, by ID, in the
// rendering's pixels (y down).
func pinnedBoxes(t *testing.T, dot string) map[string]nodeBox {
	t.Helper()
	boxes := map[string]nodeBox{}
	for _, m := range dotPinnedNode.FindAllStringSubmatch(dot, -1) {
		x, y := dotNumber(t, m[2]), -dotNumber(t, m[3])
		width, height := dotNumber(t, m[4])*72, dotNumber(t, m[5])*72
		boxes[m[1]] = nodeBox{low: Point{X: x - width/2, Y: y - height/2}, high: Point{X: x + width/2, Y: y + height/2}}
	}
	return boxes
}

// routeEndsAt is where the routed edges of the DOT text end at each head node,
// in the rendering's pixels.
func routeEndsAt(t *testing.T, dot string) map[string][]Point {
	t.Helper()
	ends := map[string][]Point{}
	for _, m := range dotRoutedEdge.FindAllStringSubmatch(dot, -1) {
		points := strings.Fields(m[3])
		xy := strings.Split(points[len(points)-1], ",")
		ends[m[2]] = append(ends[m[2]], Point{X: dotNumber(t, xy[0]), Y: -dotNumber(t, xy[1])})
	}
	return ends
}

func dotNumber(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("number %q: %v", s, err)
	}
	return v
}

// onBorder reports whether p lies on the box's border, to half a pixel.
func (b nodeBox) onBorder(p Point) bool {
	const tolerance = 0.5
	within := p.X >= b.low.X-tolerance && p.X <= b.high.X+tolerance && p.Y >= b.low.Y-tolerance && p.Y <= b.high.Y+tolerance
	onEdge := math.Abs(p.X-b.low.X) <= tolerance || math.Abs(p.X-b.high.X) <= tolerance ||
		math.Abs(p.Y-b.low.Y) <= tolerance || math.Abs(p.Y-b.high.Y) <= tolerance
	return within && onEdge
}

// overlaps reports whether the two boxes share any area.
func (b nodeBox) overlaps(other nodeBox) bool {
	return b.low.X < other.high.X && other.low.X < b.high.X && b.low.Y < other.high.Y && other.low.Y < b.high.Y
}

// child is the member of the node named parent that is named name.
func child(t *testing.T, rendering *Rendering, parent, name string) *Node {
	t.Helper()
	for _, node := range findNode(t, rendering.Roots, parent).Children {
		if node.Name == name {
			return node
		}
	}
	t.Fatalf("no member %q of %q", name, parent)
	return nil
}

// Two actor parameters of a positioned use case whose routes end at one point
// of its left border are drawn apart along that border, the one whose route
// comes from above on top, each route ending on its own symbol; a use case
// with one parameter keeps it where its route ends.
func TestDOTSpreadsBoundaryParametersAlongTheSide(t *testing.T) {
	rendering := render(t, "boundary-parameters.sysml", "BoundaryParameters::lab")
	dot, err := rendering.DOT()
	if err != nil {
		t.Fatal(err)
	}
	boxes, ends := pinnedBoxes(t, dot), routeEndsAt(t, dot)
	useCase := findNode(t, rendering.Roots, "BoundaryParameters::'Configure Instrument'").Geometry
	operator := child(t, rendering, "BoundaryParameters::'Configure Instrument'", "operator")
	administrator := child(t, rendering, "BoundaryParameters::'Configure Instrument'", "administrator")
	op, admin := boxes[operator.ID], boxes[administrator.ID]
	if op.overlaps(admin) {
		t.Errorf("actor symbols overlap: operator %v, administrator %v", op, admin)
	}
	for name, box := range map[string]nodeBox{"operator": op, "administrator": admin} {
		if math.Abs(box.low.X-useCase.X) > 0.5 {
			t.Errorf("%s symbol does not touch the left border x=%v of the use case: %v", name, useCase.X, box)
		}
	}
	if op.centre().Y >= admin.centre().Y {
		t.Errorf("operator, whose route comes from above, is drawn at y=%v, below administrator at y=%v", op.centre().Y, admin.centre().Y)
	}
	for name, node := range map[string]*Node{"operator": operator, "administrator": administrator} {
		routeEnds := ends[node.ID]
		if len(routeEnds) != 1 {
			t.Fatalf("%s: %d route(s) end at it, want 1", name, len(routeEnds))
		}
		if !boxes[node.ID].onBorder(routeEnds[0]) {
			t.Errorf("%s: route ends at %v, off its symbol %v", name, routeEnds[0], boxes[node.ID])
		}
	}
	if ends[operator.ID][0] == ends[administrator.ID][0] {
		t.Errorf("both routes end at %v", ends[operator.ID][0])
	}

	single := child(t, rendering, "BoundaryParameters::'Acquire Image'", "operator")
	if got, want := ends[single.ID], []Point{{X: 320, Y: 45}}; !slices.Equal(got, want) {
		t.Errorf("single parameter's route ends at %v, want the stated %v", got, want)
	}
	if !boxes[single.ID].onBorder(Point{X: 320, Y: 45}) {
		t.Errorf("single parameter's symbol %v does not meet its route at (320, 45)", boxes[single.ID])
	}
}

// Symbols on the right border spread along it too, in the order their routes
// come from.
func TestDOTSpreadsBoundaryParametersOnTheRight(t *testing.T) {
	rendering := render(t, "boundary-parameters.sysml", "BoundaryParameters::mirroredLab")
	dot, err := rendering.DOT()
	if err != nil {
		t.Fatal(err)
	}
	boxes, ends := pinnedBoxes(t, dot), routeEndsAt(t, dot)
	useCase := findNode(t, rendering.Roots, "BoundaryParameters::'Configure Instrument'").Geometry
	operator := child(t, rendering, "BoundaryParameters::'Configure Instrument'", "operator")
	administrator := child(t, rendering, "BoundaryParameters::'Configure Instrument'", "administrator")
	op, admin := boxes[operator.ID], boxes[administrator.ID]
	if op.overlaps(admin) {
		t.Errorf("actor symbols overlap: operator %v, administrator %v", op, admin)
	}
	right := useCase.X + useCase.Width
	for name, box := range map[string]nodeBox{"operator": op, "administrator": admin} {
		if math.Abs(box.high.X-right) > 0.5 {
			t.Errorf("%s symbol does not touch the right border x=%v of the use case: %v", name, right, box)
		}
	}
	if admin.centre().Y >= op.centre().Y {
		t.Errorf("administrator, whose route comes from above, is drawn at y=%v, below operator at y=%v", admin.centre().Y, op.centre().Y)
	}
	for name, node := range map[string]*Node{"operator": operator, "administrator": administrator} {
		if routeEnds := ends[node.ID]; len(routeEnds) != 1 || !boxes[node.ID].onBorder(routeEnds[0]) {
			t.Errorf("%s: route ends at %v, off its symbol %v", name, routeEnds, boxes[node.ID])
		}
	}
}

// A connection's name a migration made up does not label its edge; a written
// name does.
func TestSynthesizedConnectionNameIsNoEdgeLabel(t *testing.T) {
	rendering := render(t, "boundary-parameters.sysml", "BoundaryParameters::lab")
	labels := edgeLabels(rendering)
	if labels["'Operator to Acquire Image'"] {
		t.Errorf("synthesized name labels an edge; labels: %v", slices.Sorted(maps.Keys(labels)))
	}
	for _, written := range []string{"'Operator to Configure Instrument'", "'Administrator to Configure Instrument'"} {
		if !labels[written] {
			t.Errorf("written name %s labels no edge; labels: %v", written, slices.Sorted(maps.Keys(labels)))
		}
	}
}
