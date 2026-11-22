package export

import (
	"fmt"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/parser"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
	"github.com/Open-MBEE/OpenSysML/internal/translate/rdf"
	"testing"
)

func capacityDocuments(t *testing.T) []ModelDocument {
	t.Helper()
	texts := []string{
		"package A { import External::*; part unit : Thing; } package B; package C;",
		"package D; package E;",
		"library package External { part def Thing; }",
	}
	documents := make([]ModelDocument, len(texts))
	for i, text := range texts {
		file := source.New(fmt.Sprintf("capacity%d.sysml", i), []byte(text))
		p := parser.New(file)
		documents[i] = ModelDocument{File: file, Root: p.ParseFile(), Referenced: i == 2}
		if len(p.Diagnostics) > 0 {
			t.Fatal(p.Diagnostics)
		}
	}
	return documents
}

func TestMergedDocumentStorageDoesNotGrowForMultipleRoots(t *testing.T) {
	docs := capacityDocuments(t)
	_, encoders, err := modelEncodersOf(docs, IDQualifiedName)
	if err != nil {
		t.Fatal(err)
	}
	upperBound := 0
	for i, e := range encoders {
		if docs[i].Referenced {
			continue
		}
		if err := e.encodeDocument(docs[i].Root); err != nil {
			t.Fatal(err)
		}
		upperBound += e.graph.Len()
		for _, triple := range e.graph.Triples() {
			if triple.Predicate.Value == rdf.OpenSysML+xSourceLanguage {
				upperBound++
			}
		}
	}
	merged := mergeDocumentGraphs(encoders, docs)
	if cap(merged.Triples()) != graphTripleCapacity(upperBound) {
		t.Fatalf("merged graph did not retain its bounded %d-triple input class: len=%d cap=%d", upperBound, merged.Len(), cap(merged.Triples()))
	}
	for _, root := range unownedElements(merged) {
		if _, ok := merged.Lexical(root, rdf.OpenSysML+xSourceDocument); !ok {
			t.Fatalf("native root lost document provenance: %s", root)
		}
	}
}

func TestRootNamespaceStorageDoesNotGrowForMultipleRoots(t *testing.T) {
	for _, count := range []int{1, 2, 8} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			docs := capacityDocuments(t)
			if count == 1 {
				docs = docs[:1]
				file := source.New("single.sysml", []byte("package A;"))
				docs[0] = ModelDocument{File: file, Root: parser.New(file).ParseFile()}
			}
			graph, err := ModelToRDFWith(docs, IDQualifiedName)
			if err != nil {
				t.Fatal(err)
			}
			if count == 8 {
				for i := 0; i < 8; i++ {
					subject := rdf.ElementIRIForID(fmt.Sprintf("extra%d", i))
					graph.Add(subject, rdf.IRI(rdf.RDFType), rdf.SysMLTerm("Package"))
					graph.Add(subject, rdf.SysMLTerm(pElementID), rdf.String(fmt.Sprintf("extra%d", i)))
				}
			}
			roots := unownedElements(graph)
			wrappers := rootWrappers(graph, roots)
			upperBound := graph.Len() + 21*len(roots) + 2*len(wrappers)
			for _, wrapper := range wrappers {
				if len(wrapper.roots) > 1 {
					upperBound += 3
				}
			}
			wrapped, err := withRootNamespace(graph)
			if err != nil {
				t.Fatal(err)
			}
			if cap(wrapped.Triples()) != graphTripleCapacity(upperBound) {
				t.Fatalf("wrapping did not retain its bounded %d-triple class: len=%d cap=%d roots=%d documents=%d", upperBound, wrapped.Len(), cap(wrapped.Triples()), len(roots), len(wrappers))
			}
		})
	}
}
