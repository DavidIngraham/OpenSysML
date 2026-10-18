package engine

import (
	"context"
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"runtime"
	"strings"
	"testing"
)

func TestRenderIdentityAndNotationFields(t *testing.T) {
	eng, err := New()
	if err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]any{"documents": []map[string]string{{"name": "identity.sysml", "content": `package Demo { part def Thing; view diagram { expose Thing; } }`}}})
	parsed, err := eng.Call(context.Background(), "ParseSources", params)
	if err != nil {
		t.Fatal(err)
	}
	var model JParseSourcesResponse
	if err := json.Unmarshal(parsed, &model); err != nil {
		t.Fatal(err)
	}
	params, _ = json.Marshal(JRenderViewRequest{ModelHash: model.ModelHash, View: "Demo::diagram"})
	if _, err := eng.Call(context.Background(), "RenderView", params); err != nil {
		t.Fatalf("ordinary rendering unavailable: %v", err)
	}
	params, _ = json.Marshal(JRenderViewRequest{ModelHash: model.ModelHash, View: "Demo::diagram", IncludeSemanticIdentity: true})
	response, err := eng.Call(context.Background(), "RenderView", params)
	if runtime.GOOS == "js" || runtime.GOOS == "wasip1" {
		if err == nil || !strings.Contains(err.Error(), "semantic rendering identity is unavailable") {
			t.Fatalf("want explicit platform capability error, got %v", err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	var rendered JRenderViewResponse
	if err := json.Unmarshal(response, &rendered); err != nil {
		t.Fatal(err)
	}
	if len(rendered.Nodes) != 1 || rendered.Nodes[0].SemanticID != "Demo__Thing" || rendered.Nodes[0].QualifiedName != "Demo::Thing" {
		t.Fatalf("identity lost: %s", response)
	}
	data := view.Data{Rows: []view.RowData{{Cells: []string{"label"}, SemanticID: "canonical-row", QualifiedName: "Demo::thing", IdentityStatus: "mapped"}}, Nodes: []view.NodeData{{ID: "n0", Text: "payload", StandIn: true, Verdict: "inconclusive"}}, Edges: []view.EdgeData{{From: "n0", To: "n1", FromLabel: "a", ToLabel: "b", Name: "message"}}}
	response, err = marshalRenderView(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(response, &rendered); err != nil {
		t.Fatal(err)
	}
	if len(rendered.Rows) != 1 || rendered.Rows[0].SemanticID != "canonical-row" || rendered.Rows[0].QualifiedName != "Demo::thing" || rendered.Rows[0].IdentityStatus != "mapped" {
		t.Fatalf("row identity lost: %s", response)
	}
	if rendered.Nodes[0].Text != "payload" || !rendered.Nodes[0].StandIn || rendered.Nodes[0].Verdict != "inconclusive" || rendered.Edges[0].FromLabel != "a" || rendered.Edges[0].ToLabel != "b" || rendered.Edges[0].Name != "message" {
		t.Fatalf("notation fields lost: %s", response)
	}
}
