# An executable Gooo domain: remaining work budget

This example runs a small data-flow program, rather than only naming an incident
workflow. Gooo owns three computations and the two edges that carry their values:

```gooo
activity ObserveBudget(Integer) -> Integer computes "int.add:0"
activity ConsumeAttempt(Integer) -> Integer computes "int.add:-1"
activity PublishRemaining(Integer) -> Integer computes "int.add:0"

bind ObserveBudget.result -> ConsumeAttempt.input
bind ConsumeAttempt.result -> PublishRemaining.input
```

The complete source is `main.gooo`; `input.json` supplies the external root value
3. The public interface is `gooo run --json --entry ObserveBudget --input
examples/domain-observation/input.json examples/domain-observation/main.gooo`.
For this project's development, these commands execute only in GitHub Actions.

## What actually happens

1. The source interpreter compiles the declared activities and explicit bindings.
2. The first execution reads 3 and publishes 2, with three applies and two deliveries.
3. An independent execution with the same input produces comparable replay evidence.
4. The caller extracts the first observed result as the next input. That execution
   reads 2 and publishes 1. The second value is not hard-coded into the input.
5. Different inputs remain UNKNOWN for same-input replay, not a claimed improvement.
6. An int64 underflow fails at `EXECUTE/apply-int-add`, preserving the first result
   and the two attempted applies. A subsequent valid execution remains usable.
7. A separately labelled synthetic receipt corruption is REFUTED and produces
   only a non-executing repair candidate. UNKNOWN does not produce one.

The source relation chooses within-run delivery. The cross-run feedback edge is
still explicitly assembled by the caller in `observe.sh`; the language does not
own scheduling or autonomous source adoption yet. The script never runs a repair
candidate or overwrites the input source. This example consumes an observed result
in the next execution, but does not claim that the compiler changed itself.

## What is and is not generated

The current generic Go generator rejects `main.gooo` because runtime bindings are
unsupported at that output boundary. `definition.gooo` is a separate binding-free
structural projection. Comparing its two generated outputs proves only byte
replay of that projection, not execution or preservation of the bound program.
The actual computations above use `gooo run`, not the projection.

## Read the evidence, not a score

The Actions artifact includes the source/toolchain/run identities, actual CLI
receipts, replay comparisons, partial failure, next input, synthetic candidate,
build observation, first-process elapsed time and peak RSS, and package test events.
`observation.json` derives its values and operation counts from those receipts.
`report.md` presents the same observations for people. Fixture checks retain the
unchanged bytes of both source files and the input file.

The bounded example does not enforce business rules such as non-negative budgets
or actually attempt remote work. It demonstrates integer computation, source-owned
routing, explicit continuation, and failure isolation. External utility, performance
improvement and autonomous source repair/adoption are not proven here. No new
whole-language completeness percentage is introduced.
