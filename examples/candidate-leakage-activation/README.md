# Candidate leakage activation

This use case turns the merged, read-only candidate-leakage eligibility receipt
into one operating language-assurance meta-operation. It does not mint a new
metric or denominator.

- predecessor: `308233198b91a47bdfc34016f73e44905e2be582`;
- metric: `gooo.metric.semantic.candidate-leakage.v1`;
- operation: `detect-candidate-leakage`;
- before: `7/12 = 5833 bps`;
- after: `8/12 = 6666 bps`;
- transition applied: `1`;
- evaluator repository writes: `0`.

The two committed capsules are exact outputs of the predecessor's merged
GitHub Actions run. `candidateleakageactivation.Evaluate` binds their byte
digests, verifies the predecessor obligation is `NOT_IMPLEMENTED/NONE`, and
accepts only the explicit `ELIGIBLE/EXACT` report with `3/3` conformance.

The useful property is not merely another registry flag. The registry can see
the operation only through `OperatingOperation`, whose result is derived from
the evidence evaluator. Missing evidence lowers resolution to `UNKNOWN`; byte
or semantic mismatch closes at `INVARIANT_ONLY`; an unknown top decision never
activates. These are measured behaviors, not claims of language completeness.
