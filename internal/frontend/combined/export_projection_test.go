package combined

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestDeclarationProjectionIsExplicitAndKeepsParsedModel(t *testing.T) {
	server, err := New("projection-test")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := server.Call(context.Background(), "ParseSources", []byte(`{"documents":[{"name":"model.sysml","content":"package Demo { part thing; view inspect { expose thing; } }"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	var model struct {
		ModelHash string `json:"modelHash"`
	}
	if err = json.Unmarshal(parsed, &model); err != nil {
		t.Fatal(err)
	}
	params, _ := json.Marshal(map[string]string{"modelHash": model.ModelHash, "projection": "declarations"})
	body, err := server.Call(context.Background(), "ExportModel", params)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Projection   string           `json:"projection"`
		Elements     []map[string]any `json:"elements"`
		ElementCount int              `json:"elementCount"`
		Omitted      int              `json:"omittedElementCount"`
		Notice       string           `json:"coverageNotice"`
	}
	if err = json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Projection != "declarations" || response.ElementCount != len(response.Elements) || response.Omitted == 0 || response.Notice == "" {
		t.Fatalf("projection coverage missing: %s", body)
	}
	for _, element := range response.Elements {
		if _, ok := element["sysx:sourceLine"]; !ok {
			t.Fatalf("non-declaration included: %+v", element)
		}
	}
	params, _ = json.Marshal(map[string]any{"modelHash": model.ModelHash, "view": "Demo::inspect", "includeSemanticIdentity": true})
	if _, err = server.Call(context.Background(), "RenderView", params); err != nil {
		t.Fatalf("projection damaged parsed model: %v", err)
	}
	params, _ = json.Marshal(map[string]string{"modelHash": model.ModelHash, "projection": "guess"})
	if _, err = server.Call(context.Background(), "ExportModel", params); err == nil {
		t.Fatal("unknown projection accepted")
	}
}

func TestDeclarationProjectionMatchesFullResolvedActionReferences(t *testing.T) {
	server, err := New("projection-resolver-test")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := server.Call(context.Background(), "ParseSources", []byte(`{"documents":[{"name":"action.sysml","content":"package Demo { action turnOn { action light; first start then light; } }"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	var model struct {
		ModelHash string `json:"modelHash"`
	}
	if err = json.Unmarshal(parsed, &model); err != nil {
		t.Fatal(err)
	}
	var projected, full struct {
		Elements []map[string]any `json:"elements"`
	}
	for _, request := range []struct {
		projection string
		target     any
	}{{"declarations", &projected}, {"full", &full}} {
		params, _ := json.Marshal(map[string]string{"modelHash": model.ModelHash, "projection": request.projection})
		body, callErr := server.Call(context.Background(), "ExportModel", params)
		if callErr != nil {
			t.Fatal(callErr)
		}
		if err = json.Unmarshal(body, request.target); err != nil {
			t.Fatal(err)
		}
	}
	expected := map[string]map[string]any{}
	for _, element := range full.Elements {
		expected[element["@id"].(string)] = element
	}
	for _, element := range projected.Elements {
		if !reflect.DeepEqual(element, expected[element["@id"].(string)]) {
			t.Fatalf("declaration changed resolved API object: %#v", element)
		}
	}
}
