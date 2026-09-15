# Language value witness

This command compiles the explicit `computes` clause in the Gooo example and
emits a deterministic, data-only receipt.

```sh
go run ./scripts/language-value-witness \
  -source examples/language-value-witness/main.gooo \
  -activity Increment -head-sha "$HEAD_SHA" \
  -output "$OUTPUT/receipt.json" -check
```

The receipt proves only one registry-bound `int.add` value program at
`CORE_IR_ACTIVITY_VALUE_PROGRAM` resolution. Core IR preserves the program and
changes its semantic fingerprint at `1/1`; unknown attributes still fail
closed at `1/1`. The core IR does not execute or generate code from the value
program.

When the default value-witness source is selected, the receipt also contains a
`runtime_plan` object produced by invoking the native `go run` CLI on the
runtime-binding example with input `41`. The detached plan evidence records the
observed three operation applications, two deliveries, and `43` results for
both consumers; it is validated before the witness receipt is written.
