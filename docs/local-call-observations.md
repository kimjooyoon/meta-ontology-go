# Inputs observed during local candidate evaluation

Development source records the helper inputs used while evaluating IR search,
record and body-fill candidates. Public compiler 0.6.29 predates this observation. The record is useful
when a candidate calls a helper and then fails: its inputs still influenced what
the constructor tried, even though the candidate has no completed accuracy score.

`body_search.local_calls` and a caller-selected search candidate's `local_calls`
contain the original evaluator passes, in execution order. Each pass binds the
generated Go digest, root activity ID, case-suite digest, phase and candidate ID.
Calls bind a source activity ID, helper program digest, case index and typed input
tuple. Integers remain exact JSON integers. Record inputs use source field names;
an absent optional field is `null`, distinct from present zero, false or empty text.
The input-exposure reader maps this recorded `null` to an omitted optional field
when comparing with value-case inputs. The saved trace stays unchanged, and
required fields still require a typed value.

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

Body fill retains its original evaluator passes in `body_fill.local_calls`.
The phases are `fill-training`, `fill-probe`, `fill-selected-training` and
`fill-selected-holdout`. Scalar and vector probes use their actual positional
arguments. Record fill retains its value-case passes; it has no synthetic probe
stage. Each candidate binds its generated helper closure, including helpers
introduced by a hole assignment. A failed training pass retains the entered
prefix. When every assignment is rejected, `initial.fill_failure.local_calls`
keeps those observations without inventing a completed score.

Caller-selected assignments retain `local_calls` with `fill-caller-training`
and `fill-caller-holdout` phases. A rejected caller assignment retains its entered
prefix and no native execution. One fill invocation shares the 8,192-call,
1 MiB tuple and 512-pass bounds across its candidate and selected passes.
Candidate scoring continues when capture reaches its bound, with history marked
`UNKNOWN`. The existing replay pass checks the saved typed history, including
failed prefixes. Legacy absence remains absent.

The input-separation metric now includes these recorded helper arguments from
initial construction and every consumed record/search/fill candidate. This includes
local rejections that never reached native execution. An input reached before a
failure counts as used during construction, even if that candidate received no
completed score. The metric compares source activity identities and typed values;
it preserves exact integers and optional record fields. Fixed pure helpers have
no source-owned candidate selection and do not create another learned target.

Compositions also retain `search_validations` for source IR-search preflight.
Each entry binds the stage, exact Gooo source checkpoint, activity identity and
original local-call history. `BODY_PREFLIGHT` checks independent bodies before
optional model loading. `ACTIVITY_GENERATION` checks the current source after
required helpers have been assembled. Residual and quadratic grammar probes can
call helpers during either stage. A failed probe keeps its entered calls; a later
validation error keeps all earlier validation entries in the failed composition.
`completed` describes validation completion, independently of each probe outcome.

Saved replay verifies these explicit preflight boundaries against their original
source checkpoints. This repeats their pure validation probes; it does not load
a model. Older compositions with no validation
history stay readable and do not acquire invented observations. Original
record/search selection histories retain their existing replay behavior.

Compositions retain `fill_validations` for the same two stages of source body
fill. Existing fill validation already performs bounded deterministic candidate
scoring, probes and selected-case checks. Each entry now keeps those original
passes with its source checkpoint and activity identity. Failed and canceled
validation keeps its entered prefix; a failure before the evaluator has no call
history. `completed` means validation returned successfully, even when finite
case scores are below 100%.

Replay of a recorded fill validation repeats that deterministic preflight work
and compares its history. It loads no model. These replay passes are additional
work relative to older receipts that did not record fill validation; they do not
provide new independent evaluation cases. Recording during original construction
does not add another scoring pass. Older receipts retain their missing history.

A retained prefix is positive evidence of use. It does not prove that an input
absent from the prefix was never used. Path assembly, path preflight,
candidate-set planning and the remaining construction boundaries still need
complete capture. Missing or partial coverage therefore
keeps other affected inputs `UNKNOWN`. The metric reads source-replayed records;
it runs no extra candidate or model evaluation and does not rewrite old receipts.

In the `caller-search-calls` example, `Main(-1)` reaches `Seed(0)`. That same
helper input was reached by a failing construction candidate. Including the
saved trace classifies two of the three evaluation inputs as overlapping, with
one still unknown. The finite 3/3 result is unchanged; it is not three independent
cases or a general correctness claim. The saved `47b80a0f` observation remains
the original producer's report, with its earlier one-overlap/two-unknown metric.

The [caller-fill-history example](../examples/caller-fill-history/README.md)
retains `Seed(2)` and `Seed(0)` from a rejected assignment. Later successful calls
use those same helper inputs, so two of the three evaluation inputs count as
overlapping. One remains unknown. Its finite output count is separate from this
exposure count; a successful program can still be checked on already used inputs.
