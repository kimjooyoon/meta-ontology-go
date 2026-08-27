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
projection, and digest outputs. Binding source addresses resolve through a
complete module-aware, build-selected Go type check to the same `go/types`
object at declaration and registration use; package errors fail closed. Each
metric also binds its exact registration-relation literal occurrence, normalized
occurrence digest, and position-independent semantic relation digest. Comment-
only position shifts preserve semantic relations while changing raw addresses;
metric changes and unused, shadowed, cross-package, unrelated-use, and
unrelated-call counterexamples fail closed. The independent consumer reads
back its generated output artifact and records its actual path, bytes, and
digest in each binding receipt. Documentation or workflow self-search is not
sufficient. The root topology and root README remain explicit exceptions; this
is a bounded vertical slice rather than a migration of every existing global
file.

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

The fixed predicate inventory is 32 IDs: five conformance IDs
(`independent-manifest-order`, `independent-resource-digests`,
`independent-denominator-reconciliation`, `independent-binding-registry`, and
`independent-conformance-consumer`) plus 21 failure IDs
(`consumer-malformed-manifest`, `consumer-missing-manifest`,
`consumer-cross-directory-manifest`, `consumer-missing-binding`,
`consumer-stale-denominator`, `consumer-stale-generated-projection`,
`consumer-duplicate-stable-id`, `consumer-binding-self-search`,
`consumer-binding-output-digest-mismatch`, `consumer-binding-comment-only`,
`consumer-binding-unused-string`,
`consumer-binding-cross-package-same-name`, `consumer-binding-shadowed-local`,
`consumer-binding-unused-declaration`, `consumer-binding-unrelated-use`,
`consumer-binding-unresolved-import`, `consumer-binding-unrelated-type-error`,
`consumer-binding-metric-row-swap`, `consumer-binding-different-metric-literal`,
`consumer-binding-unrelated-call`, six strict receipt-boundary corruption
cases, and `classifier-success-exit-counterexample`). Claims are exactly 32;
failure and provenance predicates are exactly 27;
static failure contracts are exactly 8; typed declaration/use object tuples,
metric occurrences, semantic relation digests, and output row addresses are
each exactly 9. Each output receipt includes an exact row digest from the
consumer's embedded raw output artifact.

The conformance-consumer metric is decomposed into producer-package imports
(0/1 PASS), raw-source reconstruction (1/1 PASS), separate executable
(1/1 PASS), and algorithmic independence (0/1 UNKNOWN). Separate processes do
not by themselves prove algorithmic diversity.

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
exact diagnostic JSON bytes and digest, raw input bytes and digest, all inline in
the uploaded evidence artifact. Repository tracked/untracked path-plus-content snapshots are read
before and after the complete proof. Net equality is reported as
`NET_STATE_EQUAL`; transient mutation and mutation authority remain `UNKNOWN`
unless separately observed.

`FOUNDATION`, `COHERENCE`, and `REGRESSION` are selected from every local
manifest and each strategy gates its own evidence. CI uses Go 1.27.0 and owns
format, compile, generation, freshness, and semantic verification. Local test
execution is intentionally zero.
