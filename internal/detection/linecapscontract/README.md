# Line-cap experiment contract

This package is an independent AST oracle and fixture harness. It is safe to
build before `internal/detection/linecaps` is integrated and does not claim
that the future gooo-hosted verifier exists.

## Falsifiable hypothesis

For valid Go source, physical file lines and AST function spans are invariant
under LF versus CRLF input and under case evaluation order. A file or function
exactly at 300/75 passes; one line over fails. Unparseable source fails, while
a valid case whose future host-parity criterion is not implemented is deferred.

The hypothesis is falsified if normalized measurements differ, order changes
the evidence sequence, or a negative/deferred fixture is reported as pass.

## Reusable contract

`Measure` emits source digest, physical file lines, parse status, and sorted
function spans. `EvaluateCase` applies only the local cap criterion. `EvaluateCases`
canonicalizes case order by stable ID. `Evidence.JSON` emits
`gooo/linecaps-evidence/v1` for later IR, codegen, CI, and provenance adapters.

`pass` means only that this local measurement criterion passed. `fail` means a
cap or parse rule failed. `deferred` means the local checks passed but a named
future criterion is unavailable; it is never equivalent to a self-hosting or
promotion pass. `OutcomeMatches` compares actual decision with the declared
fixture expectation.

## Follow-up implementation contract

The production detector should consume the same `Case` source snapshot and
emit equivalent `Measurement` and `Finding` values, including SHA-256 digest,
inclusive source spans, rule IDs, and deterministic ordering. A BX/IR/codegen
adapter may add provenance or semantic IDs, but must preserve these normalized
fields. CI may compare canonical JSON evidence from this package and the
adapter; missing, mismatched, or deferred fields remain non-promotable.
