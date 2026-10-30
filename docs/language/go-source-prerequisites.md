# Go source rewrite prerequisites

The compiler's source development API inspects a proposed Go declaration split
without writing files. The operation contract retains three separate stages:

```gooo
activity ProveGoSplitRenderedCapacity(FileInput) -> RenderedCapacityObligation
activity ProveGoSplitDirectoryCapacity(RenderedCapacityObligation) -> DirectoryCapacityObligation
activity ProveGoSplitProjectedConformance(DirectoryCapacityObligation) -> ProjectedConformanceObligation
```

The [actual contract](../../internal/meta/generation/operation-input-contract.gooo)
assigns a stable ID to each stage. Native admission checks the typed dependency
chain and preserves each physical declaration's original source span. Renaming
these activity labels while retaining their IDs preserves their meaning.

`generation.InspectGoSplit(root, subject, metrics)` uses the same renderer and
directory checks as the source-splitter executor. Its result retains the exact
SHA-256 of the bytes parsed, the contract's source and semantic identities, and
the next-step policy's own source-byte identity. Directory observations keep the
scope supplied by the caller. The file hash identifies one input, rather than a
complete dependency set or a later publication authorization.

Each stage is `OBSERVED`, `UNMET`, `UNKNOWN`, or `UNOBSERVED`. A failed rendered
partition is an observation about the current planner. Missing directory data
stays unknown. Source bytes, permissions, directory entries and metric debt are
preserved on refusal. Readonly preparation leaves projected package conformance
unobserved, including when the first two stages are satisfied.

The [next-step program](../../internal/meta/generation/go-split-selection.gooo)
owns the readiness and next-operation rules:

| Native observation | Suggested next operation |
| --- | --- |
| Current planner cannot render the split | Extract a function |
| Formatted source requires no declaration split | Refresh source metrics |
| Source could not be inspected | Observe source |
| Directory capacity or shape prevents the split | Repack the directory |
| Directory observation is missing | Observe the directory |
| Both preparation stages are satisfied | Verify the projected workspace |

`preparation.NextStep()` calls its generated, typed Go library. It starts no
compiler process or model inference. The ordinary projection test compares the
committed library to its authored Gooo source. Regenerate it with
`go generate ./internal/meta/generation` after changing that source.

This is an internal source development API. Automatic candidate filtering,
execution-time source revalidation and physical successor handling still need
their consumer connections. [Issue #1782](https://github.com/kimjooyoon/meta-ontology-go/issues/1782)
tracks that complete operation; its remaining obligations retain their original
scope and denominator.
