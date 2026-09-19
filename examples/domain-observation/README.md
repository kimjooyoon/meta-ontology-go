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
compared for byte replay only; they do not demonstrate generated execution of the
original program. The actual budget computation uses the source interpreter.

## Evidence and remaining work

The Actions artifact retains exact source/head/base/workflow/toolchain/run
identities, raw execution and continuation receipts, replay comparisons, real
failures, a labelled synthetic candidate, source checksums, and native test events.
The continuation receipt records requested/completed invocations and actual
feedback deliveries. `observation.json` derives observations from these receipts;
`report.md` shows the bounded result. Build and first-invocation time/RSS are single
run measurements, not inferred improvements.

This is language-owned value continuation, not self-modifying source or autonomous
adoption. Non-negative business budgets, remote work, persistent scheduling,
source-repair acceptance and external usefulness are not proven by this example.
The full self-improvement goal still requires a source-defined repair to be
independently evaluated and its accepted change used by a later execution. No new
whole-language completeness percentage is introduced.
