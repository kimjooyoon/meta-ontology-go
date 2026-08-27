# Partial-knowledge composition

This read-only experiment treats the five `computes` programs in
[`main.gooo`](main.gooo) as source recipes. The source declares required
observations, dependency recipes, invariant capabilities, producer/consumer,
meta-operation, and proof choice. It does not declare `observed`,
`observed_available`, `invariant_evidence`, or any conclusion state.

The CI-only observer parses and lowers the source, then emits a separate raw
evidence receipt from the source digest and tracked, untracked, and status
snapshots taken before and after observation. The producer and independent
consumer each parse/lower the source and independently validate that receipt;
the consumer does not import the producer package. JSON is an observation
artifact, not the authority for which cases exist.

The calculus derives these outcomes from raw evidence:

| Case | Resolution | Decision | Claim transition |
|---|---|---|---|
| exact-pair | `EXACT` | `PASS` | `OPEN -> DISCHARGED` |
| direct-unknown | `LOWER_RESOLUTION` | `UNKNOWN` | `OPEN -> OPEN` |
| dependency-blocked | `LOWER_RESOLUTION` | `UNKNOWN` | `OPEN -> OPEN` |
| invariant-preservation | `INVARIANT_ONLY` | `HOLD` | `OPEN -> OPEN` |
| mixed-unknown-and-blocked | `LOWER_RESOLUTION` | `UNKNOWN` | `OPEN -> OPEN` |

The receipt-level decision is `CALCULUS_PROVEN` with `resolution=CALCULUS`.
It proves only that the composition rules were replayed. Subject resolution,
evidence coverage, and promotion authority remain separate fields, so the
four non-exact cases are not promoted. The source-derived fixed denominator is
five distinct predicates, with one exact case, four open claims, and
`promotion_authorized=false` because permission evidence is unavailable.

Each claim carries a normalized proposition, proposition digest, source and
semantic digests, raw evidence and case evidence digests, stage, step, reason,
producer, consumer, meta-operation, proof choice, and target operation/output.
Dependency blocking is allowed only when the linked upstream claim is
`OPEN` or `UNKNOWN` and includes its lifecycle and evidence digests.

Actions also runs two falsifiable source variants. A semantic variant changes
the actual `direct-unknown` recipe from `missing` to `exact`, then reparses and
relowers it; its semantic digest, raw evidence digest, direct decision, and
claim transition must change. A comment-only variant changes only the raw
source and raw evidence digests; semantic IR, semantic projection, decisions,
and claim transitions must remain unchanged. The Action artifact includes the
raw receipts, producer receipt, independent report, A/B digests, snapshots,
manifest, and a human-readable summary.
