# Integration promotion and release governance

Status: proposed operating policy. The rules below are the promotion contract;
the enforcement gaps in section 11 must be closed before a promotion merge is
authorized. This document complements the [governance contract](../governance.md),
the [branch policy](../../.github/branch-policy.md), and the
[staged verifier plan](../../.github/conformance-plan.md).

## 1. Operating model

`integration` is the shared validation branch. It may contain accepted work
that is not yet a public release. `main` is the public stable baseline. The only
normal path to `main` is a promotion pull request whose head is `integration`
and whose base is `main`.

The path is therefore:

```text
agent/* --validated PR--> integration --promotion PR--> main
```

Direct pushes, force pushes, branch deletion, and agent-to-main pull requests
are prohibited. A promotion PR is not a substitute for the checks on the work
PRs that admitted its commits.

Responsibilities are deliberately split:

| Role | Responsibility | Prohibited action |
| --- | --- | --- |
| Builder | Own the declared source view, branch, tests, and evidence. | Merge or weaken a gate. |
| Guardian | Review scope, stable IDs, provenance, BX laws, generated markers, and exact-SHA checks. | Implement the change under review. |
| Approver/release owner | Accept the candidate and release notes after Guardian evidence. | Bypass a failed or stale check. |
| CI manager | Maintain workflows, check names, ownership aliases, and branch protection. | Rename/remove a required gate to make a PR pass. |
| Promotion operator | Run the final gate for both agent-to-integration and integration-to-main, then merge only an eligible PR. | Implement code or merge from an unverified SHA. |
| Central observer | Collect evidence, detect drift, block/escalate, and report risks. | Merge, push, or directly implement. |

The promotion operator is the gate operator for the final merge. The central
observer may recommend a block but cannot override the operator's evidence gate.

## 2. Immutable candidate and evidence contract

For a promotion candidate, record these values in the PR before it becomes
ready:

```text
H_i = remote refs/heads/integration at candidate selection
H_m = remote refs/heads/main at candidate selection
P   = one PR with head=integration, base=main
```

The candidate is eligible only when `P.head_oid == H_i`, `P.base_oid == H_m`,
and `H_m` is an ancestor of `H_i`. If `main` is not an ancestor, first land a
normal, reviewed synchronization change through `integration`; do not resolve
the divergence by force-pushing either protected branch.

The operator records the exact SHA, event, run URL, conclusion, and completion
time for every required check. A green check for another SHA, a stale check
with the same name, a skipped job, a cancelled job, or a manually reported
result is not evidence. Any movement of either branch, PR update, rebase, or
workflow change invalidates the record and starts the gate again.

The final recheck is immediately before merge:

```sh
git ls-remote origin refs/heads/integration refs/heads/main
gh pr view <promotion-number> --json headRefName,baseRefName,headRefOid,baseRefOid,url
gh api repos/kimjooyoon/meta-ontology-go/commits/<H_i>/check-runs
```

The operator must compare the returned `head_sha` values with `H_i`, not merely
the check names. If a merge queue creates a synthetic test commit, the queue's
merge-group SHA becomes the new immutable candidate and must have its own full
evidence bundle.

## 3. Required CI checks

The current CI job names are the required baseline and must remain stable:

```text
gofmt
go vet
go test
go test -race
Semantic conformance
CI policy
```

All six checks must be successful on the promotion PR's exact head SHA and on
the final queue/test result when a queue is enabled. The semantic job remains
Stage 0: the Go verifier is authoritative, and a deferred `gooo check` is not
evidence that the CLI or a self-hosted verifier is supported. Advancing the
verifier requires parity, independent evidence, reproducible bootstrap, and
rollback evidence from the [conformance plan](../../.github/conformance-plan.md).

For a pull request, the `pull_request` event is the authoritative check source.
An `agent/**` push run may be reduced to cap-only advisory validation to remove
duplicate work, but `integration` and `main` push events and every promotion
PR event must retain all six full jobs. A successful push-only run, even on the
same SHA, never satisfies promotion evidence.

Required checks are an allowlist, not a minimum count. A new job is not
required until branch protection names it; removing or weakening a job requires
a separate reviewed CI policy change and an equivalent replacement.

## 4. Admission to integration

Every work PR must satisfy all of the following:

1. Head uses `agent/*`; base is `integration`; the base is the current
   `integration` tip when the PR becomes ready.
