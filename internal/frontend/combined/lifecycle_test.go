// Copyright 2025 Open‐MBEE Foundation. All rights reserved.
// Use of this source code is governed by the LICENSE file.

package combined

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/frontend/jsonrpc"
)

func releaseTestModel(t *testing.T, s *Server, hash string) (bool, int) {
	t.Helper()
	params, err := json.Marshal(map[string]string{"modelHash": hash})
	if err != nil {
		t.Fatal(err)
	}
	body, err := s.Call(context.Background(), "ReleaseModel", params)
	if err != nil {
		t.Fatal(err)
	}
	var answer struct {
		Released       bool `json:"released"`
		RetainedModels int  `json:"retainedModels"`
	}
	if err := json.Unmarshal(body, &answer); err != nil {
		t.Fatal(err)
	}
	return answer.Released, answer.RetainedModels
}

func TestReleaseModelDisposesBothCachesAndKeepsOtherModels(t *testing.T) {
	s := newRetentionTestServer(t)
	hashes := parseRetentionModels(t, s, 3, 0)
	callRetentionMethod(t, s, "Evaluate", hashes[0])
	callRetentionMethod(t, s, "GetSymbol", hashes[0])
	if released, count := releaseTestModel(t, s, hashes[0]); !released || count != 2 {
		t.Fatalf("release = %v, %d; want true, 2", released, count)
	}
	assertRetentionNotFound(t, s, "Evaluate", hashes[0])
	assertRetentionNotFound(t, s, "GetSymbol", hashes[0])
	assertRetentionNotFound(t, s, "GetDiagnostics", hashes[0])
	for _, hash := range hashes[1:] {
		callRetentionMethod(t, s, "Evaluate", hash)
		callRetentionMethod(t, s, "GetDiagnostics", hash)
	}
	if released, count := releaseTestModel(t, s, hashes[0]); released || count != 2 {
		t.Fatalf("repeated release = %v, %d; want false, 2", released, count)
	}
	if released, count := releaseTestModel(t, s, "unknown"); released || count != 2 {
		t.Fatalf("unknown release = %v, %d; want false, 2", released, count)
	}
	if got := parseRetentionModels(t, s, 1, 0)[0]; got != hashes[0] {
		t.Fatalf("reparse = %q; want stable hash %q", got, hashes[0])
	}
	callRetentionMethod(t, s, "Evaluate", hashes[0])
	callRetentionMethod(t, s, "GetSymbol", hashes[0])
}

func TestReleaseModelRejectsInvalidRequestsWithoutEviction(t *testing.T) {
	s := newRetentionTestServer(t)
	hash := parseRetentionModels(t, s, 1, 0)[0]
	for _, params := range []string{`{}`, `{"modelHash":42}`, `{`} {
		_, err := s.Call(context.Background(), "ReleaseModel", []byte(params))
		var callErr *jsonrpc.Error
		if !errors.As(err, &callErr) || callErr.Code != jsonrpc.CodeInvalidArgument {
			t.Errorf("ReleaseModel(%s) = %v; want InvalidArgument", params, err)
		}
	}
	callRetentionMethod(t, s, "Evaluate", hash)
	callRetentionMethod(t, s, "GetSymbol", hash)
}

func TestExplicitReleaseLeavesLRURetentionCoherent(t *testing.T) {
	s := newRetentionTestServer(t)
	hashes := parseRetentionModels(t, s, maxModels, 0)
	releaseTestModel(t, s, hashes[0])
	added := parseRetentionModels(t, s, 1, maxModels)[0]
	for _, hash := range hashes[1:] {
		callRetentionMethod(t, s, "GetDiagnostics", hash)
		callRetentionMethod(t, s, "Evaluate", hash)
	}
	parseRetentionModels(t, s, 1, maxModels+1)
	assertRetentionNotFound(t, s, "GetDiagnostics", added)
	assertRetentionNotFound(t, s, "Evaluate", added)
}

