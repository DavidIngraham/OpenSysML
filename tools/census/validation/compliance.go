package validation

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Open-MBEE/OpenSysML/tools/census/doccounts"
	"github.com/Open-MBEE/OpenSysML/tools/census/manifest"
)

// validationLibraryPath is the bundled library stating the specifications'
// constraints under their own names; a `spec:` citation must be one of them.
const validationLibraryPath = "internal/workspace/libs/stdlib/OpenSysML Libraries/SysMLValidation.sysml"

// goSourceDirs are the trees whose validate… functions a row may name.
var goSourceDirs = []string{"cmd", "internal", "tools"}

var (
	citedNamePattern     = regexp.MustCompile(`(spec:)?\bvalidate[A-Z][A-Za-z0-9]*_?`)
	censusReferencePat   = regexp.MustCompile("`(validate[A-Z][A-Za-z0-9]*_?)[^`]*` \\(census ([a-z]+(?: [a-z]+)?)[;,)]")
	constraintDefPattern = regexp.MustCompile(`(?m)^\s*constraint def (validate[A-Za-z0-9]+)`)
	goDeclPattern        = regexp.MustCompile(`(?m)^func (?:\([^)]*\) )?(validate[A-Za-z0-9_]*)\(`)
)

// rowMarkerStatus maps a compliance-map status marker to the census status a
// bare citation in that row claims; an extension row (🧩) claims none.
var rowMarkerStatus = map[string]string{
	"✅": manifest.StatusFaithful,
	"⚠": manifest.StatusApproximate,
	"❌": manifest.StatusNotImplemented,
	"⛔": manifest.StatusDeliberate,
	"🚧": manifest.StatusKnownFailure,
}

// statusWord is the text a `(census …)` reference states for a status: the
// census marker's words, without its emoji, so a status cell keeps one marker.
func statusWord(status string) string {
	marker, _ := manifest.MarkerFor(status)
	_, word, _ := strings.Cut(marker, " ")
	return word
}

func plainMarker(s string) string { return strings.ReplaceAll(s, "\ufe0f", "") }

// constraintIndex resolves every spelling the page may cite to its manifest row:
// the pilot's name, its raw identifier and the formal document's name.
func constraintIndex(m *manifest.Manifest) map[string]*manifest.Constraint {
	index := make(map[string]*manifest.Constraint, len(m.Constraints))
	for i := range m.Constraints {
		c := &m.Constraints[i]
		index[c.Name] = c
		if c.Raw != "" {
			index[c.Raw] = c
		}
		if c.Spec != nil && c.Spec.Name != "" {
			index[c.Spec.Name] = c
		}
	}
	return index
}

// libraryConstraints lists the constraint definitions of the validation library.
func libraryConstraints(root string) (map[string]bool, error) {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(validationLibraryPath))) // #nosec G304 -- fixed repository path
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool)
	for _, m := range constraintDefPattern.FindAllStringSubmatch(string(content), -1) {
		names[m[1]] = true
	}
	return names, nil
}

// goValidateFunctions lists every validate… function or method declared under
// the Go source trees, so a row may name an implementation helper.
func goValidateFunctions(root string) (map[string]bool, error) {
	names := make(map[string]bool)
	for _, dir := range goSourceDirs {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			content, err := os.ReadFile(path) // #nosec G304 -- walking the repository
			if err != nil {
				return err
			}
			for _, m := range goDeclPattern.FindAllSubmatch(content, -1) {
				names[string(m[1])] = true
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return names, nil
}

// checkCompliance holds the compliance map to the census: every validate… name
// it cites is a manifest row (by pilot, raw or formal name), a `spec:` name the
// validation library states, or a Go function of the tree; and a rule row's
// status is the census status of every row it cites bare, a citation followed
// by `(census faithful)` deferring to the census row's status instead, which it must
// state correctly.
func checkCompliance(root, content string, m *manifest.Manifest) error {
	rows, err := doccounts.RuleRows(content)
	if err != nil {
		return err
	}
	library, err := libraryConstraints(root)
	if err != nil {
		return err
	}
	goFuncs, err := goValidateFunctions(root)
	if err != nil {
		return err
	}
	index := constraintIndex(m)
	markers := make(map[int]string, len(rows))
	for _, row := range rows {
		markers[row.Line] = row.Marker
	}
	var problems []string
	for i, line := range strings.Split(content, "\n") {
		number := i + 1
		references := make(map[string]string)
		for _, ref := range censusReferencePat.FindAllStringSubmatch(line, -1) {
			references[ref[1]] = ref[2]
		}
		claimed, seen := make(map[string]bool), make(map[string]bool)
		for _, token := range citedNamePattern.FindAllString(line, -1) {
			if seen[token] {
				continue
			}
			seen[token] = true
			if name, ok := strings.CutPrefix(token, "spec:"); ok {
				if !library[name] {
					problems = append(problems, fmt.Sprintf("%s:%d: `spec:%s` is not a constraint of %s", doccounts.SpecCompliancePath, number, name, filepath.Base(validationLibraryPath)))
				}
				continue
			}
			c, ok := index[token]
			if !ok {
				if !goFuncs[token] {
					problems = append(problems, fmt.Sprintf("%s:%d: `%s` is neither a constraint of %s nor a Go function of the tree", doccounts.SpecCompliancePath, number, token, manifest.Path))
				}
				continue
			}
			if stated, ok := references[token]; ok {
				if want := statusWord(c.Status); stated != want {
					problems = append(problems, fmt.Sprintf("%s:%d: `%s` is cited as (census %s) but the census row is %s", doccounts.SpecCompliancePath, number, token, stated, want))
				}
				continue
			}
			marker, isRow := markers[number]
			if !isRow || claimed[c.Name] {
				continue
			}
			claimed[c.Name] = true
			if rowMarkerStatus[plainMarker(marker)] != c.Status {
				problems = append(problems, fmt.Sprintf("%s:%d: a %s row cites `%s`, whose census status is %s; move one status or cite it as (census %s)", doccounts.SpecCompliancePath, number, plainMarker(marker), token, c.Status, statusWord(c.Status)))
			}
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("%s does not agree with %s:\n  %s", doccounts.SpecCompliancePath, manifest.Path, strings.Join(problems, "\n  "))
}
