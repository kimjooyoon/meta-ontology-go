# Causal CI selection source authority

`main.gooo` is the policy authority. Its typed activities and semantic value
programs define:

```text
changed-file -> claim -> surface -> check
claim + prior state -> discharged | open/lower-resolution | refuted
```

The JSON passed to the producer is not a case corpus. CI creates it from the
actual pull-request `git diff` and isolation snapshots. It contains paths,
statuses, prior claim observations, and tracked/untracked Git-tree snapshots; it
contains no known flag, decision, selected check, or reason. The observation
also binds source bytes to the observed checkout SHA, Git object format, and `HEAD:path` blob.

[`prior-claims.json`](prior-claims.json) is the raw predecessor ledger
observation. CI derives a unique instance ID from template ID, proposition,
and changed path, then appends the resulting transition without inventing the
predecessor state while evaluating the plan.

The producer parses, formats, lowers, and semantically hashes the `.gooo`
source before reconstructing the policy. The separate `consumer` package does
the same from raw source and raw observation and does not import the producer.
The producer receipt is plan-only with execution `UNKNOWN`; a separate process
observation records the consumer's self-report, while a plan-adjudication
receipt compares it with shell exit/stdout/stderr evidence. This proves plan
reconstruction only; selected-check execution remains `UNKNOWN`.

`semantic-intervention.gooo` changes the semantic target from `go-test` to
`go-vet`; `nonsemantic-intervention.gooo` changes comments/layout only;
`contradiction-intervention.gooo` declares two selective targets for one
surface. CI emits intervention plans with execution `UNKNOWN`, then separately
adjudicates all four consumer processes. All outputs declare a bounded plan
capability; repository state is observed from path/content snapshots while
transient writes and global mutation authority remain `UNKNOWN`.

`ci-observations/` is a separate raw API/process evidence surface. It keeps
The normal HTTP 200, malformed HTTP 200, and missing-artifact cases are three
`SYNTHETIC_FIXTURE` cases. The observed PR #551 `proposal-promotion` HTTP 403
is the sole `HISTORICAL_FIXTURE`. The consumer derives the rows from endpoint,
required
permission, HTTP status, process exit/status, and output bytes. The 403 row is
`HISTORICAL_FIXTURE`, classified `UNKNOWN / LOWER_RESOLUTION` at
`proposal-promotion/fetch-github-evidence/CI_PERMISSION_DENIED`; it is not
the current actual case, and its three co-located artifacts plus upload-step
observation have no observed dependency edge, so they remain
`CAUSAL_RELATION_UNKNOWN`.

The current exact-head observation is `CURRENT_EVIDENCE`: run `33098087709`,
job `98608698224`, with `actions: read` present and no 403. Its root is
`OPEN / LOWER_RESOLUTION / WORKFLOW_RUN_PAGINATION_INCOMPLETE`; the three
missing predecessor files are downstream `3/3` only because workflow ordering
is observed. If current evidence is unavailable, the receipt uses
`OPEN / LOWER_RESOLUTION / CURRENT_CI_EVIDENCE_UNAVAILABLE` and never
substitutes the historical 403. Selected-check execution remains `UNKNOWN`.
