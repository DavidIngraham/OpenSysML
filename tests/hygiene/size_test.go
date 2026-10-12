package hygiene

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Caps on a test file and a fixture; the allowlists below record the files that
// predate them, each at its current size, so none of them can grow further.
const (
	maxTestFileLines = 2000
	maxFixtureKiB    = 200
)

// oversize is an allowlisted file: its tracked path, its size when recorded (in
// the unit of its check) and the one-line reason it is still over the cap.
type oversize struct {
	path   string
	size   int
	reason string
}

// oversizeTestFiles lists every `_test.go` file over maxTestFileLines, in lines.
// Split one along the feature seams of AGENTS.md §5.2 and remove its entry.
var oversizeTestFiles = []oversize{
	{"internal/exec/runtime/robustness_test.go", 17178, "shared runtime failure-mode registry; new cases go in robustness_<feature>_test.go"},
	{"tests/export/export_test.go", 4357, "one black-box suite over the whole RDF export surface"},
	{"internal/exec/runtime/classifier_behavior_test.go", 3442, "classifier-behavior scenarios on one shared harness"},
	{"internal/exec/runtime/classify_test.go", 2720, "classification-change scenarios on one shared harness"},
	{"internal/exec/runtime/conformance_test.go", 2710, "execution-conformance driver and its per-case assertions"},
	{"internal/exec/runtime/action_executor_test.go", 2629, "action-executor scenarios on one shared harness"},
	{"internal/exec/runtime/signal_test.go", 2491, "signal delivery scenarios on one shared harness"},
	{"tests/migrate/behavior_more_test.go", 2465, "SysML v1 behavior migration scenarios on one shared harness"},
	{"internal/frontend/lsp/debug_test.go", 2203, "LSP debug-protocol scenarios over one wire fixture"},
	{"internal/exec/runtime/signal_injection_test.go", 2181, "signal-injection scenarios on one shared harness"},
	{"internal/exec/runtime/adopt_test.go", 2161, "adoption scenarios on one shared harness"},
	{"internal/doc/queryexec/execute_test.go", 2114, "query-execution scenarios on one shared harness"},
	{"internal/exec/runtime/debug_api_test.go", 2094, "debug-API scenarios on one shared harness"},
	{"internal/exec/runtime/replay_test.go", 2025, "replay scenarios on one shared harness"},
}

// oversizeFixtures lists every fixture over maxFixtureKiB, in KiB rounded up.
// Shrink the model a golden comes from, or digest it, and remove its entry.
var oversizeFixtures = []oversize{
	{"tests/export/testdata/convert/then_after_members.golden.ttl", 650, "full Turtle golden of a model with many then/after successions"},
	{"tests/export/testdata/superseded/then_done_membership.ttl", 594, "superseded-form Turtle input the reader must still accept whole"},
	{"tests/export/testdata/convert/result_expressions.golden.ttl", 393, "full Turtle golden of a model dense in result expressions"},
	{"tests/export/testdata/interchange/state_entry.toolkit.full.json", 329, "complete toolkit JSON interchange export of a state model"},
	{"tests/export/testdata/interchange/transitions.toolkit.full.json", 313, "complete toolkit JSON interchange export of a transition model"},
	{"tests/export/testdata/interchange/p10.toolkit.full.json", 294, "complete toolkit JSON interchange export of the P10 sample"},
	{"tests/export/testdata/convert/end_binding_bodies.golden.ttl", 261, "full Turtle golden of a model with many end-binding bodies"},
	{"tests/export/testdata/convert/constraints.golden.ttl", 239, "full Turtle golden of a constraint-heavy model"},
	{"tests/export/testdata/convert/irregular_layout.golden.ttl", 230, "full Turtle golden of the irregular-layout sample"},
	{"tests/export/testdata/convert/loops_conditionals.golden.ttl", 213, "full Turtle golden of a model dense in loops and conditionals"},
	{"tests/export/testdata/interchange/states.toolkit.full.json", 212, "complete toolkit JSON interchange export of a state model"},
}

// TestTestFilesStayUnderTheLineCap fails when a tracked `_test.go` file is over
// maxTestFileLines without an oversizeTestFiles entry, when an allowlisted file
// has moved from its recorded size, or when its entry has nothing left to excuse.
func TestTestFilesStayUnderTheLineCap(t *testing.T) {
	var files []string
	for _, path := range trackedFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
	}
	ratchet(t, "test file", files, oversizeTestFiles, maxTestFileLines, "lines", "oversizeTestFiles",
		"split it along the feature seams AGENTS.md §5.2 describes, one file per feature",
		func(data []byte) int { return bytes.Count(data, []byte("\n")) })
}

// TestFixturesStayUnderTheSizeCap fails when a tracked fixture (a file under a
// `testdata/` directory or named `*.golden.*`) is over maxFixtureKiB without an
// oversizeFixtures entry, when an allowlisted fixture has moved from its recorded
// size, or when its entry has nothing left to excuse.
func TestFixturesStayUnderTheSizeCap(t *testing.T) {
	var files []string
	for _, path := range trackedFiles(t) {
		if isFixture(path) {
			files = append(files, path)
		}
	}
	ratchet(t, "fixture", files, oversizeFixtures, maxFixtureKiB, "KiB", "oversizeFixtures",
		"shrink the model it is generated from, or replace it with a digest of the output",
		func(data []byte) int { return (len(data) + 1023) / 1024 })
}

func isFixture(path string) bool {
	for _, component := range strings.Split(filepath.ToSlash(path), "/") {
		if component == "testdata" {
			return true
		}
	}
	base := path[strings.LastIndex(path, "/")+1:]
	i := strings.Index(base, ".golden.")
	return i > 0 && i+len(".golden.") < len(base)
}

// ratchet reports every file over the limit that the allowlist does not record at
// its current size, and every allowlist entry that no longer matches the tree.
func ratchet(t *testing.T, kind string, files []string, allowlist []oversize, limit int, unit, listName, remedy string, measure func([]byte) int) {
	t.Helper()
	recorded := map[string]oversize{}
	for _, entry := range allowlist {
		if _, dup := recorded[entry.path]; dup {
			t.Errorf("%s lists %s twice", listName, entry.path)
		}
		recorded[entry.path] = entry
	}

	seen := map[string]bool{}
	var problems []string
	for _, path := range files {
		data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		size := measure(data)
		entry, listed := recorded[path]
		seen[path] = true
		switch {
		case !listed && size > limit:
			problems = append(problems, fmt.Sprintf("%s is %d %s, over the cap of %d %s: %s; "+
				"if it has to stay whole, add it to %s with its size and the reason",
				path, size, unit, limit, unit, remedy, listName))
		case listed && size > entry.size:
			problems = append(problems, fmt.Sprintf("%s grew from %d to %d %s: %s rather than growing it further "+
				"(a justified growth updates its %s entry)",
				path, entry.size, size, unit, remedy, listName))
		case listed && size <= limit:
			problems = append(problems, fmt.Sprintf("%s is %d %s, within the cap of %d %s: remove its %s entry",
				path, size, unit, limit, unit, listName))
		case listed && size < entry.size:
			problems = append(problems, fmt.Sprintf("%s shrank from %d to %d %s: record the new size in its %s entry so it cannot grow back",
				path, entry.size, size, unit, listName))
		}
	}
	for _, entry := range allowlist {
		if !seen[entry.path] {
			problems = append(problems, fmt.Sprintf("%s is not a tracked %s: remove its %s entry", entry.path, kind, listName))
		}
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}

func trackedFiles(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	return strings.Split(strings.TrimRight(string(out), "\x00"), "\x00")
}
