package runtime

import (
	"slices"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// declaredChainValue reads `usage.member` from a usage no object stands for — a
// requirement, attribute or port usage at package level — through the value the
// member binds or inherits, as `usage::member` reads it. A feature value holds on
// every instance of its featuring type (KerML 1.0 §7.3.4.6), so for a usage of
// exactly one it is the chain's value; the usage's own value, when it has one, was
// read first and leaves nothing here. The chain may pass through valueless members
// of exactly one (`req.details.limit`) before reaching the member that binds a
// value. Reports whether the chain was read this way.
func (ec *EvalContext) declaredChainValue(base ast.Node, parts []ast.NameSegment) (Value, bool, error) {
	owner, ok := ec.chainBaseSymbol(base)
	if !ok || len(parts) == 0 || !ec.readsDeclaredMembers(owner) {
		return Value{}, false, nil
	}
	for i, part := range parts {
		member, ok := ec.ctx.declaredMember(owner, part.Text)
		if !ok {
			return Value{}, false, nil
		}
		value, decl := ec.ctx.extractDefaultValue(member), member
		if value == nil {
			value, decl = ec.ctx.redefinedDefault(member, owner)
		}
		if value != nil {
			val, err := ec.declaredMemberValue(decl, value)
			if err != nil {
				return Value{}, true, err
			}
			val, err = ec.chainMemberValue(val, parts[i+1:], member.Name)
			return val, true, err
		}
		if !ec.chainsDeclaredMembers(member) {
			return Value{}, false, nil
		}
		owner = member
	}
	return Value{}, false, nil
}

// declaredMemberValue evaluates the value decl binds for a chain, reporting a
// value that reaches back to itself as a cycle instead of recursing.
func (ec *EvalContext) declaredMemberValue(decl *symbols.Symbol, value ast.Node) (Value, error) {
	if slices.Contains(ec.ctx.chainStack, decl) {
		return Value{}, &CyclicBindingError{Usage: decl, Stated: ec.ctx.qualifiedSymbolName(decl)}
	}
	ec.ctx.chainStack = append(ec.ctx.chainStack, decl)
	defer func() { ec.ctx.chainStack = ec.ctx.chainStack[:len(ec.ctx.chainStack)-1] }()
	return ec.declaredValue(decl, value)
}

// readsDeclaredMembers reports whether a chain through sym reads the values its
// members declare: a valueless namespace-level usage of exactly one that is
// neither a type, a calculation nor a variation awaiting its choice. A usage
// nested in a type or behavior is read on that owner's instances, never at
// model level.
func (ec *EvalContext) readsDeclaredMembers(sym *symbols.Symbol) bool {
	return sym.OwnerScope != nil && namespaceScope(sym.OwnerScope) && ec.chainsDeclaredMembers(sym)
}

// chainsDeclaredMembers reports whether a chain read at model level passes
// through sym to the values its members declare: sym is a valueless usage of
// exactly one, neither a type, a calculation nor a variation awaiting its choice.
func (ec *EvalContext) chainsDeclaredMembers(sym *symbols.Symbol) bool {
	if _, usage := sym.Decl.(*ast.Usage); !usage || declaresType(sym) || isCalcUsageSymbol(sym) {
		return false
	}
	if ec.ctx.extractDefaultValue(sym) != nil || ec.ctx.model.semantics.IsVariationFeature(sym) {
		return false
	}
	n, exact := ec.ctx.featureMultiplicity(sym, nil).Exactly()
	return exact && n == 1
}
