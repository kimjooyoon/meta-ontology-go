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
- **Approver:** accepts the reviewed change after required checks pass.
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
stable CLI features yet, and the current CI does not enforce race, cache, or
provenance-publishing checks.

## Review caps

The current governance caps are soft review policy: 120 columns for ordinary
lines, 40 non-blank lines per handwritten slot, and 400 changed lines per normal
PR excluding generated output. Exceptions belong in the PR description; CI does
not enforce these caps yet.
