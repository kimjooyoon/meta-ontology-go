# Record questions owned by Gooo source

Development source, 2026-10-13. `field_value` and `field_update` assembly accept
an `observe` block with typed positional `value_input` tuples.

## The problem

A refund of zero cannot distinguish keeping an amount from negating it. Both
implementations satisfy that example. The source can carry a small set of useful
questions and a fixed reference activity:

```gooo
observe {
    value_input "[{\"amount\":17,\"approved\":false,\"label\":\"확인\"},true]"
    value_input "[{\"amount\":61,\"approved\":false,\"label\":\"keep\"},false]"
    oracle "Expected"
    max_candidates "2"
    max_rounds "1"
    reuse_outputs "true"
    resolve_unique "true"
}
```

The first input distinguishes the two bodies. The second takes a branch where
both bodies leave the record unchanged. The compiler chooses the question that
separates the most pairs of surviving candidates, retaining source order for ties.
`Expected` supplies the answer. The candidate bodies only supply possible outcomes.

See the [complete refund program](../examples/source-record-observation/README.md).

## What stays in the program

- The original `value_case` suite remains unchanged and has its own digest.
- Acquired answers form an additional, source-bound selection suite.
- Questions, answers, limits and the reference identity survive formatting,
  semantic IR, package lowering and saved replay.
- A separate fixed `computes` activity must have the same typed input and output
  signature. Its transitive callees must also be fixed pure activities.
- The reference must agree with the original cases before an additional answer
  is acquired. A contradiction is retained in the failure JSON.
- A continuation keeps the original questions and answers. Changed helpers are
  checked against those answers. The current reference must still agree.
- Caller candidates retain acquired local cases through
  `RealizeObservedRecordCandidate`. The legacy explicit candidate API reports
  a missing retained observation when this source declaration is present.

Saved generation replay reproduces the original finite observations without
invoking a model. Checking a caller candidate uses the retained answers and
current reference; it does not choose new questions. A caller expectation that
conflicts with an acquired local answer remains visible in separate counts.

## Bounds and unanswered questions

Each block allows 1..32 distinct typed inputs, 1..64 candidate masks and 1..8
acquired answers. Original cases, acquired cases and held cases together have a
128-case bound. Only the declared question pool is searched. Each record keeps
the existing scalar field and positional input limits.

`ORACLE_UNAVAILABLE` retains a distinguishing input without inventing its answer.
`ROUND_BUDGET`, `CASE_BUDGET` and `CANDIDATE_BUDGET` show the relevant bound.
Evaluation errors retain unavailable candidates. A unique resolution requires
one surviving candidate, no unobserved masks and no unavailable candidate checks.
Several survivors with no distinguishing declared input remain unresolved.

Whole-record equality preserves exact int64 values, strings, Booleans and optional
field presence. `0` and `-0` identify the same integer input. An absent optional
field and a present zero remain distinct.

## Selection, evaluation and model work

`holdout_value_case` inputs must be disjoint from selection and question inputs.
Held answers are evaluated after selection and never enter this observation loop.
Record and field success counts have separate denominators. A finite match, a
unique surviving body and additional native execution are different observations.

The optional model retains its existing input encoding. Acquired answers extend
finite checking; the five existing record encodings do not contain these answers.
With `resolve_unique`, a fully observed unique candidate can be selected before
prediction. Without a model the same source declarations work deterministically.

`record_assembly.observation` records oracle inputs actually evaluated, candidate
evaluation calls and reused outputs. These counts describe work, not latency or
whole-machine CPU usage. Construction text and Markdown also show initial and
acquired cases, surviving/unobserved/unavailable alternatives and unanswered input.
Full JSON retains every round and its values.

This feature covers bounded record field alternatives. Broader natural-language
interpretation, synthesis outside those alternatives and whole-input proofs need
their own evidence.
