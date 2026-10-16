package migrate

import (
	"reflect"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/translate/xmi/sysmlv1"
)

func TestHiddenCompartments(t *testing.T) {
	m := &migration{layoutSummary: &LayoutSummary{Unsupported: map[string]int{}}}
	sym := &sysmlv1.Symbol{Compartments: map[string]bool{
		"SUPPRESS_PARTS_COMPARTMENT":          false,
		"SUPPRESS_VALUES_COMPARTMENT":         true,
		"SUPPRESS_CLASS_ATTRIBUTES":           true,
		"SUPPRESS_PROXYPORTS_COMPARTMENT":     true,
		"SUPPRESS_FULLPORTS_COMPARTMENT":      false,
		"SUPPRESS_FLOWPROPERTIES_COMPARTMENT": true,
		"SUPPRESS_STRUCTURE":                  true,
		"SUPPRESS_BEHAVIORS":                  false,
	}}
	if got, want := m.hiddenCompartments(sym), []string{"Attributes", "Items", "Ports"}; !reflect.DeepEqual(got, want) {
		t.Errorf("hidden = %q, want %q", got, want)
	}
	if want := map[string]int{"SUPPRESS_STRUCTURE": 1}; !reflect.DeepEqual(m.layoutSummary.Unsupported, want) {
		t.Errorf("unsupported = %v, want %v", m.layoutSummary.Unsupported, want)
	}
	if got := m.hiddenCompartments(&sysmlv1.Symbol{}); len(got) != 0 {
		t.Errorf("hidden of a symbol without settings = %q, want none", got)
	}
}
