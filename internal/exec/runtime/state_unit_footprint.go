package runtime

import (
	"github.com/Open-MBEE/OpenSysML/internal/ir/lower"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
)

// unit is one queue of a unit front: the body performing its units and what
// they may touch, by which the draws among queues are classified.
type unit struct {
	run       func() error
	footprint lower.Footprint
}

// unknownFootprint stands for units whose touches are not decidable from the
// lowered graph; it conflicts with every alternative, so a draw over it is
// reported as order-dependent rather than silently taken for independent.
func unknownFootprint() lower.Footprint {
	return lower.Footprint{Dynamic: true}
}

// behaviorsFootprint is what running the behaviors touches.
func (e *StateExecutor) behaviorsFootprint(behaviors []lower.StateBehavior) lower.Footprint {
	var fp lower.Footprint
	footprints := e.graph.BehaviorFootprints()
	for _, behavior := range behaviors {
		b, ok := footprints[behavior.Node]
		if !ok {
			return unknownFootprint()
		}
		fp = unionFootprints(fp, b)
	}
	return fp
}

// exitFootprint is what exiting the states touches: their exit behaviors.
func (e *StateExecutor) exitFootprint(states []*ast.StateNode) lower.Footprint {
	var fp lower.Footprint
	for _, state := range states {
		if behaviors := e.graph.Behaviors[state]; behaviors != nil {
			fp = unionFootprints(fp, e.behaviorsFootprint(behaviors.Exit))
		}
	}
	return fp
}

// enterFootprint is what entering state and starting what it does touches.
func (e *StateExecutor) enterFootprint(state *ast.StateNode) lower.Footprint {
	behaviors := e.graph.Behaviors[state]
	if behaviors == nil {
		return lower.Footprint{}
	}
	return unionFootprints(e.behaviorsFootprint(behaviors.Entry), e.behaviorsFootprint(behaviors.Do))
}

// regionEntryFootprint is what entering the region may touch, whichever state
// it starts in: entering and starting any state under it, the effects along its
// entry transitions and those of the pseudostates it crosses, and the effects
// routed to it.
func (e *StateExecutor) regionEntryFootprint(entry *regionEntry) lower.Footprint {
	var fp lower.Footprint
	for _, eff := range entry.effects {
		fp = unionFootprints(fp, e.behaviorsFootprint([]lower.StateBehavior{eff.behavior}))
	}
	transitions := e.graph.TransitionFootprints()
	var regions []*ast.StateRegion
	for regions = append(regions, entry.region); len(regions) > 0; {
		region := regions[len(regions)-1]
		regions = regions[:len(regions)-1]
		if owner := e.graph.RegionState[region]; owner != nil {
			fp = unionFootprints(fp, e.enterFootprint(owner))
		}
		for _, start := range e.graph.StartOf(region) {
			fp = unionFootprints(fp, e.behaviorsFootprint(start.Effect))
		}
		for _, node := range region.States {
			switch v := node.(type) {
			case *ast.StateNode:
				fp = unionFootprints(fp, e.enterFootprint(v))
				for _, start := range e.graph.StartOf(v) {
					fp = unionFootprints(fp, e.behaviorsFootprint(start.Effect))
				}
				regions = append(regions, e.graph.CompositeStates[v]...)
			case *ast.PseudostateNode:
				for _, trans := range e.graph.Transitions[v] {
					fp = unionFootprints(fp, transitions[trans])
				}
			}
		}
	}
	return fp
}

// headFootprint is what the unit a head names may touch: entering a state and
// starting what it does, one token move of a do behavior, or, for a unit the
// graph does not name, whatever it touches.
func (e *StateExecutor) headFootprint(head unitHead) lower.Footprint {
	state, ok := head.at.(*ast.StateNode)
	if !ok {
		return unknownFootprint()
	}
	if head.doStep {
		return doStepFootprint(e, state)
	}
	return e.enterFootprint(state)
}
