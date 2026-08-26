# External conformance assurance eligibility

This use case asks whether the external-conformance assurance obligation may be
considered for a later activation. It does not activate the obligation.

The evaluator consumes seven sealed inputs:

1. The merged language-assurance baseline at 11/12 obligations.
2. The gomacro whole-suite report, observation, and 10/10 evaluator suite.
3. The capability-scoped report, observation, and 15/15 evaluator suite.

The fixed metric denominator is 18 indicators: 7 drivers, 4 outcomes, and 7
guardrails. The fixed regression denominator is 20 cases: 1 exact eligibility,
10 unknown fail-closed cases, and 9 invariant fail-closed cases.

Exact eligibility requires all of these facts at once:

- Whole-project gomacro compatibility remains FAIL_CLOSED at 6/8 (7500 bp).
- Selected evaluator and AST-macro capabilities remain executable at 10/10.
- The capability regression suite remains 15/15.
- The projected assurance count is 12/12 while the official count stays 11/12.
- Repository writes, official mutations, and promotions all remain zero.

An unknown top-level decision lowers resolution to UNKNOWN. A known mismatch
lowers resolution to INVARIANT_ONLY. Neither path can be treated as a fixed
point.

This evidence supports a later activation review. It does not prove whole
gomacro compatibility, overall Gooo language completion, or ecosystem adoption.
