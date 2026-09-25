# Language value witness

This experiment adds one deliberately small value-level program to Gooo:
`Increment` selects the registered pure operation `int.add` with operand `1`.
The same typed operation registry also contains `int.sub`, `int.mul`, `int.div`,
`int.mod`, `int.neg`, `int.abs`, `int.sign`, `int.max`, `int.min`, `int.iszero`, and `bool.not`, all version 1. `int.neg:0`,
`int.abs:0`, and `int.sign:0`
use the explicit zero literal as a no-configuration sentinel, reject any other
operand at IR validation, and fail closed for `MinInt64` negation or
absolute-value overflow. `int.sign:0` maps negative, zero, and positive inputs
to `-1`, `0`, and `1` without widening the literal grammar. The witness intentionally keeps using `int.add` so
its value-level evidence remains focused on one declared program while the
compiler's additional operations are exercised by separate regression cases,
including checked `int.mul` overflow, fail-closed `int.div` zero-divisor and
minimum-integer overflow cases, and `int.mod` zero-divisor handling.

`sign.gooo` is a second dogfood declaration for `int.sign:0`. It is not
substituted for the canonical witness: the value-witness workflow checks its
semantic binding and generates it twice, requiring identical Go and manifest
outputs. This keeps the new operation visible as a real Gooo source program
without changing the existing self-improvement corpus.

`mod.gooo` applies the same bounded dogfood lane to `int.mod:3`. Its focused
CI observation checks semantic binding, generates Go twice, and compares the
canonical manifests in caller-owned temporary output. The lane is classified as
`COMPILER_REQUIRED` only for this declaration and its replay identity; queued,
runner, security, or other diagnostic observations remain separate rather than
blocking unrelated compiler work or being relabeled as compiler success.
The same declaration is then executed through the existing typed plan CLI with
input `41`; CI requires the detached receipt to report `Remainder = 2`, retain
source and semantic identities, and replay byte-identically. This is execution
evidence for the registered value operation, not evidence of arbitrary generated
Go runtime support.

`neg.gooo` applies the same closed loop to `int.neg:0`. CI executes the
declaration with input `5`, requires the detached receipt to report `Negate = -5`,
and compares the generated Go, manifest, and execution receipt across two runs.
The zero operand is an explicit sentinel for the unary operation; minimum-integer
overflow remains fail-closed in the executor rather than being hidden by the
dogfood example.

`abs.gooo` extends the same declaration-to-receipt path to `int.abs:0`. CI
executes input `-9`, requires `Absolute = 9`, and replays the generated Go,
canonical manifest, and detached execution receipt. The minimum signed integer
case remains an explicit fail-closed executor boundary, not a silently coerced
dogfood result.

`add.gooo` extends the path to a parameterized `int.add:2` operation. CI
executes input `7`, requires `AddTwo = 9`, and replays the generated Go,
canonical manifest, and detached execution receipt without changing the
canonical `int.add:1` self-improvement witness.

`mul.gooo` extends the path to the registered `int.mul:3` operation. CI
executes input `4`, requires `Multiply = 12`, and replays the generated Go,
canonical manifest, and detached execution receipt.

`div.gooo` extends the path to the registered `int.div:2` operation. CI
executes input `8`, requires `Quotient = 4`, and preserves zero-divisor
handling as an explicit `FAIL_CLOSED` boundary rather than accepting an
undefined result.

`max.gooo` extends the path to the comparison primitive `int.max:0`. CI executes both `-7 -> 0` and `7 -> 7`, proving the branch while preserving deterministic generated and receipt replay.

`min.gooo` adds the inverse comparison primitive `int.min:0`. CI executes `-7 -> -7` and `7 -> 0`, so both comparison directions are represented as real `.gooo` declarations with deterministic replay.

`iszero.gooo` crosses the first explicit type boundary: `Integer -> Boolean`. The sealed result authority exposes `Boolean()` only for the declared Boolean entity and accepts only the canonical `0/1` encoding. CI proves `0 -> true` and `7 -> false` through the same replayable plan.

`not.gooo` consumes that Boolean boundary directly. The `bool.not:0` operation accepts only canonical `0/1` input, returns the opposite Boolean encoding, and fails closed for any non-Boolean integer value. CI proves `false -> true` and `true -> false` through the same generated and replayed plan path.

The typed plan regression composes these declarations as `IsZero.result -> Not.input`. It verifies that a Boolean result is delivered as a Boolean input, then observes `0 -> false` and `7 -> true` with two applies and one delivery. This is plan execution evidence, not a claim that runtime bindings are supported by the general Go generator.

The CI receipt records eleven exact input/output cases, eight fail-closed
counterexamples, three reader resolutions, and the fixed scoped coordinate
`0/1 -> 1/1`. The program participates in both the bidirectional and core IR
semantic fingerprints. Core IR preservation and fingerprint sensitivity are
each `1/1`; an unknown declaration attribute remains fail-closed at `1/1`.

Its receipt scope is `REGISTERED_VALUE_OPERATION`, which names the declared
value-evaluation boundary. The scope field alone does not prove that an
`Apply` call completed: the passing report's invoked-operation count, outputs,
and result evidence establish that narrower fact. This does not claim
handwritten Go-body execution or external effects. Historical v2 reports with
the same schema name but no scope are rejected rather than defaulted.

This does not claim a general expression language, arbitrary value types,
core IR execution or code generation, runtime memory or performance bounds, or
authority to mutate the repository.

## From one witness to a typed plan

The same `Integer -> Integer` activity shape can be composed into an explicit
typed execution plan with `CompilePlan` and `Plan.Execute`. The plan validates
every declared result-to-input edge, chooses a deterministic activity order,
records deliveries and results, and rejects missing roots, cycles, overflow, or
tampered authority before unsafe work is applied.

This is intentionally a separate capability boundary. A `.gooo` file may
declare and query a multi-activity bind while the general Go generator still
reports that runtime bindings are unsupported. The correct state is therefore
`FAIL_CLOSED`, not a fabricated generated program. A future self-improvement
candidate must preserve the source and semantic digests, generate the plan,
replay it independently, compare the receipt with the baseline, and only then
propose adoption.
