# Direct-module changes as a Gooo co-change plan

A direct dependency upgrade currently touches more than `go.mod`. Its two
`go.sum` entries, the release payload's pinned version, commit and license
digest, the committed license text, and the release document must describe the
same source. Editing one without the others gives a plausible-looking build
with an incomplete origin story.

[`plan.gooo`](../../examples/direct-module-impact/plan.gooo) names the five
affected surfaces and decides whether a proposed version has enough observed
origin data to enter detached projection. The Go observer reads the exact
checkout preimages and uses `go mod download -json` for the proposed module.
It checks the current version against the Gooo payload, module origin, both
`go.sum` entries and license bytes. The proposed version must resolve to an
exact module path, version, source commit, two sums and a bounded `LICENSE`.

For example:

```sh
go run ./scripts/direct-module-impact \
  -module github.com/kimjooyoon/gooo-jev \
  -version v0.0.0-20261019011127-aa4cc4ed706a
```

The JSON includes `cochange_outputs`, their current SHA-256 preimages,
`unresolved`, the Gooo decision and a PROV-O graph. `prov:used` links the
observation to source entities and the downloaded module. The decision's
`gooo:cochangeScope` lists the paths that must move together. These are
**planned** outputs; the graph does not claim that new files were generated.
PROV-O relation values use JSON-LD `@id` node references, so consumers follow
graph edges rather than string labels.

The first measurable coverage is five observed surfaces out of five declared
for each current direct module, with zero unresolved paths. Automatic
projection remains zero out of five. `NO_VERSION_CHANGE` is therefore the
expected current-version result, not an upgrade success. A changed version
can be `READY_FOR_DETACHED_PROJECTION` only after its origin is bound; the
present command still writes no repository file.

The next step is a detached writer that consumes this exact plan, rechecks
preimages, projects all five outputs, and records their new digests. Its
acceptance measure is the number of coherent outputs produced per change,
alongside unresolved outputs and elapsed time. The cost case should be
evaluated over repeated upgrades: modeling costs time now, while each later
upgrade should need fewer independent handwritten edits. This has not yet
been demonstrated by a measured upgrade series.
