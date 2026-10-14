package runtime

import (
	"slices"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// declaredChainValue reads `usage.member` from a usage no object stands for — a
// requirement, attribute or port usage at package level — through the value the
// member binds, as `usage::member` reads it. A feature value holds on every
// instance of its featuring type (KerML 1.0 §7.3.4.6), so for a usage of exactly
// one it is the chain's value; the usage's own value, when it has one, was read
// first and leaves nothing here. Reports whether the chain was read this way.
func (ec *EvalContext) declaredChainValue(base ast.Node, parts []ast.NameSegment) (Value, bool, error) {
	sym, ok := ec.chainBaseSymbol(base)
	if !ok || len(parts) == 0 || !ec.readsDeclaredMembers(sym) {
		return Value{}, false, nil
	}
	member, ok := ec.ctx.declaredMember(sym, parts[0].Text)
	if !ok {
		return Value{}, false, nil
	}
	value := ec.ctx.extractDefaultValue(member)
	if value == nil {
		return Value{}, false, nil
	}
	if slices.Contains(ec.ctx.chainStack, member) {
		return Value{}, true, &CyclicBindingError{Usage: member, Stated: ec.ctx.qualifiedSymbolName(member)}
	}
	ec.ctx.chainStack = append(ec.ctx.chainStack, member)
	val, err := ec.declaredValue(member, value)
	ec.ctx.chainStack = ec.ctx.chainStack[:len(ec.ctx.chainStack)-1]
	if err != nil {
		return Value{}, true, err
	}
	val, err = ec.chainMemberValue(val, parts[1:], member.Name)
	return val, true, err
}

// readsDeclaredMembers reports whether a chain through sym reads the values its
// members declare: a valueless package-level usage of exactly one that is neither
// a type, a calculation nor a variation awaiting its choice. A usage nested in a
// type or behavior is read on that owner's instances, never at model level.
func (ec *EvalContext) readsDeclaredMembers(sym *symbols.Symbol) bool {
	if _, usage := sym.Decl.(*ast.Usage); !usage || declaresType(sym) || isCalcUsageSymbol(sym) {
		return false
	}
	if sym.OwnerScope == nil || !namespaceScope(sym.OwnerScope) {
		return false
	}
	if ec.ctx.extractDefaultValue(sym) != nil || ec.ctx.model.semantics.IsVariationFeature(sym) {
		return false
	}
	n, exact := ec.ctx.featureMultiplicity(sym, nil).Exactly()
	return exact && n == 1
}
