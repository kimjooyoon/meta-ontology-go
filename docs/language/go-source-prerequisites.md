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

The existing `scripts/refactor-metrics -plan` entry point uses these observations
before assigning an independence group to a Go split. An unmet or unknown first
candidate leaves room for a later prepared candidate in that group. The plan
retains the original indicators, deferred repairs, file hashes and policy hash;
a shortfall remains unresolved. Ordinary metric collection starts no planner.
The CLI checks the repository revision and supplied metric policy. Its existing
minimum of two independent operations still applies.

The executor checks selected input bytes inside its disposable workspace before
the Git snapshot and native process. Changed bytes stop that operation. Other
operation inputs can still prevent execution; their failures retain the inspected
candidate population. Optional preparation fields leave legacy plans on their
existing validation path.

These connections are source development tooling. Readonly preparation still
leaves projected conformance unobserved. Physical successor handling and complete
operation verification remain in
[issue #1782](https://github.com/kimjooyoon/meta-ontology-go/issues/1782), with their
original scope and denominator.
