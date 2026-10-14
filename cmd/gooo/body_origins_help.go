package main

const bodyOriginsHelp = `Inspect the typed source origins of an ordinary computes body

` + bodyOriginsUsage + `

  gooo body-origins --activity Compute examples/typed-value-origins/main.gooo
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
`
