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

The v1 body profile accepts one `Integer`, `Boolean`, or `Text` input and a
supported `Integer`, `Boolean`, or `Text` result, local `let` declarations,
assignment to an existing local, `if/else`, and one-value `return`. `Integer`,
`Boolean`, and `Text` lower to Go `int64`, `bool`, and `string`. Conditions and
expressions are checked by Go's type checker after a closed syntax filter.
An inferred local initialized from an integer constant uses `int64`, keeping
Integer locals aligned with the DSL type. This rule applies at local bindings;
it adds no conversions at activity input or output boundaries. Boolean and
Text locals retain Go's `bool` and `string` inference.
Function calls, imports, loops, multiple inputs, and external effects fail
closed. The generated result is written to stdout; this command does not
mutate the repository.

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

The report also carries a
`gooo/metaprogramming-completeness-receipt/v2` record. It keeps declaration,
generation, source-AST, route-equivalence, typecheck, replay, route-protocol,
provenance, and write-boundary evidence separate from UNKNOWN dimensions such
as runtime execution, reverse observation, real workflows, full-domain
semantics, route quality, and comparable before/after baselines. Each open
dimension retains its reason and next operation; `first_unresolved` identifies
the earliest open stage. The aggregate score remains null. Compiler source
identity comes from clean Go build metadata; modified or metadata-free builds
retain an incomplete provenance status. Error JSON also retains a fail-closed
receipt and its cause.

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
AST bodies. Method `canonical_control_flow_form/v2` removes only outer
parentheses around each complete condition or return expression before comparing
these forms, because Go formatting may remove those wrappers. Inner grouping,
operators, operand order, and branch values remain part of the comparison.
A mismatch fails closed before output. This is a structural
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

## Source-bound structural paths with a local Go model

`--path-plan` is a separate local experiment for composing typed structural
decisions, including local references, assignment targets, operand order,
then/else layout, and statement order. A document declares the fallback body,
two options per decision, explicit integer cases, and a finite attempt budget.

```sh
go run ./cmd/gooo body-codegen --json --path-plan \
  examples/body-codegen/typed-path-compound-plan.json \
  --activity Combined examples/body-codegen/typed-path-compound.gooo.fixture
```

