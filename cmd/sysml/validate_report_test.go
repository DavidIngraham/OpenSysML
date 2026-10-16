package main

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/frontend/report/junit"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/report/sarif"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/report/sarif/sariftest"
)

const (
	reportCleanModel  = "package Q {\n  part def B;\n}\n"
	reportBrokenModel = "package P {\n  part def A;\n  part a : A;\n  part b : Missing;\n}\n"
	// reportNotationModel analyses cleanly with one warning, which carries a fix.
	reportNotationModel = "package S {\n  state def M {\n    entry; then pick;\n    choice pick;\n    transition pick then a;\n    state a;\n  }\n}\n"
)

// reportWorkspace is a workspace of three model files under models/: one
// clean, one with an error and one with a warning that has a fix.
func reportWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "models"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{"clean.sysml": reportCleanModel, "broken.sysml": reportBrokenModel, "notation.sysml": reportNotationModel} {
		if err := os.WriteFile(filepath.Join(dir, "models", name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func decodeSARIF(t *testing.T, data []byte) sarif.Log {
	t.Helper()
	if err := sariftest.Validate(data); err != nil {
		t.Fatalf("%v\n%s", err, data)
	}
	var log sarif.Log
	if err := json.Unmarshal(data, &log); err != nil {
		t.Fatal(err)
	}
	return log
}

// TestValidateSARIFReport: -format sarif writes one SARIF 2.1.0 run of this
// build, one rule per code, each result on its workspace-relative file with
// its line and column, and the fixes analysis knows.
func TestValidateSARIFReport(t *testing.T) {
	binary := buildCLI(t)
	dir := reportWorkspace(t)
	got := checkPathsIn(t, dir, binary, "-validate", "-format", "sarif", "-o", "report.sarif", "models")
	if got.status != exitUnevaluable || got.stdout != "" {
		t.Fatalf("exit %d, stdout %q; want %d and the report in the file\n%s", got.status, got.stdout, exitUnevaluable, got.output())
	}
	data, err := os.ReadFile(filepath.Join(dir, "report.sarif"))
	if err != nil {
		t.Fatal(err)
	}
	log := decodeSARIF(t, data)
	if len(log.Runs) != 1 || log.Runs[0].Tool.Driver.Name != "sysml" || log.Runs[0].Tool.Driver.Version != Version {
		t.Fatalf("run = %+v", log.Runs)
	}
	run := log.Runs[0]
	rules := map[string]int{}
	for i, rule := range run.Tool.Driver.Rules {
		if _, dup := rules[rule.ID]; dup {
			t.Fatalf("rule %q listed twice", rule.ID)
		}
		rules[rule.ID] = i
	}
	byRule := map[string]sarif.Result{}
	for _, r := range run.Results {
		if rules[r.RuleID] != r.RuleIndex {
			t.Fatalf("result %q has ruleIndex %d, the rule is at %d", r.RuleID, r.RuleIndex, rules[r.RuleID])
		}
		byRule[r.RuleID] = r
	}
	unresolved, ok := byRule["unresolved"]
	if !ok || unresolved.Level != "error" {
		t.Fatalf("unresolved reference not reported as an error: %+v", run.Results)
	}
	loc := unresolved.Locations[0].PhysicalLocation
	if loc.ArtifactLocation.URI != "models/broken.sysml" || loc.Region.StartLine != 4 || loc.Region.StartColumn != 12 {
		t.Fatalf("unresolved reference placed at %+v", loc)
	}
	notation, ok := byRule["nonstandard-notation"]
	if !ok || notation.Level != "warning" {
		t.Fatalf("notation warning not reported with its own level: %+v", run.Results)
	}
	if len(notation.Fixes) != 1 || len(notation.Fixes[0].ArtifactChanges) != 1 {
		t.Fatalf("notation fix = %+v", notation.Fixes)
	}
	change := notation.Fixes[0].ArtifactChanges[0]
	if change.ArtifactLocation.URI != "models/notation.sysml" || len(change.Replacements) == 0 ||
		change.Replacements[0].InsertedContent == nil || change.Replacements[0].InsertedContent.Text != "#choice state pick;" {
		t.Fatalf("notation fix change = %+v", change)
	}
	if change.Replacements[0].DeletedRegion.StartLine != 4 || change.Replacements[0].DeletedRegion.EndColumn <= change.Replacements[0].DeletedRegion.StartColumn {
		t.Fatalf("deleted region = %+v", change.Replacements[0].DeletedRegion)
	}
}

// TestValidateSARIFOnExamples: the showcase examples report as valid SARIF to
// stdout, and -format json is -json.
func TestValidateSARIFOnExamples(t *testing.T) {
	binary := buildCLI(t)
	got := checkPaths(t, binary, "-validate", "-format", "sarif", "../../examples/runtime-showcase")
	log := decodeSARIF(t, []byte(got.stdout))
	if len(log.Runs) != 1 {
		t.Fatalf("runs = %d", len(log.Runs))
	}
	asJSON := checkPaths(t, binary, "-validate", "-json", "../../examples/runtime-showcase")
	asFormat := checkPaths(t, binary, "-validate", "-format", "json", "../../examples/runtime-showcase")
	if asJSON.stdout != asFormat.stdout || asJSON.status != asFormat.status {
		t.Fatalf("-format json differs from -json:\n%s\n---\n%s", asFormat.output(), asJSON.output())
	}
}

// TestValidateJUnitReport: -format junit writes one suite per file, a failing
// case per finding typed by its severity and a passing case for a clean file.
func TestValidateJUnitReport(t *testing.T) {
	binary := buildCLI(t)
	dir := reportWorkspace(t)
	got := checkPathsIn(t, dir, binary, "-validate", "-format", "junit", "models")
	if got.status != exitUnevaluable {
		t.Fatalf("exit %d, want %d\n%s", got.status, exitUnevaluable, got.output())
	}
	var suites junit.Testsuites
	if err := xml.Unmarshal([]byte(got.stdout), &suites); err != nil {
		t.Fatalf("not JUnit XML: %v\n%s", err, got.stdout)
	}
	if len(suites.Suites) != 3 || suites.Tests != 3 || suites.Failures != 2 || suites.Errors != 0 {
		t.Fatalf("totals = %d suites, %d tests, %d failures, %d errors\n%s", len(suites.Suites), suites.Tests, suites.Failures, suites.Errors, got.stdout)
	}
	byName := map[string]junit.Testsuite{}
	for _, s := range suites.Suites {
		byName[s.Name] = s
	}
	clean := byName["models/clean.sysml"]
	if clean.Tests != 1 || clean.Failures != 0 || len(clean.Cases) != 1 || clean.Cases[0].Failure != nil {
		t.Fatalf("clean file = %+v", clean)
	}
	broken := byName["models/broken.sysml"]
	if broken.Failures != len(broken.Cases) || broken.Failures == 0 {
		t.Fatalf("broken file = %+v", broken)
	}
	for _, c := range broken.Cases {
		if c.Failure == nil || c.Failure.Type != "error" || !strings.Contains(c.Failure.Body, "models/broken.sysml:") {
			t.Fatalf("broken case = %+v", c)
		}
	}
	notation := byName["models/notation.sysml"]
	if notation.Failures != 1 || notation.Cases[0].Failure.Type != "warning" || !strings.HasPrefix(notation.Cases[0].Name, "nonstandard-notation at 4:5") {
		t.Fatalf("notation file = %+v", notation)
	}
}

// TestValidateBaselineRoundTrip: the JSON report -o wrote is read back as the
// known findings; they are suppressed wherever their lines moved, and only a
// finding the baseline does not know ends the run.
func TestValidateBaselineRoundTrip(t *testing.T) {
	binary := buildCLI(t)
	dir := reportWorkspace(t)
	recorded := checkPathsIn(t, dir, binary, "-validate", "-json", "-o", "baseline.json", "models")
	if recorded.status != exitUnevaluable {
		t.Fatalf("recording exit %d\n%s", recorded.status, recorded.output())
	}
	var baseline checkReport
	if data, err := os.ReadFile(filepath.Join(dir, "baseline.json")); err != nil {
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	known := len(baseline.Diagnostics)
	if known < 2 {
		t.Fatalf("baseline recorded %d findings\n%+v", known, baseline.Diagnostics)
	}

	again := checkPathsIn(t, dir, binary, "-validate", "-baseline", "baseline.json", "models")
	wantReport(t, again, exitHolds, "known finding(s) suppressed by baseline.json", "models: no errors")
	rejectReport(t, again, "unresolved reference")

	// The findings move down two lines and stay known.
	broken := filepath.Join(dir, "models", "broken.sysml")
	if err := os.WriteFile(broken, []byte("\n\n"+reportBrokenModel), 0o644); err != nil {
		t.Fatal(err)
	}
	moved := checkPathsIn(t, dir, binary, "-validate", "-baseline", "baseline.json", "-json", "models")
	var report checkReport
	if err := json.Unmarshal([]byte(moved.stdout), &report); err != nil {
		t.Fatalf("%v\n%s", err, moved.output())
	}
	if moved.status != exitHolds || report.Suppressed != known || len(report.Diagnostics) != 0 {
		t.Fatalf("after moving lines: exit %d, %d suppressed, %d reported\n%s", moved.status, report.Suppressed, len(report.Diagnostics), moved.output())
	}

	// A finding the baseline does not know is the only one reported, and ends the run.
	if err := os.WriteFile(broken, []byte("\n\n"+strings.Replace(reportBrokenModel, "part a : A;", "part a : A;\n  part c : AlsoMissing;", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	fresh := checkPathsIn(t, dir, binary, "-validate", "-baseline", "baseline.json", "models")
	wantReport(t, fresh, exitUnevaluable, "unresolved reference: AlsoMissing", "did not analyse cleanly")
	rejectReport(t, fresh, "unresolved reference: Missing\n")
	sarifRun := checkPathsIn(t, dir, binary, "-validate", "-baseline", "baseline.json", "-format", "sarif", "models")
	log := decodeSARIF(t, []byte(sarifRun.stdout))
	if len(log.Runs[0].Results) != 1 || log.Runs[0].Properties == nil || log.Runs[0].Properties.Suppressed != known {
		t.Fatalf("SARIF after baseline: %d results, properties %+v", len(log.Runs[0].Results), log.Runs[0].Properties)
	}
}

// TestValidateWarningsAsErrors: a warning leaves the exit status alone until
// -warnings-as-errors, which ends the run on it while the report keeps the
// warning's level.
func TestValidateWarningsAsErrors(t *testing.T) {
	binary := buildCLI(t)
	dir := reportWorkspace(t)
	lenient := checkPathsIn(t, dir, binary, "-validate", "models/notation.sysml")
	wantReport(t, lenient, exitHolds, "warning:", "no errors")
	strict := checkPathsIn(t, dir, binary, "-validate", "-warnings-as-errors", "models/notation.sysml")
	wantReport(t, strict, exitUnevaluable, "warning:", "-warnings-as-errors judges errors")
	rejectReport(t, strict, "no errors")
	asSARIF := checkPathsIn(t, dir, binary, "-validate", "-warnings-as-errors", "-format", "sarif", "models/notation.sysml")
	log := decodeSARIF(t, []byte(asSARIF.stdout))
	if asSARIF.status != exitUnevaluable || len(log.Runs[0].Results) != 1 || log.Runs[0].Results[0].Level != "warning" {
		t.Fatalf("exit %d, results %+v", asSARIF.status, log.Runs[0].Results)
	}
}

// TestValidateFormatMisuse: a format that is not one, -json beside another
// -format, -o without a document and a report flag without a check are refused.
func TestValidateFormatMisuse(t *testing.T) {
	binary := buildCLI(t)
	dir := reportWorkspace(t)
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-validate", "-format", "yaml", "models"}, "-format yaml is not a report format"},
		{[]string{"-validate", "-json", "-format", "sarif", "models"}, "-json reports JSON; it cannot be combined with -format sarif"},
		{[]string{"-validate", "-o", "report.txt", "models"}, "-o names the file the report document is written to"},
		{[]string{"-format", "sarif", "models"}, "-format reports a check; name one"},
		{[]string{"-baseline", "baseline.json", "models"}, "-baseline accepts the known findings of a check"},
		{[]string{"-warnings-as-errors", "models"}, "-warnings-as-errors judges the warnings of a check"},
		{[]string{"-validate", "-baseline", "absent.json", "models"}, "-baseline open absent.json"},
	} {
		got := checkPathsIn(t, dir, binary, tc.args...)
		wantReport(t, got, exitUnevaluable, tc.want)
	}
}
