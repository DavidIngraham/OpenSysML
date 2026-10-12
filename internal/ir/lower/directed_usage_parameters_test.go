package lower

import (
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// A directed concern, requirement, use case or other usage kind a behavior body
// admits is a parameter of the owning action, as a directed item or part is.
func TestActionGraphDirectedUsageKindsAreParameters(t *testing.T) {
	graph := actionGraphFor(t, `
		action elicit {
			in concern problem : C;
			out requirement spec : R { subject s : I; }
			out use case uses[*] : U;
			out verification checks[*] : V;
			inout item product : I;
			concern aside : C;
		}
	`)
	want := []struct {
		name      string
		kind      ast.UsageKind
		direction ast.FeatureDirection
	}{
		{"problem", ast.UsageConcern, ast.DirIn},
		{"spec", ast.UsageRequirement, ast.DirOut},
		{"uses", ast.UsageUseCase, ast.DirOut},
		{"checks", ast.UsageVerificationCase, ast.DirOut},
		{"product", ast.UsageItem, ast.DirInOut},
	}
	if len(graph.Parameters) != len(want) {
		t.Fatalf("parameters = %+v, want %d", graph.Parameters, len(want))
	}
	for i, w := range want {
		got := graph.Parameters[i]
		u, _ := got.Node.(*ast.Usage)
		if got.Name != w.name || got.Direction != w.direction || u == nil || u.Kind != w.kind {
			t.Errorf("parameter %d = %+v, want %s %v/%v", i, got, w.name, w.kind, w.direction)
		}
	}
}
