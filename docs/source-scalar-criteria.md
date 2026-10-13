# Judge Integer paths with Gooo conditions

A calculation can have a clear requirement even when writing an expected value
for every input is inconvenient. Gooo can keep that requirement as a fixed
Boolean activity beside the permitted implementation choices.

Consider “subtract eleven from the input.” At input 11, both `11-input` and
`input-11` return zero. Input 17 distinguishes them. The following complete
program declares the alternatives, the relationship to check, a next question,
and a separate evaluation input:

```gooo
package feecriteria
namespace feecriteria
entity Integer id "fee-criteria://integer"
entity Boolean id "fee-criteria://boolean"
activity Fee(Integer) -> Integer id "fee-criteria://fee" computes "return 11 - input" assembling {
 choice "direction" operand_order at "0" intent "입력에서 11을 뺀다. Subtract eleven from the input."
 criteria {
  predicate "difference" activity "Difference"
  input "[11]"
  holdout_input "[20]"
 }
 observe {
  criteria_strategy "distinguish_then_check"
  criteria_input "[17]"
  max_candidates "2"
  max_rounds "1"
 }
 attempts "2"
}
activity Difference(Integer,Integer) -> Boolean computes "return input1 == input0 - 11"
```

`Fee` retains its `Integer -> Integer` signature. `Difference` receives the
original input and the candidate's actual Integer result. The criterion can
express an equality, inequality or another supported Boolean relationship.
The compiler connects those activities and records each observed result.

All predicates in one `criteria` block share its input tuples. Declare each
selection tuple once, then declare the separate held tuples once. Two predicates
and four selection tuples produce eight predicate/input observations. Repeating
the same tuple after each predicate is a duplicate-input error; indentation does
not make an input belong to just the preceding predicate. The
[bounded fee example](../examples/scalar-source-model-fee-shared-inputs/README.md)
shows two fixed predicates over shared selection and held inputs.

## What happens

1. Check both permitted paths against `Difference` at 11. Both pass.
2. Compare their outputs at the declared question 17. The outputs differ.
3. Acquire 17 as an additional obligation. The forward path returns -6 and fails;
   the reverse path returns 6 and passes.
4. Select the surviving path and evaluate the separately declared input 20.

With `distinguish_then_check`, a question with a shared observable answer can
also be acquired when no remaining question distinguishes candidates. This
matters when every declared implementation makes the same mistake.
The [paired example](../examples/source-scalar-criteria/README.md) records that
case as well as a successful selection.

## Paths and limits

The supported profile has one Integer parameter and an Integer result. It uses
the existing typed path choices:

| Choice | Source decision |
| --- | --- |
| `operand_order` | Order of operands in a supported expression |
| `branch_layout` | Order of the existing branch bodies |
| `local_reference` | Which declared local variable to read |
| `assignment_target` | Which declared local variable to assign |
| `root_order` | Order of supported adjacent statements |

Each local alternative names an existing variable. The typed path compiler
checks each assembled body. A combination that cannot compile remains an
unavailable candidate observation. These decisions preserve the existing path
syntax and supported body subset.

There are at most six binary choices, 1–8 fixed Boolean predicates, 32 question
tuples and 1–8 acquisition rounds. The combined budget is 128 initial/acquired
criterion inputs plus held criterion inputs and declared exact cases. Candidate
and round limits retain counts for choices or questions left unexamined.

`case` and `holdout_case` may accompany criteria when an exact Integer example is
useful. They keep their own counts. A criteria-only program supplies no fabricated
expected output. Held inputs are separate from selection inputs and questions;
the final caller report also records exposure to actual consumed arguments.
`condition_case`, seeded selection, source IR search/body fill, other result
types, and mixed scalar/record choices are outside this criteria profile.

## Reading completion

`2/2` criteria means two predicate/input pairs passed in this example. An
unavailable candidate or failed predicate evaluation is `UNKNOWN`; an evaluated
Boolean false is `UNSATISFIED`. A caller output matching its expected value still
has to preserve the acquired local obligations for construction to be complete.

The separate held criteria, exact cases, and additional native caller evaluation
each have their own denominators. `COMPLETE_FINITE` describes the finite supplied
obligations. The source's choice set, question pool and criterion quality define
what this observation covers.

Criteria and questions survive formatting, semantic IR, bidirectional source
updates, workspace package lowering, caller construction and replay. If a fixed
predicate changes, continuation rechecks attempted candidates while retaining
the original question history and the earlier predicate identity.

## Model and recorded format

This profile retains deterministic selection when no model is connected. The
original paired examples make zero model calls. Record model features keep their
record scope and decline scalar criteria with
`SCALAR_CRITERIA_FEATURE_CONTRACT_UNAVAILABLE`.

Development source containing the [scalar source model adapter](scalar-source-model.ko.md)
can load the separate `gooo/scalar-criteria-model/v1` artifact from SDK
`v0.2.51-experimental`. It ranks the declared combinations once per request;
the existing source evaluator checks candidates against the fixed predicates.
The optional [criteria source flow](criteria-source-flow.ko.md) exports their
typed structure through `body-context --value-flow`, and `--model` prepares the
compatible scalar numerical features without prediction. Finite input tuples
and observed answers are excluded from those numerical rows.

The new adapter examples use synthetic parameters. Learned weights and
source-family evaluation remain separate work. Existing observations keep their
original source/model producers; adding the adapter does not reclassify the
earlier deterministic examples as model observations.

For compatibility, JSON still uses the shared `record_assembly` container.
Its schema identifies the actual profile: `gooo/scalar-criteria-assembly/v1`.
Scalar candidate, predicate and exact holdout records also have explicit scalar
schemas. Field totals are zero; the original function remains Integer-valued.
The tool-facing profile is `scalar_criteria_choices`.

Use a development build containing this change. See the
[behavior criteria reference](source-behavior-criteria.md),
[question selection rules](source-criteria-questions.md), and the
[recorded example](../examples/source-scalar-criteria/README.md).
