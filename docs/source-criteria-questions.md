# Let Gooo choose a useful next input

A refund example with an amount of zero cannot distinguish keeping the amount
from reversing its sign. Both return zero. A passing example therefore leaves
an important choice unresolved.

Gooo can now name possible follow-up inputs in the source. It compares the
permitted candidates' outputs, chooses an input that separates them, and then
applies the source's behavior criteria to that input:

```gooo
criteria {
 predicate "negated" activity "Negated"
 input "[0]"
 holdout_input "[9]"
}
observe {
 criteria_input "[7]"
 criteria_input "[-5]"
 max_candidates "2"
 max_rounds "1"
 reuse_outputs "true"
 resolve_unique "true"
}
```

This fragment belongs inside a record `assembling` declaration. `Negated` is a
fixed Boolean activity receiving the original input and the candidate's result.
See the [complete refund program](../examples/source-criteria-questions/README.md)
and [criteria reference](source-behavior-criteria.md).

## The source owns the question and the criterion

1. Evaluate each considered candidate on initial inputs. A false criterion rejects
   it. An unavailable result keeps the candidate unresolved.
2. Compare the remaining candidates' unlabelled outputs on declared questions.
   Choose the input separating the most pairs, with source order breaking ties.
3. Add only the chosen input to the criterion obligations and evaluate it.
4. Repeat within the declared limits. Record the remaining uncertainty.

Unselected question outputs can be examined without evaluating their predicates.
That distinction is recorded explicitly. The compiler does not manufacture an
expected record. Existing `value_input` and full-record oracle declarations keep
their existing behavior; a block cannot mix those with `criteria_input`.

The analogy is a small diagnostic interview: the source supplies the permitted
questions and what counts as an acceptable answer. The compiler chooses a useful
question, carries the answer forward, and records when the questions run out.

## What completion means here

The receipt separates initial and acquired inputs, candidate evaluations,
predicate evaluations, cache reuse, surviving candidates, rejected candidates,
unavailable candidates and unexamined masks. Predicate results retain their
own satisfied/unsatisfied/unknown denominator. Exact cases, fields and native
caller outputs have separate denominators.

`ONE_SURVIVING_CANDIDATE` requires the entire permitted mask space to have been
considered and no unavailable candidate to remain. With `resolve_unique`, that
candidate can be selected without a model prediction. With incomplete evidence,
selection still uses the existing bounded candidate construction path.

`NO_DISTINGUISHING_INPUT` means the remaining declared questions did not separate
the candidates. Their unasked predicates have not been established. A criterion
can be too weak, a question pool too narrow, or two different bodies can agree
on all supplied inputs. Finite completion describes the checked obligations;
behavior beyond them remains open.

Limits are 32 question tuples, 1–8 acquisition rounds, up to 64 considered masks
from six binary choices, and 128 combined effective/held criteria and exact-case
inputs. An acquired input consumes that combined budget. Held inputs cannot be
placed in the question pool. Candidate errors remain unavailable observations;
cancellation stops work.

## Evidence survives later construction

Acquired inputs remain obligations when a caller tries another body or when
construction resumes. A conflicting caller expectation cannot erase them.
Changing a fixed predicate rechecks candidates on retained inputs while keeping
the original question history and predicate program identity.

Saved evidence can be replayed without a new model prediction. Source formatting,
semantic IR, package lowering and `impact` retain the questions. Caller exposure
includes local calls made while inspecting unselected candidate outputs; partial
call records leave absence evidence unknown. Declared query pools are reserved
as selection inputs even when some questions were not consumed.

The small decision model can still order source-permitted candidates when the
questions leave several possibilities. Its current feature formats do not encode
predicate bodies. This capability has a deterministic path and adds no training.

## Check an answer even when candidates agree

A shared answer can violate the criterion. Add
`criteria_strategy "distinguish_then_check"` inside `observe` to first seek output
disagreement, then acquire the earliest untested source question with an available
output. Its predicates are evaluated only after the input has been chosen.

The [shared wrong-answer example](../examples/source-criteria-challenges/README.md)
contains two alternatives that both preserve the amount. They agree on every
question but fail the negation criterion at a nonzero input. This strategy records
that counterexample and leaves no satisfying candidate in the declared set.

Question receipts use `gooo/criteria-questions/v2` for this explicit strategy and
record `SEPARATE_OUTPUTS` or `CHECK_CRITERIA` for each acquisition. Each untested
probe records available/unavailable output counts. If every remaining question
has unavailable outputs, `QUESTION_OUTPUTS_UNAVAILABLE` leaves the question open;
it does not establish agreement. Candidate and input limits still apply.

A sole survivor can still have untested source questions. With this strategy,
those questions are checked within the round budget before the unique-resolution
shortcut is used. Exhausting the budget leaves `ROUND_BUDGET`. Initial and acquired
criterion results still describe their own finite inputs.

Omitting the strategy preserves the original distinguishing-only behavior and v1
receipt form. Both strategies can be useful: distinguishing answers which allowed
implementation to choose; the additional checks can reveal that all choices fail
a supplied criterion. Neither expands the permitted candidate set. A changed
criterion rechecks retained inputs and preserves the original acquisition reasons.

## Explain a rejected candidate set

The text and Markdown construction reports link rejected candidates to recorded
false predicates or exact-case mismatches. Each reason names its source condition,
input and actual candidate output. The [quote example](../examples/source-criteria-explanations/README.md)
shows why prices 10 and 30 fail a cost floor of 18 and a budget ceiling of 25,
while a revised source choice of 20 satisfies both conditions.

The explanation counts examined, rejected, explained, unavailable and unexamined
candidates separately. All declared candidates are reported rejected only when
the full declared mask set was considered and every rejection has a linked
decisive failure. Missing evidence leaves the explanation incomplete. It selects
one recorded failure per rejected candidate; it does not compute a minimal set
of contradictory predicates or assess other implementations.

These rows describe the initial question history, including after a predicate
change. `body-construct --report saved.json` derives the same explanation from
the stored observation without executing a candidate, a predicate or a model.
Its output identifies recorded claims as unverified. The original JSON form and
saved bytes stay unchanged; existing source replay remains the verification path.
