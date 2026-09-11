# Bounded predecessor observation

The native predecessor selector can observe an exact-route workflow that is
still running without claiming that its proposal is available.

The existing five selection proofs and eight proposal-contract cells are
unchanged. The existing meta operations `select-merged-change-proposal` and
`lower-resolution-on-unknown-proposal-predecessor` continue to bind the
selection and unresolved-candidate indicators. This adds scheduling behavior,
not a new language-completeness score or promotion authority.

## State and authority

- Legacy `Collect` and `Select` retain their completed-run query and replay.
- Opt-in `CollectPending` and `SelectPending` preserve exact SHA, route,
  workflow, attempt, synthesis-job, artifact and proposal-contract requirements.
- Only observed in-flight runs receive `UNKNOWN / DEPENDENCY_BLOCKED`.
- Every such record keeps stage, step, reason, unknown_class, next_operation and
  a sorted run/attempt frontier in blocked_by.
- Pending observations never return a selected proposal or authorize promotion.
- Contradiction, ambiguity, absent runs, incomplete pagination, malformed input,
  mixed unresolved causes and API errors are not scheduling permission.

## Bounded CI adapter

`-predecessor-observation-attempts 13` enables pending-aware observation.
The adapter makes at most 13 observations, waits 15 seconds only between
validated dependency-blocked observations, and has a four-minute context budget.
The default remains one legacy observation. No workflow is restarted or mutated.

Each attempt is retained beside the final output as
`selection.json.attempt-NN.json`. The final selection still has to pass all
existing checks. Budget expiry preserves non-selection and exits with failure.

The CI regression policy has exactly seven cases: immediate selection, pending
then selected, exhausted observations, missing evidence, refutation,
cancellation, and an invalid success return for UNKNOWN. Invalid budgets are
tested separately. Existing route-identity and synthesis-job cases are retained.
Performance improvement and external utility remain UNKNOWN.
