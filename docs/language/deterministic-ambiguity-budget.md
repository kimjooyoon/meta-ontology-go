# Deterministic ambiguity budget

This experiment is a read-only meta-operation. It counts finite semantic
structures rather than imitating confidence language. The [executable source](../../examples/ambiguity-budget/main.gooo)
declares graph elements; it does not declare the resulting integers or labels.

The producer and the independent verifier each parse and lower that source
with `syntax.ParseFile -> bidir.Lower`, then derive distinct candidate IDs,
unresolved/resolved branch IDs, branch-observation facts, and linked evidence
paths. The three integer coordinates are the cardinalities of the lowered
candidate, unresolved-branch, and evidence-path sets. The contract’s versioned
policy is authoritative (`CONTRACT_POLICY`, limits `2,1,2`); the `.gooo`
`FixedBudget` program binds to policy version `v1`.

| case | derived cardinalities | derived class | subject result | claim lifecycle |
| --- | --- | --- | --- | --- |
| zero | `1,0,1` | `ZERO` | `PASS / EXACT` | `OPEN -> DISCHARGED` |
| boundary | `2,1,2` | `BOUNDARY` | `PASS / EXACT` | `OPEN -> DISCHARGED` |
| over | `3,2,3` | `OVER` | `FAIL_CLOSED / LOWER_RESOLUTION` | `OPEN -> REFUTED` |
| unknown | `2,0,2`; branch coordinate absent | `UNKNOWN` | `UNKNOWN / LOWER_RESOLUTION` | `OPEN -> OPEN` |

The unknown case is missing the branch-observation relation. The evaluator
derives the exact gap coordinate
`AMBIGUITY_OBSERVATION / unresolved_branches /
AMBIGUITY_COORDINATE_UNOBSERVED`. No `.gooo` `?`, `UNKNOWN`, `ZERO`,
`BOUNDARY`, or `OVER` label is trusted as an observation. `EXACT` and
`LOWER_RESOLUTION` are subject-resolution fields only; they never replace the
claim state.

Each claim is the canonical proposition
`counts-within-budget(case:<case-id>,budget:<policy-id>)`. Persistent
transitions start at `OPEN` and carry proposition digest, stage, step, reason,
and semantic evidence digest. Raw source digest is a separate provenance
field. Thus comment-only edits can change raw provenance without changing
claim evidence.

## Formal choices

Scott and Johnstone’s [SPPF-style parsing paper](https://doi.org/10.1016/j.scico.2009.07.001)
supports retaining alternative derivations as a shared finite structure.
Economopoulos’s [*Generalised LR parsing algorithms*](https://ir.cwi.nl/pub/14233/14233B.pdf),
§§4.4–4.4.1, describes local ambiguity, conflicts, and packed forests. We
adopt their structural separation between alternatives and resolution, but
count integer sets instead of assigning weights. We reject confidence,
probability, default choice, intent recognition, and semantic-correctness
claims. An SPPF/GLR engine is not required: the lowered semantic graph is the
independent observation boundary.

The denominator is versioned and multidimensional: `cases=4`,
`integer_observations=12`, `claims=4`, `interventions=2`, and
`authority_observations=1`. Numerators are reported separately; there is no
aggregate score. The ordinary receipt reports `4` conforming cases, `11`
observed integer coordinates, `2` discharged claims, `1` refuted claim, `1`
open claim, `2` satisfied interventions, and `0` observed authority claims.

The semantic intervention adds one candidate and one linked evidence-path
structure, crossing the boundary case from `2,1,2` to `3,1,3` and changing its
claim transition from `OPEN -> DISCHARGED` to `OPEN -> REFUTED`. The
comment-only intervention changes only raw source digest. It preserves the
semantic digest, structural elements, counts, class, subject decision and
resolution, proposition, evidence, and claim transition.

CI computes tracked-plus-untracked snapshots before and after the run and
binds both snapshot digests into the final workspace-effects artifact. The
artifact digest is included in the final producer receipt and independently
checked by the consumer. The write-set must be equal with repository writes
`0`; mutation authority remains `UNKNOWN / NOT_OBSERVED` because equality is
not a capability observation.

The claim is falsified by nondeterministic replay, declared counts accepted as
evidence, a missing unknown coordinate, an over-budget case remaining exact,
an altered denominator, a forged digest accepted by the consumer, a producer
dependency in the consumer, or any nonzero repository write-set.
