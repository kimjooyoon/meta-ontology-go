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
unknown values are fail-closed. Re-run only after the current head is confirmed.

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
