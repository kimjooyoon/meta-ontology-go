# Source declaration split checks

Use the existing native Go command to inspect a candidate before scheduling a
write. It reads the same source plan and topology constraints as write mode:

```sh
go run ./scripts/source-splitter -root "$ROOT" -metrics "$METRICS" \
  -sha "$SHA" -check -subject path/to/source.go
```

An explicit subject exits unsuccessfully when its declaration cannot fit a
rendered part, the projected directory exceeds its limit, the directory has
mixed topology, or the source no longer needs splitting. The error retains the
subject and original planning reason. Missing subjects, incomplete directory
evidence, and mismatched repository/evidence revisions are also errors. No
source writes occur during this check.

Without `-subject`, `-check` remains an inventory operation: it reports planned
and blocked counts, and a blocked candidate alone does not fail that inventory.
The existing repository-wide CI uses this form. Agents deciding whether to
attempt one split should use the explicit-subject form and its exit status.

The check matches the repository HEAD with the metrics revision and reads the
working source. It does not attest that the working tree is clean, reserve
future output paths, test filesystem writes, or compile the resulting package.
Write mode rechecks the plan and still owns output collision and write errors.
Keep the source and metrics fixed between planning and execution.

This exposes an existing operation's read-only prerequisite check. Connecting
such checks to Gooo's operation selection, preserving measured debt while
selecting another ready candidate, remains tracked in
[#1782](https://github.com/kimjooyoon/meta-ontology-go/issues/1782).
