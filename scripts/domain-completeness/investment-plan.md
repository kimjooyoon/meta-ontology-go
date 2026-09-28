# Domain Investment Plan

The domain capability layer does not report a universal completeness score. `PlanDomainInvestment` compares observed capability coverage with an explicitly supplied investment target and preserves evidence boundaries.

- `UNKNOWN` means the domain, scope, counts, or target is not bound.
- `DEFERRED` means the evidence is incomplete or incomparable.
- `INVESTIGATE` means unresolved capability boundaries take priority over broadening scope.
- `INVEST` means observed coverage is below the declared target.
- `MAINTAIN` means only that the supplied target is met; it is not a proof of correctness or completeness.

The signed `gap_numerator` compares ratios without floating-point rounding. Scope and evidence digests remain part of the plan digest so a plan cannot be detached from the measurement context.
