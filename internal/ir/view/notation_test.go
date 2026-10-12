package view

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

// The optional notation elements are named, "all" names every one, and a name
// none has is refused on every form by an error naming the elements there are.
func TestNotationElementsAreNamed(t *testing.T) {
	if got := ElementNames(); got != "parameter-symbols, port-symbols, compartments, edge-labels, keywords, multiplicities, typings, qualifiers, all" {
		t.Errorf("ElementNames() = %q", got)
	}
	shown, err := ParseShown("edge-labels, parameter-symbols")
	if err != nil || !shown.Has(ElementEdgeLabels) || !shown.Has(ElementParameterSymbols) || shown.Has(ElementPortSymbols) {
		t.Errorf("ParseShown(edge-labels, parameter-symbols) = %v, %v", shown, err)
	}
	if shown, err := ParseShown(""); err != nil || shown != nil {
		t.Errorf("ParseShown(\"\") = %v, %v, want nil: the default", shown, err)
	}
	if all, err := ParseShown("all"); err != nil || len(all) != len(Elements()) || all.String() != "all" {
		t.Errorf("ParseShown(all) = %v, %v", all, err)
	}
	_, err = ParseShown("edge-labels,labels")
	var unknown *UnknownElementError
	if !errors.As(err, &unknown) || !errors.Is(err, ErrUnknownElement) || unknown.Name != "labels" {
		t.Fatalf("ParseShown(labels): err = %v, want an UnknownElementError", err)
	}
	if !strings.Contains(err.Error(), ElementNames()) {
		t.Errorf("%v does not name the elements", err)
	}
	for _, element := range Elements() {
		if got, ok := ElementOfAttribute(element.Attribute()); !ok || got != element {
			t.Errorf("ElementOfAttribute(%s) = %q, %v", element.Attribute(), got, ok)
		}
	}
	if got := ElementEdgeLabels.Attribute(); got != "showEdgeLabels" {
		t.Errorf("Attribute() = %q", got)
	}
	if got := DefaultShown.Names(); strings.Join(got, ",") != "port-symbols,compartments" {
		t.Errorf("DefaultShown = %v", got)
	}
}

// useCaseEnds is where each connection of the use-case fixture ends, by the
// actor it starts from: the node ID and the pin ID, "" for an end at the node.
func useCaseEnds(t *testing.T, rendering *Rendering) map[string][2]string {
	t.Helper()
	byID := map[string]*Node{}
	for _, node := range everyNode(rendering.Roots) {
		byID[node.ID] = node
	}
	ends := map[string][2]string{}
	for _, edge := range rendering.Edges {
		if edge.Kind != EdgeConnection {
			continue
		}
		ends[byID[edge.From].Name] = [2]string{edge.To, edge.ToPort}
	}
	if len(ends) != 2 {
		t.Fatalf("edges = %+v, want the two connections", rendering.Edges)
	}
	return ends
}

// By default a use case's actor parameters are left out: the actors are drawn
// once each, the use case has no parameter symbol, and each connection ends
// at the use case node, one line per connection, in every form.
func TestParameterSymbolsAreLeftOutByDefault(t *testing.T) {
	rendering := render(t, "use-cases.sysml", "LabUseCases::useCases")
	if rendering.Shown != nil {
		t.Errorf("Shown = %v, want nil: the view's rendering states none", rendering.Shown)
	}
	for _, form := range []Form{FormText, FormDot, FormMermaid, FormPlantUML, FormD2} {
		out, err := rendering.WriteWith(form, Options{})
		if err != nil {
			t.Fatalf("%s: %v", form, err)
		}
		for _, parameter := range []string{"operator", "administrator"} {
			if strings.Contains(out, parameter) {
				t.Errorf("%s draws the parameter %s:\n%s", form, parameter, out)
			}
		}
		for _, actor := range []string{"Operator", "Administrator"} {
			if n := len(regexp.MustCompile(`\b`+actor+`\b`).FindAllString(out, -1)); n == 0 {
				t.Errorf("%s does not draw the actor %s:\n%s", form, actor, out)
			}
		}
	}
	shown := rendering.showing(nil)
	byName := nodesByName(shown)
	useCase := byName["LabUseCases::'Configure Instrument'"]
	if useCase == nil || len(useCase.Ports) != 0 || len(useCase.Children) != 0 {
		t.Fatalf("use case = %+v, want one with no parameter symbol", useCase)
	}
	if len(shown.Roots) != 3 {
		t.Errorf("roots = %d, want the two actors and the use case", len(shown.Roots))
	}
	for name, end := range useCaseEnds(t, shown) {
		if end != [2]string{useCase.ID, ""} {
			t.Errorf("%s ends at %v, want the use case node %s", name, end, useCase.ID)
		}
	}
	if len(shown.Edges) != 2 {
		t.Errorf("edges = %+v, want one line per connection", shown.Edges)
	}
}