2. The ownership verifier accepts every changed path. An unknown branch or
   missing ownership alias fails closed; it is not fixed by broadening a path
   wildcard.
3. The six CI checks pass on the PR head SHA, with Guardian review for semantic,
   provenance, generated-region, or policy changes.
4. The PR is not stacked on another agent branch. A dependent PR stays draft,
   names its parent, and is retargeted to `integration` and retested after the
   parent merges.
5. A core dependency PR is either the current serialized core item or is draft.

The operator may merge independent ready PRs in any order, but every merge
creates a new `integration` SHA. Earlier checks never authorize a later SHA.
The operator therefore rechecks the resulting integration push before selecting
it for promotion.

## 5. Parallel lanes and the core dependency gate

Review and CI may run in parallel. Merge admission is serialized at the shared
branch, unless a verified merge queue supplies equivalent merge-group evidence.

The core lane currently includes changes to `go.mod`, `go.sum`, `.github/**`,
`scripts/**`, `internal/verify/**`, `internal/semantic/**`, and
`internal/syntax/**`, plus any future generated boundary or verifier authority
path registered by the CI manager. Unknown shared infrastructure paths default
to core. A core item must merge before dependent feature lanes and must be
revalidated against the current integration tip.

Two lanes are independent only when their changed paths do not overlap, neither
declares a dependency on the other, neither changes shared generated output or
policy, and both are tested against the same integration base. Independent
lanes may be reviewed and tested concurrently; they do not share or borrow
checks after either lane merges. A core merge, toolchain change, workflow change,
or conflict resolution invalidates all queued candidate evidence.

This rule handles follow-up bursts without pretending that parallel CI makes a
shared branch immutable. The queue order is recorded as PR number, base SHA,
head SHA, merge SHA, and the next required revalidation run.

## 6. Promotion PR and release cadence

The promotion operator opens at most one promotion PR at a time. It is draft
until the Builder's candidate record, Guardian review, Approver decision, and
all required checks are present. The PR body must include the checklist in
section 10 and link the exact Actions runs.

The default cadence is one promotion window per business day, with the release
owner choosing a window only after the candidate has remained green and stable
for that window. A skipped window is safer than a relaxed gate. Additional
promotions require an explicit release decision; an integration movement during
the window invalidates the candidate rather than being silently absorbed.

The release owner publishes a release note or candidate summary from `main`
only after the promotion merge and a successful `main` push run. Tags and other
external artifacts must point to that immutable main SHA. There is no claim of
automated release publishing until a checked-in entry point and CI evidence
exist.

## 7. Merge queue and branch protection

Before enabling a merge queue, CI must listen for the queue's merge-group event
and map its synthetic SHA into the same evidence contract. A queue that only
has `pull_request` and `push` coverage is not a promotion gate.

The minimum GitHub configuration is:

- protect `integration` and `main`; require pull requests, one Guardian/Approver
  review, conversation resolution, current branches, and all six named checks;
- dismiss stale approvals after updates, prohibit force pushes and deletion, and
  disallow administrator bypass for normal merges;
- allow `integration` as the only promotion head for a `main` promotion rule;
- allow the promotion operator to merge, but do not grant the central observer
  merge or push authority;
- configure a queue only after merge-group CI is green and its SHA is recorded.

The CI manager verifies the settings through the GitHub API after every change.
Until `protected: true`, required checks, and review rules are observable,
passing Actions runs are advisory evidence and no operator may merge on the
assumption that branch protection exists. This is a deferred proposal, not a
request to change settings in this governance PR: first keep the policy and
canonical check names stable, then propose the settings as a separately
reviewed administrative change.

## 8. Hotfix, rollback, and recovery

Hotfixes still use the normal authority path: `agent/hotfix-*` targets
`integration`, is reviewed and tested, and is promoted through the sole
`integration`-to-`main` PR. The release owner may shorten the cadence, but may
not bypass the path. A hotfix that starts from `main` must first be replayed or
merged into `integration` and must satisfy the ancestor rule before promotion.

Rollback is a forward, reviewable revert. Never reset, force-push, delete, or
rewrite `main` or `integration`. To recover a bad promotion:

1. Record the bad promotion merge SHA and the last-known-good `main` SHA; pause
   promotion and dependent merges.
2. Create an explicit revert change through an agent PR to `integration`, with
   the incident, affected evidence, and recovery test.
