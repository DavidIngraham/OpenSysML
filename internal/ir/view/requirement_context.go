package view

import (
	"fmt"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
)

const requirementContextMetadata = "OpenSysMLRenderings::RequirementContext"

// requirementContextDepth reads the nearest view's declared policy through native metadata facts.
func (r *Renderer) requirementContextDepth(view *symbols.Symbol) (int, bool, error) {
	if view == nil {
		return 0, false, nil
	}
	for _, subject := range append([]*symbols.Symbol{view}, r.model.AllSupertypes(view)...) {
		found := false
		depth := 0
		for _, annotation := range r.model.AnnotationFactsOf(subject) {
			if !r.isRequirementContextType(annotation.TypeFQN) {
				continue
			}
			if found {
				return 0, true, fmt.Errorf("%s has several RequirementContext annotations", r.fqn(subject))
			}
			found = true
			valid := false
			for _, feature := range annotation.Values {
				if feature.Feature != "relationshipDepth" {
					continue
				}
				if len(feature.Values) == 1 {
					value := feature.Values[0]
					if value.Kind == symbols.FilterValueInt && value.BigInt == nil && value.Int >= 0 && value.Int <= 64 {
						depth = int(value.Int)
						valid = true
					}
				}
			}
			if !valid {
				return 0, true, fmt.Errorf("relationshipDepth of RequirementContext on %s must be a constant Integer from 0 through 64", r.fqn(subject))
			}
		}
		if found {
			return depth, true, nil
		}
	}
	return 0, false, nil
}

func (r *Renderer) isRequirementContextType(name string) bool {
	if name == requirementContextMetadata {
		return true
	}
	if r.resolver == nil || r.resolver.Index() == nil {
		return false
	}
	for _, typ := range r.resolver.Index().LookupQualified(name) {
		for _, base := range r.model.AllSupertypes(typ) {
			if r.fqn(base) == requirementContextMetadata {
				return true
			}
		}
	}
	return false
}
