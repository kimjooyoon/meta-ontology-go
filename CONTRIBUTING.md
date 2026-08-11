# Contributing

The language is deliberately experimental. Every change should state which
authority it changes, which semantic identities it preserves, and how the
bidirectional (BX) laws remain true. Keep proposals narrow and deterministic;
do not document an API or command that is not implemented and tested.

## Scope and authority

- `.gooo` declarations own business intent and explicit semantic IDs.
- Stable IDs own identity. Names and aliases are presentation or lookup metadata.
- Handwritten Go owns irreducible implementation logic only.
- Generated Go owns structural boundaries and must be regenerated from source.
- Provenance facts and verification records are evidence; they do not become
  business intent merely because a tool observed them.
- Ontology vocabulary, verifier semantics, and CI policy are protected kernel
  surfaces. Changes require a design note and focused regression evidence.

See [docs/governance.md](docs/governance.md) for the complete SSOT matrix and
[docs/spec.md](docs/spec.md) for the current language contract.

## Agent roles

The author is the Builder: they change the scoped source view and add evidence.
A Guardian reviews the semantic diff, provenance, generated-region integrity, and
checks without weakening the gate. An Approver decides whether the change is
acceptable. One person or agent must not implement a feature, weaken its verifier,
and approve the same change.

For documentation or example work, the allowed ownership is `docs/**`,
`examples/**`, and the root governance files `README.md`, `CONTRIBUTING.md`, and
`AGENTS.md`. Do not repair core package code as part of a documentation PR.

## Branch and PR workflow

1. Start from the repository default branch and create `agent/<area>`; for this
   documentation area, use `agent/docs`.
2. Inspect `git status` before editing. Mixed worktrees must be staged by explicit
   path; never silently include another agent's changes.
3. Keep one semantic concern per PR. Explain the authority boundary, affected
   IDs, generated regions, and evidence in the PR body.
4. Push the named branch and open a draft PR unless the owner explicitly requests
   a ready-for-review PR. Request a Guardian review before approval.
5. If a generated output needs a structural change, change the DSL or generator
   in its owning PR and regenerate it. Never hand-edit generated regions.

## Required local checks

Run the checks that match the files changed:

```sh
gofmt -w .
go vet ./...
go test ./...
go run ./cmd/gooo check examples/billing/main.gooo
```

For a documentation-only change, `gofmt -l .` is a non-mutating equivalent for
the formatting check when another agent owns dirty Go files. The current workflow
in `.github/workflows/ci.yml` runs exactly these CI steps:

- `gofmt -l .` and `go test ./...` in the Go test job;
- `go vet ./...` in the Go vet job;
- `go run ./cmd/gooo check examples/billing/main.gooo` in the semantic job.

CI does not currently run race tests, static analysis, LSP checks, cache checks,
generated-output snapshots, or automatic provenance publishing. Those are future
work, not current guarantees.

## Line caps and review size

These are review-policy caps, not compiler features, and are not currently
machine-enforced:

- keep ordinary source and Markdown lines within 120 columns; URLs, tables, and
  generated markers may exceed the soft cap when wrapping would reduce clarity;
- keep a handwritten implementation slot to 40 non-blank lines; extract named
  logic or write a design note when it grows beyond that boundary;
- keep a normal PR to 400 changed lines excluding generated output; split larger
  changes by authority boundary or semantic concern;
- do not impose a line cap on generated output by hand. Its contract is marker
  integrity and deterministic regeneration.

If a change must exceed a cap, explain why in the PR description and add the
focused evidence that makes the larger review safe.
