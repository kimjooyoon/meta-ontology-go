# Non-monotonic refutation experiment

This experiment makes evidence-driven knowledge revision a meta-semantic
operation. It is deliberately narrower than an event-sourcing system: the
producer and consumer do not replicate a mutation log. They independently
parse and lower the same `.gooo` source, reconstruct the observation model,
classify values, and retain every prior claim state in an append-only ledger.

The source is authoritative for three claim propositions, observable subjects
and inputs, predicates, expected values, ordered observations, provenance, and
evidence digests. A `computes` value contains no conclusion, state, decision,
relation, or revision-policy label. The independent consumer derives
`SUPPORTS`, `CONTRADICTS`, `INSUFFICIENT`, or `UNKNOWN` from the proposition and
observation instead of repeating a source label.

All claim ledgers begin at `OPEN`. Sufficient support discharges a claim,
direct contradiction refutes it, and insufficient or unknown evidence leaves
it `OPEN` with `LOWER_RESOLUTION` plus a stage, step, and reason. Every
observation appends a ledger row; no earlier claim state is deleted.

The fixed denominator is three claims and six ordered observations/ledger
rows:

| Claim | Source observations | Replayed status history | Current status |
| --- | ---: | --- | --- |
| `alpha` | 1 | `OPEN -> DISCHARGED` | `DISCHARGED` |
| `beta` | 2 | `OPEN -> DISCHARGED -> REFUTED` | `REFUTED` |
| `gamma` | 3 | `OPEN -> DISCHARGED -> REFUTED -> DISCHARGED` | `DISCHARGED` |

The baseline has `SUPPORTS=4`, `CONTRADICTS=2`,
`OPEN->DISCHARGED=3`, `DISCHARGED->REFUTED=2`, and
`REFUTED->DISCHARGED=1`. Two claims are currently `DISCHARGED`, one is
`REFUTED`, and nine status values are retained, including the three initial
states. Current discharge coverage is `2/3 = 6666` basis points; it is not a
claim that every subject is true.

## Formal choices

The experiment adopts two principles from the primary literature:

1. Doyle's *A Truth Maintenance System* treats beliefs as supported by
   recorded justifications and updates those dependencies when information
   changes. We adopt justification-bearing observations, explicit provenance,
   and append-only historical annotations. Source: [Doyle, *A Truth
   Maintenance System*, Artificial Intelligence 12(3), 1979](https://doi.org/10.1016/0004-3702(79)90008-0).
2. Reiter's *A Logic for Default Reasoning* formalizes defaults that may be
   defeated by later information. We adopt non-monotonic revision and reject
   the assumption that a prior `DISCHARGED` result remains valid after direct
   counterevidence. Source: [Reiter, *A Logic for Default Reasoning*,
   Artificial Intelligence 13(1-2), 1980](https://doi.org/10.1016/0004-3702(80)90014-4).

Adopted implementation principles:

- `syntax.ParseFile -> bidir.Lower` is performed independently by producer
  and oracle. The oracle reconstructs a private wire model and does not
  import or trust producer semantics.
- Raw and lowered semantic digests are bound together. A receipt also binds
  the reconstructed source model, so a coherently resealed producer payload
  cannot replace source evidence.
- A transition records `before`, `after`, relation, accepted/rejected status,
  evidence digest, provenance, proof choice, `stage`, `step`, generated
  reason, and the previous transition digest. This makes order and refutation
  basis replayable.
- Conformance (did source reconstruction and replay pass?) is reported
  separately from subject resolution (what current ledger statuses were
  observed?).
- The experiment records no mutation authority, promotion, repository write,
  confidence ranking, or global conflict policy.

Rejected principles:

- no Go `CanonicalContract` table duplicated by the oracle;
- no conclusion labels or basis sentences supplied as evidence input;
- no refutation string counting, self-declared decision JSON, generic event
  log replication, arbitrary default logic, confidence ranking, or repository
  mutation presented as knowledge revision.

## Interventions and falsifiability

Actions-only CI runs the baseline twice, then independently runs four source
interventions:

- changing gamma's observed value from `1` to `0` changes its relation from
  `SUPPORTS` to `CONTRADICTS`, changes the final subject distribution from
  `DISCHARGED=2;REFUTED=1;OPEN=0` to
  `DISCHARGED=1;REFUTED=2;OPEN=0`, and changes the final ledger transition;
- changing alpha's predicate to an unknown predicate yields `UNKNOWN`, keeps
  the claim `OPEN`, and lowers subject resolution;
- removing an observed value yields `INSUFFICIENT`, keeps that claim `OPEN`,
  and records the lower-resolution coordinate and reason;
- adding only a comment changes the raw source digest but preserves the
  lowered semantic digest, relation sequence, subject resolution, and ledger
  transition chain;
- changing and resealing producer observations and its receipt digest is
  rejected because the oracle reconstructs the raw source and independent
  observation model.

CI exposes fixed denominators individually rather than hiding them in an
aggregate score: `source_claims=3/3`, `source_observations=6/6`,
`producer_imports=0/0`, `semantic_causality=1/1`,
`nonsemantic_preservation=1/1`, `persistent_ledger_rows=6/6`,
`persistent_transitions=6/6`, `repository_writes=0/0`,
`mutation_promotions=0/0`, and `coherent_tamper_rejected=1/1`.

This is a read-only semantic experiment, not a promotion gate and not a
general-purpose knowledge base.
