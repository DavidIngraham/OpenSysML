package migrate_test

import (
	"bytes"
	"strings"
	"testing"
)

// The compartments a modeler collapsed on a symbol are written as the symbol's
// Layout show… switches, from the symbol's own SUPPRESS_* properties and then
// its symbol style's chain; a property with no value is the tool's default,
// shown, and overrides an inherited suppression. A suppressed compartment no
// v2 compartment answers to is an unsupported presentation property.
func TestGoldenSuppressedCompartments(t *testing.T) {
	r := migrateStream(t, "suppressed_compartments")
	l := r.Report.Layout
	if l == nil || l.StreamDiagrams != 1 || l.PlacementsWritten != 4 || l.CompartmentsSuppressed != 8 || l.SymbolsSuppressing != 2 {
		t.Fatalf("layout summary: %+v", l)
	}
	if l.Unsupported["SUPPRESS_STRUCTURE"] != 2 {
		t.Errorf("unsupported = %v, want SUPPRESS_STRUCTURE twice", l.Unsupported)
	}
	checkGolden(t, "testdata/xmi/suppressed_compartments.golden.sysml", r.Notation)
	var report bytes.Buffer
	if err := r.Report.WriteText(&report); err != nil {
		t.Fatal(err)
	}
	checkGolden(t, "testdata/xmi/suppressed_compartments.golden.report.txt", report.Bytes())
	for _, d := range errors(t, "suppressed_compartments.sysml", r.Notation) {
		t.Errorf("%v", d)
	}
	notation := string(r.Notation)
	for _, want := range []string{
		// Camera: its own parts suppression, attributes and references from the
		// style, ports from the style's parent; its own unvalued operations
		// property overrides the parent's suppression.
		"metadata DiagramLayout::Layout about Camera { x = 40; y = 20; width = 200; height = 100; showAttributes = false; showParts = false; showPorts = false; showReferences = false; }",
		// Lens: the style's settings alone.
		"metadata DiagramLayout::Layout about Lens { x = 40; y = 220; width = 120; height = 33; showActions = false; showAttributes = false; showPorts = false; showReferences = false; }",
		// Mount: no style, one unvalued property: nothing suppressed.
		"metadata DiagramLayout::Layout about Mount { x = 300; y = 220; width = 120; height = 60; }",
		"metadata DiagramLayout::Layout about Optics::Camera::data { x = 235; y = 60; width = 10; height = 10; }",
		"# compartments suppressed: 8 on 2 symbols",
	} {
		if !strings.Contains(notation, want) && !strings.Contains(report.String(), want) {
			t.Errorf("output lacks %q", want)
		}
	}
}
