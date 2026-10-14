package runtime

import (
	"slices"
	"sort"
	"strings"

	"github.com/Open-MBEE/OpenSysML/internal/ir/lower"
)

// A default run resolves each choice point by a fixed rule. Whether that rule
// decided anything is read off the alternatives' footprints, the same
// projection the checker's reduction commutes moves by (check_reduce.go): two
// alternatives neither of which writes what the other reads or writes, nor meet
// on the bus, are one outcome in any order, and the choice is Independent. Any
// other pair leaves the features its alternatives write standing at a value
// another order may change; those are the choice's Touches.

// classifyChoice sets c.Independent and c.Touches for a choice of the order of
// its alternatives, from their footprints, one per alternative in order; rel is
// what is known of the executors alternatives i and j belong to, nil for one
// executor's. Footprints not matching the alternatives leave the choice
// conflicting, touching nothing named.
func classifyChoice(c *ChoicePoint, fps []lower.Footprint, rel func(i, j int) lower.Relation) {
	if len(fps) != len(c.Alternatives) || len(fps) < 2 {
		return
	}
	found := false
	for i := range fps {
		for j := i + 1; j < len(fps); j++ {
			var r lower.Relation
			if rel != nil {
				r = rel(i, j)
			}
			if !onValues(fps[i]).DependentBy(onValues(fps[j]), r) {
				continue
			}
			found = true
			c.touchMeeting(fps[i], fps[j], r)
		}
	}
	c.Independent = !found
}

// onValues is the footprint less what the order of two moves changes nothing of
// in a run's values: meeting at a control node (a join fires once whichever
// token arrives first) or both completing a performance.
func onValues(fp lower.Footprint) lower.Footprint {
	fp.Control, fp.Completion = nil, false
	return fp
}

// touchMeeting adds the writes of two dependent alternatives that meet a place of
// the other's; when none does (the alternatives meet on the bus, on control, or
// one is dynamic) every write of both may stand at an order-dependent value.
func (c *ChoicePoint) touchMeeting(f, g lower.Footprint, rel lower.Relation) {
	met := false
	for _, p := range f.Writes {
		if rel.PlaceMeets(p, g.Reads) || rel.PlaceMeets(p, g.Writes) {
			c.touch(p)
			met = true
		}
	}
	for _, q := range g.Writes {
		if rel.PlaceMeets(q, f.Reads) || rel.PlaceMeets(q, f.Writes) {
			c.touch(q)
			met = true
		}
	}
	if !met {
		c.touchWrites(f, g)
	}
}

// classifyExclusive marks a choice among alternatives of which one runs and the
// others do not (a decision's branches, a state's transitions): never
// independent, since the alternatives are different continuations, and touching
// whatever any of them writes.
func classifyExclusive(c *ChoicePoint, fps []lower.Footprint) {
	c.Independent = false
	c.touchWrites(fps...)
}

func (c *ChoicePoint) touchWrites(fps ...lower.Footprint) {
	for _, fp := range fps {
		for _, p := range fp.Writes {
			c.touch(p)
		}
	}
}

func (c *ChoicePoint) touch(p lower.Place) {
	if p.Local || p.State != nil {
		return
	}
	name := p.String()
	i := sort.SearchStrings(c.Touches, name)
	if i < len(c.Touches) && c.Touches[i] == name {
		return
	}
	c.Touches = slices.Insert(c.Touches, i, name)
}

// futureFootprints is what the run's executors may still touch, made on first use.
func (ctx *Context) futureFootprints() *futureFootprints {
	if ctx.futures == nil {
		ctx.futures = newFutureFootprints(ctx)
	}
	return ctx.futures
}

// touchesFeature reports whether the choice may have left the feature's value
// order-dependent: a touch naming it, alone or as the last segment of a path.
func (c ChoicePoint) touchesFeature(name string) bool {
	for _, t := range c.Touches {
		if t == name || strings.HasSuffix(t, "."+name) {
			return true
		}
	}
	return false
}

// OrderDependence is what a run's unresolved choice points leave open: how many
// there are, and the features they may have left at an order-dependent value.
type OrderDependence struct {
	// Unresolved counts the choice points whose alternatives may not commute.
	Unresolved int
	// Features maps a feature a conflicting alternative wrote to the number of
	// unresolved choice points touching it.
	Features map[string]int
}

// Dependent reports how many unresolved choice points touch the feature; 0 for a
// value nothing order-dependent wrote.
func (d OrderDependence) Dependent(feature string) int {
	return d.Features[feature]
}

// orderDependenceOf folds the choice points among notes.
func orderDependenceOf(notes []RunNote) OrderDependence {
	d := OrderDependence{Features: make(map[string]int)}
	for _, n := range notes {
		c, ok := n.(ChoicePoint)
		if !ok || c.Independent {
			continue
		}
		d.Unresolved++
		for _, t := range c.Touches {
			d.Features[t]++
		}
	}
	return d
}

// OrderDependence is what the run's choice points so far leave open.
func (ctx *Context) OrderDependence() OrderDependence {
	return orderDependenceOf(ctx.run.notes)
}

// OrderDependent reports how many unresolved choice points of the run touch the
// feature, by its name alone or as the last segment of a written path.
func (ctx *Context) OrderDependent(feature string) int {
	count := 0
	for _, n := range ctx.run.notes {
		if c, ok := n.(ChoicePoint); ok && !c.Independent && c.touchesFeature(feature) {
			count++
		}
	}
	return count
}

// UnresolvedChoices counts the run's choice points whose alternatives may not
// commute: the ones a default run's result stands only up to.
func (ctx *Context) UnresolvedChoices() int {
	return orderDependenceOf(ctx.run.notes).Unresolved
}
