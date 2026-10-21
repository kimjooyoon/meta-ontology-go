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
generated. The observer also scans every regular file listed by the Git index
for the exact current-version byte sequence. It records the tracked path-list
digest, scanned file count, matching file count, outside paths and scan time.
The Gooo decision treats every outside match as unresolved. Untracked files
and nonregular tracked paths are not part of this scan; a nonregular or missing
tracked path stops observation rather than producing a readiness decision.
PROV-O relation values use JSON-LD `@id` node references, so consumers follow
graph edges rather than string labels.

In a local scan, the `gooo-jev` version appeared in four tracked files, all
inside the five declared outputs. A proposed Jev
upgrade remains `READY_FOR_DETACHED_PROJECTION`. The decision-runtime version
appeared in 28 files; 24 are outside the declared outputs, including three
known active documents and historical records. Gooo returns
`UNRESOLVED_OUTPUTS` for its proposed upgrade. An outside byte match is a
review obligation, not a claim that a historical record should be rewritten.
Without `-materialize-dir`, the command only observes and writes no repository
file.

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
scan covers Git-tracked working-tree bytes, not a semantic classification of
every mention. The next step is source-owned classification of active consumers
and immutable historical mentions, with independent review of that boundary.

For each subsequent upgrade, record declared, observed, projected and verified
output counts separately, plus unresolved references, independent manual edits,
candidate verification time, CI time and whether the candidate was accepted.
The scanned local Jev candidate verified five files (four changed) in 1.63 s
with a warm module cache; that number includes two tracked-file scans and
excludes build, CI and human review. It is not
an end-to-end upgrade time. The prepayment only pays off when the accumulated
time saved on later accepted upgrades exceeds the time spent defining and
maintaining this plan. No repeated-upgrade series has established that point.
