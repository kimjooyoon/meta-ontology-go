# Compare the outcomes of a workflow change

The published [0.6.27 development release](releases/0.6.27-dev.md) adds `gooo body-outcomes-delta`. It reads saved results from
`body-construct` or `body-compose` and shows what changed for each caller input.
The public binary supports text and JSON output. The current development source
also supports Markdown; build it with Go1.27.2 for all commands below.

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o /tmp/gooo-outcomes ./cmd/gooo
/tmp/gooo-outcomes body-outcomes-delta --before before.json --after after.json
/tmp/gooo-outcomes body-outcomes-delta --before before.json --after after.json --json
```

The default output gives counts and changed/unassessed inputs with their old and
new values. JSON retains every row, including unchanged values, original case
indices and source identities. Exit0 means the comparison was produced; inspect
the assessment counts for regressions. The command does not call a model,
execute generated programs, write files, or contact a provider.

## A small rule change

The [batch example](../examples/workflow-outcome-delta/README.md) normalizes
negative queue counts to0 and chooses a batch size bounded by16, then32. The
source declares four candidate combinations and output/condition requirements.
Its intentionally incorrect starting recipe makes candidate selection visible.
This fixed rule is also easy to implement directly in Go; adoption value needs
evidence from repeated changes in an actual workflow.

```sh
/tmp/gooo-outcomes body-construct \
  --source examples/workflow-outcome-delta/batch16.gooo.fixture --entry Main \
  --construction-cases examples/workflow-outcome-delta/construction.json \
  --cases examples/workflow-outcome-delta/evaluation16.json \
  --attempts 4 --out batch16-run > before.json

/tmp/gooo-outcomes body-construct \
  --source examples/workflow-outcome-delta/batch32.gooo.fixture --entry Main \
  --construction-cases examples/workflow-outcome-delta/construction.json \
  --cases examples/workflow-outcome-delta/evaluation32.json \
  --attempts 4 --out batch32-run > after.json

/tmp/gooo-outcomes body-outcomes-delta --before before.json --after after.json
```

Use fresh output directories. Both constructions are deterministic here. Adding
a compatible released model to construction only changes candidate order; the
comparison itself never needs a model file.

### Share a change report

The development source after 0.6.27 also exports a Markdown report for a pull
request or team discussion. This flag is absent from the 0.6.27 public binary:

```sh
/tmp/gooo-outcomes body-outcomes-delta --before before.json --after after.json --markdown > change.md
```

Choose either `--markdown` or `--json`. The report includes every aligned group,
before/after input digests and recorded runtime stages. Long values are shortened
at 120 characters; JSON retains full values, original indices and fault details.
The report is written to stdout; the shell redirection above creates the file.

Each share names its denominator. Regressions are divided by previously matching
groups with unchanged expectations; improvements use previously failing groups.
Changed requirements, unobserved assessments and conflicting duplicates use all
observed groups. A missing denominator is `n/a`, and percentages are descriptive
shares of these saved groups. They do not estimate unseen behavior, model
accuracy or savings. The [adoption plan](adoption-plan.ko.md) explains which
real workflow observations would make this useful to another team.

On the original saved0.6.26 observations, ten input groups align. Five outputs
and their expectations change together (17,31,32,33 and9007199254740993); five
remain matching. The report contains five `REQUIREMENT_CHANGED`, five
`STILL_MATCHING` and no regression under an unchanged expectation. This is a
read-only recount of those saved observations, not a repeated model experiment.

## How alignment works

Each key contains the **whole external caller input tuple** and stable activity
ID. Every externally supplied port of every root participates. Bound producer
values are intermediate observations, so they do not replace the original
caller input key. Reordering cases, graph deliveries, ports or JSON record keys
does not create a false addition. Integers remain exact `int64`, including the
one-unit difference above2^53. Strings, booleans, arrays, records and explicit
null retain their JSON types. Floating-point runtime values are unsupported.

An activity or caller tuple found on only one side is `BEFORE_ONLY` or
`AFTER_ONLY`. Renamed stable IDs represent different activities. The tool does
not guess equivalent identities or unseen inputs. Conflicting duplicate
observations for the same tuple/activity are `AMBIGUOUS`; all occurrences remain
in JSON. Identical duplicates form one group with their original case indices.

| Assessment | Meaning |
| --- | --- |
| `REQUIREMENT_CHANGED` | Both expectations exist and differ. Each side keeps its own pass state. |
| `REGRESSION` | The same expectation matched before and fails after. |
| `IMPROVEMENT` | The same expectation failed before and matches after. |
| `STILL_MATCHING` / `STILL_FAILING` | Both observations retain their outcome against the same expectation. |
| `UNOBSERVED` | A side, expected value or actual observation is missing. |
| `AMBIGUOUS` | Repeated observations disagree; no arbitrary row is chosen. |

`outcome_change` separately reports actual value/fault/blocked-state changes.
An expected value of JSON null is an observation; an absent expectation is
unobserved. A recorded arithmetic fault or blocked dependency is a known failed
outcome against a supplied expectation. An absent output stays unobserved.
Fault comparison uses its kind, expression ID, operator and integer operands;
full source-site details remain in each original snapshot.

## Records and bounds

Accepted inputs are full `body-construct` stdout, its saved `evaluation.json`,
full `body-compose` stdout, or bare composition runtime v1/v2/v3. Only the latest
evaluation is compared. Construction attempts and runtime history are preserved
by their original files and are not counted as new execution here. Each input
is bounded to32MiB and8192 observed deliveries. Duplicate JSON fields, ambiguous
envelopes, inconsistent pass flags and complete-runtime totals are errors.

The report includes the before/after input digests, producer/source/suite
identities and observed runtime stages. Failed or partial runs keep their stage;
missing observations do not become successful zeroes. Counts refer to unique
activity/input groups, not a whole-program success rate or model accuracy.
This checks the supplied records' consistency; it does not authenticate their
producer, establish unseen behavior, or estimate financial savings.

The report structure is authored in
[`delta.gooo`](../internal/outcomedelta/delta.gooo), then generated into Go and
JSON Schema by Gooo's receipt projection. Exact input matching and record
comparison use Go. Existing `completeness-delta` continues to compare measurement
dimensions with its original scope rules.
