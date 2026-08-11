# Staged verifier conformance plan

`meta-ontology-go` starts as a Go-hosted implementation and is intended to
reimplement itself with `gooo`. The verifier must migrate in stages while
retaining an independently reproducible trust boundary.

| Stage | Authoritative result | Required CI behavior | Promotion gate |
| --- | --- | --- | --- |
| 0 — Go baseline | Go verifier | Run the Go verifier and all deterministic gates. The `gooo check` and generated-freshness checks remain explicitly deferred while the CLI is a stub. | Current default; no parity claim is made. |
| 1 — dual evidence | Neither implementation alone | Run Go and `gooo` verifiers in parallel on the same pinned checkout and fixtures. Normalize and compare semantic results, scope decisions, generated freshness, and evidence manifests; any mismatch fails. | Reproducible builds, identical evidence across repeated runs, independent comparison logic, and a tested rollback to Stage 0. |
| 2 — gooo authority with fallback | `gooo` verifier | Make `gooo` the required result while retaining Go as a fallback and independent comparator. Preserve both evidence bundles and roll back when the authoritative result cannot be independently checked. | Sustained Stage 1 parity, reviewed evidence samples, reproducible bootstrap output, rollback rehearsal, and an approved promotion decision. |
| 3 — fallback removal | `gooo` verifier | Remove the Go fallback only after preserving the previous verifier as a pinned, independently runnable artifact. Keep reproducible build inputs, evidence manifests, and rollback automation. | Independent parity audit, reproducible rebuild, recovery from a forced mismatch, and a reviewed governance change. |

## Evidence and trust rules

- Go and `gooo` produce separate append-only evidence bundles before comparison.
  A verifier must not certify its own output as the only trust root.
- Comparisons use stable semantic IDs, canonical serialization, fixture inputs,
  toolchain identity, and content digests. Re-running a stage from the same
  checkout must produce the same evidence and decision.
- An unavailable verifier, missing revision, evidence mismatch,
  generated-output drift, or non-reproducible build blocks promotion and leaves
  the previous stage active.
- Every stage after Stage 0 retains a documented rollback to the last known-good
  stage. Rollback is exercised in CI or an equivalent reproducible verification
  job before fallback removal.
- The workflow stage variable is fail-closed. Setting it to a future stage
  before its implementation and gates are reviewed fails CI.
