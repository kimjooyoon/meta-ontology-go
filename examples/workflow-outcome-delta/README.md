# Change a queue batch rule and inspect the difference

`batch16.gooo` and `batch32.gooo` keep the same semantic IDs and change the batch
cap and requirements. `Normalize` handles negative inputs, `BatchSize` assembles
its branch from declared choices, and `Main` calls both. These are synthetic
rules with finite checks; no queue service or customer environment is connected.

Use the development [comparison guide](../../docs/workflow-outcome-delta.md) for
the complete construction and before/after commands. Output directories must be
new. The separate caller evaluation includes large signed integers and cap
boundaries. Some inputs overlap the source checks, and the same inputs are used
in both modes; this is not a held-out model accuracy benchmark.

The inputs are copied byte-for-byte from the source/protocol frozen before the
[public queue batch observation](https://github.com/kimjooyoon/meta-ontology-go/wiki/Small-Workflow-Pilot).
That original observation ran once with public0.6.26 and is retained intact.
The new comparison command can read its saved results without another execution.
