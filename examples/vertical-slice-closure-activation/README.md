# Vertical slice closure activation

This use case converts the merged, read-only vertical-slice eligibility
receipt into exactly one operating language-assurance meta-operation. It does
not widen either denominator or claim that the language is complete.

- predecessor: `3920551d8db1226810832f6f924783b2fddf4ccd`;
- metric: `gooo.metric.capability.vertical-slice-closure.v1`;
- operation: `close-vertical-slice`;
- before: `10/12 = 8333 bps`;
- after: `11/12 = 9166 bps`;
- activation cases: `4/4`;
- eligibility boundaries: `6/6`;
- eligibility links: `12/12`;
- eligibility indicators: `8/8`;
- activation indicators: `6 = 1 OUTCOME + 2 DRIVER + 3 GUARDRAIL`;
- transition applied: `1`;
- evaluator repository writes: `0`.

The committed capsules are byte-exact outputs of merged GitHub Actions.
`verticalsliceclosureactivation.Evaluate` binds both digests before exposing
`OperatingOperation`. Missing evidence lowers resolution to `UNKNOWN`; byte
drift closes at `INVARIANT_ONLY`; an unknown eligibility decision remains
`FAIL_CLOSED / UNKNOWN / BLOCK` and cannot become `FIXED_POINT`.
