# Continue constructing a program with called helpers

Build from this source revision with Go 1.27.1. This example requires the
dependency-aware continuation added after `v0.6.6-dev`.

`Seed` adds one; `Wrap` doubles that result; `Main` calls `Wrap` twice. Each
helper declares two candidate bodies and two finite expectations. A Gooo policy
first saves an incomplete construction, then another Gooo policy continues it:

```sh
go build -o /tmp/gooo-dependent ./cmd/gooo
/tmp/gooo-dependent body-compose \
  --source examples/dependent-continuation/main.gooo.fixture \
  --cases examples/dependent-continuation/cases.json --entry Main \
  --assembly-policy examples/assembly-policy/checkpoint.gooo.fixture \
  --policy-activity Checkpoint --out /tmp/gooo-dependent-first

/tmp/gooo-dependent body-compose \
  --source examples/dependent-continuation/main.gooo.fixture \
  --cases examples/dependent-continuation/cases.json \
  --resume-composition /tmp/gooo-dependent-first/composition.json \
  --assembly-policy examples/assembly-explainer/main.gooo.fixture \
  --policy-activity Explain --out /tmp/gooo-dependent-next

/tmp/gooo-dependent body-compose \
  --source examples/dependent-continuation/main.gooo.fixture \
  --cases examples/dependent-continuation/cases.json \
  --composition /tmp/gooo-dependent-next/composition.json
```

Use new output directories. The saved composition supplies the entry activity
during continuation and replay. All three commands retain the original source.

The first run keeps one attempt for each helper. Continuing selects `Seed`'s
incremented body before evaluating `Wrap`. That change makes `Wrap`'s previous
scores stale: its attempted mask is evaluated again against the new `Seed`, then
the remaining candidate is considered. The expected native result moves from
0/2 to 2/2. Each helper's construction cases remain a separate 2-case measurement.

## What the record means

| Field | Meaning |
| --- | --- |
| `retained_attempts` | Candidate identities attempted before this continuation |
| `rechecked_attempts` | Retained candidates scored under changed callable dependencies |
| `added_attempts` | Previously unattempted candidates explored in this continuation |
| `new_model_calls` | New predictions during continuation; always zero in this route |

The cumulative attempt budget counts distinct attempted masks. Rechecks consume
computation but do not create new masks or reset that budget. Historical receipt
validation also evaluates candidates; these counters are not total CPU work.
When a callable dependency changes, every retained mask of the affected activity
is rechecked. Fixed helpers used only by alternatives are included. An unrelated
helper change preserves the score prefix.

`control_history` retains the original policy boundary and changed source
checkpoints. Replay reconstructs those stages with their original dependencies,
then verifies current scores. Historical model context and ranking keep their
original source identity. Comparing a saved stage with a newly constructed
helper is therefore explicit, even if both stages selected the same mask.

## Scope and next use

The supported route is `body-compose` with record-choice preparations and
record-choice graph activities. Ordinary typed callers are projected again.
Nested helpers and a helper used as both a bound producer and a function are
supported. Histories are bounded to 16 saved stages, with 128 KiB per source.
Each round is checked against the helper bodies that round actually constructed.

Source-fill/search continuation and package-workspace continuation remain open
integration work. This command is explicitly invoked; it does not schedule its
own next run. Nested calls made during candidate scoring do not yet contribute
input-history observations, so the runtime report retains that input-separation
limitation. Finite matches describe the named cases.

### 한국어

조립 중인 작은 부품을 바꾸면 그 부품을 쓰는 다음 단계도 다시 확인해야 합니다.
이 예제는 보조 함수를 먼저 이어 만들고, 영향을 받은 호출자의 기존 후보를
새 부품에 맞춰 재검사합니다. 과거 점수는 당시 소스와 함께 재현하고, 현재 점수는
현재 소스에서 계산합니다. 모델의 기존 제안 순서를 활용하면서 새 추론 없이
남은 조립을 진행하는 방식입니다.
