package engine

import (
	"bytes"
	"encoding/json"

	"github.com/Open-MBEE/OpenSysML/internal/frontend/jsonrpc"
	"github.com/Open-MBEE/OpenSysML/internal/translate/export"
)

// JExportModelRequest selects a parsed model for API-JSON export.
type JExportModelRequest struct {
	ModelHash  string `json:"modelHash"`
	Projection string `json:"projection,omitempty"`
}

// JExportModelResponse preserves the experimental API mapping and its notice.
type JExportModelResponse struct {
	Elements            json.RawMessage `json:"elements"`
	Experimental        bool            `json:"experimental"`
	Notice              string          `json:"notice"`
	Projection          string          `json:"projection,omitempty"`
	ElementCount        *int            `json:"elementCount,omitempty"`
	OmittedElementCount *int            `json:"omittedElementCount,omitempty"`
	CoverageNotice      string          `json:"coverageNotice,omitempty"`
}

func (e *Engine) exportModel(req *JExportModelRequest) ([]byte, error) {
	if req.Projection != "" && req.Projection != "full" && req.Projection != "declarations" {
		return nil, statusError(codeInvalidArgument, "projection must be full or declarations")
	}
	cached, ok := e.get(req.ModelHash)
	if !ok {
		return nil, statusErrorf(codeNotFound, msgModelNotFound, req.ModelHash)
	}
	documents := make([]export.ModelDocument, 0, len(cached.Documents))
	for _, doc := range cached.Documents {
		if len(doc.Errors) > 0 {
			return nil, statusError(codeFailedPrecondition, "cannot export a model with syntax errors")
		}
		documents = append(documents, export.ModelDocument{File: doc.Source, Root: doc.Root})
	}
	if req.Projection == "declarations" {
		typed := cached.typedModel()
		elements, counts, err := export.ModelToAPIJSONDeclarationsInModel(documents, cached.resolver, typed)
		if err != nil {
			return nil, statusError(codeInvalidArgument, err.Error())
		}
		return marshalExportResponse(JExportModelResponse{Elements: elements, Experimental: true,
			Notice:     "The SysML to API-JSON mapping is experimental; preserve source and provenance.",
			Projection: "declarations", ElementCount: &counts.ElementCount, OmittedElementCount: &counts.OmittedElementCount,
			CoverageNotice: "Only source-backed declaration objects are included. References to omitted expression, membership and other infrastructure objects retain their native IDs. The complete parsed model remains available for native views and symbol queries."})
	}
	graph, err := export.ModelElementGraphWith(documents, export.IDQualifiedName)
	if err != nil {
		return nil, statusError(codeInvalidArgument, err.Error())
	}
	elements, err := export.WriteAPIJSON(graph)
	if err != nil {
		return nil, statusError(codeInvalidArgument, err.Error())
	}
	return marshalExportResponse(JExportModelResponse{Elements: elements, Experimental: true, Notice: "The SysML to API-JSON mapping is experimental; preserve source and provenance."})
}

// marshalExportResponse avoids an exact-length clone of a large encoded result.
func marshalExportResponse(response JExportModelResponse) ([]byte, error) {
	if len(response.Elements) < 16<<20 {
		return json.Marshal(response)
	}
	metadata := response
	metadata.Elements = nil
	encodedMetadata, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	// The metadata contains a four-byte null; Encode adds one final newline.
	size := len(response.Elements) + len(encodedMetadata) - 3
	var out bytes.Buffer
	out.Grow(jsonrpc.ResponseCapacity(size))
	if err := json.NewEncoder(&out).Encode(response); err != nil {
		return nil, err
	}
	return out.Bytes()[:out.Len()-1], nil
}
