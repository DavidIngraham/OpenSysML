# Native declaration export memory profile

Build from the repository root with go build -o build/export-memory/profile ./scripts/export-memory.
Run build/export-memory/profile documents.json declarations.json > stages.jsonl.

The input is a JSON array of native ParseSources document records, each containing
name and content. Source bytes remain local. The optional second argument saves
the exact declaration API JSON for semantic comparison. Standard output records
per-document authored, import, normative, source, projection and annotation stages,
then merged and JSON stages, with elapsed milliseconds and native coverage counts.

HeapAlloc is allocated Go heap at that checkpoint and can include unswept garbage.
A projected-after-gc stage follows the exporter's explicit collection and is a
retained-heap approximation. HeapSys is acquired Go arenas; neither value is WASM
linear memory or a continuous peak sample. Run the same source and binary target
before and after a change, and pair browser/WASM measurements with these native
stage records. A native run does not represent all browser worker model caches.
