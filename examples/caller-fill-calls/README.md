# Fill a body that uses another Gooo activity

This example lets a complete hole assignment call a fixed pure activity from
the same Gooo source. It requires Gooo 0.6.30 or later; see the
[release and installation guide](../../docs/releases/0.6.30-dev.md).

`Step` doubles its argument. `Build` chooses where to put the declared offset
and bias in `Step(input + offset) + bias`. Its local case requires `2 -> 6`;
the assignment `offset=1, bias=0` matches. The compiler includes `Step` while
checking every candidate, comparing behavior, emitting and replaying the body.
The helper can also appear inside a candidate expression. Integer, multiple
argument and declared record helper signatures use the existing pure-call rules.

From a compiler checkout containing this feature, with Go 1.27.2:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o ./gooo ./cmd/gooo
./gooo body-construct --source examples/caller-fill-calls/main.gooo.fixture \
  --entry Main --construction-cases examples/caller-fill-calls/construction-cases.json \
  --cases examples/caller-fill-calls/evaluation-cases.json \
  --attempts 5 --out fill-with-helper --format json > fill-with-helper.json

./gooo body-construct --source fill-with-helper/original.gooo \
  --construction fill-with-helper/construction.json \
  --cases examples/caller-fill-calls/evaluation-cases.json --format text
```

Use a new output directory. Selection is deterministic with no model configured.
Source-owned cases and declared assignments determine the finite scores; the
separate holdout is evaluated after selection. The emitted report binds the
helper identity and body digest. Saved replay reconstructs the same source and
candidate evaluation without calling a model.

Helpers must have fixed pure bodies when an assignment is evaluated. Composition
can assemble required helpers first. Existing type checks, recursion rejection
and call bounds apply. Gooo 0.6.30 records helper arguments during fill preflight
and selection; the original observation below predates that
capture. Other construction boundaries still have incomplete coverage, so
constructed-helper input independence can remain `UNKNOWN`. Finite passing
examples cover only their declared rows.

The [saved synthetic observation](observed-helper-fill-deb93c0a.json) binds clean
development source `deb93c0a` to one original construction and saved replay. It
selected `shifted`, matching 1/1 local training and 1/1 holdout cases. Native
evaluation and saved replay each matched 3/3 supplied rows, with zero model
calls. The report retained `Step`'s source identity and body digest.

Command accounting on macOS ARM64 reported construction plus evaluation at
0.62 seconds wall, 0.28 seconds user CPU, 0.22 seconds system CPU and 86,228,992
bytes maximum RSS. Replay reported 0.35 seconds wall, 0.16 seconds user CPU,
0.11 seconds system CPU and 86,736,896 bytes maximum RSS. These single runs
include compilation and subprocesses; cache state and startup affect the totals.
Instantaneous system CPU utilization was not measured.

The [model-to-helper-fill example](../caller-model-fill/README.md) uses the public
source-relative model to rank an upstream helper's branch choices, then runs the
fill stage deterministically. It records the model's weak initial proposal as
well as the later passing construction.
