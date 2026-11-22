package combined

import (
	"context"
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/core"
	"testing"
)

// Both the core and combined native wire paths retain arbitrary-precision integers.
func TestGetSymbolExactIntegerBounds(t *testing.T) {
	c, err := core.New()
	if err != nil {
		t.Fatal(err)
	}
	s, err := New("integer-bounds")
	if err != nil {
		t.Fatal(err)
	}
	for _, server := range []struct {
		name string
		call func(context.Context, string, []byte) ([]byte, error)
	}{{"core", c.Call}, {"combined", s.Call}} {
		t.Run(server.name, func(t *testing.T) {
			payload, _ := json.Marshal(map[string]any{"documents": []map[string]string{{"name": "integer.sysml", "content": `package Integers { private import ScalarValues::*; part source { attribute bounded:Integer=9223372036854775807; attribute minimum:Integer=-9223372036854775808; attribute huge:Integer=18446744073709551616; attribute small:Integer=-9223372036854775809; } }`}}})
			parsed, err := server.call(context.Background(), "ParseSources", payload)
			if err != nil {
				t.Fatal(err)
			}
			var model struct {
				ModelHash string `json:"modelHash"`
			}
			if err := json.Unmarshal(parsed, &model); err != nil {
				t.Fatal(err)
			}
			payload, _ = json.Marshal(map[string]string{"modelHash": model.ModelHash, "symbolId": "Integers::source"})
			body, err := server.call(context.Background(), "GetSymbol", payload)
			if err != nil {
				t.Fatal(err)
			}
			var reply struct {
				Symbol struct {
					Attributes []struct {
						Name  string
						Value map[string]string
					}
				}
			}
			if err := json.Unmarshal(body, &reply); err != nil {
				t.Fatalf("%v: %s", err, body)
			}
			values := map[string]map[string]string{}
			for _, a := range reply.Symbol.Attributes {
				values[a.Name] = a.Value
			}
			for _, want := range []struct{ name, arm, value string }{{"bounded", "intValue", "9223372036854775807"}, {"minimum", "intValue", "-9223372036854775808"}, {"huge", "bigIntValue", "18446744073709551616"}, {"small", "bigIntValue", "-9223372036854775809"}} {
				if len(values[want.name]) != 1 || values[want.name][want.arm] != want.value {
					t.Errorf("%s=%v want %s:%s", want.name, values[want.name], want.arm, want.value)
				}
			}
		})
	}
}
