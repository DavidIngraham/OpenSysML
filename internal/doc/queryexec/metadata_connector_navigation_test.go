package queryexec

import (
	"testing"
)

func TestMetadataFilterAcceptsConnectorEndNavigation(t *testing.T) {
	fixture := loadExecutionFixture(t, navigationBody+`
package Declared { requirement a; requirement b;
connection d {
end #RequirementDerivation::OriginalRequirementMetadata source ::> a;
end #RequirementDerivation::DerivedRequirementMetadata target ::> b; } }

calc def Q :> Query {
 in root : Element;
 Project(source = WhereType(source = Descendants(source = root, maxDepth = 1), type = "ConnectionUsage"),
 columns = (Column(name = "Ends", cell = { in row : KerML::Kernel::Connector;
 WhereMetadata(source = row.connectorEnd, 'metadata' = "RequirementDerivation::OriginalRequirementMetadata") })))
}`)
	result, err := fixture.execute(t, "Q", Bindings{"root": {ElementValue(fixture.symbol(t, "Declared"))}}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows()) != 1 {
		t.Fatalf("got %d rows", len(result.Rows()))
	}
	got := cellElementNames(t, result, "Ends")
	if len(got) != 1 || len(got[0]) != 1 || got[0][0] != "Observatory::Declared::d::source" {
		t.Fatalf("metadata selected %v", got)
	}

}
