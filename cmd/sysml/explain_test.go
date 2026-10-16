package main

import (
	"strings"
	"testing"
)

// TestExplainPrintsACode checks that -explain=<code> prints the catalogue entry
// of a code and exits without reading a model, and that -explain alone lists
// every code with its summary.
func TestExplainPrintsACode(t *testing.T) {
	binary := buildCLI(t)

	got := check(t, binary, "part x", "-explain=unresolved")
	wantReport(t, got, 0, "unresolved: error (name-resolution)", "Specification: KerML")
	if strings.Contains(got.output(), "syntax") {
		t.Errorf("-explain read the model:\n%s", got.output())
	}

	out := run(t, binary, "-explain")
	for _, want := range []string{"Syntax\n", "Name resolution\n", "  unresolved ", "  syntax "} {
		if !strings.Contains(out, want) {
			t.Errorf("-explain is missing %q:\n%s", want, out)
		}
	}

	got = check(t, binary, "", "-explain=unresolved", "-explain=syntax")
	wantReport(t, got, 0, "unresolved: error", "\n\nsyntax: error (syntax)")
}

// TestExplainRefusesAnUnknownCode checks that a code the catalogue lacks is a
// usage error that names the listing.
func TestExplainRefusesAnUnknownCode(t *testing.T) {
	binary := buildCLI(t)

	got := check(t, binary, "", "-explain=bogus")
	wantReport(t, got, 2, `unknown diagnostic code "bogus"`, "-explain")
	if got.stdout != "" {
		t.Errorf("stdout = %q, want nothing", got.stdout)
	}
}
