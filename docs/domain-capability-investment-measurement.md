# Domain Capability Investment Measurement

This document defines what the domain-capability measurements can and cannot say.

## Measurement unit

One observation is a tuple of `domain`, `capability_id`, `outcome`, `evidence_digest`, and the source/contract identity that produced it. A repeated observation with the same identity is not new coverage.

## Ratios

- `coverage_ratio` is observed capability IDs divided by the expected capability IDs for the declared domain.
- `useful_ratio` is useful outcomes divided by observed outcomes after the observation policy has been fixed.
- `unresolved_ratio` is unresolved capability IDs divided by expected capability IDs.

These ratios describe the current observation set. They do not prove semantic completeness, correctness, safety, authorization, or product readiness.

## Investment signals

- `COLLECT_MORE_OBSERVATIONS`: the sample is too small or the evidence boundary is incomplete.
- `REVIEW_MISSING_CAPABILITIES`: expected capabilities remain unobserved.
- `REVIEW_INVESTMENT`: the observed boundary is stable enough to decide whether additional implementation work is worthwhile.
- `NO_ACTIONABLE_SIGNAL`: the current evidence does not justify changing scope.

A signal is a review prompt. It never executes work, changes permissions, or selects a feature automatically.

## Scope discipline

A domain must declare its expected capability boundary before ratios are calculated. If the boundary changes, the measurement identity changes and the previous result must not be compared as if it were the same domain contract.

A missing capability remains explicit. Unknown evidence remains `UNKNOWN`; it must not be converted into zero coverage or success merely because no failure was observed.