# Non-monotonic refutation experiment

This experiment makes evidence-driven knowledge revision a meta-semantic
operation. It is deliberately narrower than an event-sourcing system: the
producer and consumer do not replicate a mutation log. They independently
parse and lower the same `.gooo` source, reconstruct the observation model,
classify values, and retain every prior claim state in an append-only ledger.

The source is authoritative for three subjects, six ordered observations, the
equality predicate, expected and observed values, provenance and evidence
digests, prior state, and explicit revision policy. A `computes` value never
contains `SUPPORT` or `REFUTE`; the oracle derives `SUPPORT` when the observed
value equals the expected value and `CONTRADICTING` otherwise.

The fixed denominator is three claims and six observations/transitions:

| Claim | Source observations | Replayed status history | Policy |
| --- | ---: | --- | --- |
| `alpha` | 1 | `OPEN -> DISCHARGED` | `NEVER_REOPEN` |
| `beta` | 2 | `OPEN -> DISCHARGED -> REFUTED` | `NEVER_REOPEN` |
| `gamma` | 3 | `OPEN -> DISCHARGED -> REFUTED -> DISCHARGED` | `REOPEN_IF_NEWER_ADMISSIBLE` |

The resulting counts are `OPEN->DISCHARGED=3`,
`DISCHARGED->REFUTED=2`, and `REFUTED->DISCHARGED=1`. Two claims are currently
`DISCHARGED`, one is currently `REFUTED`, and nine status values are retained
including the three initial states. Current discharge coverage is `2/3 = 6666`
basis points; it is not a claim that every subject is true.

## Formal choices

The experiment adopts two principles from the primary literature:

1. Doyle's *A Truth Maintenance System* treats beliefs as supported by
   recorded justifications and updates those dependencies when information
   changes. We adopt justification-bearing observations, explicit provenance,
   and append-only historical annotations. Source: [Doyle, *A Truth
   Maintenance System*, Artificial Intelligence 12(3), 1979](https://doi.org/10.1016/0004-3702(79)90008-0).
2. Reiter's *A Logic for Default Reasoning* formalizes defaults that may be
   defeated by later information. We adopt non-monotonic revision and reject
   the assumption that a prior `DISCHARGED` result remains valid after
   counterevidence. Source: [Reiter, *A Logic for Default Reasoning*,
   Artificial Intelligence 13(1-2), 1980](https://doi.org/10.1016/0004-3702(80)90014-4).

Adopted implementation principles:

- `syntax.ParseFile -> bidir.Lower` is performed independently by producer and
  oracle. The oracle reconstructs a private wire model and does not import or
  trust producer semantics.
- A transition records `before`, `after`, accepted/rejected status, evidence
  digest, provenance, proof choice, `stage`, `step`, generated reason, and the
  previous transition digest. This makes order and refutation basis replayable.
- `REFUTED -> DISCHARGED` is admissible only when the source explicitly says
  `REOPEN_IF_NEWER_ADMISSIBLE`. An unknown or forbidding policy preserves the
  current state, records a rejected attempted transition, and returns
  `FAIL_CLOSED` with `LOWER_RESOLUTION`.
- Conformance (did the source reconstruct and replay exactly?) is reported
  separately from subject resolution (what current statuses were observed?).

Rejected principles:

- no Go `CanonicalContract` table duplicated by the oracle;
- no conclusion labels or basis sentences supplied as evidence input;
- no arbitrary default logic, confidence ranking, probability, global conflict
  resolution, or repository mutation;
- no generic event-log replication presented as knowledge revision.

## Interventions and falsifiability

Actions-only CI runs the baseline twice, then runs three source interventions:

- changing a real observed value changes the semantic digest, classifier,
  transition result, and decision;
- adding only a comment changes the raw source digest but preserves the lowered
  semantic digest, transition chain, and verdict;
- changing and resealing producer observations and its receipt digest is still
  rejected because the oracle reconstructs the source model independently.

CI observes repository status before and after every read-only run; the fixed
suite denominator is `source_claims=3/3`, `source_observations=6/6`,
`producer_imports=0/0`, `semantic_causality=1/1`,
`nonsemantic_preservation=1/1`, `persistent_transitions=6/6`, and
`coherent_tamper_rejected=1/1`.

This is a read-only semantic experiment, not a promotion gate and not a
general-purpose knowledge base.
