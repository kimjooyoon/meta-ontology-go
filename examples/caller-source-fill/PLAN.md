# Caller-guided complete IR assignments

Frozen before implementation and own-model observation. Baseline compiler source:
`772acf4abaa93c01e666909e1c8b9029406f003a` (public Gooo 0.6.13-dev).

## Language gap

`source_fill` can fill condition, assignment and return expressions together, for
integer and record bodies. `body-construct` currently rejects these bodies even
though ordinary composition can execute them. Connect their complete declared
assignments to caller feedback without changing local expectations or the local
body-fill chooser's existing behavior.

## Implementation

- Add a source-bound planning/realization API for one complete fill assignment.
  Re-derive explicit or grammar-generated candidates from the original source;
  preserve the existing 2..16 candidate and hole/type contracts. Record every
  hole expression, exact local training values and separate local holdout values.
- A selected assignment is a checked caller proposal, not a rewritten historical
  local-winner receipt. Retain original source, selected source and plan digests.
- Extend bounded joint construction across fills, record masks and integer search.
  Initial local selection comes first, then source order for fill assignments;
  use existing 1..64 whole-program budget and 1..16 body limit. Emit v4 whenever
  fills participate; retain v1/v2/v3 compatibility and v3-style search rejections.
- Only training and caller cases influence selection. Source-local holdout rows
  remain measured separately, including mismatches, and never enter ranking or
  completion decisions. Clearly name the finite scope.
- Carry the existing optional `--fill-model` through initial construction only;
  saved replay loads no model. Recompute assignments, cases and whole-program
  histories when replaying. Preserve cancellation and terminal tool errors.
- Connect the v4 observations to the public Go/Gooo ecosystem workbench reader
  and its automatic feedback loop, so users can consume the new language path.

## Use case and observation

Build a Gooo next-budget helper with two holes: the boundary condition and the
value to keep at the limit. Its local example cannot distinguish the alternatives;
actual caller feedback must select a bounded result. Exercise local variables,
conditional assignment, record fields and another activity calling the helper.
Keep a too-small-budget partial result. Add independent final inputs, including
an exact integer above 2^53. Retain local holdout results separately.

Also exercise integer multi-hole fills, grammar-derived fills, explicit binds,
conflicting local expectations and mixed fill/search/record combinations. Reuse
the existing own graph QAT model for a frozen mixed program; do not train or
change weights after observing results. If a compatible fill model is available,
record its separate role; never label a synthetic test provider as trained-model
performance. No heavy model downloads or restored repositories.

## Verification

Test source/candidate bounds, hole grouping, exact values, local conflict,
holdout independence, saved replay and modified histories. Check that cancellation
and invalid source stop before loading models; verify fill-model flags and no
new inference on replay. Compare native outputs against fixed original cases.
Run focused and race tests and authoritative CI; retain failures, model decisions,
counts and source identities. Report programs, runs, rounds, combinations,
rejections and native executions separately. Finite cases and grammar coverage
must not be presented as all-domain correctness or general intent accuracy.
