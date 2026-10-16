package migrate

import (
	"sort"

	"github.com/Open-MBEE/OpenSysML/internal/translate/xmi/sysmlv1"
)

// compartmentKinds maps a Cameo symbol's SUPPRESS_* compartment property to
// the DiagramLayout::Layout show… switch of the v2 compartment its members
// become: parts → part usages, references → ref usages, properties and values
// (and a class's attributes) → attributes, constraints → constraint usages,
// flow properties → item features, proxy, full and class ports → port usages,
// operations → action usages. The IBD_ properties are a part symbol's
// compartments on an internal block diagram and map the same way.
var compartmentKinds = map[string]string{
	"SUPPRESS_PARTS_COMPARTMENT":              "Parts",
	"SUPPRESS_IBD_PARTS_COMPARTMENT":          "Parts",
	"SUPPRESS_REFERENCES_COMPARTMENT":         "References",
	"SUPPRESS_IBD_REFERENCES_COMPARTMENT":     "References",
	"SUPPRESS_PROPERTIES_COMPARTMENT":         "Attributes",
	"SUPPRESS_VALUES_COMPARTMENT":             "Attributes",
	"SUPPRESS_CLASS_ATTRIBUTES":               "Attributes",
	"SUPPRESS_IBD_VALUES_COMPARTMENT":         "Attributes",
	"SUPPRESS_IBD_UMLPROPS_COMPARTMENT":       "Attributes",
	"SUPPRESS_IBD_INITIALVALUES_COMPARTMENT":  "Attributes",
	"SUPPRESS_CONSTRAINTS_COMPARTMENT":        "Constraints",
	"SUPPRESS_IBD_CONSTRAINTS_COMPARTMENT":    "Constraints",
	"SUPPRESS_FLOWPROPERTIES_COMPARTMENT":     "Items",
	"SUPPRESS_IBD_FLOWPROPERTIES_COMPARTMENT": "Items",
	"SUPPRESS_PROXYPORTS_COMPARTMENT":         "Ports",
	"SUPPRESS_FULLPORTS_COMPARTMENT":          "Ports",
	"SUPPRESS_CLASS_PORTS":                    "Ports",
	"SUPPRESS_CLASS_OPERATIONS":               "Actions",
	"SUPPRESS_IBD_OPERATIONS_COMPARTMENT":     "Actions",
}

// hiddenCompartments names the v2 compartment kinds sym suppresses, sorted: a
// kind is hidden when any Cameo compartment mapping to it is suppressed. The
// suppressed compartments no kind maps to are named apart, sorted, for the
// report's unsupported presentation properties.
func hiddenCompartments(sym *sysmlv1.Symbol) (kinds, unsupported []string) {
	hidden := map[string]bool{}
	for id, suppressed := range sym.Compartments {
		if !suppressed {
			continue
		}
		if kind, ok := compartmentKinds[id]; ok {
			hidden[kind] = true
		} else {
			unsupported = append(unsupported, id)
		}
	}
	for kind := range hidden {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	sort.Strings(unsupported)
	return kinds, unsupported
}
