# Imported record constructors and source-owned choices

The money and count modules both declare a record named `Value` with different stable IDs.
Two application files use the same `대상` import alias for those different modules.
Constructor types follow the imports of their original file.

`Round` adds one to quantity. `Quote` clamps a negative price to zero and declares two
record-result candidates for the remaining branch: product and sum. Its source selection
case chooses the product; held-out and native caller cases have separate inputs.

```sh
go run ./cmd/gooo package execute --json --cases examples/package-qualified-record-constructors/cases.json examples/package-qualified-record-constructors/workspace.json
```

The six files here preserve the exact bytes of one fresh external-manifest CLI observation.
Its [record](cli-observation.json) retains the actual base revision, overlay hashes and
binary hash. That observation selected `product` with selection 1/1, held-out 1/1 and
native outputs 8/8 across two caller cases. The two native executions agreed; model and
provider calls were zero. The original files were unchanged and the temporary binary was removed.

Two separate API observations cover [file-local scopes](file-scope-observation.json) and
[candidate slot scopes](source-fill-observation.json). Their producers and denominators
remain distinct. Compatibility of this submission is established by its own original CI.

This scope uses flat records of single integer, string and boolean fields. Business
units, external data origin, nested records and general semantic equivalence require
their own declarations and observations.

See the [Korean guide](../../docs/language/package-qualified-record-constructors.ko.md).
