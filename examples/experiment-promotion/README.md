# Experiment promotion ledger

This is a separate measurement-governance experiment from denominator-evolution v1. The existing fixed denominator remains `5/5`; this portfolio has its own fixed denominators: `experiments=30/30` and `gate_slots=150/150`. They are never added or converted into a score.

## Authority boundary

`main.gooo` declares exactly the source vocabulary: `experiment-01` through `experiment-30`, and the five gates `source-bound`, `semantic-causality`, `independent-consumer`, `persistent-claim-transition`, and `exact-actions`. It declares no outcome, decision, resolution, expected reason, claim state, or final PASS.

The producer and the independent consumer each perform `syntax.ParseFile -> bidir.Lower` on the raw source and decode the two value programs themselves. The JSON contract is a validator expectation, not the authority for the portfolio. Observation input is also separate from deterministic replay so a later common observer (for example, #550) can supply raw GitHub observations without turning a cached network conclusion, PR title, or PR body into evidence.

Each external receipt must bind:

* PR number and exact 40-character head SHA;
* raw-source and semantic-source digests;
* producer ID and consumer package/import boundary;
* claim-transition digest;
* exact Actions run URL, job URL, and conclusion;
* artifact byte count, path, and digest.

An absent receipt is `UNKNOWN`, never a pass. A receipt with malformed or contradictory evidence is `REFUTED`. `OPEN` is reserved for a valid observation whose Actions conclusion is still `in_progress` or `queued`. The persistent ledger retains one `OPEN -> DISCHARGED`, `OPEN -> OPEN`, or `OPEN -> REFUTED` transition for every one of the 150 slots.

## Fixture corpus

The fixtures are a deterministic corpus, not a claim about the current GitHub portfolio:

* `fully-proven.json`: one experiment has all five valid receipts, so it is `PROVEN` and the remaining 29 experiments are `UNKNOWN`;
* `missing-common-gate.json`: a common `exact-actions` receipt is absent, and one separate slot is still `OPEN`;
* `malformed-evidence.json`: an artifact digest is invalid and is calculated as `REFUTED`;
* `contradictory-semantic.json`: the semantic intervention changes raw bytes but not semantic meaning and is calculated as `REFUTED`.

No receipt is bound for the other 29 experiments. Therefore the fixtures do not assert that the current portfolio has generated `30/30`, passed a first audit `30/30`, has ten semantic candidates, has six dedicated Actions successes, or has any overall promotion. Those historical-looking numbers can appear only as separately bound observations in a later observer input.

## Guardrails and exact output

The report contains no aggregate metric. Guardrails expose direction, observation, allowed value, and conformance separately:

```text
experiments=30/30 gate_slots=150/150
guardrail_forbidden_aggregate_claim observed=0 allowed_max=0 conformance=1/1
guardrail_repository_writes observed=0 allowed_max=0 conformance=1/1
```

The first observation is computed by scanning the emitted claim ledger for `ASSERTED` claims in forbidden aggregate classes. The second is computed from the CI wrapper's before/after repository snapshots. Neither zero is a hard-coded success label. `aggregate_metrics` must be an empty array and repository mutation authority must be false.

## Design research and adoption rule

The design adopts the Apache Avro specification's schema-resolution discipline: a reader must have the writer schema and incompatible schema differences are errors. Here the source projection and receipt artifact are explicit writer-side inputs, and the consumer resolves them independently by digest rather than trusting a report. See [Apache Avro schema resolution](https://avro.apache.org/docs/++version++/specification/#schema-resolution).

It also adopts JSON Schema's versioned dialect boundary: a schema identifier belongs at the document root and tells tooling which vocabulary applies. The contract therefore has an explicit schema/version and is validated before use. We reject the idea that a schema or contract should be the runtime authority: the `.gooo` source and raw observation receipts are independently reconstructed first. See [JSON Schema dialect and `$schema`](https://json-schema.org/understanding-json-schema/reference/schema).

Finally, the ledger adopts NIST's measurement-traceability principle of a documented, unbroken chain to a reference. Predecessor/source/artifact/claim digests form that chain, with missing links failing closed into `UNKNOWN`. We reject replacing that chain with an improvement rate, weighted score, or uncertainty-free aggregate estimate; this experiment reports only exact numerators, denominators, and state counts. See [NIST metrological traceability](https://www.nist.gov/metrology/metrological-traceability).

## Falsifiers

The claim is falsified if a comment-only source edit changes the semantic digest, if a source ID or gate is accepted from the contract but not reconstructed from raw `.gooo`, if the consumer imports the producer package, if a missing receipt becomes `PROVEN`, if a coherent reseal with a changed source digest is accepted, if an intervention changes raw bytes without changing semantic digest but remains proven, if a claim transition is deleted, or if any aggregate score/rate is emitted.

The CI workflow checks Go 1.27, exact pull-request head checkout, producer/consumer replay, forbidden imports `0/0`, all four fixture states, the fixed denominators, persistent claims `150/150`, and the actual repository write-set. It does not merge the pull request.
