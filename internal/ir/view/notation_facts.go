package view

import (
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// SemanticRef identifies an actual native declaration used by notation.
// Identity fields are populated only by the canonical export adapter.
type SemanticRef struct {
	Kind, Name, LifelineID, Direction         string
	SemanticID, QualifiedName, IdentityStatus string
	Origin                                    Origin
}

// Notation records native production and endpoint roles without changing EdgeKind.
type Notation struct {
	Production, DeclarationKind, Classification          string
	SourceRole, TargetRole, SourceEndName, TargetEndName string
	SourceEvent, TargetEvent                             *SemanticRef
}

// StateNotation preserves native state behavior compartments before text flattening.
type StateNotation struct {
	Parallel  bool
	Behaviors []StateBehaviorNotation
}

// StateBehaviorNotation describes a lowered state behavior, not a runtime instance.
type StateBehaviorNotation struct {
	Role, Text string
	SemanticRef
}

// SequenceOrder is an explicit native partial-order constraint, not an adjacent row.
type SequenceOrder struct {
	SemanticRef
	SourceEvent, TargetEvent *SemanticRef
}

func cloneRef(ref *SemanticRef) *SemanticRef {
	if ref == nil {
		return nil
	}
	copied := *ref
	return &copied
}
func cloneNotation(n *Notation) *Notation {
	if n == nil {
		return nil
	}
	copied := *n
	copied.SourceEvent, copied.TargetEvent = cloneRef(n.SourceEvent), cloneRef(n.TargetEvent)
	return &copied
}
func cloneState(n *StateNotation) *StateNotation {
	if n == nil {
		return nil
	}
	copied := *n
	copied.Behaviors = append([]StateBehaviorNotation(nil), n.Behaviors...)
	return &copied
}
func cloneSequenceOrders(orders []SequenceOrder) []SequenceOrder {
	out := append([]SequenceOrder(nil), orders...)
	for i := range out {
		out[i].SourceEvent = cloneRef(orders[i].SourceEvent)
		out[i].TargetEvent = cloneRef(orders[i].TargetEvent)
	}
	return out
}

// flowNotation classifies the native flow declaration, never its displayed label.
func flowNotation(decl ast.Node) *Notation {
	n := &Notation{Production: "flow", SourceRole: "source", TargetRole: "target"}
	if u, ok := decl.(*ast.Usage); ok {
		n.DeclarationKind = ast.Notation(u)
		if u.Keyword == "message" {
			n.Production = "message"
		} else if u.IsSuccessionFlow() {
			n.Production = "succession-flow"
		}
	}
	return n
}

func connectorNotation(sym *symbols.Symbol, kind EdgeKind) *Notation {
	if kind == EdgeFlow {
		return flowNotation(sym.Decl)
	}
	return nil
}
