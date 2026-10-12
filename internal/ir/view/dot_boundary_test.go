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
	dotClusterBB  = regexp.MustCompile(`(?s)subgraph "cluster_(n\d+)" \{.*?bb="([-0-9.]+),([-0-9.]+),([-0-9.]+),([-0-9.]+)"`)
)

// clusterBoxes is the box of every cluster the DOT text gives a `bb`, by ID,
// in the rendering's pixels (y down).
func clusterBoxes(t *testing.T, dot string) map[string]nodeBox {
	t.Helper()
	boxes := map[string]nodeBox{}
	for _, m := range dotClusterBB.FindAllStringSubmatch(dot, -1) {
		boxes[m[1]] = nodeBox{low: Point{X: dotNumber(t, m[2]), Y: -dotNumber(t, m[5])}, high: Point{X: dotNumber(t, m[4]), Y: -dotNumber(t, m[3])}}
	}
	return boxes
}

// symbolBoxes renders the view and returns the DOT text's pinned boxes, cluster
// boxes and route ends with the boxes of the two named members of parent.
func symbolBoxes(t *testing.T, view, parent, first, second string) (rendering *Rendering, boxes, clusters map[string]nodeBox, ends map[string][]Point, a, b *Node) {
	t.Helper()
	rendering = render(t, "boundary-parameters.sysml", view)
	dot, err := rendering.DOT()
	if err != nil {
		t.Fatal(err)
	}
	a, b = child(t, rendering, parent, first), child(t, rendering, parent, second)
	return rendering, pinnedBoxes(t, dot), clusterBoxes(t, dot), routeEndsAt(t, dot), a, b
}

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

// Two actor parameters whose routes reach one corner from different sides are
// drawn apart, each on the side its route comes to, each route ending on its
// own symbol.
func TestDOTSpreadsBoundaryParametersAtACorner(t *testing.T) {
	_, boxes, _, ends, operator, technician := symbolBoxes(t, "BoundaryParameters::corner", "BoundaryParameters::'Load Sample'", "operator", "technician")
	op, tech := boxes[operator.ID], boxes[technician.ID]
	if op.overlaps(tech) {
		t.Errorf("actor symbols overlap: operator %v, technician %v", op, tech)
	}
	if math.Abs(op.low.X-320) > 0.5 {
		t.Errorf("operator, reached from the left, does not touch the left border x=320: %v", op)
	}
	if math.Abs(tech.low.Y-120) > 0.5 {
		t.Errorf("technician, reached from above, does not touch the top border y=120: %v", tech)
	}
	for name, node := range map[string]*Node{"operator": operator, "technician": technician} {
		if routeEnds := ends[node.ID]; len(routeEnds) != 1 || !boxes[node.ID].onBorder(routeEnds[0]) {
			t.Errorf("%s: route ends at %v, off its symbol %v", name, routeEnds, boxes[node.ID])
		}
	}
}

// A use case with a position but no stated size is boxed round its parameters:
// the two the routes place over one another are drawn apart, and the box drawn
// for the use case is refitted round them.
func TestDOTSpreadsBoundaryParametersOfAnUnsizedNode(t *testing.T) {
	rendering, boxes, clusters, _, operator, administrator := symbolBoxes(t, "BoundaryParameters::unsized", "BoundaryParameters::'Configure Instrument'", "operator", "administrator")
	useCase := findNode(t, rendering.Roots, "BoundaryParameters::'Configure Instrument'")
	op, admin := boxes[operator.ID], boxes[administrator.ID]
	if op.overlaps(admin) {
		t.Errorf("actor symbols overlap: operator %v, administrator %v", op, admin)
	}
	cluster, ok := clusters[useCase.ID]
	if !ok {
		t.Fatalf("no bb for the use case %s; clusters: %v", useCase.ID, clusters)
	}
	margin := Point{X: dotClusterMargin, Y: dotClusterMargin}
	want := nodeBox{low: Point{X: 320, Y: 120}, high: Point{X: 320, Y: 120}}.union(op.union(admin))
	want = nodeBox{low: Point{X: want.low.X - margin.X, Y: want.low.Y - margin.Y}, high: Point{X: want.high.X + margin.X, Y: want.high.Y + margin.Y}}
	for _, corner := range [][2]Point{{cluster.low, want.low}, {cluster.high, want.high}} {
		if math.Abs(corner[0].X-corner[1].X) > 0.5 || math.Abs(corner[0].Y-corner[1].Y) > 0.5 {
			t.Fatalf("use case boxed at %v, want the margin round its spread symbols %v", cluster, want)
		}
	}
}

