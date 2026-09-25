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
  files. Guardians review them; Builders do not weaken them to make a change pass.

## Agent roles

- **Builder:** changes only the assigned authority view and supplies focused tests
  or runnable evidence.
- **Guardian:** inspects scope, IDs, provenance, BX laws, generated markers, and
  freshness; does not implement the feature under review.
- **Approver:** records review acceptance after required checks pass. Review
  roles do not authorize a protected-branch promotion; that decision is made by
  the CI-only proof and branch-protection contract below.
- **Docs/example Builder:** owns `docs/**`, `examples/**`, and root `README.md`,
  `CONTRIBUTING.md`, and `AGENTS.md`; it must not modify core package source.

No single agent should implement a feature, weaken its verification, and approve
the same change.

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

Work branches use `agent/* -> dev`. The only promotion route is the exact
same-repository `dev -> main` route; no intermediary branch participates in the
current contract. Governance mode is `ci_only`: review roles, approval actors,
and last-push approval fields are not CI proof inputs.

The six canonical proof jobs are `gofmt`, `go vet`, `go test`, `go test -race`,
`Semantic conformance`, and `CI policy`. Protected `dev` requires those six
contexts plus `CI guardian shadow`; protected `main` requires those six plus
`CI guardian`. Both are seven-context protections.

For a `dev -> main` promotion, CI emits a digest-bound `promotion_authorization`
with `source=dev`, `target=main`, and `operation=fast_forward`. It is `PASS`
only for a current, open, non-draft, unmerged, clean, mergeable same-repository
PR whose live `dev` ref is ahead of `main`, has `behind=0`, and has `main` as
its merge base, with exact proof, Guardian, artifact, and protection evidence.
The authorization never writes refs or protection. After a final exact reread,
only a normal compare-and-swap/fast-forward operation may update `main`; force
pushes and force updates are not permitted. Missing or stale evidence is
`FAIL_CLOSED`.

## Review caps

The current governance caps are soft review policy: 120 columns for ordinary
lines, 40 non-blank lines per handwritten slot, and 400 changed lines per normal
PR excluding generated output. Exceptions belong in the PR description; CI does
not enforce these caps yet.
