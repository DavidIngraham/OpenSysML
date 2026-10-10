package engine

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRenderViewRefusesArtifact(t *testing.T) {
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range []string{`{}`, `{"form":"dot"}`} {
		_, err := e.Call(context.Background(), "RenderView", []byte(`{"modelHash":"missing","view":"V","artifact":`+artifact+`}`))
		var refusal *Error
		if !errors.As(err, &refusal) || refusal.Code != codeUnimplemented || !strings.Contains(refusal.Message, "render_view_artifact") {
			t.Fatalf("artifact request = %v", err)
		}
	}
}
