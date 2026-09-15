# Source authority activation

This use case turns one merged, read-only eligibility receipt into one active
language-assurance meta-operation. It does not mint a new denominator.

- predecessor: `240bb8019af2f5488701cd00797a4d3598bda213`;
- metric: `gooo.metric.semantic.source-backed-authority.v1`;
- operation: `bind-source-backed-authority`;
- before: `6/12 = 5000 bps`;
- after: `7/12 = 5833 bps`;
- transition applied: `1`;
- repository writes by the evaluator: `0`.

The three committed capsules are byte-for-byte outputs of the predecessor's
merged GitHub Actions run. `sourceauthorityactivation.Evaluate` checks all three
capsule digests and the eligibility report's semantic digest before exposing the
operation to the language-assurance registry.

The fixed activation denominator has three cases: exact evidence applies one
transition, unavailable evidence fails closed at `UNKNOWN`, and a byte mismatch
fails closed at `INVARIANT_ONLY`. A same-head promotion attempt is separately
blocked because the source-authority obligation is already operating.
