//go:build js && wasm

package combined

import (
	"runtime"
	"testing"
)

func TestBrowserMemoryPhaseCollectsLargeWorkAndPreservesReachableData(t *testing.T) {
	s := &Server{}
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	data := make([]byte, memoryPhaseBudget+1)
	for i := range data {
		data[i] = byte(i)
	}
	s.finishMemoryPhase("ParseSources", before.TotalAlloc)
	if s.phaseCollections != 1 || s.lastPhaseCollection == nil {
		t.Fatalf("large work did not collect: %+v", s)
	}
	phase := s.lastPhaseCollection
	if phase.Method != "ParseSources" || phase.AllocatedBytes < memoryPhaseBudget || phase.BeforeHeapAllocBytes < memoryPhaseBudget || phase.AfterHeapAllocBytes < uint64(len(data)) {
		t.Fatalf("missing phase evidence or reachable data: %+v", phase)
	}
	for i := range data {
		if data[i] != byte(i) {
			t.Fatalf("reachable data changed at %d", i)
		}
	}
	runtime.KeepAlive(data)
}

func TestBrowserMemoryPhaseSkipsSmallWorkAndQueries(t *testing.T) {
	s := &Server{}
	for _, method := range []string{"GetSymbol", "GetDiagnostics", "Evaluate", "GetServerInfo", "GetMemoryStats", "ReleaseModel"} {
		if s.beginMemoryPhase(method) != nil {
			t.Fatalf("query %s starts a collection phase", method)
		}
	}
	var current runtime.MemStats
	runtime.ReadMemStats(&current)
	s.finishMemoryPhase("ParseSources", current.TotalAlloc)
	if s.phaseCollections != 0 {
		t.Fatal("zero-allocation work forced collection")
	}
}

func TestBrowserReleaseCollectionRequiresMaterialHeap(t *testing.T) {
	s := &Server{}
	data := make([]byte, memoryPhaseBudget+1)
	data[0], data[len(data)-1] = 17, 42
	s.collectReleasedModel()
	if s.phaseCollections != 1 || s.lastPhaseCollection.Method != "ReleaseModel" || s.lastPhaseCollection.AllocatedBytes != 0 {
		t.Fatalf("release collection not recorded: %+v", s.lastPhaseCollection)
	}
	if data[0] != 17 || data[len(data)-1] != 42 {
		t.Fatal("release collection corrupted reachable data")
	}
	runtime.KeepAlive(data)
}
