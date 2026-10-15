package migrate_test

import (
	"strings"
	"testing"
)

// A v1 name that holds a quote or a newline is spelled in every query the
// migration writes the way the index reads it: escapes kept, quotes dropped.
func TestTableQueriesResolveEscapedNames(t *testing.T) {
	r := migrateDocument(t, `
    <packagedElement xmi:type="uml:Package" xmi:id="_pkg" name="Frank's Braindump">
      <packagedElement xmi:type="uml:Package" xmi:id="_composition" name="Composition Example">
        <packagedElement xmi:type="uml:Class" xmi:id="_instrument" name="Instrument">
          <ownedAttribute xmi:type="uml:Property" xmi:id="_bench_part" name="bench" type="_bench" aggregation="composite" association="_association"/>
        </packagedElement>
        <packagedElement xmi:type="uml:Class" xmi:id="_bench" name="Optical&#10;Bench"/>
        <packagedElement xmi:type="uml:Association" xmi:id="_association" memberEnd="_bench_part _instrument_end">
          <ownedEnd xmi:type="uml:Property" xmi:id="_instrument_end" name="instrument" type="_instrument" association="_association"/>
        </packagedElement>
      </packagedElement>
    </packagedElement>`, `
  <xmi:Extension extender="Example UML Tool 1.0">
    <modelExtension>
      <ownedDiagram xmi:type="uml:Diagram" xmi:id="_property_diagram" name="Properties" ownerOfDiagram="_composition">
        <xmi:Extension extender="Example UML Tool 1.0">
          <diagramRepresentation>
            <diagram:DiagramRepresentationObject xmlns:diagram="http://www.example.com/tool/diagram" type="Generic Table" umlType="Class Diagram">
              <diagramContents/>
            </diagram:DiagramRepresentationObject>
          </diagramRepresentation>
        </xmi:Extension>
      </ownedDiagram>
      <ownedDiagram xmi:type="uml:Diagram" xmi:id="_instance_diagram" name="Bench Instances" ownerOfDiagram="_composition">
        <xmi:Extension extender="Example UML Tool 1.0">
          <diagramRepresentation>
            <diagram:DiagramRepresentationObject xmlns:diagram="http://www.example.com/tool/diagram" type="SysML Instance Table" umlType="Class Diagram">
              <diagramContents/>
            </diagram:DiagramRepresentationObject>
          </diagramRepresentation>
        </xmi:Extension>
      </ownedDiagram>
    </modelExtension>
  </xmi:Extension>
  <MagicDraw_Profile:DiagramTable xmlns:MagicDraw_Profile="http://www.omg.org/spec/UML/20131001/MagicDrawProfile" xmi:id="_property_table" base_Diagram="_property_diagram" scope="_composition" excludedElements="_instrument_end">
    <rowElementType href="http://www.omg.org/spec/UML/20131001/UML.xmi#Property"/>
    <columnIds>QPROP:Element:qualifiedName</columnIds>
  </MagicDraw_Profile:DiagramTable>
  <MagicDraw_Profile:InstanceTable xmlns:MagicDraw_Profile="http://www.omg.org/spec/UML/20131001/MagicDrawProfile" xmi:id="_instance_table" base_Diagram="_instance_diagram" scope="_composition" includeSubtypesOfRowTypes="true" showScopeAsRoot="false">
    <classifiers xmi:idref="_bench"/>
    <columnIds>QPROP:Element:qualifiedName</columnIds>
  </MagicDraw_Profile:InstanceTable>`)
	wantClean(t, "escaped_table_names.sysml", r)
	s := session(t, r)
	propertyRows := rows(t, s, `'Frank\'s Braindump'::'Composition Example'::'Properties Rows'`)
	if !strings.Contains(propertyRows, `Frank\\'s Braindump::Composition Example::Instrument::bench`) {
		t.Errorf("Properties rows omit the class-owned part property:\n%s", propertyRows)
	}

	instanceRows := rows(t, s, `'Frank\'s Braindump'::'Composition Example'::'Bench Instances Rows'`)
	if strings.Contains(instanceRows, "::bench") {
		t.Errorf("Bench Instances rows list the part, not an instance:\n%s", instanceRows)
	}

	notation := string(r.Notation)
	for _, spelled := range []string{
		`DocumentQueries::Named(qualifiedName = ("Frank\\'s Braindump::Composition Example"))`,
		`type = ("Frank\\'s Braindump::Composition Example::Optical\\nBench")`,
	} {
		if !strings.Contains(notation, spelled) {
			t.Errorf("the queries do not spell %s with escaped names:\n%s", spelled, notation)
		}
	}
	for _, document := range []string{
		`'Frank\'s Braindump'::'Composition Example'::'Properties Document'`,
		`'Frank\'s Braindump'::'Composition Example'::'Bench Instances Document'`,
	} {
		page := html(t, s, document)
		if strings.Contains(page, "names no single element") {
			t.Errorf("%s rendered an error page:\n%s", document, page)
		}
	}
}
