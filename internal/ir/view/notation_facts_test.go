package view

import (
	"strings"
	"testing"
)

func TestNativeNotationFactsAndCloneIsolation(t *testing.T) {
	seq := render(t, "sequence.sysml", "SequenceViews::pubSubView")
	if seq.SequenceOrdering != "partial-order-linearization" || len(seq.SequenceOrders) != 3 {
		t.Fatalf("explicit sequence ordering missing: %s %d", seq.SequenceOrdering, len(seq.SequenceOrders))
	}
	parents := map[string]string{}
	for _, node := range seq.Data().Nodes {
		parents[node.ID] = node.Parent
	}
	messages := 0
	for _, edge := range seq.Edges {
		if edge.Kind == EdgeSuccession {
			continue
		}
		messages++
		n := edge.Notation
		if n == nil || n.Production != "message" || n.SourceEvent == nil || n.TargetEvent == nil || n.SourceEvent.Kind != "event occurrence" || n.SourceEvent.LifelineID != parents[edge.From] || n.TargetEvent.LifelineID != parents[edge.To] || !n.SourceEvent.Origin.Located() {
			t.Fatalf("native message facts lost: %+v", edge)
		}
	}
	if messages != 3 || len(seq.Edges) != 6 {
		t.Fatalf("messages=%d edges=%d", messages, len(seq.Edges))
	}
	copy := seq.Clone()
	copy.Edges[0].Notation.SourceEvent.Kind = "changed"
	copy.SequenceOrders[0].SourceEvent.Kind = "changed"
	if seq.Edges[0].Notation.SourceEvent.Kind == "changed" || seq.SequenceOrders[0].SourceEvent.Kind == "changed" {
		t.Fatal("clone aliases sequence semantic facts")
	}
	states := render(t, "state-do.sysml", "InstrumentViews::peas")
	count := 0
	var walk func(*Node)
	walk = func(n *Node) {
		if n.State != nil {
			for _, b := range n.State.Behaviors {
				count++
				if !b.Origin.Located() || b.Kind == "" || b.Role == "" {
					t.Fatalf("unlocated state behavior: %+v", b)
				}
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range states.Roots {
		walk(n)
	}
	if count != 7 {
		t.Fatalf("state behaviors=%d want7", count)
	}
	general := render(t, "general.sysml", "GeneralViews::requirementView")
	derive, refine := false, false
	for _, e := range general.Edges {
		if e.Kind == EdgeDerive {
			derive = e.Notation != nil && e.Notation.Production == "derivation" && e.Notation.SourceRole == "derived" && e.Notation.TargetRole == "original"
		}
		if e.Kind == EdgeRefine {
			refine = e.Notation != nil && e.Notation.Production == "refinement" && e.Notation.SourceRole == "client" && strings.Contains(e.Notation.Classification, "Refinement")
		}
	}
	if !derive || !refine {
		t.Fatalf("derive=%v refine=%v", derive, refine)
	}
}
