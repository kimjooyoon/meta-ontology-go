# Optional records across a package boundary

This workspace carries an EntityFields V4 `Profile` from an imported producer
to the local entry activity. It checks three distinct states: every field absent,
explicit empty/false/zero values, and a partial record with a large exact
integer.

Run the full package path, including generated Go compilation and two runtime
replays per case:

```sh
gooo package execute --json --cases cases.json gooo.workspace.json
```

The receipt should report all six activity expectations as passed. Each run
preserves omitted fields as absent and keeps explicitly supplied zero values
present across the typed bind. This finite example demonstrates only the listed
record shapes and inputs.
