package view

import (
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
)

func TestGeneralGraphDeduplicatesBySourceOrigin(t *testing.T) {
	first := Origin{Doc: "parallel.sysml", Span: source.Span{Offset: 10, Len: 20}}
	second := Origin{Doc: "parallel.sysml", Span: source.Span{Offset: 40, Len: 20}}
	for _, kind := range []EdgeKind{EdgeSatisfy, EdgeDependency, EdgeAllocate, EdgeImport} {
		t.Run(kind.String(), func(t *testing.T) {
			output := &Rendering{}
			graph := &generalGraph{out: output, edges: map[generalEdgeKey]bool{}}
			a := Edge{From: "part", To: "requirement", Kind: kind, Label: kind.String(), Origin: first}
			b := a
			b.Origin = second
			graph.add(a)
			graph.add(b)
			graph.add(a)
			graph.add(b)
			sameDeclaration := a
			sameDeclaration.Origin.Name = source.Span{Offset: 12, Len: 3}
			graph.add(sameDeclaration)
			if len(output.Edges) != 2 {
				t.Fatalf("distinct source declarations became %d edges, want 2", len(output.Edges))
			}
			if output.Edges[0].Origin != first || output.Edges[1].Origin != second {
				t.Fatalf("source origins or insertion order changed: %+v", output.Edges)
			}
		})
	}
}

func TestGeneralViewParallelSatisfactionOrigins(t *testing.T) {
	rendering := renderSource(t, "Parallel::requirements", `package Parallel {
        private import StandardViewDefinitions::*;
        requirement target;
        part vehicle;
        satisfy target by vehicle;
        satisfy target by vehicle;
        view requirements : GeneralView {
            @OpenSysMLRenderings::RequirementContext { relationshipDepth = 2; }
            filter @SysML::RequirementUsage;
            expose target;
            expose target;
        }
    }`)
	origins := map[Origin]bool{}
	for _, edge := range rendering.Edges {
		if edge.Kind != EdgeSatisfy {
			continue
		}
		if !edge.Origin.Located() || origins[edge.Origin] {
			t.Fatalf("missing or repeated satisfaction origin: %+v", edge)
		}
		origins[edge.Origin] = true
	}
	if len(origins) != 2 {
		t.Fatalf("got %d satisfaction declarations, want 2: %+v", len(origins), rendering.Edges)
	}
	if len(rendering.Roots) != 2 {
		t.Fatalf("parallel edges duplicated their shared endpoint nodes: %+v", rendering.Roots)
	}
}

func TestGeneralGraphUnlocatedEdgesRemainDeduplicated(t *testing.T) {
	output := &Rendering{}
	graph := &generalGraph{out: output, edges: map[generalEdgeKey]bool{}}
	edge := Edge{From: "part", To: "requirement", Kind: EdgeSatisfy}
	graph.add(edge)
	edge.Origin = Origin{Doc: "synthetic.sysml", Name: source.Span{Offset: 1, Len: 3}}
	graph.add(edge)
	if len(output.Edges) != 1 {
		t.Fatalf("unlocated artifacts acquired synthetic relationship identity: %+v", output.Edges)
	}
}
