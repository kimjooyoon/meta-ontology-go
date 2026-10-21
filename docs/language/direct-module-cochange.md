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
  -version '<exact-proposed-version>'
```

The JSON includes `cochange_outputs`, their current SHA-256 preimages,
`unresolved`, the Gooo decision and a PROV-O graph. `prov:used` links the
observation to source entities and the downloaded module. The decision's
`gooo:cochangeScope` lists the paths that must move together. These are
**planned** outputs; the observation graph does not claim that new files were
generated.
PROV-O relation values use JSON-LD `@id` node references, so consumers follow
graph edges rather than string labels.

For the current `gooo-jev` pin, the first measurable coverage is five observed
surfaces out of five declared, with zero unresolved paths. Its current-version
decision is `NO_VERSION_CHANGE`. The decision runtime also has three active
version references in the root README and scalar-source guides outside this
five-file candidate. Their preimages are observed, and Gooo returns
`UNRESOLVED_OUTPUTS` for a proposed runtime version. Historical release and
observation records remain outside the write scope. Without `-materialize-dir`,
the command only observes and writes no repository file.

`-materialize-dir /tmp/gooo-module-candidate` asks the detached writer to
replay the exact plan, recheck preimages and create all five files plus a
`candidate-receipt.json` outside the checkout. The receipt binds every output
to its old and new digest with `prov:wasDerivedFrom` and
`prov:wasGeneratedBy`. The release document also names the projected Gooo
payload as its origin. An unchanged license is copied and recorded rather
than pretending its bytes changed. It separately reports five files written,
five output bytes read back and verified, the number whose content changed,
elapsed microseconds through byte verification (including module download),
`build_status: not_run`, and `application_status: not_applied`. For the
`gooo-jev` commit
`678e5bc3eb6f714ebfec79f9ba3e05e9312bf920`, the local candidate wrote
five coherent outputs; four changed and the license text stayed identical.
An earlier decision-runtime dry run covered only those five files; the current
Gooo decision blocks that incomplete candidate.
The release archive's `NOTICE` is generated from the projected payload at
release time; it is a downstream artifact rather than a sixth committed file.

The writer does not apply the candidate to a branch or claim a passing build.
It refuses a proposed module whose own `go.mod` adds graph directives such as
`require` or `replace`: those would need more than the five declared outputs.

PROV-O records the origin of each input and output occurrence. The Gooo
activity owns the affected paths and the decision; the Go adapter only observes
and projects exact bytes. A provenance edge alone never changes a file. The
known active references outside the declared write set are recorded as
unresolved and block projection. The current runtime proposal demonstrates
this with three active documents. New references are not discovered globally
yet; the next closure step needs an independent search of tracked files.

For each subsequent upgrade, record declared, observed, projected and verified
output counts separately, plus unresolved references, independent manual edits,
candidate verification time, CI time and whether the candidate was accepted.
The first local Jev candidate verified five files (four changed) in 54 ms with
a warm module cache; that number excludes build, CI and human review and is not
an end-to-end upgrade time. The prepayment only pays off when the accumulated
time saved on later accepted upgrades exceeds the time spent defining and
maintaining this plan. No repeated-upgrade series has established that point.
