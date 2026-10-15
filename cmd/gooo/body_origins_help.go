package main

const bodyOriginsHelp = `Inspect the typed source origins of an ordinary computes body

` + bodyOriginsUsage + `

  gooo body-origins --activity Compute examples/typed-value-origins/main.gooo
  gooo body-origins --activity Adjust --format text examples/compound-updates/main.gooo
  gooo body-origins --activity Choose --fill-candidate wrong_type \
    examples/fill-candidate-origins/source.gooo
  gooo body-origins --activity Choose --fill-candidate wrong_type \
    --inputs-for origin/fill-candidates:Inspect \
    examples/fill-candidate-origins/source.gooo > candidate-origins.json
  gooo package execute --json --inputs candidate-origins.json \
    examples/fill-candidate-origins/gooo.workspace.json
  gooo body-origins --activity Select --include-control \
    --inputs-for typed/ancestry:Inputs examples/typed-value-origins/main.gooo > origins.json
  gooo package execute --json --inputs origins.json \
    examples/typed-value-origins/ancestry.workspace.json

Text output connects return roots, named source expressions, original line/column
locations and separate data/control edges. Gooo source defines the displayed role
labels. Both output modes describe possible source dependencies. JSON is the default.

The JSON report contains field-sensitive value definitions, return roots, input
ports and original literal byte ranges. Fixed same-source helpers retain their
argument bindings. Scalar results use an empty output field ID; for a record
result, select its stable field ID with --field when exporting input arrays.

Inspection reads source without choosing candidates, evaluating cases or calling
a model. It reports possible source dependencies, with up to 512 nodes. Unsupported
analysis returns UNRESOLVED and a nonzero exit code. The supplied Gooo consumer
accepts 1..60 nodes and follows parent edges; --include-control adds guard and
condition edges. Larger graphs and inputs above 32 KiB are refused during export.

See examples/typed-value-origins/README.md and docs/language/typed-value-origins.ko.md.

Use --fill-candidates to list source-owned finite assignments and their IDs.
With --fill-candidate, inspect one retained source_fill assignment before selection.
Each row connects the hole declaration and candidate expression to the original
template token. An absent or repeated token is UNRESOLVED with no invented use
range. Comments and string contents are not token uses. A declaration can map even
when its candidate would fail type checking; no type result or score is claimed.
The export contains up to 8 primitive rows for a Gooo consumer. For derived grammar assignments, a distinct report and 19-field input row retain
the grammar location. No direct candidate or expression literal location is invented.
Grammar coverage measures enumeration; type and functional outcomes are separate.
The --field and --include-control options apply to typed value origins only.
Use --inputs-shape source-control-columns with --inputs-for to export retained
branch polarity, constant truth and node activity IDs for a Gooo reader. The
default ancestry-columns shape is unchanged. The control shape keeps the whole
retained activity graph and refuses --field. See examples/source-control-reader.
Use --inputs-shape input-declaration-columns to export original signature positions,
type IDs and byte ranges for source/control:InspectInputs. Record field nodes share
one signature position. The bound is 16 inputs and 32 KiB; --field and
--include-control do not apply. Text output also shows signature line/column locations.
See examples/fill-candidate-origins/README.md and
examples/derived-candidate-origins/README.md.
`
