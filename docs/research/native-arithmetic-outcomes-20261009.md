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

This changes the runtime observation and bounded construction protocol. Ordinary
generated programs retain their original arithmetic behavior. Support for other
language faults, general correctness, unseen-program model accuracy and a public
release containing this change require separate evidence.