// Shown — by a rendering declared in the model with showParameterSymbols true,
// or by Options adding the element — the actor parameters are symbols on the
// use case's boundary, and each connection ends at its parameter.
func TestParameterSymbolsAreShownWhenAsked(t *testing.T) {
	cases := map[string]struct {
		view    string
		options Options
	}{
		"by the model's rendering":    {"LabUseCases::useCasesWithParameters", Options{}},
		"by an inherited rendering":   {"LabUseCases::useCasesInheritingParameters", Options{}},
		"by Options":                  {"LabUseCases::useCases", Options{Show: Shown{ElementParameterSymbols: true}}},
		"by Options over the model's": {"LabUseCases::useCasesWithoutParameters", Options{Show: Shown{ElementParameterSymbols: true}}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rendering := render(t, "use-cases.sysml", tc.view)
			shown := rendering.showing(tc.options.Show)
			useCase := nodesByName(shown)["LabUseCases::'Configure Instrument'"]
			if useCase == nil || len(useCase.Children) != 3 {
				t.Fatalf("use case = %+v, want one with its subject and two actors as parameter symbols", useCase)
			}
			byActor := map[string]string{}
			for _, child := range useCase.Children {
				if child.Kind == "actor" {
					byActor[child.Name] = child.ID
				}
			}
			if len(byActor) != 2 {
				t.Fatalf("actors = %v, want operator and administrator", byActor)
			}
			ends := useCaseEnds(t, shown)
			if got := ends["LabUseCases::Operator"]; got != [2]string{byActor["operator"], ""} {
				t.Errorf("the operator's connection ends at %v, want the operator parameter %s", got, byActor["operator"])
			}
			if got := ends["LabUseCases::Administrator"]; got != [2]string{byActor["administrator"], ""} {
				t.Errorf("the administrator's connection ends at %v, want the administrator parameter %s", got, byActor["administrator"])
			}
			for _, form := range []Form{FormText, FormDot, FormMermaid, FormPlantUML, FormD2} {
				out, err := rendering.WriteWith(form, tc.options)
				if err != nil {
					t.Fatalf("%s: %v", form, err)
				}
				if !strings.Contains(out, "operator") || !strings.Contains(out, "administrator") {
					t.Errorf("%s does not draw the parameters:\n%s", form, out)
				}
			}
		})
	}
}

// A rendering turning the parameter symbols off again under one turning them
// on leaves them out: the nearest declaration decides.
func TestTheNearestRenderingDecidesAnElement(t *testing.T) {
	rendering := render(t, "use-cases.sysml", "LabUseCases::useCasesWithoutParameters")
	if rendering.Shown == nil || rendering.Shown.Has(ElementParameterSymbols) || !rendering.Shown.Has(ElementPortSymbols) {
		t.Fatalf("Shown = %v, want the default without parameter symbols", rendering.Shown)
	}
	useCase := nodesByName(rendering.showing(nil))["LabUseCases::'Configure Instrument'"]
	if useCase == nil || len(useCase.Children) != 0 {
		t.Errorf("use case = %+v, want one with no parameter symbol", useCase)
	}
}

