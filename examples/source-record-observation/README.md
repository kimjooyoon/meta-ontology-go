# Asking a useful refund question

The initial example refunds zero. Keeping and negating zero both satisfy it.
`main.gooo.fixture` supplies two possible questions and a separate fixed
`Expected` activity. Only the approved refund of 17 distinguishes the bodies.

From a compiler checkout containing this development feature:

```sh
go run ./cmd/gooo body-codegen --json --activity Decide examples/source-record-observation/main.gooo.fixture
go run ./cmd/gooo body-compose --source examples/source-record-observation/main.gooo.fixture --entry Decide --cases examples/source-record-observation/evaluation.json --out /tmp/gooo-record-observation-example
```

The JSON keeps the initial case, acquired answer, question history and separate
held results. `body-codegen` reports a selected finite suite of 2/2: the original
zero case and the reference's answer for 17. Its two held cases are evaluated
afterward, with six output fields counted separately.

`evaluation.json` supplies two additional native inputs: approved amount 103 and
unapproved amount 109. They are absent from the source's selection, question and
held pools. The runtime result has its own denominator.

The first development CLI run produced these observations:

| Stage | Result |
| --- | --- |
| Original + acquired selection cases | 1 + 1, matched 2/2 |
| Held records / held fields | 2/2 records, 6/6 fields |
| Additional native outputs | 2/2 |
| Model predictions | 0 |
| Candidate evaluations / cached outputs reused | 6 / 5 |
| Reference inputs evaluated | 2, including the original case |

These are counts for this program. Latency and CPU utilization were not measured
in this feature run.

Remove `oracle "Expected"` to keep a useful unanswered question. The initial
example still matches; the report preserves the unresolved alternatives. Set
`max_candidates "1"` to see an unobserved candidate rather than a unique claim.

Tests also cover a caller requiring the opposite answer. The two bodies retain
local/caller scores of 2/2 and 0/1, or 1/2 and 1/1. That construction stays partial.

See [the contract and limits](../../docs/record-source-observation.md).
