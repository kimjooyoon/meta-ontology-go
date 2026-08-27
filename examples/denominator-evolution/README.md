# Denominator evolution: a self-measuring Gooo experiment

This experiment defines when a measurement denominator may legally move. It is
not a generic version-control clone: the object under test is the measurement
basis itself. A successor is accepted only when the exact predecessor digest,
successor digest, every addition/deletion reason, and a replayable migration
receipt agree.

## Exact contract

- Fixed denominator: `gooo/measurement-denominator/v1`, exactly `5/5` declared obligations.
- Legal advance: `1/1` case, `v1 -> v2`, one addition with `NEW_MEASURABLE_OBLIGATION`, and one deletion with `DEPRECATED_DUPLICATE`.
- Unauthorized change: `1/1` case rejected as `BLOCK / INVARIANT_ONLY / MIGRATION_RECEIPT_MISSING`.
- Unknown predecessor: `1/1` case remains `FAIL_CLOSED / UNKNOWN / PREDECESSOR_DIGEST_UNKNOWN`.
- Guardrail `gooo.guardrail.denominator.forbidden-estimate.v1`: direction `AT_MOST`, `observed=0`, `allowed_max=0`, `conformance=1/1`, `conforms=true`.
- Guardrail `gooo.guardrail.denominator.repository-writes.v1`: direction `AT_MOST`, `observed=0`, `allowed_max=0`, `conformance=1/1`, `conforms=true`.

These are separate exact predicates. There is no improvement rate, weighted
score, percentage, or composite estimate.

The guardrail fields deliberately separate direction, observed value, allowed
boundary, and pass evidence. A zero is not treated as a pass by implication:
`observed=0` is compared with `allowed_max=0`, and only the explicit
`conformance=1/1` plus `conforms=true` records that the guardrail passed. The
same two guardrails are carried by the legal migration receipt, the producer
summary, the CI step summary, and the independent consumer output.

## Gooo and evidence flow

`examples/denominator-evolution/main.gooo` declares the basis, the change
reasons, predecessor evidence, receipt, claim transition, and independent
decision. The producer `denominatorevolution.Evaluate` projects that source and
emits a report containing the three cases and the migration receipt. The
consumer `denominatorevolutionverify.Verify` is a separate package and has no
import edge to the producer; it decodes the report, recomputes the digests, and
replays the legal case before accepting the report.

Each claim has an explicit `stage`, `step`, `reason`, `meta_operation`, and
`proof_choice` (`FOUNDATION`, `COHERENCE`, or `REGRESSION`). The transitions are
`PROPOSED -> ACCEPTED`, `PROPOSED -> REJECTED`, and `PROPOSED -> UNKNOWN`.

## Adopted and rejected principles

The experiment adopts these official principles:

1. [Apache Avro specification](https://avro.apache.org/docs/1.11.3/specification/): schema resolution compares writer and reader schemas, so a version label alone is insufficient. We adopt explicit version-plus-digest resolution and reject name-only advancement.
2. [W3C PROV-DM](https://www.w3.org/TR/prov-dm/): a derivation can be an update producing a new entity and provenance records the entities and activities involved. We adopt predecessor/successor entities, an issuing activity, and a replayable receipt; we reject an untraceable “latest wins” state.
3. [NIST metrological traceability policy](https://www.nist.gov/metrology/metrological-traceability): a result needs a documented unbroken chain and the provider must support the claim. We adopt a receipt that binds every link and reject producer self-attestation without independent replay.
4. [Confluent schema evolution guidance](https://docs.confluent.io/platform/current/schema-registry/fundamentals/schema-evolution.html): compatibility is checked against prior versions and transitive modes exist. We adopt explicit prior-version checks, but reject treating compatibility as permission to alter this denominator; permission additionally requires change reasons and the migration receipt.

The sources guide the governance shape; they do not certify this repository's
measurement semantics. The exact contract and independent verifier remain the
authority for this experiment.

## Falsifiers and boundary

The claim is falsified by any one of: a report whose denominator digest does not
recompute, a receipt whose predecessor or successor digest differs, a missing or
inadmissible add/delete reason, an unknown predecessor becoming `ACCEPTED`, a
producer import in the consumer, a changed report across deterministic replay,
or any repository write. The experiment does not claim that the five
obligations measure language quality or that a legal denominator advance is an
improvement in a product metric.
