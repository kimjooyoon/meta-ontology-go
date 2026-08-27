# Causal CI selection source authority

`main.gooo` is the policy authority. Its typed activities and semantic value
programs define:

```text
changed-file -> claim -> surface -> check
claim + prior state -> discharged | open/lower-resolution | refuted
```

The JSON passed to the producer is not a case corpus. CI creates it from the
actual pull-request `git diff` and isolation snapshots. It contains paths,
statuses, prior claim observations, and before/after repository status; it
contains no known flag, decision, selected check, or reason.

[`prior-claims.json`](prior-claims.json) is the raw predecessor ledger
observation. CI joins its `OPEN` state to each observed changed path; the
producer appends the resulting transition and does not invent the predecessor
state while evaluating the plan.

The producer parses, formats, lowers, and semantically hashes the `.gooo`
source before reconstructing the policy. The separate `consumer` package does
the same from raw source and raw observation and does not import the producer.

`semantic-intervention.gooo` changes the semantic target from `go-test` to
`go-vet`; `nonsemantic-intervention.gooo` changes comments/layout only;
`contradiction-intervention.gooo` declares two selective targets for one
surface. CI emits an intervention artifact rather than claiming that any
selected check ran. All outputs are plan-only and repository read-only.
