package view

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// ErrInvalidFaultTree identifies an unsupported gate, malformed causation or cycle.
var ErrInvalidFaultTree = errors.New("invalid qualitative fault tree")

// failureKind classifies by semantic specialization, never by a project's names.
func (r *Renderer) failureKind(sym *symbols.Symbol) string {
	kinds := map[string]string{
		"FailureAnalysis::AndGate": "AND", "FailureAnalysis::OrGate": "OR",
		"FailureAnalysis::BasicEvent": "basic event", "FailureAnalysis::HouseEvent": "house event",
		"FailureAnalysis::UndevelopedEvent": "undeveloped event", "FailureAnalysis::Gate": "unsupported gate",
		"FailureAnalysis::Event": "event",
	}
	result := ""
	gate := ""
	for _, typ := range append([]*symbols.Symbol{sym}, r.model.AllSupertypes(sym)...) {
		kind := kinds[r.fqn(typ)]
		if kind == "AND" || kind == "OR" {
			if gate != "" && gate != kind {
				return "unsupported gate"
			}
			gate = kind
		}
		if kind != "" && (result == "" || result == "event") {
			result = kind
		}
	}
	if gate != "" {
		return gate
	}
	return result
}

func (r *Renderer) hasFailureType(sym *symbols.Symbol, name string) bool {
	for _, typ := range append([]*symbols.Symbol{sym}, r.model.AllSupertypes(sym)...) {
		if r.fqn(typ) == name {
			return true
		}
	}
	return false
}

