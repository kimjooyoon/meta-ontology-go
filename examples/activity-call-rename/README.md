# Select an activity through its call

This development source example points at byte 289, inside the Unicode escape in
`Runner`'s call to `Charge`. The compiler resolves the declared activity ID, then
uses the same source operation as declaration rename.

```sh
gooo refactor rename-activity --at-byte 289 --to 요금 --json main.gooo.fixture
```

The command was run once with an empty PATH. The retained receipt changes two
free calls and the declaration in three edits, retaining `urn:cursor:charge`.
The input file remained unchanged. The string value `Charge(input) stays text`,
the raw-body comment and the raw delimiters are retained in the saved output.

The [observation](observation.json) records the actual development binary and
production source overlays. It is source-only evidence: zero Gooo program
executions, model predictions and training runs. The editor adapters compiled;
this observation does not execute LSP prepare/rename requests. Cross-file callers
and cursor positions inside assembly alternatives remain outside this profile.
