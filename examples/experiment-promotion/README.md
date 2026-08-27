# Experiment promotion ledger

This is a separate measurement-governance experiment from denominator-evolution v1. The v1 fixed denominator remains `5/5`; this portfolio uses the independent fixed denominators `declared_experiments=30/30` and `materialized_claim_slots=150/150`. The denominators are not summed, scored, weighted, or converted into an improvement rate.

## Source authority

[`main.gooo`](./main.gooo) declares the fixed identity tuple `experiment_id + exact PR number + topic/claim address` for the actual portfolio: #549, #548, #545, #555, #544, #543, #546, #559, #550, #552, #564, #547, #558, #542, #569, #567, #551, #560, #553, #562, #563, #566, #554, #570, #541, #556, #561, #568, #565, and #557. It also declares the five gate IDs. It declares no decision, resolution, expected reason, claim state, or final PASS.

The producer and independent consumer each consume raw `.gooo` bytes through `syntax.ParseFile -> bidir.Lower`, then reconstruct the identity records and semantic digest in different procedures and data structures. The JSON contract is a validator expectation only. A duplicate PR, cross-gate head, reused observation/run/job/artifact/target relation, or contract/source mismatch fails closed.

The five predicates are distinct. `source-bound` requires the actual candidate `.gooo` attachment bytes. `semantic-causality` lowers baseline, semantic, and comment-only source bytes and checks the derived semantic and claim digests. `independent-consumer` binds captured consumer procedure bytes, source path, algorithm ID, and import graph. `persistent-claim-transition` consumes a hashed append-only prior ledger. `exact-actions` reconstructs raw captured API bytes and binds repository, PR, exact head, workflow/job IDs, conclusion, and artifact identity.

## Evidence and states

An observation receipt contains actual raw bytes, not only metadata: procedure bytes, Actions API bytes, artifact bytes, and their lengths/digests. A metadata-only artifact reseal fails. Actions URLs or conclusions copied into JSON without the corresponding captured API bytes do not prove anything. `CURRENT_EVIDENCE`, `HISTORICAL_FIXTURE`, and `UNKNOWN` are separate classes; fixture receipts never promote a current portfolio slot.

The report exposes actual promotion states and fixture replay states separately:

```text
declared_experiments=30/30
materialized_claim_slots=150/150
experiment_states PROVEN n/30 OPEN n/30 UNKNOWN n/30 REFUTED n/30
gate_states PROVEN n/150 OPEN n/150 UNKNOWN n/150 REFUTED n/150
```

Every gate retains an explicit `OPEN -> DISCHARGED`, `OPEN -> OPEN`, or `OPEN -> REFUTED` evidence transition. A missing receipt is `UNKNOWN` with `stage`, `step`, and `reason`; an observation ID is retained even when its fields are incomplete. A historical fixture's valid or refuted evidence appears in the fixture-only counts, while its promotion state remains `UNKNOWN`.

The four checked-in fixtures are a deterministic corpus, not current portfolio progress: `fully-proven.json` has one complete historical experiment, `missing-common-gate.json` has a missing gate and one in-progress receipt, `malformed-evidence.json` contains a digest mismatch, and `contradictory-semantic.json` has a raw-changing but semantically unchanged intervention. Their results are not added to actual evidence counts.

The fixed counterexample denominator is `9/9` slots: duplicate PR mapping, cross-gate head mismatch, fake import list, metadata-only artifact reseal, random semantic digests, stale ledger deletion/reset, fixture claiming current evidence, reused run/job/artifact relation, and forbidden aggregate injection. Each rejection records its stage, step, and reason.

## Guardrails and research basis

Forbidden aggregate claims are counted from emitted claim bytes and shown directionally as `observed`, `allowed_max`, and `conformance`; zero is not a hard-coded success. Repository writes are counted from separate before/after snapshot bytes and changed paths. A no-write observation does not prove broader mutation authority. `aggregate_metrics` is an empty array.

The design adopts [Apache Avro schema resolution](https://avro.apache.org/docs/++version++/specification/#schema-resolution): writer-side bytes and reader-side resolution are explicit, and incompatible resolution fails. It adopts [JSON Schema's dialect/version boundary](https://json-schema.org/understanding-json-schema/reference/schema) for the versioned contract, while rejecting the contract as runtime authority. It adopts [NIST metrological traceability](https://www.nist.gov/metrology/metrological-traceability) for the documented predecessor/source/artifact/claim digest chain, while rejecting weighted scores and aggregate improvement estimates.

## Falsifiers

The experiment is falsified if a comment-only edit changes semantic meaning, a source identity is accepted from contract metadata instead of raw `.gooo`, a fake import list passes without matching captured procedure bytes, an artifact digest passes after its bytes change, a fixture promotes current evidence, a prior ledger is deleted/reset, the same run/job/artifact relation proves multiple gates, a forbidden aggregate claim is not counted, or any aggregate rate/score is emitted.

The workflow checks Go 1.27, exact pull-request head checkout, raw-source and observation replay, consumer producer-import `0/0`, the four fixture outcomes, exact denominators, persistent claims `150/150`, and the actual repository write-set. It never merges the pull request.
