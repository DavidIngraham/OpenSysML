// Copyright 2025 Open‐MBEE Foundation. All rights reserved.
// Use of this source code is governed by the LICENSE file.

package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/frontend/jsonrpc"
)

func TestExportResponseEncodingMatchesStandardMarshal(t *testing.T) {
	zero, count, omitted := 0, 2, 19
	large := json.RawMessage(`[{"@id":"A","text":"<>&\u2028\u2029` + strings.Repeat("x", 17<<20) + `"}]`)
	responses := []JExportModelResponse{
		{},
		{Elements: json.RawMessage(`[]`), Experimental: true, Notice: "full <>&\nnotation"},
		{Elements: json.RawMessage(`[ { "@id": "a", "text": "<>&" } ]`), Experimental: true, Projection: "declarations", ElementCount: &count, OmittedElementCount: &omitted, CoverageNotice: "source-backed\nrecords"},
		{Elements: json.RawMessage(`null`), ElementCount: &zero},
		{Elements: large, Experimental: true, Notice: "large <&> export", Projection: "declarations", ElementCount: &count, OmittedElementCount: &omitted},
		{Elements: json.RawMessage("invalid")},
		{Elements: json.RawMessage(`"` + strings.Repeat("x", 17<<20))},
	}
	for i, response := range responses {
		want, wantErr := json.Marshal(response)
		got, gotErr := MarshalExportResponse(response)
		if (gotErr == nil) != (wantErr == nil) || gotErr != nil && gotErr.Error() != wantErr.Error() {
			t.Fatalf("case %d error=%v; want %v", i, gotErr, wantErr)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("case %d export bytes differ", i)
		}
		if gotErr == nil && len(response.Elements) >= 16<<20 {
			expected := jsonrpc.ResponseCapacity(len(got) + 1)
			if cap(got) != expected {
				t.Fatalf("large result cap=%d; want allocation class %d", cap(got), expected)
			}
		}
	}
}

func TestNearbyLargeExportResponsesShareAllocationClass(t *testing.T) {
	var capacity int
	for _, extra := range []int{0, 205, 411, 4096} {
		response := JExportModelResponse{Elements: json.RawMessage(`["` + strings.Repeat("x", (17<<20)+extra) + `"]`)}
		got, err := MarshalExportResponse(response)
		if err != nil {
			t.Fatal(err)
		}
		if capacity == 0 {
			capacity = cap(got)
		}
		if cap(got) != capacity {
			t.Fatalf("small model edit changed large buffer class from %d to %d", capacity, cap(got))
		}
	}
}

var exportResponseBenchmarkResult []byte

func BenchmarkLargeExportResponseBuffers(b *testing.B) {
	response := JExportModelResponse{Elements: json.RawMessage(`["` + strings.Repeat("x", 26<<20) + `"]`), Experimental: true, Projection: "declarations"}
	b.Run("marshal-baseline", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(response.Elements)))
		for i := 0; i < b.N; i++ {
			result, err := json.Marshal(response)
			if err != nil {
				b.Fatal(err)
			}
			exportResponseBenchmarkResult = result
		}
	})
	b.Run("bounded-encoder", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(response.Elements)))
		for i := 0; i < b.N; i++ {
			result, err := MarshalExportResponse(response)
			if err != nil {
				b.Fatal(err)
			}
			exportResponseBenchmarkResult = result
		}
	})
}

var exportWireBenchmarkResult string

func BenchmarkLargeExportWireBuffers(b *testing.B) {
	response := JExportModelResponse{Elements: json.RawMessage(`["` + strings.Repeat("x", 26<<20) + `"]`), Experimental: true, Projection: "declarations"}
	b.Run("marshal-formatter-baseline", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(response.Elements)))
		for i := 0; i < b.N; i++ {
			body, err := json.Marshal(response)
			if err != nil {
				b.Fatal(err)
			}
			exportWireBenchmarkResult = fmt.Sprintf(`{"jsonrpc":"2.0","id":null,"result":%s}`, body)
		}
	})
	b.Run("bounded-encoder-builder", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(response.Elements)))
		for i := 0; i < b.N; i++ {
			body, err := MarshalExportResponse(response)
			if err != nil {
				b.Fatal(err)
			}
			exportWireBenchmarkResult = jsonrpc.Envelope(body, nil)
		}
	})
}
