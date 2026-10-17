# Qualitative failure analysis

`FailureAnalysis` is a non-normative, RAAML-aligned SysML v2 library. It reuses
RAAML's basic-event, house-event, undeveloped-event and AND/OR distinctions;
it is not an import of RAAML's UML profile or a claim of RAAML conformance.
See [OMG RAAML 1.1](https://www.omg.org/spec/RAAML/1.1).

`Event.present` is unbound until supplied. `OrGate` and `AndGate` compute it
from their `inputs` references. These are propositions about one declared
exposure interval: conjunction implies neither temporal ordering nor statistical
independence. The library assigns no probabilities or failure rates.

Declare causal links with standard `CauseAndEffect::Causation` connections.
Specialize `FailureAnalysisViews::FaultTreeView` and expose the event usages and
connections to draw. Gate types, including specializations, determine the symbols;
model names do not. A short name supplies the display label. Shared causes remain
one node with multiple outgoing edges.

The native `fault-tree` rendering provides text/IR and Mermaid output with embedded
SVG AND/OR symbols. It requires a Mermaid release supporting image nodes (11.3+).
Other event kinds currently use labelled rectangles, not a complete fault-tree
symbol alphabet. DOT, PlantUML and D2 return a typed unsupported-form error.
Layout, palette and style decorations are not applied to this rendering.

This is a declaration view. Owned nested events are flattened, while `inputs` references are not extra event nodes. Single statically bound
event references share their target node; unresolved or computed references fail. Instantiated inherited
failure members are rejected: expose their declarations explicitly or redeclare
them in the usage. Gate inputs must be statically bound event references (one
reference or a sequence), not computed expressions. Binary causation is supported;
multicausation, ambiguous AND/OR specializations, cycles, duplicate inputs and
links contradicting executable inputs are rejected. Inputs or links omitted by
the view produce notices and a visible **partial view** label on the gate.
A partial view is not evidence that the complete model has no other causes.

`ReviewRow` records the function, affected part, causal events, detection proposal,
action, owner and evidence disposition. `Review` groups rows for any system.
`DesignFailureReview` reports counts; its readiness result requires nonempty,
reviewed scope and accepted disposition/evidence for every safety-critical row.
It does not certify physical safety or close every noncritical action.

Specialize `FailureAnalysisViews::ReviewRows` and bind `reviewName` to a review's
qualified name. The query selects descendant row declarations by semantic type,
including specializations, and projects referenced descriptions and actions.
New owned rows appear without editing a list of row names. As a declaration
query it does not expand inherited runtime instances. A native `DocumentQueries`
Table can consume this query with no application-specific renderer.

[The redundant-power example](../../examples/fault-tree.sysml) demonstrates the
shared cause, both gates, a local view specialization and a reusable DFMEA query:

```sh
sysml examples/fault-tree.sysml -render PowerSupply::tree
sysml examples/fault-tree.sysml -run-query PowerReview::Rows
sysml examples/fault-tree.sysml -render-document PowerReport::Report
```
