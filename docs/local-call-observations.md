# Inputs observed during local candidate evaluation

Development source records the helper inputs used while evaluating IR search
candidates. Public compiler 0.6.29 predates this observation. The record is useful
when a candidate calls a helper and then fails: its inputs still influenced what
the constructor tried, even though the candidate has no completed accuracy score.

`body_search.local_calls` and a caller-selected search candidate's `local_calls`
contain the original evaluator passes, in execution order. Each pass binds the
generated Go digest, root activity ID, case-suite digest, phase and candidate ID.
Calls bind a source activity ID, helper program digest, case index and typed input
tuple. Integers remain exact JSON integers. Record inputs use source field names;
an absent optional field is `null`, distinct from present zero, false or empty text.

The phases are `candidate-generation`, `candidate`, `selected-training` and
`selected-holdout`. Probes retain their own generated program and case-suite
digests. Repeated calls remain separate entries. Argument expressions run before
the enclosing call is entered; if a later argument fails, only the helper calls
already entered are recorded. Values are copied after coercion, before the callee
body runs, so later record updates cannot rewrite an earlier input.

`RECORDED` describes capture of the passes in this engine invocation. A failed
evaluation can have recorded inputs and `evaluation_completed: false`.
`cases_completed` counts completed evaluations; it does not replace the existing
scoring contract, which keeps a failed suite incomplete. Missing history remains
unobserved. Cancellation, unavailable identity/value bindings or exceeded bounds
produce `UNKNOWN` with the retained prefix and a reason.

One search invocation retains at most 8,192 helper call entries, 1 MiB of encoded
input tuples and 512 evaluator-pass records. Reaching a recording bound does not
change candidate evaluation or its score. The bound covers the recorded tuple
payload; object, identity and report overhead requires additional memory.

Existing replay passes compare the recorded history without another model call
or an additional scoring pass. Saved JSON indentation does not change typed
values. Historical receipts without this field remain readable with their
original absence; replay does not add a claim about past execution.

Record assembly also captures each original candidate evaluation in
`record_assembly.attempts[].local_calls`, including caller-selected record
candidates. Its phase is `record-candidate`; the candidate ID is the decimal
choice mask. Each record candidate retains at most 8,192 calls and 1 MiB of
encoded tuples. The source-owned attempt budget bounds the number of candidates.
An evaluation failure consumes one attempt and retains the calls already entered;
it receives no completed case score. Cancellation stops construction. Exhausting
the budget before any candidate completes produces an error with its original
record receipt. `body-construct` exposes this as `initial.record_failure`.

Input separation still needs the rest of the construction history: separate
preflight, path/fill assembly evaluations, and every consumed caller
candidate. The current input-separation metric continues to report `UNKNOWN`
when those inputs are unavailable. The new local trace provides evidence for
that next integration; it supplies no general correctness or independence score.
