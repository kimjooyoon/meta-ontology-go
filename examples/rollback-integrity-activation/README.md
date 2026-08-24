# Rollback integrity activation

This use case converts the merged, read-only rollback-integrity eligibility
receipt into exactly one operating language-assurance meta-operation. It does
not mint a metric, widen the denominator, or claim language completeness.

- predecessor: `0908088ab35447098671263ffa7290c0363f7404`;
- metric: `gooo.metric.operation.rollback-integrity.v1`;
- operation: `verify-rollback-integrity`;
- before: `9/12 = 7500 bps`;
- after: `10/12 = 8333 bps`;
- activation cases: `4/4`;
- indicators: `6 = 1 OUTCOME + 2 DRIVER + 3 GUARDRAIL`;
- meta-operations: `6 = 3 FOUNDATION + 2 COHERENCE + 1 REGRESSION`;
- transition applied: `1`;
- evaluator repository writes: `0`.

The two committed capsules are exact outputs of the predecessor's merged
GitHub Actions run. `rollbackintegrityactivation.Evaluate` binds their byte
digests and validates `3/3` evidence capsules, `7/7` shadow cases,
`2/2` deterministic replays, and all indicator-to-meta-operation bindings.

The useful property is evidence-derived activation. The language assurance
registry can observe rollback integrity only through `OperatingOperation`.
Missing evidence lowers resolution to `UNKNOWN`; a byte mismatch closes at
`INVARIANT_ONLY`; an unknown eligibility decision is preserved as
`FAIL_CLOSED / UNKNOWN / BLOCK` and never becomes a fixed point.
