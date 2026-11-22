# Combined service model lifecycle

The combined native and WASM service retains parsed models by content hash. Its validation and execution frontends retain separate model state over one frozen standard-library index. The shared least-recently-used bound remains 16 models. Clients that prepare temporary or replacement versions should release hashes once no current or pending operation needs them, rather than waiting for that bound.

The combined service handles calls sequentially. Lifecycle methods apply between calls; they do not cancel a running parse, export, render or evaluation. A client must coordinate asynchronous ownership before releasing a model.

## Explicit release

Capability `release_model` provides `ReleaseModel`:

```json
{"modelHash":"<hash returned by ParseSources or ParseFile>"}
```

The response contains `released` and `retainedModels`. A known hash is removed from both the validation and execution caches and from shared retention bookkeeping. Unknown or already released hashes return `released: false`; an empty or nonstring hash is an invalid argument. In a browser WASM build, disposing an existing model collects at that safe boundary when the current allocated Go heap is at least 64 MiB. Unknown releases do not trigger collection.

Subsequent calls naming a released hash return the existing model-not-found status. Parsing the same source again restores the same content hash and valid native state. Releasing another hash leaves current models and the frozen library available.

## Memory measurements

Capability `memory_stats` provides `GetMemoryStats`:

```json
{"collect":false}
```

The response contains:

- `collected`: whether this call explicitly requested a completed Go garbage collection.
- `retainedModels` and `modelHashes`: combined cache count and hashes in most-recently-used order. Inspection does not change that order.
- `nextGCBytes`, `lastGCUnixMillis`, `pauseTotalNanos`, `gcCPUFraction` and `goroutines`: collection goal, last completion time, cumulative pause duration, CPU fraction and goroutine count.
- `gcPercent` and `memoryLimitBytes`: exact decimal strings for the active Go GC settings; the unlimited setting exceeds JavaScript integer precision.
- `phaseCollections` and optional `lastPhaseCollection`: browser work-boundary collections, including method, allocated bytes, heap allocation before/after and duration.
- `heapAllocBytes`, `heapInUseBytes`, `heapSysBytes`, `heapIdleBytes`, `heapReleasedBytes`, `totalAllocBytes`, `heapObjects`, `sysBytes` and `numGC`: the corresponding Go `runtime.MemStats` values.

Set `collect: true` only when an explicit collection is appropriate for diagnosis. `heapAllocBytes` includes unreachable allocations not yet collected; after collection it measures remaining Go heap allocation, not semantic model size. `totalAllocBytes` is cumulative and does not decrease when models are released.

WASM linear memory is a distinct measurement available from the host's `WebAssembly.Memory.buffer.byteLength`. Go heap arena reservation and linear memory can remain at a previous high-water mark after live heap usage decreases. Neither that high-water mark nor the runtime out-of-memory message alone establishes the size of live SysML objects.

These methods extend the combined JSON service; they do not change the standalone core, engine or gRPC protocol.


## Browser work boundaries

Browser WASM has a 32-bit linear address space. Parse, render and export can create substantial transient native state that the normal collector eventually reclaims, after linear memory has already grown close to that address limit. On the browser target, the combined service samples cumulative allocation around `ParseSources`, `ParseFile`, `RenderView` and `ExportModel`. After the native work frames return, it collects when that call allocated at least 64 MiB and the current allocated Go heap is at least 64 MiB. Small cache hits and ordinary symbol/value queries do not force collection.

This 64 MiB threshold is a work-phase garbage-collection budget, not a live-model limit or a guarantee that a larger model will fit. It does not remove reachable semantic state, change export membership or lower the model retention bound. Explicit ownership and release still control which versions remain reachable. Native nonbrowser builds retain their usual Go GC behavior.


## Large response buffers

JSON response buffers of at least 16 MiB use 4 MiB allocation classes, adding less than 25 percent spare capacity. Nearby edited models can reuse the same large allocation size instead of requiring a slightly larger contiguous block for each response. Smaller responses retain their existing encoding path. Normal buffer growth still handles unusually escape-heavy input; content is never truncated to the reservation.

Native export responses still use the standard JSON encoder, including RawMessage validation, HTML escaping and field order. Large responses return the encoder destination's storage directly. JavaScript success envelopes use a preallocated string builder, avoiding the separate formatter buffer and final string copy. Wire bytes remain unchanged; spare capacity is neither serialized nor a model-content limit. These buffers are ordinary garbage-collected allocations, without a persistent response pool.
