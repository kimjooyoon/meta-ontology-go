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
reports `REGISTERED_VALUE_OPERATION` only when a registered operation applies.
Runner wall time and maximum RSS are observed by the user journey scorecard but
are not called improvements across runs.
