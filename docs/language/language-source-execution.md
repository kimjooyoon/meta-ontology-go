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

A source that declares runtime `bind` edges cannot use this declaration-only
path; it fails closed with `SOURCE_RUNTIME_BINDINGS_UNSUPPORTED`. Supply an
explicit `--input` to select the value-plan path, which validates and executes
the declared edges for supported registered operations.

For a generated runtime-bound source, `gooo generate <file.gooo> --out <dir>`
emits `runtime-plan.json` automatically when the source declares typed binds.
The generated Go file also includes one deterministic `GoooCompose...`
function per connected group of explicit `bind` edges. It calls the activity
functions in validated dependency order, forwards each bound result to its
declared consumer, accepts unbound inputs as parameters, and returns terminal
outputs. No edge is inferred. The activity implementation slots remain the
place where domain behavior is supplied; the composition function describes
data flow, not successful domain execution.
Use `--runtime-plan <name>` only to change that artifact's relative filename.
When executing the source with `--input`, pass the generated artifact with
`--runtime-plan <dir>/runtime-plan.json` to attach source, semantic, and typed
plan identity to the execution receipt. If that file is named
`runtime-plan.json` in the current working directory, `gooo run` discovers and
validates it for a bound source without requiring the flag. Explicit
`--runtime-plan` always takes precedence. A plan file in the working directory
is ignored for sources with no typed runtime binds.

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
