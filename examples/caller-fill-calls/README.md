# Fill a body that uses another Gooo activity

This development example lets a complete hole assignment call a fixed pure
activity from the same Gooo source. Public compiler 0.6.29 predates this change.

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
and call bounds apply. Full recording of helper arguments throughout fill
preflight and selection remains work in progress, so constructed-helper input
independence can remain `UNKNOWN`. Finite passing examples cover only their
declared rows.
