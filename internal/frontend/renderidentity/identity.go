// Package renderidentity binds drawings to conversion identities on every platform.
package renderidentity

import (
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/resolve"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/semantics"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
	"github.com/Open-MBEE/OpenSysML/internal/translate/export"
)

// Inventory is a lightweight native identity inventory with no emitted RDF graph.
type Inventory struct{ ids *export.ElementIDs }

// NewInventory reuses native semantic identity rules without serializing the model.
func NewInventory(documents []Document, res *resolve.Resolver, model *semantics.Model) (*Inventory, error) {
	converted := make([]export.ModelDocument, len(documents))
	for i, document := range documents {
		converted[i] = export.ModelDocument{File: document.File, Root: document.Root}
	}
	ids, err := export.NewDeclarationElementIDsInModel(converted, res, model)
	if err != nil {
		return nil, err
	}
	return &Inventory{ids: ids}, nil
}

// Bind annotates declared nodes, edges and grid records; synthetic artifacts remain unidentified.
func Bind(data *view.Data, documents []Document, index *symbols.Index) error {
	inventory, err := NewInventory(documents, nil, nil)
	if err != nil {
		return err
	}
	return BindInventory(data, inventory, index)
}

// BindInventory reuses the loaded model's canonical declaration inventory.
func BindInventory(data *view.Data, inventory *Inventory, index *symbols.Index) error {
	ids := inventory.ids
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
	bindRef := func(ref *view.SemanticRef) {
		if ref == nil {
			return
		}
		ref.IdentityStatus = "unmapped"
		if identity, ok := lookup(ref.Origin); ok {
			ref.SemanticID, ref.QualifiedName, ref.IdentityStatus = identity.ID, identity.QualifiedName, "mapped"
		}
	}
	for i := range data.Nodes {
		node := &data.Nodes[i]
		for j := range node.Ports {
			port := &node.Ports[j]
			port.IdentityStatus = "unmapped"
			if identity, ok := lookup(port.Origin); ok {
				port.SemanticID, port.QualifiedName, port.IdentityStatus = identity.ID, identity.QualifiedName, "mapped"
			}
		}
		if node.State != nil {
			for j := range node.State.Behaviors {
				bindRef(&node.State.Behaviors[j].SemanticRef)
			}
		}
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
		if edge.Notation != nil {
			bindRef(edge.Notation.SourceEvent)
			bindRef(edge.Notation.TargetEvent)
		}
		edge.IdentityStatus = "unmapped"
		if identity, ok := lookup(edge.Origin); ok {
			edge.SemanticID, edge.QualifiedName = identity.ID, identity.QualifiedName
			edge.IdentityStatus = "mapped"
		}
	}
	for i := range data.SequenceOrders {
		order := &data.SequenceOrders[i]
		bindRef(&order.SemanticRef)
		bindRef(order.SourceEvent)
		bindRef(order.TargetEvent)
	}
	for i := range data.Rows {
		row := &data.Rows[i]
		row.IdentityStatus = "unmapped"
		if identity, ok := lookup(row.Origin); ok {
			row.SemanticID, row.QualifiedName = identity.ID, identity.QualifiedName
			row.IdentityStatus = "mapped"
		}
	}
	for i := range data.ColumnElements {
		column := &data.ColumnElements[i]
		column.IdentityStatus = "unmapped"
		if identity, ok := lookup(column.Origin); ok {
			column.SemanticID, column.QualifiedName = identity.ID, identity.QualifiedName
			column.IdentityStatus = "mapped"
		}
	}
	for i := range data.RelationshipCells {
		for j := range data.RelationshipCells[i].Relationships {
			relation := &data.RelationshipCells[i].Relationships[j]
			relation.IdentityStatus = "unmapped"
			if identity, ok := lookup(relation.Origin); ok {
				relation.SemanticID, relation.QualifiedName = identity.ID, identity.QualifiedName
				relation.IdentityStatus = "mapped"
			}
			if identity, ok := lookup(relation.SourceOrigin); ok {
				relation.SourceSemanticID = identity.ID
			}
			if identity, ok := lookup(relation.TargetOrigin); ok {
				relation.TargetSemanticID = identity.ID
			}
		}
	}
	return nil
}
