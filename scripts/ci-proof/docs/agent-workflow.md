# Agent CI workflow

Each PR is an independent goal. An agent must continue implementing, testing,
committing, pushing, and auditing its own PR even when another PR is blocked.
Do not wait for another agent's permission or external evidence.

The governance contract is explicitly `mode=ci_only`. Dev CI closure
does not consume human reviews, approval actors, or last-push approval fields.
It is determined by exact tuple identity, the six GitHub-app proof jobs, registered
scope, current artifact digest/binding, checked-in policy digest, and the
no-write/provenance contract. Branch protection is retained as a separate
promotion predicate and remains fail-closed when its observer is unavailable.

PR failure ownership and protected-branch push ownership are separate. A
pull-request manifest must use the exact registered `agent/*` owner branch.
For a protected `push`, the owner is the normalized event branch itself and
must be one of `dev` or `main`; the event ref must be exactly
`refs/heads/<owner>`, and `pr_number` must be zero. Unknown, stale, omitted, or
malformed refs fail closed and route remediation through the catalog/gate
handoff rather than inventing an agent owner.

The Guardian is a separate `pull_request_target` context. Guardian v2 freezes
the union of protected-kernel paths (`.github/workflows/**`, governance and
transition policy, `scripts/ci-proof/**`, `scripts/ci-evidence/**`,
`scripts/verify/**`, `internal/verify/**`, `go.mod`, and `go.sum`) and fails closed
on any add, modify, delete, or rename, including `previous_filename` renames. It
checks out only the immutable base SHA and does not parse candidate YAML as an
authorization decision. Comments and inert YAML markers therefore cannot bypass
the path gate, and fork PRs receive the same read-only treatment. A Guardian PASS
is kernel-safety evidence only; CI scope ownership and the exact PR policy remain
separate conjunctions.

The default branch is `dev`. Feature `agent/* -> dev` runs emit exactly
`CI guardian shadow`; exact `dev -> main` runs emit exactly `CI guardian`.
Both routes re-read stable live dev/main refs before and after inspection. A
promotion additionally requires `ahead > 0`, `behind = 0`, and live main as the
merge base. Guardian v2 PASS is `head_binding_status=verified` only after these
checks; FAIL_CLOSED artifacts retain `CI-GUARDIAN-HEAD-BINDING-UNVERIFIED`.
The main promotion proof carries an independent Guardian run/job/check/artifact
tuple; it is not a seventh `proofJob`. Ordinary PRs cannot modify
the kernel. Future kernel rotation requires a maintenance ledger with before/after
policy digests and code/tests proving the new base-pinned guardian; it is not a
human-review predicate or an `agent/ci-workflow` exemption.

The Guardian producer manifest is `gooo/ci-guardian/v2`; the proof's immutable
observer envelope is the distinct `gooo/ci-guardian-evidence/v1` schema and must
carry `head_binding_status=verified`, the selected app/check/job IDs, exact
action/topology, and an RFC3339 run timestamp. An old or unverified envelope is
not promotion evidence.

Bootstrap caveat: this PR is evaluated by the already-integrated default-dev
workflow, so its `pull_request_target` run may retain the old `CI guardian` job
name. That result is recorded as expected-negative trust-root evidence; it does
not prove the candidate `CI guardian shadow` route. The shadow name is probed by
a later non-kernel feature PR after this workflow is integrated.

Feature PRs target `dev`; only exact `dev -> main` is a promotion. The former
`integration` ref is retired and must not be used for routing or ownership.

Post-bootstrap dev protection is strict and app-bound: it must contain the six
canonical contexts plus exactly `CI guardian shadow`. Main protection must
contain the same six plus exactly `CI guardian`. The current bootstrap PR is
allowed to be audited against the live six-context dev policy only as the
explicit one-time migration exception; it does not establish steady-state
eligibility. Before any subsequent merge, the gate must activate and re-read
dev's seven-context protection. A main promotion is fail-closed unless both
trusted dev-shadow and main-Guardian protection snapshots are exact seven-
context observations.