// Nested parts are no boundary symbols: the routes that reach them inside their
// owner leave them where they end, off the owner's border.
func TestDOTLeavesInteriorMembersWhereRoutesPlaceThem(t *testing.T) {
	_, boxes, _, _, detector, filament := symbolBoxes(t, "BoundaryParameters::interior", "BoundaryParameters::Instrument", "detector", "filament")
	owner := nodeBox{low: Point{X: 320, Y: 120}, high: Point{X: 480, Y: 220}}
	for name, box := range map[string]nodeBox{"detector": boxes[detector.ID], "filament": boxes[filament.ID]} {
		inside := box.low.X > owner.low.X+1 && box.high.X < owner.high.X-1 && box.low.Y > owner.low.Y+1 && box.high.Y < owner.high.Y-1
		if !inside {
			t.Errorf("%s %v was moved to the border of %v", name, box, owner)
		}
		if !box.holds(Point{X: 370, Y: 170}) && !box.onBorder(Point{X: 370, Y: 170}) {
			t.Errorf("%s %v does not meet its route's end (370, 170)", name, box)
		}
	}
}

// Two symbols from above, together wider than the top side leaves beyond the
// symbol from the left, run from that symbol past the far corner rather than
// back over it.
func TestDOTRunsCrowdedCornerSymbolsPastTheFarCorner(t *testing.T) {
	rendering, boxes, _, ends, technician, administrator := symbolBoxes(t, "BoundaryParameters::crowdedCorner", "BoundaryParameters::'Load Sample'", "technician", "administrator")
	operator := child(t, rendering, "BoundaryParameters::'Load Sample'", "operator")
	op, tech, admin := boxes[operator.ID], boxes[technician.ID], boxes[administrator.ID]
	for _, pair := range [][2]nodeBox{{op, tech}, {op, admin}, {tech, admin}} {
		if pair[0].overlaps(pair[1]) {
			t.Errorf("symbols overlap: %v and %v", pair[0], pair[1])
		}
	}
	if math.Abs(op.low.X-320) > 0.5 {
		t.Errorf("operator does not touch the left border x=320: %v", op)
	}
	for name, box := range map[string]nodeBox{"technician": tech, "administrator": admin} {
		if math.Abs(box.low.Y-120) > 0.5 {
			t.Errorf("%s does not touch the top border y=120: %v", name, box)
		}
		if box.low.X < op.high.X+dotBoundaryGap-0.5 {
			t.Errorf("%s %v is not clear of the operator %v", name, box, op)
		}
	}
	if tech.centre().X >= admin.centre().X {
		t.Errorf("technician, whose route comes from the left of the administrator's, is drawn right of it: %v, %v", tech, admin)
	}
	for name, node := range map[string]*Node{"operator": operator, "technician": technician, "administrator": administrator} {
		if routeEnds := ends[node.ID]; len(routeEnds) != 1 || !boxes[node.ID].onBorder(routeEnds[0]) {
			t.Errorf("%s: route ends at %v, off its symbol %v", name, routeEnds, boxes[node.ID])
		}
	}
}

// Two symbols from above that fit between neither corner's symbol are centred
// between the two, the unavoidable overlap split evenly, each still touching
// the top and each route still ending on its own symbol.
func TestDOTSplitsTheOverlapOfACrowdedSideBetweenItsCorners(t *testing.T) {
	rendering, boxes, _, ends, technician, administrator := symbolBoxes(t, "BoundaryParameters::crowdedCorners", "BoundaryParameters::'Load Sample'", "technician", "administrator")
	operator := child(t, rendering, "BoundaryParameters::'Load Sample'", "operator")
	supervisor := child(t, rendering, "BoundaryParameters::'Load Sample'", "supervisor")
	op, sup, tech, admin := boxes[operator.ID], boxes[supervisor.ID], boxes[technician.ID], boxes[administrator.ID]
	if math.Abs(op.low.X-320) > 0.5 || math.Abs(sup.high.X-480) > 0.5 {
		t.Fatalf("corner symbols off their sides: operator %v, supervisor %v", op, sup)
	}
	if tech.overlaps(admin) {
		t.Errorf("symbols from above overlap: %v and %v", tech, admin)
	}
	low, high := op.high.X+dotBoundaryGap, sup.low.X-dotBoundaryGap
	if tech.high.X-tech.low.X+dotBoundaryGap+admin.high.X-admin.low.X <= high-low {
		t.Fatalf("the symbols from above fit between the corners' (%v..%v): %v, %v", low, high, tech, admin)
	}
	if math.Abs((tech.low.X-low)-(high-admin.high.X)) > 0.5 {
		t.Errorf("overlap not split between the corners (%v..%v): technician %v, administrator %v", low, high, tech, admin)
	}
	for name, box := range map[string]nodeBox{"technician": tech, "administrator": admin} {
		if math.Abs(box.low.Y-120) > 0.5 {
			t.Errorf("%s does not touch the top border y=120: %v", name, box)
		}
	}
	for name, node := range map[string]*Node{"operator": operator, "supervisor": supervisor, "technician": technician, "administrator": administrator} {
		if routeEnds := ends[node.ID]; len(routeEnds) != 1 || !boxes[node.ID].onBorder(routeEnds[0]) {
			t.Errorf("%s: route ends at %v, off its symbol %v", name, routeEnds, boxes[node.ID])
		}
	}
}
