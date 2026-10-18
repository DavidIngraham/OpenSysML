//go:build !js && !wasip1

// Package renderidentity binds native drawings to native conversion identities.
package renderidentity

import (
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
	"github.com/Open-MBEE/OpenSysML/internal/translate/export"
)

// Bind annotates declared nodes and edges; synthetic artifacts remain unidentified.
func Bind(data *view.Data, documents []Document, index *symbols.Index) error {
	var ids *export.ElementIDs
	var err error
	if len(documents) == 1 {
		ids, err = export.NewElementIDsOfDocument(documents[0].File.Name(), documents[0].File.Bytes())
	} else {
		converted := make([]export.ModelDocument, len(documents))
		for i, document := range documents {
			converted[i] = export.ModelDocument{File: document.File, Root: document.Root}
		}
		ids, err = export.NewElementIDs(converted)
	}
	if err != nil {
		return err
	}
	library := map[symbols.Origin]export.ElementIdentity{}
	if index != nil {
		index.Registered(func(_ string, syms []*symbols.Symbol) {
			for _, sym := range syms {
				if !index.Library(sym) {
					continue
				}
				origin := sym.Origin()
				origin.Name = source.Span{}
				if !origin.Located() {
					continue
				}
				name := index.GetFQN(sym)
				if id, ok := ids.Of(name); ok {
					library[origin] = export.ElementIdentity{ID: id, QualifiedName: name}
				}
			}
		})
	}
	lookup := func(origin symbols.Origin) (export.ElementIdentity, bool) {
		if identity, ok := ids.OfOrigin(origin); ok {
			return identity, true
		}
		origin.Name = source.Span{}
		identity, ok := library[origin]
		return identity, ok
	}
	for i := range data.Nodes {
		node := &data.Nodes[i]
		node.IdentityStatus = "unmapped"
		if node.Synthetic || node.StandIn || (node.NameSynthesized && !node.Origin.Located()) {
			node.IdentityStatus = "synthetic"
			continue
		}
		if !node.StandIn {
			if identity, ok := lookup(node.Origin); ok {
				node.SemanticID, node.QualifiedName = identity.ID, identity.QualifiedName
				node.IdentityStatus = "mapped"
			}
		}
	}
	for i := range data.Edges {
		edge := &data.Edges[i]
		edge.IdentityStatus = "unmapped"
		if identity, ok := lookup(edge.Origin); ok {
			edge.SemanticID, edge.QualifiedName = identity.ID, identity.QualifiedName
			edge.IdentityStatus = "mapped"
		}
	}
	for i := range data.Rows {
		row := &data.Rows[i]
		row.IdentityStatus = "unmapped"
		if identity, ok := lookup(row.Origin); ok {
			row.SemanticID, row.QualifiedName = identity.ID, identity.QualifiedName
			row.IdentityStatus = "mapped"
		}
	}
	return nil
}
