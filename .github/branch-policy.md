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

The CI/verification change itself is intentionally scoped to `.github/**`,
`scripts/**`, and `internal/verify/**`. A change outside those paths needs its owning
agent and its own review boundary.
