# Non-monotonic refutation experiment

This experiment makes evidence-driven knowledge revision a meta-semantic
operation. It is deliberately narrower than an event-sourcing system: the
producer and consumer independently parse and lower the same `.gooo` source,
reconstruct a policy and fixture model, classify evidence, and retain every
ledger attempt without deleting prior claim state.

The source declares three propositions, observable subjects and inputs, a
revision-policy object, and six ordered `HISTORICAL_FIXTURE` evidence recipes.
The fixture's `observed` values exercise the semantics but do not establish
external domain truth. The source supplies no computed evidence digest;
producer and consumer independently hash canonical evidence material binding
claim ID, proposition, target address, observed material/value, fixture class,
sequence, superseded claim ID, and any superseded evidence digest.

The policy is itself source-bound metadata, not an implicit Go switch:

- `UNKNOWN` and `INSUFFICIENT` retain the prior accepted state and lower
  resolution;
- ordinary `SUPPORTS` cannot erase `REFUTED`;
- `SUPERSEDES` may correct `REFUTED` only when its target claim ID and evidence
  digest identify the currently active accepted refutation; the receipt also
  records the resolved target transition digest and current/stale status;
- `FOUNDATION`, `COHERENCE`, and `REGRESSION` each have bounded deterministic
  admission rules recorded on every ledger attempt.

All claim ledgers begin at `OPEN`. The baseline's final gamma event is an
explicit targeted correction, not last-observation-wins:

| Claim | Source observations | Replayed status history | Current status |
| --- | ---: | --- | --- |
| `alpha` | 1 | `OPEN -> DISCHARGED` | `DISCHARGED` |
| `beta` | 2 | `OPEN -> DISCHARGED -> REFUTED` | `REFUTED` |
| `gamma` | 3 | `OPEN -> DISCHARGED -> REFUTED -> DISCHARGED` | `DISCHARGED` |

The baseline has six observation attempts and six accepted state transitions:
`OPEN->DISCHARGED=3`, `DISCHARGED->REFUTED=2`, and the one targeted
`REFUTED->DISCHARGED` correction. Rejected attempts are counted separately and
never overwrite `current_evidence` or the prior accepted state. Every attempt
is still an append-only ledger row with a complete previous-digest chain.

## Formal choices

The experiment adopts two principles from the primary literature:

1. Doyle's *A Truth Maintenance System* treats beliefs as supported by
   recorded justifications and updates those dependencies when information
   changes. We adopt justification-bearing observations, explicit provenance,
   and retained historical annotations. Source: [Doyle, *A Truth
   Maintenance System*, Artificial Intelligence 12(3), 1979](https://doi.org/10.1016/0004-3702(79)90008-0).
2. Reiter's *A Logic for Default Reasoning* formalizes defaults that may be
   defeated by later information. We adopt non-monotonic revision but reject
   unqualified last-observation-wins: a correction must name the exact prior
   evidence it supersedes. Source: [Reiter, *A Logic for Default Reasoning*,
   Artificial Intelligence 13(1-2), 1980](https://doi.org/10.1016/0004-3702(80)90014-4).

Rejected principles include source-declared conclusions, source-supplied
evidence digests, copied producer contracts, refutation-string counting,
implicit revision policy, decorative proof choices, global mutation-authority
claims, and repository mutation presented as knowledge revision.

## Vocabulary and effects

Receipts distinguish `HISTORICAL_FIXTURE_ONLY` knowledge from
`ACCEPTED_APPEND_ONLY_LEDGER_EVIDENCE`; `UNKNOWN_OR_INSUFFICIENT_NOT_CURRENT_EVIDENCE`
is never promoted to current evidence. The repository effect report is scoped:
`NONE_OBSERVED_IN_NET_STATUS` means only that before/after porcelain status was
unchanged, while mutation authority is explicitly `UNKNOWN` because no global
capability probe was performed. Promotion operations observed remain scoped to
the run.

## Actions-only correction matrix

The dedicated GitHub Actions job runs Go 1.27 tests and the following fixed
regressions without local Go execution:

- `UNKNOWN` after `DISCHARGED` retains `DISCHARGED`;
- `INSUFFICIENT` after `REFUTED` retains `REFUTED`;
- ordinary `SUPPORTS` after `REFUTED` remains `REFUTED`;
- the valid gamma target discharges only with the exact prior evidence digest;
- wrong or missing correction targets remain `REFUTED` with lower resolution;
- an exact refutation digest from another claim is rejected;
- a same-claim refutation that is stale because later accepted evidence exists
  is rejected;
- only the same-claim current exact refutation is accepted for correction;
- a proof-rule rejection remains at the prior state;
- changing gamma's fixture value changes semantic digest, relation, and state;
- comment-only changes raw digest only and preserve semantic digest, relation,
  resolution, history, and transition chain;
- a coherently resealed producer receipt is rejected by raw-source
  reconstruction and the independent evidence model.

CI exposes each of the ten correction numerators and denominators separately,
with `correction_count=10/10`, and the three target regressions separately as
`target_correction_count=3/3`, plus source counts, producer-import
independence, persistent rows/transitions, tamper rejection, scoped effects,
and no mutation/promotion.

This is a read-only semantic experiment, not a promotion gate or a claim about
general-purpose external knowledge.
