# Captured native interruption, not a simulated execution

The two NDJSON inputs are byte-preserved files from GitHub Actions artifact
10167275876, produced by repository `kimjooyoon/meta-ontology-go`, run
34510267647, job 102985404638, source
`d7635e11a2e1c70cf9cd1ac1fe38436cc51802ad`.

Source run: https://github.com/kimjooyoon/meta-ontology-go/actions/runs/34510267647
Tracking issue: https://github.com/kimjooyoon/meta-ontology-go/issues/713

The first invocation completed. Its separate replay was cancelled at
2026-09-10T18:32:14Z; journal preservation succeeded at 18:32:25Z. The external
cancellation cause is not inferred from the journal. Capture metadata records
the original artifact names, invocation identities and byte digests.

The first file has 24 events. Replay has 16 events and ends after entering the
`CollapseAssignReturn` verifier for `CollapseFixture`. The action entry is
event 11 and the verifier entry is event 16. Both invocations share head, plan
and manifest digests but have different invocation IDs. A completed earlier
attempt must not supply the missing terminal of another attempt.

This separate regression cohort has six cases: completed input, interrupted
input, combined distinct attempts, a foreign-attempt terminal, a terminal with
a mismatched head, and truncated terminal JSON. The last three are explicitly
mutated corpus inputs, not additional native compiler executions. Expected
diagnostic reconstruction counts are asserted in CI, not declared as passed
by this document.

Use the existing production `readCostReport` consumer. Keep
`source_authenticity=UNVERIFIED`, `improvement=UNKNOWN` and the diagnostic,
non-additive scope. Captured input and a digest are not a signature, permission
grant, semantic-effect proof, or evidence of automatic resume. These cases do
not reproduce cancellation injection or close the original pagination
`ExtractFunction` counterexample. No operation or activity denominator changes.
