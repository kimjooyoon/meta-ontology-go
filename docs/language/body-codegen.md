# Experimental activity body code generation

An activity can keep intent, typed paths, finite cases and a search budget in an
[`assembling` block](../source-assembly.md). `body-codegen` detects it directly;
context export and native execution read the same source contract.
For source assembly, JSON generation also includes a reusable `gooo_source`.
[`body-realize`](../source-assembly.md#continue-developing-from-the-selected-body)
replays a saved selection into a new directory and preserves the baseline paths
for the next generation, including results with partial finite completeness.

[`body-compose`](../native-body-composition.md) constructs and immediately
executes several activities joined by explicit `bind` declarations. It preserves
all selected checkpoints and records actual intermediate values across Integer,
Boolean and Text bodies. Optional local ranking is retained across integer
assembly activities; saved composition replay needs no model calls.
Ordinary bodies accept 1..16 scalar or declared record inputs in source order. Use
`input` for one input and `input0`, `input1`, ... for several; each parameter is
read-only while local `let` values may be assigned. See the
[multiple-input example and case keys](../native-body-composition.md) and
[record construction and actual value delivery](../native-record-values.md).

### Model-guided record body fill

Gooo source can declare candidate expressions for a record-valued activity
body. The compiler checks each complete assignment against the record shape
and declared `value_case` examples, then emits only a listed assignment. The
default chooser is deterministic. An optional local `--tiny-model` path can
select assignments when a hole exposes at least two distinct supported
operations. The local model sees the source intent and predicts one of its eight
operation labels. Gooo keeps assignments whose focused hole has that operation,
then uses finite training scores to choose within the group. It does not send
record cases or candidate summaries to the model. If no hole offers more than
one supported operation class, generation fails before inference. An optional
`GOOO_LAYA_URL` endpoint remains available for experiments that need an
external chooser. In every path, Gooo typechecks the complete assignments,
scores them on declared cases, emits the selected body and evaluates the finite
record cases again. For example,
[`source-ir-fill-record.gooo.fixture`](../../examples/body-codegen/source-ir-fill-record.gooo.fixture)
uses `Candidate.state` to construct an accepted or rejected `Review`:

```sh
go run ./cmd/gooo body-codegen --json --activity ReviewCandidate \
  examples/body-codegen/source-ir-fill-record.gooo.fixture
```

Without a model endpoint, Gooo deterministically keeps the highest-scoring
declared assignment. To use the own-model path with source-derived record
candidates, try
[`source-ir-fill-record-tiny.gooo.fixture`](../../examples/body-codegen/source-ir-fill-record-tiny.gooo.fixture):

```sh
gooo body-codegen --json --tiny-model /path/to/model.json \
  --activity ReviewCandidate \
  examples/body-codegen/source-ir-fill-record-tiny.gooo.fixture
```

The `model.json` and its `weights.bin` must come from a compatible local TinyGo
bundle. The example gives the model an `and`/`or` choice over two typed record
fields while Gooo derives candidate bodies, checks types and measures two
training rows plus two separate holdout rows whose `queued` state was not used
to derive candidates. The receipt reports the selected
operation, focused hole, model timings, finite accuracy, and the two retained
assignments out of the four declared combinations. A model proposal that scores
below another candidate is still replaced by Gooo's best-scoring choice.
Record body fill supports required single string, boolean and integer fields,
one declared record result, and pure body expressions.

For external model experiments, set `GOOO_LAYA_URL` to a Laya `/v1/systemone`
endpoint. Its request includes the typed Gooo body, candidate summaries and
finite-case scores; it omits raw case objects. Derived expressions can contain
literals observed in those cases because those are the choices being ranked.
The receipt records the proposal, final candidate, each record case result, and
whether Gooo replaced an inferior proposal with the best-scoring candidate.

The source can also derive record candidates from its declared `value_case`s
instead of listing each assignment. Per-hole `record-field-predicate/v1`
enumerates equality and inequality checks over observed scalar inputs;
`record-field-predicate-composition/v1` also enumerates pairwise `&&` and `||`
combinations, preferring predicates that read different fields. It uses no
arbitrary formulas or nested Boolean syntax, and its completeness denominator
includes the full finite pairwise grammar before the per-hole cap.
The versioned `record-field-predicate/v2` adds `<`, `<=`, `>`, and `>=` for
Integer inputs and integer fields while keeping v1's equality-only candidate
set stable. Its comparison thresholds come from observed training values; it
does not invent unobserved midpoints. Multiple comparison holes can be combined
by control flow already declared in `computes`, and the complete assignment
space remains explicit in the receipt. The
[`integer boundary fixture`](../../examples/body-codegen/source-ir-fill-record-integer-boundary.gooo.fixture)
generates a positive-score predicate, then checks separate training and holdout
records.
`record-field-predicate-composition/v2` composes those ordered atoms with
pairwise `&&` and `||`. Its bounded prefix interleaves cross-field predicates
with same-field integer intervals, so a small cap can retain both record
routing and range candidates. The receipt still reports the full pairwise
grammar size and the retained prefix; a generated interval that is not retained
remains outside the search. The
[`integer range fixture`](../../examples/body-codegen/source-ir-fill-record-integer-range.gooo.fixture)
builds a two-sided score condition, typechecks it, replays it, and scores
separate holdout records. Version 1 keeps its original candidate order and
equality-only atoms.
`record-field-relation/v1` adds comparisons between distinct fields on declared
record inputs. It compares only fields with the same scalar type: string and
Boolean fields get equality/inequality, while Integer fields also get ordered
comparisons. Selectors keep their input and field declaration order; the
receipt reports the complete relation grammar and retained prefix. Unlike
literal predicates, these expressions compare caller-supplied values directly.
The [record relation fixture](../../examples/body-codegen/source-ir-fill-record-relations.gooo.fixture)
derives `input0.key == input1.key`, checks four training pairs, and measures two
separate holdouts.
`record-field-relation-composition/v1` combines two distinct relation atoms
with `&&` or `||`, prioritizing pairs that compare four independent fields.
Its finite denominator includes all pairwise combinations and the atomic
relations; `max_expressions` retains a deterministic prefix. The grammar does
not build nested conditions. The
[relation composition fixture](../../examples/body-codegen/source-ir-fill-record-relation-composition.gooo.fixture)
generates key-and-state matching across two records, retains 8 of 16
expressions, and reports the separate training, holdout and assignment coverage.
`record-string-literal/v1`, `record-integer-literal/v1`, and
`record-boolean-literal/v1` draw typed literals from the expected output record.
The assignment cap and omitted search space appear in the same completeness
receipt as integer derivation. Put independent record examples in
`holdout_value_case` rows. They are not used to derive expressions, score
candidates, or form Laya's request; Gooo evaluates the final selected body
against them after selection and reports a separate holdout accuracy and suite
digest. This measures only those withheld rows and does not establish
full-domain behavior. See
[`source-ir-fill-record-derived.gooo.fixture`](../../examples/body-codegen/source-ir-fill-record-derived.gooo.fixture).
Run it without Laya to use the deterministic best-scoring candidate:

```sh
go run ./cmd/gooo body-codegen --json --activity ReviewCandidate \
  examples/body-codegen/source-ir-fill-record-derived.gooo.fixture
```

[`source-ir-fill-record-composed.gooo.fixture`](../../examples/body-codegen/source-ir-fill-record-composed.gooo.fixture)
shows the bounded two-field predicate grammar in a complete record body-fill.

For multiple requests in one process, use
[`gooo body-path-stream`](../native-body-worker.md). It accepts source recipes,
keeps an optional small model loaded and emits each result as soon as it is ready.
Its `response` has the same shape as the JSON output below.

`gooo body-codegen` is the first source-to-body projection experiment. It reads
an activity's existing `computes` string and emits one deterministic Go
function, bound to the activity's stable semantic ID by generated-region
markers:

```sh
go run ./cmd/gooo body-codegen --activity ClampBelowZero examples/body-codegen/main.gooo.fixture
```

For several body statements, a backtick string can hold the code on separate
lines. Backslashes and line endings are literal, and inner text literals keep
their own Go body syntax. Existing double-quoted strings decode escapes as before.

```gooo
activity ClampBelowZero(Integer) -> Integer computes `
let value = input
value = input
let accepted = value >= 0
if accepted {
    return value
} else {
    return 0
}
`
```

The [complete raw-body fixture](../../examples/body-codegen/raw-computes.gooo.fixture)
includes the package, namespace and Integer declaration. Generate it with:

```sh
go run ./cmd/gooo body-codegen --activity ClampBelowZero examples/body-codegen/raw-computes.gooo.fixture
```

A closing backtick ends the outer literal. Use the double-quoted form for body
text containing a backtick. Canonical Gooo formatting writes an escaped quoted
string while preserving the decoded contents, stable IDs and body semantics.
Whitespace remains part of the decoded body, so different layout can change
generated formatting and source digests. Unterminated literals and invalid UTF-8
retain source diagnostics and cannot become accepted body generation.

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

The pure body profile accepts 1..16 `Integer`, `Boolean`, `Text` or declared
required-string record inputs and one scalar/record result, local `let` declarations,
assignment to an existing local, `if/else`, and one-value `return`. `Integer`,
`Boolean`, and `Text` lower to Go `int64`, `bool`, and `string`. Conditions and
expressions are checked by Go's type checker after a closed syntax filter.
An inferred local initialized from an integer constant uses `int64`, keeping
Integer locals aligned with the DSL type. This rule applies at local bindings;
it adds no conversions at activity input or output boundaries. Boolean and
Text locals retain Go's `bool` and `string` inference.
Function calls, imports, loops and external effects fail
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

For Gooo-source-declared IR candidate search, `gooo body-search-run` connects
candidate selection directly to native execution:

```sh
gooo body-search-run \
  --source examples/body-codegen/ir-search-source.gooo.fixture \
  --activity ClampNegativeToZero \
  --cases examples/body-codegen/ir-search-runtime-cases.json
```

The command finishes one generation request before compiling the selected Go
program, then executes it twice against the independent runtime cases. It emits
one JSON object containing the generation and runtime receipts. Laya is optional;
without it, Gooo uses its stable candidate order. Selection training cases,
withheld holdout cases and independent runtime cases retain separate scores, so
the result shows which part of the pipeline was actually exercised.

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

Optional `--path-observation observation.json` asks which supplied input best
separates the candidates. A separately declared pure Gooo activity can provide
the expected result, extending the finite contract before search. The original
cases and each additional observation remain separately recorded. See the
[observation loop](../path-observation-loop.md) for an executable example, budgets
and unresolved states. This path adds no training and works with a model disabled.

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

## Filling typed body IR holes

The separate `--fill-plan` experiment lets Laya select one listed expression or
complete assignment to fill typed holes in a Gooo-authored body skeleton. The activity keeps the control
flow and a typed hole; the plan supplies a natural-language intent, a finite
list of expression candidates, and explicit input/output cases:

```sh
GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone \
  go run ./cmd/gooo body-codegen --json --fill-plan \
  examples/body-codegen/ir-fill-clamp-plan.json \
  --activity ClampNegativeToZero \
  examples/body-codegen/ir-fill-clamp.gooo.fixture
```

The v1 plan accepts one `Integer -> Integer` hole named
`__GOOO_BODY_HOLE_<hole_id>__`. The v2 plan accepts two to eight holes and
declares two to sixteen complete assignments. Every assignment must fill every
hole exactly once. A single decision selects the whole assignment, which lets
the model consider how expression choices fit together while keeping each piece
explicit in Gooo's IR skeleton. For example, `seed` and `step` can be selected
together from the
[multi-hole fixture](../../examples/body-codegen/ir-fill-multi-hole.gooo.fixture)
and [plan](../../examples/body-codegen/ir-fill-multi-hole-plan.json):

```sh
go run ./cmd/gooo body-codegen --json --fill-plan \
  examples/body-codegen/ir-fill-multi-hole-plan.json \
  --activity Lift examples/body-codegen/ir-fill-multi-hole.gooo.fixture
```

Set `GOOO_LAYA_URL` to let Laya rank the declared assignments. Without a model,
Gooo selects deterministically and still scores every assignment. The compact
`--tiny-model` path also accepts v2 plans when at least one hole gives every
complete assignment a distinct supported root operation. Gooo picks the first
such hole in declaration order and records it as `tiny_model_focus_hole`. The
current model predicts one of eight operation labels from the intent; Gooo maps
that label to a complete assignment through the focused hole's expression. It
does not score the remaining holes jointly, so assignments that cannot be
distinguished through any one hole are rejected before inference. Source-derived
grammar candidate sets also remain on Laya or deterministic selection. The model
never invents a fill, and every selected assignment still passes Gooo's type and
finite-case checks.

Both plan versions support activities with one to sixteen `Integer` inputs and
one `Integer` output. Unary cases retain the `input` field. Multi-input cases
provide a positional `inputs` array matching the activity's `input0`, `input1`,
... parameters:

```sh
GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone \
  go run ./cmd/gooo body-codegen --json --fill-plan \
  examples/body-codegen/multi-input-fill-plan.json \
  --activity Combine examples/body-codegen/multi-input-fill.gooo.fixture
```

This fixture asks the model to choose between sum and difference for two inputs.
Gooo scores both candidates on the supplied vectors before the model call,
records the input types and vector probes in the decision context, then checks
the emitted body against training and holdout vectors. Probe vectors perturb one
input at a time and exclude all declared holdout vectors. Probe disagreement
measures whether alternatives differ on those constructed inputs; it does not
measure intent correctness. Without a configured model, candidate order remains
deterministic.

All plans require closed expression candidates and one to 4096 training cases,
with at most 4096 holdout cases. Each case provides either an integer `input` or
an integer `inputs` array, plus integer `expected`; when both input forms are
present, `input` must equal the first array value. Intent text is limited to
2000 Unicode characters. Gooo typechecks every candidate and
evaluates it with a closed, side-effect-free integer AST interpreter before
asking Laya. Laya receives the Gooo
body IR skeleton, the intent, candidate assignments, and each candidate's
measured test score. It can return only a listed candidate ID. Gooo then fills
all declared IR holes in the selected assignment, emits the final Go function,
typechecks and replays it, and reports the emitted body's exact pass fraction on
the same suite. If the
proposal scores below another declared candidate, Gooo emits the highest-scoring
candidate instead; an equal-scoring Laya proposal is retained. The receipt keeps
both the model proposal and the actual emitted candidate, so the deterministic
test gate cannot hide a poor model selection.
The v1 fixture places its hole inside a conditional assignment to a local `let`
binding, then returns that value, exercising condition, assignment and return.
The v2 fixture fills two local `let` expressions before returning their sum.
The model receives candidate score summaries, the training count and digest,
and per-candidate output profiles on up to 128 synthetic integer probes derived
from training inputs. Declared holdout inputs are excluded from those probes;
the chooser receives neither holdout rows nor expected probe outputs. The profile
lets Laya reason about candidate behavior alongside Gooo intent without treating
any synthetic output as a correctness oracle. Individual training and holdout
input/output rows stay in Gooo's local receipt.

See the [2026-10-06 local Laya observation](laya-local-observation-2026-10-06.md)
for one measured run, including latency, deterministic fallback, and the limits
of the CPU observation.

The `functional_accuracy_percent` field means passed cases divided by declared
cases. It is an exact score for that finite suite under the named bounded AST
interpreter; generated Go is typechecked but not executed by this command. The
score does not claim whole-domain correctness. `candidate_scores` reports every
candidate's score; `best_candidate_accuracy_percent` records the best score
available in this candidate set, and `selection_regret_percentage_points`
measures how far the proposal falls below it. Those values separate
candidate-set coverage from Laya's selection quality on the declared suite.
For v2, the receipt records each `hole_id` and selected expression in
`hole_fills`; `selected_candidate_id` continues to identify the complete
assignment. The completeness receipt adds `declared_suite_functional_accuracy` to the
body-fill core and binds its plan identity to the fill-plan digest. Known
failing cases keep completeness at `PROGRESS`, even when emitted code compiles.
Reports identify the repaired evaluator as `gooo/bodycodegen-int64-ast-interpreter/v2`.
The checked-in plan includes `int64` minimum and maximum values as well as
inputs on both sides of zero.

Source-owned multi-hole fills may also declare `holdout_case` rows. Their inputs
must be disjoint from training inputs. Candidate scores and the Laya request use
training cases only; after selection, the final body is checked against the held-out
cases and the receipt adds `body_fill_holdout_accuracy` as a separate completeness
dimension. The score is exact for that held-out set and is not a proof over the full
integer domain. Holdout input/output values stay in the local report; the chooser
receives neither those values nor a holdout score.

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

### Letting Gooo construct the finite search space

The `candidate_generation` option replaces the hand-written candidate list
with a small, named Gooo expression grammar derived from training examples:

```sh
GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone \
  go run ./cmd/gooo body-codegen --json --fill-search \
  examples/body-codegen/ir-search-generated-candidates-plan.json \
  --activity ClampNegativeToZero \
  examples/body-codegen/ir-fill-clamp.gooo.fixture
```

For `integer-offset-constant/v1`, Gooo constructs deduplicated choices from
the input, expected training constants, negated input, and input-plus/minus
offsets observed in the training examples. `max_candidates` bounds the set to
2..16 expressions. Laya may select only one of these typed IR expressions;
Gooo checks and scores each selected expression before another model call.
Without Laya, the same candidates are tried in their deterministic generated
order.

The receipt hashes the generated candidate set and reports how many expressions
the finite grammar produced, retained, or omitted. `grammar_coverage_percent`
measures coverage of this named grammar after applying the declared cap. It is
not a percentage of user intent, all possible Gooo expressions, or correctness
over the integer domain. Training examples shape the candidate set; holdout
examples remain unavailable to Laya and are evaluated only after selection.
This reduces manual candidate authoring while keeping code authority with the
Gooo checker and evaluator. It does not infer a general-purpose program from
natural language.

The same grammar can be declared inside a Gooo `assembling` block, binding the
intent, training cases, optional withheld holdout, and search budget to the
source's semantic identity. See [source-declared IR search](../source-assembly.md#declare-a-typed-ir-search-in-the-gooo-source)
and the runnable [Gooo fixture](../../examples/body-codegen/ir-search-source.gooo.fixture).
