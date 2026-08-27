# Denominator evolution: a self-measuring Gooo experiment

This experiment defines when a measurement denominator may legally move. It is
not a generic version-control clone: the object under test is the measurement
basis itself. A successor is accepted only when the exact predecessor digest,
successor digest, every addition/deletion reason, and a replayable receipt
agree.

## Exact contract

- Fixed denominator: `gooo/measurement-denominator/v1`, exactly `5/5` declared members.
- Legal proposal: computed `ADVANCE / EXACT`, with one addition (`NEW_MEASURABLE_OBLIGATION`) and one deletion (`DEPRECATED_DUPLICATE`), and `OPEN -> DISCHARGED`.
- Unreceipted proposal: computed `BLOCK / INVARIANT_ONLY`, with `OPEN -> REFUTED` for the calculated `migration-authorized` claim.
- Unknown predecessor: computed `FAIL_CLOSED / LOWER_RESOLUTION / PREDECESSOR_DIGEST_UNKNOWN`, with `OPEN -> OPEN` preserved.
- The computed successor is `gooo/measurement-denominator/v2`; its record has five members and a predecessor reference to the v1 digest.
- Forbidden-estimate guardrail: `direction=AT_MOST`, `proposition_present=true`, `observed=0`, `allowed_max=0`, `conformance=1/1`, `conforms=true`.
- Repository-write guardrail: `direction=AT_MOST`, `proposition_present=false`, `observed=0`, `allowed_max=0`, `conformance=1/1`, `conforms=true`.

CI exposes exact predicates: `source_cases=3/3`, `producer_imports=0/0`,
`semantic_causality=1/1`, `nonsemantic_preservation=1/1`,
`persistent_claims=3/3`, `guardrail_observations=2/2`,
`version_records=2/2`, and `v1_nonretroactive=1/1`. There is no improvement
rate, weighted score, percentage, or composite estimate.

## Source and calculation boundary

`main.gooo` contains only the denominator version, exact member definitions,
three proposal inputs, predecessor references, proposed additions/deletions
with reasons, and optional receipt binding material. It contains no case
outcomes, expected reasons, claim states, emitted statuses, or independent
result labels.

The producer and independent consumer each run
`syntax.ParseFile -> bidir.Lower`, decode their own wire model, calculate the
v1 digest, apply the proposed change set, calculate the v2 digest, resolve
predecessor `(version,digest)` pairs against a registry, enforce duplicate /
missing-member checks and reason allowlists, and validate receipt binding.
`contract.json` is a checked expectation only; it cannot create source inputs
or authorize a transition. The consumer does not import the producer package.

The legal receipt carries the computed predecessor and successor references,
change set, and receipt digest. The v1 and v2 results are separate immutable
records. The v2 record chains to v1, while the v1 record remains the exact
five-member result that existed before activation.

Claim ledger entries are generated from the calculated result. The ledger is
sealed with previous digests and records stage, step, and reason. The emitted
claim list is also calculated; the forbidden-estimate guardrail counts actual
`ASSERTED` entries of class `FORBIDDEN_ESTIMATE`, not source prose. The artifact
schema carries an empty `aggregate_metrics` list and rejects any aggregate
metric artifact.

## Adopted and rejected principles

1. [Apache Avro specification](https://avro.apache.org/docs/1.11.3/specification/): schema resolution compares writer and reader schemas. We adopt explicit version-plus-digest resolution and reject name-only advancement.
2. [W3C PROV-DM](https://www.w3.org/TR/prov-dm/): derivation relates entities and activities. We adopt predecessor/successor entities and a replayable receipt; we reject an untraceable “latest wins” state.
3. [NIST metrological traceability policy](https://www.nist.gov/metrology/metrological-traceability): a result needs a documented unbroken chain. We adopt receipt-bound links and independent replay; we reject producer self-attestation.
4. [Confluent schema evolution guidance](https://docs.confluent.io/platform/current/schema-registry/fundamentals/schema-evolution.html): compatibility is checked against prior versions. We adopt prior-version checks but reject treating compatibility as permission; reasons and receipt binding are additionally required.

These sources guide the governance shape; they do not certify this
repository's measurement semantics.

## Falsifiers and boundary

The claim is falsified by a digest that does not recompute, a receipt whose
binding differs from calculated records, a missing or inadmissible reason, a
duplicate or missing member, an unknown predecessor becoming terminal, a
producer import in the consumer, a coherent reseal accepted despite source
disagreement, a semantic mutation that leaves the result unchanged, a
comment-only mutation that changes semantic output, an aggregate metric
artifact, or any repository write. The experiment does not claim that five
members measure language quality or that a legal denominator advance is a
product improvement.
