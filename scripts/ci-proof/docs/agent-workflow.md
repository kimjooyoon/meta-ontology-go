# Agent CI workflow

Each PR is an independent goal. An agent must continue implementing, testing,
committing, pushing, and auditing its own PR even when another PR is blocked.
Do not wait for another agent's approval, permission, or external evidence.

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
`owner_branch`, `artifact_status`, `artifact_reason`, bound artifact records,
and protection/approval/provenance `missing_reasons`. A proof artifact that is
missing, malformed, stale, or bound to another run is never treated as a
successful proof: emit `CI-ARTIFACT-001` or `CI-FRESHNESS-001` with the exact
reason and require a fresh artifact. Every terminal canonical job has an
explicit mapping; an unrecognized terminal job is a fail-closed workflow
error, not an inferred success or guessed category.

The provenance receipt is bound to the same repository, event, base/head/ref,
PR, run/attempt, workflow, canonical job records, artifact inventory, branch
protection snapshot, domain evidence, digests, and predecessor list as the
proof bundle. A mismatch is a receipt failure, not a reason to accept the
underlying proof.

Normal operations are:

1. Reconfirm the existing PR, branch, base, and current head.
2. Work only in the registered path scope and run the required local checks.
3. Commit intentionally and push normally to the existing branch.
4. Audit the fresh PR-authoritative run, canonical jobs, artifacts, reviews, and
   branch protection using GitHub's current state.
5. Merge only when exact-head CI, required independent review, and actual branch
   protection all permit it.

Self-approval, overlapping approval, admin bypass, force-push, branch aliasing,
CI-policy weakening, and reuse of stale evidence are prohibited. CI green is
necessary but never sufficient for a protected merge.
