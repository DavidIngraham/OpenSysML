//go:build js || wasip1

package renderidentity

import (
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
)

// Bind rejects optional conversion identities on lightweight WASM engines.
// Ordinary rendering does not call Bind and remains available.
func Bind(_ *view.Data, _ []Document, _ *symbols.Index) error {
	return &UnsupportedPlatformError{}
}
