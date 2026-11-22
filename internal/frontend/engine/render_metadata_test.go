package engine

import (
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"testing"
)

func TestRenderMetadataEncoding(t *testing.T) {
	var rendered JRenderViewResponse
	data := view.Data{Rows: []view.RowData{{Cells: []string{"label"}, SemanticID: "canonical-row", QualifiedName: "Demo::thing", IdentityStatus: "mapped"}}, Nodes: []view.NodeData{{ID: "n0", Text: "payload", StandIn: true, Verdict: "inconclusive"}}, Edges: []view.EdgeData{{From: "n0", To: "n1", FromLabel: "a", ToLabel: "b", Name: "message"}}}
	response, err := marshalRenderViewExtended(&cachedModel{}, data, "", true)
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
