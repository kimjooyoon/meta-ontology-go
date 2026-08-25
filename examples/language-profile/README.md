# Gooo language profile experiment

This experiment profiles the `PayOrder` activity declared in `examples/billing/main.gooo`.
The compiler emits runner-scoped wall-time and Go `TotalAlloc` observations while preserving
the deterministic source-execution digest as a separate coordinate.

```sh
gooo profile --json --samples 5 --entry PayOrder examples/billing/main.gooo
```

The fixed contract covers two profile receipts and ten executions. It does not claim RSS,
cross-run performance improvement, production readiness, or business correctness.
