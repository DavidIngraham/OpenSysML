- The combined browser engine retains full native API export, declaration projection,
canonical rendering identities and exact model release on the current upstream.
Conversion is supplied by the combined frontend so the standalone execution engine
keeps its converter-free dependency boundary and unchanged 11.3 MB gzip budget.

The combined profile has a 12.6 MB gzip budget (Go 1.27.4), measured at about
12.45 MB with the required full export contract; the unextended upstream combined
module is about 11.45 MB. Default rendering keeps the protobuf JSON shape; native
state/parameter metadata is requested through semantic identity or the independent
`includeRenderingMetadata` flag. Fault-tree gate SVG glyphs remain model notation.
