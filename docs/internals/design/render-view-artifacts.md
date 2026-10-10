# Formatted view artifacts over the service

## Contract and compatibility

Extend `RenderView` with an optional `RenderArtifactOptions artifact` request (field 4) and optional `RenderArtifact artifact` response (field 11). An absent request keeps the existing structured-data-only behavior. A present request requires an explicit form. An empty options message is an invalid request, not an implicit format choice. The new `render_view_artifact` capability is checked by the service and every client before sending artifact options; old servers must never silently ignore them.

Options are `form`, `palette`, `drawing_style`, `overlay`, `direction`, `unplaced`, `width`, and `link_template`. The existing top-level `ports` selection applies to both data and artifact. Defaults match the core writers: no palette, pilot style, no overlay, kind-specific direction, omit unplaced nodes, unbounded text width, and no links. Width is nonnegative. All named options are validated even when the chosen format does not use them. Invalid options and incompatible form/kind pairs produce INVALID_ARGUMENT; missing models or view targets retain NOT_FOUND; unavailable capabilities produce UNIMPLEMENTED.

The artifact contains its actual `form`, `media_type`, and UTF-8 `text`. Supported forms are text, Mermaid, Markdown, DOT, PlantUML, D2, CSV, and TSV, restricted by the existing kind/form support matrix. Plain diagram source uses `text/plain; charset=utf-8`; Markdown, CSV, and TSV use their corresponding text media types. Consumers use `form` to identify the diagram language. There are no subprocesses or image converters. PNG, SVG, PDF, view discovery, targetless pseudo-views, and bulk output are separate features.

## Shared implementation

The service resolves a declared view or targeted pseudo-view through the existing `view.RenderNamed` path. It uses the same `view.Rendering.WriteWith` writers as CLI/LSP, never reimplements a renderer in a client. Verdict overlays use the existing runtime requirement-verdict provider and reject unsupported view kinds. Source links are generated from the renderer's source sites and the requested template; the server does not visit links or write artifacts to its filesystem.

Structured nodes, edges, rows, optional layout, and notices remain present. Ports and semantic overlays are consistent across data and artifact. Presentation-only settings can omit unplaced nodes from an artifact while the structured response retains their underlying data. An additive `RenderNode.verdict` field preserves the semantic verdict the core overlay already produces, with mappings in every client. Existing modeled styles are not overwritten merely to reflect a writer palette. The response is assembled after writing to retain rendering notices. Writer-specific notices, such as DOT placement omissions, remain comments in the artifact. Overlay runtimes use `Service.newRuntimeOver` to retain service budgets and the context-bound tool runner. A client rejects a successful artifact request whose response omits the artifact message. Runtime contexts and model workers are released on every exit, and request cancellation follows the existing service runtime conventions.

## Language clients and transport adapters

Regenerate Go, Java, Python, TypeScript, and Rust protobuf outputs and the Rust conformance descriptor using repository-pinned generators. Add native optional artifact options/results and capability constants to Python, Go (in-process and remote), Node/TypeScript, Java, Rust, Julia, and MATLAB/Octave without changing existing signatures incompatibly. Preserve omitted versus present artifact messages and empty artifact text. Extend conformance adapters so requests and responses retain every new field. The internal browser/WASM engine uses a separate drawing-data contract and must not falsely advertise or silently accept the new capability.

## Test contract

Test data-only compatibility; presence semantics; all eight forms using real graph/table models; wrong-kind forms; invalid option names, width, and templates; port selection; Unicode and empty renderings; styles, palettes, positioned nodes, links, and verdict overlays; artifact/data consistency; notices; withheld capabilities; and source/model-not-found errors. Compare service artifacts with shared-writer or CLI output rather than maintaining a second implementation. Add transport conformance and native client encoding/decoding and live-service tests, including optional response absence and explicit empty text.

## Delivery and validation

Implement schema/server first, then clients and conformance runners, then reference documentation and a changelog fragment. Independently review this design before implementation and the completed diff before delivery. Address findings with regression tests where appropriate.

Required gates are `go build ./...`, `go vet ./...`, empty `gofmt -l .`, and `go test ./...`; run focused race tests, protobuf lint/breaking checks, changelog validation, language client suites and conformance checks. Provision the repository's required corpus fixtures and report unavailable external-tool gates honestly. The PR targets `develop`; it describes implemented behavior and actual validation, with no implied claim for skipped checks.
