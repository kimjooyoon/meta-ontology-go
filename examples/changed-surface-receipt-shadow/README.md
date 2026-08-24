# Changed-surface receipt totality shadow

This shadow turns the existing coupling contract's changed surfaces and receipt
failure vocabulary into a language-assurance metric. It does not activate the
obligation or claim that the detector is externally complete.

The fixed denominator has six cases: exact, zero-change, missing, orphan,
duplicate, and unknown-top. Exact set equality is the only
`FIXED_POINT/EXACT` path. Missing evidence and unknown receipt decisions lower
resolution to `UNKNOWN`; structural contradictions close at `INVARIANT_ONLY`.

The useful meta-programming property is that the metric is consumed by the same
operation it describes: `totalize-changed-surface-receipts`. Six indicators
separate one outcome, two evidence drivers, and three guardrails. The proof
choices are FOUNDATION `3`, COHERENCE `2`, and REGRESSION `1`; evaluator writes
remain `0`.
