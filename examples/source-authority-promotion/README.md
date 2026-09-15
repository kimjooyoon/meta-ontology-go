# Source authority promotion eligibility

This use case consumes two independently produced CI artifacts:

- the fixed `gooo/language-assurance-denominator/v1` baseline;
- the fixed `gooo/upstream-source-conformance-denominator/v1` report.

The evaluator checks one proposed transition:

- metric: `gooo.metric.semantic.source-backed-authority.v1`;
- before: `NOT_IMPLEMENTED / NONE`;
- eligible after: `OPERATING / EXACT`;
- denominator: unchanged at 12;
- operating count: 6 to 7;
- coverage: 5000 to 5833 basis points.

An `ELIGIBLE` receipt is read-only. It reports `promotion_applied=0` and
`repository_writes=0`. A later, separately reviewed kernel change is required
to apply the transition.

Unknown or mismatched evidence lowers resolution to `INVARIANT_ONLY` and
returns `BLOCK`; it is never converted to a fixed point.
