package export

import (
	"bytes"
	"context"
	"encoding/json"
	"runtime"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/resolve"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/semantics"
	"github.com/Open-MBEE/OpenSysML/internal/translate/rdf"
)

// APIJSONExportStage records an optional declaration-export memory checkpoint.
// HeapAlloc includes unswept garbage; HeapSys is acquired arenas, not live objects.
type APIJSONExportStage struct {
	Document       string
	Stage          string
	Triples        int
	TripleCapacity int
	Subjects       int
	HeapAlloc      uint64
	HeapSys        uint64
	TotalAlloc     uint64
	NumGC          uint32
}

// APIJSONDeclarationsOptions reuses native semantics and optionally observes export stages.
type APIJSONDeclarationsOptions struct {
	Resolver *resolve.Resolver
	Model    *semantics.Model
	Observe  func(APIJSONExportStage)
}

func observeAPIJSONExport(observe func(APIJSONExportStage), document, stage string, graph *rdf.Graph) {
	if observe == nil {
		return
	}
	point := APIJSONExportStage{Document: document, Stage: stage}
	if graph != nil {
		point.Triples = graph.Len()
		point.TripleCapacity = cap(graph.Triples())
		point.Subjects = len(graph.Subjects())
	}
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	point.HeapAlloc, point.HeapSys, point.TotalAlloc, point.NumGC = stats.HeapAlloc, stats.HeapSys, stats.TotalAlloc, stats.NumGC
	observe(point)
}

// APIJSONProjectionCounts reports the declarations written and other elements omitted.
type APIJSONProjectionCounts struct {
	ElementCount        int
	OmittedElementCount int
}

// WriteAPIJSONDeclarations writes unchanged API objects for source-backed declarations.
// References retain their native IDs; omitted subjects acquire no API objects.
func WriteAPIJSONDeclarations(graph *rdf.Graph) ([]byte, APIJSONProjectionCounts, error) {
	var out bytes.Buffer
	w := apiJSONWriter{buf: &out, enc: json.NewEncoder(&out)}
	counts := APIJSONProjectionCounts{}
	out.WriteByte('[')
	err := apiJSONEachSelectedElement(graph, APIJSONOptions{}, func(current *rdf.Graph, subject rdf.Term) bool {
		if _, ok := current.Lexical(subject, rdf.OpenSysML+xSourceLine); !ok {
			counts.OmittedElementCount++
			return false
		}
		return true
	}, func(element apiJSONObject) error {
		if counts.ElementCount > 0 {
			out.WriteByte(',')
		}
		if err := w.object(element); err != nil {
			return err
		}
		counts.ElementCount++
		return nil
	})
	if err != nil {
		return nil, counts, err
	}
	out.WriteByte(']')
	return out.Bytes(), counts, nil
}

// ModelToAPIJSONDeclarations writes native source declarations with shared model identities.
// Derived properties use the complete native graph before subjects are selected.
func ModelToAPIJSONDeclarations(documents []ModelDocument) ([]byte, APIJSONProjectionCounts, error) {
	return ModelToAPIJSONDeclarationsInModel(documents, nil, nil)
}

// ModelToAPIJSONDeclarationsInModel reuses a caller's authoritative resolver and semantic model.
func ModelToAPIJSONDeclarationsInModel(documents []ModelDocument, res *resolve.Resolver, model *semantics.Model) ([]byte, APIJSONProjectionCounts, error) {
	return ModelToAPIJSONDeclarationsWithOptions(documents, APIJSONDeclarationsOptions{Resolver: res, Model: model})
}

// ModelToAPIJSONDeclarationsWithOptions exports declarations with optional stage observations.
func ModelToAPIJSONDeclarationsWithOptions(documents []ModelDocument, options APIJSONDeclarationsOptions) ([]byte, APIJSONProjectionCounts, error) {
	defer runtime.GC()
	res, model := options.Resolver, options.Model
	observeAPIJSONExport(options.Observe, "", "start", nil)
	if res != nil {
		for _, document := range documents {
			res.ResolveDocument(document.File.Name(), document.Root)
		}
	}
	if res == nil {
		res, model = analyzeModel(documents)
	}
	graph, err := declarationElementGraph(documents, res, model, options.Observe)
	if err != nil {
		return nil, APIJSONProjectionCounts{}, err
	}
	observeAPIJSONExport(options.Observe, "", "merged", graph)
	collectLargeExportPhase(graph)
	content, counts, err := WriteAPIJSONDeclarations(graph)
	observeAPIJSONExport(options.Observe, "", "json", nil)
	return content, counts, err
}

// declarationElementGraph returns without retaining per-document encoders beside the merged graph.
func declarationElementGraph(documents []ModelDocument, res *resolve.Resolver, model *semantics.Model, observe func(APIJSONExportStage)) (*rdf.Graph, error) {
	_, encoders, err := modelEncodersOver(documents, IDQualifiedName, res, model)
	if err != nil {
		return nil, err
	}
	return encodeModelDocumentsObserved(context.Background(), documents, encoders, addLinkedLibraryStubs, observe)
}
