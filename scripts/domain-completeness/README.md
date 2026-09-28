# Domain completeness receipt validator

This package validates the machine-readable receipt defined by
`docs/domain-completeness-receipt.schema.json`.

The validator is intended to run in GitHub Actions after a domain observation
has produced its receipt. It returns a JSON report with `valid` and a
canonical `receipt_digest`; invalid input exits non-zero and preserves the
reason for rejection.

It checks only evidence and contract invariants:

- declaration, observation, and digest identities are present and unique;
- `AVAILABLE`, `DEFERRED`, `UNKNOWN`, and `OUT_OF_SCOPE` states are explicit;
- non-available observations retain a reason;
- metric values equal their numerator divided by denominator;
- available coverage is bound to `AVAILABLE` observations;
- provenance preserves the first missing stage and next operation;
- execution and authorization remain false.

The validator does not infer domain correctness, turn coverage into a
completion score, or treat `UNKNOWN` and `DEFERRED` as success. A new domain
declaration, catalog, schema, or toolchain identity creates a new measurement
context and must be compared as a new receipt lineage.