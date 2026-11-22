# Downstream compatibility on the November baseline

The integration derives from official Open-MBEE/OpenSysML develop commit
`318c3e02b059fe67f869cfc2dc7a9624e5a39536`. The verified November 22 nightly
`00a68e3a61` is an ancestor. Downstream manifests identify the exact fork commit,
Git tree and toolchain used for an artifact; a dated nightly asset is not a source
pin or a substitute for the combined browser engine.

## Fork changes

The Blue Dog rollback pin is `23ca2e2f0f1fae4da6bc3e1488951cb017ed43e2`.
Its changes were deliberately replayed, retaining upstream's newer multi-document
Python generation, native rendering options and semantic/export infrastructure.

| Original change | Replay |
|---|---|
| `4fd6fe7f48722b71ef607febbef1aabebbe16f4c`, scalar dataclasses | `5b2e30cd2` |
| `e002f1668834f9409ee62200c8b89c0f5bf5e309`, declaration order and numeric scalars | `f61e52205` |
| `1a44d05cd6c8f3b2ff508ee16c4caa5d0aa55cd4`, failure analysis and fault-tree rendering | `bba91721e` |
| `d2962794588fb8904c43d60497e7c69c525124a0`, causal traversal | `bd646ae64`; obsolete self-model edits were omitted |
| `23ca2e2f0`, library pin | Replaced by a replay on current extension-library base `3faa2bf1ddda4f35c3ea1f0e2600b55a10f0e0f1` |

The extension library is pinned to
`38097bbc0198767590c58f89796a436678319924` in DavidIngraham's personal fork.
It preserves FailureAnalysis, FailureAnalysisViews and FaultTreeView, and the
browser's RequirementContext metadata. DesignFailureReview explicitly declares
its Boolean return multiplicity as `[1]`, preserving the scalar readiness API.
It establishes inventory and disposition checks, not engineering acceptance or
quantitative failure probabilities.

The separately reviewable browser layer starts at `dfc161b204d641a679b93838c0d956864c197a7e`.
It replays the combined-service contracts from SysPresent's fork base
`af37bbf155aea785b8ac494a14afe659244bbcdf` and carried patch SHA-256
`e8332e02f984358dd8c3da41fb9d3fa53cb8e140fa6b2a7a44bff0c43a76d908`.
The previous fully patched source tree was
`d85adf5b2e95aad223e944768fc37769e32acab4`.
The old gRPC/editor semantic-identity option was not replayed: these downstreams
use the combined JSON identity contract, while their gRPC consumers use the
current upstream declaration and execution interfaces.

## Native and browser interfaces

| Interface | Contract |
|---|---|
| Combined RenderView | Current upstream geometry, compartments and sequence proxies, with opt-in canonical semantic identities |
| Combined ExportModel | Native API JSON; declarations retain the complete graph's derived-property context and select source-backed objects |
| Combined ReleaseModel | Exact hash disposal in validation, execution and shared retention bookkeeping |
| Combined GetMemoryStats | Explicit native heap/cache and browser work-phase measurements |
| Combined GetSymbol and Evaluate | Native declarations, typed values, unset values and arbitrary-precision integer representations |
| GetElementProperties and Convert | Served by native gRPC; the combined JSON service already refused them before this upgrade and continues to report Unimplemented |

Declaration export uses upstream's element-form graph path. The general RDF
path eagerly completes inherited library memberships and is unsuitable for this
browser projection. Filtering graph triples before completing API properties
would change native objects, so the projection selects subjects at the native
API writer instead. Collection order and references to omitted objects remain
native.

Sequence event proxies and succession arrows are actual upstream records.
SysPresent places their native event identities on their native parent lifelines,
retains explicit orders separately, and does not turn proxies into participants
or infer new ordering from labels. Language-owned start/done and completion nodes
remain synthetic rather than acquiring project identities.

## Diagnostic and application boundaries

The unchanged eight Blue Dog inputs exit successfully on both the rollback pin
and the new baseline. The rollback pin emits no diagnostics; the new baseline
emits seven `name-conflict` warnings for a subject named `result` and its
redefinitions under BlueDog::Analyses::Weather::Sensitivity. The inherited
Case::result member already exists in the old library. The newer distinguishability
checks diagnose this existing name overlap. Exact diagnostic parity therefore
fails; the warning difference must remain visible in comparison evidence.
No model field is renamed and no diagnostic is suppressed for the upgrade.

The newer engine reports omitted parameter multiplicities as collections.
Blue Dog's application interfaces retain native parameter values and enforce one
value at their scalar boundaries. Explicitly ranged outputs remain sequences;
zero or multiple values cannot silently become a Boolean or number. Python still
owns SciPy/SLSQP, and no native experimental optimizer replaces it.

The C bridge passes the existing Python float inputs through the generated
compiler's numeric constructors and copies results through its numeric conversion
helpers. It keeps tagged-number and arbitrary-precision Integer layouts inside C,
rather than guessing those layouts in ctypes. Single calls and batches share that
bridge and retain explicit error and output-width checks.
