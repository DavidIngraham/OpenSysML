package export

import (
	"cmp"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/naming"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/resolve"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/semantics"
	"slices"
)

// DeclarationIndexEntry is native discovery metadata, not an API JSON element object.
type DeclarationIndexEntry struct {
	ID              string `json:"id"`
	QualifiedName   string `json:"qualifiedName"`
	RequirementSeed bool   `json:"requirementSeed,omitempty"`
}

// ModelDeclarationIndex inventories names and requirement seeds without deriving API properties.
func ModelDeclarationIndex(documents []ModelDocument, res *resolve.Resolver, model *semantics.Model) ([]DeclarationIndexEntry, error) {
	ids, err := NewDeclarationElementIDsInModel(documents, res, model)
	if err != nil {
		return nil, err
	}
	entries := make([]DeclarationIndexEntry, 0, len(ids.byNode))
	for _, encoder := range ids.encoders {
		for _, node := range encoder.writtenDeclarations() {
			address := encoder.names.FQN[node]
			if address == "" || isPositionalName(address) {
				continue
			}
			declared, _ := naming.DeclaredNameAndMembers(node)
			class := declaredMetaclass(node)
			entries = append(entries, DeclarationIndexEntry{ID: ids.byNode[node], QualifiedName: naming.Escaped(address), RequirementSeed: declared.Named && declared.Text != "" && (class == "RequirementUsage" || class == "RequirementDefinition")})
		}
	}
	slices.SortFunc(entries, func(a, b DeclarationIndexEntry) int {
		if c := cmp.Compare(a.QualifiedName, b.QualifiedName); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return entries, nil
}
