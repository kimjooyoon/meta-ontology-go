# Manual-source-registration-edit-free registry projection

This experiment keeps concept registration local and makes the shared view a
deterministic projection. The bounded slice has exactly three production
inputs:

- `examples/language-syntax-roundtrip/concept.manifest.json`
- `examples/language-semantic-model/concept.manifest.json`
- `examples/toolchain-conformance/concept.manifest.json`

Each manifest owns its concept-local code and metric bindings, use cases, raw
resource references, and declared denominators. The producer discovers and
sorts those inputs, verifies ownership paths, non-empty bindings, structured
metric bindings, resource digests, and source-derived denominators, then
generates catalog, corpus, registry, denominator, documentation, manifest index,
projection, and digest outputs. Binding entries reconnect raw source, semantic
digest, consumer entry point, and observed output digest; documentation or
workflow self-search is not sufficient. The root topology and root README remain
explicit exceptions; this is a bounded vertical slice rather than a migration
of every existing global file.

## Fixed regression denominator

The baseline has exactly 12 observed manual shared-source touchpoints. A real
temporary filesystem intervention adds one fourth local manifest and runs
discovery plus generation. The existing 12 source paths are compared by bytes
digest before and after. The same intervention measures changed generated
outputs among exactly 8 outputs, and runs the independent conformance consumer
against the generated projection. No separate production/compiler adoption
evidence is available in this bounded slice.

The three surfaces are intentionally distinct:

| Surface | Measured result | Meaning |
| --- | ---: | --- |
| existing shared source touchpoints | 12/12 baseline | current human-edited registration surface |
| generator-changed shared outputs | 6/8 for the new fixture | generated projection change surface, not zero conflict |
| independent conformance consumer | 1/1 | independent raw-manifest reconstruction equals projection |
| production adoption | 0/1 UNKNOWN | no separate production/compiler consumer evidence |
| manual source registration edits required | 0/12 after fixture | no existing source file was edited |

The corrected toolchain denominator reconciles `181` cases from the parsed
machine-readable corpus artifact. Use-case execution receipt evidence is not
observable in this bounded run, so it is recorded as `UNKNOWN` with completed
numerator `0/1`; no number is extracted from prose. A regression clone with
the prior declared `160` cases must produce `FAIL_CLOSED / DENOMINATOR_SOURCE_MISMATCH`,
with declared and calculated values side by side in the receipt.

## Meaning gates

The proof keeps raw manifest bytes separate from the semantic manifest view.
Semantic and comment-only interventions render actual changed JSON bytes,
compute the digest from those bytes, decode them again, and only then project.
Replay twice and reversed manifest order must be byte-identical. The independent
consumer independently checks malformed, missing, stale, cross-directory,
missing-binding, duplicate-ID, and stale-denominator inputs fail closed while
preserving stage, step, and reason.

Claims carry distinct observed predicates, target addresses, target digests, and
an `OPEN` transition. Subject/conformance decision is separate from observed
predicate truth: an observed valid rejection can have decision `FAIL_CLOSED` and
still discharge the rejection proposition. `TRUE`, `FALSE`, and `UNKNOWN` select
`DISCHARGED`, `REFUTED`, and `OPEN`. Failure observations preserve exit code,
exact diagnostic JSON bytes and digest, raw input digest, and durable provenance
address. Repository tracked/untracked path-plus-content snapshots are read
before and after the complete proof. Net equality is reported as
`NET_STATE_EQUAL`; transient mutation and mutation authority remain `UNKNOWN`
unless separately observed.

`FOUNDATION`, `COHERENCE`, and `REGRESSION` are selected from every local
manifest and each strategy gates its own evidence. CI uses Go 1.27.0 and owns
format, compile, generation, freshness, and semantic verification. Local test
execution is intentionally zero.
