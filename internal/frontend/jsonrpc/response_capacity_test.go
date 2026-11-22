// Copyright 2025 Open‐MBEE Foundation. All rights reserved.
// Use of this source code is governed by the LICENSE file.

package jsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestResponseCapacityKeepsNearbyLargeResponsesInOneBoundedClass(t *testing.T) {
	for _, size := range []int{0, 1, 1024, (16 << 20) - 1} {
		if got := ResponseCapacity(size); got != size {
			t.Errorf("small capacity(%d)=%d", size, got)
		}
	}
	for _, size := range []int{16 << 20, (16 << 20) + 1, 26 << 20, (26 << 20) + 123, 28 << 20, (28 << 20) + 1} {
		got := ResponseCapacity(size)
		if got < size || got-size >= 4<<20 || got%(4<<20) != 0 {
			t.Errorf("large capacity(%d)=%d; want bounded 4 MiB class", size, got)
		}
	}
	if ResponseCapacity(26<<20) != ResponseCapacity((26<<20)+4096) {
		t.Fatal("nearby model responses use different classes")
	}
	maxInt := int(^uint(0) >> 1)
	if got := ResponseCapacity(maxInt); got != maxInt {
		t.Fatalf("overflow capacity=%d; want %d", got, maxInt)
	}
}

func TestEnvelopePreservesExistingSuccessAndErrorBytes(t *testing.T) {
	for _, body := range [][]byte{
		nil, {}, []byte("null"), []byte(`{"value":"<>&\u2028","n":1}`),
		[]byte(" invalid JSON\x00\xff"), bytes.Repeat([]byte("x"), (16<<20)+17),
	} {
		got := Envelope(body, nil)
		want := fmt.Sprintf(`{"jsonrpc":"2.0","id":null,"result":%s}`, body)
		if got != want {
			t.Fatalf("success envelope changed for %d bytes", len(body))
		}
	}
	for _, err := range []error{Errorf(CodeInvalidArgument, "invalid <value> & \"quote\""), errors.New("unknown\nerror")} {
		code, message := ErrorParts(err)
		// json.Marshal supplies the existing message escaping contract.
		quoted, marshalErr := json.Marshal(message)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		want := fmt.Sprintf(`{"jsonrpc":"2.0","id":null,"error":{"code":%d,"message":%s}}`, code, quoted)
		if got := Envelope([]byte("ignored"), err); got != want {
			t.Fatalf("error envelope=%s; want %s", got, want)
		}
	}
}

var envelopeBenchmarkResult string

func BenchmarkLargeEnvelopeBuffers(b *testing.B) {
	body := []byte(`["` + strings.Repeat("x", 26<<20) + `"]`)
	b.Run("formatter-baseline", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(body)))
		for i := 0; i < b.N; i++ {
			envelopeBenchmarkResult = fmt.Sprintf(`{"jsonrpc":"2.0","id":null,"result":%s}`, body)
		}
	})
	b.Run("bounded-builder", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(body)))
		for i := 0; i < b.N; i++ {
			envelopeBenchmarkResult = Envelope(body, nil)
		}
	})
}
