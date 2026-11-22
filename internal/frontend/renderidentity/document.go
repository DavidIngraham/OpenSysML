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
