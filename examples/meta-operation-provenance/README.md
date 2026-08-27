# Meta-operation provenance experiment

This is an executable Gooo experiment, not a general-purpose lineage
library. Its subject is the Gooo-specific relation between a metric,
producer, consumer, meta-operation, and evidence path. The source stores the
experiment contract as actual `computes` values. Both the producer and the
independent consumer parse and lower the raw `.gooo` file before recovering
those values; declaration names are not the data contract.

## Fixed contract

There are three semantic metric records, one each in `FOUNDATION`, `COHERENCE`,
and `REGRESSION`, and four semantic scenario-mutation records. The producer and
consumer enforce those cardinalities and `OPEN` prior claims. Each metric has
four required relation slots, so every scenario has a fixed denominator of
`3 * 4 = 12`. A relation enters the numerator only after its own raw artifact
is read, parsed, digest-recorded, and shown to agree with the declared endpoint.
The four artifacts are a producer receipt, consumer reconstruction receipt,
executed meta-operation receipt, and evidence artifact. A missing or mismatched
artifact is `UNKNOWN`; the disconnected scenario is a verified counterexample
to the explicit lineage-completeness proposition and is `FAIL_CLOSED`.

The continuous claim ledger starts every source claim at `OPEN`:

| Conformance decision | Next claim | Ledger transition |
| --- | --- | --- |
| `PASS` | `DISCHARGED` | `DISCHARGED` |
| `UNKNOWN` | `OPEN` | `PRESERVED_OPEN` |
| verified `FAIL_CLOSED` | `REFUTED` | `REFUTED` |
| unverified failure | `OPEN` | `PRESERVED_OPEN` |

Conformance decision is separate from source and lineage resolution. A parsed
source has `source_resolution=EXACT`; a metric with missing lineage evidence has
`lineage_resolution=LOWER_RESOLUTION` while its claim remains `OPEN`. A
consumer run without raw source returns `UNKNOWN` with both resolutions lowered,
stage `CONSUMER`, step `parse-source`, and reason
`REQUIRED_RAW_SOURCE_MISSING`. Every ledger row explicitly states the
proposition: `metric M has complete executable lineage for this subject digest
D`; underlying metric-value claims are not refuted by lineage absence.

## Evidence and interventions

The receipt records both the raw source digest and the canonical semantic
digest from `semantic.IR.StableHash()`. Repository writes are derived from
isolated CI before/after repository-status observations. Mutation authority is
a separate capability-envelope observation and is `UNKNOWN` when CI supplies
no authoritative envelope; it is never inferred from the absence of writes.

The CI artifact contains two counterfactual interventions:

- The semantic intervention changes the producer endpoint value in one
  `computes` record. It must change the canonical semantic digest and at least
  one decision/claim transition.
- The nonsemantic intervention appends a comment. It must change only the raw
  source digest while preserving the semantic digest, decisions, and
  transitions.
- The no-op intervention attempts to remove a nonexistent relation and must
  be rejected with `SCENARIO/apply-mutation/NOOP_MUTATION_REJECTED`.

To reproduce the artifact path with Go 1.27:

```sh
go run ./scripts/meta-operation-provenance \
  -mode build -source examples/meta-operation-provenance/main.gooo \
  -artifact-root examples/meta-operation-provenance/artifacts \
  -workspace-root "$PWD" -out /tmp/meta-operation-provenance/receipt.json
go run ./scripts/meta-operation-provenance \
  -mode verify -source examples/meta-operation-provenance/main.gooo \
  -receipt /tmp/meta-operation-provenance/receipt.json \
  -artifact-root examples/meta-operation-provenance/artifacts \
  -consumer-source internal/meta/operationprovenance/verify \
  -out /tmp/meta-operation-provenance/verification.json
```

The checked-in [`receipt.json`](receipt.json) is the producer result. The
consumer reads the lowered source and all four raw artifact classes in its own
package, without importing the producer or copying a canonical source string.
CI proves producer independence from a digest of every `.go` file in that
package (not one selected file).

## Research and design choices

The experiment adopts the useful separation of entities, activities, and
derivation from the [W3C PROV-O Recommendation](https://www.w3.org/TR/prov-o/),
but rejects RDF/OWL as the execution surface: Gooo's typed entity-to-activity
chain is the subject under experiment. It also adopts explicit producer
identity and additive event/evidence orientation from the [OpenLineage
specification](https://github.com/OpenLineage/OpenLineage/blob/main/spec/OpenLineage.md),
but rejects job/run/dataset terminology as the core model because this PR
needs metric-to-meta-operation relations and fail-closed decision semantics.

Those sources motivate vocabulary only. Neither source is treated as proof
that an unbound metric is meaningful. The narrower claim tested here is that a
fixed-ratio metric is actionable only when its four Gooo-specific lineage
relations resolve, and that the independent consumer preserves direct,
dependency-blocked, and lower-resolution uncertainty.
