package combined

import (
	"context"
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/engine"
	"testing"
)

func TestRenderIdentityAndNotationFields(t *testing.T) {
	eng, err := New("identity-contract-test")
	if err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]any{"documents": []map[string]string{{"name": "identity.sysml", "content": `package Demo { part def Thing; view diagram { expose Thing; } }`}}})
	parsed, err := eng.Call(context.Background(), "ParseSources", params)
	if err != nil {
		t.Fatal(err)
	}
	var model engine.JParseSourcesResponse
	if err := json.Unmarshal(parsed, &model); err != nil {
		t.Fatal(err)
	}
	params, _ = json.Marshal(engine.JRenderViewRequest{ModelHash: model.ModelHash, View: "Demo::diagram"})
	if _, err := eng.Call(context.Background(), "RenderView", params); err != nil {
		t.Fatalf("ordinary rendering unavailable: %v", err)
	}
	params, _ = json.Marshal(engine.JRenderViewRequest{ModelHash: model.ModelHash, View: "Demo::diagram", IncludeSemanticIdentity: true})
	response, err := eng.Call(context.Background(), "RenderView", params)
	if err != nil {
		t.Fatal(err)
	}
	var rendered engine.JRenderViewResponse
	if err := json.Unmarshal(response, &rendered); err != nil {
		t.Fatal(err)
	}
	if len(rendered.Nodes) != 1 || rendered.Nodes[0].SemanticID != "Demo__Thing" || rendered.Nodes[0].QualifiedName != "Demo::Thing" {
		t.Fatalf("identity lost: %s", response)
	}

}

func TestLanguageControlNodesRemainSynthetic(t *testing.T) {
	eng, err := New("identity-contract-test")
	if err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]any{"documents": []map[string]string{{"name": "flow.sysml", "content": `package P { action flow { action a; first start then a; then done; } }`}}})
	body, err := eng.Call(context.Background(), "ParseSources", params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed engine.JParseSourcesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}
	params, _ = json.Marshal(engine.JRenderViewRequest{ModelHash: parsed.ModelHash, View: "#action:P::flow", IncludeSemanticIdentity: true})
	body, err = eng.Call(context.Background(), "RenderView", params)
	if err != nil {
		t.Fatal(err)
	}
	var drawing engine.JRenderViewResponse
	if err := json.Unmarshal(body, &drawing); err != nil {
		t.Fatal(err)
	}
	done, action := false, false
	for _, node := range drawing.Nodes {
		if node.Name == "done" {
			done = true
			if node.IdentityStatus != "synthetic" || node.SemanticID != "" {
				t.Fatalf("completion acquired identity: %+v", node)
			}
		}
		if node.Name == "a" {
			action = node.IdentityStatus == "mapped" && node.SemanticID == "P__flow__a"
		}
	}
	if !done || !action {
		t.Fatalf("control or declared action missing: %s", body)
	}
}
