package resolve_test

import (
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/resolve"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// bindingOf resolves the reference written as text in the document and
// returns the fully-qualified name it binds, or "" when it is unresolved.
func bindingOf(t *testing.T, r *resolve.Resolver, root *ast.RootNamespace, scope *symbols.Scope, text string) string {
	t.Helper()
	var found *resolve.Reference
	for _, ref := range resolve.References(root, scope) {
		if nameText(ref.QN) == text {
			ref := ref
			if found != nil {
				t.Fatalf("%q is written more than once", text)
			}
			found = &ref
		}
	}
	if found == nil {
		t.Fatalf("%q is not written", text)
	}
	sym, ok := r.ResolveReference(*found)
	if !ok {
		return ""
	}
	return symbols.FQNOf(sym)
}

// diagnosticsOf splits the resolver's diagnostics into warning and error texts.
func diagnosticsOf(r *resolve.Resolver) (warnings, errors []string) {
	for _, d := range r.Diagnostics {
		if d.Warning {
			warnings = append(warnings, d.Message)
		} else {
			errors = append(errors, d.Message)
		}
	}
	return warnings, errors
}
