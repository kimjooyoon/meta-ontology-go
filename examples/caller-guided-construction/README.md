# Reconsider a helper using its caller's result

The helper's example `0 -> {value:0}` fits both `input` and `input * 2`.
The caller expects `3 -> 6`. `body-construct` tries permitted combinations,
immediately compiles each whole program, and uses actual caller failures to
continue. The helper's original obligation stays visible in every attempt.

From the compiler repository, using this development source:

```sh
go run ./cmd/gooo body-construct \
  --source examples/caller-guided-construction/main.gooo.fixture --entry Main \
  --construction-cases examples/caller-guided-construction/construction-cases.json \
  --cases examples/caller-guided-construction/evaluation-cases.json \
  --attempts 2 --out /tmp/my-caller-construction
```

Use a new output directory. The first whole program returns 3. The second keeps
the helper's `1/1` and returns the caller's 6. After selection, four evaluation
rows check positive, negative and large exact-int64 values. One root input was
already consumed during construction; three are different caller input tuples.
These counts do not establish model-training independence or general correctness.

Replay without a model:

```sh
go run ./cmd/gooo body-construct \
  --source /tmp/my-caller-construction/original.gooo \
  --construction /tmp/my-caller-construction/construction.json \
  --cases examples/caller-guided-construction/evaluation-cases.json
```

Saved replay reconstructs the source choices, rechecks local obligations and
re-executes **every recorded whole-program attempt** before the supplied cases.
Fresh construction evaluates its selected program directly, without repeating
its completed search. Neither evaluation nor saved replay makes a prediction.
The selected ordinary Gooo source and generated Go files are also saved.

## Bounds and partial results

The route supports 1..16 record-choice bodies in one typed composition, including
called helpers and multiple interacting bodies. Source-fill and IR-search bodies
use their existing commands. There are at most 64 whole-program attempts and a
three-minute construction deadline. The source's `attempts` bounds the eligible
prefix of each activity's initial ranking. The CLI budget bounds combinations
of those prefixes. Initial local attempts and later local rechecks are separate
observations; repeated checks are not new candidate identities.

The initial locally selected combination is tried first. Remaining combinations
follow the Cartesian order of the retained per-activity rankings. This crosses
plateaus where changing only one of two helpers cannot improve the caller.
No model means deterministic ordering. An explicit `--model` can rank during
initial construction; caller failures subsequently advance this bounded order.
There is no new model call for each failure in this route.

Selection prefers a combination satisfying all local obligations, then more
caller expectations, then more local expectations, keeping the first exact tie.
`COMPLETE_FINITE` requires every local and supplied caller obligation to match.
Insufficient budgets or contradictory expectations retain `PARTIAL_FINITE`, the
best observed program and the full attempt history. Compilation/execution errors
stop with an explicit failure. Logical partial results still return exit code 0;
inspect `construction.decision` and the separate evaluation counts.

This command does not yet perform package-wide joint search or apply an external
Gooo assembly policy to joint combinations. Existing package continuation and
local Gooo policies retain their separate semantics.

## Three-choice model observation

`model.gooo.fixture` exposes the three choices expected by the current graph
model. It constructs two terms and a condition in a helper; the caller returns
their sum for positive inputs. Use the `model-` case files, `--attempts 8`, and
optionally the independently trained graph QAT model. Those files are fixed
before dogfooding. Compare the initial prediction, complete program attempts,
local and caller scores separately. The seven evaluation rows include one
consumed construction input and six other root tuples; zero is also a local
helper example. The evaluation score cannot be labeled wholly unseen.

The [frozen observation and raw records](../../docs/research/caller-guided-construction-20261008)
compare fixed and model ordering, include both partial and complete outcomes,
and preserve a model-free saved replay. These are local development observations;
the recorded compiler revision identifies the implementation used.
