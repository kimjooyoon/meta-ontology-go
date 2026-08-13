# Agent CI workflow

Each PR is an independent goal. An agent must continue implementing, testing,
committing, pushing, and auditing its own PR even when another PR is blocked.
Do not wait for another agent's permission or external evidence.

The governance contract is explicitly `mode=ci_only`. Integration CI closure
does not consume human reviews, approval actors, or last-push approval fields.
It is determined by exact tuple identity, the six GitHub-app jobs, registered
scope, current artifact digest/binding, checked-in policy digest, and the
no-write/provenance contract. Branch protection is retained as a separate
promotion predicate and remains fail-closed when its observer is unavailable.

PR failure ownership and protected-branch push ownership are separate. A
pull-request manifest must use the exact registered `agent/*` owner branch.
For a protected `push`, the owner is the normalized event branch itself and
must be one of `integration`, `dev`, or `main`; the event ref must be exactly
`refs/heads/<owner>`, and `pr_number` must be zero. Unknown, stale, omitted, or
malformed refs fail closed and route remediation through the catalog/gate
handoff rather than inventing an agent owner.

The `CI guardian` is a separate `pull_request_target` context. Guardian v1 freezes
the union of protected-kernel paths (`.github/workflows/**`, governance and
transition policy, `scripts/ci-proof/**`, `scripts/ci-evidence/**`,
`scripts/verify/**`, `internal/verify/**`, `go.mod`, and `go.sum`) and fails closed
on any add, modify, delete, or rename, including `previous_filename` renames. It
checks out only the immutable base SHA and does not parse candidate YAML as an
authorization decision. Comments and inert YAML markers therefore cannot bypass
the path gate, and fork PRs receive the same read-only treatment.

The current integration base predates this `pull_request_target` workflow. GitHub
loads that workflow from the protected base/default topology, so this PR is the
explicit one-time `CI-ROOT-OF-TRUST-BOOTSTRAP-001` migration and cannot produce its
own authoritative guardian context. No default-branch or protection mutation is
performed here. After bootstrap, ordinary PRs cannot modify the kernel. Future
kernel rotation requires a maintenance ledger with before/after policy digests and
the code/tests proving the new base-pinned guardian; it is not a human-review
predicate or an `agent/ci-workflow` exemption. Until a real guardian run is proven
from the protected dev/default topology, the context is not claimed as an existing
required check.

During the transition, feature PRs may target `integration` or `dev`; only exact
`dev -> main` is a promotion. `integration` is temporary compatibility and remains
in the protected push owner set while migration evidence is collected.

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

The provenance receipt is bound to the same repository, event, base/head/ref,
PR, run/attempt, workflow, canonical job records, artifact inventory, branch
protection snapshot, domain evidence, digests, and predecessor list as the
proof bundle. A mismatch is a receipt failure, not a reason to accept the
underlying proof.

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
