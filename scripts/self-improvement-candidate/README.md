# Self-improvement candidate

This command compiles `examples/self-improvement/candidate.gooo` and consumes an
exact read-only language observation receipt. It deterministically emits one
data-only experiment candidate when the explicit `value-level computation`
non-claim is present.

```sh
go run ./scripts/self-improvement-candidate \
  -contract examples/self-improvement/candidate.gooo \
  -head-sha "$HEAD_SHA" \
  -source-run-id "$SOURCE_RUN_ID" \
  -observation "$INPUT/first.json" \
  -output "$OUTPUT/candidate.json" \
  -check
```

`PROPOSED` means only that one fixed-schema candidate was produced. The receipt
keeps achieved improvement at `0`, the experiment target at `1`, repository
writes at `0`, and all execution, mutation, promotion, and adoption authorities
at `false`.
