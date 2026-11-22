package engine

import (
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/resolve"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/semantics"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/ast"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
)

// ModelDocument exposes immutable parsed source to a composed frontend.
type ModelDocument struct {
	File *source.SourceFile
	Root *ast.RootNamespace
}

// ModelContext shares native facts without making the execution engine link a converter.
type ModelContext struct {
	Documents []ModelDocument
	Resolver  *resolve.Resolver
	Index     *symbols.Index
	cached    *cachedModel
}

// SemanticModel materializes the same semantic side tables used by native rendering.
func (c *ModelContext) SemanticModel() *semantics.Model { return c.cached.typedModel() }

// RenderingIdentities binds a drawing through a composed frontend's native identity inventory.
type RenderingIdentities interface {
	Bind(*view.Data, *symbols.Index) error
}

// ModelExtensions are conversion contracts supplied by a frontend that owns export.
type ModelExtensions struct {
	RenderingMetadata bool
	ExportModel       func(*ModelContext, *JExportModelRequest) ([]byte, error)
	RenderingIdentity func(*ModelContext) (RenderingIdentities, error)
}

func (m *cachedModel) extensionContext() *ModelContext {
	c := &ModelContext{Resolver: m.resolver, Index: m.Index, cached: m, Documents: make([]ModelDocument, 0, len(m.Documents))}
	for _, d := range m.Documents {
		c.Documents = append(c.Documents, ModelDocument{File: d.Source, Root: d.Root})
	}
	return c
}
