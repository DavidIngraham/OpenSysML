package combined

import (
	"encoding/json"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/engine"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/jsonrpc"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/renderidentity"
	"github.com/Open-MBEE/OpenSysML/internal/translate/export"
)

func modelExtensions() engine.ModelExtensions {
	return engine.ModelExtensions{RenderingMetadata: true, ExportModel: exportModel, RenderingIdentity: func(model *engine.ModelContext) (engine.RenderingIdentities, error) {
		documents := make([]renderidentity.Document, len(model.Documents))
		for i, d := range model.Documents {
			documents[i] = renderidentity.Document{File: d.File, Root: d.Root}
		}
		return renderidentity.NewInventory(documents, model.Resolver, model.SemanticModel())
	}}
}

func exportModel(model *engine.ModelContext, req *engine.JExportModelRequest) ([]byte, error) {
	documents := make([]export.ModelDocument, 0, len(model.Documents))
	for _, doc := range model.Documents {
		documents = append(documents, export.ModelDocument{File: doc.File, Root: doc.Root})
	}
	if req.Projection == "declaration-index" {
		entries, err := export.ModelDeclarationIndex(documents, model.Resolver, model.SemanticModel())
		if err != nil {
			return nil, jsonrpc.Errorf(jsonrpc.CodeInvalidArgument, "%s", err.Error())
		}
		return json.Marshal(struct {
			Projection       string                         `json:"projection"`
			DeclarationIndex []export.DeclarationIndexEntry `json:"declarationIndex"`
			Notice           string                         `json:"notice"`
		}{"declaration-index", entries, "Native declaration discovery metadata; request declarations or full for API JSON properties."})
	}
	if req.Projection == "declarations" {
		typed := model.SemanticModel()
		elements, counts, err := export.ModelToAPIJSONDeclarationsInModel(documents, model.Resolver, typed)
		if err != nil {
			return nil, jsonrpc.Errorf(jsonrpc.CodeInvalidArgument, "%s", err.Error())
		}
		return engine.MarshalExportResponse(engine.JExportModelResponse{Elements: elements, Experimental: true,
			Notice:     "The SysML to API-JSON mapping is experimental; preserve source and provenance.",
			Projection: "declarations", ElementCount: &counts.ElementCount, OmittedElementCount: &counts.OmittedElementCount,
			CoverageNotice: "Only source-backed declaration objects are included. References to omitted expression, membership and other infrastructure objects retain their native IDs. The complete parsed model remains available for native views and symbol queries."})
	}
	graph, err := export.ModelElementGraphWith(documents, export.IDQualifiedName)
	if err != nil {
		return nil, jsonrpc.Errorf(jsonrpc.CodeInvalidArgument, "%s", err.Error())
	}
	elements, err := export.WriteAPIJSON(graph)
	if err != nil {
		return nil, jsonrpc.Errorf(jsonrpc.CodeInvalidArgument, "%s", err.Error())
	}
	return engine.MarshalExportResponse(engine.JExportModelResponse{Elements: elements, Experimental: true, Notice: "The SysML to API-JSON mapping is experimental; preserve source and provenance."})
}