3. Run the six checks and Guardian/Approver review on the revert.
4. Open a new promotion PR and repeat the immutable-SHA gate; publish the
   recovery result against the new `main` SHA.

If checks are unavailable, a SHA is missing, or evidence disagrees, stop and
recover the verifier/runner first. The previous Go verifier remains the
rollback authority during self-hosting; a candidate `gooo` verifier cannot
certify its own promotion.

## 9. Current evidence and quantified risks

Snapshot observed after CI audit #85 (GitHub timestamps 2026-08-11 UTC; local
review date 2026-08-12):

- `integration` is `17d6b7a2dc4a96cf85c024763bd923942d9b72b6`; `main` is
  `c557daf1fd6748b2e61afca10fb632792683061f`. They remain divergent: main is
  not an ancestor of integration.
- GitHub reported both branches unprotected, zero repository rulesets, and
  automatic merge disabled. The six canonical CI names exist in
  [.github/workflows/ci.yml](../../.github/workflows/ci.yml), but are not yet
  configured as required checks by repository settings.
- There were 65 open PRs targeting `integration`: 64 drafts and one ready PR.
  There was no promotion PR targeting `main` in this snapshot. This is a
  queue, not a release candidate.
- Audit #85 now labels PR-event runs as authoritative, labels agent pushes
  cap-only, and retains all six full jobs for `integration`/`main` pushes and
  PR events. The latest `integration` push run for `17d6b7a` succeeded on all
  six jobs at [Actions run 31540055891](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/31540055891).
- PR #63 remains an open draft with base `6c6208e` and head `0a494e6`; it is
  stale and must not be force-pushed or reused as the current candidate.
- The current `agent/integration-governance` alias is present in the ownership
  map, but a fresh branch name needs its own exact alias. A missing alias must
  fail closed rather than be solved with a wildcard.

Reproduce the snapshot before acting; never treat these counts as permanent:

```sh
git ls-remote origin refs/heads/integration refs/heads/main
gh pr list --state open --base integration --limit 100 \
  --json number,isDraft,headRefName,baseRefName,headRefOid,updatedAt
gh run list --workflow CI --branch integration --limit 20 \
  --json event,conclusion,headSha,createdAt,url
gh api repos/kimjooyoon/meta-ontology-go/branches/integration/protection
gh api repos/kimjooyoon/meta-ontology-go/branches/main/protection
```

## 10. Promotion checklist

- [ ] Exactly one PR has `head=integration` and `base=main`; no agent PR targets
      `main`.
- [ ] The recorded `H_i` equals the current remote integration SHA and `H_m`
      equals the current main SHA; `H_m` is an ancestor of `H_i`.
- [ ] The six required checks are successful on `H_i`; no skipped, cancelled,
      stale, or same-name result is being reused.
- [ ] Candidate PRs have current integration bases; stacked PRs are draft or
      have been retargeted and retested.
- [ ] Core dependency work is merged or explicitly blocked; no core item is
      racing the candidate freeze.
- [ ] Guardian review, Approver/release-owner acceptance, and release notes are
      recorded.
- [ ] The promotion operator rechecked branch refs and check-run `head_sha`
      immediately before merge.
- [ ] After merge, the main push run succeeds; the merge SHA is recorded as the
      release or rollback reference.

## 11. Remaining enforcement and administrative follow-up

Audit #85 already implemented the event-source metadata, agent push cap-only
behavior, and the exact ownership alias for the original governance branch.
This fresh follow-up intentionally does not modify CI policy or branch
protection. Remaining work is:

1. Add an exact ownership alias for this fresh branch name, or select an
   already-registered branch name. Do not broaden the allowlist and do not
   force-push PR #63.
2. Extend the verifier to allow only the exact promotion pair
   `(head=integration, base=main)` in addition to the existing agent-to-
   integration rule, with regression tests. All other main PRs remain rejected.
3. Observe the canonical six names across fresh PR-event and
   `integration`/`main` push runs for a stable review window. Only then propose
   and API-verify branch protection in a separate administrative change.
4. If a merge queue is adopted, add merge-group CI and synthetic-SHA evidence
   before enabling it.

Until the exact fresh-branch alias and promotion-pair verifier exception land,
the promotion operator must report the block rather than merge around it. Main
direct pushes and early promotion remain prohibited.
