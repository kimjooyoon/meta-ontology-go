# Evaluate a selected record body with source-owned examples

A record field can have several permitted expressions. `value_case` supplies
examples for choosing among them. `holdout_value_case` supplies a separate check
of the body after that choice has ended. Both stay in the Gooo source.

The development implementation supports this declaration with `field_value` and
`field_update` choices. The source-fill profile retains its existing evaluation
route. Use a compiler built from the revision containing this feature.

```gooo
activity Choose(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "direction" field_value at "0" alternative "0 - input" intent "Negate the input."
 value_case "[0]" -> "{\"value\":0}"
 holdout_value_case "[5]" -> "{\"value\":-5}"
 attempts "2"
}
```

Both alternatives match the selection example at zero. Deterministic selection
keeps the first candidate. Its held evaluation is 0/1. That outcome records a
remaining question; it does not restart selection or change the expected answer.
The [complete example](../examples/record-field-holdouts/README.md) adds a caller
whose own requirements can distinguish the candidates.

## Three stages

1. Local record selection scores only `value_case` rows. Held inputs and expected
   values are absent from every supported record model context. Type and shape
   validation occurs before selection. Explicit candidate projection does not
   evaluate held cases.
2. `body-compose --resume-composition` evaluates the body selected at the new
   checkpoint. A changed helper can change this measurement. The preceding
   composition remains an independent record with its preceding result.
3. `body-construct` evaluates the final callable body after the caller has
   finished selecting. `final_record_holdouts` is distinct from each initial
   local evaluation. A construction containing this field uses
   `gooo/joint-construction/v9`; constructions without it retain their schema.

Whole-record matches and matching fields have separate denominators. Both are
finite observations. A failed evaluator records its reason and an unavailable
accuracy; it does not silently become a measured zero. A canceled operation
remains canceled.

The top-level generation `decision` describes completion of typed code generation.
The record assembly status and counts describe its selection cases; the nested
`holdout` describes separate evaluation. Native runtime results carry their own
case counts. Read these stages together when assessing the implementation.

Selection and held inputs must be disjoint canonical positional tuples. Duplicate
held inputs are rejected. The existing total bound of 128 declared cases includes
both sets. Required/optional fields retain their ordinary typed rules.
Input comparison preserves exact int64 values, including values above 2^53.
Integer `0` and `-0` count as the same input; text `"0"` and `"-0"` remain
different. This comparison leaves the source's JSON spelling intact.

## Changed callers and saved results

Final caller exposure compares held tuples with actual arguments recorded in
all consumed native caller attempts, including losing attempts. `observed_inputs`
means such a call occurred, `other_inputs` means no overlap in those recorded
calls, and `unknown_inputs` preserves an unavailable call history. These counts
cover that caller history. Independent local helper evaluation and model-training
exposure remain outside the measure.

Saved replay recomputes selection and held observations from their source. It
checks source bindings, declared expected values, selected body, case/field
results and final caller exposure. JSON whitespace can change without changing
a typed value. Changed or omitted held results fail replay.

Changing the source's own rules still requires a new source comparison and
construction. Continuation currently retains the declared target contract while
its callable dependencies may change. These held scores do not authorize reuse
under an edited contract or prove behavior on all possible inputs.
