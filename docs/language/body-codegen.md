# Experimental activity body code generation

`gooo body-codegen` is the first source-to-body projection experiment. It reads
an activity's existing `computes` string and emits one deterministic Go
function, bound to the activity's stable semantic ID by generated-region
markers:

```sh
go run ./cmd/gooo body-codegen --activity ClampBelowZero examples/body-codegen/main.gooo.fixture
```

For an explicit seeded probability experiment, pass a seed:

```sh
go run ./cmd/gooo body-codegen --json --sample-seed experiment-17 \
  --activity ChooseBranch examples/body-codegen/guard-route.gooo.fixture
```

To let the local Laya service select among eligible routes, set the same
endpoint used by `gooo decide`:

```sh
GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone \
  go run ./cmd/gooo body-codegen --json --activity ChooseBranch \
  examples/body-codegen/guard-route.gooo.fixture
```

The v1 body profile accepts one `Integer` or `Boolean` input and one matching
result, local `let` declarations, assignment to an existing local, `if/else`,
and one-value `return`. Conditions and expressions are checked by Go's type
checker after a closed syntax filter. Function calls, imports, loops, multiple
inputs, and external effects fail closed. The generated result is written to
stdout; this command does not mutate the repository.

The v3 JSON report records the source/program/generated digests, source and
lowered semantic-unit counts, candidate and final route, Laya decision and
route-selection receipts, provider latency, typecheck result, deterministic
replay digest, and completeness over the accepted source-body units. A
semantic unit is a non-container statement or expression
node in the activity body. Completeness is the fraction of those source units represented
by the selected lowering; it measures coverage of this body profile, not the
percentage of a person's unstated intent. A `PASS` also requires that the
closed syntax filter, control-flow termination check, Go typecheck, and replay
all succeed.

When `GOOO_LAYA_URL` points to a Laya `/v1/systemone` endpoint, this command
uses Laya to choose only from routes already proven eligible for that source
shape. For a single pure `if/else` whose two branches each return once, it can
choose among preserving the branch returns, a `guard-return` rewrite, or an
explicit `merge-result` join. The guard rewrite changes
`if c { return a } else { return b }` into `if c { return a }; return b`.
The join route assigns each branch value to a typed local and returns it after
the conditional, making the control-flow merge explicit. Each emitted report
includes a `route_equivalence` receipt with schema, decision, method, rule,
source and generated semantic digests, and the declared proof scope. The
compiler canonicalizes only the direct `if/else`-return, guard-return, and
typed-result-join shapes; other accepted bodies must have matching formatted Go
AST bodies. A mismatch fails closed before output. This is a structural
equivalence witness for the closed pure body profile, not a proof of unstated
intent, full-domain behavior, or which route is clearer. Each route is also
typechecked before output. Laya receives a source digest and structural
summary, not the activity body. By default, an absent,
unavailable, malformed, or over-budget provider uses the declared `preserve`
fallback; the planner has a three-second budget. With no Laya setting and no
sample seed, repeated runs therefore select the same route and produce the
same generated digest.

`--sample-seed` opts into a probability draw. With Laya available, the command
normalizes its per-route probabilities and samples only among the eligible
routes. Without Laya, it samples from an equal prior over those routes. The
receipt records the sampling method, seed digest, normalized weights, draw
value, proposed route, and final emitted route; it never stores the raw seed.
The draw binds the seed to the canonical decision-request digest and ordered
candidate set. Repeating the same source, seed, and recorded weights therefore
replays the same route. If the seed is omitted, the existing Laya top-choice
and deterministic `preserve` fallback behavior remains unchanged. Every
selected route still passes Gooo's closed syntax checks, type checking, and
internal emission replay before it is returned.

The experiment does not change `gooo generate`'s package projection or claim
that the current runtime executes these bodies. The model selects a bounded
construction route; the Gooo emitter, typechecker, and replay check retain
authority over generated Go.

The checked-in `.gooo.fixture` intentionally uses a finite, side-effect-free
body. The suffix keeps this experimental input outside the repository's fixed
`.gooo` conformance inventory until the language corpus itself is revised.
Later experiments can compare behavior across codegen routes without granting
the model authority to produce code.
