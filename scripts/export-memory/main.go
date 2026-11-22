package main

import (
	"encoding/json"
	"fmt"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/parser"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
	"github.com/Open-MBEE/OpenSysML/internal/translate/export"
	"os"
	"runtime"
	"time"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var raw []struct {
		Name    string
		Content string
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(err)
	}
	docs := make([]export.ModelDocument, len(raw))
	for i, doc := range raw {
		f := source.New(doc.Name, []byte(doc.Content))
		p := parser.New(f)
		docs[i] = export.ModelDocument{File: f, Root: p.ParseFile()}
		if len(p.Diagnostics) > 0 {
			panic(fmt.Sprintf("%s: %v", doc.Name, p.Diagnostics))
		}
	}
	runtime.GC()
	started := time.Now()
	enc := json.NewEncoder(os.Stdout)
	result, counts, err := export.ModelToAPIJSONDeclarationsWithOptions(docs, export.APIJSONDeclarationsOptions{Observe: func(stage export.APIJSONExportStage) {
		_ = enc.Encode(struct {
			ElapsedMS int64
			export.APIJSONExportStage
		}{time.Since(started).Milliseconds(), stage})
	}})
	if err != nil {
		panic(err)
	}
	if len(os.Args) > 2 {
		if err := os.WriteFile(os.Args[2], result, 0600); err != nil {
			panic(err)
		}
	}
	_ = enc.Encode(struct {
		Stage     string
		ElapsedMS int64
		Bytes     int
		Counts    export.APIJSONProjectionCounts
	}{"complete", time.Since(started).Milliseconds(), len(result), counts})
}
