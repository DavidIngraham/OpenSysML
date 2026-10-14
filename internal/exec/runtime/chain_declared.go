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
// first and leaves nothing here. The walk passes through valueless members of
// exactly one, reads a value a member inherits from what it redefines, and
// applies the nested redefinitions (`:>> details.limit = 9.0`) the usage and its
// types declare, as materializing the object would. Reports whether the chain
// was read this way.
func (ec *EvalContext) declaredChainValue(base ast.Node, parts []ast.NameSegment) (Value, bool, error) {
	owner, ok := ec.chainBaseSymbol(base)
	if !ok || len(parts) == 0 || !ec.readsDeclaredMembers(owner) {
		return Value{}, false, nil
	}
	var pending []pendingRedefinition
	for i, part := range parts {
		member, ok := ec.ctx.declaredMember(owner, part.Text)
		if !ok {
			return Value{}, false, nil
		}
		overrides, carry := ec.ctx.splitRedefinitions(pending)
		if redefining, ok := overrides[part.Text]; ok &&
			!(hasRedefines(member) && ec.ctx.blocksChain(member, redefining)) {
			member = redefining
		}
		value, decl := ec.ctx.extractDefaultValue(member), member
		if value == nil {
			value, decl = ec.ctx.redefinedDefault(member, owner)
		}
		below := ec.ctx.declaredNestedRedefinitions(owner, part.Text, carry)
		// A valued chain from a more specific body governs the member's
		// inherited binding, as it does on a materialized object.
		if value != nil && slices.ContainsFunc(below, func(p pendingRedefinition) bool {
			return valuedChain(p) && ec.ctx.chainGovernsValue(p.sym, decl)
		}) {
			value = nil
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
		owner, pending = member, below
	}
	return Value{}, false, nil
}

// declaredMemberValue evaluates the value decl declares, reporting a value that
// reaches back to a member the chain is already reading as a cycle.
func (ec *EvalContext) declaredMemberValue(decl *symbols.Symbol, value ast.Node) (Value, error) {
	if slices.Contains(ec.ctx.chainStack, decl) {
		return Value{}, &CyclicBindingError{Usage: decl, Stated: ec.ctx.qualifiedSymbolName(decl)}
	}
	ec.ctx.chainStack = append(ec.ctx.chainStack, decl)
	defer func() { ec.ctx.chainStack = ec.ctx.chainStack[:len(ec.ctx.chainStack)-1] }()
	return ec.declaredValue(decl, value)
}

// declaredNestedRedefinitions lists the nested redefinitions reaching below the
// member feature of owner: the tails carried down from above, plus the chains
// owner and its member sources declare whose first segment names feature under
// any of its names (the symbol-level counterpart of pendingNestedRedefinitions).
func (ctx *Context) declaredNestedRedefinitions(owner *symbols.Symbol, feature string, carry []pendingRedefinition) []pendingRedefinition {
	names := map[string]bool{feature: true}
	for _, group := range ctx.redefinitionGroups(owner) {
		if slices.Contains(group, feature) {
			for _, name := range group {
				names[name] = true
			}
		}
	}
	var out []pendingRedefinition
	for _, p := range carry {
		if names[p.rest[0]] {
			out = append(out, pendingRedefinition{rest: p.rest[1:], sym: p.sym})
		}
	}
	for _, src := range append([]*symbols.Symbol{owner}, ctx.model.semantics.MemberSources(owner)...) {
		for _, nr := range ctx.model.semantics.NestedRedefinitionsOf(src) {
			if len(nr.Path) > 1 && names[nr.Path[0]] {
				out = append(out, pendingRedefinition{rest: nr.Path[1:], sym: nr.Feature})
			}
		}
	}
	return out
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
