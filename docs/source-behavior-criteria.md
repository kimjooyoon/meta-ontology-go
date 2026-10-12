# Describe acceptable record behavior in Gooo

A refund has several requirements: reverse the amount when approved, record the
approval, and preserve the customer's explanation. An exact expected record is
useful for an example. A named predicate also expresses a relationship between
the original input and the returned result.

`criteria` keeps these relationships with the source-owned alternatives:

```gooo
activity Choose(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "direction" field_value at "0" alternative "0-input" intent "Negate."
 criteria {
  predicate "negated" activity "Negated"
  input "[7]"
  holdout_input "[9]"
 }
 attempts "2"
}
activity Negated(Integer,Box) -> Boolean computes "return input1.value == 0-input0"
```

The [refund example](../examples/source-behavior-criteria/README.md) includes the
entity declarations and three independent predicates. Use a development build
containing this feature.

## What each declaration means

- Each predicate is a separate fixed `computes` activity. Its positional inputs
  are the assembly's original input types, followed by its result type. Its
  result is Boolean. Fixed, acyclic pure helper calls are supported; a predicate
  cannot call an unresolved assembling activity.
- `input` supplies a canonical positional JSON tuple. The candidate produces an
  actual record. Each predicate receives that record alongside the original
  inputs. No expected record is manufactured.
- `holdout_input` supplies a separate tuple evaluated after selection. Its result
  does not change ranking, stopping, or the selected candidate. Held tuples must
  differ from selection, observation-probe and other held tuples.
- `value_case` may coexist with criteria. Its exact-value cases and field counts
  retain their own denominators. Criteria-only programs need no `value_case`.

The current scope is declared record construction and field updates, with 1–8
predicates, up to 15 original parameters, and the existing 128-input combined
budget. There are at most six binary field choices. Integer search and body-fill
profiles keep their existing contracts. `observe { criteria_input "[...]" }` can acquire additional criterion inputs
from source-declared questions. See [criteria questions](source-criteria-questions.md)
for the selection rule, retained obligations and unresolved outcomes.

## Partial results are useful results

Each predicate/input pair records `SATISFIED`, `UNSATISFIED`, or `UNKNOWN`.
Unknown includes a candidate that did not return or a predicate that could not
be evaluated. Later inputs and candidates can still be considered within the
declared budget. Cancellation ends the operation and preserves the attempted
prefix. A candidate must complete at least one input to be emitted when criteria
are its only obligations.

`passed / total` counts satisfied predicate/input pairs. `failed` counts actual
Boolean false results; `unknown` counts unavailable or not-yet-evaluated pairs.
The observed denominator also reports `(passed + failed) / total`. These numbers
describe the supplied predicates on the supplied inputs. The quality of those
predicates and behavior on other inputs require further evidence.

Criteria selection compares whole exact-value matches first, satisfied predicate
pairs next, and matching fields next. Ties retain the prior candidate order.
Full local completion requires all declared exact cases and predicates to pass.
With an explicit Gooo control policy, its `matched`, `total`, and `best` refer to
exact-value case rows plus criteria input rows that satisfy every predicate.
The policy decision records this unit. The detailed predicate counts remain
separate.

## Source, callers and saved evidence

Predicates, inputs, and held inputs survive source formatting, semantic IR,
round trips, and workspace package lowering. `impact` includes criterion
dependencies. The record includes the predicate's stable activity identity,
generated program digest, actual result, and local call observations.

Changing a fixed predicate or its helper invalidates the dependency checkpoint.
Continuation rechecks the already attempted candidates and retains the preceding
source. Editing the assembly's criterion declarations changes its contract and
requires a new construction.

Caller construction retains each activity's predicates when trying explicit
candidates. Caller outputs and local criteria must both pass for completion.
After caller selection, `final_criteria_holdouts` evaluates the final callable
body and reports overlap with actual consumed caller arguments. Constructions
with criteria use `gooo/joint-construction/v10`.

Replay reconstructs candidate, predicate, held, policy and caller evidence. It
checks the actual saved bindings and counts, including absent or altered results.
Saved model observations are carried without a new prediction.

## The small decision model's role

The model can order source-permitted field combinations. Existing model feature
formats remain unchanged: they do not encode predicate bodies or held inputs.
The expanded plan export retains the declarations separately. This feature adds
no training and makes no claim that the model understands the predicates.

Without a model, the same source uses deterministic candidate order. Gooo owns
the choices, obligations, unresolved outcomes and evidence in either path.
