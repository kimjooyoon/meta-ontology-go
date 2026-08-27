# Causal CI selection meta-program

This experiment tests a stronger proposition than changed-file pattern
matching:

```text
PR changed-file observation
  -> Gooo claim
  -> Gooo surface
  -> registered check
  -> proof choice and plan receipt
```

The policy is the actual
[`main.gooo`](../examples/causal-ci-selection/main.gooo) source. Its typed
activity contracts and semantic value programs define `changed-file-to-claim`,
`claim-to-surface`, `surface-to-check`, and prior claim state operations. The
producer reconstructs that graph only after canonical parse, canonical format,
lowering, and semantic digest. The raw observation is generated from the PR's
actual `git diff`; it has no final known/decision/choice/reason fields.
It also records the observed checkout SHA, the `HEAD:path` blob object ID, and
the source-byte digest so a later source mutation is fail-closed.
Its predecessor state comes from the raw
[`prior-claims.json`](../examples/causal-ci-selection/prior-claims.json)
ledger observation and is joined to each observed path by CI.

## Decision boundary

The receipt separates conformance from subject resolution:

- `conformance.decision=PLAN_RECONSTRUCTION_CONFORMANCE_PASS` means the registered Gooo policy was parsed,
  lowered, and reconstructed. It does not mean any check ran.
- A subject with a complete claim-mediated path is `SELECTED` and receives a
  selective plan with `CLAIM_IMPACT_REASON`.
- A subject not bound to the source authority is `UNKNOWN`; the plan descends
  to the fixed six-check suite and carries the exact
  `CAUSAL_SELECTION/observe-subject/SOURCE_NOT_BOUND_TO_POLICY` cause.
- A contradictory policy path is `FAIL_CLOSED` only for its structurally linked
  subject; unrelated changed paths remain `UNKNOWN` and descend to the full
  suite. The exact contradiction target and claim-instance IDs are retained.

This is explicitly a `PLAN_ONLY` artifact. The producer receipt records
`execution.result=UNKNOWN`; it does not claim that a selected check ran. A
separate consumer process reparses and relowers the raw source and observation.
Only its separate adjudication receipt may report an observed exit/result and
source reconstruction `1/1` or `4/4`. That receipt preserves one exact
source-plan binding tuple per variant: plan digest, expected raw/source bytes
digest, actual consumer source digest/object ID, logical path, and binding mode.

## Claim ledger

CI observes prior claim templates and creates a content-addressed instance for
each subject from template ID, proposition, and subject path. The producer
appends a digest-linked transition, preserving stage, step, reason, evidence
digest, proposition, and provenance:

- the exact proposition `complete policy route reconstructed`: `OPEN -> DISCHARGED`;
- a selected-check sufficiency proposition remains `OPEN` until execution is observed;
- unresolved paths preserve the prior state with a state-specific persistence reason;
- an explicit semantic contradiction refutes only the structurally linked route proposition.

Evidence digests are SHA-256 values over canonical observed values, not
placeholders. The transition chain is append-only and independently replayed.

## Interventions and falsification

CI emits four source reconstructions from one raw observation, each with a
producer plan whose execution is `UNKNOWN`. Four separate consumer processes
then produce the adjudication evidence:

1. base policy selects `go-test` for the source subject;
2. semantic intervention changes the policy target to `go-vet`, changing the
   semantic digest, plan digest, and subject proof choice;
3. nonsemantic comment/layout intervention changes the raw digest while
   preserving parsed/semantic/plan digests;
4. contradiction intervention declares two selective targets for one surface,
   producing `FAIL_CLOSED` and `REFUTED` claim transitions.

The proposition is falsified if a filename-only rule selects a check without a
reconstructed semantic path, a raw observation can declare a conclusion, an
unknown subject omits the full six-check descent, a contradiction produces a
plan, a comment changes the semantic/plan digest, a receipt reports execution
before a consumer process ran, or the independent consumer can verify a
producer receipt without reparsing and relowering the raw source.

## Isolation and fixed inventories

