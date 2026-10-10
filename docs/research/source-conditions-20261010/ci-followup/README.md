# Follow-up to the original source-condition CI

Original compiler revision: `1928459f8f9c370c6f9fd876d5469acaada9ec01`.
Original pull-request run: [38013552778, attempt 1](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38013552778),
check suite `103000608175`. The run completed with failure. Unit, race, vet,
semantic conformance and CI policy checks succeeded. The format job failed its
modernizer/module fixed-point check; dependent evidence/proof checks failed too.
The original run was neither rerun nor canceled.

The format diff requires `strings.SplitSeq` in the recount and condition test,
plus removal of two stale v0.2.26 module checksums. This follow-up applies those
changes after the original run terminated. The original run/check metadata and
format log are preserved here with deterministic gzip compression; decompressed
bytes were compared with the downloaded originals.

The compiler also consumes public decision-runtime `v0.2.28-experimental`, source
`302fa1c032ce8f31b2e6f330fb632b4f2c9a620a`. SDK [PR15](https://github.com/kimjooyoon/gooo-decision-runtime/pull/15)
passed original push run `38014681349` and PR run `38014706839`; accepted main run
`38014862061` also passed before the tag was published. This carries one committed
condition counterexample to explicit frozen-model reconsideration.

Two new compiler regressions cover:

- Conditions reading a reassigned local value, in deterministic and model modes,
  followed by saved-projection verification.
- Gooo-declared condition failures and caller CI context reaching all three
  subsequent model inputs, with exact `-9007199254740995`, final output/condition
  checks and saved-projection verification.

The paired native study in the parent directory retains its original compiler
`62f2252d`, SDK v0.2.27, weights, inputs and outputs. It was not repeated or
relabeled as an observation of this follow-up. The frozen initial PR description
describes that earlier revision; this file records the later dependency change.

The original local full macOS suite also had failures in pre-existing platform
tests, a diagnostic assertion and native subprocess time limits under concurrent
execution. Isolated new native tests and the original Linux unit/race jobs passed.
These observations do not establish a root cause for every local timeout.

Verify the archived payload from this directory with `shasum -a 256 -c SHA256SUMS`.
