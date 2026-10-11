# Continue after a record candidate fails

`Build` first tries a field value containing `Step(input) / input`. Its local
case for zero fails after actually calling `Step(0)`. The declared alternative
uses `Step(input)` directly and completes both local cases. `Main` then consumes
the assembled record.

Gooo 0.6.30 retains the first attempt as `EVALUATION_FAILED`, consumes
one of the two source attempts, and proceeds to the next candidate. Its
incomplete suite has no accuracy denominator. The actual helper calls reached
before failure remain in `record_assembly.attempts[].local_calls`; they are
compared during saved replay. Cancellation still stops the request.

From a compiler checkout, run with Gooo 0.6.30 or later:

```sh
gooo body-construct --source examples/caller-record-failures/main.gooo.fixture \
  --entry Main \
  --construction-cases examples/caller-record-failures/construction-cases.json \
  --cases examples/caller-record-failures/evaluation-cases.json \
  --attempts 2 --out /tmp/gooo-record-recovery --format json
```

The construction case is `Main(3) = 4`. The separate finite evaluation cases are
`Main(-1) = 0`, `Main(4) = 5` and `Main(0) = 1`. These examples describe a bounded
contract; passing them does not establish correctness for all integer inputs.
No model is requested. See the [0.6.30 release guide](../../docs/releases/0.6.30-dev.md)
for installation and the release evidence.

When the source budget is reduced to one, construction stops with
`initial.record_failure` containing the attempted prefix. It emits no native
candidate and no finite execution score. A caller-selected failing record mask
likewise consumes a program attempt and retains a `LOCAL_RECORD_ASSEMBLY`
rejection. Successful alternatives remain eligible within the original budget.

## Original development observation

[The saved observation](observed-recovery-3a1628ed.json) binds clean compiler
source `3a1628ed413fbbaa9d9168acdee2b9fb13f4749a`. One construction invocation
recorded `Step(2)` and `Step(0)` before mask 0 failed, then selected mask 1 after
it completed both local cases. Native evaluation and one saved-JSON replay each
passed 3/3 supplied cases, with zero model calls. The earlier producer stopped
before native execution; it has no comparable native accuracy denominator.
