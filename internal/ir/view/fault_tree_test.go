package view

import (
	"errors"
	"strings"
	"testing"
)

func TestFaultTreeTypesSharedCausesAndRenderings(t *testing.T) {
	drawing := render(t, "fault-tree.sysml", "PowerSupply::tree")
	if drawing.Kind != KindFaultTree || len(drawing.Roots) != 6 || len(drawing.Edges) != 6 {
		t.Fatalf("kind=%s nodes=%d edges=%d notices=%v", drawing.Kind, len(drawing.Roots), len(drawing.Edges), drawing.Notices)
	}
	common := findNode(t, drawing.Roots, "PowerSupply::common")
	count := 0
	for _, edge := range drawing.Edges {
		if edge.From == common.ID {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("shared cause has %d outputs", count)
	}
	if findNode(t, drawing.Roots, "PowerSupply::left").Kind != "OR" {
		t.Fatal("gate specialization lost")
	}
	text, err := drawing.Write(FormMermaid)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(text, "data:image/svg+xml;base64,") != 3 || !strings.Contains(text, "flowchart BT") {
		t.Fatalf("missing symbols: %s", text)
	}
	if strings.Contains(text, "subgraph") || strings.Contains(text, "description") {
		t.Fatal("attribute compartments leaked")
	}
	for _, form := range []Form{FormDot, FormPlantUML, FormD2} {
		if _, err := drawing.Write(form); !errors.Is(err, ErrWrongForm) {
			t.Fatalf("unsupported %s: %v", form, err)
		}
	}
}

func TestFaultTreePartialAndCycle(t *testing.T) {
	partial := render(t, "fault-tree.sysml", "PowerSupply::partial")
	if len(partial.Roots) != 1 || len(partial.Edges) != 0 || len(partial.Notices) != 1 {
		t.Fatalf("partial: %+v", partial)
	}
	renderer, index := loadFixture(t, "fault-tree.sysml")
	if _, err := renderer.Render(lookup(t, index, "InvalidTree::tree")); !errors.Is(err, ErrInvalidFaultTree) {
		t.Fatalf("cycle: %v", err)
	}
}

func TestFaultTreeIntegrity(t *testing.T) {
	renderer, index := loadFixture(t, "fault-tree.sysml")
	for _, name := range []string{"ExtraCause::tree", "Inherited::tree", "Ambiguous::tree"} {
		t.Run(name, func(t *testing.T) {
			if _, err := renderer.Render(lookup(t, index, name)); !errors.Is(err, ErrInvalidFaultTree) {
				t.Fatalf("expected integrity error, got %v", err)
			}
		})
	}
	partial := render(t, "fault-tree.sysml", "PartialGate::tree")
	text, err := partial.Write(FormMermaid)
	if err != nil || !strings.Contains(text, "Loss of supply<br/>[partial view]") {
		t.Fatalf("partial label missing: %s %v", text, err)
	}
	nested := render(t, "fault-tree.sysml", "Nested::tree")
	if len(nested.Roots) != 3 || len(nested.Edges) != 1 {
		t.Fatalf("nested nodes=%d edges=%d", len(nested.Roots), len(nested.Edges))
	}
}

func TestFaultTreeEndRoles(t *testing.T) {
	drawing := render(t, "fault-tree.sysml", "Reordered::tree")
	if len(drawing.Edges) != 1 {
		t.Fatalf("edges=%v", drawing.Edges)
	}
	cause := findNode(t, drawing.Roots, "Reordered::a")
	effect := findNode(t, drawing.Roots, "Reordered::b")
	if drawing.Edges[0].From != cause.ID || drawing.Edges[0].To != effect.ID {
		t.Fatalf("reversed edge: %v", drawing.Edges)
	}
}

func TestFaultTreeReferenceIdentity(t *testing.T) {
	drawing := render(t, "fault-tree.sysml", "ReferenceCause::tree")
	if len(drawing.Roots) != 2 || len(drawing.Edges) != 1 || len(drawing.Notices) != 0 {
		t.Fatalf("reference tree: %+v", drawing)
	}
	if drawing.Edges[0].From != findNode(t, drawing.Roots, "ReferenceCause::a").ID {
		t.Fatal("reference lost causal identity")
	}
}
