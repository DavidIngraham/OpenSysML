package resolve_test

import (
	"strings"
	"testing"
)

// A feature with no declared name or short name takes the short name of the
// feature it redefines (KerML 8.3.3.3 Feature::effectiveShortName), so the
// redefining feature answers to that short name from the redefining type and
// from whatever specializes, subsets or redefines it, hiding the redefined
// feature's membership (8.3.3.1 Type::inheritedMemberships).
func TestEffectiveShortNameOfARedefiningFeatureResolves(t *testing.T) {
	const model = `package P {
		part def TMS { attribute mass; }
		part def ADS { part <tms> t : TMS; }
		part gear : ADS { part :>> t { } }
		part g2 :> gear;
		part g2b :> g2 { part :>> t; }
		abstract part <ads> adsFull { part <tmsU> thermalManagementSystem : TMS; }
		part gearU subsets ads { part :>> thermalManagementSystem { } }
	`
	cases := []struct {
		name, body, ref, binds string
	}{
		{"redefining type", `part c { attribute a :> gear.tms.mass; }`, "tms", "P::gear::t"},
		{"one level of specialization", `part c { attribute a :> g2.tms.mass; }`, "tms", "P::gear::t"},
		{"two levels of specialization", `part c :> g2 { attribute a :> c.tms.mass; }`, "tms", "P::gear::t"},
		{"redefined again two levels down", `part c { attribute a :> g2b.tms.mass; }`, "tms", "P::g2b::t"},
		{"through redefines", `part c { ref part sua redefines gear; attribute a :> sua.tms.mass; }`, "tms", "P::gear::t"},
		{"through subsets", `part c { ref part sua :> gear; attribute a :> sua.tms.mass; }`, "tms", "P::gear::t"},
		{"redefined feature on a usage", `part c { ref part sua :> gearU; attribute a :> sua.tmsU.mass; }`, "tmsU", "P::gearU::thermalManagementSystem"},
		{"qualified name", `attribute q :> gear::tms;`, "gear::tms", "P::gear::t"},
		{"qualified name one level down", `attribute q :> g2::tms;`, "g2::tms", "P::gear::t"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, root, scope := resolvedDoc(t, model+tc.body+"\n}")
			if got := bindingOf(t, r, root, scope, tc.ref); got != tc.binds {
				t.Errorf("%s binds %q, want %q", tc.ref, got, tc.binds)
			}
			if _, errs := diagnosticsOf(r); len(errs) != 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}

// A redefining feature that declares a short name of its own has that short
// name, not the redefined feature's (effectiveShortName returns the
// declaredShortName when one is given), and the redefinition hides the
// redefined feature, so the old short name no longer resolves.
func TestDeclaredShortNameOfARedefiningFeatureReplacesTheInheritedOne(t *testing.T) {
	const model = `package P {
		part def TMS { attribute mass; }
		part def ADS { part <tms> t : TMS; }
		part gear : ADS { part :>> t { } }
		part g3 :> gear { part <t3> :>> t; }
	`
	cases := []struct {
		name, body, ref, binds, unresolved string
	}{
		{"new short name", `part c { attribute a :> g3.t3.mass; }`, "t3", "P::g3::t3", ""},
		{"old short name", `part c { attribute a :> g3.tms.mass; }`, "tms", "", "unresolved member: tms"},
		{"old short name qualified", `attribute q :> g3::tms;`, "g3::tms", "", "unresolved reference: g3::tms"},
		{"old short name one level down", `part g4 :> g3; part c { attribute a :> g4.tms.mass; }`, "tms", "", "unresolved member: tms"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, root, scope := resolvedDoc(t, model+tc.body+"\n}")
			if got := bindingOf(t, r, root, scope, tc.ref); got != tc.binds {
				t.Errorf("%s binds %q, want %q", tc.ref, got, tc.binds)
			}
			_, errs := diagnosticsOf(r)
			if tc.unresolved == "" {
				if len(errs) != 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
				return
			}
			if len(errs) != 1 || !strings.Contains(errs[0], tc.unresolved) {
				t.Errorf("errors %v, want one containing %q", errs, tc.unresolved)
			}
		})
	}
}
