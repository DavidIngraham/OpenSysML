package grpc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"connectrpc.com/connect"
	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	"github.com/Open-MBEE/OpenSysML/internal/exec/runtime"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/symbolfacts"
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"google.golang.org/protobuf/proto"
)

func artifactModel(t *testing.T, s *Service) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "conformance", "fixtures", "views.sysml"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := parseContent(t, s, string(content))
	return parsed.ModelHash
}

func TestRenderViewArtifactsMatchSharedWriters(t *testing.T) {
	s := mustNewService(t, 10)
	t.Cleanup(s.Close)
	hash := artifactModel(t, s)
	cached, _ := s.cache.Get(hash)
	worker, release := cached.worker()
	defer release()
	renderer := view.NewRenderer(worker.Model.Semantics(), worker.Model.Resolver(), cachedSourceText(cached))
	for _, form := range view.Forms() {
		name := "RenderViewDemo::connections"
		if form == view.FormMarkdown || form == view.FormCSV || form == view.FormTSV {
			name = "#table:RenderViewDemo::Assembly"
		}
		for _, ports := range []string{"minimal", "full"} {
			t.Run(string(form)+"/"+ports, func(t *testing.T) {
				base, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: hash, View: name, Ports: ports})
				if err != nil {
					t.Fatal(err)
				}
				if base.Artifact != nil {
					t.Fatal("data-only request returned an artifact")
				}
				a := &pb.RenderArtifactOptions{Form: string(form), Palette: "okabe-ito", DrawingStyle: "cameo", Direction: "LR", Unplaced: "strip", Width: 40}
				got, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: hash, View: name, Ports: ports, Artifact: a})
				if err != nil {
					t.Fatal(err)
				}
				rendered, err := view.RenderNamed(renderer, name, func(name string) []*symbols.Symbol { return symbolfacts.LookupNamed(cached.Index, name) })
				if err != nil {
					t.Fatal(err)
				}
				p, _ := view.ParsePorts(ports)
				want, err := rendered.WriteWith(form, view.Options{Ports: p, Palette: "okabe-ito", Style: view.StyleCameo, Direction: view.DirectionLeftRight, Unplaced: view.UnplacedStrip, Width: 40})
				if err != nil {
					t.Fatal(err)
				}
				if got.Artifact == nil || got.Artifact.Form != string(form) || got.Artifact.Text != want || got.Artifact.MediaType != renderArtifactMediaType(string(form)) {
					t.Fatalf("artifact = %v; want %q", got.Artifact, want)
				}
				got.Artifact = nil
				if !proto.Equal(got, base) {
					t.Fatal("presentation options changed structured data")
				}
			})
		}
	}
}

func TestRenderViewArtifactRejectsInvalidOptions(t *testing.T) {
	s := mustNewService(t, 10)
	t.Cleanup(s.Close)
	hash := artifactModel(t, s)
	for name, options := range map[string]*pb.RenderArtifactOptions{
		"missing form": {}, "unknown form": {Form: "png"}, "wrong kind": {Form: "csv"},
		"palette": {Form: "text", Palette: "missing"}, "style": {Form: "text", DrawingStyle: "missing"},
		"direction": {Form: "text", Direction: "sideways"}, "unplaced": {Form: "text", Unplaced: "missing"},
		"width": {Form: "text", Width: -1}, "link": {Form: "text", LinkTemplate: "{unknown}"},
		"overlay": {Form: "text", Overlay: "missing"}, "overlay kind": {Form: "text", Overlay: "verdicts"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: hash, View: "RenderViewDemo::connections", Artifact: options})
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("error = %v", err)
			}
		})
	}
	// Default options must be usable, including an omitted unplaced choice.
	got, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: hash, View: "RenderViewDemo::connections", Artifact: &pb.RenderArtifactOptions{Form: "dot"}})
	if err != nil || got.Artifact == nil || !strings.Contains(got.Artifact.Text, "digraph") {
		t.Fatalf("default artifact: %v, %v", got, err)
	}
}

