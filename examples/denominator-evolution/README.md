# Denominator evolution: a self-measuring Gooo experiment

This experiment defines when a measurement denominator may legally move. It is
not a generic version-control clone: the object under test is the measurement
basis itself. A successor is accepted only when the exact predecessor digest,
successor digest, every addition/deletion reason, and a replayable migration
receipt agree.

## Exact contract

- Fixed denominator: `gooo/measurement-denominator/v1`, exactly `5/5` declared obligations.
- Legal advance: `1/1` case, `v1 -> v2`, one addition with `NEW_MEASURABLE_OBLIGATION`, and one deletion with `DEPRECATED_DUPLICATE`.
- Unauthorized change: `1/1` case rejected as `BLOCK / INVARIANT_ONLY / MIGRATION_RECEIPT_MISSING`, with `OPEN -> REFUTED`.
- Unknown predecessor: `1/1` case is `FAIL_CLOSED / LOWER_RESOLUTION / PREDECESSOR_DIGEST_UNKNOWN`, with `OPEN -> OPEN` preserved in the ledger.
- Guardrail `gooo.guardrail.denominator.forbidden-estimate.v1`: direction `AT_MOST`, `observed=0`, `allowed_max=0`, `conformance=1/1`, `conforms=true`.
- Guardrail `gooo.guardrail.denominator.repository-writes.v1`: direction `AT_MOST`, `observed=0`, `allowed_max=0`, `conformance=1/1`, `conforms=true`.

Additional exact CI predicates are `source_cases=3/3`, `producer_imports=0/0`,
`semantic_causality=1/1`, `nonsemantic_preservation=1/1`,
`persistent_claims=3/3`, and `guardrail_observations=2/2`. There is no
improvement rate, weighted score, percentage, or composite estimate.

The guardrail fields deliberately separate direction, observed value, allowed
boundary, and pass evidence. A zero is not treated as a pass by implication:
`observed=0` is compared with `allowed_max=0`, and only the explicit
`conformance=1/1` plus `conforms=true` records that the guardrail passed. The
same observations are carried by the legal migration receipt, producer
summary, CI step summary, and independent consumer output.

## Gooo and evidence flow

`main.gooo` declares all five obligations, all three case inputs, add/delete
reasons, predecessor versions, receipt fields, and prior claim states in
`computes` values. The producer `denominatorevolution.Evaluate` runs
`syntax.ParseFile -> bidir.Lower`, builds its own wire model, derives the
denominator/cases/ledger, and treats `contract.json` only as a verification
expectation. The consumer `denominatorevolutionverify.Verify` is a separate
package with no import edge to the producer; it receives raw source bytes,
performs the same parser/lowerer boundary with an independent wire model and
decision function, and rejects a coherently resealed report that disagrees with
the source.

Each claim has explicit `stage`, `step`, and `reason` in a persistent sealed
ledger. The transitions are `OPEN -> DISCHARGED`, `OPEN -> REFUTED`, and
`OPEN -> OPEN`. Unknown predecessor evidence is fail-closed with
`LOWER_RESOLUTION`; it does not invent a terminal claim state. The
`meta_operation` and `proof_choice` (`FOUNDATION`, `COHERENCE`, or `REGRESSION`)
remain attached to each case check.

The forbidden-estimate guardrail counts actual `ASSERTED` entries of class
`FORBIDDEN_ESTIMATE` in the structured emitted-claim ledger. The repository
write guardrail is bound to before/after repository snapshots captured by the
CI wrapper. Neither zero is a producer constant.

## Adopted and rejected principles

The experiment adopts these official principles:

1. [Apache Avro specification](https://avro.apache.org/docs/1.11.3/specification/): schema resolution compares writer and reader schemas, so a version label alone is insufficient. We adopt explicit version-plus-digest resolution and reject name-only advancement.
2. [W3C PROV-DM](https://www.w3.org/TR/prov-dm/): a derivation can be an update producing a new entity and provenance records the entities and activities involved. We adopt predecessor/successor entities, an issuing activity, and a replayable receipt; we reject an untraceable “latest wins” state.
3. [NIST metrological traceability policy](https://www.nist.gov/metrology/metrological-traceability): a result needs a documented unbroken chain and the provider must support the claim. We adopt a receipt that binds every link and reject producer self-attestation without independent replay.
4. [Confluent schema evolution guidance](https://docs.confluent.io/platform/current/schema-registry/fundamentals/schema-evolution.html): compatibility is checked against prior versions and transitive modes exist. We adopt explicit prior-version checks, but reject treating compatibility as permission to alter this denominator; permission additionally requires change reasons and the migration receipt.

The sources guide the governance shape; they do not certify this repository's
measurement semantics. `main.gooo` is the source authority for this experiment;
the JSON contract is a checked expectation and the independent verifier is a
second judgment.

## Falsifiers and boundary

The claim is falsified by any one of: a report whose denominator digest does not
recompute, a receipt whose predecessor or successor digest differs, a missing or
inadmissible add/delete reason, an unknown predecessor becoming `DISCHARGED` or
`REFUTED`, a producer import in the consumer, a changed report across
deterministic replay, a semantic source mutation that leaves the decision
unchanged, a comment-only mutation that changes the semantic result, or any
repository write. The experiment does not claim that the five obligations
measure language quality or that a legal denominator advance is an improvement
in a product metric.
