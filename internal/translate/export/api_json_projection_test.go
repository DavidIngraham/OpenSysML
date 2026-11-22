package export

import (
	"encoding/json"
	"fmt"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/parser"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
	"reflect"
	"testing"
)

func TestAPIJSONDeclarationsPreservesNativeObjects(t *testing.T) {
	file := source.New("declarations.sysml", []byte(`package Example { part def Unit { attribute mass : ScalarValues::Real = 2+3; } part unit : Unit; requirement safety; part implementation { satisfy safety; } }`))
	root := parser.New(file).ParseFile()
	graph, err := ToRDF(file, root)
	if err != nil {
		t.Fatal(err)
	}
	full, err := WriteAPIJSON(graph)
	if err != nil {
		t.Fatal(err)
	}
	projected, counts, err := WriteAPIJSONDeclarations(graph)
	if err != nil {
		t.Fatal(err)
	}
	var all, got []map[string]any
	if err = json.Unmarshal(full, &all); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(projected, &got); err != nil {
		t.Fatal(err)
	}
	var want []map[string]any
	for _, element := range all {
		if _, ok := element["sysx:sourceLine"]; ok {
			want = append(want, element)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("projected API objects changed: got %s want declarations from %s", projected, full)
	}
	if counts.ElementCount != len(want) || counts.OmittedElementCount != len(all)-len(want) || counts.OmittedElementCount == 0 {
		t.Fatalf("wrong coverage counts: %+v full=%d declarations=%d", counts, len(all), len(want))
	}
}

func TestModelDeclarationsMatchesFullCrossDocumentConversion(t *testing.T) {
	texts := []string{
		`package Example { part def Unit { attribute mass : ScalarValues::Real = 2+3; } requirement safety; verification def Check { objective { verify safety; } } }`,
		`package Product { part unit : Example::Unit { satisfy Example::safety; } verification check : Example::Check; dependency motivation from unit to Example::safety; }`,
	}
	documents := make([]ModelDocument, len(texts))
	for i, text := range texts {
		file := source.New(fmt.Sprintf("document%d.sysml", i), []byte(text))
		documents[i] = ModelDocument{File: file, Root: parser.New(file).ParseFile()}
	}
	graph, err := ModelToRDFWith(documents, IDQualifiedName)
	if err != nil {
		t.Fatal(err)
	}
	full, err := WriteAPIJSON(graph)
	if err != nil {
		t.Fatal(err)
	}
	projected, counts, err := ModelToAPIJSONDeclarations(documents)
	if err != nil {
		t.Fatal(err)
	}
	var all, got []map[string]any
	if err = json.Unmarshal(full, &all); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(projected, &got); err != nil {
		t.Fatal(err)
	}
	var want []map[string]any
	for _, element := range all {
		if _, ok := element["sysx:sourceLine"]; ok {
			want = append(want, element)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("streamed declaration objects changed: got %s want %s", projected, mustProjectionJSON(t, want))
	}
	if counts.ElementCount != len(want) || counts.OmittedElementCount != len(all)-len(want) {
		t.Fatalf("streamed counts differ: %+v full=%d declarations=%d", counts, len(all), len(want))
	}
	fullIDs, err := NewElementIDs(documents)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := NewDeclarationElementIDs(documents)
	if err != nil {
		t.Fatal(err)
	}
	for origin, want := range fullIDs.byOrigin {
		got, ok := inventory.OfOrigin(origin)
		if !ok || got != want {
			t.Fatalf("canonical origin identity changed: origin=%+v got=%+v want=%+v", origin, got, want)
		}
	}
	for name, want := range fullIDs.byName {
		got, ok := inventory.Of(name)
		if !ok || got != want {
			t.Fatalf("canonical name identity changed: %s got=%s want=%s", name, got, want)
		}
	}
}

func mustProjectionJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestDeclarationIdentityInventoryPreservesScopesAndNormativeIDs(t *testing.T) {
	texts := []string{
		`package P { @IdentityMetadata::ProjectRef { projectId = "project-one"; org = "team"; } part 'left::right' { @IdentityMetadata::ElementId { id = "shared"; } attribute value : ScalarValues::Real; } }`,
		`package Q { @IdentityMetadata::ProjectRef { projectId = "project-two"; org = "team"; } part unit { @IdentityMetadata::ElementId { id = "shared"; } } }`,
	}
	documents := make([]ModelDocument, len(texts))
	for i, text := range texts {
		file := source.New(fmt.Sprintf("scope%d.sysml", i), []byte(text))
		parsed := parser.New(file)
		root := parsed.ParseFile()
		if len(parsed.Diagnostics) > 0 {
			t.Fatal(parsed.Diagnostics)
		}
		documents[i] = ModelDocument{File: file, Root: root}
	}
	full, err := NewElementIDs(documents)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := NewDeclarationElementIDs(documents)
	if err != nil {
		t.Fatal(err)
	}
	for origin, want := range full.byOrigin {
		got, ok := inventory.OfOrigin(origin)
		if !ok || got != want {
			t.Fatalf("scope origin identity changed: got=%+v want=%+v", got, want)
		}
	}
	for origin, got := range inventory.byOrigin {
		want, ok := full.OfOrigin(origin)
		if !ok || got != want {
			t.Fatalf("inventory introduced an identity absent from conversion: origin=%+v got=%+v want=%+v", origin, got, want)
		}
	}
	for _, name := range []string{"ScalarValues::Real", "P", "Q", "Q::unit"} {
		want, ok := full.Of(name)
		got, have := inventory.Of(name)
		if ok != have || want != got {
			t.Fatalf("canonical identity changed for %s: got=%s want=%s", name, got, want)
		}
	}
}

func TestDeclarationInventoryUsesSuppliedLibraryOverride(t *testing.T) {
	texts := []string{`library package ScalarValues { attribute def Real { @IdentityMetadata::ElementId { id = "supplied-real"; } } }`, `package Product { attribute reading : ScalarValues::Real; }`}
	var documents []ModelDocument
	for i, text := range texts {
		file := source.New(fmt.Sprintf("override%d.sysml", i), []byte(text))
		parsed := parser.New(file)
		root := parsed.ParseFile()
		if len(parsed.Diagnostics) > 0 {
			t.Fatal(parsed.Diagnostics)
		}
		documents = append(documents, ModelDocument{File: file, Root: root})
	}
	full, err := NewElementIDs(documents)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := NewDeclarationElementIDs(documents)
	if err != nil {
		t.Fatal(err)
	}
	want, ok := full.Of("ScalarValues::Real")
	got, have := inventory.Of("ScalarValues::Real")
	if !ok || !have || got != want || got != "supplied-real" {
		t.Fatalf("supplied library identity lost: got=%s want=%s", got, want)
	}
}

func TestDeclarationsProjectionPreservesCollectionsImportsAndParallelLinks(t *testing.T) {
	tests := []struct {
		name  string
		texts []string
	}{
		{"ordered-members-and-parallel-links", []string{
			"package Types { part def Unit { attribute size : ScalarValues::Real = 2+3; port a; port b; } requirement constraintA; requirement constraintB; }",
			"package Product { import Types::*; part beta : Unit; part alpha : Unit { satisfy constraintB; satisfy constraintA; satisfy constraintB; } dependency d1 from alpha to beta; dependency d2 from alpha to beta; }",
		}},
		{"scoped-identities-and-primitives", []string{
			"package P { @IdentityMetadata::ProjectRef { projectId = \"project-one\"; org = \"team\"; } part 'left::right' { @IdentityMetadata::ElementId { id = \"shared\"; } attribute text : ScalarValues::String = \"<read & write>\"; attribute answer : ScalarValues::Integer = 9007199254740993; } }",
			"package Q { @IdentityMetadata::ProjectRef { projectId = \"project-two\"; org = \"team\"; } part unit { @IdentityMetadata::ElementId { id = \"shared\"; } } }",
		}},
		{"supplied-library-and-reference", []string{
			"library package ScalarValues { attribute def Real { @IdentityMetadata::ElementId { id = \"supplied-real\"; } } }",
			"package Product { attribute reading : ScalarValues::Real; attribute extra : ScalarValues::Real; }",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			documents := make([]ModelDocument, len(test.texts))
			for i, text := range test.texts {
				file := source.New(fmt.Sprintf("collection%d.sysml", i), []byte(text))
				parsed := parser.New(file)
				documents[i] = ModelDocument{File: file, Root: parsed.ParseFile()}
				if len(parsed.Diagnostics) != 0 {
					t.Fatal(parsed.Diagnostics)
				}
			}
			full, err := ModelToRDFWith(documents, IDQualifiedName)
			if err != nil {
				t.Fatal(err)
			}
			want, wantCounts, err := WriteAPIJSONDeclarations(full)
			if err != nil {
				t.Fatal(err)
			}
			got, gotCounts, err := ModelToAPIJSONDeclarations(documents)
			if err != nil {
				t.Fatal(err)
			}
			var gotObjects, wantObjects any
			if err := json.Unmarshal(got, &gotObjects); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(want, &wantObjects); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotObjects, wantObjects) {
				t.Fatalf("projected objects, reference identities or collection order changed: got %s want %s", got, want)
			}
			if gotCounts != wantCounts {
				t.Fatalf("coverage changed: got %+v want %+v", gotCounts, wantCounts)
			}
		})
	}
}

func TestDeclarationsExportObservationsDoNotChangeContent(t *testing.T) {
	file := source.New("observed.sysml", []byte("package P { part a; part b; }"))
	documents := []ModelDocument{{File: file, Root: parser.New(file).ParseFile()}}
	want, wantCounts, err := ModelToAPIJSONDeclarations(documents)
	if err != nil {
		t.Fatal(err)
	}
	var points []APIJSONExportStage
	got, counts, err := ModelToAPIJSONDeclarationsWithOptions(documents, APIJSONDeclarationsOptions{Observe: func(point APIJSONExportStage) { points = append(points, point) }})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) || counts != wantCounts {
		t.Fatal("observing export changed content or counts")
	}
	if len(points) < 7 || points[0].Stage != "start" || points[len(points)-1].Stage != "json" {
		t.Fatalf("missing stage observations: %+v", points)
	}
	for _, point := range points {
		if point.HeapSys < point.HeapAlloc || point.TotalAlloc < point.HeapAlloc {
			t.Fatalf("invalid memory sample: %+v", point)
		}
	}
}

