# Meta-operation provenance experiment

This is an executable Gooo experiment, not a general-purpose lineage
library. Its subject is the Gooo-specific relation between a metric,
producer, consumer, meta-operation, and evidence path. The source stores the
experiment contract as actual `computes` values. Both the producer and the
independent consumer parse and lower the raw `.gooo` file before recovering
those values; declaration names are not the data contract.

## Fixed contract

There are three semantic metric records, one each in `FOUNDATION`, `COHERENCE`,
and `REGRESSION`, and four semantic scenario-mutation records. Each metric has
four required relation slots, so every scenario has a fixed denominator of
`3 * 4 = 12`. The numerator counts uniquely resolved `PRODUCES`, `CONSUMES`,
`OPERATES`, and `EVIDENCED_BY` relations. An absent consumer is a direct
`FAIL_CLOSED`; other missing lineage is `UNKNOWN`. A dependency on an unknown
upstream metric is separately marked `DEPENDENCY_BLOCK`.

The continuous claim ledger starts every source claim at `OPEN`:

| Conformance decision | Next claim | Ledger transition |
| --- | --- | --- |
| `PASS` | `DISCHARGED` | `DISCHARGED` |
| `UNKNOWN` | `OPEN` | `PRESERVED_OPEN` |
| `FAIL_CLOSED` | `REFUTED` | `REFUTED` |

Conformance decision and subject resolution are separate fields. A parsed
source has `EXACT` resolution, while a consumer run without raw source returns
`UNKNOWN` with `LOWER_RESOLUTION`, stage `CONSUMER`, step `parse-source`, and
reason `REQUIRED_RAW_SOURCE_MISSING`.

## Evidence and interventions

The receipt records both the raw source digest and the canonical semantic
digest from `semantic.IR.StableHash()`. Repository writes and mutation
authority are derived from isolated CI before/after repository-status
observations; the receipt does not assert them as constants.

The CI artifact contains two counterfactual interventions:

- The semantic intervention changes the `consumer` value in one `computes`
  record. It must change the canonical semantic digest and at least one
  decision/claim transition.
- The nonsemantic intervention appends a comment. It must change only the raw
  source digest while preserving the semantic digest, decisions, and
  transitions.

To reproduce the artifact path with Go 1.27:

```sh
go run ./scripts/meta-operation-provenance \
  -mode build -source examples/meta-operation-provenance/main.gooo \
  -workspace-root "$PWD" -out /tmp/meta-operation-provenance/receipt.json
go run ./scripts/meta-operation-provenance \
  -mode verify -source examples/meta-operation-provenance/main.gooo \
  -receipt /tmp/meta-operation-provenance/receipt.json \
  -consumer-source internal/meta/operationprovenance/verify/core.go \
  -out /tmp/meta-operation-provenance/verification.json
```

The checked-in [`receipt.json`](receipt.json) is the producer result. The
consumer reconstructs the expected graph and ledger from the lowered source,
without importing the producer or copying a canonical source string or
expected scenario table.

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