func TestRenderViewArtifactUnicodeLinksAndEmptyViews(t *testing.T) {
	s := mustNewService(t, 10)
	t.Cleanup(s.Close)
	parsed, _ := parseContent(t, s, `package Unicode { part def 'Moteur é'; view diagram { expose 'Moteur é'; } view empty; }`)
	for _, name := range []string{"Unicode::diagram", "Unicode::empty"} {
		result, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: parsed.ModelHash, View: name, Artifact: &pb.RenderArtifactOptions{Form: "dot", LinkTemplate: "https://example.test/{qname}#L{line}"}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Artifact == nil {
			t.Fatal("artifact presence lost")
		}
		if name == "Unicode::diagram" && (!strings.Contains(result.Artifact.Text, "Moteur é") || !strings.Contains(result.Artifact.Text, "https://example.test/") || !strings.Contains(result.Artifact.Text, "#L1")) {
			t.Fatalf("unicode missing: %s", result.Artifact.Text)
		}
	}
}

func TestRenderViewArtifactCapabilityRefusal(t *testing.T) {
	s := mustNewServiceWithout(t, CapabilityRenderViewArtifact)
	t.Cleanup(s.Close)
	hash := artifactModel(t, s)
	if _, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: hash, View: "RenderViewDemo::connections"}); err != nil {
		t.Fatal(err)
	}
	_, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: hash, View: "RenderViewDemo::connections", Artifact: &pb.RenderArtifactOptions{Form: "dot"}})
	if connect.CodeOf(err) != connect.CodeUnimplemented || !strings.Contains(err.Error(), CapabilityRenderViewArtifact) {
		t.Fatalf("refusal = %v", err)
	}
}

func TestRenderViewArtifactVerdictOverlay(t *testing.T) {
	s := mustNewService(t, 10)
	t.Cleanup(s.Close)
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "general-views-demo", "vehicle.sysml"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := parseContent(t, s, string(content))
	result, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: parsed.ModelHash, View: "GeneralViews::requirementView", Artifact: &pb.RenderArtifactOptions{Form: "mermaid", Overlay: "verdicts"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Artifact.Text, "verdict pass by VehicleVerification::lightMassTest, fail by VehicleVerification::heavyMassTest") {
		t.Fatalf("overlay missing: %s", result.Artifact.Text)
	}
	found := false
	for _, node := range result.Nodes {
		if node.Verdict == "fail" {
			found = true
			if node.Style == nil || node.Style.Line != "#D55E00" {
				t.Fatalf("verdict style = %v", node.Style)
			}
		}
	}
	if !found {
		t.Fatal("semantic verdict missing from structured data")
	}
	plain, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: parsed.ModelHash, View: "GeneralViews::requirementView"})
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range plain.Nodes {
		if node.Verdict != "" {
			t.Fatal("overlay leaked into later request")
		}
	}
}

func TestRenderViewArtifactOverlayRespectsBudget(t *testing.T) {
	t.Setenv(runtime.MaxStepsEnvVar, "1")
	t.Setenv(runtime.MaxActionStepsEnvVar, "1")
	s := mustNewService(t, 10)
	t.Cleanup(s.Close)
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "general-views-demo", "vehicle.sysml"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := parseContent(t, s, string(content))
	result, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: parsed.ModelHash, View: "GeneralViews::requirementView", Artifact: &pb.RenderArtifactOptions{Form: "mermaid", Overlay: "verdicts"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Artifact.Text, "evaluation step limit exceeded (1 steps") {
		t.Fatalf("budget refusal absent: %s", result.Artifact.Text)
	}
}

func TestRenderViewArtifactPositionedAndUnplaced(t *testing.T) {
	s := mustNewService(t, 10)
	t.Cleanup(s.Close)
	parsed, _ := parseContent(t, s, `package P {
 private import Views::*;
 private import DiagramLayout::*;
 part def A; part def B;
 view drawing {
 expose A; expose B;
 render asInterconnectionDiagram;
 metadata Layout about A { x = 10; y = 20; width = 100; height = 40; }
 }
 }`)
	var artifacts []string
	for _, unplaced := range []string{"", "strip"} {
		result, err := s.RenderView(context.Background(), &pb.RenderViewRequest{ModelHash: parsed.ModelHash, View: "P::drawing", Artifact: &pb.RenderArtifactOptions{Form: "dot", Unplaced: unplaced}})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Nodes) != 2 || result.Nodes[0].Geometry == nil {
			t.Fatalf("positioned data lost: %v", result.Nodes)
		}
		artifacts = append(artifacts, result.Artifact.Text)
	}
	if !strings.Contains(artifacts[0], "left undrawn") || strings.Contains(artifacts[1], "left undrawn") || !strings.Contains(artifacts[1], "<b>B</b>") {
		t.Fatalf("unplaced options not honored:\n%q\n%q", artifacts[0], artifacts[1])
	}
}
