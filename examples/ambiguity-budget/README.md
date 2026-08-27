# Deterministic ambiguity budget

This is a read-only philosophy experiment. It measures interpretation
ambiguity with integer cardinalities, not probability or natural-language
confidence. The budget policy is authoritative in the versioned JSON contract;
the executable `.gooo` `FixedBudget` declaration binds the source to policy
version `v1`.

## Executable observation

The real [`.gooo` source](main.gooo) declares four case anchors and finite
semantic graph elements: candidate IDs, resolved/unresolved branch IDs,
branch-observation facts, and evidence-path edges. It declares no observed
integer, class, decision, resolution, or `?` label. The producer and
independent consumer both parse with `syntax.ParseFile -> bidir.Lower`, collect
the lowered graph, deduplicate IDs, and compute the three cardinalities:

`(interpretation_candidates, unresolved_branches, evidence_paths)`.

The unknown case contains candidate/path elements but no branch-observation
relation. The evaluator therefore discovers the missing
`unresolved_branches` coordinate at
`AMBIGUITY_OBSERVATION / unresolved_branches /
AMBIGUITY_COORDINATE_UNOBSERVED`; this is not a source-declared unknown label.

The contract policy is `(2,1,2)`. The derived case results are:

| case | computed set | class | subject decision / resolution | claim |
| --- | --- | --- | --- | --- |
| zero | `(1,0,1)` | `ZERO` | `PASS / EXACT` | `OPEN -> DISCHARGED` |
| boundary | `(2,1,2)` | `BOUNDARY` | `PASS / EXACT` | `OPEN -> DISCHARGED` |
| over | `(3,2,3)` | `OVER` | `FAIL_CLOSED / LOWER_RESOLUTION` | `OPEN -> REFUTED` |
| unknown | `(2,0,2)` plus an unobserved coordinate | `UNKNOWN` | `UNKNOWN / LOWER_RESOLUTION` | `OPEN -> OPEN` |

`EXACT` and `LOWER_RESOLUTION` are resolution fields, never claim states.
Every claim proposition is canonical, for example
`counts-within-budget(case:over-budget-ambiguity,budget:gooo://ambiguity-budget/policy/v1)`.
The transition stores its proposition digest, stage, step, reason, and semantic
evidence digest. Raw source provenance is stored separately, so a comment-only
source change cannot change semantic claim evidence.

## Formal accounting choices

Scott and Johnstone, [“Recognition is not parsing — SPPF-style parsing from
cubic recognisers”](https://doi.org/10.1016/j.scico.2009.07.001), motivate
retaining alternatives as a shared finite structure. Economopoulos,
[*Generalised LR parsing algorithms*](https://ir.cwi.nl/pub/14233/14233B.pdf),
§§4.4–4.4.1, describes local conflicts and packed derivation forests. This
experiment adopts explicit alternatives, branch observations, and evidence
paths as countable structural facts. It rejects probability, confidence
weights, silently selecting one derivation, and treating a parser choice as
proof of semantic correctness. SPPF/GLR implementation is not required here;
the canonical lowered graph is the shared observation boundary.

## Accounting and interventions

The versioned denominator is not a score: `cases=4`,
`integer_observations=12`, `claims=4`, `interventions=2`, and
`authority_observations=1`. Numerators remain separate. The normal receipt has
four conforming cases, eleven observed integer coordinates, two discharged
claims, one refuted claim, one open claim, two satisfied interventions, and
zero observed authority claims.

The semantic intervention adds a candidate and a linked evidence-path graph
structure to the boundary case. It changes `(2,1,2)` to `(3,1,3)`, changes the
subject resolution to `LOWER_RESOLUTION`, and changes that proposition from
`OPEN -> DISCHARGED` to `OPEN -> REFUTED`. The nonsemantic intervention adds
only a comment: raw digest changes, while semantic digest, elements, counts,
class, decision, resolution, proposition, and claim transition are preserved.

CI builds a tracked-plus-untracked workspace snapshot before and after the
probe, binds both snapshot digests into the final effects artifact, and binds
that artifact into both the producer receipt and independent judge. Repository
writes must be zero. Mutation authority is separately reported as
`UNKNOWN / NOT_OBSERVED`, rather than inferred from write-set equality.
