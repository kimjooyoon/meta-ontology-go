# Causal CI selection source authority

`main.gooo` is the policy authority. Its typed activities and semantic value
programs define:

```text
changed-file -> claim -> surface -> check
claim + prior state -> discharged | open/lower-resolution | refuted
```

The JSON passed to the producer is not a case corpus. CI creates it from the
actual pull-request `git diff` and isolation snapshots. It contains paths,
statuses, prior claim observations, and tracked/untracked content snapshots; it
contains no known flag, decision, selected check, or reason. The observation
also binds source bytes to the observed checkout SHA and `HEAD:path` blob.

[`prior-claims.json`](prior-claims.json) is the raw predecessor ledger
observation. CI derives a unique instance ID from template ID, proposition,
and changed path, then appends the resulting transition without inventing the
predecessor state while evaluating the plan.

The producer parses, formats, lowers, and semantically hashes the `.gooo`
source before reconstructing the policy. The separate `consumer` package does
the same from raw source and raw observation and does not import the producer.
The producer receipt is plan-only with execution `UNKNOWN`; a separate consumer
process writes the adjudication receipt and is the only source of observed
consumer result fractions.

`semantic-intervention.gooo` changes the semantic target from `go-test` to
`go-vet`; `nonsemantic-intervention.gooo` changes comments/layout only;
`contradiction-intervention.gooo` declares two selective targets for one
surface. CI emits intervention plans with execution `UNKNOWN`, then separately
adjudicates all four consumer processes. All outputs declare a bounded plan
capability; repository state is observed from path/content snapshots while
transient writes and global mutation authority remain `UNKNOWN`.
