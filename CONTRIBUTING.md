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

For Go changes, run the repository gate:

```sh
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
go test -race ./...
GOOO_CONFORMANCE_STAGE=0 ./scripts/semantic-conformance.sh
go run ./scripts/verify
```

For a documentation-only change, `gofmt -l .` remains a non-mutating formatting
check. Do not use `gofmt -w .` to include unrelated Go edits in a docs PR. The
current workflow in [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs
these jobs:

- `gofmt -l .`;
- `go vet ./...`;
- `go test ./...`;
- `go test -race ./...`;
- `./scripts/semantic-conformance.sh` with `GOOO_CONFORMANCE_STAGE=0`;
- `go run ./scripts/verify` for pull-request scope, branch target, and Go caps.

The semantic script first runs the Go verifier. Because `cmd/gooo` currently
contains only a command stub, it prints an explicit deferred message for
`check` and exits without claiming semantic parity. `generate`, `analyze`, and
`lsp` are also deferred; none is a successful local check until its entry point,
tests, and CI evidence exist. Stages 1--3 are rejected until a reviewed CI
promotion change enables them.

## Line caps and review size

The CI verifier enforces the Go caps; the other limits are review policy:

- keep Go files at or below 300 lines;
- keep Go functions and methods at or below 75 lines;
- keep ordinary source and Markdown lines within 120 columns; URLs, tables, and
  generated markers may exceed the soft cap when wrapping would reduce clarity;
- keep a handwritten implementation slot to 40 non-blank lines;
- keep a normal PR to 400 changed lines excluding generated output;
- do not impose a line cap on generated output by hand. Its contract is marker
  integrity and deterministic regeneration.

If a review limit must be exceeded, explain why in the PR description and add the
focused evidence that makes the larger review safe. A Go cap failure is a CI
failure. See [contracts.md](docs/contracts.md) for the implemented/deferred
status of AST, semantic, PROV, BX, codegen, LSP, cache, and self-hosting.