// Port symbols left out, a connector that ended at a part's port ends at the
// part, in every form, and no pin is drawn.
func TestPortSymbolsLeftOutEndConnectorsAtTheirParts(t *testing.T) {
	rendering := render(t, "interconnection-ports.sysml", "ToasterViews::dualView")
	none := Shown{}
	rendering.Shown = none
	shown := rendering.showing(nil)
	byName := nodesByName(shown)
	for _, name := range []string{"primary", "backup", "heating"} {
		if node := byName[name]; node == nil || len(node.Ports) != 0 {
			t.Errorf("%s = %+v, want a part with no pin", name, node)
		}
	}
	var standby *Edge
	for i := range shown.Edges {
		if shown.Edges[i].Kind == EdgeConnection {
			standby = &shown.Edges[i]
		}
	}
	if standby == nil || standby.From != byName["backup"].ID || standby.To != byName["heating"].ID || standby.FromPort != "" || standby.ToPort != "" {
		t.Errorf("standby = %+v, want it from backup %s to heating %s", standby, byName["backup"].ID, byName["heating"].ID)
	}
	for _, form := range []Form{FormText, FormDot, FormMermaid, FormPlantUML, FormD2} {
		out, err := rendering.WriteWith(form, Options{Ports: PortsFull})
		if err != nil {
			t.Fatalf("%s: %v", form, err)
		}
		if strings.Contains(out, "durationOut") || strings.Contains(out, "durationIn") {
			t.Errorf("%s draws a port:\n%s", form, out)
		}
	}
	for _, form := range []Form{FormText, FormDot, FormMermaid, FormPlantUML, FormD2} {
		out, err := rendering.WriteWith(form, Options{Show: Shown{ElementPortSymbols: true}})
		if err != nil {
			t.Fatalf("%s: %v", form, err)
		}
		if !strings.Contains(out, "durationOut") || !strings.Contains(out, "durationIn") {
			t.Errorf("%s with port symbols added does not draw the ports:\n%s", form, out)
		}
	}
}

// Edge labels are left out by default and shown when asked: a connection's
// name labels its line only then, and a line without a label has no empty
// label box in any form.
func TestEdgeLabelsAreShownWhenAsked(t *testing.T) {
	rendering := render(t, "interconnection-ports.sysml", "ToasterViews::dualView")
	for _, form := range []Form{FormText, FormDot, FormMermaid, FormPlantUML, FormD2} {
		out, err := rendering.WriteWith(form, Options{})
		if err != nil {
			t.Fatalf("%s: %v", form, err)
		}
		if strings.Contains(out, "standby") {
			t.Errorf("%s labels the connection by default:\n%s", form, out)
		}
		for _, line := range strings.Split(out, "\n") {
			if !strings.Contains(line, " -> ") && !strings.Contains(line, " -- ") && !strings.Contains(line, "---") {
				continue
			}
			for _, empty := range []string{`label=""`, `|  |`, `: ""`, `-- "" --`, ": ''"} {
				if strings.Contains(line, empty) {
					t.Errorf("%s writes an empty label %s:\n%s", form, empty, out)
				}
			}
		}
		out, err = rendering.WriteWith(form, Options{Show: Shown{ElementEdgeLabels: true}})
		if err != nil {
			t.Fatalf("%s: %v", form, err)
		}
		if !strings.Contains(out, "standby") {
			t.Errorf("%s with edge labels does not label the connection:\n%s", form, out)
		}
	}
}

// Every form refuses an element name none has, by the error naming them.
func TestUnknownElementsAreRefusedOnEveryForm(t *testing.T) {
	_, err := ParseShown("ports")
	var unknown *UnknownElementError
	if !errors.As(err, &unknown) || unknown.Name != "ports" {
		t.Fatalf("err = %v, want an UnknownElementError for ports", err)
	}
	if !strings.Contains(err.Error(), "port-symbols") {
		t.Errorf("%v does not name the elements", err)
	}
}
