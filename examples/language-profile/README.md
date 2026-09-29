# Gooo language profile experiment

This experiment profiles the `PayOrder` activity declared in `examples/billing/main.gooo`.
The compiler emits runner-scoped wall-time and Go `TotalAlloc` observations while preserving
the deterministic source-execution digest as a separate coordinate.

```sh
gooo profile --json --samples 5 --entry PayOrder examples/billing/main.gooo
```

The fixed contract covers two profile receipts and ten executions. It does not claim RSS,
cross-run performance improvement, production readiness, or business correctness.

The exact-head source-execution CI job also emits `comparison.json`. That paired
receipt compares the `PayOrder` declaration with a Go AST baseline on the same
runner and requires the projected package, namespace, operation signature, and
stable entity IDs to match in every sample. It reports median wall time and Go
`TotalAlloc` for each path plus Go/Gooo ratios in parts per million. The detailed
measurement boundaries and limits are in
[`docs/language/language-comparison.md`](../../docs/language/language-comparison.md).

Compiler conformance and delivery credit are separate stages: this profile capability remains
outside the delivery score until GitHub Actions emits and inspects an external receipt.
