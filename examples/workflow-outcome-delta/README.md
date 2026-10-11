# Change a queue batch rule and inspect the difference

`batch16.gooo.fixture` and `batch32.gooo.fixture` keep the same semantic IDs and
change the batch cap and requirements. `Normalize` handles negative inputs, `BatchSize` assembles
its branch from declared choices, and `Main` calls both. These are synthetic
rules with finite checks; no queue service or customer environment is connected.

Use the development [comparison guide](../../docs/workflow-outcome-delta.md) for
the complete construction and before/after commands. Output directories must be
new. The separate caller evaluation includes large signed integers and cap
boundaries. Some inputs overlap the source checks, and the same inputs are used
in both modes; this is not a held-out model accuracy benchmark.

The `.gooo.fixture` files use the repository's construction-example convention;
the CLI reads them directly. Their contents are copied byte-for-byte from the
source/protocol frozen before the
[public queue batch observation](https://github.com/kimjooyoon/meta-ontology-go/wiki/Small-Workflow-Pilot).
That original observation ran once with public0.6.26 and is retained intact.
The new comparison command can read its saved results without another execution.

## Compare the ordinary Go implementation

[`referencego/batch.go`](referencego/batch.go) implements the same two rules in
ordinary Go. Its test reads the original `evaluation16.json` and
`evaluation32.json` with `int64` values, including the exact large integers:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./examples/workflow-outcome-delta/referencego
```

This checks the Go counterpart against the same twenty caller cases. It does not
run Gooo construction or a model, extend the original observation, or measure
developer effort. When comparing a real team's workflow, include writing the
initial declarations and cases, making the next rule change, diagnosing failures
and explaining the result. The [adoption plan](../../docs/adoption-plan.ko.md)
keeps those unmeasured costs visible.
