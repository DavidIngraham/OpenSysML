//go:build !js && !wasip1

package renderidentity

import (
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/parser"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"

	"testing"
)

func TestIdentityCoverage(t *testing.T) {
	file := source.New("coverage.sysml", []byte("package Demo { part thing; }"))
	root := parser.New(file).ParseFile()
	data := view.Data{Nodes: []view.NodeData{
		{ID: "synthetic", NameSynthesized: true},
		{ID: "unknown"},
		{ID: "standin", StandIn: true},
		{ID: "control", Synthetic: true, Origin: symbols.OriginAt("coverage.sysml", source.Span{Offset: 1, Len: 2})},
		{ID: "other", Origin: symbols.OriginAt("other.sysml", source.Span{Offset: 1, Len: 2})},
	}}
	data.Rows = []view.RowData{{Cells: []string{"Demo::thing"}}, {Cells: []string{"thing"}, Origin: symbols.OriginAt("other.sysml", source.Span{Offset: 1, Len: 2})}}
	if err := Bind(&data, []Document{{File: file, Root: root}}, nil); err != nil {
		t.Fatal(err)
	}
	for _, row := range data.Rows {
		if row.IdentityStatus != "unmapped" || row.SemanticID != "" {
			t.Fatalf("row identity inferred from label or unrelated origin: %+v", row)
		}
	}
	for i, want := range []string{"synthetic", "unmapped", "synthetic", "synthetic", "unmapped"} {
		if data.Nodes[i].IdentityStatus != want || data.Nodes[i].SemanticID != "" {
			t.Fatalf("node %d: %+v", i, data.Nodes[i])
		}
	}
}
