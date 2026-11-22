//go:build js && wasm

package combined

import (
	"runtime"
	"time"
)

// A work-phase budget bounds transient garbage in the 32-bit browser heap.
const memoryPhaseBudget = 64 << 20

func (s *Server) beginMemoryPhase(method string) func() {
	switch method {
	case "ParseSources", "ParseFile", "RenderView", "ExportModel":
		var before runtime.MemStats
		runtime.ReadMemStats(&before)
		return func() { s.finishMemoryPhase(method, before.TotalAlloc) }
	default:
		return nil
	}
}

func (s *Server) finishMemoryPhase(method string, startAllocated uint64) {
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - startAllocated
	if allocated < memoryPhaseBudget || after.HeapAlloc < memoryPhaseBudget {
		return
	}
	s.collectMemoryPhase(method, allocated, after.HeapAlloc)
}

func (s *Server) collectReleasedModel() {
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	if after.HeapAlloc >= memoryPhaseBudget {
		s.collectMemoryPhase("ReleaseModel", 0, after.HeapAlloc)
	}
}

func (s *Server) collectMemoryPhase(method string, allocated, beforeHeap uint64) {
	started := time.Now()
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	s.phaseCollections++
	s.lastPhaseCollection = &phaseCollection{
		Method:               method,
		AllocatedBytes:       allocated,
		BeforeHeapAllocBytes: beforeHeap,
		AfterHeapAllocBytes:  after.HeapAlloc,
		ElapsedMillis:        float64(time.Since(started)) / float64(time.Millisecond),
	}
}
