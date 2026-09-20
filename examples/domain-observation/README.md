# Gooo owns the continuation: a remaining-work budget

This example executes a data-flow program and declares how its successful result
becomes the next invocation's input. The source owns both phases:

```gooo
activity ObserveBudget(Integer) -> Integer computes "int.add:0"
activity ConsumeAttempt(Integer) -> Integer computes "int.add:-1"
activity PublishRemaining(Integer) -> Integer computes "int.add:0"

bind ObserveBudget.result -> ConsumeAttempt.input
bind ConsumeAttempt.result -> PublishRemaining.input
feedback PublishRemaining.result -> ObserveBudget.input
```

The complete program is `main.gooo`. `input.json` provides the initial value 3.
The caller selects an explicit finite iteration budget, not a replacement value:

```sh
gooo run --json --entry ObserveBudget \
  --input examples/domain-observation/input.json --iterations 2 \
  examples/domain-observation/main.gooo
```

For this project's development, commands execute only in GitHub Actions.

## What the two edges mean

`bind` delivers a result inside the current invocation. Those edges must form a
DAG. `feedback` delivers a result only after the entire invocation succeeds, to
an external root input of the next invocation. It cannot target an already bound
input or compete with another incoming edge. Ports, activity identity, arity and
entity types are checked through the same canonical binding model.

Feedback is part of semantic identity, not a comment or an inferred loop. The
IR uses `gooo.runtime-feedback/v1` rather than changing the meaning of an existing
`gooo.runtime-binding/v1` edge. Source spans survive lowering. Formatting, cloning,
semantic equality and Get-Put preserve the phase. Unknown schemas fail closed.

A normal `gooo run` still executes one invocation. `--iterations N` requests N
invocations; N must be positive, and N > 1 requires explicit feedback. Other
caller-supplied roots stay fixed unless a feedback edge names them. The Go runtime
API also accepts a cancellation context, checked between invocations. CLI record
and package continuation are not supported and are rejected rather than ignored.

## Observed behavior and the trust boundary

1. The first invocation reads 3 and publishes 2: three applies, two bind deliveries.
2. The runtime delivers the private produced result along the source feedback edge.
3. The second invocation reads 2 and publishes 1, without the shell manufacturing
   a new input or choosing a continuation endpoint.
4. Same-source/input replay remains CLOSED. Different-input replay is UNKNOWN,
   not an improvement claim and not a trigger for a repair candidate.
5. A real int64 underflow retains partial execution, stage and failing step. If
   the second invocation fails, the third does not run or receive feedback.
6. An independent valid invocation remains usable after a failed one.
7. A separately labelled synthetic receipt corruption is REFUTED and produces
   only a non-executing repair candidate.

Continuation accepts no JSON receipt as execution authority. Only private result
handles issued by this compiled plan can carry feedback, and failed invocations
release none. Returned receipts are detached observations; editing them or making
a digest match cannot resume the plan. External values remain caller inputs, not
claims that the compiler produced them.

## Generation is still a different capability

The generic Go generator rejects the bound/feedback program. `definition.gooo`
is a separate binding-free structural projection. Its two generated outputs are
compared for byte replay and then reverse-observed with `gooo analyze` against
the same Gooo authority. That reverse observation proves only that the generated
markers and registered semantic facts retain the same IR identity with
`write_effect=no-write`; it does not demonstrate generated execution of the
original program. The actual budget computation uses the source interpreter.

When generation is requested with `--runtime-plan`, the generated runtime-plan
artifact now carries the validated typed-plan digest and deterministic activity
order. These fields are compiler evidence for the explicit `bind` graph; they
are not an execution grant, source adoption, or repository-write authority.

## Evidence and remaining work

The Actions artifact retains exact source/head/base/workflow/toolchain/run
identities, raw execution and continuation receipts, replay comparisons, real
failures, a labelled synthetic candidate, source checksums, and native test events.
The continuation receipt records requested/completed invocations and actual
feedback deliveries. `observation.json` derives observations from these receipts;
`report.md` shows the bounded result. Build and first-invocation time/RSS are single
run measurements, not inferred improvements.

This is language-owned value continuation, not self-modifying source or autonomous
adoption. Non-negative business budgets, remote work, persistent scheduling and
external usefulness are not proven by this example.

## Exact source revision boundary

`repair.gooo` is a deliberately small repair fixture. Its baseline operation
adds one to the maximum signed integer and therefore fails with an explicit
`VALUE_INTEGER_OVERFLOW` observation. The caller asks `revise-source` for one
exact activity replacement, receives `candidate.gooo` and `revision.json` in an
external directory, and then calls `evaluate-revision` with the same source
input. The evaluator reaches `CLOSED` only when the baseline failed for the
declared reason and the candidate succeeds. It records both source digests,
the input digest, the baseline failure, candidate execution and the next
operation.

The final `run-accepted-revision --accept` of `candidate.gooo` is an explicit
caller action. The command requires the revision/evaluation/source/input digests
to agree and compares its reexecution with the accepted evaluation receipt. It
returns `next_operation: CAPTURE_NEXT_RUN_COMPARISON` with an empty
`blocked_by` frontier so a later run can resume from an explicit causal point.
The dogfood workflow first calls compare-accepted-revision with the same
baseline, candidate, and input. It reproduces the declared baseline failure,
reproduces the accepted candidate execution, and records CLOSED/IMPROVED with
RECORD_IMPROVEMENT_EVIDENCE; a digest mismatch is REFUTED and an identity
mismatch is UNKNOWN.
The dogfood workflow then generates Go from that accepted candidate and runs the
same reverse observation against `candidate.gooo`, preserving the equality and
no-write evidence as a separate generated-artifact fact.
is not an automatic source write, commit, merge, deployment or compiler
self-modification. The example therefore demonstrates an executable repair loop, while
`source_repair_adoption` remains `EXPLICIT_CALLER` and utility/improvement
claims remain separate observations.

The full self-improvement goal still requires a later system boundary to use
accepted changes under an explicit policy. No new whole-language completeness
percentage is introduced.

The accepted candidate can now cross an explicit external staging boundary.
When the next-run comparison is exactly `CLOSED/IMPROVED`, the example writes
the candidate, comparison receipt, and stage manifest to a caller-owned output
directory, then executes that staged candidate on the next run. The execution
digest must match the accepted receipt. This is not source adoption: the
repository and original `.gooo` input remain unchanged, while UNKNOWN and
REFUTED comparisons are rejected before staging output is created.

## Independent semantic domain fixture

`incident.gooo` keeps the incident-resolution domain separate from the
executable continuation in `main.gooo`. It declares an observed service
symptom, a bounded diagnosis, an explicit resolution, and an observed outcome:

```gooo
activity ObserveSymptom(Service) -> Symptom computes "observation=service-latency;fact=explicit"
activity ProposeDiagnosis(Symptom) -> Diagnosis computes "proposal=bounded;inference=explicit"
activity ApplyResolution(Diagnosis) -> Resolution computes "effect=configuration-change;authority=explicit"
activity RecordOutcome(Resolution) -> Outcome computes "result=observed;claim=explicit"
```

This fixture is used for semantic checking only. Its bindings are intentionally
not presented as generated runtime support. The workflow records its semantic
check separately from the executable continuation, so a failure in one scope
does not become an unsupported claim about the other.
