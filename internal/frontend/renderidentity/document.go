package renderidentity

import (
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
)

// Document supplies analyzed source without importing a conversion backend.
type Document struct {
	File *source.SourceFile
	Root *ast.RootNamespace
}

// UnsupportedPlatformError reports an explicitly requested unavailable capability.
type UnsupportedPlatformError struct{}

func (*UnsupportedPlatformError) Error() string {
	return "semantic rendering identity is unavailable on JS/WASI; use the native engine or gRPC service"
}
