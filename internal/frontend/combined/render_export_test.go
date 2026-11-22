package combined

import (
	"context"
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/engine"
	"testing"
)

func TestRenderingAndExportShareConversionIdentities(t *testing.T) {
	server, err := New("view-test")
	if err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]any{"documents": []map[string]string{
		{"name": "model.sysml", "content": `package Demo { part def Device; part device : Device; }`},
		{"name": "views.sysml", "content": `package Views { view device { expose Demo::device; } }`},
	}})
	parsed, err := server.Call(context.Background(), "ParseSources", params)
	if err != nil {
		t.Fatal(err)
	}
	var model struct {
		ModelHash string `json:"modelHash"`
	}
	if err := json.Unmarshal(parsed, &model); err != nil {
		t.Fatal(err)
	}
	params, _ = json.Marshal(map[string]any{"modelHash": model.ModelHash, "view": "Views::device", "includeSemanticIdentity": true})
	rendered, err := server.Call(context.Background(), "RenderView", params)
	if err != nil {
		t.Fatal(err)
	}
	var drawing engine.JRenderViewResponse
	if err := json.Unmarshal(rendered, &drawing); err != nil {
		t.Fatal(err)
	}
	if len(drawing.Nodes) != 1 || drawing.Nodes[0].SemanticID != "Demo__device" || drawing.Nodes[0].IdentityStatus != "mapped" {
		t.Fatalf("identity absent: %s", rendered)
	}
	params, _ = json.Marshal(map[string]any{"modelHash": model.ModelHash})
	exported, err := server.Call(context.Background(), "ExportModel", params)
	if err != nil {
		t.Fatal(err)
	}
	var graph struct {
		Elements     []map[string]any `json:"elements"`
		Experimental bool             `json:"experimental"`
		Notice       string           `json:"notice"`
	}
	if err := json.Unmarshal(exported, &graph); err != nil {
		t.Fatal(err)
	}
	if !graph.Experimental || graph.Notice == "" {
		t.Fatalf("experimental mapping not disclosed: %s", exported)
	}
	for _, element := range graph.Elements {
		if element["@id"] == drawing.Nodes[0].SemanticID {
			return
		}
	}
	t.Fatalf("rendered identity does not occur in exported graph: %s", exported)
}

func TestExportRefusesPartialParse(t *testing.T) {
	server, err := New("invalid-test")
	if err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]any{"documents": []map[string]string{{"name": "broken.sysml", "content": "package Demo { part"}}})
	parsed, err := server.Call(context.Background(), "ParseSources", params)
	if err != nil {
		t.Fatal(err)
	}
	var model struct {
		ModelHash string `json:"modelHash"`
	}
	if err := json.Unmarshal(parsed, &model); err != nil {
		t.Fatal(err)
	}
	params, _ = json.Marshal(map[string]string{"modelHash": model.ModelHash})
	if _, err := server.Call(context.Background(), "ExportModel", params); err == nil {
		t.Fatal("exported a partial parse")
	}
}
