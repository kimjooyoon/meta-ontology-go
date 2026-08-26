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
`BIDIR_ACTIVITY_SEMANTIC` resolution. It records the lower core IR boundary as
`0/1` and requires that unsupported lowering fail closed.