func TestDeclarationsProjectionPreservesReferencedDocumentIdentities(t *testing.T) {
	texts := []string{
		"library package External { part def Unit { attribute mass : ScalarValues::Real; } }",
		"package Product { part beta : External::Unit; part alpha : External::Unit; dependency trace from alpha to beta; }",
	}
	documents := make([]ModelDocument, len(texts))
	for i, text := range texts {
		file := source.New(fmt.Sprintf("referenced%d.sysml", i), []byte(text))
		p := parser.New(file)
		documents[i] = ModelDocument{File: file, Root: p.ParseFile(), Referenced: i == 0}
		if len(p.Diagnostics) > 0 {
			t.Fatal(p.Diagnostics)
		}
	}
	graph, err := ModelToRDFWith(documents, IDQualifiedName)
	if err != nil {
		t.Fatal(err)
	}
	want, wantCounts, err := WriteAPIJSONDeclarations(graph)
	if err != nil {
		t.Fatal(err)
	}
	got, counts, err := ModelToAPIJSONDeclarations(documents)
	if err != nil {
		t.Fatal(err)
	}
	var a, b any
	if err := json.Unmarshal(want, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(got, &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) || counts != wantCounts {
		t.Fatalf("referenced identities or coverage changed: counts %+v want %+v; got %s want %s", counts, wantCounts, got, want)
	}
}
