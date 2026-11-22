package engine

import (
	"context"
	"encoding/json"
	"testing"
)

func TestMatrixIdentityPreservesAxesDirectionAndDuplicateRelationships(t *testing.T) {
	eng, err := New()
	if err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]any{"documents": []map[string]string{{
		"name": "grid.sysml",
		"content": `package Grid {
			part fromPart;
			part toPart;
			dependency linkOne from fromPart to toPart;
			dependency linkTwo from fromPart to toPart;
			dependency backLink from toPart to fromPart;
		}`,
	}}})
	response, err := eng.Call(context.Background(), "ParseSources", params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed JParseSourcesResponse
	if err := json.Unmarshal(response, &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("grid fixture has diagnostics: %+v", parsed.Diagnostics)
	}
	render := func(identity bool) JRenderViewResponse {
		t.Helper()
		params, _ := json.Marshal(JRenderViewRequest{ModelHash: parsed.ModelHash, View: "#matrix:Grid", IncludeSemanticIdentity: identity})
		response, err := eng.Call(context.Background(), "RenderView", params)
		if err != nil {
			t.Fatal(err)
		}
		var rendered JRenderViewResponse
		if err := json.Unmarshal(response, &rendered); err != nil {
			t.Fatal(err)
		}
		return rendered
	}
	rendered := render(true)
	if rendered.Kind != "matrix" || len(rendered.Rows) != 2 || len(rendered.Columns) != 3 || len(rendered.ColumnElements) != 2 || len(rendered.RelationshipCells) != 2 {
		t.Fatalf("unexpected grid shape: %+v", rendered)
	}
	targets := map[int]string{}
	for _, column := range rendered.ColumnElements {
		if column.Index < 1 || column.Index >= len(rendered.Columns) || column.Name != rendered.Columns[column.Index] || column.IdentityStatus != "mapped" || column.SemanticID == "" {
			t.Fatalf("column is not an identified target at its existing index: %+v", column)
		}
		targets[column.Index] = column.SemanticID
	}
	want := map[string][2]string{
		"Grid__linkOne":  {"Grid__fromPart", "Grid__toPart"},
		"Grid__linkTwo":  {"Grid__fromPart", "Grid__toPart"},
		"Grid__backLink": {"Grid__toPart", "Grid__fromPart"},
	}
	seen := map[string]bool{}
	for _, cell := range rendered.RelationshipCells {
		if cell.Row < 0 || cell.Row >= len(rendered.Rows) || targets[cell.Column] == "" {
			t.Fatalf("invalid cell index: %+v", cell)
		}
		if rendered.Rows[cell.Row].Cells[cell.Column] != "dependency" {
			t.Fatalf("legacy cell changed: %+v", cell)
		}
		for _, relation := range cell.Relationships {
			endpoints, ok := want[relation.SemanticID]
			if !ok || seen[relation.SemanticID] || relation.IdentityStatus != "mapped" || relation.Kind != "dependency" || relation.QualifiedName == "" {
				t.Fatalf("relationship identity lost or duplicated: %+v", relation)
			}
			if relation.SourceSemanticID != endpoints[0] || relation.TargetSemanticID != endpoints[1] ||
				relation.SourceSemanticID != rendered.Rows[cell.Row].SemanticID || relation.TargetSemanticID != targets[cell.Column] {
				t.Fatalf("relationship direction/axis changed: %+v in %+v", relation, cell)
			}
			seen[relation.SemanticID] = true
		}
	}
	if len(seen) != len(want) {
		t.Fatalf("same-kind relationship collapsed: %v", seen)
	}
	withoutIdentity := render(false)
	for _, column := range withoutIdentity.ColumnElements {
		if column.SemanticID != "" {
			t.Fatalf("identity leaked into an unbound render: %+v", column)
		}
	}
	for _, cell := range withoutIdentity.RelationshipCells {
		for _, relation := range cell.Relationships {
			if relation.SemanticID != "" || relation.SourceSemanticID != "" || relation.TargetSemanticID != "" {
				t.Fatalf("identity binding mutated the source rendering: %+v", relation)
			}
		}
	}
	params, _ = json.Marshal(JRenderViewRequest{ModelHash: parsed.ModelHash, View: "#table:Grid", IncludeSemanticIdentity: true})
	response, err = eng.Call(context.Background(), "RenderView", params)
	if err != nil {
		t.Fatal(err)
	}
	var table JRenderViewResponse
	if err := json.Unmarshal(response, &table); err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) == 0 || len(table.ColumnElements) != 0 || len(table.RelationshipCells) != 0 {
		t.Fatalf("ordinary table properties acquired matrix identities: %+v", table)
	}
}

func TestMatrixDerivationIdentityUsesOriginalToDerivedDirection(t *testing.T) {
	eng, err := New()
	if err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]any{"documents": []map[string]string{{
		"name": "derive-grid.sysml",
		"content": `package DeriveGrid {
			private import RequirementDerivation::*;
			requirement sourceReq;
			requirement childReq;
			#derivation connection derivationLink {
				end :> derivedRequirements ::> childReq;
				end :> originalRequirements ::> sourceReq;
			}
		}`,
	}}})
	response, err := eng.Call(context.Background(), "ParseSources", params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed JParseSourcesResponse
	if err := json.Unmarshal(response, &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("derivation fixture has diagnostics: %+v", parsed.Diagnostics)
	}
	params, _ = json.Marshal(JRenderViewRequest{ModelHash: parsed.ModelHash, View: "#matrix:DeriveGrid", IncludeSemanticIdentity: true})
	response, err = eng.Call(context.Background(), "RenderView", params)
	if err != nil {
		t.Fatal(err)
	}
	var rendered JRenderViewResponse
	if err := json.Unmarshal(response, &rendered); err != nil {
		t.Fatal(err)
	}
	if len(rendered.Rows) != 1 || len(rendered.ColumnElements) != 1 || len(rendered.RelationshipCells) != 1 ||
		len(rendered.RelationshipCells[0].Relationships) != 1 {
		t.Fatalf("unexpected derivation matrix: %s", response)
	}
	cell := rendered.RelationshipCells[0]
	relation := cell.Relationships[0]
	if rendered.Rows[0].SemanticID != "DeriveGrid__sourceReq" || rendered.ColumnElements[0].SemanticID != "DeriveGrid__childReq" ||
		cell.Row != 0 || cell.Column != 1 || relation.SemanticID != "DeriveGrid__derivationLink" || relation.Kind != "derive" ||
		relation.SourceSemanticID != "DeriveGrid__sourceReq" || relation.TargetSemanticID != "DeriveGrid__childReq" {
		t.Fatalf("matrix derivation must preserve original-to-derived direction: %s", response)
	}
}
