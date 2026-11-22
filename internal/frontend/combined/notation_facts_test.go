package combined

import (
	"context"
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/engine"
	"os"
	"testing"
)

func TestNativeNotationWireIdentityAndDisclosure(t *testing.T) {
	eng, err := New("native-metadata-contract-test")
	if err != nil {
		t.Fatal(err)
	}
	call := func(method string, params any, out any) {
		t.Helper()
		p, _ := json.Marshal(params)
		body, err := eng.Call(context.Background(), method, p)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(body, out); err != nil {
			t.Fatalf("%v:%s", err, body)
		}
	}
	sequence, err := os.ReadFile("../../ir/view/testdata/sequence.sysml")
	if err != nil {
		t.Fatal(err)
	}
	action := `package Actions { private import StandardViewDefinitions::*; item def Value; action def Source {out value:Value;} action def Sink {in value:Value;} action def Pipe {action source:Source; then action sink:Sink; flow source.value to sink.value;} view pipe:ActionFlowView {expose Pipe;} }`
	state, err := os.ReadFile("../../ir/view/testdata/state-do.sysml")
	if err != nil {
		t.Fatal(err)
	}
	general, err := os.ReadFile("../../ir/view/testdata/general.sysml")
	if err != nil {
		t.Fatal(err)
	}
	var parsed engine.JParseSourcesResponse
	call("ParseSources", map[string]any{"documents": []map[string]string{{"name": "sequence.sysml", "content": string(sequence)}, {"name": "action.sysml", "content": action}, {"name": "state-do.sysml", "content": string(state)}, {"name": "general.sysml", "content": string(general)}}}, &parsed)
	for _, identity := range []bool{true, false, true} {
		var out engine.JRenderViewResponse
		call("RenderView", engine.JRenderViewRequest{ModelHash: parsed.ModelHash, View: "SequenceViews::pubSubView", IncludeSemanticIdentity: identity, IncludeRenderingMetadata: true}, &out)
		if len(out.Edges) != 6 || len(out.SequenceOrders) != 3 {
			t.Fatalf("wire lost native occurrences: %+v", out)
		}
		parents := map[string]string{}
		for _, node := range out.Nodes {
			parents[node.ID] = node.Parent
		}
		messages := 0
		for _, edge := range out.Edges {
			if edge.Kind == "succession" {
				continue
			}
			messages++
			n := edge.Notation
			if n == nil || n.Production != "message" || n.SourceEvent == nil || n.TargetEvent == nil || n.SourceEvent.Kind != "event occurrence" || n.SourceEvent.LifelineID != parents[edge.From] {
				t.Fatalf("wire facts absent: %+v", edge)
			}
			if identity {
				if n.SourceEvent.SemanticID == "" || n.SourceEvent.IdentityStatus != "mapped" {
					t.Fatalf("unbound event: %+v", n)
				}
			} else if n.SourceEvent.SemanticID != "" {
				t.Fatal("cached event identity leaked")
			}
		}

		if messages != 3 {
			t.Fatalf("messages=%d", messages)
		}
		for _, order := range out.SequenceOrders {
			if order.SourceEvent == nil || order.TargetEvent == nil || order.SourceEvent.LifelineID == "" || order.TargetEvent.LifelineID == "" {
				t.Fatalf("order endpoint association missing %+v", order)
			}
			if identity && (order.SourceEvent.SemanticID == "" || order.TargetEvent.SemanticID == "") {
				t.Fatal("order endpoint identity missing")
			}
			if !identity && (order.SourceEvent.SemanticID != "" || order.TargetEvent.SemanticID != "") {
				t.Fatal("cached order identity leaked")
			}
		}
		call("RenderView", engine.JRenderViewRequest{ModelHash: parsed.ModelHash, View: "InstrumentViews::peas", IncludeSemanticIdentity: identity, IncludeRenderingMetadata: true}, &out)
		behaviors := 0
		for _, node := range out.Nodes {
			if node.State != nil {
				for _, b := range node.State.Behaviors {
					behaviors++
					if b.Role == "" || b.Kind == "" {
						t.Fatal("state behavior roles absent")
					}
					if !identity && b.SemanticID != "" {
						t.Fatal("cached state behavior identity leaked")
					}
				}
			}
		}
		if behaviors != 7 {
			t.Fatalf("state behavior wire count=%d", behaviors)
		}
		call("RenderView", engine.JRenderViewRequest{ModelHash: parsed.ModelHash, View: "GeneralViews::requirementView", IncludeSemanticIdentity: identity, IncludeRenderingMetadata: true}, &out)
		derived := false
		for _, edge := range out.Edges {
			if edge.Kind == "derive" {
				derived = edge.Notation != nil && edge.Notation.Production == "derivation" && edge.Notation.SourceRole == "derived"
			}
		}
		if !derived {
			t.Fatal("derivation native production lost on wire")
		}
		call("RenderView", engine.JRenderViewRequest{ModelHash: parsed.ModelHash, View: "Actions::pipe", IncludeSemanticIdentity: identity, IncludeRenderingMetadata: true, Show: []string{"parameter-symbols", "typings"}}, &out)
		ports := 0
		for _, node := range out.Nodes {
			for _, port := range node.Ports {
				ports++
				if !port.Parameter {
					t.Fatal("native parameter flag lost")
				}
				if identity && port.IdentityStatus != "mapped" {
					t.Fatalf("unbound native parameter %+v", port)
				}
				if !identity && port.SemanticID != "" {
					t.Fatal("cached port identity leaked")
				}
			}
		}
		if ports != 2 {
			t.Fatalf("ports=%d want2", ports)
		}
		flow := false
		for _, edge := range out.Edges {
			if edge.Kind == "flow" {
				flow = edge.FromPort != "" && edge.ToPort != "" && edge.Notation != nil && edge.Notation.Production == "flow"
			}
		}
		if !flow {
			t.Fatalf("native directed flow ports absent %+v", out.Edges)
		}
	}
}
