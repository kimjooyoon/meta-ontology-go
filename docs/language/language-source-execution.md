# Language source execution

## User path

`gooo run --entry PayOrder examples/billing/main.gooo` executes the activity
contract declared in the source. `--json` emits the versioned
`gooo/source-execution-receipt/v1` envelope.

## Runtime meaning

The first Gooo runtime is symbolic and ontology-native. It parses the source,
lowers it to semantic IR, binds each activity input to its declared stable
entity ID, invokes the activity transition, and produces the declared output
entity. Four ordered events make this transition replayable.

This is not Go source evaluation and does not delegate to an embedded Go
interpreter. It therefore preserves Gooo's own semantic authority.

## Failure boundary

Unknown entries and invalid syntax return exit code 1 with an explicit
`FAIL_CLOSED` receipt. Unknown top decisions are not accepted as a fixed point;
the meta evaluator lowers resolution instead.

## Claims intentionally absent

Value-level computation, multi-file execution, external dependencies,
language-level tests, debugging, and profiling remain outside this receipt.
Runner wall time and maximum RSS are observed by the user journey scorecard but
are not called improvements across runs.
