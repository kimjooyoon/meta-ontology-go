# Branch and pull-request policy

The repository uses a two-stage integration path:

```text
agent/*  ->  integration  ->  main
```

- Work branches use the `agent/` prefix and contain one coherent change.
- Pull requests from work branches target `integration`. The CI policy job rejects a
  pull request that targets another branch.
- `integration` is the shared validation branch. It is the base for follow-up work and
  is promoted to `main` through a separate pull request.
- `main` is protected from direct pushes. Releases and other externally visible
  changes must arrive through the `integration` promotion path.
- CI checks are required before either merge. The workflow uses the same deterministic
  gate for work branches, `integration`, and `main`.

The scaffold baseline does not yet expose a working semantic CLI. Until
`cmd/gooo` implements its `check` command, the semantic CLI and generated-freshness
step prints an explicit deferred status and exits successfully. Formatting, vet, unit
tests, race tests, changed-path scope, and the DAMP/DRY caps remain mandatory
throughout this transition. Once the CLI exists, the same step automatically runs the
round-trip, evidence, scope, and generated-region checks under the
`semantic_conformance` build tag.

## Staged verifier promotion

Self-hosting is an explicit CI migration, not a replacement of the trust root
with a single self-checking compiler. The workflow currently pins
`GOOO_CONFORMANCE_STAGE=0`: the Go verifier is authoritative and the existing
format, vet, test, race, scope, DAMP/DRY, and deferred-CLI gates remain
required. Promotion criteria are documented in `.github/conformance-plan.md`
and require a reviewed CI-owned change.

The CI/verification change itself is intentionally scoped to `.github/**`,
`scripts/**`, and `internal/verify/**`. A change outside those paths needs its owning
agent and its own review boundary.

The ownership map for `agent/*` branches is explicit and lives in
`internal/verify/scope.go`. Unknown agent branches fail closed. The map assigns
each feature branch to its package, maps research branches to their individual
`docs/research/<slug>.md` file, and reserves shared CI paths for
`agent/ci-workflow`. The review table of registered aliases is maintained in
`.github/agent-scope-table.md`; the Go map remains the executable policy. The
`agent/go-version` maintenance branch is the sole
toolchain exception: it may change `go.mod` only when the diff contains `go` or
`toolchain` directives. Dependency, module, `go.sum`, and other source changes
remain outside that exception.

Pushes run the Go caps without using the push event's `before` SHA, because a
rebased-before revision may not exist in the checkout. An `agent/**` push is a
cap-only auxiliary run; it never substitutes for the six-job full matrix. Pushes
to `integration` or `main` run all six full jobs. Pull-request events also run
all six jobs; the existing required check names remain canonical (`gofmt`, `go
vet`, `go test`, `go test -race`, `Semantic conformance`, and `CI policy`) so
branch protection does not break during ownership updates. The workflow run name
and the CI policy step summary identify `PR authoritative`, `push full`, or
`agent push cap-only`, so an auxiliary push result cannot be mistaken for the
authoritative PR result. Pull-request events retain the complete base-to-head
changed-path and branch-ownership check; the verifier also reports unavailable
revisions deterministically.
