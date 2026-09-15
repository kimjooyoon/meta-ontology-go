# Feedback predecessor cycle use cases

These fixtures describe how a new self-improvement cycle consumes the previous
canonical feedback receipt. They are executable inputs, not narrative examples:
`internal/meta/feedbackpredecessor/usecase_fixture_test.go` loads the JSON and
evaluates every case through `feedbackpredecessor.Select`.

| Use case | Causal source | Expected result | Operator action |
| --- | --- | --- | --- |
| PR starts from current dev | PR base SHA | `SELECTED` | Start the current cycle |
| Dev push follows a merge | Push before SHA | `SELECTED` | Continue the canonical chain |
| Prior push has not emitted | Push before SHA | `FAIL_CLOSED / NOT_FOUND` | Wait for canonical CI |
| Prior push failed | Push before SHA | `FAIL_CLOSED / UNSUCCESSFUL` | Repair the prior cycle |
| Artifact expired | PR base SHA | `FAIL_CLOSED / UNAVAILABLE` | Re-establish canonical evidence |
| Receipt is malformed | PR base SHA | `FAIL_CLOSED / RECEIPT_UNBOUND` | Reject the evidence |
| Reruns conflict | PR base SHA | `FAIL_CLOSED / AMBIGUOUS` | Resolve run identity |
| Observer reports writes | Push before SHA | `FAIL_CLOSED / WRITE_EFFECT` | Stop and investigate |
| Candidate uses another branch | PR base SHA | `FAIL_CLOSED / CANONICAL_UNBOUND` | Reject the candidate |

The fixture binds outcome, driver, and guardrail metrics to the selector meta
operation. Every case also requires FOUNDATION, COHERENCE, and REGRESSION proof
choices. Project-root README policy remains `NOT_APPLICABLE`; this nested example
contains its own documentation.
