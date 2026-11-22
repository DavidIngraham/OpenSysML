package engine

import (
	"bytes"
	"encoding/json"

	"github.com/Open-MBEE/OpenSysML/internal/frontend/jsonrpc"
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
	if req.Projection != "" && req.Projection != "full" && req.Projection != "declarations" && req.Projection != "declaration-index" {
		return nil, statusError(codeInvalidArgument, "projection must be full, declarations or declaration-index")
	}
	cached, ok := e.get(req.ModelHash)
	if !ok {
		return nil, statusErrorf(codeNotFound, msgModelNotFound, req.ModelHash)
	}
	for _, doc := range cached.Documents {
		if len(doc.Errors) > 0 {
			return nil, statusError(codeFailedPrecondition, "cannot export a model with syntax errors")
		}
	}
	if e.extensions.ExportModel == nil {
		return nil, statusError(codeUnimplemented, "model export requires the combined frontend")
	}
	return e.extensions.ExportModel(cached.extensionContext(), req)

}

// MarshalExportResponse avoids an exact-length clone of a large encoded result.
func MarshalExportResponse(response JExportModelResponse) ([]byte, error) {
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
