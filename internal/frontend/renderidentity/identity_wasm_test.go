//go:build js || wasip1

package renderidentity

import (
	"errors"
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"testing"
)

func TestIdentityExplicitlyUnavailable(t *testing.T) {
	data := view.Data{Nodes: []view.NodeData{{ID: "unchanged"}}}
	var unsupported *UnsupportedPlatformError
	if err := Bind(&data, nil, nil); !errors.As(err, &unsupported) {
		t.Fatalf("want explicit unsupported capability, got %v", err)
	}
	if data.Nodes[0].ID != "unchanged" || data.Nodes[0].SemanticID != "" {
		t.Fatal("unsupported binding changed native data")
	}
}
