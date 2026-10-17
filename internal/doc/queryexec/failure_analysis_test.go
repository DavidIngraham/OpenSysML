package queryexec

import (
	"slices"
	"testing"
)

func TestFailureAnalysisReusableRows(t *testing.T) {
	fixture := loadExecutionFixture(t, `
    occurrence cause : FailureAnalysis::BasicEvent { attribute :>> description = "Common power loss"; }
    occurrence effect : FailureAnalysis::Event { attribute :>> description = "Both buses unavailable"; }
    part def LocalRow :> FailureAnalysis::ReviewRow {
        attribute :>> intendedFunction = "Supply power";
        ref occurrence :>> cause = Observatory::cause;
        ref occurrence :>> failureMode = effect;
        attribute :>> recommendedAction = "Inject a supply fault";
    }
    part review : FailureAnalysis::Review {
        part rowOne : LocalRow subsets modes { attribute :>> identifier = "P-01"; }
        part rowTwo : LocalRow subsets modes { attribute :>> identifier = "P-02"; }
        part unrelated;
    }
    calc def Rows :> FailureAnalysisViews::ReviewRows { in :>> reviewName = "Observatory::review"; }
    `)
	rows, err := fixture.execute(t, "Rows", nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := cellTexts(t, rows, 0); !slices.Equal(got, []string{"P-01", "P-02"}) {
		t.Fatalf("rows=%v", got)
	}
	if got := cellTexts(t, rows, 2); !slices.Equal(got, []string{"Common power loss", "Common power loss"}) {
		t.Fatalf("causes=%v", got)
	}
	if got := cellTexts(t, rows, 3); !slices.Equal(got, []string{"Both buses unavailable", "Both buses unavailable"}) {
		t.Fatalf("effects=%v", got)
	}
}
