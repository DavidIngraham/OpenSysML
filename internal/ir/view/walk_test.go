package view

import (
	"strings"
	"testing"
)

// A view's depth is its own depth attribute, else the nearest view definition's;
// 0 is no bound, and the caller's depth overrides them all for that render.
func TestWalkDepthNearestWinsAndZeroIsUnbounded(t *testing.T) {
	r, idx := loadFixtures(t, "depth-tree.sysml")
	cases := []struct {
		view     string
		override int
		deepest  string
	}{
		{"DepthViews::depth1", 0, "wheel"},
		{"DepthViews::depth2", 0, "hub"},
		{"DepthViews::depth3", 0, "bearing"},
		{"DepthViews::unlimited", 0, "grade"},
		{"DepthViews::unlimited", 1, "wheel"},
		{"DepthViews::depth1", 3, "bearing"},
	}
	for _, tc := range cases {
		r.SetDepth(tc.override)
		rendering, err := r.Render(lookup(t, idx, tc.view))
		if err != nil {
			t.Fatalf("render %s: %v", tc.view, err)
		}
		names := nodeNames(rendering.Roots)
		if !names[tc.deepest] {
			t.Errorf("%s with override %d: %s is not drawn; nodes: %v", tc.view, tc.override, tc.deepest, names)
		}
		if tc.deepest != "grade" && names["grade"] {
			t.Errorf("%s with override %d draws grade, below the bound", tc.view, tc.override)
		}
	}
}

// A depth bound is noticed once, counting the elements it kept from the drawing.
func TestWalkDepthIsNoticedOnce(t *testing.T) {
	rendering := render(t, "depth-tree.sysml", "DepthViews::depth1")
	var cuts int
	for _, notice := range rendering.Notices {
		if strings.Contains(notice, "nested deeper than 1 level") {
			cuts++
			if !strings.HasPrefix(notice, "4 elements") {
				t.Errorf("notice counts %q, want the 4 elements below wheel and engine", notice)
			}
		}
	}
	if cuts != 1 {
		t.Errorf("depth noticed %d times, want once: %v", cuts, rendering.Notices)
	}
}

// The view's filter conditions prune the walk: a member a condition rejects is
// not drawn, nor anything under it, while the exposed root is admitted as before.
func TestWalkFiltersPruneContainedMembers(t *testing.T) {
	rendering := render(t, "walk-filter.sysml", "FilteredWalk::structure")
	names := nodeNames(rendering.Roots)
	for _, want := range []string{"FilteredWalk::rover", "chassis", "axle", "radio", "rx"} {
		if !names[want] {
			t.Errorf("%s is not drawn; nodes: %v", want, names)
		}
	}
	for _, hidden := range []string{"mass", "length", "telemetry", "tx"} {
		if names[hidden] {
			t.Errorf("%s is drawn, though the view's filters reject it; nodes: %v", hidden, names)
		}
	}
}

// A connector ending at a pruned element is reported, not drawn, and not
// rerouted to the pruned element's owner.
func TestWalkFilteredConnectorEndIsNoticed(t *testing.T) {
	rendering := render(t, "walk-filter.sysml", "FilteredWalk::links")
	if len(rendering.Edges) != 0 {
		t.Errorf("edges = %+v, want none: telemetry is pruned", rendering.Edges)
	}
	if !strings.Contains(strings.Join(rendering.Notices, "\n"), "which the view's filter conditions exclude") {
		t.Errorf("notices = %v, want the pruned connector end reported", rendering.Notices)
	}
}

// A connector crossing the depth bound ends at the drawn element standing in
// for what lies below it, and the ports of an element at the bound stay pinned.
func TestWalkDepthKeepsPortsAndReroutesCrossingConnectors(t *testing.T) {
	rendering := render(t, "depth-interconnection.sysml", "DepthLinks::links")
	names := nodeNames(rendering.Roots)
	for _, want := range []string{"plant", "status", "controller", "cmd"} {
		if !names[want] {
			t.Errorf("%s is not drawn; nodes: %v", want, names)
		}
	}
	for _, hidden := range []string{"pump", "tank", "out"} {
		if names[hidden] {
			t.Errorf("%s is drawn below the bound; nodes: %v", hidden, names)
		}
	}
	if len(rendering.Edges) != 2 {
		t.Errorf("edges = %d, want the two connectors from controller", len(rendering.Edges))
	}
}

// A collapsed Layout folds an element's children away in every form, keeping
// the collapsed tag; its ports stay, and a connector still ends at them.
func TestCollapsedLayoutFoldsChildrenInEveryForm(t *testing.T) {
	for _, viewName := range []string{"FoldedViews::folded", "FoldedViews::links"} {
		rendering := render(t, "layout-collapsed.sysml", viewName)
		for _, form := range []Form{FormText, FormMermaid, FormDot, FormPlantUML, FormD2} {
			artifact, err := rendering.Write(form)
			if err != nil {
				t.Fatalf("%s as %s: %v", viewName, form, err)
			}
			for _, folded := range []string{"piston", "pump"} {
				if strings.Contains(artifact, folded) {
					t.Errorf("%s as %s draws %s under the collapsed engine:\n%s", viewName, form, folded, artifact)
				}
			}
			if !strings.Contains(artifact, "collapsed") {
				t.Errorf("%s as %s lost the collapsed tag:\n%s", viewName, form, artifact)
			}
		}
		if rendering.Kind == KindInterconnection && len(rendering.Edges) != 1 {
			t.Errorf("%s draws %d edges, want the connector to engine's port", viewName, len(rendering.Edges))
		}
	}
}

// The text form omits the members a Layout's showParts = false hides, as DOT does.
func TestTextOmitsHiddenCompartments(t *testing.T) {
	rendering := render(t, "layout-collapsed.sysml", "FoldedViews::noParts")
	text, err := rendering.Write(FormText)
	if err != nil {
		t.Fatal(err)
	}
	for _, hidden := range []string{"wheel", "engine", "tank"} {
		if strings.Contains(text, hidden) {
			t.Errorf("text lists %s, which showParts = false hides:\n%s", hidden, text)
		}
	}
}