Add `--path-model /path/to/model.json` to load a downloaded
`gooo/tiny-path-decision-model/v1` bundle, such as the public
[Gooo typed path models](https://huggingface.co/asketeddy/gooo-typed-path-tiny-v1).
The operation-model ABI used by `--tiny-model` is deliberately rejected here.
No model is downloaded implicitly. Runtime and orchestration are Go; this path
makes zero external provider calls even if Laya environment settings exist.

The compiler validates the fallback against the original activity body using
its existing canonical equivalence witness before loading the model. The
activity must have one Integer input and Integer result. Edits replace only
the in-memory `computes` literal and preserve the source package, unrelated
declarations, and stable semantic identity; repository writes remain zero.

The typed plan is prepared once into an owned immutable snapshot. Its already
checked fallback is used for source binding, and search reuses the same snapshot.
Every combined candidate still passes the existing arena type/scope compiler;
there is no global cache or reuse of earlier test outcomes. The receipt records
`plan_prepare_ms` separately. Earlier receipts included preparation in source
binding/search costs, so compare total processing before comparing stage costs.

`typed-path-conditional-assignment.gooo.fixture` and its English/Korean plan
documents exercise six interacting decisions across comparison operands,
Boolean local references and assignment targets, Integer assignment targets,
branch bodies and Boolean-update order. Their 64 combinations use the same
structural ABI. Seven explicit cases select behavior; they do not prove every
natural-language structural request. An inconsistent finite case retains a
6/7 functional result while native lowering remains complete.

One synchronous prediction per decision ranks its two typed options before
candidate tests. Conditional probabilities are ranking weights, not calibrated
odds that a request is fulfilled. The bounded search checks combined scope and
types, evaluates finite cases, and retains the best observed partial candidate.
The selected body then passes native Gooo emission, equivalence, type checking,
replay, and an independent native AST evaluator compared with the arena results.
There is no background codegen goroutine waiting for a model response.

`body_paths` records original/selected source, plan, test suite, model metadata
and weight hashes; actual predictions; candidate attempts, type rejections and
unattempted alternatives; per-stage timings; and finite functional completeness.
Native lowering `PASS` can coexist with functional `PARTIAL`: 100% lowering
means all selected source constructs were emitted, while e.g. 2/3 finite cases
means 66.67% measured functional completeness. Neither is full natural-language
or all-input correctness. Default ranking does not receive test cases.
Optional `--path-feedback-rounds` can pass finite outcome counts and the first
observed mismatch to the same frozen model between candidate batches. It requires
`--path-model` and `--path-step-attempts`; a bounded `--path-feedback-ci` file can
provide caller CI context without edit or merge authority. See
[incremental typed paths](../typed-path-sessions.md) for budgets and receipts.

Bounds are 16 binary decisions, 128 expression and statement arena nodes, 128
cases, 64 attempted combinations, 128 KiB source/document, and a cooperative
eight-second operation deadline. Without a model, enumeration starts from the
declared fallback in deterministic order; it can still change that fallback in
response to the explicit finite contract. A document `seed` is optional with a
model and enables a reproducible probability draw. Failure JSON retains the
structural receipt where available. These flags cannot combine with route,
fill-plan, fill-search, or operator-model modes.

The compound fixture has three decisions and eight combinations, including
two combined scope failures. Its expected arithmetic is `5*input+2`, subtracting
3 for negative inputs and adding 3 otherwise. The original fallback deliberately
differs, so the example measures structural assembly and finite feedback rather
than asserting semantic equivalence between the original and requested edit.

## Filling a typed body IR hole

The separate `--fill-plan` experiment lets Laya select one listed expression to
fill a hole in a Gooo-authored body skeleton. The activity keeps the control
flow and a typed hole; the plan supplies a natural-language intent, a finite
list of expression candidates, and explicit input/output cases:

```sh
GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone \
  go run ./cmd/gooo body-codegen --json --fill-plan \
  examples/body-codegen/ir-fill-clamp-plan.json \
  --activity ClampNegativeToZero \
  examples/body-codegen/ir-fill-clamp.gooo.fixture
```

This first slice accepts one `Integer -> Integer` hole named
`__GOOO_BODY_HOLE_<hole_id>__`, two to sixteen closed expression candidates,
and one to 4096 integer test cases. Intent text is limited to 2000 Unicode
characters. Every JSON test case must explicitly provide non-null integer
`input` and `expected` fields. Gooo typechecks every candidate and
evaluates it with a closed, side-effect-free integer AST interpreter before
asking Laya. Laya receives the Gooo
body IR skeleton, the intent, candidate expressions, and each candidate's
measured test score. It can return only a listed candidate ID. Gooo then fills
that IR hole, emits the final Go function, typechecks and replays it, and
reports the emitted expression's exact pass fraction on the same suite. If the
proposal scores below another declared candidate, Gooo emits the highest-scoring
candidate instead; an equal-scoring Laya proposal is retained. The receipt keeps
both the model proposal and the actual emitted candidate, so the deterministic
test gate cannot hide a poor model selection.
This fixture places the hole inside a conditional assignment to a local `let`
binding, then returns that value, so it exercises condition, assignment, and
return paths together. The model receives candidate score summaries plus the test count and digest;
the individual input/output cases stay in Gooo's local receipt.

The `functional_accuracy_percent` field means passed cases divided by declared
cases. It is an exact score for that finite suite under the named bounded AST
interpreter; generated Go is typechecked but not executed by this command. The
score does not claim whole-domain correctness. `candidate_scores` reports every
candidate's score; `best_candidate_accuracy_percent` records the best score
available in this candidate set, and `selection_regret_percentage_points`
measures how far the proposal falls below it. Those values separate
candidate-set coverage from Laya's selection quality on the declared suite.
The completeness receipt adds `declared_suite_functional_accuracy` to the
body-fill core and binds its plan identity to the fill-plan digest. Known
failing cases keep completeness at `PROGRESS`, even when emitted code compiles.
Reports identify the repaired evaluator as `gooo/bodycodegen-int64-ast-interpreter/v2`.
The checked-in plan includes `int64` minimum and maximum values as well as
inputs on both sides of zero.

Candidate expressions retain their grouping when inserted into a larger
expression: filling `2 * HOLE` with `input + 1` emits `2 * (input + 1)`.
The evaluator uses Go typechecker bindings and constant values. Local names
such as `true` or `false` resolve to their declared variables; constant
arithmetic keeps Go's arbitrary precision until conversion, while runtime
integer operations use their declared widths. Boolean and string locals are
also supported inside an integer-input/integer-output body. Unsupported
runtime value types return an error instead of a functional score. CLI
generation failures return a nonzero exit status and a failure receipt.

Regression tests compare interpreter results against independently compiled
Go for constant boundaries, local bindings, integer overflow, and compound
expressions. That comparison tests the evaluator's fidelity on those cases;
it does not turn the declared user suite into a full-domain proof.

This path is deliberately sequential: Gooo creates and scores the typed IR
plan, then makes one synchronous Laya call, then fills and emits the final
body. No body-emission goroutines run beside the model request. The request
inherits cancellation and has an eight-second decision budget for the larger
IR-and-test context; a timeout or unavailable provider records deterministic
fallback, then the same score gate picks the best declared candidate. The report records IR preparation,
decision, final-emission, and total wall latency. Tests cover the blocking
call order, provider cancellation, exact functional percentages, and disconnected
replay. The eight seconds bounds provider work only; parsing, candidate scoring,
and final emission do not have an operation-wide deadline. An expired provider
request still allows deterministic fallback and emission to finish.
This bounded evaluator accepts only the existing pure expression and
statement profile; it does not run arbitrary model-authored source.

The evidence sent in this call is candidate-level TDD data computed before
emission, not a GitHub Actions result. CI runs after a generated revision
exists, so it can verify this output and inform a later generation round, but
it cannot be an outcome of the same decision that produced the revision.

## Sequential before-score search

The experimental `--fill-search` mode uses a separate plan schema to let Laya
propose one declared expression candidate at a time before Gooo scores it:

```sh
GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone \
  go run ./cmd/gooo body-codegen --json --fill-search \
  examples/body-codegen/ir-search-clamp-plan.json \
  --activity ClampNegativeToZero \
  examples/body-codegen/ir-fill-clamp.gooo.fixture
```

The `gooo/body-codegen-ir-search-plan/v1` plan names the intent and typed IR
hole, declares an ordered candidate list, provides `test_cases` for training,
may provide `holdout_test_cases`, and must set `max_attempts` from one through
the number of candidates. This is still a single `int64 -> int64` expression
hole in a Gooo-authored body skeleton. It does not let the model author
arbitrary code or control flow.

On each attempt, the model proposes from candidates that have not yet been
tried. Gooo scores only that selected candidate against the training cases. A
perfect training candidate ends the search; otherwise the next model round
receives the prior candidate's training failure before it proposes again. The
search stops at the first perfect training candidate or after
`max_attempts`. If it does not find a perfect candidate, it emits the best
observed candidate; ties follow the declared candidate order. Without a
provider, candidates are tried in declared order. When only one candidate
remains, Gooo selects it without another model call.

Each attempt appears in the receipt. `attempted_candidates` counts every
candidate tried, while `evaluated_candidates` counts only completed training
scores. A candidate rejected before scoring has `scoring_completed: false` and
a null `accuracy_percent`, rather than a misleading zero. Feedback for the
next model round is bounded: it includes at most the first eight failed
training cases, along with `failed_cases_total` and `failed_cases_truncated`.
The local receipt retains the complete case results.

`provider_operations` counts resolver calls made with a nonblank endpoint,
including a call canceled before a candidate attempt is recorded. It excludes
sole-candidate selection and fallback after the shared provider budget is spent.
This is not an HTTP request count: a resolver operation can also request provider
health metadata. Whitespace-only endpoints behave as an unconfigured provider.
The external-network completeness dimension uses this counter separately from
completed candidate attempts.

All provider rounds share one eight-second budget accumulated only during
provider decision calls; `provider_budget_used_ms` reports the time charged to
those calls. Plan parsing, candidate checking and scoring, the final training
rescore, holdout evaluation, and final emission do not consume that budget.
There is no end-to-end generation deadline. After the final candidate is
chosen and emitted, Gooo rescores the emitted source against the training
suite; receipt training metrics and `training_case_results` describe that
final source. The
holdout suite is evaluated only after final selection; its cases are never
sent to the model and never affect search arbitration. The receipt keeps
training and holdout results separate and reports
`global_best_accuracy_percent` as null (`UNKNOWN`): candidates outside the
declared finite set and behavior outside the supplied cases are not
established. These scores come from the bounded integer AST interpreter;
generated Go is typechecked and replayed but not executed by this CLI. A
runtime functional metric requires a separate compiled oracle. The
deterministic offline fixture includes bad identity and negation candidates
before the correct zero expression, with `int64` extrema held out from
training.

This experiment establishes only exact training and holdout results on the
declared cases under Gooo's bounded integer interpreter. It does not establish
whole-domain correctness, model quality in general, or a speed benefit over
the score-first `--fill-plan` experiment.

The checked-in `.gooo.fixture` intentionally uses a finite, side-effect-free
body. The suffix keeps this experimental input outside the repository's fixed
`.gooo` conformance inventory until the language corpus itself is revised.
Later experiments can compare behavior across codegen routes without granting
the model authority to produce code.

### Reusing source-bound external training observations

An `ir-search` plan may include an `external_training_feedback` object when a
separate run produced actual results for the exact same DSL bytes and typed
training cases. The Go API is `IRBodySearchPlan.ExternalTrainingFeedback
*ExternalTrainingFeedback`; the JSON form is:

```json
{
  "source_digest": "sha256:<lowercase hex of the original .gooo bytes>",
  "training_suite_sha256": "sha256:<lowercase hex of canonical typed training cases>",
  "candidate_id": "identity",
  "observations": [
    { "input": -2, "expected": 0, "actual": -2, "passed": false },
    { "input": 1, "expected": 1, "actual": 1, "passed": true }
  ]
}
```

`source_digest` binds the exact original DSL byte sequence using the same
`sha256:<hex>` notation as receipts. `training_suite_sha256` binds Go's
canonical `encoding/json` encoding of the typed `[]IRBodyFillTestCase` in
`test_cases`. An observation must name a declared training input and its exact
expected value; duplicate inputs, holdout rows, stale digests, unknown
candidates, and a `passed` value that disagrees with `actual == expected` are
rejected before any provider call. There must be 1 through 4096 observations,
and each observation must explicitly supply non-null `input`, `expected`,
`actual`, and `passed` fields. The observation list contains measured values,
not an attestation that a particular CI system produced them.

External observations are advisory prior feedback. They are not authenticated
CI proof, a semantic authority, or a score for this search invocation. Gooo
does not score a candidate before asking the chooser; after a candidate is
selected, Gooo typechecks and measures it against the current declared training
cases as usual. Holdout cases cannot be supplied as external observations and
remain withheld from the chooser. The chooser receives only the declared
candidate ID and up to eight failing `{input, expected, actual}` triples, with
`failed_cases_total` and `failed_cases_truncated`. The local receipt retains the
feedback hash, both digest bindings, candidate ID, and observation/failure
counts. With no provider endpoint, declared candidate order remains the
deterministic selection rule even when feedback is attached.

An optional `prompt_profile` selects the request packaging. Omission or an
empty value keeps the legacy package, including its `training_suite_sha256`
state field. The value `compact` omits that field whether or not external
feedback is attached; compact requests use the same concise chooser
instructions in both cases, so a comparison of compact prompts isolates the
presence of external failure observations from packaging and wording changes.
Compact model state contains no external digest strings. Other profile values
are rejected before body selection or a provider call. The receipt echoes a
non-empty `prompt_profile` for local provenance.

The optional plan field `provider_model` is passed through to the typed
decision request. Supported values are `english`, `multilingual`, and
`typed-decisions`; omission or an empty value keeps the provider's automatic
default. Any other value fails plan validation before source-body selection or
a provider call.

The text fixtures in `examples/body-codegen/` cover string equality and
conditional results on the bounded Laya route, plus local assignment on the
deterministic route. This adds scalar text behavior to the closed source-body
profile; it does not add records, collections, multiple inputs, or dynamic
model-authored code.
