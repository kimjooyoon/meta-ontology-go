# Partial-knowledge composition

This is a read-only Gooo meta-value experiment. It composes two operation
observations while preserving knowledge resolution:

| Case | Composition result | Case decision | Claim transition |
| --- | --- | --- | --- |
| exact pair | `EXACT` | `PASS` | `OPEN -> DISCHARGED` |
| direct unknown | `DIRECT_UNKNOWN` | `FAIL_CLOSED` | `OPEN -> UNKNOWN` |
| dependency block | `DEPENDENCY_BLOCKED` | `FAIL_CLOSED` | `OPEN -> BLOCKED` |
| invariant preservation | `INVARIANT_ONLY` | `HOLD` | `OPEN -> INVARIANT_PRESERVED` |
| mixed unknown + block | `MIXED_UNRESOLVED` | `FAIL_CLOSED` | `OPEN -> UNRESOLVED` |

The only top-success value is `EXACT`. The four non-exact cases remain
non-promotable, including the known `repository-writes-zero` invariant. The
mixed case retains both its direct-unknown operation and its blocked
dependency, so composition does not erase causal resolution.

The fixed denominator is 5. The receipt exposes 10 indicators, 5 case
results, and 5 digest-linked claim transitions. Its producer is
`partial-knowledge-producer`, its consumer is
`partial-knowledge-composition-consumer`, and its central meta-operation is
`compose-partial-knowledge`. Proof choices are explicit per case and at the
receipt boundary. Repository writes and promotion authority are both zero.

The producer and independent verifier are runnable from CI:

```sh
go run ./cmd/partial-knowledge-composition-witness \
  --head-sha "$HEAD_SHA" \
  --output "$RUNNER_TEMP/partial-knowledge-receipt.json"
go run ./cmd/partial-knowledge-composition-verifier \
  --head-sha "$HEAD_SHA" \
  --receipt "$RUNNER_TEMP/partial-knowledge-receipt.json" \
  --output "$RUNNER_TEMP/partial-knowledge-verification.json"
```

The independent verifier reconstructs the source vocabulary, fixture, result
states, indicators, claim chain, and receipt digest without importing the
producer package. Research decisions and the falsification boundary are in
[`docs/research/partial-knowledge-composition.md`](../../docs/research/partial-knowledge-composition.md).
