package combined

import (
	"context"
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/engine"
	"testing"
)

func TestRenderParallelSatisfactionIdentities(t *testing.T) {
	eng, err := New("identity-contract-test")
	if err != nil {
		t.Fatal(err)
	}
	call := func(method string, params any, response any) {
		t.Helper()
		payload, err := json.Marshal(params)
		if err != nil {
			t.Fatal(err)
		}
		result, err := eng.Call(context.Background(), method, payload)
		if err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		if err := json.Unmarshal(result, response); err != nil {
			t.Fatal(err)
		}
	}
	const source = `package Parallel {
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
    }`
	var parsed engine.JParseSourcesResponse
	call("ParseSources", map[string]any{"documents": []map[string]string{{"name": "parallel.sysml", "content": source}}}, &parsed)
	if parsed.ModelHash == "" {
		t.Fatal("no native model identity")
	}
	var exported struct {
		Elements []map[string]any `json:"elements"`
	}
	call("ExportModel", map[string]any{"modelHash": parsed.ModelHash, "projection": "declarations"}, &exported)
	expected := map[string]bool{}
	for _, element := range exported.Elements {
		if element["@type"] == "SatisfyRequirementUsage" {
			expected[element["@id"].(string)] = true
		}
	}
	if len(expected) != 2 {
		t.Fatalf("native declaration export has %d satisfaction assertions, want 2", len(expected))
	}
	for iteration, identity := range []bool{true, false, true} {
		var rendered engine.JRenderViewResponse
		call("RenderView", engine.JRenderViewRequest{ModelHash: parsed.ModelHash, View: "Parallel::requirements", IncludeSemanticIdentity: identity}, &rendered)
		byNode := map[string]string{}
		for _, node := range rendered.Nodes {
			byNode[node.ID] = node.SemanticID
		}
		if len(rendered.Nodes) != 2 {
			t.Fatalf("render %d duplicated endpoint nodes: %+v", iteration, rendered.Nodes)
		}
		seen, count := map[string]bool{}, 0
		for _, edge := range rendered.Edges {
			if edge.Kind != "satisfy" {
				continue
			}
			count++
			if identity {
				if !expected[edge.SemanticID] || seen[edge.SemanticID] || edge.IdentityStatus != "mapped" {
					t.Fatalf("render %d lost assertion identity: %+v", iteration, edge)
				}
				if byNode[edge.From] != "Parallel__vehicle" || byNode[edge.To] != "Parallel__target" {
					t.Fatalf("render %d changed native endpoints: %+v", iteration, edge)
				}
				seen[edge.SemanticID] = true
			} else if edge.SemanticID != "" {
				t.Fatalf("unrequested identity leaked from cached rendering: %+v", edge)
			}
		}
		if count != 2 {
			t.Fatalf("render %d has %d satisfaction edges, want 2", iteration, count)
		}
	}
}
