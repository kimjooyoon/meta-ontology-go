# Gooo source-owned release identity

The selected development version lives in
[`version.gooo`](../../internal/meta/releaseidentity/version.gooo), activity
`release://activity/current-version`. The initial value for this contract is
`0.6.39-dev`; publication is a separate operation after exact-source acceptance.
Published tags and the [v3 contract](gooo-release-publication-v3.md) retain their
historical identities.

`go generate ./internal/meta/releaseidentity` uses the existing computed-library
generator to project that zero-input activity into Go. The CLI imports its
`Evaluate()` result. Generation selects the activity by stable ID, so a display
rename preserves the selected entry. No model, network request or source parser
runs when the CLI reads the compiled version.

The publication workflow checks the committed projection against current Gooo
source before reading `gooo version --json`. Its conformance output binds that
version to the exact candidate SHA. The requested tag must equal `v` followed
by the source-owned version. Preparation and publication receive this same
value through job outputs and verify the native archive's reported version and
the manifest against it. The contract and manifest retain the original version
source digest and activity identity.

The next version change edits the Gooo return value and regenerates its Go
projection. Independent expected-output tests keep their own expectation.
Workflow checks, manifest fields, tag messages, release titles and claim IDs
derive the selected version rather than carrying a separate release literal.
Previously generated artifacts retain their original source hashes.

The eight publication activities, nine public assets, eight checksum entries,
native platform witnesses, exact dev/readiness binding, isolated archive probe,
annotated tag check, verified draft and immutable-tag behavior remain as
specified in v3. This source-owned identity adds no merge authority or CI gate.

Local regression checks cover entry selection across a display rename, refusal
of an unknown entry, a stale projection check that preserves prior output,
explicit regeneration from changed source, native execution of that generated
library, and preservation of output after invalid source. These are functional
cases; no build-time, latency or memory reduction has been measured.