The Guardian promotion observer validates the `guardian-observer` environment
with the base token before minting a current-repository GitHub App installation
token with Administration:read. The App private key is environment-scoped,
never a regular PR secret, and is not exposed to feature shadow routes. Token
rotation, revocation, and environment deployment policy are provisioned by the
gate outside this owner PR; absence or API failure is not inferred as verified.

Cross-scope relationships are local dependencies. Record them as
`CI-DEPENDENCY-001` with `blocking_scope=local` and `parallelizable=true`; keep
unrelated work moving. A blocker in one PR must never be promoted into a global
stop for other PRs.

If the requested code meaning conflicts with the meaning already present in the
repository, do not overwrite either meaning and do not create a duplicate or
alias. Record `CI-CONTRACT-001`, preserve the evidence, return to the requestor,
and take another independent task.

For every CI failure, read the versioned `gooo/ci-failure/v1` manifest first.
Use its exact source/base/head/run/attempt/job tuple and evidence references;
unknown values are fail-closed. The workflow summary repeats the manifest's
code, severity, scope, blocking scope, parallelization, handoff requirement,
catalog path, and evidence references. Re-run only after the current head is
confirmed.

The failure manifest also exposes the immutable `catalog_digest`, sorted
`failure_codes` (the complete mapped set), all proof `rejections`, exact
`owner_branch`, `artifact_status`, `artifact_reason`, bound evidence and proof
artifact records, ordered terminal failure records/codes, and
protection/provenance `missing_reasons`. A proof artifact that is
missing, malformed, stale, or bound to another run is never treated as a
successful proof: emit `CI-ARTIFACT-001` or `CI-FRESHNESS-001` with the exact
reason and require a fresh artifact. Every terminal job is collected in
deterministic ID/name order with an explicit mapping; an unrecognized terminal
job emits `CI-UNCLASSIFIED-001`, never an inferred success or guessed category.
The dedicated DAMP/DRY cap step emits `CI-CAPS-001` separately from scope
failures.

The proof and provenance receipt schemas are v3 after the mandatory promotion
GuardianEvidence migration. v2 proof/receipt bundles fail closed and cannot be
used as promotion evidence. The provenance receipt is bound to the same repository, event, base/head/ref,
PR, run/attempt, workflow, canonical job records, artifact inventory, branch
protection snapshot, domain evidence, digests, and predecessor list as the
proof bundle. A mismatch is a receipt failure, not a reason to accept the
underlying proof.

Main promotion bundles also carry a live PR observation and a separate pure
`promotion_authorization`: only an open, non-draft, unmerged, mergeable, clean
same-repository `dev -> main` PR with exact current refs/topology, seven-context
protection, six canonical jobs, Guardian evidence, and immutable artifacts can
produce `PASS` with `operation=fast_forward`. The authorization is bound to the
proof bundle digest; it never mutates refs or protection. Any missing or stale
state remains `FAIL_CLOSED`.

Normal operations are:

1. Reconfirm the existing PR, branch, base, and current head.
2. Work only in the registered path scope and run the required local checks.
3. Commit intentionally and push normally to the existing branch.
4. Audit the fresh PR-authoritative run, canonical jobs, artifacts, and
   branch protection using GitHub's current state.
5. Merge or promote only when exact-head CI and the separate actual branch
   protection predicate permit it.

Admin bypass, force-push, branch aliasing, CI-policy weakening, and reuse of
stale evidence are prohibited. CI green is necessary but never sufficient for a
protected promotion.

When the terminal failure set is empty, the failure report emits a separate
`gooo/ci-closure/v1` artifact. Its `NO_TERMINAL_FAILURE` status is bound to the
same repository/base/head/event/run/attempt tuple and all six terminal
canonical-success jobs. `HEALTH_PASS_ONLY`, `write_effect=none`, and the empty
terminal-failure arrays are explicit: this artifact never asserts promotion,
provenance, or mergeability. Missing or mismatched canonical jobs
must fail closed instead of silently producing a note-only result.

Immediately before any promotion, the gate must perform a final exact re-read
of the live dev and main refs and both current dev/main protection snapshots.
Only that exact tuple may be followed by a normal CAS/fast-forward operation;
force-push and force-update operations are never permitted.
