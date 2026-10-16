package validation

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Open-MBEE/OpenSysML/tools/census/manifest"
	"github.com/Open-MBEE/OpenSysML/tools/oracle/errata"
	"github.com/Open-MBEE/OpenSysML/tools/oracle/repo"
)

// loadCommittedManifest reads the committed manifest or fails the test.
func loadCommittedManifest(t *testing.T) (string, *manifest.Manifest) {
	t.Helper()
	root, err := repo.Root()
	if err != nil {
		t.Fatal(err)
	}
	m, err := manifest.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRecord(m); err != nil {
		t.Fatal(err)
	}
	return root, m
}

// TestManifestCodesAreTheProbesDiagnostics: the diagnostic codes and passes a
// row records are exactly what its probes observe, so every code the manifest
// names is one the tree emits and no emitted code goes unrecorded.
func TestManifestCodesAreTheProbesDiagnostics(t *testing.T) {
	root, m := loadCommittedManifest(t)
	codes, passes := probeCodes(t, root)
	for _, c := range m.Constraints {
		if got, want := sorted(c.Codes), keys(codes[c.Name]); got != want {
			t.Errorf("%s records diagnostic codes %s, its probes observe %s", c.Name, got, want)
		}
		if got, want := sorted(c.Passes), keys(passes[c.Name]); got != want {
			t.Errorf("%s records passes %s, its probes observe %s", c.Name, got, want)
		}
	}
}

func sorted(list []string) string {
	copied := append([]string(nil), list...)
	sort.Strings(copied)
	return strings.Join(copied, ",")
}

func keys(set map[string]bool) string {
	var list []string
	for k := range set {
		list = append(list, k)
	}
	return sorted(list)
}

// TestManifestRoundTripsByteIdentical: decoding and re-encoding the committed
// manifest reproduces it byte for byte, so no field the tree records is one
// this program would drop on -update.
func TestManifestRoundTripsByteIdentical(t *testing.T) {
	root, m := loadCommittedManifest(t)
	committed, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(manifest.Path)))
	if err != nil {
		t.Fatal(err)
	}
	if encoded := manifest.Encode(m); !bytes.Equal(encoded, committed) {
		t.Fatalf("%s is not what this program encodes; a field it does not know, or a non-canonical encoding:\n%s", manifest.Path, diffHint(committed, encoded))
	}
}

// diffHint names the first differing line of two encodings.
func diffHint(a, b []byte) string {
	al, bl := strings.Split(string(a), "\n"), strings.Split(string(b), "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return fmt.Sprintf("line %d:\n  committed: %s\n  encoded:   %s", i+1, al[i], bl[i])
		}
	}
	return fmt.Sprintf("committed has %d lines, encoded %d", len(al), len(bl))
}

// constraintTierPasses lists the production files under internal/check/passes
// declaring a pass whose Level is LevelConstraint, keyed by repository path.
func constraintTierPasses(t *testing.T, root string) map[string][]string {
	t.Helper()
	dir := filepath.Join(root, "internal", "check", "passes")
	entries, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]string{}
	for _, path := range entries {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "Level" || fn.Recv == nil || fn.Body == nil {
				continue
			}
			constraint := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if ident, ok := n.(*ast.Ident); ok && ident.Name == "LevelConstraint" {
					constraint = true
				}
				return !constraint
			})
			if constraint {
				rel := "internal/check/passes/" + filepath.Base(path)
				files[rel] = append(files[rel], receiverType(fn))
			}
		}
	}
	if len(files) == 0 {
		t.Fatal("no constraint-tier pass found under internal/check/passes")
	}
	return files
}

// TestConstraintTierPassesAreAccountedFor: every constraint-tier pass is cited
// by a row's implementation or listed among the manifest's extensions with its
// basis, and no extension names a file that is not such a pass.
func TestConstraintTierPassesAreAccountedFor(t *testing.T) {
	root, m := loadCommittedManifest(t)
	passes := constraintTierPasses(t, root)
	cited := map[string]bool{}
	for _, c := range m.Constraints {
		for _, ref := range implementationRef.FindAllStringSubmatch(c.Implementation, -1) {
			cited[ref[1]] = true
		}
	}
	extensions := map[string]manifest.Extension{}
	for _, e := range m.Extensions {
		extensions[e.File] = e
		types, ok := passes[e.File]
		if !ok {
			t.Errorf("extension %s is not a constraint-tier pass file under internal/check/passes", e.File)
			continue
		}
		if !slicesContains(types, e.Pass) {
			t.Errorf("extension %s names pass %s; the file declares %s", e.File, e.Pass, strings.Join(types, ", "))
		}
		if cited[e.File] {
			t.Errorf("extension %s is also cited by a row; a pass is one or the other", e.File)
		}
		switch e.Basis {
		case "spec", "extension":
		default:
			t.Errorf("extension %s has basis %q, want spec or extension", e.File, e.Basis)
		}
	}
	var unaccounted []string
	for file := range passes {
		if _, ok := extensions[file]; !ok && !cited[file] {
			unaccounted = append(unaccounted, file)
		}
	}
	sort.Strings(unaccounted)
	if len(unaccounted) > 0 {
		t.Errorf("constraint-tier passes neither cited by a row nor listed as extensions in %s:\n  %s", manifest.Path, strings.Join(unaccounted, "\n  "))
	}
}

func slicesContains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// TestManifestErrataMatchTheRegistry: the errata the manifest lists are exactly
// the declared overlay, entry for entry, so neither can gain or lose one alone.
func TestManifestErrataMatchTheRegistry(t *testing.T) {
	_, m := loadCommittedManifest(t)
	registry := errata.Registry()
	if len(m.Errata) != len(registry) {
		t.Fatalf("%s lists %d errata, the registry declares %d", manifest.Path, len(m.Errata), len(registry))
	}
	for i, want := range registry {
		got := m.Errata[i]
		if got.ID != want.ID || got.Heading != want.Heading || got.Path != want.Path || got.Line != want.Line {
			t.Errorf("erratum %d: manifest records %+v, the registry declares %s %q at %s:%d", i, got, want.ID, want.Heading, want.Path, want.Line)
		}
	}
}
