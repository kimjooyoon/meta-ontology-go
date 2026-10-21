# Direct-module changes as a Gooo co-change plan

A direct dependency upgrade currently touches more than `go.mod`. Its two
`go.sum` entries, the release payload's pinned version, commit and license
digest, the committed license text, and the release document must describe the
same source. Editing one without the others gives a plausible-looking build
with an incomplete origin story.

[`plan.gooo`](../../examples/direct-module-impact/plan.gooo) names five shared
surfaces, plus three live documentation version rules for the decision runtime, and decides
whether a proposed version has enough observed
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
Gooo source declares two archival trees, `docs/releases/` and
`docs/observations/`. A match in either is classified as historical only when
its working-tree bytes exactly equal that file at Git `HEAD`. The report records
the `HEAD` commit, those paths, and their SHA-256 preimages. These records stay
unchanged when a new candidate is projected. Edited archive files and every
other outside match remain unresolved. The Gooo decision receives separate
historical and unresolved counts; PROV-O retains the source rule, `HEAD` origin,
observed paths and file hashes. This is a path-and-byte role declaration, not a
semantic proof that every historical-looking sentence is inactive. Untracked files
and nonregular tracked paths are not part of this scan; a nonregular or missing
tracked path stops observation rather than producing a readiness decision.
PROV-O relation values use JSON-LD `@id` node references, so consumers follow
graph edges rather than string labels.

The source also names nine exact historical files whose recorded release or
execution identity must stay fixed. Each has its own stable rule ID and SHA-256
in Gooo. A matching current-version mention counts as historical only when the
working-tree bytes equal both the declared digest and Git `HEAD`. Editing one
file revokes that classification until its origin is reviewed and the source
rule is updated. The report links the historical file to both its rule and its
committed preimage.

Each active rule names one path, one exact prefix and one stable rule ID in
Gooo. The byte projector replaces only `prefix + current version` with
`prefix + proposed version` after checking a unique match and unchanged file
preimage. The README's older SDK study link stays as it was. Each candidate
output derives from its original bytes and the Gooo rule entity.
The rule columns can express further unique current-version spans without a
new Go replacement function for each document.
The README also has a preserved SDK study link. Gooo declares that link's
prefix, fixed release version and stable rule ID separately from the live pin.
For every document with a live rule, the observer checks each occurrence of the
current version against one declared live or preserved span. An extra occurrence
has no source-owned role and blocks projection. The fixed historical link stays
valid when a later upgrade changes the current pin.

In a local scan, the `gooo-jev` version appeared in four tracked files, all
inside the five declared outputs. A proposed Jev
upgrade remains `READY_FOR_DETACHED_PROJECTION`. The decision-runtime version
appeared in 28 files in the earlier scan; 24 were outside the declared outputs.
The archive role can exclude committed historical records from the write
obligation, while current documentation and other unclassified matches still
block projection. An outside byte match is an obligation to classify or model,
not a reason to rewrite it.
The README and two current model documents now have live rules. The nine
digest-bound files add to the two archival trees; mixed documents retain their
original bytes rather than being rewritten wholesale. Before these rules, the
runtime version occurred in 26 tracked files: five of six declared outputs,
ten archive records and eleven unclassified files. A read-only observation of
the unchanged current pin found 27 reference files: seven of eight declared
outputs, 19 historical files and the Gooo rule source, with zero outside
references. Its Gooo decision was
`NO_VERSION_CHANGE`, as expected for the same proposed version. This confirms
reference classification for that checkout; it does not establish upgrade
compatibility. A new version still needs a verified origin and fresh observation.
Without `-materialize-dir`, the command only observes and writes no repository
file.

`-materialize-dir /tmp/gooo-module-candidate` asks the detached writer to
replay the exact plan, recheck preimages and create every declared output plus a
`candidate-receipt.json` outside the checkout. The receipt binds every output
to its old and new digest with `prov:wasDerivedFrom` and
`prov:wasGeneratedBy`. The release document also names the projected Gooo
payload as its origin. An unchanged license is copied and recorded rather
than pretending its bytes changed. It separately reports files written,
output bytes read back and verified, the number whose content changed,
elapsed microseconds through byte verification (including module download),
`build_status: not_run`, and `application_status: not_applied`. For the
`gooo-jev` commit
`678e5bc3eb6f714ebfec79f9ba3e05e9312bf920`, the local candidate wrote
five coherent outputs; four changed and the license text stayed identical.
The decision-runtime plan now declares eight candidate files. A stale file
digest, missing unique active span or new outside reference still blocks an
incomplete candidate.
The release archive's `NOTICE` is generated from the projected payload at
release time; it is a downstream artifact rather than a sixth committed file.

The detached writer does not apply the candidate to a branch or claim a passing build.
It refuses a proposed module whose own `go.mod` adds graph directives such as
`require` or `replace`: those would need more than the declared outputs.

After reviewing a detached candidate, `-commit-branch agent/<name>` can create
a **local** branch with its verified outputs in one Git commit. This mode needs
a clean tracked checkout and a branch name that does not already exist. It
replays the observation, checks the current `HEAD` against the candidate's
origin, and asks Gooo's `AssessBranchCommit` activity to accept the exact
proposal. A temporary Git index builds the tree without changing the current
checkout. The adapter verifies the commit's tree, every output digest and the
complete changed-path set before creating the branch ref atomically. The JSON
receipt records the base and new commit, tree IDs, Gooo decision and PROV-O
origin edges. It reports `build_status: not_run`; normal branch CI remains the
independent check. This command does not push the branch or open a PR. The
current runtime pin has no unresolved exact-version references in the observed
checkout. A new version can still be refused for a changed preimage, incomplete
module origin or unmodeled module graph. Versionless API compatibility remains
outside this decision and needs independent checks.

PROV-O records the origin of each input and output occurrence. The Gooo
activity owns the affected paths and the decision; the Go adapter only observes
and projects exact bytes. A provenance edge alone never changes a file. The
scan covers Git-tracked working-tree bytes. Active references outside the
declared outputs, aliases, constructed versions and versionless API dependencies
remain unmodeled. Subsequent work should give those active consumers explicit
source-owned projection rules and verify their resulting bytes together.

For each subsequent upgrade, record declared, observed, projected and verified
output counts separately, plus unresolved references, independent manual edits,
candidate verification time, CI time and whether the candidate was accepted.
The scanned local Jev candidate verified five files (four changed) in 1.63 s
with a warm module cache; that number includes two tracked-file scans and
excludes build, CI and human review. It is not
an end-to-end upgrade time. The prepayment only pays off when the accumulated
time saved on later accepted upgrades exceeds the time spent defining and
maintaining this plan. No repeated-upgrade series has established that point.
