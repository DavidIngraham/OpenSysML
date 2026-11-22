package combined

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/translate/export"
)

func TestDeclarationIndexMatchesFullNativeDiscovery(t *testing.T) {
	fixtures := [][]string{
		{`package '@Root' { requirement <short>; requirement 'Owner\'s req'; }`},
		{`package 'A::B' { requirement def R; requirement 'two words' : R; requirement : R; part unit { satisfy 'two words'; } }`},
		{`package P { requirement def R; requirement r : R; part def T { requirement nested; } part x : T { requirement :> T::nested; } }`, `package Q { requirement s : P::R; }`},
		{`package SyspresentRuntimeContext { part item; } package P { alias 'another name' for SyspresentRuntimeContext::item; requirement r; }`},
	}
	for _, texts := range fixtures {
		server, err := New("index-test")
		if err != nil {
			t.Fatal(err)
		}
		documents := []map[string]string{}
		for i, text := range texts {
			documents = append(documents, map[string]string{"name": string(rune('a'+i)) + ".sysml", "content": text})
		}
		params, _ := json.Marshal(map[string]any{"documents": documents})
		parsed, err := server.Call(context.Background(), "ParseSources", params)
		if err != nil {
			t.Fatal(err)
		}
		var model struct {
			ModelHash string `json:"modelHash"`
			Roots     []struct {
				ChildIDs []string `json:"childIds"`
			} `json:"roots"`
		}
		if err = json.Unmarshal(parsed, &model); err != nil {
			t.Fatal(err)
		}
		call := func(projection string) []byte {
			t.Helper()
			p, _ := json.Marshal(map[string]string{"modelHash": model.ModelHash, "projection": projection})
			b, e := server.Call(context.Background(), "ExportModel", p)
			if e != nil {
				t.Fatal(e)
			}
			return b
		}
		var indexed struct {
			Projection string                         `json:"projection"`
			Entries    []export.DeclarationIndexEntry `json:"declarationIndex"`
			Elements   json.RawMessage                `json:"elements"`
		}
		if err = json.Unmarshal(call("declaration-index"), &indexed); err != nil {
			t.Fatal(err)
		}
		if indexed.Projection != "declaration-index" || len(indexed.Elements) > 0 {
			t.Fatalf("index masquerades as API objects: %+v", indexed)
		}
		var full struct {
			Elements []struct {
				ID            string `json:"@id"`
				Type          string `json:"@type"`
				Name          string `json:"declaredName"`
				QualifiedName string `json:"qualifiedName"`
			} `json:"elements"`
		}
		if err = json.Unmarshal(call("declarations"), &full); err != nil {
			t.Fatal(err)
		}
		rootIDs := map[string]bool{}
		for _, root := range model.Roots {
			for _, id := range root.ChildIDs {
				rootIDs[id] = true
			}
		}
		var wantRoots, wantSeeds, gotRoots, gotSeeds []string
		for _, el := range full.Elements {
			if el.QualifiedName == "" {
				continue
			}
			if rootIDs[el.ID] {
				wantRoots = append(wantRoots, el.QualifiedName)
			}
			if el.Name != "" && (el.Type == "RequirementUsage" || el.Type == "RequirementDefinition") {
				wantSeeds = append(wantSeeds, el.QualifiedName)
			}
		}
		for _, el := range indexed.Entries {
			if rootIDs[el.ID] {
				gotRoots = append(gotRoots, el.QualifiedName)
			}
			if el.RequirementSeed {
				gotSeeds = append(gotSeeds, el.QualifiedName)
			}
		}
		for _, v := range [][]string{wantRoots, wantSeeds, gotRoots, gotSeeds} {
			slices.Sort(v)
		}
		if !reflect.DeepEqual(gotRoots, wantRoots) || !reflect.DeepEqual(gotSeeds, wantSeeds) {
			t.Fatalf("discovery differs: roots %q want %q; seeds %q want %q", gotRoots, wantRoots, gotSeeds, wantSeeds)
		}
		bad, _ := json.Marshal(map[string]string{"modelHash": "absent", "projection": "declaration-index"})
		if _, err = server.Call(context.Background(), "ExportModel", bad); err == nil {
			t.Fatal("unknown model accepted")
		}
	}
}
