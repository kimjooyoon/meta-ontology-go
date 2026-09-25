# Language source execution

## User path

`gooo run --entry PayOrder examples/billing/main.gooo` resolves the activity
declaration in the source. The receipt's `scope` is
`DECLARATION_RESOLUTION_ONLY`; `--json` emits the versioned
`gooo/source-execution-receipt/v1` envelope.

## Runtime meaning

The first Gooo runtime is symbolic and ontology-native. It parses the source,
lowers it to semantic IR, binds each activity input to its declared stable
entity ID, and records the declared activity/output transition as four
ordered events. The events are declaration-resolution evidence, not a claim
that a handwritten Go body ran.

This is not Go source evaluation and does not delegate to an embedded Go
interpreter. It therefore preserves Gooo's own semantic authority.

## Failure boundary

Unknown entries and invalid syntax return exit code 1 with an explicit
`FAIL_CLOSED` receipt. Unknown top decisions are not accepted as a fixed point;
the meta evaluator lowers resolution instead.

## Claims intentionally absent

Registered-value operation execution, handwritten Go-body execution, external
effects, multi-file execution, language-level tests, debugging, and profiling
remain outside this receipt. The separately scoped `--input` value-plan path
uses `REGISTERED_VALUE_OPERATION` as its declared evaluation scope. That field
does not by itself prove that a registered `Apply` call completed; the value
report's decision, invocation counts, outputs, and result evidence establish
that narrower fact. Runner wall time and maximum RSS are observed by the user
journey scorecard but are not called improvements across runs.

## Input compatibility

The `gooo/source-execution-receipt/v1` contract requires `scope`. A historical
receipt with the same v1 schema name but no scope is not evidence of
`DECLARATION_RESOLUTION_ONLY` and is rejected by the validator. No permissive
default or retrofit is applied to old evidence.
