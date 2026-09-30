# Agent and contributor contract

This repository is a semantic compiler, not only a text generator. Every change
should make the source view, semantic IR, Go projection, and verification evidence
more consistent. The detailed policy is in [docs/governance.md](docs/governance.md).

## Authority boundaries

- Business intent is authoritative in `.gooo` DSL declarations.
- Stable semantic IDs are authoritative identity; display names and aliases may
  change without changing meaning.
- The semantic IR is a normalized intermediary, not a replacement SSOT for
  business intent.
- Handwritten Go owns irreducible implementation logic only.
- Generated Go owns structural boundaries and must use stable generated-region
  markers. Never hand-edit those regions.
- Go analysis produces syntactic observations, candidate facts, or deterministic
  source-backed facts. Only accepted deterministic facts with provenance may
  change semantic state.
- Provenance facts and verification evidence are append-only records during a
  build; they cannot silently rewrite source intent.
- Ontology vocabulary, verifier semantics, and CI policy are protected kernel
  files. CI validates exact source scope, semantic IDs, provenance, BX laws,
  generated markers, and evidence freshness from the submitted revisions.

## Machine authority

The CI result is derived from the exact source revision and six canonical checks.
Review identities, last-push approval, and a separate Guardian result are not
inputs to the proof. Documentation and example ownership remains scoped to
`docs/**`, `examples/**`, and the root governance files.

## BX gate

The bidirectional transformation must preserve the laws in
[docs/governance.md](docs/governance.md): Get-Put, Put-Get, semantic round-trip,
locality, and provenance. Presentation changes may normalize formatting, but they
must not change stable IDs or unrelated semantic facts.

## Required checks

```sh
gofmt -w .
go vet ./...
go test ./...
go run ./cmd/gooo check examples/billing/main.gooo
```

Do not claim that a command or subsystem is supported unless it has an implemented
entry point and runnable evidence. In particular, `analyze` and `lsp` are not
stable CLI features yet. CI runs the canonical format, vet, unit-test, race,
semantic-conformance, and policy jobs; cache conformance and durable provenance
publishing are not current guarantees.

## CI-only branch flow

Work branches use `agent/* -> dev`. Promotion uses a same-repository PR to
`main` with either the exact `dev` head or a snapshot branch named
`agent/main-promotion-snapshot-<dev-sha>`. CI accepts the snapshot only when its
tree equals the live `dev` tree and its sole parent is the live `main` commit.
No other main-target head is accepted. Governance mode is `ci_only`: review
roles, approval actors, and last-push approval fields are not CI proof inputs.

The six canonical proof jobs are `gofmt`, `go vet`, `go test`, `go test -race`,
`Semantic conformance`, and `CI policy`. The live `main` protection rule
requires exactly those six; `dev` has no required status contexts.

For a promotion, CI emits a digest-bound `promotion_authorization` with
`source=dev`, `target=main`, and `operation=fast_forward`. It is `PASS` only for
a current, open, non-draft, unmerged, clean, mergeable same-repository PR whose
candidate is either a direct fast-forward `dev` head or a source-bound snapshot
with the exact live `dev` tree and live `main` parent. Topology must have
`behind=0` and `main` as its merge base, with the exact proof and artifacts.
GitHub's native branch
rules enforce the six required statuses; proof does not make a second app-bound
policy snapshot.
The authorization never writes refs or protection. Once the promotion workflow
is installed on the default branch, a successful exact-head `dev` run opens or
reuses the matching `main` PR and dispatches CI with the PR number and exact
head/base SHAs. CI re-reads that tuple, emits the digest-bound proof, and the
executor merges through GitHub's normal PR API with the expected head SHA.
The original successful `dev` promotion run owns the dispatched CI through a
bounded poll of its exact workflow, head, branch, run ID and current attempt.
A dependent merge job downloads that attempt's proof and rechecks the live
source/target tuple. Completion callbacks from dispatched CI do not own merges.
Native branch protection remains authoritative; force pushes and force updates
are not permitted. Missing or stale evidence is `FAIL_CLOSED`.

## Review caps

The current governance caps are soft review policy: 120 columns for ordinary
lines, 40 non-blank lines per handwritten slot, and 400 changed lines per normal
PR excluding generated output. Exceptions belong in the PR description; CI does
not enforce these caps yet.
