package passes

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
)

// The documents of one batch settle a namespace's import collisions once: an
// analysis taking what an earlier one settled reports the same and still
// records the reads that settling made, so its interface record depends on
// everything along the chain of re-exports the collisions were decided over.
func TestBatchSharesImportCollisionsWithTheirReads(t *testing.T) {
	const n = 6
	docs := map[string]string{}
	var names []string
	for k := 0; k < n; k++ {
		var src strings.Builder
		fmt.Fprintf(&src, "package P%d {\n", k)
		if k > 0 {
			fmt.Fprintf(&src, "    public import P%d::*;\n    part q%d : D%d;\n", k-1, k, k-1)
		}
		fmt.Fprintf(&src, "    part def D%d;\n}\n", k)
		name := fmt.Sprintf("p%d.sysml", k)
		docs[name] = src.String()
		names = append(names, name)
	}
	last := names[n-1]
	idx, roots := indexedBatch(t, docs)
	batch := &Batch{Documents: names, Record: true}
	PrepareBatch(idx, batch)
	if batch.Collisions == nil {
		t.Fatal("preparing a batch shares no collisions memo")
	}
	diags, settled := AnalyzeInBatch(last, source.KindSysML, roots[last], nil, idx, Options{}, batch)
	again, taken := AnalyzeInBatch(last, source.KindSysML, roots[last], nil, idx, Options{}, batch)
	if !reflect.DeepEqual(diags, again) {
		t.Errorf("the analysis taking the shared collisions reports %v, want what settling them reported: %v", again, diags)
	}
	if settled == nil || taken == nil || len(settled.Namespaces) == 0 {
		t.Fatalf("recorded reads %v and %v, want both with namespaces", settled, taken)
	}
	// The audits' gathers are shared too, so only the chain's own reads compare.
	for _, chain := range []struct{ settled, taken []string }{{settled.Namespaces, taken.Namespaces}, {settled.Names, taken.Names}} {
		if want, got := ofChain(chain.settled), ofChain(chain.taken); !reflect.DeepEqual(want, got) {
			t.Errorf("the analysis taking the shared collisions read %v, want what settling them read: %v", got, want)
		}
	}
}

func ofChain(reads []string) []string {
	var out []string
	for _, read := range reads {
		if strings.HasPrefix(read, "P") {
			out = append(out, read)
		}
	}
	return out
}
