# Manual-source-registration-edit-free registry projection

This experiment keeps concept registration local and makes the shared view a
deterministic projection. The bounded slice has exactly three production
inputs:

- `examples/language-syntax-roundtrip/concept.manifest.json`
- `examples/language-semantic-model/concept.manifest.json`
- `examples/toolchain-conformance/concept.manifest.json`

Each manifest owns its concept-local code and metric bindings, use cases, raw
resource references, and declared denominators. The producer discovers and
sorts those inputs, verifies ownership paths, non-empty bindings, real code and
metric registrations, resource digests, and source-derived denominators, then
generates catalog, corpus, registry, denominator, documentation, manifest index,
projection, and digest outputs. The root topology and root README remain
explicit exceptions; this is a bounded vertical slice rather than a migration
of every existing global file.

## Fixed regression denominator

The baseline has exactly 12 observed manual shared-source touchpoints. A real
temporary filesystem intervention adds one fourth local manifest and runs
discovery plus generation. The existing 12 source paths are compared by bytes
digest before and after. The same intervention measures changed generated
outputs among exactly 8 outputs, and runs the independent production consumer
against the generated projection.

The three surfaces are intentionally distinct:

| Surface | Measured result | Meaning |
| --- | ---: | --- |
| existing shared source touchpoints | 12/12 baseline | current human-edited registration surface |
| generator-changed shared outputs | 6/8 for the new fixture | generated projection change surface, not zero conflict |
| production consumer adoption | 1/1 | independent raw-manifest reconstruction equals projection |
| manual source registration edits required | 0/12 after fixture | no existing source file was edited |

The corrected toolchain denominator reconciles `181` corpus cases and the
`152` use-case string count. A regression clone with the prior declared `160`
cases must produce `FAIL_CLOSED / DENOMINATOR_SOURCE_MISMATCH`, with declared
and calculated values side by side in the receipt.

## Meaning gates

The proof keeps raw manifest bytes separate from the semantic manifest view.
Semantic and comment-only interventions render actual changed JSON bytes,
compute the digest from those bytes, decode them again, and only then project.
Replay twice and reversed manifest order must be byte-identical. The independent
consumer independently checks malformed, missing, stale, cross-directory,
missing-binding, duplicate-ID, and stale-denominator inputs fail closed while
preserving stage, step, and reason.

Claims carry distinct observed predicates, target addresses, target digests, and
an `OPEN` transition. Only the independent consumer's recomputation may move a
positive predicate to `DISCHARGED`; rejection claims remain `REFUTED` with the
failure diagnostic. Repository status is read before and after the complete
proof. Net equality is reported as `NET_STATE_EQUAL`; transient mutation and
mutation authority remain `UNKNOWN` unless separately observed.

`FOUNDATION`, `COHERENCE`, and `REGRESSION` are selected from every local
manifest and each strategy gates its own evidence. CI uses Go 1.27.0 and owns
format, compile, generation, freshness, and semantic verification. Local test
execution is intentionally zero.
