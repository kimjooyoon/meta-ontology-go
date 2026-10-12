# Gooo declares its next question

The initial example `11 -> 0` fits both `input - 11` and `11 - input`.
The source gives Gooo two further inputs to inspect and names `ExpectedFee`
as the activity that supplies expected answers. Gooo compares the remaining
candidates on those inputs, selects a distinguishing question, and extends the
finite suite with the declared activity's answer.

```gooo
observe {
    input "17"
    input "-5"
    oracle "ExpectedFee"
    max_candidates "2"
    max_rounds "1"
    reuse_outputs "true"
    resolve_unique "true"
}
```

The reference here is deliberately small and readable. In a larger application
the reference describes the known rule while the assembled activity represents
the implementation under construction. Its answers apply to inspected inputs;
the declaration does not establish equivalence on every integer.

## Run from a compiler checkout containing this feature

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-codegen --json \
  --activity Fee examples/source-owned-observation/main.gooo.fixture

GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-compose \
  --source examples/source-owned-observation/main.gooo.fixture --entry Main \
  --cases examples/source-owned-observation/evaluation.json \
  --out /tmp/gooo-source-owned-observation
```

Both commands work without a model or an external observation JSON file. The
first keeps the initial ambiguity and the acquired answer in its receipt. The
second evaluates four separate native inputs after construction. Its saved
`composition.json` can be replayed with `body-compose --composition`.

Remove the `oracle` line to keep the next question open. The receipt then reports
`ORACLE_UNAVAILABLE` and retains the original suite; no expected answer is
invented. Ordinary finite selection can still match that suite. Read the
observation status alongside the finite score to see the remaining ambiguity.

With `max_candidates "1"`, one alternative is still unobserved. The status is
`CANDIDATE_BUDGET`; `resolve_unique` cannot turn this into a unique solution.

This development feature reuses the existing candidate-observation algorithm.
The language change makes its questions, reference and bounds part of the source
contract, semantic IR and saved replay. It currently supports Integer typed-path
assembly. The path profile does not accept inline `holdout_case`; separate final
evaluation inputs are used here. IR search and record assembly keep their existing
interfaces. Public compiler 0.6.34 does not contain this syntax.

Workspace preparation also retains a local oracle declared in another source file
and rewrites its name with the owning package. Its dependency is recorded as
`observe:oracle` in the existing pure-call inventory.

## When the caller asks for a conflicting result

`caller-conflict.json` asks for `Main(17) = 6`. The source reference already
answered `Fee(17) = -6`, and Main returns Fee directly. Both requirements cannot
hold for either declared candidate.

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-construct \
  --source examples/source-owned-observation/main.gooo.fixture --entry Main \
  --construction-cases examples/source-owned-observation/caller-conflict.json \
  --cases examples/source-owned-observation/evaluation.json --attempts 2 \
  --format text --out /tmp/gooo-source-owned-conflict
```

The expected decision is `PARTIAL_FINITE`. The selected program keeps its two
local answers (2/2) and misses the conflicting caller answer (0/1). The other
candidate meets the caller answer (1/1) but loses the acquired local answer
(1/2). The retained observation belongs to the local contract throughout caller
selection. Its provenance and these separate counts remain in the saved result.

The reference must be a fixed `computes` activity. An activity whose body still
has `assembling` choices cannot supply the expected answers. Fixed helper calls
are supported by the existing pure-call closure.
