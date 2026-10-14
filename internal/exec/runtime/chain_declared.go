package runtime

import (
	"errors"
	"slices"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// declaredChainValue reads `usage.member` from a usage no object stands for — a
// requirement, attribute or port usage at package level — through the values
// its members bind, as `usage::member` reads them. A feature value holds on
// every instance of its featuring type (KerML 1.0 §7.3.4.6), so for a usage of
// exactly one the chain is read from the object the usage would materialize:
// its bindings and defaults, the values its members inherit from what they
// redefine, and the nested redefinitions (`:>> details.limit = 9.0`) the usage
// and its types declare, all as materializing it gives them. The object is made
// under a probe and taken back with it, so the usage stays unmaterialized and
// a later read sees no occurrence. Reports whether the chain was read this way.
func (ec *EvalContext) declaredChainValue(base ast.Node, parts []ast.NameSegment) (Value, bool, error) {
	owner, ok := ec.chainBaseSymbol(base)
	if !ok || len(parts) == 0 || !ec.readsDeclaredMembers(owner) {
		return Value{}, false, nil
	}
	// A member whose value reaches back to a chain over its own usage is a cycle.
	if slices.Contains(ec.ctx.chainStack, owner) {
		return Value{}, true, &CyclicBindingError{Usage: owner, Stated: ec.ctx.qualifiedSymbolName(owner)}
	}
	ec.ctx.chainStack = append(ec.ctx.chainStack, owner)
	defer func() { ec.ctx.chainStack = ec.ctx.chainStack[:len(ec.ctx.chainStack)-1] }()
	defer ec.ctx.beginProbe()()
	inst, err := ec.ctx.materialize(owner, 0, nil, "")
	if err != nil {
		return Value{}, true, err
	}
	val, err := ec.chainMemberValue(Value{Kind: ValInstance, Instance: inst.ID}, parts, owner.Name)
	// A chain the object does not answer as a value — a member it leaves unbound
	// or does not have, one ending on a calc or on an object the probe takes
	// back — is not read here.
	if (err != nil && !errors.Is(err, ErrCyclicFeatureValue)) || val.Kind == ValInstance || val.Kind == ValFunction {
		return Value{}, false, nil
	}
	return val, true, err
}

// readsDeclaredMembers reports whether a chain through sym reads the values its
// members declare: a valueless namespace-level usage of exactly one that is
// neither a type, a calculation nor a variation awaiting its choice. A usage
// nested in a type or behavior is read on that owner's instances, never at
// model level.
func (ec *EvalContext) readsDeclaredMembers(sym *symbols.Symbol) bool {
	if sym.OwnerScope == nil || !namespaceScope(sym.OwnerScope) {
		return false
	}
	if _, usage := sym.Decl.(*ast.Usage); !usage || declaresType(sym) || isCalcUsageSymbol(sym) {
		return false
	}
	if ec.ctx.extractDefaultValue(sym) != nil || ec.ctx.model.semantics.IsVariationFeature(sym) {
		return false
	}
	n, exact := ec.ctx.featureMultiplicity(sym, nil).Exactly()
	return exact && n == 1
}