The before/after repository observations are snapshots of tracked and
untracked paths plus Git-tree-style kind, mode, symlink-target digest, object
format, object ID, and content digests. They produce
`NET_REPOSITORY_STATE_UNCHANGED` and changed path/content counts; transient
writes and global mutation authority remain `UNKNOWN`, because a net snapshot
cannot prove that no transient write occurred. The six checks, six indicators,
and four intervention variants use exact expected/observed ID inventories.
Changed-file counts are reported as a PR-SHA subject-universe digest/count and
coverage, not as a fixed improvement denominator.

The plan-adjudication artifact preserves exact `go1.27.0`, `go version`,
`go env GOVERSION`, the registered `go tool fix help` inventory, and
`go fix -diff` stdout/stderr/exit/digests, exact command argv/cwd, subject
HEAD/tree and module digests, package-universe count/digest, and a live old
syntax fixture. Full-suite conformance requires exit `0` and an empty diff;
the fixture separately requires a nonzero exit and nonempty diff (`1/1`).

## Actual CI API observation

The predecessor SHA itself is deterministic from the pull-request event
payload (`pull_request.base.sha`) or checked-out Git history (`HEAD^` for a
push), so this experiment does not use an API to invent that coordinate. The
promotion artifact, however, is external CI evidence and cannot be recovered
from Git history. The existing promotion workflow therefore retains the
minimum declared permission `actions: read` for its GitHub Actions API calls.

The raw observations in
[`ci-observations/`](../examples/causal-ci-selection/ci-observations/) keep
transport facts and process facts only. The normal HTTP 200, malformed 200,
and missing-artifact cases are `SYNTHETIC_FIXTURE`; the observed PR #551 run
`33088310894` / job `98574425650` HTTP 403 is the sole
`HISTORICAL_FIXTURE`. The 403 is `UNKNOWN / LOWER_RESOLUTION` at
`proposal-promotion / fetch-github-evidence / CI_PERMISSION_DENIED`; it is
not a semantic contradiction or a fixed point. Its three co-located artifact
files and the upload-step observation have no observed dependency edge and
therefore remain `CAUSAL_RELATION_UNKNOWN`.

The current exact-head observation is separate. Run `33098087709`, job
`98608698224`, had `actions: read`, did not reproduce 403, and failed at
`Select exact readiness predecessor` with
`WORKFLOW_RUN_PAGINATION_INCOMPLETE`. It is recorded as
`CURRENT_EVIDENCE`, `OPEN / LOWER_RESOLUTION`, with current root `1/1` and
the three missing predecessor files as downstream `3/3` only because the
workflow step ordering is observed. If current evidence is absent, the
receipt uses `OPEN / LOWER_RESOLUTION / CURRENT_CI_EVIDENCE_UNAVAILABLE`
with root `0/1`; a historical 403 is never substituted for current evidence.

The predecessor selector now follows every GitHub `Link: rel="next"` page
until actual termination, records page URL/status/body and next-link
digests, and distinguishes incomplete totals, repeated links, page caps,
malformed links, duplicate run IDs, and malformed responses. The
[`pagination-fixtures.json`](../examples/causal-ci-selection/pagination-fixtures.json)
keeps normal multi-page, last-page, and each fail-closed case separate.
Process exit and HTTP status are both preserved, so an arbitrary exit code
alone cannot be treated as a counterexample.

## Build-graph research boundary

The experiment adopts explicit path reasoning from Bazel's first-party query
model (`deps`, `rdeps`, `somepath`, `allpaths`) and records an explanation for
the path. It rejects Bazel's build target/configuration as CI semantic
authority. See the [Bazel Query Guide](https://bazel.build/query/guide) and
[Bazel Query Reference](https://bazel.build/reference/query).

It adopts Nix's first-party transitive closure intuition for conservative full
descent when a path is unknown. It rejects store reachability and cache reuse
as evidence that a CI proof was justified. See the [Nix Derivation
Reference](https://nix.dev/manual/nix/2.22/language/derivations.html) and
[Nix Closure Glossary](https://nix.dev/manual/nix/2.35/glossary.html).
