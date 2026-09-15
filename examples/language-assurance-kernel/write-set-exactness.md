# Write-set exactness use cases

The metric `gooo.metric.effects.write-set-exactness.v1` is meaningful only with
the meta-operation `observe-exact-write-set` and its independent receipt.

| Raw evidence | Declared | Observed | Mismatch | BPS | Decision | Resolution |
| --- | ---: | ---: | ---: | ---: | --- | --- |
| valid, equal sets | 1 | 1 | 0 | 10000 | PASS | EXACT |
| valid, unequal sets | 0 | 1 | 1 | 0 | BLOCK | EXACT |
| missing or invalid snapshot | null | null | null | 0 | FAIL_CLOSED | INVARIANT_ONLY |

The observer compares canonical before/after snapshots and does not import the
assurance evaluator. Receipts bind the subject SHA, frozen denominator digest,
observer identity, both snapshot digests, and all exact path sets.
