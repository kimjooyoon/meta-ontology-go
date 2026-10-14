package main

const bodyOriginsHelp = `Inspect the typed source origins of an ordinary computes body

` + bodyOriginsUsage + `

  gooo body-origins --activity Compute examples/typed-value-origins/main.gooo
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

With --fill-candidate, inspect an explicit source_fill candidate before selection.
Each row connects the hole declaration and candidate expression to the original
template token. An absent or repeated token is UNRESOLVED with no invented use
range. Comments and string contents are not token uses. A declaration can map even
when its candidate would fail type checking; no type result or score is claimed.
The export contains up to 8 primitive rows for a Gooo consumer. Generated grammar
candidates have no explicit assignment location and are refused in this mode.
The --field and --include-control options apply to typed value origins only.
See examples/fill-candidate-origins/README.md.
`
