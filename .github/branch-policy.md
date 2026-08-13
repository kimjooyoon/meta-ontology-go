# Branch and pull-request policy

The repository uses the steady-state promotion path:

```text
agent/*  ->  dev  ->  main
```

- Work branches use the `agent/` prefix and contain one coherent change.
- Pull requests from work branches target `dev`. The CI policy job
  rejects a work branch targeting `main` or another branch.
- The former `integration` ref is retired and is not a current branch, trigger,
  protected-push owner, or promotion source. The only valid promotion is an exact
  `dev` head targeting `main`.
- `main` is protected from direct pushes. Releases and other externally visible
  changes must arrive through the separate `dev`-to-`main` promotion path.
- CI checks are required before either merge. The workflow uses the same deterministic
  gate for work branches, `dev`, and `main`.
- The checked-in governance mode is `ci_only`: the proof still contains exactly
  six canonical jobs, while steady-state dev protection requires those six plus
  the app-bound `CI guardian shadow` (seven contexts). Main requires those six
  plus app-bound `CI guardian` (seven contexts). Human reviews and last-push
  approvals are not CI proof inputs. This bootstrap PR is the one-time
  expected-negative under the currently observed six-context dev protection;
  after it lands, the gate must activate dev's shadow context before any later
  PR is merged. Main promotion requires trusted seven-context snapshots for
  both dev and main; inaccessible protection cannot be inferred from CI.

## Immutable CI trust kernel

The `CI guardian` job is a read-only `pull_request_target` check. It checks out
only the immutable base SHA and uses base-pinned code to paginate the changed-file
API. It inspects both `filename` and `previous_filename`; any add, modify, delete,
or rename touching the protected kernel fails with `CI-ROOT-OF-TRUST-001`. It does
not parse candidate YAML, inspect PR text, execute candidate code, or use write
permissions, so comments and inert workflow markers cannot authorize a change.

The default branch is now `dev`. A probe must compare runtime SHA/ref, workflow
SHA/ref, event PR head SHA, and the external expected tuple before any
required-context decision. The guardian summary retains
`CI-GUARDIAN-HEAD-BINDING-UNVERIFIED` until that probe proves the binding; it is
not a required status or merge/promotion authorization. A future kernel rotation
is an explicit maintenance operation with before/after policy digests and an issue
ledger, not a human review predicate or ordinary PR exemption.

During this bootstrap PR, `pull_request_target` still executes the Guardian
workflow integrated on the current default `dev`; it cannot emit or prove the
candidate's new `CI guardian shadow` route name. Record that old-base Guardian
result as bootstrap expected-negative evidence. A later non-kernel feature PR,
after this workflow is integrated, is the probe for `CI guardian shadow`; do not
pretend this PR has already activated that name or make it required here.
The promotion route validates the named `guardian-observer` environment before
minting a current-repository GitHub App installation token with only
Administration:read. The environment/App configuration is provisioned outside
this owner lane; missing or malformed configuration remains fail-closed.

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

Pushes to `dev` or `main` run all six full jobs. Pull-request events also run
all six jobs; the existing required check names remain canonical (`gofmt`, `go
vet`, `go test`, `go test -race`, `Semantic conformance`, and `CI policy`) so
branch protection does not break during ownership updates. The workflow run name
and the CI policy step summary identify `PR authoritative` or `push full`.
Pull-request events retain the complete base-to-head
changed-path and branch-ownership check; the verifier also reports unavailable
revisions deterministically.

The pull-request trigger explicitly includes `ready_for_review`, ensuring a
draft-to-review transition receives the same six-job authoritative matrix.

Every audit must re-read the live PR base/head, event/run/attempt, canonical
jobs, artifacts, and protection predicates. Historical PR narratives, preview
merge SHAs, stale workflow runs, and auxiliary push results are not evidence for
the current gate. Unknown agent branch names remain rejected by the executable
scope map.

Observer freshness is bounded to exactly ten minutes: every verified
`guardian-observer`, dev-protection, and main-protection snapshot records the
GitHub response `Date` as `observed_at` and derives `valid_until` from that
timestamp. Missing, malformed, future, or expired freshness evidence is
fail-closed. Steady-state dev/main are seven contexts each: the six canonical
proof contexts plus the app-bound Guardian context for that branch.
