# Symbolic reader request user observation

This use case consumes `symbolic-reader-request-result.json` without granting
write or promotion authority. It turns one compiler result into a smaller USER
observation rather than treating a green compiler job as proof of usefulness.

The fixed denominator is 10 indicators: OUTCOME 3, DRIVER 3, and GUARDRAIL 4.
The proof choices are FOUNDATION 4, COHERENCE 3, and REGRESSION 3. Every input
decision other than the literal `PASS` lowers the result to `FAIL_CLOSED` while
retaining the exact satisfied count.

The command is intentionally an independent consumer:

```sh
go run ./scripts/symbolic-invocation-usecase/reader-observation \
  --input symbolic-reader-request-result.json \
  --output symbolic-reader-request-user-observation.json \
  --expected-subject-sha "$GITHUB_SHA"
```

This stage does not yet claim exact-head cross-job artifact transport. CI
transport is a separate extension boundary so a fixture-level success cannot
silently become an end-to-end language claim.
