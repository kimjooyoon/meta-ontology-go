# Domain record identity

This example uses two single-field records with the same display name and integer layout.
The money and quantity entities have separate stable IDs. Imported activity ports retain
that distinction; a quantity result bound to the money input is rejected before execution.

The source-authored calculation multiplies price by quantity. Two finite cases expect
29 × 4 = 116 and −3 × 5 = −15. These cases include the two upstream `Keep` outputs.

```sh
go run ./cmd/gooo package execute --json --cases examples/domain-record-identity/cases.json examples/domain-record-identity/workspace.json
```

The [observation](observation.json) belongs to the original fresh in-memory Go test at
compiler `d8779545bf82b8688993d77ef50c1aa0feec2b82` plus one hashed test overlay.
The public `.fixture` files preserve source content; their external manifest has different
filenames. This packaged CLI command has not been locally executed. Its compatibility
is checked by this submission's original CI, separately from the saved observation.

Incoming JSON does not identify a currency or the origin of a number. The example checks
source-authored entity IDs and bindings. Currency, unit conversion and value obligations
need their own source declarations and observations.

See the [Korean guide](../../docs/language/domain-record-identity.ko.md).
