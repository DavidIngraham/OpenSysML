package view

import (
	"fmt"
	"strings"
	"testing"
)

func TestRequirementContextDepthIsDeterministic(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, depth := range []int{0, 1, 2} {
			t.Run(fmt.Sprintf("reverse_%t_depth_%d", reverse, depth), func(t *testing.T) {
				relationships := `dependency linkAB from b to a; dependency linkBC from c to b;`
				if reverse {
					relationships = `dependency linkBC from c to b; dependency linkAB from b to a;`
				}
				text := fmt.Sprintf(`package Demo { private import StandardViewDefinitions::*; requirement a; requirement b; requirement c; %s view selected : GeneralView { filter @SysML::RequirementUsage; @OpenSysMLRenderings::RequirementContext { relationshipDepth = %d; } expose a; } }`, relationships, depth)
				rendered := renderSource(t, "Demo::selected", text)
				data := rendered.Data()
				if len(data.Nodes) != depth+1 || len(data.Edges) != depth {
					t.Fatalf("depth not bounded: nodes=%d edges=%d", len(data.Nodes), len(data.Edges))
				}
				for _, edge := range data.Edges {
					if edge.Kind != EdgeDependency {
						t.Fatalf("dependency changed kind: %+v", edge)
					}
				}
			})
		}
	}
}

func TestRequirementContextInducedSelectionAndLegacy(t *testing.T) {
	source := `package Demo { private import StandardViewDefinitions::*; requirement a; requirement b; requirement c; dependency linkAB from b to a; dependency linkBC from c to b; view selected : GeneralView { filter @SysML::RequirementUsage; %s expose a; expose b; } }`
	bounded := renderSource(t, "Demo::selected", fmt.Sprintf(source, `@OpenSysMLRenderings::RequirementContext { relationshipDepth = 0; }`)).Data()
	if len(bounded.Nodes) != 2 || len(bounded.Edges) != 1 || bounded.Edges[0].Kind != EdgeDependency {
		t.Fatalf("induced policy changed: %+v", bounded)
	}
	legacy := renderSource(t, "Demo::selected", fmt.Sprintf(source, "")).Data()
	if len(legacy.Nodes) != 2 || len(legacy.Edges) != 0 {
		t.Fatalf("legacy behavior changed: %+v", legacy)
	}
}

func TestRequirementContextRejectsInvalidDepth(t *testing.T) {
	for _, value := range []string{"-1", "65", "true", "1.5"} {
		source := fmt.Sprintf(`package Demo { private import StandardViewDefinitions::*; requirement a; view selected : GeneralView { filter @SysML::RequirementUsage; @OpenSysMLRenderings::RequirementContext { relationshipDepth = %s; } expose a; } }`, value)
		renderer, index := loadSources(t, []string{"policy.sysml"}, [][]byte{[]byte(source)})
		_, err := renderer.Render(lookup(t, index, "Demo::selected"))
		if err == nil || !strings.Contains(err.Error(), "relationshipDepth") {
			t.Fatalf("invalid %s depth accepted: %v", value, err)
		}
	}
}

func TestRequirementContextDoesNotExpandRequirementParties(t *testing.T) {
	text := `package Demo {
		private import StandardViewDefinitions::*;
		part def Vehicle;
		requirement def Limit { subject machine : Vehicle; }
		view legacy : GeneralView {
			filter @SysML::RequirementDefinition; expose Limit;
		}
		view bounded : GeneralView {
			@OpenSysMLRenderings::RequirementContext { relationshipDepth = 0; }
			filter @SysML::RequirementDefinition; expose Limit;
		}
	}`
	legacy := renderSource(t, "Demo::legacy", text).Data()
	bounded := renderSource(t, "Demo::bounded", text).Data()
	if len(legacy.Nodes) != 2 || len(bounded.Nodes) != 1 {
		t.Fatalf("requirement party expansion: legacy=%d bounded=%d", len(legacy.Nodes), len(bounded.Nodes))
	}
}