func (r *Renderer) renderFaultTree(view *symbols.Symbol, exposed []*symbols.Symbol, out *Rendering) error {
	w := &featureWalk{r: r, view: view, ids: &nodeIDs{}, nodes: map[*symbols.Symbol]*Node{},
		pins: map[*Node]map[*symbols.Symbol]string{}, parent: map[*Node]*Node{}, pruned: map[*symbols.Symbol]bool{}, out: out}
	seen := map[*symbols.Symbol]bool{}
	var collect func(*symbols.Symbol) error
	collect = func(sym *symbols.Symbol) error {
		if r.libraryDeclared(sym) || seen[sym] {
			return nil
		}
		seen[sym] = true
		if r.hasFailureType(sym, "FailureAnalysis::Gate::inputs") {
			return nil
		}
		if u, ok := sym.Decl.(*ast.Usage); ok && u.IsReference && r.failureKind(sym) != "" {
			target, err := r.faultEventTarget(sym)
			if err != nil {
				return err
			}
			return collect(target)
		}
		// This declaration view cannot assign separate identities to inherited instances.
		if _, usage := sym.Decl.(*ast.Usage); usage {
			owned := map[*symbols.Symbol]bool{}
			for _, member := range r.containedMembers(sym) {
				owned[member] = true
			}
			for _, member := range r.model.MembersOf(sym) {
				if !owned[member] && !r.libraryDeclared(member) &&
					(r.failureKind(member) != "" || r.hasFailureType(member, "CausationConnections::Causation")) {
					return fmt.Errorf("%w: %s inherits failure members; expose their declarations explicitly or redeclare them in the usage", ErrInvalidFaultTree, r.notationName(sym))
				}
			}
		}
		kind := r.failureKind(sym)
		_, usage := sym.Decl.(*ast.Usage)
		switch {
		case usage && r.hasFailureType(sym, "CausationConnections::Causation"):
			w.connectors = append(w.connectors, sym)
		case usage && r.hasFailureType(sym, "CausationConnections::Multicausation"):
			return fmt.Errorf("%w: %s requires binary Causation connections", ErrInvalidFaultTree, r.notationName(sym))
		case kind != "" && usage:
			if kind == "unsupported gate" {
				return fmt.Errorf("%w: %s has no supported AND/OR gate type", ErrInvalidFaultTree, r.notationName(sym))
			}
			node := &Node{ID: w.ids.take(), Kind: kind, Name: r.notationName(sym), Type: declType(sym),
				Typings: r.declTypings(sym), Text: sym.ShortName, Origin: symbolOrigin(sym)}
			w.nodes[sym] = node
			w.drawn = append(w.drawn, sym)
			out.Roots = append(out.Roots, node)
			for _, member := range w.walkMembers(sym) {
				if err := collect(member); err != nil {
					return err
				}
			}
		default:
			for _, member := range w.walkMembers(sym) {
				if err := collect(member); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, sym := range exposed {
		if err := collect(sym); err != nil {
			return err
		}
	}
	gateInputs := map[string]map[string]bool{}
	for sym, node := range w.nodes {
		if node.Kind != "AND" && node.Kind != "OR" {
			continue
		}
		inputs, err := r.faultGateInputs(sym)
		if err != nil {
			return err
		}
		gateInputs[node.ID] = map[string]bool{}
		for _, input := range inputs {
			if inputNode := w.nodes[input]; inputNode != nil {
				gateInputs[node.ID][inputNode.ID] = true
			} else {
				node.Detail = "partial: one or more gate inputs are outside this view"
			}
		}
	}
	connected := map[string]map[string]bool{}
	for _, connector := range w.connectors {
		ends := r.model.ConnectorEndAttachments(connector)
		if len(ends) != 2 {
			return fmt.Errorf("%w: %s must connect one cause and one effect", ErrInvalidFaultTree, r.notationName(connector))
		}
		var cause, effect *Node
		for _, end := range ends {
			target, ok := r.resolver.ResolveTarget(connector.OwnerScope, end.Attachment)
			if !ok {
				return fmt.Errorf("%w: unresolved endpoint in %s", ErrInvalidFaultTree, r.notationName(connector))
			}
			if r.failureKind(target) != "" {
				var err error
				target, err = r.faultEventTarget(target)
				if err != nil {
					return err
				}
			}
			// Exact identities only: an unexposed child must never become its parent.
			node := w.nodes[target]
			role := end.Name
			if end.EndFeature != nil {
			roles:
				for _, base := range append([]*symbols.Symbol{end.EndFeature}, r.model.AllRedefinedFeatures(end.EndFeature)...) {
					switch r.fqn(base) {
					case "CausationConnections::Causation::theCause":
						role = "theCause"
						break roles
					case "CausationConnections::Causation::theEffect":
						role = "theEffect"
						break roles
					}
				}
			}
			switch role {
			case "theCause", "theCauses", "source":
				cause = node
			case "theEffect", "theEffects", "target":
				effect = node
			default:
				return fmt.Errorf("%w: cannot identify causal role of %s in %s", ErrInvalidFaultTree, role, r.notationName(connector))
			}
		}
		if cause == nil || effect == nil {
			out.Notices = append(out.Notices, fmt.Sprintf("%s has an endpoint outside the exposed failure events; no edge is drawn", r.notationName(connector)))
			continue
		}
		if expected, gate := gateInputs[effect.ID]; gate && !expected[cause.ID] {
			return fmt.Errorf("%w: %s connects a cause absent from the executable gate inputs", ErrInvalidFaultTree, r.notationName(connector))
		}
		if connected[effect.ID] == nil {
			connected[effect.ID] = map[string]bool{}
		}
		if connected[effect.ID][cause.ID] {
			return fmt.Errorf("%w: duplicate causal input in %s", ErrInvalidFaultTree, r.notationName(connector))
		}
		connected[effect.ID][cause.ID] = true
		out.Edges = append(out.Edges, Edge{From: cause.ID, To: effect.ID, Kind: EdgeConnection, Origin: symbolOrigin(connector)})
	}
	for _, node := range out.Roots {
		for input := range gateInputs[node.ID] {
			if !connected[node.ID][input] {
				node.Detail = "partial: one or more gate input connections are outside this view"
			}
		}
		if node.Detail != "" {
			out.Notices = append(out.Notices, node.Name+": "+node.Detail)
		}
	}
	// Shared causes are a DAG, not duplicate independent leaves. Cycles have no static-tree meaning.
	adjacency := map[string][]string{}
	for _, edge := range out.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	state := map[string]int{}
	var visit func(string) bool
	visit = func(id string) bool {
		if state[id] == 1 {
			return false
		}
		if state[id] == 2 {
			return true
		}
		state[id] = 1
		for _, to := range adjacency[id] {
			if !visit(to) {
				return false
			}
		}
		state[id] = 2
		return true
	}
	for _, node := range out.Roots {
		if !visit(node.ID) {
			return fmt.Errorf("%w: causal cycle at %s", ErrInvalidFaultTree, node.Name)
		}
	}
	return nil
}

// faultGateImage is a self-contained conventional AND/OR symbol; no external asset is loaded.
func faultGateImage(kind string) string {
	path := "M 10 54 L 10 28 A 22 22 0 0 1 54 28 L 54 54 Z"
	if kind == "OR" {
		path = "M 10 54 Q 14 20 32 6 Q 50 20 54 54 Q 32 43 10 54 Z"
	}
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="` + path + `" fill="white" stroke="black" stroke-width="2"/></svg>`
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

func (r *Rendering) writeFaultTreeMermaid(b *strings.Builder, direction Direction, labels labeller) {
	if direction == "" {
		direction = DirectionBottomTop
	}
	fmt.Fprintf(b, "flowchart %s\n", direction)
	gates := map[string]bool{}
	for _, node := range r.Roots {
		title := bareNames(labels.name(node))
		if node.Text != "" {
			title = bareNames(node.Text)
		}
		title = mermaidText(title)
		if node.Detail != "" {
			title += "<br/>[partial view]"
		}
		fmt.Fprintf(b, "  %s[\"%s\"]\n", node.ID, title)
		if node.Kind == "AND" || node.Kind == "OR" {
			gates[node.ID] = true
			fmt.Fprintf(b, "  %s_gate@{ img: \"%s\", label: \"%s\", pos: \"b\", h: 56, constraint: \"on\" }\n", node.ID, faultGateImage(node.Kind), node.Kind)
			fmt.Fprintf(b, "  style %s_gate fill:transparent,stroke:none\n", node.ID)
			fmt.Fprintf(b, "  %s_gate --> %s\n", node.ID, node.ID)
		}
	}
	for _, edge := range r.Edges {
		target := edge.To
		if gates[target] {
			target += "_gate"
		}
		fmt.Fprintf(b, "  %s --> %s\n", edge.From, target)
	}
}

// faultGateInputs reads the statically bound references defining an executable gate.
func (r *Renderer) faultGateInputs(gate *symbols.Symbol) ([]*symbols.Symbol, error) {
	for _, member := range r.model.MembersOf(gate) {
		if member.Name != "inputs" {
			continue
		}
		usage, ok := member.Decl.(*ast.Usage)
		if !ok || usage.Value == nil {
			continue
		}
		values := []ast.Node{usage.Value}
		if seq, ok := usage.Value.(*ast.SequenceExpr); ok {
			values = seq.Elements
		}
		var result []*symbols.Symbol
		for _, value := range values {
			target, ok := r.resolver.ResolveTarget(member.OwnerScope, value)
			if !ok || r.failureKind(target) == "" {
				return nil, fmt.Errorf("%w: %s requires statically bound event references", ErrInvalidFaultTree, r.notationName(gate))
			}
			target, err := r.faultEventTarget(target)
			if err != nil {
				return nil, err
			}
			result = append(result, target)
		}
		if len(result) == 0 {
			break
		}
		return result, nil
	}
	return nil, fmt.Errorf("%w: %s has no bound gate inputs", ErrInvalidFaultTree, r.notationName(gate))
}

// faultEventTarget preserves the identity of a single statically bound event reference.
func (r *Renderer) faultEventTarget(sym *symbols.Symbol) (*symbols.Symbol, error) {
	seen := map[*symbols.Symbol]bool{}
	for {
		u, ok := sym.Decl.(*ast.Usage)
		if !ok || !u.IsReference {
			return sym, nil
		}
		if seen[sym] {
			return nil, fmt.Errorf("%w: cyclic event reference %s", ErrInvalidFaultTree, r.notationName(sym))
		}
		seen[sym] = true
		target, resolved := r.resolver.ResolveTarget(sym.OwnerScope, u.Value)
		if !resolved || r.failureKind(target) == "" {
			return nil, fmt.Errorf("%w: %s requires a single statically bound event reference", ErrInvalidFaultTree, r.notationName(sym))
		}
		sym = target
	}
}
