# Native arithmetic outcomes in bounded Gooo construction

## Original failure and test-first change

The published 0.6.15 path ends after a native zero-divisor exit even when a later
source assignment and caller budget remain. The original six-candidate source,
construction inputs and final expectations are retained under
`examples/caller-native-failure`. The two initial regressions were executed
before implementation: independent deliveries were lost at EXECUTE_1, and joint
construction stopped at attempt five. Both now pass locally.

## Authority and execution

The source still owns assignments, typed activities and graph edges. Replay
reconstructs the saved pure Go projection before deriving an observed projection.
Only nonconstant, typed int64 division and remainder expressions are wrapped.
The wrapper observes evaluated operands at the reached operation. Constant
folding, operand evaluation and short-circuit order are retained. The fault binds
the activity/root IDs, original projected expression bytes, byte span and hash.
Observation-only helper names are chosen outside source identifiers.

A typed zero-divisor outcome unwinds the current graph activity. Independent
activities and later case rows continue. A consumer with a faulted or blocked
producer is recorded as blocked; unavailable values are never sent to a callee.
The two fresh executions must produce identical complete native output. Unknown
panics and process/tool errors still return errors; diagnosis text is not used
to classify a language fault. No model is asked to decide whether a fault occurred.

## Selection and replay

Local candidate scoring stays separate. Preflight rejections consume their
original program attempt without native execution; a native fault is an executed
combination with measured expected-output categories. Selection ranks actual
caller matches, preferring a fault-free program on an equal score. A program
with a fault outside the named expectations is not complete. Exhausted budgets
and all-fault spaces retain partial results and every original expected value.

`body-composition-runtime/v3` marks native observations containing faults;
`joint-construction/v6` records their construction histories. Historical
v1–v5 successful runtime semantics are replayed without requiring instrumentation
fields that did not exist then. New fault metadata, native observation hashes,
operands, case identity, expected values and counts are compared during replay.
Current-source replay executes every saved combination and invokes no new model.

## Finite verification scope

Tests cover the original caller-only failure; explicit dependency blocking and
independent results; division/remainder; reached and skipped branches; `&&`/`||`;
first reached nested operation; unused local evaluation; MinInt64 / -1 and values
above 2^53; all-fault partial histories; observer/source name collisions; unknown
panics and cancellation; malformed/missing/contradictory protocol; modified fault
metadata and expectations; and the original variable-zero search counterexample.
The latter now retains a native fault and reaches its original third candidate.

## Actual clean-source observations

The native CLI was built from clean source
`717d876e481c233ae3587488077be87843175fe6` with Go 1.27.1. Its development version
string is still 0.6.15-dev; its source revision distinguishes it from the public
0.6.15 executable. The full bodyexecution race suite passed in 147.685 seconds,
vet passed, and focused CLI regressions passed. Raw results are in the
[observation directory](native-arithmetic-outcomes-20261009/).

| Run | Program attempts | Preflight rejected | Native combinations | Combinations with a native fault | Final cases |
| --- | ---: | ---: | ---: | ---: | ---: |
| Budget 5 | 5 | 2 | 3 | 1 | 1/4 |
| Budget 6 | 6 | 2 | 4 | 1 | 4/4 |
| Saved budget replay | 6 | 2 | 4 | 1 | 4/4 |
| Mixed source order | 48 | 16 | 32 | 8 | 4/4 |
| Mixed own-model order | 6 | 2 | 4 | 1 | 4/4 |
| Saved mixed replay | 6 | 2 | 4 | 1 | 4/4 |

Faulted combinations are a subset of native combinations. Each completed native
observation runs twice; saved attempt counts are historical, while replay executes
that history again. The unchanged graph chooser made one prediction in 33,833 ns.
Record choices use that prediction; subsequent fill assignments use source order.
Both modes selected identical Gooo source. The model is the existing all-data
demonstration QAT ternary model, with 2,096 bytes of resident tensors. Metadata SHA-256
is `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202` and weights SHA-256
is `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
There was no new training or download.

The whole model-mode command took 1.50 seconds wall time, 0.74 seconds user CPU
and 0.50 seconds system CPU; maximum RSS was 87,113,728 bytes (about 83.1 MiB).
These command-level observations include native building/execution. CPU time
divided by wall time is about 82.7% of one core; this is not whole-machine CPU
utilization or a controlled speed comparison. The model prediction interval is
recorded separately.

The recount uses exact JSON numbers and retains all original final expectations,
including integers above 2^53. Historical v3, v4 and v5 records were executed
again with zero inference and equal saved constructions, traces, expectations,
selected source and plan identities. Two selected sources also ran as ordinary
Gooo programs with the same final values. The text copies of the recount and
comparison Go programs accompany the raw compressed outputs.

At the time of these compiler observations, the workbench pinned public 0.6.15.
The subsequent [workbench PR 40](https://github.com/kimjooyoon/gooo-ecosystem-workbench/pull/40)
added v6/runtime-v3 reading and Gooo feedback policies. It passed exact-source CI
and merged as `83ca2df6560f81961512daf2d426b025a2c4d74d`. Its original five-round
feedback and separate evaluation-only fault examples retain the original inputs
and expectations; see its `publication/native-outcomes-20261009` directory.

This changes the runtime observation and bounded construction protocol. Ordinary
generated programs retain their original arithmetic behavior. Support for other
language faults, general correctness, unseen-program model accuracy and a public
release containing this change require separate evidence.
