# Guardianless CI and bottleneck findings

## Change made

The separate Guardian workflow and its privileged evidence chain have been
removed from the active CI path. Promotion evidence is gathered in the regular
CI run from the current pull request and live `dev`/`main` refs. A `main`
promotion now accepts only `head=dev`; the one-time Foundation authorization,
review identity, last-push approval, reconciliation route, Guardian artifact,
branch-protection snapshot, and the hard-coded PR 602 human decision are not
inputs to that decision.

The repository's six canonical contexts remain `gofmt`, `go vet`, `go test`,
`go test -race`, `Semantic conformance`, and `CI policy`. They are ordinary
GitHub status checks. The live `main` protection rule now requires exactly
these six; the `dev` rule has no required checks. The removed app token and
duplicate observer did not add a seventh verification of the program.

## What the run records show

| Evidence | Observation | Interpretation |
| --- | --- | --- |
| [Guardian run 35458233471](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/35458233471), 2026-09-19 | Its job log spans about 7 seconds. App-token minting was skipped; inspection failed with `CI-FOUNDATION-AUTHORIZATION-001: Guardian dispatch live candidate tuple is not exact`. | This sample is not evidence of a long Guardian runtime. It shows a brittle one-time tuple check creating a failed promotion path, so rerunning or repairing the tuple was the work it imposed. |
| [PR #739](https://github.com/kimjooyoon/meta-ontology-go/pull/739), CI run [34052986715](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/34052986715), policy job 101541360114, 2026-09-06 | The policy job spans roughly 48 minutes, from 19:03:09 to 19:51:09 UTC in its step logs. The declared meta-operation plan execution and replay each occupy roughly 23 minutes; receipt validation and final verification follow. | This is the measured runtime bottleneck in the inspected sample. The long work is inside `CI policy`, not the Guardian job. One run identifies a candidate for profiling; it does not establish a stable baseline or prove how much a change would save. |

These are observations from two individual runs. Job and step timestamps are
reported to whole seconds, parallel work cannot be added as if it were serial,
and neither example establishes a causal savings estimate.

## System metric

`gooo.metric.ci.required-check-bottleneck.v1` is emitted in the CI effort
observation report. It reads the six canonical terminal jobs for one exact
workflow run and head, reports the slowest and next-slowest check, and reports
their duration difference. The result is `UNKNOWN` if a required job is missing,
duplicated, skipped, not terminal, attached to another head, or below source
timestamp resolution. The metric is descriptive and does not block CI or
promotion. Its adoption remains `UNOBSERVED` until reports provide repeatable
coverage.

The result narrows follow-up to the slow check. Step observations in the same
report can then show which operation dominates that check. A useful baseline
needs repeated runs and should retain run identity, head SHA, job and step
intervals, cache state, queue time, and parallelism. Optimization should be
credited only after same-scope before/after observations; lower check time alone
does not prove a particular code change caused the improvement.

The CI effort observer also emits `gooo.metric.ci.required-check-baseline.v1`.
It reads up to 20 recent successful Actions runs with the same repository,
workflow ID, event, full ref, and head branch. It reports each check's median
duration and the current run's delta after at least five complete, unique,
successful samples. Run IDs, attempts, head SHAs, excluded-sample reasons, and
an input digest are retained in an artifact configured for 90 days. If history
is unavailable, fewer than five samples qualify, or the current run is not fully
successful, the corresponding result remains `UNKNOWN`. The measurement uses
job wall time, excludes queue wait, and remains descriptive. A workflow ID does
not prove that the workflow definition stayed the same, so workflow edits can
create discontinuities and comparisons do not attribute cause.

Failure manifests now use `gooo/ci-failure/v2`: `handoff_required`, handoff
owners, and the branch-registration pointer are gone. Reports bind the exact
`head_branch` as identity and carry a closed `next_operation` code. The current
workflow reports that code with its evidence; it does not yet run an agent that
repairs source or updates refs from the report.

## Remaining work toward no human intervention

- The branch-to-path ownership check has been removed from the active PR path.
  Any valid `agent/*` branch receives the same six full checks; legacy tables
  remain only for verifier compatibility fixtures and can be retired separately.
- CI emits promotion authorization as evidence but does not merge or update
  refs. The remote `main` rule now requires exactly the six canonical checks;
  its up-to-date, linear-history, admin-enforcement, force-push, and deletion
  settings were preserved. The `dev` rule still has required checks disabled.
- The v2 failure report removes person assignment and emits a bounded system
  action key. An executor that performs safe source repair and replays the
  checks is still needed; ambiguous contracts remain unresolved and fail closed.

Accordingly, this change removes human approval evidence, the active per-branch
scope-registration block, and the separate Guardian workflow from checked-in
CI. It adds system-derived per-run bottleneck and repeated-run baseline reports.
GitHub's `main` rule was updated separately to remove only the Guardian check.
Automatic promotion and the self-repair of failed checks remain future work.