func TestReleaseModelBoundsRepeatedCandidateLifecycle(t *testing.T) {
	s := newRetentionTestServer(t)
	current := parseRetentionModels(t, s, 1, 0)[0]
	for i := 1; i <= 40; i++ {
		candidate := parseRetentionModels(t, s, 1, i)[0]
		callRetentionMethod(t, s, "Evaluate", current)
		callRetentionMethod(t, s, "Evaluate", candidate)
		if released, count := releaseTestModel(t, s, current); !released || count != 1 {
			t.Fatalf("cycle %d release = %v, %d; want true, 1", i, released, count)
		}
		assertRetentionNotFound(t, s, "GetDiagnostics", current)
		assertRetentionNotFound(t, s, "Evaluate", current)
		current = candidate
	}
	callRetentionMethod(t, s, "Evaluate", current)
	callRetentionMethod(t, s, "GetDiagnostics", current)
}

func TestMemoryStatsReportHeapAndDoNotTouchRetention(t *testing.T) {
	s := newRetentionTestServer(t)
	hashes := parseRetentionModels(t, s, 2, 0)
	var previousGC uint32
	for _, collect := range []bool{false, true} {
		params, err := json.Marshal(map[string]bool{"collect": collect})
		if err != nil {
			t.Fatal(err)
		}
		body, err := s.Call(context.Background(), "GetMemoryStats", params)
		if err != nil {
			t.Fatal(err)
		}
		var answer struct {
			Collected       bool     `json:"collected"`
			RetainedModels  int      `json:"retainedModels"`
			ModelHashes     []string `json:"modelHashes"`
			HeapAllocBytes  uint64   `json:"heapAllocBytes"`
			HeapSysBytes    uint64   `json:"heapSysBytes"`
			TotalAllocBytes uint64   `json:"totalAllocBytes"`
			NumGC           uint32   `json:"numGC"`
		}
		if err := json.Unmarshal(body, &answer); err != nil {
			t.Fatal(err)
		}
		if answer.Collected != collect || answer.RetainedModels != 2 || !reflect.DeepEqual(answer.ModelHashes, []string{hashes[1], hashes[0]}) {
			t.Fatalf("memory stats lost retention state: %s", body)
		}
		if answer.HeapAllocBytes == 0 || answer.HeapSysBytes < answer.HeapAllocBytes || answer.TotalAllocBytes < answer.HeapAllocBytes {
			t.Fatalf("invalid Go heap metrics: %s", body)
		}
		if collect && answer.NumGC <= previousGC {
			t.Fatalf("explicit collection did not advance GC count: %s", body)
		}
		previousGC = answer.NumGC
	}
	callRetentionMethod(t, s, "Evaluate", hashes[0])
	callRetentionMethod(t, s, "GetSymbol", hashes[0])
}

func TestMemoryStatsExposeExactGCSettingsAndPhaseEvidence(t *testing.T) {
	s := newRetentionTestServer(t)
	body, err := s.Call(context.Background(), "GetMemoryStats", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	var stats struct {
		NextGCBytes         uint64           `json:"nextGCBytes"`
		LastGCUnixMillis    uint64           `json:"lastGCUnixMillis"`
		PauseTotalNanos     uint64           `json:"pauseTotalNanos"`
		GCCPUFraction       float64          `json:"gcCPUFraction"`
		Goroutines          int              `json:"goroutines"`
		GCPercent           string           `json:"gcPercent"`
		MemoryLimitBytes    string           `json:"memoryLimitBytes"`
		PhaseCollections    uint64           `json:"phaseCollections"`
		LastPhaseCollection *phaseCollection `json:"lastPhaseCollection"`
	}
	if err := json.Unmarshal(body, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.NextGCBytes == 0 || stats.Goroutines < 1 || stats.GCCPUFraction < 0 || stats.GCCPUFraction > 1 {
		t.Fatalf("invalid GC diagnostics: %s", body)
	}
	for _, setting := range []string{stats.GCPercent, stats.MemoryLimitBytes} {
		if _, err := strconv.ParseUint(setting, 10, 64); err != nil {
			t.Fatalf("GC setting is not an exact decimal string: %q", setting)
		}
	}
	if stats.PhaseCollections != 0 || stats.LastPhaseCollection != nil {
		t.Fatalf("inspection itself created a phase collection: %s", body)
	}
}
