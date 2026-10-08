# Gooo selects a declared next search setting

This example starts with `input + hole`, an offset grammar and a two-candidate
cap. The source permits two alternatives: the same grammar with more candidates,
then a contextual residual grammar. A Gooo policy decides when to advance.

```sh
gooo body-refine --source examples/search-feedback/source.gooo.fixture \
  --activity Add --feedback-cases examples/search-feedback/feedback-cases.json \
  --evaluation-cases examples/search-feedback/evaluation-cases.json \
  --policy examples/search-feedback/policy.gooo.fixture --search-policy \
  --max-attempts 8 --max-rounds 6 --out out/search-feedback
```

The declaration is part of the activity's assembly contract:

```gooo
search hole "offset" grammar "integer-offset-constant/v1" intent "Add one to the input." max_candidates "2"
search_alternative "wider" grammar "integer-offset-constant/v1" max_candidates "16"
search_alternative "contextual" grammar "integer-hole-residual/v1" max_candidates "16"
```

Up to four alternatives may name distinct grammar/candidate-bound pairs, in
source order. Each keeps the existing hole, intent and expectations. Revision
updates only the active grammar, candidate bound and attempt budget within that
activity. Alternatives remain declared after selection and contribute to the
semantic contract. Other declarations and the body remain intact.

## What happens in this finite example

1. The first candidate fails; Gooo incorporates the failed direct-input case.
2. Both retained candidates fail, while four expressions are omitted. Gooo
   selects `ADVANCE_SEARCH` with the `wider` alternative.
3. The six retained expressions still miss the needed inner constant. Gooo
   advances to the `contextual` alternative.
4. The contextual grammar derives the inner value `1`. The retained program
   satisfies both feedback inputs, and then both separate final evaluation inputs.

The policy sees current grammar, candidate cap, attempted/retained/omitted counts
and the next unvisited declared alternative. It returns `next_search_id` with
`ADVANCE_SEARCH`. Candidate bounds stay in 2..16. The next attempt limit is the
smaller of the caller's limit and the alternative's cap. A grammar/cap pair is
visited at most once, while budgets and cases can be revised within that pair.
The total run still has at most eight rounds. No model is needed for the policy.

`--search-policy` opts into this extended 16-field policy input and requires an
integer search target. Existing budget/case policies keep their eight-field input.
Their `STOP`, `CONTINUE` and `INCORPORATE` decisions retain their behavior.
Source alternatives alone do not change construction: a policy must select the
transition. Without an unused declared alternative, the example policy keeps the
partial result. The supplied policy also stops at the requested round limit.

Every round retains its source, runtime observations, Gooo decision and revised
source. The selected program replays with zero new inference. Final evaluation
is run only after selection and is never sent to the policy. The two feedback
inputs are adaptive examples; the two evaluation inputs are finite observations,
not evidence of correctness for all integers.
