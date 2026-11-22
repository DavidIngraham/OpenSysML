// Copyright 2025 Open‐MBEE Foundation. All rights reserved.
// Use of this source code is governed by the LICENSE file.

package combined

import (
	"encoding/json"
	"runtime"
	"runtime/metrics"
	"strconv"

	"github.com/Open-MBEE/OpenSysML/internal/frontend/jsonrpc"
)

func (s *Server) releaseModel(params []byte) ([]byte, error) {
	var req struct {
		ModelHash string `json:"modelHash"`
	}
	if err := jsonrpc.Decode(params, &req); err != nil {
		return nil, err
	}
	if req.ModelHash == "" {
		return nil, jsonrpc.Errorf(jsonrpc.CodeInvalidArgument, "modelHash is required")
	}
	elem, released := s.byHash[req.ModelHash]
	if released {
		s.models.Remove(elem)
		delete(s.byHash, req.ModelHash)
	}
	s.core.Evict(req.ModelHash)
	s.engine.Evict(req.ModelHash)
	if released {
		s.collectReleasedModel()
	}
	return json.Marshal(struct {
		Released       bool `json:"released"`
		RetainedModels int  `json:"retainedModels"`
	}{Released: released, RetainedModels: s.models.Len()})
}

func (s *Server) memoryStats(params []byte) ([]byte, error) {
	var req struct {
		Collect bool `json:"collect"`
	}
	if err := jsonrpc.Decode(params, &req); err != nil {
		return nil, err
	}
	if req.Collect {
		runtime.GC()
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	gcSettings := []metrics.Sample{{Name: "/gc/gogc:percent"}, {Name: "/gc/gomemlimit:bytes"}}
	metrics.Read(gcSettings)
	hashes := make([]string, 0, s.models.Len())
	for elem := s.models.Front(); elem != nil; elem = elem.Next() {
		hashes = append(hashes, elem.Value.(string))
	}
	return json.Marshal(struct {
		Collected           bool             `json:"collected"`
		RetainedModels      int              `json:"retainedModels"`
		ModelHashes         []string         `json:"modelHashes"`
		HeapAllocBytes      uint64           `json:"heapAllocBytes"`
		HeapInUseBytes      uint64           `json:"heapInUseBytes"`
		HeapSysBytes        uint64           `json:"heapSysBytes"`
		HeapIdleBytes       uint64           `json:"heapIdleBytes"`
		HeapReleasedBytes   uint64           `json:"heapReleasedBytes"`
		TotalAllocBytes     uint64           `json:"totalAllocBytes"`
		HeapObjects         uint64           `json:"heapObjects"`
		SysBytes            uint64           `json:"sysBytes"`
		NumGC               uint32           `json:"numGC"`
		NextGCBytes         uint64           `json:"nextGCBytes"`
		LastGCUnixMillis    uint64           `json:"lastGCUnixMillis"`
		PauseTotalNanos     uint64           `json:"pauseTotalNanos"`
		GCCPUFraction       float64          `json:"gcCPUFraction"`
		Goroutines          int              `json:"goroutines"`
		GCPercent           string           `json:"gcPercent"`
		MemoryLimitBytes    string           `json:"memoryLimitBytes"`
		PhaseCollections    uint64           `json:"phaseCollections"`
		LastPhaseCollection *phaseCollection `json:"lastPhaseCollection,omitempty"`
	}{
		Collected:           req.Collect,
		RetainedModels:      len(hashes),
		ModelHashes:         hashes,
		HeapAllocBytes:      m.HeapAlloc,
		HeapInUseBytes:      m.HeapInuse,
		HeapSysBytes:        m.HeapSys,
		HeapIdleBytes:       m.HeapIdle,
		HeapReleasedBytes:   m.HeapReleased,
		TotalAllocBytes:     m.TotalAlloc,
		HeapObjects:         m.HeapObjects,
		SysBytes:            m.Sys,
		NumGC:               m.NumGC,
		NextGCBytes:         m.NextGC,
		LastGCUnixMillis:    m.LastGC / 1000000,
		PauseTotalNanos:     m.PauseTotalNs,
		GCCPUFraction:       m.GCCPUFraction,
		Goroutines:          runtime.NumGoroutine(),
		GCPercent:           strconv.FormatUint(gcSettings[0].Value.Uint64(), 10),
		MemoryLimitBytes:    strconv.FormatUint(gcSettings[1].Value.Uint64(), 10),
		PhaseCollections:    s.phaseCollections,
		LastPhaseCollection: s.lastPhaseCollection,
	})
}

// phaseCollection records a collection at an application work boundary.
type phaseCollection struct {
	Method               string  `json:"method"`
	AllocatedBytes       uint64  `json:"allocatedBytes"`
	BeforeHeapAllocBytes uint64  `json:"beforeHeapAllocBytes"`
	AfterHeapAllocBytes  uint64  `json:"afterHeapAllocBytes"`
	ElapsedMillis        float64 `json:"elapsedMillis"`
}
