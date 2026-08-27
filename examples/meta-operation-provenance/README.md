# Meta-operation provenance experiment

This is an executable Gooo experiment, not a general-purpose lineage library.
The source declares a Gooo-specific relation chain: a metric is progressively
bound to a producer, consumer, meta-operation, and evidence path. The Go
evaluator treats that chain as a graph and keeps the metric's claim state
(`OPEN`, `DISCHARGED`, or `REFUTED`) unchanged while judging the graph.

## Fixed contract

The registry contains exactly three indicators: one each in `FOUNDATION`,
`COHERENCE`, and `REGRESSION`. Each has a fixed denominator of four required
lineage relations, so every scenario has a fixed total denominator of 12.
The numerator is the count of uniquely resolved relations; duplicate or absent
edges do not count. A metric with no consumer is `FAIL_CLOSED`, because its
ratio has no meaning. A missing evidence path is `UNKNOWN` with a direct cause.
An otherwise complete metric whose upstream metric is `UNKNOWN` is `UNKNOWN`
with an explicit dependency block.

## Evidence

Run the experiment through the repository's Go 1.27 toolchain:

```sh
go run ./scripts/meta-operation-provenance \
  -mode build -source examples/meta-operation-provenance/main.gooo \
  -out /tmp/meta-operation-provenance/receipt.json
go run ./scripts/meta-operation-provenance \
  -mode verify -source examples/meta-operation-provenance/main.gooo \
  -receipt /tmp/meta-operation-provenance/receipt.json \
  -out /tmp/meta-operation-provenance/verification.json
```

The checked-in [`receipt.json`](receipt.json) records four cases:

| Case | Fixed numerator/denominator | Decisions | Meaning |
| --- | ---: | --- | --- |
| `complete` | 12/12 | 3 PASS | all four relations resolve for all metrics |
| `disconnected` | 11/12 | 2 PASS, 1 FAIL_CLOSED | consumer edge is absent |
| `direct-unknown` | 11/12 | 2 PASS, 1 UNKNOWN | evidence path is directly missing |
| `dependency-blocked` | 11/12 | 1 PASS, 2 UNKNOWN | regression is blocked by foundation UNKNOWN |

`verification` is produced by a separate package that reconstructs the source,
fixed metric contract, expected graph cardinalities, decision cases, and digest;
it does not accept the evaluator's decision as authority. The receipt defaults
to `repository_workspace_writes=false` and `mutation_authority=false`.

## Research and design choices

The experiment adopts the useful separation of entities, activities, and
derivation from the [W3C PROV-O Recommendation](https://www.w3.org/TR/prov-o/),
but rejects RDF/OWL as the execution surface: Gooo's typed entity-to-activity
chain is the subject under experiment. It also adopts OpenLineage's explicit
producer identity and additive event/evidence orientation from its
[official specification](https://github.com/OpenLineage/OpenLineage/blob/main/spec/OpenLineage.md),
but rejects job/run/dataset terminology as the core model because this PR
needs metric-to-meta-operation relations and fail-closed decision semantics.

Those sources motivate the vocabulary; neither source is treated as proof that
an unbound metric is meaningful. The experiment's novel claim is narrower:
fixed-ratio metrics become executable only when all four Gooo-specific lineage
edges resolve, and the independent judge preserves the distinction between a
direct missing input and a dependency-blocked result.
