# Deterministic ambiguity budget

This is a read-only philosophy experiment. The executable source declares a
budget and four integer observations with canonical `computes` programs:

`(interpretation_candidates, unresolved_branches, evidence_paths)`

The fixed budget is `(2, 1, 2)`. A known set at or below every bound remains
`PASS / EXACT` and moves its claim `OPEN -> DISCHARGED`; a known set over any
bound becomes `FAIL_CLOSED / LOWER_RESOLUTION` and moves its claim `OPEN ->
REFUTED`. The unknown case uses `?` for its missing `unresolved_branches`
coordinate, is derived as `UNKNOWN / LOWER_RESOLUTION`, and preserves
`OPEN -> OPEN` with an `AMBIGUITY_OBSERVATION / unresolved_branches /
AMBIGUITY_COORDINATE_UNOBSERVED` coordinate.

The producer parses and lowers [main.gooo](main.gooo), then observes those
programs. The JSON [contract](contract.json) supplies only source identity,
case/intervention IDs, and the fixed denominator of `2`; it has no observed
counts or result labels. The independent
verifier repeats the canonical parse/lowering and rule evaluation without
importing the producer. `CONFORMANCE_DECISION` is separate from the subject
decision, so a conforming receipt still reports the subject vector as
`MIXED / LOWER_RESOLUTION`.

## Formal accounting choices

The design adopts the “retain alternatives as a finite structural object” idea
from two formal parsing sources:

- Scott and Johnstone, [“Recognition is not parsing — SPPF-style parsing from
  cubic recognisers”](https://doi.org/10.1016/j.scico.2009.07.001), motivate an
  SPPF containing possible derivations while sharing common substructure. We
  adopt separation between alternatives and a chosen interpretation, but count
  alternatives instead of assigning weights.
- Economopoulos, [*Generalised LR parsing algorithms*](https://ir.cwi.nl/pub/14233/14233B.pdf), §§4.4–4.4.1, describes reduce/reduce conflicts,
  local ambiguity, and packed forests. We adopt explicit branch and
  evidence-path accounting and reject silently selecting one derivation as a
  confidence claim. The source does not carry ZERO/BOUNDARY/OVER/UNKNOWN or
  KNOWN/UNKNOWN labels; those are derived from observed counts and coordinate
  availability.

Probability, “likely” labels, default disambiguation, semantic correctness,
and intent recognition are explicitly not claimed. The semantic intervention
changes the boundary count from `2` to `3` and crosses the budget; the
nonsemantic intervention adds only a comment and preserves the lowered digest,
counts, class, decision, resolution, and claim transition. Both are replayed in
memory with zero repository writes, and CI compares tracked plus untracked
workspace snapshots before and after execution.
